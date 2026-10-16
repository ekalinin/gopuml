// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const testTimeout = 200 * time.Millisecond

func TestLayoutRunsDotTsvg(t *testing.T) {
	dir := t.TempDir()
	args, stdin := filepath.Join(dir, "args"), filepath.Join(dir, "stdin")
	exe := writeScript(t, `printf '%s\n' "$@" > '`+args+`'
cat > '`+stdin+`'
echo '`+testSVG+`'
`)
	src := "digraph g {\n  a -> b;\n}\n"
	out, err := (&Exec{Path: exe}).Layout(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != testSVG {
		t.Errorf("output = %q", got)
	}
	if got := readFile(t, args); got != "-Tsvg\n" {
		t.Errorf("arguments = %q, want -Tsvg", got)
	}
	if got := readFile(t, stdin); got != src {
		t.Errorf("stdin = %q, want %q", got, src)
	}
}

func TestLayoutIgnoresWarningsOnStderr(t *testing.T) {
	exe := writeScript(t, `cat >/dev/null
echo 'Warning: some warning' >&2
echo '`+testSVG+`'
`)
	out, err := (&Exec{Path: exe}).Layout(context.Background(), "digraph g { a }")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "Warning") {
		t.Errorf("stderr leaked into the SVG: %q", out)
	}
}

func TestRunDotStopsProcessAfterTimeout(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	exe := writeScript(t, "echo $$ > '"+pidFile+"'\nexec sleep 30\n")
	start := time.Now()
	_, err := (&Exec{Path: exe, Timeout: testTimeout}).Layout(context.Background(), "digraph g { a -> b }")
	elapsed := time.Since(start)

	if !errors.Is(err, ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
	te, ok := errors.AsType[*TimeoutError](err)
	if !ok || te.Timeout != testTimeout || te.Path != exe {
		t.Fatalf("err = %#v", err)
	}
	if errors.Is(err, ErrFailed) || errors.Is(err, ErrNotFound) {
		t.Errorf("timeout must not match the other sentinels: %v", err)
	}
	if elapsed > 5*time.Second {
		t.Errorf("Layout returned after %v", elapsed)
	}
	assertStopped(t, pidFile)
}

func TestRunDotKillsProcessIgnoringTerm(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	exe := writeScript(t, "trap '' TERM\necho $$ > '"+pidFile+"'\nexec sleep 30\n")
	start := time.Now()
	_, err := (&Exec{Path: exe, Timeout: testTimeout}).Layout(context.Background(), "digraph g { a }")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Layout returned after %v", elapsed)
	}
	assertStopped(t, pidFile)
}

func TestRunDotReturnsWhenChildKeepsOutputOpen(t *testing.T) {
	// The shell is stopped, but its sleep child keeps stdout open.
	exe := writeScript(t, "sleep 5\n")
	start := time.Now()
	_, err := (&Exec{Path: exe, Timeout: testTimeout}).Layout(context.Background(), "digraph g { a }")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Errorf("Layout waited for the child: %v", elapsed)
	}
}

// assertStopped checks that the process whose PID is in pidFile is gone.
func assertStopped(t *testing.T, pidFile string) {
	t.Helper()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Logf("no PID recorded before the timeout: %v", err)
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	if err := p.Signal(syscall.Signal(0)); err == nil {
		t.Errorf("dot process %d is still running", pid)
	}
}

