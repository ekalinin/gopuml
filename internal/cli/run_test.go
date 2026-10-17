// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/code"
)

// fakeEngine splits a source into @start/@end blocks. A line "foo bar
// baz" is a syntax error, a line "unsupported" makes the block unsupported
// and each "newpage" line adds a page. Render writes the block text.
type fakeEngine struct {
	mu      sync.Mutex
	sources []Source
}

type fakeBlock struct {
	lines   []string
	err     *BlockError
	pages   int
	renders *int
}

func (e *fakeEngine) Read(src *Source) ([]Block, error) {
	e.mu.Lock()
	e.sources = append(e.sources, *src)
	e.mu.Unlock()
	var blocks []Block
	var cur *fakeBlock
	for i, line := range strings.Split(src.Text, "\n") {
		switch {
		case strings.HasPrefix(line, "@start"):
			cur = &fakeBlock{pages: 1}
		case cur == nil:
			continue
		case line == "foo bar baz" && cur.err == nil:
			cur.err = &BlockError{Line: i, Messages: []string{"Syntax Error? (Assumed diagram type: sequence)"}}
		case line == "unsupported":
			cur.err = &BlockError{Unsupported: true, Line: i}
		case line == "newpage":
			cur.pages++
		}
		cur.lines = append(cur.lines, line)
		if strings.HasPrefix(line, "@end") {
			blocks = append(blocks, cur)
			cur = nil
		}
	}
	return blocks, nil
}

func (b *fakeBlock) Name() (string, bool) {
	_, name, ok := strings.Cut(b.lines[0], " ")
	return name, ok
}

func (b *fakeBlock) Error() *BlockError { return b.err }

func (b *fakeBlock) Type() string {
	if b.err != nil {
		return ""
	}
	return "SEQUENCE"
}

func (b *fakeBlock) Description() string {
	if b.err != nil {
		return "(Error)"
	}
	return "(fake)"
}

func (b *fakeBlock) EncodedURL() string { return code.Encode(strings.Join(b.lines, "\n")) }

func (b *fakeBlock) Preprocessed() []string { return b.lines }

func (b *fakeBlock) Pages() int { return b.pages }

func (b *fakeBlock) Metadata() (string, bool) { return strings.Join(b.lines, "\n") + "\n", true }

func (b *fakeBlock) Log() []string {
	var out []string
	for _, l := range b.lines {
		if msg, ok := strings.CutPrefix(l, "!log "); ok {
			out = append(out, "[Log] "+msg)
		}
	}
	return out
}

func (b *fakeBlock) Render(w io.Writer, page int, opt RenderOptions) (int, error) {
	if b.lines[1] == "renderfail" {
		io.WriteString(w, "partial")
		return 0, errors.New("boom")
	}
	if opt.Format == "SVG" {
		fmt.Fprintf(w, "<svg page=%d>%s</svg>", page, b.lines[1])
	} else {
		fmt.Fprintf(w, "%s %d %s\n", opt.Format, page, b.lines[1])
	}
	if b.err != nil {
		return statusError, nil
	}
	return 0, nil
}

// writeTree writes files relative to a fresh working directory.
func writeTree(t *testing.T, files map[string]string) {
	t.Helper()
	t.Chdir(t.TempDir())
	for name, content := range files {
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

type result struct {
	code           int
	stdout, stderr string
	engine         *fakeEngine
}

func runCLI(t *testing.T, stdin string, args ...string) result {
	t.Helper()
	return runCLIEnv(t, nil, stdin, args...)
}

func runCLIEnv(t *testing.T, env map[string]string, stdin string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	e := &fakeEngine{}
	getenv := func(name string) string { return env[name] }
	cmd := &Command{Stdin: strings.NewReader(stdin), Stdout: &out, Stderr: &errOut, Engine: e, Getenv: getenv}
	code := cmd.Run(args)
	return result{code, out.String(), errOut.String(), e}
}

// outputs lists the files created in the working directory, except the
// inputs.
func outputs(t *testing.T, inputs map[string]string) []string {
	t.Helper()
	var got []string
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if _, ok := inputs[filepath.ToSlash(path)]; !ok {
				got = append(got, filepath.ToSlash(path))
			}
		}
		return nil
	})
	return got
}

