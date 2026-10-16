// SPDX-License-Identifier: MIT
// Provenance: original

// Command refgen renders conformance references with Java PlantUML v1.2026.8
// in the reference image (tools/reference/Dockerfile).
//
// Curated diagram, CLI and HTTP cases from testdata/conformance/cases are
// recorded into testdata/conformance/reference/<id>/. Vega cases from
// .cache/vega are recorded into .cache/reference/vega/ and are not committed.
// With -png, PNG references go to .cache/reference-png/<id>/.
//
// With -out, the same cases are rendered by the implementation under test
// (-impl) into <out>/<id>/ (PNG into <out>-png/), in the layout of the
// references, for the conformance runner to compare.
//
// With -impl go, the Go implementation renders the cases natively (no
// container); HTTP cases are skipped until the server is wired up.
//
// Usage: go run ./tools/refgen [-suite all|curated|vega] [-png] [-impl java|go] [-out dir]
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/conformance"
)

const (
	casesDir     = "testdata/conformance/cases"
	referenceDir = "testdata/conformance/reference"
	vegaDir      = ".cache/vega"
	vegaRefRoot  = ".cache/reference"
	pngRefDir    = ".cache/reference-png"
	// workRoot holds one scratch directory per run, so that concurrent
	// runs do not share files.
	workRoot = ".cache/reference-work"
	platform     = "linux/amd64"
	jar          = "/opt/plantuml.jar"
	// cliClasspath runs the CLI through RunCli, which disables the
	// time-based decorations of error images.
	cliClasspath = jar + ":/opt/tools"

	plantumlServerImage = "plantuml/plantuml-server:v1.2026.8@sha256:5f6f99ec2fc11c8236e80630d622b0ba22b8d9301b39b18db4bb2b95cae87781"
)

type generator struct {
	root  string
	image string
	png   bool
	user  string
	impl  string
	// binary is the Go command line, built on first use.
	binary string
	// workDir is the scratch directory of this run, relative to root.
	workDir string

	// Output roots, relative to the repository root.
	outCurated, outVega, outPNG string
}

