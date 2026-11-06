// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakePlantUML logs its dot and arguments, and the first line of each input
// file with whether its fixture is next to it. It fails on the class case.
const fakePlantUML = `
echo "$(basename "$0") dot=$GRAPHVIZ_DOT args=$*" >> "$BENCH_LOG"
status=0
for f in "$@"; do
	case "$f" in
	*.puml)
		echo "$(basename "$0") file=$(head -n 1 "$f") fixture=$(test -f "$(dirname "$f")/inc.iuml" && echo yes || echo no)" >> "$BENCH_LOG"
		case "$f" in *class-diagram/c/c.puml) echo "Error line 1 in file: $f" >&2; status=200;; esac
		;;
	esac
done
exit $status
`

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// fakeBench returns a comparison of the fake commands over two sequence
// cases, one class case and one unsupported case, with 3 start runs and 2
// rendering runs, and the log file of the commands.
func fakeBench(t *testing.T) (bench, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake commands are shell scripts")
	}
	dir := t.TempDir()
	header := func(id, typ string) string {
		return "---\nid: " + id + "\nfamily: uml\ntype: " + typ + "\nformats: [svg]\n---\n"
	}
	writeFiles(t, dir, map[string]string{
		"cases/sequence-diagram/a.puml":           header("sequence-diagram/a", "sequence") + "@startuml\nA -> B\n@enduml\n",
		"cases/sequence-diagram/b.puml":           header("sequence-diagram/b", "sequence") + "@startuml\n!include inc.iuml\n@enduml\n",
		"cases/sequence-diagram/b.files/inc.iuml": "A -> B\n",
		"cases/class-diagram/c.puml":              header("class-diagram/c", "class") + "@startuml\nclass C\n@enduml\n",
		"cases/diagram-source/u.puml":             header("diagram-source/u", "unsupported") + "@startgantt\n@endgantt\n",
		"bin/java":                                "#!/bin/sh\nif [ \"$1\" = -version ]; then echo 'openjdk version \"21.0.2\" 2024-01-16' >&2; echo 'OpenJDK Runtime' >&2; exit 0; fi\n" + fakePlantUML,
		"bin/plantuml":                            "#!/bin/sh\n" + fakePlantUML,
		"bin/dot":                                 "#!/bin/sh\necho 'dot - graphviz version 2.43.0 (0)' >&2\n",
	})
	logFile := filepath.Join(dir, "log")
	t.Setenv("BENCH_LOG", logFile)
	return bench{
		cli:        filepath.Join(dir, "bin", "plantuml"),
		java:       filepath.Join(dir, "bin", "java"),
		jar:        filepath.Join(dir, "plantuml.jar"),
		dot:        filepath.Join(dir, "bin", "dot"),
		cases:      filepath.Join(dir, "cases"),
		startRuns:  3,
		renderRuns: 2,
	}, logFile
}

// column returns the column of a command of the comparison: the base name
// of the command, and for the CLI, whether it runs with a named dot.
func column(cmd *exec.Cmd) string {
	name := filepath.Base(cmd.Path)
	if name != "plantuml" {
		return name
	}
	if slices.ContainsFunc(cmd.Env, func(kv string) bool { return strings.HasPrefix(kv, "GRAPHVIZ_DOT=") }) {
		return name + " (dot)"
	}
	return name + " (gogv)"
}

