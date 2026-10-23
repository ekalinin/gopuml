// SPDX-License-Identifier: MIT
// Provenance: original

// Command site assembles the landing page and the WebAssembly module of its
// playground into a directory that GitHub Pages or any static file server
// can serve.
//
// It copies the page (-src, site/ by default) and reads the release it names
// from its gopuml-release meta tag. The module comes either from a local
// build (-wasm, the output of make wasm) or from the WASM archive of that
// release (-download, the base URL of the release downloads), checked
// against the checksums file of the release. When the release is not
// published yet (its checksums file answers 404), it prints a notice, writes
// nothing and succeeds. Nothing is written when a check fails.
//
// When GITHUB_OUTPUT is set, it appends deploy=true, or deploy=false when it
// wrote nothing, to the file it names.
//
// Usage: go run ./tools/site -out dir (-wasm dir | -download url) [-src dir]
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// wasmFiles are the files of the module, in the wasm directory of the site.
var wasmFiles = []string{"plantuml.wasm", "wasm_exec.js", "plantuml.mjs"}

var releaseMeta = regexp.MustCompile(`<meta name="gopuml-release" content="(v\d+\.\d+\.\d+)">`)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout io.Writer) error {
	fl := flag.NewFlagSet("site", flag.ContinueOnError)
	src := fl.String("src", "site", "directory of the page")
	out := fl.String("out", "", "output directory, replaced by the site")
	local := fl.String("wasm", "", "directory of a local build of the module")
	download := fl.String("download", "", "base URL of the release downloads")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if *out == "" || (*local == "") == (*download == "") {
		return errors.New("usage: site -out dir (-wasm dir | -download url) [-src dir]")
	}

	tag, err := release(*src)
	if err != nil {
		return err
	}
	var module map[string][]byte
	if *local != "" {
		module, err = readModule(*local)
	} else {
		module, err = downloadModule(strings.TrimSuffix(*download, "/"), tag)
	}
	if errors.Is(err, errUnpublished) {
		fmt.Fprintf(stdout, "::notice::%v: nothing to deploy\n", err)
		return setOutput(false)
	}
	if err != nil {
		return err
	}
	if err := write(*src, *out, module); err != nil {
		os.RemoveAll(*out)
		return err
	}
	fmt.Fprintf(stdout, "site of %s written to %s\n", tag, *out)
	return setOutput(true)
}

// release returns the release that the page in src names.
func release(src string) (string, error) {
	page, err := os.ReadFile(filepath.Join(src, "index.html"))
	if err != nil {
		return "", err
	}
	m := releaseMeta.FindAllSubmatch(page, -1)
	if len(m) != 1 {
		return "", fmt.Errorf("%s: %d gopuml-release meta tags, want 1", filepath.Join(src, "index.html"), len(m))
	}
	return string(m[0][1]), nil
}

// readModule reads the files of the module from a local build.
func readModule(dir string) (map[string][]byte, error) {
	module := map[string][]byte{}
	for _, name := range wasmFiles {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		module[name] = data
	}
	return module, nil
}

var errUnpublished = errors.New("release not published")

// downloadModule downloads the WASM archive of the release tag from base,
// checks it against the checksums file of the release and returns its
// files.
func downloadModule(base, tag string) (map[string][]byte, error) {
	sums, status, err := get(base + "/" + tag + "/checksums.txt")
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s has no %s/checksums.txt", errUnpublished, tag, base)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%s/%s/checksums.txt: status %d", base, tag, status)
	}
	archive := "gopuml_" + strings.TrimPrefix(tag, "v") + "_wasm.tar.gz"
	var want string
	for line := range strings.SplitSeq(string(sums), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[1] == archive {
			want = f[0]
		}
	}
	if want == "" {
		return nil, fmt.Errorf("checksums.txt of %s has no %s", tag, archive)
	}
	data, status, err := get(base + "/" + tag + "/" + archive)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%s/%s/%s: status %d", base, tag, archive, status)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("%s: SHA-256 %s, checksums.txt has %s", archive, got, want)
	}
	return unpack(archive, data)
}

// get returns the body and the status of a GET request.
func get(url string) ([]byte, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// unpack returns the files of the module in a .tar.gz archive.
func unpack(name string, data []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	module := map[string][]byte{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		base := path.Base(h.Name)
		if h.Typeflag != tar.TypeReg || !slices.Contains(wasmFiles, base) {
			continue
		}
		if module[base], err = io.ReadAll(tr); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	for _, f := range wasmFiles {
		if module[f] == nil {
			return nil, fmt.Errorf("%s has no %s", name, f)
		}
	}
	return module, nil
}

// write replaces out with the files of src and the module in out/wasm.
func write(src, out string, module map[string][]byte) error {
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(out, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, rel), data, 0o644)
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(out, "wasm"), 0o755); err != nil {
		return err
	}
	for name, data := range module {
		if err := os.WriteFile(filepath.Join(out, "wasm", name), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// setOutput appends the deploy output of the step to the file that
// GITHUB_OUTPUT names, when it is set.
func setOutput(deploy bool) error {
	name := os.Getenv("GITHUB_OUTPUT")
	if name == "" {
		return nil
	}
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "deploy=%t\n", deploy); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