func main() {
	suite := flag.String("suite", "all", "cases to render: curated, vega or all")
	png := flag.Bool("png", false, "also render PNG references into "+pngRefDir)
	image := flag.String("image", "plantuml-go-reference", "reference image name")
	impl := flag.String("impl", "java", "implementation that renders the cases: java or go")
	out := flag.String("out", "", "render into this directory instead of the reference locations")
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("refgen: ")

	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		log.Fatal("run refgen from the repository root")
	}
	switch *impl {
	case "java":
		if err := exec.Command("docker", "image", "inspect", *image).Run(); err != nil {
			log.Fatalf("image %s not found; build it with: docker build --platform %s -t %s tools/reference", *image, platform, *image)
		}
	case "go":
		if *out == "" {
			log.Fatal("-impl go needs -out: references are recorded with Java only")
		}
	default:
		log.Fatalf("unknown implementation %q", *impl)
	}
	g := &generator{
		root: root, image: *image, png: *png, user: fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), impl: *impl,
		outCurated: referenceDir, outVega: vegaRefRoot, outPNG: pngRefDir,
	}
	if *out != "" {
		rel, err := filepath.Rel(root, filepath.Join(root, *out))
		if err != nil || strings.HasPrefix(rel, "..") {
			log.Fatal("-out must be inside the repository")
		}
		g.outCurated, g.outVega, g.outPNG = rel, rel, rel+"-png"
	}

	if err := os.MkdirAll(filepath.Join(root, workRoot), 0o755); err != nil {
		log.Fatal(err)
	}
	work, err := os.MkdirTemp(filepath.Join(root, workRoot), "run-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)
	g.workDir, _ = filepath.Rel(root, work)

	switch *suite {
	case "curated":
		err = g.curated()
	case "vega":
		err = g.vega()
	case "all":
		err = errors.Join(g.curated(), g.vega())
	default:
		err = fmt.Errorf("unknown suite %q", *suite)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// diagramJob is one diagram case prepared in the scratch directory.
type diagramJob struct {
	c       *conformance.Case
	file    string // scratch file, relative to the repository root
	outRoot string // reference root, relative to the repository root
}

func (g *generator) curated() error {
	if err := removeExcept(filepath.Join(g.root, g.outCurated), "vega"); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(g.root, casesDir)); errors.Is(err, fs.ErrNotExist) {
		log.Printf("no curated cases in %s", casesDir)
		return nil
	}
	fsys := os.DirFS(filepath.Join(g.root, "testdata/conformance"))
	cases, err := conformance.LoadCases(fsys, "cases")
	if err != nil {
		return err
	}
	cliCases, err := conformance.LoadCLICases(fsys, "cases")
	if err != nil {
		return err
	}
	httpCases, err := conformance.LoadHTTPCases(fsys, "cases")
	if err != nil {
		return err
	}

	scratch := filepath.Join(g.workDir, "curated")
	if err := os.RemoveAll(filepath.Join(g.root, scratch)); err != nil {
		return err
	}
	var jobs []diagramJob
	for _, c := range cases {
		dir := filepath.Join(scratch, filepath.FromSlash(c.ID))
		name := path.Base(c.ID) + ".puml"
		if err := writeFile(filepath.Join(g.root, dir, name), []byte(c.Source)); err != nil {
			return err
		}
		fixtures := filepath.Join(g.root, casesDir, filepath.FromSlash(c.ID)+".files")
		if err := copyTree(fixtures, filepath.Join(g.root, dir), nil); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		jobs = append(jobs, diagramJob{c: c, file: filepath.Join(dir, name), outRoot: g.outCurated})
	}
	if err := g.renderDiagrams("curated", jobs); err != nil {
		return err
	}
	if err := g.recordCLI(cliCases); err != nil {
		return err
	}
	if err := g.recordHTTP(httpCases); err != nil {
		return err
	}
	log.Printf("curated: %d diagram, %d CLI, %d HTTP case(s) recorded in %s", len(cases), len(cliCases), len(httpCases), g.outCurated)
	return nil
}

func (g *generator) vega() error {
	if _, err := os.Stat(filepath.Join(g.root, vegaDir, ".complete")); err != nil {
		return fmt.Errorf("vega corpus missing; run tools/corpus/fetch-vega.sh first")
	}
	cases, err := conformance.LoadVegaCases(os.DirFS(filepath.Join(g.root, vegaDir)), ".")
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(g.root, g.outVega, "vega")); err != nil {
		return err
	}
	scratch := filepath.Join(g.workDir, "vega")
	if err := os.RemoveAll(filepath.Join(g.root, scratch)); err != nil {
		return err
	}
	// Copy the whole tree so relative includes keep working, then overwrite
	// each case file with its source without the YAML header.
	if err := copyTree(filepath.Join(g.root, vegaDir), filepath.Join(g.root, scratch), func(rel string) bool { return rel == ".complete" }); err != nil {
		return err
	}
	var jobs []diagramJob
	for _, c := range cases {
		file := filepath.Join(scratch, filepath.FromSlash(c.Path))
		if err := writeFile(filepath.Join(g.root, file), []byte(c.Source)); err != nil {
			return err
		}
		jobs = append(jobs, diagramJob{c: c, file: file, outRoot: g.outVega})
	}
	if err := g.renderDiagrams("vega", jobs); err != nil {
		return err
	}
	log.Printf("vega: %d case(s) recorded in %s", len(cases), filepath.Join(g.outVega, "vega"))
	return nil
}