func TestComparisonTable(t *testing.T) {
	// Spec scenario "Comparison table", with fake commands.
	b, logFile := fakeBench(t)

	// The durations of the runs of each column, in the order of the runs:
	// 3 start runs, then 2 runs of sequence and 2 of class.
	ms := time.Millisecond
	durations := map[string][]time.Duration{
		"java":            {300 * ms, 100 * ms, 110 * ms, 1 * time.Second, 3 * time.Second, 500 * ms, 500 * ms},
		"plantuml (dot)":  {9 * ms, 1 * ms, 2 * ms, 10 * ms, 30 * ms, 40 * ms, 60 * ms},
		"plantuml (gogv)": {9 * ms, 3 * ms, 4 * ms, 10 * ms, 20 * ms, 30 * ms, 50 * ms},
	}
	b.timed = func(cmd *exec.Cmd) (time.Duration, error) {
		_, err := timed(cmd)
		name := column(cmd)
		if len(durations[name]) == 0 {
			t.Fatalf("one run too many of %s", name)
		}
		d := durations[name][0]
		durations[name] = durations[name][1:]
		return d, err
	}
	var out strings.Builder
	if err := b.run(&out); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"| | PlantUML 1.2026.9 (Java) | gopuml (dot) | gopuml (gogv) |\n|---|---:|---:|---:|\n",
		"| Start time (one small sequence diagram) | 110 ms | 2.0 ms | 4.0 ms |\n",
		"| Rendering: sequence (2 cases) | 2.00 s (1.00 s per case) | 20 ms (10 ms per case) | 15 ms (7.5 ms per case) |\n",
		"| Rendering: class (1 case) | 500 ms (500 ms per case) | 50 ms (50 ms per case) | 40 ms (40 ms per case) |\n",
		"- Machine: " + runtime.GOOS + "/" + runtime.GOARCH + ", ",
		"- Java: openjdk version \"21.0.2\" 2024-01-16\n",
		"- Go: " + runtime.Version() + "\n",
		"- Graphviz: dot - graphviz version 2.43.0 (0)\n",
		"- gogv: unknown\n",
		"- Runs: median of 3 start runs and of 2 rendering runs per type\n",
		"- Exit status of PlantUML 1.2026.9 (Java), class: exit status 200; messages about class-diagram/c\n",
		"- Exit status of gopuml (dot), class: exit status 200; messages about class-diagram/c\n",
		"- Exit status of gopuml (gogv), class: exit status 200; messages about class-diagram/c\n",
	}
	for _, w := range want {
		if !strings.Contains(out.String(), w) {
			t.Errorf("no %q in\n%s", w, out.String())
		}
	}
	if strings.Contains(out.String(), "unsupported") || strings.Contains(out.String(), "description") {
		t.Errorf("a row of a type without supported cases:\n%s", out.String())
	}
	for name, left := range durations {
		if len(left) > 0 {
			t.Errorf("%d run(s) of %s missing", len(left), name)
		}
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	log := string(data)
	for _, w := range []string{
		"java file=@startuml fixture=yes\n",
		"plantuml file=@startuml fixture=yes\n",
	} {
		if !strings.Contains(log, w) {
			t.Errorf("no %q in the log of the commands:\n%s", w, log)
		}
	}
	for line := range strings.SplitSeq(log, "\n") {
		if strings.Contains(line, "file=") && !strings.Contains(line, "file=@startuml ") {
			t.Errorf("a case with its YAML header: %s", line)
		}
	}
}

func TestComparisonLayout(t *testing.T) {
	// Spec scenario "Layout of each column", with fake commands: -dot names
	// the fake dot, and GRAPHVIZ_DOT of the environment another one.
	b, logFile := fakeBench(t)
	other := filepath.Join(filepath.Dir(b.dot), "other-dot")
	t.Setenv("GRAPHVIZ_DOT", other)
	if err := b.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	log := string(data)

	// The 3 start runs, then 2 rendering runs of each of the sequence and
	// class types, per column.
	jar := "java dot=" + b.dot + " args=-Djava.awt.headless=true -jar " + b.jar + " -tsvg "
	for _, w := range []struct {
		prefix string
		runs   int
	}{
		{jar + "-pipe\n", 3},
		{jar + "-o ", 4},
		{"plantuml dot=" + b.dot + " args=-tsvg -pipe\n", 3},
		{"plantuml dot=" + b.dot + " args=-tsvg -o ", 4},
		{"plantuml dot= args=-tsvg -pipe\n", 3},
		{"plantuml dot= args=-tsvg -o ", 4},
	} {
		if n := strings.Count(log, w.prefix); n != w.runs {
			t.Errorf("%d runs of %q, want %d, in the log of the commands:\n%s", n, w.prefix, w.runs, log)
		}
	}
	if n := strings.Count(log, " dot="); n != 3*(3+4) {
		t.Errorf("%d runs, want %d, in the log of the commands:\n%s", n, 3*(3+4), log)
	}
	if strings.Contains(log, other) {
		t.Errorf("a command with the GRAPHVIZ_DOT of the environment:\n%s", log)
	}
}

func TestGogvVersion(t *testing.T) {
	const gogv = "github.com/ekalinin/gogv"
	for _, c := range []struct {
		name string
		deps []*debug.Module
		want string
	}{
		{"dependency", []*debug.Module{{Path: "go.yaml.in/yaml/v3", Version: "v3.0.4"}, {Path: gogv, Version: "v0.2.0"}}, "v0.2.0"},
		{"replacement", []*debug.Module{{Path: gogv, Version: "v0.2.0", Replace: &debug.Module{Path: gogv, Version: "v0.2.1-0.20261101000000-abcdef123456"}}}, "v0.2.1-0.20261101000000-abcdef123456"},
		{"local replacement", []*debug.Module{{Path: gogv, Version: "v0.2.0", Replace: &debug.Module{Path: "../gogv"}}}, "../gogv"},
		{"no gogv", []*debug.Module{{Path: "go.yaml.in/yaml/v3", Version: "v3.0.4"}}, "unknown"},
	} {
		if got := moduleVersion(&debug.BuildInfo{Deps: c.deps}, gogv); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
	if got := gogvVersion(filepath.Join(t.TempDir(), "no-cli")); got != "unknown" {
		t.Errorf("no CLI: %q, want unknown", got)
	}
}
