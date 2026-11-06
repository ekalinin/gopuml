// SPDX-License-Identifier: MIT
// Provenance: original

package directdot_test

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/directdot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

const graph = "@startdot\ndigraph G {\na -> b\n}\n@enddot\n"

// renderContext returns a context in a security profile, which lays graphs
// out with dotPath, or with the built-in engine when it is "".
func renderContext(t *testing.T, profile, dotPath string) *core.RenderContext {
	t.Helper()
	t.Setenv("GRAPHVIZ_DOT", "")
	rc := core.NewRenderContext()
	rc.SecurityProfile = profile
	rc.DotPath = dotPath
	return rc
}

// exportDot parses the first block of src as a dot diagram and writes it in
// format, with metadata.
func exportDot(t *testing.T, rc *core.RenderContext, src string, format font.FileFormat) ([]byte, export.ImageData) {
	t.Helper()
	d, ok := parseDot(t, rc, src).(*directdot.PSystemDot)
	if !ok {
		t.Fatalf("%q is not a dot diagram", src)
	}
	var buf bytes.Buffer
	data, err := d.ExportDiagramContext(t.Context(), &buf, 0, export.NewFileFormatOptionMetadata(format, true))
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), data
}

// fakeDot writes a dot executable that runs script, and returns its path.
func fakeDot(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake dot is a shell script")
	}
	exe := filepath.Join(t.TempDir(), "dot")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

// TestDotBuiltinEngine checks the dot-diagram scenario "Laid out by the
// built-in engine".
func TestDotBuiltinEngine(t *testing.T) {
	out, data := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", ""), graph, font.FormatSVG)
	if s := string(out); !strings.HasPrefix(s, "<?xml") || !strings.Contains(s, "(gogv)") || !strings.Contains(s, ">a</text>") {
		t.Errorf("output %.300q, want the SVG of gogv with the node a", out)
	}
	if conformance.HasSVGSourceMetadata(out) || data.Status() != 0 {
		t.Errorf("metadata %v, status %d, want neither", conformance.HasSVGSourceMetadata(out), data.Status())
	}
}

// TestDotNamedDot checks the dot-diagram scenario "Laid out by the named
// dot".
func TestDotNamedDot(t *testing.T) {
	args := filepath.Join(t.TempDir(), "args")
	exe := fakeDot(t, `echo "$@" > '`+args+"'\ncat >/dev/null\necho '<svg>fake</svg>'\n")
	out, data := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", exe), graph, font.FormatSVG)
	if string(out) != "<svg>fake</svg>\n" || data.Status() != 0 {
		t.Errorf("output %q, status %d, want the output of the named dot", out, data.Status())
	}
	if got, _ := os.ReadFile(args); string(got) != "-Tsvg\n" {
		t.Errorf("arguments %q, want -Tsvg", got)
	}
}

// TestDotSyntaxErrorOutput checks the dot-diagram scenario "Syntax error in
// the DOT source".
func TestDotSyntaxErrorOutput(t *testing.T) {
	out, data := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", ""), "@startdot\ndigraph G {\na ->\n}\n@enddot\n", font.FormatSVG)
	if string(out) != "Error: <stdin>: syntax error in line 3 near '}'\n" || data.Status() != 0 {
		t.Errorf("output %q, status %d, want the error line of dot and status 0", out, data.Status())
	}
}

// TestDotGraphvizWarningOutput checks the dot-diagram scenario "Graphviz
// warning in the output".
func TestDotGraphvizWarningOutput(t *testing.T) {
	out, _ := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", ""), "@startdot\ndigraph G {\na -> 1b\n}\n@enddot\n", font.FormatSVG)
	const want = "Warning: syntax ambiguity - badly delimited number '1b' in line 2 of <stdin> splits into two tokens\n<?xml"
	if !strings.HasPrefix(string(out), want) {
		t.Errorf("output %.200q, want the warning line, then the SVG document", out)
	}
}

