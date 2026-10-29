// SPDX-License-Identifier: MIT
// Provenance: original

package timing

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/svg/svgtest"
)

const (
	casesDir      = "../../testdata/conformance/cases"
	referencesDir = "../../testdata/conformance/reference"
)

// caseSource returns the diagram source of a curated case, without its
// front matter.
func caseSource(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(data), "---\n", 3)
	return parts[2]
}

func svgOf(t *testing.T, d *TimingDiagram) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	_, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOptionMetadata(font.FormatSVG, true))
	return buf.String(), err
}

// firstDifference returns the parts of a and b around their first
// difference.
func firstDifference(a, b string) (string, string) {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	from := max(0, i-120)
	return a[from:min(len(a), i+200)], b[from:min(len(b), i+200)]
}

// compareCase compares the SVG of a curated case that draws a timing
// diagram with its reference; it reports false when the case does not draw
// one.
func compareCase(t *testing.T, id string) bool {
	t.Helper()
	refDir := filepath.Join(referencesDir, id)
	meta, err := os.ReadFile(filepath.Join(refDir, "meta.json"))
	if err != nil || !strings.Contains(string(meta), `"type": "TIMING"`) {
		return false
	}
	want, err := os.ReadFile(filepath.Join(refDir, "block0-page0.svg"))
	if err != nil {
		return false
	}
	d := mustParse(t, caseSource(t, filepath.Join(casesDir, id+".puml")))
	got, err := svgOf(t, d)
	if err != nil {
		t.Errorf("%s: %v", id, err)
		return true
	}
	if svgtest.Normalize(got) != svgtest.Normalize(string(want)) {
		g, w := firstDifference(svgtest.Normalize(got), svgtest.Normalize(string(want)))
		t.Errorf("%s: SVG differs\n got: %s\nwant: %s", id, g, w)
	}
	return true
}

// TestExportCuratedReferences requires the SVG of every curated case that
// draws a timing diagram to be the reference, apart from the encoding of
// the source.
func TestExportCuratedReferences(t *testing.T) {
	var cases []string
	for _, pattern := range []string{"timing-diagram/*.puml", "diagram-source/timing-*.puml"} {
		found, err := filepath.Glob(filepath.Join(casesDir, pattern))
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, found...)
	}
	total := 0
	for _, c := range cases {
		id := filepath.Join(filepath.Base(filepath.Dir(c)), strings.TrimSuffix(filepath.Base(c), ".puml"))
		if compareCase(t, id) {
			total++
		}
	}
	t.Logf("%d references compared", total)
}

// TestExportDrawingFailure checks that a diagram whose times are all below
// zero, which upstream's drawing crashes on, returns an error and writes
// nothing.
func TestExportDrawingFailure(t *testing.T) {
	d := mustParse(t, source(`concise "User" as U`, `@-20`, `U is Idle`, `@-10`, `U is Busy`))
	for _, format := range []font.FileFormat{font.FormatSVG, font.FormatPNG} {
		var buf bytes.Buffer
		_, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOptionMetadata(format, true))
		if err == nil || !strings.Contains(err.Error(), "java.lang.IllegalStateException") {
			t.Errorf("%v: error %v, want the IllegalStateException of the ruler", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written, want none", format, buf.Len())
		}
	}
}
