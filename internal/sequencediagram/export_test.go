// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/svg/svgtest"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/tim"
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
		Index   int    `json:"index"`
		Type    string `json:"type"`
		Renders []struct {
			File *string `json:"file"`
		} `json:"renders"`
	} `json:"blocks"`
}

// sequenceReference is a reference SVG of block 0 of a sequence case.
type sequenceReference struct {
	name   string
	file   string
	source string
	page   int
	svg    string
}

// exportExceptions are the reference pages that differ, with the reason.
var exportExceptions = map[string]string{
	"gantt-diagram/embedded-in-note/0":             "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"mindmap-diagram/embedded-in-note/0":           "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"wbs-diagram/embedded-in-note/0":               "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"timing-diagram/embedded-in-note/0":            "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"salt-diagram/embedded-in-note/0":              "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"salt-diagram/old-embedded-form/0":             "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"chen-eer-diagram/embedded-in-note/0":          "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"packet-diagram/embedded-in-note/0":            "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"styling/embedded-use-case-in-note/0":          "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"preprocessor/option-unknown/0":                "!option warnings do not reach the diagram (core.Factory has no preprocessing artifact)",
	"styling/theme-cerulean/0":                     "themes are not bundled yet (tim.Options.Themes is never set)",
	"vega/svg/option/sequence_desc/0":              "!option svgDesc does not reach the diagram (core.Factory has no preprocessing artifact)",
	"vega/svg/option/sequence_title/0":             "!option svgTitle does not reach the diagram (core.Factory has no preprocessing artifact)",
	"vega/svg/option/sequence_title_title/0":       "!option svgTitle does not reach the diagram (core.Factory has no preprocessing artifact)",
	"vega/nonreg/warnings/option_novalue/0":        "!option warnings do not reach the diagram (core.Factory has no preprocessing artifact)",
	"vega/nonreg/warnings/option_unknown/0":        "!option warnings do not reach the diagram (core.Factory has no preprocessing artifact)",
	"vega/nonreg/warnings/skinparam_handwritten/0": "handwritten drawing is not ported (UGraphicHandwritten)",
	"vega/tikz/tikz-scale-down/0":                  "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"vega/tikz/tikz-scale-up/0":                    "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
	"vega/tikz/tikz-scale-nested-scale/0":          "embedded diagram {{ }}: built without the embedded-diagram renderer of internal/engine; the conformance suite compares it",
}