// TestDotNotSupportedYetWarningDropped checks the dot-diagram scenario
// "Feature not supported by the built-in engine".
func TestDotNotSupportedYetWarningDropped(t *testing.T) {
	out, _ := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", ""), "@startdot\ndigraph G {\na [shape=record]\n}\n@enddot\n", font.FormatSVG)
	if !strings.HasPrefix(string(out), "<?xml") || strings.Contains(string(out), "Warning") {
		t.Errorf("output %.200q, want the SVG document without a warning", out)
	}
}

// TestDotPNGNamedDot checks the dot-diagram scenario "PNG with the named
// dot".
func TestDotPNGNamedDot(t *testing.T) {
	args := filepath.Join(t.TempDir(), "args")
	exe := fakeDot(t, `echo "$@" > '`+args+"'\ncat >/dev/null\nprintf 'PNG bytes'\n")
	out, data := exportDot(t, renderContext(t, "LEGACY", exe), graph, font.FormatPNG)
	if string(out) != "PNG bytes" || data.Status() != 0 {
		t.Errorf("output %q, status %d, want the output of the named dot", out, data.Status())
	}
	if got, _ := os.ReadFile(args); string(got) != "-Tpng\n" {
		t.Errorf("arguments %q, want -Tpng", got)
	}
}

// TestDotMissingNamedDot checks the dot-diagram scenario "Named dot is
// missing".
func TestDotMissingNamedDot(t *testing.T) {
	out, data := exportDot(t, renderContext(t, "INTERNET_WITH_DOTSVG", "/missing/dot"), graph, font.FormatSVG)
	if !strings.Contains(string(out), ">There is an issue with your Dot/Graphviz installation</text>") || data.Status() != 503 {
		t.Errorf("output %.300q, status %d, want the installation error image and status 503", out, data.Status())
	}
	if conformance.HasSVGSourceMetadata(out) {
		t.Error("the error image has the source metadata")
	}
}

// TestDotLayoutTimeout checks the dot-diagram scenario "Layout timeout in a
// dot diagram".
func TestDotLayoutTimeout(t *testing.T) {
	rc := renderContext(t, "INTERNET_WITH_DOTSVG", fakeDot(t, "exec sleep 30\n"))
	rc.DotTimeout = time.Second
	start := time.Now()
	out, data := exportDot(t, rc, graph, font.FormatSVG)
	if !strings.Contains(string(out), ">GraphViz has crashed</text>") || data.Status() != 503 {
		t.Errorf("output %.300q, status %d, want the crash image and status 503", out, data.Status())
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("the layout stopped after %s, want about 1s", elapsed)
	}
}

func TestDotSuppressedSVG(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "conformance", "reference", "dot-diagram", "svg-suppressed", "block0-page0.svg"))
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "ran")
	exe := fakeDot(t, "touch '"+marker+"'\ncat >/dev/null\necho '<svg/>'\n")
	for _, profile := range []string{"LEGACY", "INTERNET", "ALLOWLIST", "SANDBOX"} {
		for _, src := range []string{graph, "@startdot\ndigraph G {\na ->\n}\n@enddot\n"} {
			out, data := exportDot(t, renderContext(t, profile, exe), src, font.FormatSVG)
			if !bytes.Equal(out, want) || data.Status() != 0 {
				t.Errorf("%s, %q: output %q, status %d, want upstream's suppressed-feature document", profile, src, out, data.Status())
			}
		}
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("the layout engine ran")
	}
}

func TestDotPNGWithoutNamedDot(t *testing.T) {
	out, data := exportDot(t, renderContext(t, "LEGACY", ""), graph, font.FormatPNG)
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	// Upstream's "GraphViz has crashed" image is 181 by 26 pixels.
	if b := img.Bounds(); b.Dx() != 181 || b.Dy() != 26 || data.Status() != 503 {
		t.Errorf("image of %dx%d, status %d, want the crash image of 181x26 and status 503", b.Dx(), b.Dy(), data.Status())
	}
	if conformance.HasPNGSourceMetadata(out) {
		t.Error("the crash image has the source metadata")
	}
}
