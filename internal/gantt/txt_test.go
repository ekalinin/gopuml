// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
)

const casesDir = "../../testdata/conformance/cases"

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

func parseSource(t *testing.T, src string) *GanttDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	d, ok := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true)).(*GanttDiagram)
	if !ok {
		t.Fatalf("not a gantt diagram")
	}
	return d
}

func txtOf(t *testing.T, d *GanttDiagram) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.exportTxt(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// TestTxtReferences checks the TXT and UTXT references of the curated
// cases: the table of tasks, in ASCII for both.
func TestTxtReferences(t *testing.T) {
	for _, id := range []string{"gantt-diagram/utxt-table", "gantt-diagram/table-rows", "gantt-diagram/table-french",
		"gantt-diagram/table-no-project-start", "gantt-diagram/table-working-hours", "output-formats/gantt-txt"} {
		d := parseSource(t, caseSource(t, filepath.Join(casesDir, id+".puml")))
		matches, _ := filepath.Glob(filepath.Join("../../testdata/conformance/reference", id, "block0-page0.*txt"))
		if len(matches) != 1 {
			t.Fatalf("%s: references %v", id, matches)
		}
		want, err := os.ReadFile(matches[0])
		if err != nil {
			t.Fatal(err)
		}
		if got := txtOf(t, d); got != string(want) {
			t.Errorf("%s:\n%s\nwant\n%s", id, got, want)
		}
	}
}

// TestTxtOfCuratedCases checks the dates of every curated gantt case that
// draws a diagram against the TXT output of the reference jar for the same
// source (testdata/txt), before the drawing.
func TestTxtOfCuratedCases(t *testing.T) {
	goldens, err := filepath.Glob("testdata/txt/*.atxt")
	if err != nil || len(goldens) == 0 {
		t.Fatalf("no golden: %v", err)
	}
	for _, golden := range goldens {
		id := strings.TrimSuffix(filepath.Base(golden), ".atxt")
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		d := parseSource(t, caseSource(t, filepath.Join(casesDir, "gantt-diagram", id+".puml")))
		if got := txtOf(t, d); got != string(want) {
			t.Errorf("%s:\n%s\nwant\n%s", id, got, want)
		}
	}
}
