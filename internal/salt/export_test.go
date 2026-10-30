// SPDX-License-Identifier: MIT
// Provenance: original

package salt

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

// exportExceptions are the references that differ, with the reason.
var exportExceptions = map[string]string{
	"salt-diagram/sprite-in-cell": "the sprite is scaled by 12/13 for the font of the widgets, and the bilinear " +
		"scaling of klimt rounds some pixels one off Java's fixed point arithmetic; the conformance suite compares it",
}

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

// parseSource returns the diagram of src, a salt diagram or an error.
func parseSource(t *testing.T, src string) core.Diagram {
	t.Helper()
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	return Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
}

func parseSalt(t *testing.T, src string) *PSystemSalt {
	t.Helper()
	d, ok := parseSource(t, src).(*PSystemSalt)
	if !ok {
		t.Fatalf("not a salt diagram: %#v", parseSource(t, src))
	}
	return d
}

func exportOf(t *testing.T, d *PSystemSalt, format font.FileFormat) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	_, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOptionMetadata(format, true))
	return buf.String(), err
}

func svgOf(t *testing.T, d *PSystemSalt) (string, error) { return exportOf(t, d, font.FormatSVG) }

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

// TestExportCuratedReferences requires the SVG of every curated case that
// draws a salt diagram to be the reference, apart from the encoding of the
// source.
func TestExportCuratedReferences(t *testing.T) {
	var cases []string
	for _, pattern := range []string{"salt-diagram/*.puml", "diagram-source/salt-*.puml"} {
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
		if err != nil || !strings.Contains(string(meta), `"type": "SALT"`) {
			continue
		}
		want, err := os.ReadFile(filepath.Join(refDir, "block0-page0.svg"))
		if err != nil {
			continue
		}
		total++
		got, err := svgOf(t, parseSalt(t, caseSource(t, c)))
		if err != nil {
			t.Errorf("%s: %v", id, err)
			continue
		}
		if svgtest.Normalize(got) == svgtest.Normalize(string(want)) {
			identical++
			if _, ok := exportExceptions[id]; ok {
				t.Errorf("%s: matches its reference, remove it from exportExceptions", id)
			}
			continue
		}
		if _, ok := exportExceptions[id]; ok {
			continue
		}
		g, w := firstDifference(svgtest.Normalize(got), svgtest.Normalize(string(want)))
		t.Errorf("%s: SVG differs\n got: %s\nwant: %s", id, g, w)
	}
	t.Logf("%d of %d references identical", identical, total)
}