func TestLayoutReportsDotStderrOnFailure(t *testing.T) {
	const msg = "Error: <stdin>: syntax error in line 1 near 'x'"
	exe := writeScript(t, "cat >/dev/null\necho \""+msg+"\" >&2\nexit 1\n")
	_, err := (&Exec{Path: exe}).Layout(context.Background(), "x")
	if !errors.Is(err, ErrFailed) {
		t.Fatalf("err = %v, want ErrFailed", err)
	}
	fe, ok := errors.AsType[*FailureError](err)
	if !ok || fe.ExitCode != 1 || fe.Stderr != msg || fe.Path != exe {
		t.Fatalf("err = %#v", err)
	}
	if got, want := err.Error(), "Graphviz dot failed: exit status 1: "+msg; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestLayoutRejectsUnusableOutput(t *testing.T) {
	for _, tc := range []struct{ name, script, want string }{
		{"empty", "cat >/dev/null\necho 'Warning: odd' >&2\n", "Graphviz dot failed: empty output: Warning: odd"},
		{"not svg", "cat >/dev/null\necho hello\n", "Graphviz dot failed: output is not SVG"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&Exec{Path: writeScript(t, tc.script)}).Layout(context.Background(), "digraph g { a }")
			if !errors.Is(err, ErrFailed) {
				t.Fatalf("err = %v, want ErrFailed", err)
			}
			if err.Error() != tc.want {
				t.Errorf("Error() = %q, want %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLayoutWithoutUsableExecutable(t *testing.T) {
	_, err := (&Exec{}).Layout(context.Background(), "digraph g { a }")
	if nf, ok := errors.AsType[*NotFoundError](err); !ok || nf.State != NullUndefined {
		t.Errorf("empty path: err = %v", err)
	}
	missing := filepath.Join(t.TempDir(), "dot")
	_, err = (&Exec{Path: missing}).Layout(context.Background(), "digraph g { a }")
	if nf, ok := errors.AsType[*NotFoundError](err); !ok || nf.State != DoesNotExist || nf.Path != missing {
		t.Errorf("missing file: err = %v", err)
	}
	notExec := filepath.Join(t.TempDir(), "dot")
	if err := os.WriteFile(notExec, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = (&Exec{Path: notExec}).Layout(context.Background(), "digraph g { a }")
	if fe, ok := errors.AsType[*FailureError](err); !ok || fe.Err == nil || fe.ExitCode != -1 {
		t.Errorf("not executable: err = %v", err)
	}
}

func TestLayoutHonorsCanceledContext(t *testing.T) {
	exe := writeScript(t, workingDot("x"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&Exec{Path: exe}).Layout(ctx, "digraph g { a }")
	if !errors.Is(err, context.Canceled) || errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestExecTimeoutDefault(t *testing.T) {
	if got := (&Exec{}).timeout(); got != 30*time.Second {
		t.Errorf("default timeout = %v", got)
	}
	if got := (&Exec{Timeout: time.Second}).timeout(); got != time.Second {
		t.Errorf("timeout = %v", got)
	}
}

func TestDotVersion(t *testing.T) {
	exe := writeScript(t, workingDot("x"))
	if got := (&Exec{Path: exe}).DotVersion(context.Background()); got != testVersion {
		t.Errorf("DotVersion = %q, want %q", got, testVersion)
	}
	multi := writeScript(t, "echo '  first'\necho 'second  ' >&2\n")
	if got := (&Exec{Path: multi}).DotVersion(context.Background()); got != "first second" {
		t.Errorf("DotVersion = %q, want %q", got, "first second")
	}
	missing := filepath.Join(t.TempDir(), "dot")
	if got := (&Exec{Path: missing}).DotVersion(context.Background()); got != "?" {
		t.Errorf("DotVersion of a missing dot = %q, want ?", got)
	}
	slow := writeScript(t, "exec sleep 30\n")
	if got := (&Exec{Path: slow, Timeout: testTimeout}).DotVersion(context.Background()); got != "?" {
		t.Errorf("DotVersion of a hanging dot = %q, want ?", got)
	}

	c := Config{DotPath: exe}
	if got := c.DotVersion(context.Background()); got != testVersion {
		t.Errorf("Config.DotVersion = %q", got)
	}
	c = Config{DotPath: "/missing/dot", Host: fakeHost("linux", nil, nil)}
	if got, want := c.DotVersion(context.Background()), "Error:File /missing/dot does not exist"; got != want {
		t.Errorf("Config.DotVersion = %q, want %q", got, want)
	}
}

func TestRetrieveVersion(t *testing.T) {
	for s, want := range map[string]int{
		testVersion:                                     1400,
		"dot - graphviz version 2.43.0 (0)":             243,
		"dot - graphviz version 2.26.3 (20100126.1600)": 226,
		"dot - graphviz version 15.0.0 (20260523.1842)": 1500,
		"dot - Graphviz version 2.8 (Tue Jan 1 2008)":   208,
		"?": -1,
		"Error:File /opt/local/bin/dot does not exist": -1,
		"version 2.38":                       -1, // a non-digit must follow
		"dot version2.38.0":                  -1, // a space must precede
		"dot - graphviz version 2.123.0 (0)": -1,
	} {
		if got := RetrieveVersion(s); got != want {
			t.Errorf("RetrieveVersion(%q) = %d, want %d", s, got, want)
		}
	}
}

func TestNewGraphvizVersion(t *testing.T) {
	type flags struct{ shield, protection, xlabel, vizjs, ignoreHorizontal bool }
	get := func(v GraphvizVersion) flags {
		return flags{v.UseShieldForQuantifier(), v.UseProtectionWhenThereALinkFromOrToGroup(), v.UseXLabelInsteadOfLabel(), v.IsVizjs(), v.IgnoreHorizontalLinks()}
	}
	for s, want := range map[string]flags{
		"?":                                 {true, true, false, false, false},
		"dot - graphviz version 2.28.0 (0)": {true, true, false, false, false},
		"dot - graphviz version 2.30.1 (0)": {false, true, false, false, true},
		"dot - graphviz version 2.39.0 (0)": {false, false, false, false, false},
		"dot - graphviz version 2.40.1 (0)": {false, false, false, false, false},
		testVersion:                         {false, true, false, false, false},
		"version2.38":                       {false, true, false, false, false},
	} {
		if got := get(NewGraphvizVersion(s)); got != want {
			t.Errorf("NewGraphvizVersion(%q) = %+v, want %+v", s, got, want)
		}
	}
	if NewGraphvizVersion("?") != DefaultGraphvizVersion {
		t.Error("unparsable versions must give DefaultGraphvizVersion")
	}
	exe := writeScript(t, workingDot("x"))
	if got := get((&Exec{Path: exe}).GraphvizVersion(context.Background())); got != (flags{false, true, false, false, false}) {
		t.Errorf("Exec.GraphvizVersion = %+v", got)
	}
}

// TestLayoutWithInstalledDot runs the real Graphviz when it is installed.
func TestLayoutWithInstalledDot(t *testing.T) {
	path, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("Graphviz dot is not installed")
	}
	e := &Exec{Path: path}
	out, err := e.Layout(context.Background(), "digraph g { a -> b }")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<svg") || !strings.Contains(string(out), ">a</text>") {
		t.Errorf("unexpected SVG: %s", out)
	}
	_, err = e.Layout(context.Background(), "digraph g { a -> }")
	fe, ok := errors.AsType[*FailureError](err)
	if !ok || fe.ExitCode == 0 || !strings.Contains(fe.Stderr, "syntax error") {
		t.Errorf("syntax error: err = %v", err)
	}
	if v := RetrieveVersion(e.DotVersion(context.Background())); v < dotVersionLimit {
		t.Errorf("version %d", v)
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