// sequenceReferences lists the SVG pages of block 0 of the sequence cases
// under refRoot, whose sources are under casesRoot.
func sequenceReferences(t *testing.T, refRoot, casesRoot, prefix string) []sequenceReference {
	t.Helper()
	var result []sequenceReference
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
		if len(meta.Blocks) == 0 || meta.Blocks[0].Type != "SEQUENCE" {
			return nil
		}
		dir := filepath.Dir(p)
		rel, _ := filepath.Rel(refRoot, dir)
		file := filepath.Join(casesRoot, rel+".puml")
		src, err := os.ReadFile(file)
		if err != nil {
			return nil
		}
		for _, r := range meta.Blocks[0].Renders {
			if r.File == nil || !strings.HasSuffix(*r.File, ".svg") {
				continue
			}
			// The page comes from the file name: renders of other formats
			// may come first.
			var block, page int
			if _, err := fmt.Sscanf(*r.File, "block%d-page%d.svg", &block, &page); err != nil {
				return fmt.Errorf("%s: %w", *r.File, err)
			}
			svg, err := os.ReadFile(filepath.Join(dir, *r.File))
			if err != nil {
				return err
			}
			result = append(result, sequenceReference{name: prefix + filepath.ToSlash(rel), file: file,
				source: frontMatterBody(string(src)), page: page, svg: string(svg)})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(result, func(a, b sequenceReference) int {
		return strings.Compare(fmt.Sprint(a.name, a.page), fmt.Sprint(b.name, b.page))
	})
	return result
}

// parseFile builds the diagram of the first block of a source file, as
// the engine does: the includes are relative to the file.
func parseFile(file, source string) (core.Diagram, error) {
	rc := core.NewRenderContext()
	rc.CurrentDir = filepath.Dir(file)
	// The fixtures of a curated case are staged next to it by refgen
	if fixtures := strings.TrimSuffix(file, ".puml") + ".files"; isDir(fixtures) {
		rc.CurrentDir = fixtures
	}
	blocks := core.BuildBlocks(source, rc.Desc, rc.Config)
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no block")
	}
	access := &security.Access{FS: rc.FS}
	access.Config.Profile, _ = security.ParseProfile(rc.SecurityProfile)
	modTime := time.UnixMilli(0)
	if st, err := os.Stat(file); err == nil {
		modTime = st.ModTime()
	}
	res, err := tim.PreprocessWith(rc, access, blocks[0].Lines, tim.Options{Definitions: &tim.BlockDefinitions{},
		Defines: tim.DefinesWithFileName(filepath.Base(file), modTime, "")})
	if err != nil {
		return nil, err
	}
	return Factory.CreateSystem(rc, core.NewUmlSource(res.Lines, blocks[0].Lines, true)), nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func renderSVG(t *testing.T, source string, page int) (string, error) {
	t.Helper()
	d, _, err := parseCase(source)
	if err != nil {
		return "", err
	}
	return exportSVG(d, page)
}

func exportSVG(d core.Diagram, page int) (string, error) {
	sd, ok := d.(*SequenceDiagram)
	if !ok {
		return "", fmt.Errorf("not a sequence diagram: %T", d)
	}
	var buf bytes.Buffer
	_, err := sd.ExportDiagram(&buf, page, export.NewFileFormatOptionMetadata(font.FormatSVG, true))
	return buf.String(), err
}

func checkReferences(t *testing.T, refs []sequenceReference) {
	identical := 0
	for _, ref := range refs {
		key := fmt.Sprintf("%s/%d", ref.name, ref.page)
		d, err := parseFile(ref.file, ref.source)
		var got string
		if err == nil {
			got, err = exportSVG(d, ref.page)
		}
		reason, known := exportExceptions[key]
		// PNG data URLs compare by their pixels: Go's PNG encoder differs
		// from Java's ImageIO (internal/klimt/svg).
		same := err == nil && svgtest.Normalize(plantumlSrc.ReplaceAllString(got, "")) ==
			svgtest.Normalize(plantumlSrc.ReplaceAllString(ref.svg, ""))
		switch {
		case same:
			identical++
			if known {
				t.Errorf("%s: identical, but listed as an exception (%s)", key, reason)
			}
		case known:
		case err != nil:
			t.Errorf("%s: %v", key, err)
		default:
			t.Errorf("%s: SVG differs from the reference\ngot:  %.400s\nwant: %.400s", key, got, ref.svg)
		}
	}
	t.Logf("%d of %d reference pages identical", identical, len(refs))
}

// TestExportCuratedReferences requires the SVG of every curated sequence
// case to be byte-identical to its reference. The references are rendered
// on Linux and the test runs on macOS and Linux: the output-formats
// scenario "Same output on two machines".
func TestExportCuratedReferences(t *testing.T) {
	refs := sequenceReferences(t, curatedReferenceRoot, conformanceRoot, "")
	if len(refs) == 0 {
		t.Fatal("no curated sequence reference")
	}
	checkReferences(t, refs)
}

func TestExportVegaReferences(t *testing.T) {
	if _, err := os.Stat(vegaReferenceRoot); err != nil {
		t.Skip("Vega references not fetched (make reference)")
	}
	checkReferences(t, sequenceReferences(t, vegaReferenceRoot, vegaRoot, "vega/"))
}

func TestExportPages(t *testing.T) {
	d := parseModel(t, "A -> B : one\nnewpage second\nA -> B : two")
	if d.Pages() != 2 {
		t.Fatalf("Pages = %d", d.Pages())
	}
	for page := range 2 {
		var buf bytes.Buffer
		data, err := d.ExportDiagram(&buf, page, export.NewFileFormatOption(font.FormatSVG))
		if err != nil || data.Width() == 0 || !strings.Contains(buf.String(), "<svg") {
			t.Fatalf("page %d: %v", page, err)
		}
		if page == 1 && !strings.Contains(buf.String(), ">second</text>") {
			t.Errorf("page 1 has no title")
		}
	}
}

// sourceLine matches the source line attribute of a group.
var sourceLine = regexp.MustCompile(` data-source-line="\d+"`)

func TestExportTeozPragmaHasNoEffect(t *testing.T) {
	render := func(body string) string {
		got, err := renderSVG(t, "@startuml\n"+body+"Alice -> Bob : a\n& Bob -> Carol : b\n@enduml\n", 0)
		if err != nil {
			t.Fatal(err)
		}
		// The pragma line shifts the source lines of the groups.
		got = sourceLine.ReplaceAllString(got, "")
		return strings.Join(strings.Split(plantumlSrc.ReplaceAllString(got, ""), "\n"), "")
	}
	base := render("")
	for _, pragma := range []string{"!pragma teoz false\n", "!pragma teoz true\n", "!pragma teoz anything\n"} {
		if got := render(pragma); got != base {
			t.Errorf("%q changes the drawing", pragma)
		}
	}
}

func TestExportCrashIsAnError(t *testing.T) {
	// Upstream also reports "Infinite Loop?" (as a crash report) for this
	// cyclic set of constraints.
	_, err := renderSVG(t, "@startuml\ncollections P4\nP5 --> P5 : self 0\n& P0 -> P4 : par 1\n@enduml\n", 0)
	if err == nil || !strings.Contains(err.Error(), "Infinite Loop?") {
		t.Errorf("err = %v", err)
	}
}

// TestExportConcurrent renders diagrams in parallel, each parsed by its
// own goroutine (a diagram, like upstream's, is not safe for concurrent
// use: its skin parameters cache their values).
func TestExportConcurrent(t *testing.T) {
	const body = "box \"B\"\nparticipant A\nend box\nA -> B ++ : call\nB -> B : self\nnote right : n\n" +
		"alt x\nB --> A : r\nend\nreturn ok"
	var want string
	var wg sync.WaitGroup
	results := make([]string, 8)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := renderSVG(t, "@startuml\n"+body+"\n@enduml\n", 0)
			if err != nil {
				t.Error(err)
			}
			results[i] = got
		}()
	}
	wg.Wait()
	want = results[0]
	for i, r := range results {
		if r != want {
			t.Errorf("export %d differs", i)
		}
	}
}

