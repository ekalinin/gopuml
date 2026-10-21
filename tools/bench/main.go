// SPDX-License-Identifier: MIT
// Provenance: original

// Command bench compares the start time and the rendering time of this
// implementation with the upstream jar, on the same machine and with the same
// dot for both (passed through GRAPHVIZ_DOT).
//
// The start time is the median time of separate processes that each render a
// small sequence diagram with -tsvg -pipe. The rendering time of a diagram
// type is the median time of one process that renders the curated cases of
// that type (their sources without the YAML header, each in its own
// directory with its fixtures) with -tsvg -o <dir>. It prints a Markdown
// table and the environment of the run.
//
// Usage: go run ./tools/bench -cli <plantuml binary> [-jar file] [-java cmd] [-dot file] [-start-runs n] [-render-runs n]
package main

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/version"
)

// types are the supported diagram types, as the type field of the curated
// cases names them, in the order of the table.
var types = []string{"sequence", "class", "description", "state", "activity"}

// startDiagram is the small diagram of the start time.
const startDiagram = "@startuml\nAlice -> Bob : hello\nBob --> Alice : ok\n@enduml\n"

type bench struct {
	cli, java, jar, dot   string
	cases                 string
	startRuns, renderRuns int
	// timed runs cmd and returns its wall time; tests replace it.
	timed func(cmd *exec.Cmd) (time.Duration, error)
}

// impl is one implementation under comparison.
type impl struct {
	name string
	// command returns the command line that runs it with args.
	command func(args ...string) *exec.Cmd
}

// row is one line of the table: a median per implementation.
type row struct {
	label   string
	cases   int
	medians []time.Duration
}