// renderDiagrams renders all jobs in one container run: the in-process
// renderer for svg/png/txt/utxt, then the CLI in batch mode for preproc.
func (g *generator) renderDiagrams(suite string, jobs []diagramJob) error {
	if len(jobs) == 0 {
		return nil
	}
	if g.impl == "go" {
		return g.renderDiagramsGo(jobs)
	}
	scratch := filepath.Join(g.workDir, suite)
	var manifest bytes.Buffer
	var preproc []diagramJob
	outRoot := jobs[0].outRoot
	for _, j := range jobs {
		var formats []string
		for _, f := range j.c.Formats {
			switch {
			case f == "preproc":
				preproc = append(preproc, j)
			case f == "png" && !g.png:
			default:
				formats = append(formats, f)
			}
		}
		if len(formats) == 0 {
			formats = []string{"none"} // still records meta.json
		}
		fmt.Fprintf(&manifest, "%s\t%s\t%s\n", j.c.ID, container(j.file), strings.Join(formats, ","))
	}
	if err := writeFile(filepath.Join(g.root, scratch, "manifest.tsv"), manifest.Bytes()); err != nil {
		return err
	}

	var script strings.Builder
	script.WriteString("set -e\n")
	render := fmt.Sprintf("java -cp %s:/opt/tools ReferenceRender %s %s", jar, container(filepath.Join(scratch, "manifest.tsv")), container(outRoot))
	if g.png {
		render += " " + container(g.outPNG)
	}
	fmt.Fprintln(&script, render)
	if len(preproc) > 0 {
		fmt.Fprintf(&script, "java -jar %s --preproc --ignore-startuml-filename", jar)
		for _, j := range preproc {
			fmt.Fprintf(&script, " %s", shellQuote(container(j.file)))
		}
		script.WriteString(" || true\n")
	}
	if err := writeFile(filepath.Join(g.root, scratch, "run.sh"), []byte(script.String())); err != nil {
		return err
	}
	if err := g.docker(nil, nil, "-v", g.root+":/work", "-w", "/work", g.image, "sh", container(filepath.Join(scratch, "run.sh"))); err != nil {
		return fmt.Errorf("%s: rendering failed: %w", suite, err)
	}

	// The CLI names preproc outputs <base>.preproc, <base>_001.preproc, ...
	for _, j := range preproc {
		dir := filepath.Join(g.root, filepath.Dir(j.file))
		base := strings.TrimSuffix(filepath.Base(j.file), ".puml")
		out := filepath.Join(g.root, outRoot, filepath.FromSlash(j.c.ID))
		for i := 0; ; i++ {
			name := base + ".preproc"
			if i > 0 {
				name = fmt.Sprintf("%s_%03d.preproc", base, i)
			}
			src := filepath.Join(dir, name)
			if _, err := os.Stat(src); err != nil {
				break
			}
			if err := os.Rename(src, filepath.Join(out, fmt.Sprintf("block%d.preproc", i))); err != nil {
				return err
			}
		}
	}
	return nil
}

// textOutputs are created files whose content is stored in the reference.
var textOutputs = []string{".svg", ".atxt", ".utxt", ".preproc", ".txt"}

