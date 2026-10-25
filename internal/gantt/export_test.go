// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

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

const referencesDir = "../../testdata/conformance/reference"

func svgOf(t *testing.T, d *GanttDiagram) (string, error) {
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

// TestExportCuratedReferences requires the SVG of every curated gantt case
// that draws a diagram to be the reference, apart from the encoding of the
// source.
func TestExportCuratedReferences(t *testing.T) {
	cases, err := filepath.Glob(filepath.Join(casesDir, "gantt-diagram", "*.puml"))
	if err != nil {
		t.Fatal(err)
	}
	identical, total := 0, 0
	for _, c := range cases {
		id := strings.TrimSuffix(filepath.Base(c), ".puml")
		refDir := filepath.Join(referencesDir, "gantt-diagram", id)
		meta, err := os.ReadFile(filepath.Join(refDir, "meta.json"))
		if err != nil || !strings.Contains(string(meta), `"type": "GANTT"`) {
			continue
		}
		want, err := os.ReadFile(filepath.Join(refDir, "block0-page0.svg"))
		if err != nil {
			continue
		}
		total++
		got, err := svgOf(t, parseSource(t, caseSource(t, c)))
		if err != nil {
			t.Errorf("%s: %v", id, err)
			continue
		}
		if svgtest.Normalize(got) == svgtest.Normalize(string(want)) {
			identical++
			continue
		}
		g, w := firstDifference(svgtest.Normalize(got), svgtest.Normalize(string(want)))
		t.Errorf("%s: SVG differs\n got: %s\nwant: %s", id, g, w)
	}
	t.Logf("%d of %d references identical", identical, total)
}
