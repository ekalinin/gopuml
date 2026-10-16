// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/cli"
)

// run runs the command line with the real engine in dir.
func run(t *testing.T, dir, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Chdir(dir)
	var out, errOut bytes.Buffer
	cmd := &cli.Command{Stdin: strings.NewReader(stdin), Stdout: &out, Stderr: &errOut, Engine: cliEngine{}}
	code = cmd.Run(args)
	return code, out.String(), errOut.String()
}

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out
}

func TestNoArgumentsPrintsUsage(t *testing.T) {
	// Spec scenario "No arguments".
	code, stdout, _ := run(t, t.TempDir(), "")
	if code != 0 || !strings.HasPrefix(stdout, "plantuml - generate diagrams") || !strings.Contains(stdout, "--help-more") {
		t.Errorf("exit %d, stdout %.200q", code, stdout)
	}
}

func TestUnsupportedFlagPDFIsRejected(t *testing.T) {
	// Spec scenario "PDF requested".
	dir := t.TempDir()
	write(t, dir, map[string]string{"diagram.puml": "@startuml\nAlice -> Bob\n@enduml\n"})
	code, _, stderr := run(t, dir, "", "--pdf", "diagram.puml")
	if code != cli.ExitUnsupported || !strings.Contains(stderr, "--pdf") {
		t.Errorf("exit %d, stderr %q", code, stderr)
	}
	if got := listFiles(t, dir); !slices.Equal(got, []string{"diagram.puml"}) {
		t.Errorf("files %v", got)
	}
}

func TestThreadsProduceByteIdenticalOutput(t *testing.T) {
	// Spec scenario "Parallel rendering".
	files := map[string]string{}
	for i := range 12 {
		files[fmt.Sprintf("docs/d%02d.puml", i)] = fmt.Sprintf(
			"@startuml\nparticipant A%d\nA%d -> B : message %d\nnewpage\nB -> A%d : page 2\n@enduml\n", i, i, i, i)
	}
	outputs := map[string]map[string][]byte{}
	for _, threads := range []string{"1", "auto"} {
		dir := t.TempDir()
		write(t, dir, files)
		if code, _, stderr := run(t, dir, "", "--svg", "--threads", threads, "docs/"); code != 0 {
			t.Fatalf("--threads %s: exit %d, stderr %q", threads, code, stderr)
		}
		outputs[threads] = map[string][]byte{}
		for _, f := range listFiles(t, dir) {
			if strings.HasSuffix(f, ".svg") {
				data, _ := os.ReadFile(filepath.Join(dir, f))
				outputs[threads][f] = data
			}
		}
	}
	if len(outputs["1"]) != 24 {
		t.Fatalf("%d outputs, want 24", len(outputs["1"]))
	}
	for f, data := range outputs["1"] {
		if !bytes.Equal(data, outputs["auto"][f]) {
			t.Errorf("%s differs between --threads 1 and --threads auto", f)
		}
	}
}

func TestSkipFreshDoesNotRewriteOutput(t *testing.T) {
	// Spec scenario "Skip fresh output".
	dir := t.TempDir()
	src := "@startuml\nAlice -> Bob : hello\n@enduml\n"
	write(t, dir, map[string]string{"diagram.puml": src})
	if code, _, stderr := run(t, dir, "", "--svg", "diagram.puml"); code != 0 {
		t.Fatalf("first run: exit %d, stderr %q", code, stderr)
	}
	out := filepath.Join(dir, "diagram.svg")
	old := time.Unix(1_000_000_000, 0)
	if err := os.Chtimes(out, old, old); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := run(t, dir, "", "--svg", "--skip-fresh", "diagram.puml"); code != 0 {
		t.Fatalf("skip-fresh run: exit %d, stderr %q", code, stderr)
	}
	if st, _ := os.Stat(out); !st.ModTime().Equal(old) {
		t.Error("a fresh output was rewritten")
	}
	// A changed source makes the output stale.
	write(t, dir, map[string]string{"diagram.puml": "@startuml\nAlice -> Bob : changed\n@enduml\n"})
	run(t, dir, "", "--svg", "--skip-fresh", "diagram.puml")
	if st, _ := os.Stat(out); st.ModTime().Equal(old) {
		t.Error("a stale output was not rewritten")
	}
	// An SVG written by upstream for the same source is fresh too: its
	// metadata is compared decoded (cli/skip-fresh case).
	upstream := `<svg xmlns="http://www.w3.org/2000/svg"><g><?plantuml-src Syp9J4vLqBLJSCfFKh1Io4ZDoSa70000?></g></svg>`
	write(t, dir, map[string]string{"diagram.puml": src, "diagram.svg": upstream})
	run(t, dir, "", "--svg", "--skip-fresh", "diagram.puml")
	if data, _ := os.ReadFile(out); string(data) != upstream {
		t.Error("an upstream output of the same source was rewritten")
	}
}

func TestExtractSourceFromOwnPNG(t *testing.T) {
	// Spec scenario "Extract source from PNG".
	dir := t.TempDir()
	src := "@startuml\nAlice -> Bob : hello\n@enduml\n"
	write(t, dir, map[string]string{"diagram.puml": src})
	if code, _, stderr := run(t, dir, "", "diagram.puml"); code != 0 {
		t.Fatalf("PNG run: exit %d, stderr %q", code, stderr)
	}
	code, stdout, stderr := run(t, dir, "", "--extract-source", "diagram.png")
	want := "------------------------\ndiagram.png\n\n" + src + "\n1.2026.8\n------------------------\n"
	if code != 0 || stdout != want {
		t.Errorf("exit %d, stdout %q, want %q, stderr %q", code, stdout, want, stderr)
	}
}