func main() {
	log.SetFlags(0)
	var b bench
	flag.StringVar(&b.cli, "cli", "", "plantuml binary of this implementation")
	flag.StringVar(&b.jar, "jar", ".cache/upstream/plantuml-mit-"+version.Upstream+".jar", "upstream jar")
	flag.StringVar(&b.java, "java", "java", "java command")
	flag.StringVar(&b.dot, "dot", os.Getenv("GRAPHVIZ_DOT"), "dot of both implementations (default: GRAPHVIZ_DOT, else dot in PATH)")
	flag.StringVar(&b.cases, "cases", "testdata/conformance/cases", "curated cases")
	flag.IntVar(&b.startRuns, "start-runs", 10, "runs of the start time")
	flag.IntVar(&b.renderRuns, "render-runs", 5, "runs of the rendering time of each type")
	flag.Parse()
	if err := b.run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func (b *bench) run(w io.Writer) error {
	if b.cli == "" {
		return errors.New("-cli is required")
	}
	if b.startRuns < 1 || b.renderRuns < 1 {
		return errors.New("the numbers of runs must be positive")
	}
	if b.dot == "" {
		dot, err := exec.LookPath("dot")
		if err != nil {
			return fmt.Errorf("no dot: set -dot or GRAPHVIZ_DOT: %w", err)
		}
		b.dot = dot
	}
	if b.timed == nil {
		b.timed = timed
	}
	for _, p := range []*string{&b.cli, &b.jar, &b.dot} {
		abs, err := filepath.Abs(*p)
		if err != nil {
			return err
		}
		*p = abs
	}
	impls := []impl{
		{"PlantUML " + version.Upstream + " (Java)", func(args ...string) *exec.Cmd {
			return exec.Command(b.java, append([]string{"-Djava.awt.headless=true", "-jar", b.jar}, args...)...)
		}},
		{"gopuml", func(args ...string) *exec.Cmd { return exec.Command(b.cli, args...) }},
	}

	cases, err := conformance.LoadCases(os.DirFS(b.cases), ".")
	if err != nil {
		return err
	}
	work, err := os.MkdirTemp("", "gopuml-bench-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	var rows []row
	var notes []string
	start := row{label: "Start time (one small sequence diagram)"}
	for _, im := range impls {
		var times []time.Duration
		for range b.startRuns {
			cmd := b.command(im, "-tsvg", "-pipe")
			cmd.Stdin = strings.NewReader(startDiagram)
			d, err := b.timed(cmd)
			if err != nil {
				return fmt.Errorf("%s: start: %w", im.name, err)
			}
			times = append(times, d)
		}
		start.medians = append(start.medians, median(times))
	}
	rows = append(rows, start)

	for _, typ := range types {
		ids, files, err := stage(cases, typ, b.cases, filepath.Join(work, typ))
		if err != nil {
			return err
		}
		if len(files) == 0 {
			continue
		}
		r := row{label: "Rendering: " + typ, cases: len(files)}
		for _, im := range impls {
			out := filepath.Join(work, "out")
			var times []time.Duration
			var failure string
			for range b.renderRuns {
				if err := os.RemoveAll(out); err != nil {
					return err
				}
				cmd := b.command(im, append([]string{"-tsvg", "-o", out}, files...)...)
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				d, err := b.timed(cmd)
				var exit *exec.ExitError
				if err != nil && !errors.As(err, &exit) {
					return fmt.Errorf("%s: %s: %w", im.name, typ, err)
				}
				if err != nil && failure == "" {
					// A case with an error still renders its error image;
					// the run counts, and the note names the cases that
					// the messages name.
					var named []string
					for i, f := range files {
						if strings.Contains(stderr.String(), f) {
							named = append(named, ids[i])
						}
					}
					about := strings.Join(named, ", ")
					if about == "" {
						about, _, _ = strings.Cut(strings.TrimSpace(stderr.String()), "\n")
					}
					failure = fmt.Sprintf("%s, %s: %v; messages about %s", im.name, typ, err, about)
				}
				times = append(times, d)
			}
			if failure != "" {
				notes = append(notes, failure)
			}
			r.medians = append(r.medians, median(times))
		}
		rows = append(rows, r)
	}

	fmt.Fprintf(w, "| | %s | %s |\n|---|---:|---:|\n", impls[0].name, impls[1].name)
	for _, r := range rows {
		label, cells := r.label, make([]string, len(r.medians))
		if r.cases > 0 {
			label = fmt.Sprintf("%s (%d %s)", r.label, r.cases, plural(r.cases, "case"))
		}
		for i, m := range r.medians {
			cells[i] = duration(m)
			if r.cases > 0 {
				cells[i] += fmt.Sprintf(" (%s per case)", duration(m/time.Duration(r.cases)))
			}
		}
		fmt.Fprintf(w, "| %s | %s |\n", label, strings.Join(cells, " | "))
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Machine: %s/%s, %s, %d CPUs\n", runtime.GOOS, runtime.GOARCH, cpu(), runtime.NumCPU())
	fmt.Fprintf(w, "- Java: %s\n", firstLine(exec.Command(b.java, "-version")))
	fmt.Fprintf(w, "- Go: %s\n", runtime.Version())
	fmt.Fprintf(w, "- Graphviz: %s\n", firstLine(exec.Command(b.dot, "-V")))
	fmt.Fprintf(w, "- Runs: median of %d start %s and of %d rendering %s per type\n", b.startRuns, plural(b.startRuns, "run"), b.renderRuns, plural(b.renderRuns, "run"))
	for _, n := range notes {
		fmt.Fprintf(w, "- Exit status of %s\n", n)
	}
	return nil
}

// command returns the command of im with the dot of the comparison.
func (b *bench) command(im impl, args ...string) *exec.Cmd {
	cmd := im.command(args...)
	cmd.Env = append(os.Environ(), "GRAPHVIZ_DOT="+b.dot)
	return cmd
}

func timed(cmd *exec.Cmd) (time.Duration, error) {
	start := time.Now()
	err := cmd.Run()
	return time.Since(start), err
}

// stage writes the sources of the cases of typ, each with its fixtures in its
// own directory below dir, and returns their ids and files.
func stage(cases []*conformance.Case, typ, casesDir, dir string) (ids, files []string, err error) {
	for _, c := range cases {
		if c.Type != typ {
			continue
		}
		caseDir := filepath.Join(dir, filepath.FromSlash(c.ID))
		if err := os.MkdirAll(caseDir, 0o755); err != nil {
			return nil, nil, err
		}
		file := filepath.Join(caseDir, path.Base(c.ID)+".puml")
		if err := os.WriteFile(file, []byte(c.Source), 0o644); err != nil {
			return nil, nil, err
		}
		fixtures := filepath.Join(casesDir, filepath.FromSlash(c.ID)+".files")
		if err := os.CopyFS(caseDir, os.DirFS(fixtures)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, nil, err
		}
		ids = append(ids, c.ID)
		files = append(files, file)
	}
	return ids, files, nil
}

func median(times []time.Duration) time.Duration {
	s := slices.Clone(times)
	slices.SortFunc(s, cmp.Compare)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

func duration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.2f s", d.Seconds())
	case d >= 10*time.Millisecond:
		return fmt.Sprintf("%d ms", d.Round(time.Millisecond).Milliseconds())
	default:
		return fmt.Sprintf("%.1f ms", float64(d)/float64(time.Millisecond))
	}
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// firstLine returns the first line that cmd prints on stdout or stderr.
func firstLine(cmd *exec.Cmd) string {
	out, err := cmd.CombinedOutput()
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	if err != nil && line == "" {
		return err.Error()
	}
	return line
}

// cpu returns the model of the processor, when the system tells it.
func cpu() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "linux":
		data, _ := os.ReadFile("/proc/cpuinfo")
		for line := range strings.SplitSeq(string(data), "\n") {
			if name, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(name) == "model name" {
				return strings.TrimSpace(value)
			}
		}
	}
	return "unknown CPU"
}