const (
	good = "@startuml\nAlice -> Bob\n@enduml\n"
	bad  = "@startuml\nAlice -> Bob\nfoo bar baz\n@enduml\n"
)

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		args       []string
		code       int
		stderr     string
		wantOutput []string
	}{
		{
			name: "missing file", args: []string{"does-not-exist.puml"},
			code: ExitNoFileFound, stderr: "No file found\n",
		},
		{
			name: "success", files: map[string]string{"a.puml": good}, args: []string{"--svg", "a.puml"},
			code: ExitOK, wantOutput: []string{"a.svg"},
		},
		{
			name: "syntax error", files: map[string]string{"bad.puml": bad}, args: []string{"--svg", "bad.puml"},
			code: ExitSomeErrors, stderr: "Error line 3 in file: bad.puml\nSome diagram description contains errors\n",
			wantOutput: []string{"bad.svg"},
		},
		{
			name: "syntax error and missing file", files: map[string]string{"bad.puml": bad},
			args: []string{"--svg", "missing.puml", "bad.puml"},
			code: ExitSomeErrors, stderr: "Error line 3 in file: bad.puml\nSome diagram description contains errors\n",
			wantOutput: []string{"bad.svg"},
		},
		{
			name: "missing file next to an existing one", files: map[string]string{"nodiag.puml": "Alice -> Bob\n"},
			args: []string{"--svg", "nodiag.puml", "missing.puml"},
			code: ExitNoDiagram, stderr: "Warning: no image in nodiag.puml\nNo diagram found\n",
		},
		{
			name: "unsupported diagram", files: map[string]string{"u.puml": "@startuml\nunsupported\n@enduml\n"},
			args: []string{"--svg", "u.puml"},
			code: ExitSomeErrors, stderr: "Some diagram description contains errors\n", wantOutput: []string{"u.svg"},
		},
		{
			name: "check syntax", files: map[string]string{"good.puml": good, "bad.puml": bad},
			args: []string{"--check-syntax", "good.puml", "bad.puml"},
			code: ExitSomeErrors, stderr: "Some diagram description contains errors\n",
		},
		{
			name: "check before run", files: map[string]string{"good.puml": good, "bad.puml": bad},
			args: []string{"--check-before-run", "good.puml", "bad.puml"},
			code: ExitSomeErrors, stderr: "Some diagram description contains errors\nNo diagram found\n",
		},
		{
			name: "check before run without errors", files: map[string]string{"good.puml": good},
			args: []string{"--svg", "--check-before-run", "good.puml"},
			code: ExitOK, wantOutput: []string{"good.svg"},
		},
		{
			name: "stop on error", files: map[string]string{"a.puml": bad, "b.puml": good, "c.puml": good},
			args: []string{"--svg", "--stop-on-error", "--threads", "4", "a.puml", "b.puml", "c.puml"},
			code: ExitSomeErrors, stderr: "Error line 3 in file: a.puml\nSome diagram description contains errors\n",
			wantOutput: []string{"a.svg"},
		},
		{
			name: "no error image", files: map[string]string{"bad.puml": bad, "good.puml": good},
			args: []string{"--svg", "--no-error-image", "bad.puml", "good.puml"},
			code: ExitSomeErrors, stderr: "Error line 3 in file: bad.puml\nSome diagram description contains errors\n",
			wantOutput: []string{"good.svg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeTree(t, tt.files)
			r := runCLI(t, "", tt.args...)
			if r.code != tt.code || r.stderr != tt.stderr || r.stdout != "" {
				t.Errorf("exit %d, stdout %q, stderr %q; want exit %d, stderr %q", r.code, r.stdout, r.stderr, tt.code, tt.stderr)
			}
			if got := outputs(t, tt.files); !slices.Equal(got, tt.wantOutput) {
				t.Errorf("outputs %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

func TestRunOutputs(t *testing.T) {
	files := map[string]string{
		"flows.puml": good + "@startuml login\nA -> B\n@enduml\n" + good,
		"pages.puml": "@startuml\nA -> B\nnewpage\nB -> C\n@enduml\n" + good,
	}
	writeTree(t, files)
	if r := runCLI(t, "", "--svg", "flows.puml", "pages.puml"); r.code != ExitOK || r.stderr != "" {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	want := []string{"flows.svg", "flows_001.svg", "login.svg", "pages.svg", "pages_001.svg", "pages_002.svg"}
	if got := outputs(t, files); !slices.Equal(got, want) {
		t.Errorf("outputs %v, want %v", got, want)
	}
	if data, _ := os.ReadFile("pages_001.svg"); string(data) != "<svg page=1>A -> B</svg>" {
		t.Errorf("second page: %q", data)
	}
	if data, _ := os.ReadFile("pages_002.svg"); string(data) != "<svg page=0>Alice -> Bob</svg>" {
		t.Errorf("next block after two pages: %q", data)
	}
}

func TestRunDefaultFormatAndFormatFlag(t *testing.T) {
	files := map[string]string{"a.puml": good}
	writeTree(t, files)
	if r := runCLI(t, "", "a.puml"); r.code != ExitOK || !slices.Equal(outputs(t, files), []string{"a.png"}) {
		t.Errorf("default format: exit %d, outputs %v", r.code, outputs(t, files))
	}
	writeTree(t, files)
	if r := runCLI(t, "", "-f", "utxt", "a.puml"); r.code != ExitOK || !slices.Equal(outputs(t, files), []string{"a.utxt"}) {
		t.Errorf("-f utxt: exit %d, outputs %v", r.code, outputs(t, files))
	}
	// A name upstream does not know is ignored.
	writeTree(t, files)
	if r := runCLI(t, "", "-f", "SVG", "--txt", "a.puml"); r.code != ExitOK || !slices.Equal(outputs(t, files), []string{"a.atxt"}) {
		t.Errorf("-f SVG: exit %d, outputs %v", r.code, outputs(t, files))
	}
	writeTree(t, files)
	if r := runCLI(t, "", "-f", "pdf", "a.puml"); r.code != ExitUnsupported || !strings.Contains(r.stderr, `"pdf"`) {
		t.Errorf("-f pdf: exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunPreproc(t *testing.T) {
	files := map[string]string{"a.puml": good + "@startuml second\nB -> C\n@enduml\n"}
	writeTree(t, files)
	if r := runCLI(t, "", "--preproc", "a.puml"); r.code != ExitOK {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	if got := outputs(t, files); !slices.Equal(got, []string{"a.preproc", "second.preproc"}) {
		t.Errorf("outputs %v", got)
	}
	if data, _ := os.ReadFile("a.preproc"); string(data) != good {
		t.Errorf("a.preproc = %q", data)
	}
}

func TestRunRenderFailure(t *testing.T) {
	files := map[string]string{"a.puml": "@startuml\nrenderfail\n@enduml\n"}
	writeTree(t, files)
	r := runCLI(t, "", "--svg", "a.puml")
	if r.code != ExitSomeErrors || r.stderr != "plantuml: a.puml: boom\nSome diagram description contains errors\n" {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
	if got := outputs(t, files); len(got) != 0 {
		t.Errorf("partial output left: %v", got)
	}
}

func TestRunReadOnlyOutput(t *testing.T) {
	files := map[string]string{"a.puml": good, "a.svg": "old"}
	writeTree(t, files)
	if err := os.Chmod("a.svg", 0o444); err != nil {
		t.Fatal(err)
	}
	if f, err := os.OpenFile("a.svg", os.O_WRONLY, 0); err == nil {
		f.Close()
		t.Skip("read-only files are writable (running as root)")
	}
	r := runCLI(t, "", "--svg", "a.puml")
	abs, _ := filepath.Abs("a.svg")
	if r.code != ExitOK || !strings.HasPrefix(r.stderr, "Cannot write to file "+abs+"\nWarning: no image in a.puml\n") {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
	if r := runCLI(t, "", "--svg", "--overwrite", "a.puml"); r.code != ExitOK || r.stderr != "" {
		t.Errorf("--overwrite: exit %d, stderr %q", r.code, r.stderr)
	}
	if data, _ := os.ReadFile("a.svg"); string(data) == "old" {
		t.Error("--overwrite did not replace the file")
	}
}

func TestRunParallelOutputOrder(t *testing.T) {
	files := map[string]string{}
	var want strings.Builder
	var args []string
	for i := range 40 {
		name := fmt.Sprintf("f%02d.puml", i)
		files[name] = bad
		args = append(args, name)
		fmt.Fprintf(&want, "Error line 3 in file: %s\n", name)
	}
	want.WriteString("Some diagram description contains errors\n")
	writeTree(t, files)
	r := runCLI(t, "", append([]string{"--svg", "--threads", "8"}, args...)...)
	if r.code != ExitSomeErrors || r.stderr != want.String() {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
	if got := len(outputs(t, files)); got != 40 {
		t.Errorf("%d outputs, want 40", got)
	}
}

func TestRunSourcesAndOutputDir(t *testing.T) {
	files := map[string]string{"src/a.puml": good}
	writeTree(t, files)
	r := runCLI(t, "", "--svg", "-DUSER=Alice", "-o", "out", "src/a.puml")
	if r.code != ExitOK {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	if got := outputs(t, files); !slices.Equal(got, []string{"src/out/a.svg"}) {
		t.Errorf("outputs %v", got)
	}
	src := r.engine.sources[0]
	if src.File != "src/a.puml" || src.Text != good || !slices.Equal(src.Defines, []KeyValue{{"USER", "Alice", true}}) {
		t.Errorf("source %+v", src)
	}
}

func TestRunEncodeDecodeURL(t *testing.T) {
	files := map[string]string{"a.puml": good + "@startuml\nB -> C\n@enduml\n"}
	writeTree(t, files)
	r := runCLI(t, "", "--encode-url", "a.puml")
	want := code.Encode("@startuml\nAlice -> Bob\n@enduml") + "\n" + code.Encode("@startuml\nB -> C\n@enduml") + "\n"
	if r.code != ExitOK || r.stdout != want || len(outputs(t, files)) != 0 {
		t.Errorf("--encode-url: exit %d, stdout %q", r.code, r.stdout)
	}
	// Decoding adds @startuml and @enduml to a source without them, and
	// --decode-url wraps the result again, as upstream.
	encoded := code.Encode("Alice -> Bob")
	if r := runCLI(t, "", "--decode-url", encoded); r.code != ExitOK || r.stdout != "@startuml\n@startuml\nAlice -> Bob\n@enduml\n@enduml\n" {
		t.Errorf("--decode-url: exit %d, stdout %q", r.code, r.stdout)
	}
	if r := runCLI(t, "", "--decode-url", "~1zz"); r.code != ExitUnsupported || r.stdout != "" || r.stderr == "" {
		t.Errorf("--decode-url invalid: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestRunParseError(t *testing.T) {
	r := runCLI(t, "", "-D")
	if r.code != ExitParseError || r.stderr != "-D [DEFINE]: missing value\n" {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunDuration(t *testing.T) {
	writeTree(t, map[string]string{"a.puml": good})
	r := runCLI(t, "", "--svg", "--duration", "a.puml")
	if r.code != ExitOK || !strings.HasPrefix(r.stderr, "Duration = ") || !strings.HasSuffix(r.stderr, " seconds\n") {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunWithoutEngine(t *testing.T) {
	writeTree(t, map[string]string{"a.puml": good})
	var out, errOut bytes.Buffer
	cmd := &Command{Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errOut}
	if code := cmd.Run([]string{"a.puml"}); code != ExitUnsupported || !strings.Contains(errOut.String(), "not implemented") {
		t.Errorf("exit %d, stderr %q", code, errOut.String())
	}
	errOut.Reset()
	if code := cmd.Run([]string{"missing.puml"}); code != ExitNoFileFound {
		t.Errorf("missing file without engine: exit %d", code)
	}
}

func TestPipe(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		stdin          string
		code           int
		stdout, stderr string
	}{
		{
			name: "unmarked input", args: []string{"--svg", "-p"}, stdin: "Alice -> Bob\n",
			stdout: "<svg page=0>Alice -> Bob</svg>",
		},
		{
			name: "delimiter after svg", args: []string{"--svg", "-p", "-pipedelimitor", "==="},
			stdin:  good + "@startuml\nBob -> Carol\n@enduml\n",
			stdout: "<svg page=0>Alice -> Bob</svg>===\n<svg page=0>Bob -> Carol</svg>===\n",
		},
		{
			name: "delimiter after text", args: []string{"--txt", "-p", "-pipedelimitor", "==="},
			stdin:  good + good,
			stdout: "ATXT 0 Alice -> Bob\n===\nATXT 0 Alice -> Bob\n===\n",
		},
		{
			name: "empty input", args: []string{"-p"}, stdin: "\n  \n",
			code: ExitNoFileFound,
		},
		{
			name: "error image to stdout and report to stderr", args: []string{"--svg", "-p"}, stdin: bad,
			code: ExitSomeErrors, stdout: "<svg page=0>Alice -> Bob</svg>",
			stderr: "ERROR\n2\nSyntax Error? (Assumed diagram type: sequence)\n",
		},
		{
			name: "pipenostderr", args: []string{"--svg", "-p", "-pipenostderr"}, stdin: bad + good,
			code:   ExitSomeErrors,
			stdout: "ERROR\n2\nSyntax Error? (Assumed diagram type: sequence)\n<svg page=0>Alice -> Bob</svg>",
		},
		{
			name: "no error image", args: []string{"--svg", "-p", "--no-error-image"}, stdin: bad,
			code: ExitSomeErrors, stderr: "ERROR\n2\nSyntax Error? (Assumed diagram type: sequence)\n",
		},
		{
			name: "image index", args: []string{"--svg", "-p", "--pipe-image-index", "1"},
			stdin:  "@startuml\nA -> B\nnewpage\nB -> C\n@enduml\n",
			stdout: "<svg page=1>A -> B</svg>",
		},
		{
			name: "image index too big", args: []string{"--svg", "-p", "--pipe-image-index", "3"},
			stdin: good, stderr: "numImage is too big = 2\n",
		},
		{
			name: "format line", args: []string{"--svg", "-p"}, stdin: "@@@format utxt\n" + good,
			stdout: "UTXT 0 Alice -> Bob\n",
		},
		{
			name: "preproc", args: []string{"--preproc", "-p"}, stdin: good,
			stdout: good,
		},
		{
			name: "encode url", args: []string{"-p", "--encode-url"}, stdin: good,
			stdout: code.Encode("@startuml\nAlice -> Bob\n@enduml") + "\n",
		},
		{
			name: "syntax", args: []string{"-syntax"}, stdin: good,
			stdout: "SEQUENCE\n(fake)\n",
		},
		{
			name: "syntax error", args: []string{"-syntax"}, stdin: bad,
			code: ExitSomeErrors, stdout: "ERROR\n2\nSyntax Error? (Assumed diagram type: sequence)\n",
		},
		{
			name: "syntax of an unsupported diagram", args: []string{"-syntax"}, stdin: "@startuml\nunsupported\n@enduml\n",
			code: ExitSomeErrors,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := runCLI(t, tt.stdin, tt.args...)
			if r.code != tt.code || r.stdout != tt.stdout || r.stderr != tt.stderr {
				t.Errorf("exit %d, stdout %q, stderr %q\nwant exit %d, stdout %q, stderr %q",
					r.code, r.stdout, r.stderr, tt.code, tt.stdout, tt.stderr)
			}
		})
	}
}

func TestPipeSource(t *testing.T) {
	r := runCLI(t, "Alice -> Bob\n", "-p", "-filedir", "/tmp/x", "-filename", "f.puml", "-DA=1")
	if len(r.engine.sources) != 1 {
		t.Fatalf("%d sources", len(r.engine.sources))
	}
	src := r.engine.sources[0]
	if src.Text != "@startuml\nAlice -> Bob\n@enduml\n" || src.File != "" || src.FileDir != "/tmp/x" || src.Filename != "f.puml" ||
		!slices.Equal(src.Defines, []KeyValue{{"A", "1", true}}) {
		t.Errorf("source %+v", src)
	}
}

func TestReadSingleDiagram(t *testing.T) {
	read := func(input string) []string {
		r := &run{p: parse(t)}
		in := &javaLineReader{r: bufio.NewReader(strings.NewReader(input))}
		var got []string
		for s, ok := r.readSingleDiagram(in, true); ok; s, ok = r.readSingleDiagram(in, false) {
			got = append(got, s)
		}
		return got
	}
	for _, tt := range []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"A -> B", []string{"@startuml\nA -> B\n@enduml\n"}},
		{"\n\nA -> B\n", []string{"@startuml\nA -> B\n@enduml\n"}},
		// Lines before a start line are dropped.
		{"junk\n@startuml\nA\n@enduml\n", []string{"@startuml\nA\n@enduml\n"}},
		// A start line with a name ends at any @end line.
		{"@startuml x\nA\n@endfoo\n", []string{"@startuml x\nA\n@endfoo\n"}},
		{"@startmindmap\n* a\n@enduml\n@endmindmap\n", []string{"@startmindmap\n* a\n@enduml\n@endmindmap\n"}},
		// A missing end line is added without a newline.
		{"@startuml\nA", []string{"@startuml\nA\n@enduml"}},
		// Text after the first diagram without a start line is ignored.
		{"@startuml\nA\n@enduml\ntrailing\n", []string{"@startuml\nA\n@enduml\n"}},
		{"@startuml\r\nA\r@enduml\r\n@startuml\nB\n@enduml", []string{"@startuml\nA\n@enduml\n", "@startuml\nB\n@enduml\n"}},
	} {
		if got := read(tt.input); !slices.Equal(got, tt.want) {
			t.Errorf("%q: got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRunConfigLines(t *testing.T) {
	files := map[string]string{
		"a.puml":       good,
		"default.cfg":  "skinparam a 1\r\n",
		"inc/one.iuml": "' one\n",
		"inc/two.iuml": "' two",
		"conf.txt":     "title from config\n",
	}
	writeTree(t, files)
	r := runCLIEnv(t, map[string]string{"PLANTUML_DEFAULT_CONFIG_FILENAME": "default.cfg"}, "",
		"--svg", "-Iinc/*.iuml", "--theme", "cerulean", "--config", "conf.txt", "-Pteoz=true", "-Pflag",
		"-SarrowColor=red", "a.puml")
	if r.code != ExitOK {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	want := []string{"skinparam a 1", "' one", "' two", "!theme cerulean", "title from config",
		"!pragma teoz true", "!pragma flag null", "skinparamlocked arrowColor red"}
	if got := r.engine.sources[0].Config; !slices.Equal(got, want) {
		t.Errorf("config %q\nwant %q", got, want)
	}
	// Pipe mode gets the same lines.
	r = runCLI(t, good, "-p", "-Pteoz=true")
	if got := r.engine.sources[0].Config; !slices.Equal(got, []string{"!pragma teoz true"}) {
		t.Errorf("pipe config %q", got)
	}
}

func TestRunConfigErrors(t *testing.T) {
	writeTree(t, map[string]string{"a.puml": good, "dir/x": ""})
	abs, _ := filepath.Abs("missing.cfg")
	if r := runCLI(t, "", "--config", "missing.cfg", "a.puml"); r.code != ExitParseError || r.stderr != "Cannot read "+abs+"\n" {
		t.Errorf("missing config: exit %d, stderr %q", r.code, r.stderr)
	}
	abs, _ = filepath.Abs("dir")
	if r := runCLI(t, "", "-Idir", "--version"); r.code != ExitParseError || r.stderr != "Cannot have a directory here "+abs+"\n" {
		t.Errorf("directory include: exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunCharset(t *testing.T) {
	files := map[string]string{"a.puml": "@startuml\ncaf\xe9\n@enduml\n"}
	writeTree(t, files)
	r := runCLI(t, "", "--svg", "--charset", "ISO-8859-1", "a.puml")
	if r.code != ExitOK || r.engine.sources[0].Text != "@startuml\ncafé\n@enduml\n" {
		t.Errorf("ISO-8859-1: exit %d, text %q", r.code, r.engine.sources[0].Text)
	}
	r = runCLI(t, "", "--svg", "a.puml")
	if r.code != ExitOK || r.engine.sources[0].Text != "@startuml\ncaf\uFFFD\n@enduml\n" {
		t.Errorf("UTF-8 default: exit %d, text %q", r.code, r.engine.sources[0].Text)
	}
	r = runCLI(t, "\xfe\xff\x00A\x00\n", "-p", "--charset", "UTF-16")
	if len(r.engine.sources) != 1 || r.engine.sources[0].Text != "@startuml\nA\n@enduml\n" {
		t.Errorf("UTF-16 pipe: sources %+v", r.engine.sources)
	}
	if r := runCLI(t, "", "--charset", "nope", "a.puml"); r.code != ExitUnsupported || !strings.Contains(r.stderr, `"nope"`) {
		t.Errorf("unknown charset: exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunQuotedOutputDir(t *testing.T) {
	files := map[string]string{"a.puml": good}
	writeTree(t, files)
	if r := runCLI(t, "", "--svg", "-o", `"out"`, "a.puml"); r.code != ExitOK {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	if got := outputs(t, files); !slices.Equal(got, []string{"out/a.svg"}) {
		t.Errorf("outputs %v", got)
	}
}

func TestRunStdlib(t *testing.T) {
	// No standard library is bundled (design D14): the list is empty.
	if r := runCLI(t, "", "-stdlib"); r.code != ExitOK || r.stdout != "" {
		t.Errorf("exit %d, stdout %q", r.code, r.stdout)
	}
}

func TestRunVersionAndLicense(t *testing.T) {
	r := runCLI(t, "", "--version")
	if r.code != ExitOK || !strings.HasPrefix(r.stdout, "gopuml ") || !strings.HasSuffix(r.stdout, " (PlantUML 1.2026.8 compatible)\n") {
		t.Errorf("--version: exit %d, stdout %q", r.code, r.stdout)
	}
	r = runCLI(t, "", "-license")
	if r.code != ExitOK || !strings.Contains(r.stdout, "under the terms of the MIT License.\n") {
		t.Errorf("-license: exit %d, stdout %q", r.code, r.stdout)
	}
}

// fakeDot writes a dot script that reports Graphviz 14.0.1 and lays out
// any graph as an empty SVG.
func fakeDot(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake dot scripts need a POSIX shell")
	}
	path := filepath.Join(t.TempDir(), "dot")
	script := `#!/bin/sh
if [ "$1" = "-V" ]; then echo "dot - graphviz version 14.0.1 (20251006.0113)" >&2; exit 0; fi
cat >/dev/null
echo '<svg width="62pt" height="44pt" viewBox="0.00 0.00 62.00 44.00" xmlns="http://www.w3.org/2000/svg"></svg>'
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCheckGraphviz(t *testing.T) {
	// The reference output of graph-layout/check-graphviz.
	want := "GraphViz: dot - graphviz version 14.0.1 (20251006.0113)\nInstallation seems OK. File generation OK\n"
	dotPath := fakeDot(t)
	r := runCLIEnv(t, map[string]string{"GRAPHVIZ_DOT": dotPath}, "", "--check-graphviz")
	if r.code != ExitOK || r.stdout != want || r.stderr != "" {
		t.Errorf("GRAPHVIZ_DOT: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	// --dot-path is used too, unlike upstream.
	r = runCLIEnv(t, map[string]string{"GRAPHVIZ_DOT": "/missing/dot"}, "", "--check-graphviz", "--dot-path", `"`+dotPath+`"`)
	if r.code != ExitOK || r.stdout != want {
		t.Errorf("--dot-path: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	r = runCLIEnv(t, map[string]string{"GRAPHVIZ_DOT": "/missing/dot"}, "", "--check-graphviz")
	if r.code == ExitOK || r.stdout != "" || !strings.Contains(r.stderr, "Error: ") {
		t.Errorf("missing dot: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestRunDotSettings(t *testing.T) {
	writeTree(t, map[string]string{"a.puml": good})
	r := runCLI(t, "", "--svg", "--dot-path", "/opt/dot", "--graphviz-timeout", "7", "a.puml")
	if src := r.engine.sources[0]; src.DotPath != "/opt/dot" || src.DotTimeout != 7*time.Second {
		t.Errorf("dot settings %q %v", src.DotPath, src.DotTimeout)
	}
	r = runCLI(t, "", "--svg", "--graphviz-timeout", "7s", "a.puml")
	if src := r.engine.sources[0]; src.DotTimeout != 0 {
		t.Errorf("a timeout that is not a number is ignored, got %v", src.DotTimeout)
	}
}

func TestRunLogLines(t *testing.T) {
	files := map[string]string{"a.puml": "@startuml\n!log hello\nA -> B\n@enduml\n"}
	writeTree(t, files)
	if r := runCLI(t, "", "--svg", "a.puml"); r.code != ExitOK || r.stderr != "[Log] hello\n" {
		t.Errorf("file: exit %d, stderr %q", r.code, r.stderr)
	}
	if r := runCLI(t, files["a.puml"], "-p", "--svg"); r.code != ExitOK || r.stderr != "[Log] hello\n" {
		t.Errorf("pipe: exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestRunExtractSource(t *testing.T) {
	// The SVG metadata of cli/extract-source-svg, written by upstream.
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><g><?plantuml-src Syp9J4vLqBLJSCfFKh1Io4ZDoSa70000?></g></svg>`
	old := `<svg><!--SRC=[Syp9J4vLqBLJSCfFKh1Io4ZDoSa70000]--></svg>`
	writeTree(t, map[string]string{"diagram.svg": svg, "old.svg": old, "none.svg": "<svg/>", "x.png": "not a png"})
	r := runCLI(t, "", "--extract-source", "diagram.svg", "old.svg", "none.svg")
	want := "------------------------\ndiagram.svg\n\n@startuml\nAlice -> Bob : hello\n@enduml\n------------------------\n" +
		"------------------------\nold.svg\n\n@startuml\nAlice -> Bob : hello\n@enduml\n------------------------\n" +
		"------------------------\nnone.svg\n\n------------------------\n"
	if r.code != ExitOK || r.stdout != want || r.stderr != "" {
		t.Errorf("exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	r = runCLI(t, "", "--extract-source", "x.png")
	if r.code != ExitOK || r.stdout != "------------------------\nx.png\n\nnull\n------------------------\n" || !strings.Contains(r.stderr, "not a PNG") {
		t.Errorf("bad PNG: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}
