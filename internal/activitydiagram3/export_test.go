// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// The export tests compare the SVG output with the reference renders of
// the upstream jar: the curated conformance references (committed) and,
// when they are fetched, the Vega references in .cache. The comparison is
// byte for byte, except the <?plantuml-src ...?> instruction, whose deflate
// stream differs from Java's (design D10).

const (
	curatedReferenceRoot = "../../testdata/conformance/reference"
	vegaReferenceRoot    = "../../.cache/reference/vega"
)

var plantumlSrc = regexp.MustCompile(`<\?plantuml-src [^?]*\?>`)

// referenceMeta is the part of a reference meta.json the tests read.
type referenceMeta struct {
	Blocks []struct {
		Index       int    `json:"index"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Renders     []struct {
			File *string `json:"file"`
		} `json:"renders"`
	} `json:"blocks"`
}

// activityReference is a reference SVG of block 0 of an activity case.
type activityReference struct {
	name   string
	source string
	svg    string
}

// exportExceptions are the reference pages that differ, with the reason.
var exportExceptions = map[string]string{
	"vega/tikz/tikz-1926":     "embedded diagram {{ }} in a note (EmbeddedDiagramMaker not ported)",
	"vega/tikz/tikz-1926-svg": "embedded diagram {{ }} in a note (EmbeddedDiagramMaker not ported)",
}

// activityReferences lists the SVG of block 0 of the activity cases under
// refRoot, whose sources are under casesRoot.
func activityReferences(t *testing.T, refRoot, casesRoot, prefix string) []activityReference {
	t.Helper()
	var result []activityReference
	err := filepath.WalkDir(refRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "meta.json" {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var meta referenceMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		// The legacy activity diagrams ("(n activities)") are another
		// diagram type
		if len(meta.Blocks) == 0 || meta.Blocks[0].Type != "ACTIVITY" ||
			meta.Blocks[0].Description != "activity3" || len(meta.Blocks[0].Renders) == 0 {
			return nil
		}
		r := meta.Blocks[0].Renders[0]
		if r.File == nil || !strings.HasSuffix(*r.File, ".svg") {
			return nil
		}
		dir := filepath.Dir(p)
		rel, _ := filepath.Rel(refRoot, dir)
		src, err := os.ReadFile(filepath.Join(casesRoot, rel+".puml"))
		if err != nil {
			return nil
		}
		svg, err := os.ReadFile(filepath.Join(dir, *r.File))
		if err != nil {
			return err
		}
		result = append(result, activityReference{name: prefix + filepath.ToSlash(rel),
			source: frontMatterBody(string(src)), svg: string(svg)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(result, func(a, b activityReference) int { return strings.Compare(a.name, b.name) })
	return result
}

// exportSVG draws page 0 of an activity diagram.
func exportSVG(d core.Diagram) (string, error) {
	ad, ok := d.(*ActivityDiagram3)
	if !ok {
		return "", fmt.Errorf("not an activity diagram: %T", d)
	}
	var buf bytes.Buffer
	_, err := ad.ExportDiagram(&buf, 0, export.NewFileFormatOptionMetadata(font.FormatSVG, true))
	return buf.String(), err
}

// renderSVG parses and draws a source.
func renderSVG(source string) (string, error) {
	d, preproc, err := parseCase(source)
	if err != nil {
		return "", err
	}
	if preproc {
		return "", fmt.Errorf("preprocessor error")
	}
	return exportSVG(d)
}

func sameSVG(got, want string) bool {
	return plantumlSrc.ReplaceAllString(got, "") == plantumlSrc.ReplaceAllString(want, "")
}

func checkReferences(t *testing.T, refs []activityReference) {
	identical := 0
	for _, ref := range refs {
		got, err := renderSVG(ref.source)
		reason, known := exportExceptions[ref.name]
		same := err == nil && sameSVG(got, ref.svg)
		switch {
		case same:
			identical++
			if known {
				t.Errorf("%s: identical, but listed as an exception (%s)", ref.name, reason)
			}
		case known:
		case err != nil:
			t.Errorf("%s: %v", ref.name, err)
		default:
			t.Errorf("%s: SVG differs from the reference\n%s", ref.name, firstDifference(got, ref.svg))
		}
	}
	t.Logf("%d of %d reference pages identical", identical, len(refs))
}

// firstDifference shows the SVG around the first differing byte.
func firstDifference(got, want string) string {
	got = plantumlSrc.ReplaceAllString(got, "")
	want = plantumlSrc.ReplaceAllString(want, "")
	i := 0
	for i < len(got) && i < len(want) && got[i] == want[i] {
		i++
	}
	from := max(0, i-150)
	return fmt.Sprintf("got:  ...%.400s\nwant: ...%.400s", got[from:], want[from:])
}

func TestExportCuratedReferences(t *testing.T) {
	refs := activityReferences(t, curatedReferenceRoot, conformanceRoot, "")
	if len(refs) == 0 {
		t.Fatal("no curated activity reference")
	}
	checkReferences(t, refs)
}

func TestExportVegaReferences(t *testing.T) {
	if _, err := os.Stat(vegaReferenceRoot); err != nil {
		t.Skip("Vega references not fetched (make reference)")
	}
	checkReferences(t, activityReferences(t, vegaReferenceRoot, vegaRoot, "vega/"))
}
