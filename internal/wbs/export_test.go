// SPDX-License-Identifier: MIT
// Provenance: original

package wbs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
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

func parseSource(t *testing.T, src string) *WBSDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	d, ok := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true)).(*WBSDiagram)
	if !ok {
		t.Fatalf("not a WBS diagram")
	}
	return d
}

func svgOf(t *testing.T, d *WBSDiagram) (string, error) {
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

// TestExportCuratedReferences requires the SVG of every curated WBS case
// that draws a WBS diagram to be the reference, apart from the encoding of
// the source.
func TestExportCuratedReferences(t *testing.T) {
	var cases []string
	for _, pattern := range []string{"wbs-diagram/*.puml", "diagram-source/wbs-*.puml"} {
		found, err := filepath.Glob(filepath.Join(casesDir, pattern))
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, found...)
	}
	identical, total := 0, 0
	for _, c := range cases {
		id := filepath.Join(filepath.Base(filepath.Dir(c)), strings.TrimSuffix(filepath.Base(c), ".puml"))
		refDir := filepath.Join(referencesDir, id)
		meta, err := os.ReadFile(filepath.Join(refDir, "meta.json"))
		if err != nil || !strings.Contains(string(meta), `"type": "WBS"`) {
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
