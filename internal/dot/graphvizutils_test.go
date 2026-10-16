// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// TestCheckGraphvizMatchesUpstreamReference compares the check with a fake
// dot reporting the reference image's Graphviz version against the output
// recorded from upstream's --check-graphviz.
func TestCheckGraphvizMatchesUpstreamReference(t *testing.T) {
	data, err := os.ReadFile("../../testdata/conformance/reference/graph-layout/check-graphviz/cli.json")
	if err != nil {
		t.Fatal(err)
	}
	var ref struct {
		Exit   int    `json:"exit"`
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatal(err)
	}
	exe := writeScript(t, workingDot("x"))
	lines, code := CheckGraphviz(context.Background(), Config{DotPath: exe}, false)
	stdout, stderr := strings.Join(lines, "\n")+"\n", ""
	if code != 0 {
		stdout, stderr = "", stdout
	}
	if code != ref.Exit || stdout != ref.Stdout || stderr != ref.Stderr {
		t.Errorf("CheckGraphviz = %d, stdout %q, stderr %q; upstream = %d, %q, %q", code, stdout, stderr, ref.Exit, ref.Stdout, ref.Stderr)
	}
}

func TestAddDotStatus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		script string
		code   int
		lines  []string
	}{
		{"old version", `echo "dot - graphviz version 2.20.2 (Tue Mar  2 19:03:41 UTC 2010)" >&2`, 13,
			[]string{"GraphViz: dot - graphviz version 2.20.2 (Tue Mar  2 19:03:41 UTC 2010)", "Warning : Your dot installation seems old", "Some diagrams may have issues"}},
		{"unknown version", `echo "garbage"`, 12,
			[]string{"GraphViz: garbage", "Warning : cannot determine dot version"}},
		{"failing layout", `if [ "$1" = "-V" ]; then echo "` + testVersion + `" >&2; exit 0; fi
cat >/dev/null; echo "Error: something" >&2; exit 1`, 14,
			[]string{"GraphViz: " + testVersion, "Error: dot generates unreadable SVG file. Check you dot installation."}},
		{"empty layout", `if [ "$1" = "-V" ]; then echo "` + testVersion + `" >&2; exit 0; fi
cat >/dev/null`, 14,
			[]string{"GraphViz: " + testVersion, "Error: dot generates empty file. Check you dot installation."}},
		{"hanging layout", `if [ "$1" = "-V" ]; then echo "` + testVersion + `" >&2; exit 0; fi
exec sleep 30`, 14,
			[]string{"GraphViz: " + testVersion, "Error: timeout TIMEOUT"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exe := writeScript(t, tc.script)
			lines, code := AddDotStatus(context.Background(), Config{DotPath: exe, Timeout: testTimeout}, false, false)
			if code != tc.code || !slices.Equal(lines, tc.lines) {
				t.Errorf("AddDotStatus = %d %q, want %d %q", code, lines, tc.code, tc.lines)
			}
		})
	}
}

func TestAddDotStatusRichText(t *testing.T) {
	exe := writeScript(t, workingDot("x"))
	lines, code := AddDotStatus(context.Background(), Config{DotPath: exe}, false, true)
	want := []string{`<b>GraphViz</b> ""` + testVersion + `""`, "<b>Installation seems OK. File generation OK"}
	if code != 0 || !slices.Equal(lines, want) {
		t.Errorf("AddDotStatus = %d %q, want 0 %q", code, lines, want)
	}

	h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": "/missing/dot"}, nil)
	lines, code = AddDotStatus(context.Background(), Config{Host: h}, false, true)
	want = []string{"<b><color:red>Error: Dot executable does not exist", "Error: only sequence diagrams will be generated"}
	if code != 16 || !slices.Equal(lines, want) {
		t.Errorf("AddDotStatus = %d %q, want 16 %q", code, lines, want)
	}
}

func TestCheckGraphvizMissingDot(t *testing.T) {
	h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": "/missing/dot", "PATH": "/usr/bin"},
		map[string]fs.FileMode{"/usr/bin/dot": 0o755})
	lines, code := CheckGraphviz(context.Background(), Config{Host: h}, false)
	want := []string{"Error: Dot executable does not exist", "Error: only sequence diagrams will be generated"}
	if code != 16 || !slices.Equal(lines, want) {
		t.Errorf("CheckGraphviz = %d %q, want 16 %q", code, lines, want)
	}
}

func TestCheckGraphvizInsecure(t *testing.T) {
	h := fakeHost("linux", nil, nil)
	lines, code := CheckGraphviz(context.Background(), Config{Host: h}, true)
	want := []string{
		"The environment variable GRAPHVIZ_DOT has not been set",
		"Dot executable is /opt/local/bin/dot",
		"Error: Dot executable does not exist",
		"Error: only sequence diagrams will be generated",
	}
	if code != 16 || !slices.Equal(lines, want) {
		t.Errorf("CheckGraphviz = %d %q, want 16 %q", code, lines, want)
	}

	// The explicit path is reported as the GRAPHVIZ_DOT setting, as upstream.
	exe := writeScript(t, workingDot("x"))
	lines, code = CheckGraphviz(context.Background(), Config{DotPath: exe, Host: envHost(nil)}, true)
	want = []string{
		"The environment variable GRAPHVIZ_DOT has been set to " + exe,
		"Dot executable is " + exe,
		"GraphViz: " + testVersion,
		"Installation seems OK. File generation OK",
	}
	if code != 0 || !slices.Equal(lines, want) {
		t.Errorf("CheckGraphviz = %d %q, want 0 %q", code, lines, want)
	}

	// Windows without Graphviz: upstream prints the null File.
	h = fakeHost("windows", nil, nil)
	lines, _ = CheckGraphviz(context.Background(), Config{Host: h}, true)
	if lines[1] != "Dot executable is null" || lines[2] != "Error: No dot executable found" {
		t.Errorf("CheckGraphviz = %q", lines)
	}
}

// TestCheckGraphvizWithInstalledDot runs the check on the real Graphviz when
// it is installed.
func TestCheckGraphvizWithInstalledDot(t *testing.T) {
	path, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("Graphviz dot is not installed")
	}
	lines, code := CheckGraphviz(context.Background(), Config{DotPath: path}, false)
	if code != 0 || len(lines) != 2 || !strings.HasPrefix(lines[0], "GraphViz: dot - graphviz version ") || lines[1] != "Installation seems OK. File generation OK" {
		t.Errorf("CheckGraphviz = %d %q", code, lines)
	}
}