func (g *generator) recordCLI(cases []*conformance.CLICase) error {
	for _, c := range cases {
		dir := filepath.Join(g.root, g.workDir, "cli", filepath.FromSlash(c.ID))
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for name, content := range c.Files {
			if err := writeFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(content)); err != nil {
				return err
			}
		}
		var stdout, stderr bytes.Buffer
		exit, work := 0, "/case"
		if g.impl == "go" {
			var err error
			if exit, err = g.runCLIGo(c, dir, &stdout, &stderr); err != nil {
				return fmt.Errorf("%s: %w", c.ID, err)
			}
			work = dir
		} else {
			args := []string{"-i", "-v", dir + ":/case", "-w", "/case"}
			for _, k := range slices.Sorted(maps.Keys(c.Env)) {
				args = append(args, "-e", k+"="+c.Env[k])
			}
			args = append(append(args, g.image, "java", "-cp", cliClasspath, "RunCli"), c.Args...)
			err := g.docker(strings.NewReader(c.Stdin), &cliOutput{&stdout, &stderr}, args...)
			if ee, ok := errors.AsType[*exec.ExitError](err); ok {
				exit = ee.ExitCode()
			} else if err != nil {
				return fmt.Errorf("%s: %w", c.ID, err)
			}
		}

		rec := conformance.CLIRecord{
			Args:   c.Args,
			Exit:   exit,
			Stdout: strings.ReplaceAll(stdout.String(), work, "$WORK"),
			Stderr: strings.ReplaceAll(stderr.String(), work, "$WORK"),
		}
		out := filepath.Join(g.root, g.outCurated, filepath.FromSlash(c.ID))
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(dir, p)
			rel = filepath.ToSlash(rel)
			if _, input := c.Files[rel]; input {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(data)
			fr := conformance.FileRecord{Path: rel, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])}
			if slices.Contains(textOutputs, path.Ext(rel)) {
				fr.Stored = true
				if err := writeFile(filepath.Join(out, "files", filepath.FromSlash(rel)), data); err != nil {
					return err
				}
			}
			rec.Files = append(rec.Files, fr)
			return nil
		})
		if err != nil {
			return err
		}
		if err := writeJSON(filepath.Join(out, "cli.json"), rec); err != nil {
			return err
		}
	}
	return nil
}

// cliOutput routes the docker command's stdout and stderr separately.
type cliOutput struct{ stdout, stderr io.Writer }

// Response headers kept in HTTP references.
var keptHeaders = []string{"Content-Type", "Access-Control-Allow-Origin", "Cache-Control", "Location"}

func (g *generator) recordHTTP(cases []*conformance.HTTPCase) error {
	if len(cases) == 0 {
		return nil
	}
	var builtin, pserver *server
	var err error
	if g.impl == "go" {
		// One Go server implements the routes of both reference servers.
		if builtin, err = g.startGoServer(); err != nil {
			return err
		}
		defer builtin.stop()
		pserver = builtin
	} else {
		// The builtin server is always started: it also encodes sources.
		builtin, err = g.startServer(g.image, "java", "-cp", cliClasspath, "RunCli", "--http-server:8080")
		if err != nil {
			return err
		}
		defer builtin.stop()
		if slices.ContainsFunc(cases, func(c *conformance.HTTPCase) bool { return c.Server == "plantuml-server" }) {
			if pserver, err = g.startServer(plantumlServerImage); err != nil {
				return err
			}
			defer pserver.stop()
		}
	}

	client := &http.Client{
		Timeout:       2 * time.Minute,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	for _, c := range cases {
		srv := builtin
		if c.Server == "plantuml-server" {
			srv = pserver
		}
		rec := conformance.HTTPRecord{Server: c.Server, Method: c.Method, Path: c.RequestPath}
		if g.impl == "go" && strings.Contains(c.RequestPath, "{encoded}") {
			// Send the request of the reference: encodings of Java and Go
			// may differ byte-wise.
			var ref conformance.HTTPRecord
			if err := readJSONFile(filepath.Join(g.root, referenceDir, filepath.FromSlash(c.ID), "http.json"), &ref); err != nil {
				return fmt.Errorf("%s: %w", c.ID, err)
			}
			rec.Path, rec.Encoded = ref.Path, ref.Encoded
		} else if strings.Contains(c.RequestPath, "{encoded}") {
			if rec.Encoded, err = g.encode(builtin, c.Source); err != nil {
				return fmt.Errorf("%s: %w", c.ID, err)
			}
			rec.Path = strings.ReplaceAll(c.RequestPath, "{encoded}", rec.Encoded)
		}
		req, err := http.NewRequest(c.Method, srv.url+rec.Path, strings.NewReader(c.Body))
		if err != nil {
			return fmt.Errorf("%s: %w", c.ID, err)
		}
		for k, v := range c.Headers {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("%s: %w", c.ID, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", c.ID, err)
		}
		rec.Status = resp.StatusCode
		rec.Headers = map[string]string{}
		for k, v := range resp.Header {
			if slices.Contains(keptHeaders, k) || strings.HasPrefix(k, "X-Plantuml-") {
				rec.Headers[k] = strings.Join(v, ", ")
			}
		}
		sum := sha256.Sum256(body)
		rec.Body = conformance.FileRecord{Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:])}
		out := filepath.Join(g.root, g.outCurated, filepath.FromSlash(c.ID))
		if ext := bodyExtension(resp.Header.Get("Content-Type")); ext != "" {
			rec.Body.Path, rec.Body.Stored = "body"+ext, true
			if err := writeFile(filepath.Join(out, rec.Body.Path), body); err != nil {
				return err
			}
		}
		if err := writeJSON(filepath.Join(out, "http.json"), rec); err != nil {
			return err
		}
	}
	return nil
}