// TestSVGElementGroups checks the groups of the lifelines, heads, tails and
// messages: the start tags of the groups of each page are those of the
// reference written by upstream, for sequence-diagram/svg-element-groups and
// every other committed sequence reference but the exceptions (scenario
// "Groups of a message").
func TestSVGElementGroups(t *testing.T) {
	refs := sequenceReferences(t, curatedReferenceRoot, conformanceRoot, "")
	groups := regexp.MustCompile(`<g[ >][^>]*>`)
	seen := false
	for _, ref := range refs {
		if _, known := exportExceptions[fmt.Sprintf("%s/%d", ref.name, ref.page)]; known {
			continue
		}
		d, err := parseFile(ref.file, ref.source)
		if err != nil {
			t.Errorf("%s: %v", ref.name, err)
			continue
		}
		got, err := exportSVG(d, ref.page)
		if err != nil {
			t.Errorf("%s/%d: %v", ref.name, ref.page, err)
			continue
		}
		want := groups.FindAllString(ref.svg, -1)
		if g := groups.FindAllString(got, -1); !slices.Equal(g, want) {
			t.Errorf("%s/%d: groups\n got  %q\n want %q", ref.name, ref.page, g, want)
		}
		if strings.HasSuffix(ref.name, "svg-element-groups") {
			seen = true
		}
	}
	if !seen {
		t.Error("no reference for sequence-diagram/svg-element-groups")
	}
}