func bodyExtension(contentType string) string {
	switch ct, _, _ := strings.Cut(contentType, ";"); {
	case ct == "image/svg+xml":
		return ".svg"
	case ct == "application/json":
		return ".json"
	case strings.HasPrefix(ct, "text/"):
		return ".txt"
	}
	return "" // binary bodies are recorded by checksum only
}

type server struct {
	g   *generator
	id  string
	url string
	// cmd is the Go server process; id is empty then.
	cmd *exec.Cmd
}

func (g *generator) startServer(image string, command ...string) (*server, error) {
	args := append([]string{"run", "-d", "--rm", "--platform", platform, "-p", "127.0.0.1::8080", image}, command...)
	out, err := exec.Command("docker", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("starting %s: %w", image, err)
	}
	s := &server{g: g, id: strings.TrimSpace(string(out))}
	port, err := exec.Command("docker", "port", s.id, "8080/tcp").Output()
	if err != nil {
		s.stop()
		return nil, err
	}
	addr, _, _ := strings.Cut(strings.TrimSpace(string(port)), "\n")
	s.url = "http://" + addr
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(s.url + "/"); err == nil {
			resp.Body.Close()
			return s, nil
		}
		time.Sleep(time.Second)
	}
	s.stop()
	return nil, fmt.Errorf("server %s did not start", image)
}

func (s *server) stop() {
	if s.cmd != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
		return
	}
	_ = exec.Command("docker", "rm", "-f", s.id).Run()
}

// encode returns upstream's URL encoding of source.
func (g *generator) encode(s *server, source string) (string, error) {
	cmd := exec.Command("docker", "exec", "-i", s.id, "java", "-jar", jar, "-p", "--encode-url")
	cmd.Stdin = strings.NewReader(source)
	out, err := cmd.Output()
	// Exit status 200 only means the diagram has an error; the encoding is
	// still printed.
	if ee, ok := errors.AsType[*exec.ExitError](err); err != nil && !(ok && ee.ExitCode() == 200) {
		return "", fmt.Errorf("encoding source: %w", err)
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	if line == "" {
		return "", errors.New("encoding source: empty result")
	}
	return line, nil
}

// docker runs a container from the reference image as the current user.
func (g *generator) docker(stdin io.Reader, out *cliOutput, args ...string) error {
	full := append([]string{"run", "--rm", "--platform", platform, "--user", g.user, "-e", "HOME=/tmp"}, args...)
	cmd := exec.CommandContext(context.Background(), "docker", full...)
	cmd.Stdin = stdin
	if out != nil {
		cmd.Stdout, cmd.Stderr = out.stdout, out.stderr
	} else {
		cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	}
	return cmd.Run()
}

// container maps a path relative to the repository root to the container.
func container(rel string) string { return "/work/" + filepath.ToSlash(rel) }

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func writeFile(p string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func writeJSON(p string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(p, append(data, '\n'))
}

// removeExcept deletes the entries of dir, keeping the one named keep.
func removeExcept(dir, keep string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == keep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// copyTree copies src into dst. skip, when not nil, excludes relative paths.
func copyTree(src, dst string, skip func(rel string) bool) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() || (skip != nil && skip(filepath.ToSlash(rel))) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(dst, rel), data)
	})
}
