// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// The text output tests compare ATXT and UTXT with the reference renders
// of the upstream jar, byte for byte, as the SVG tests do.

// textReference is a reference text page of block 0 of a sequence case.
type textReference struct {
	name   string
	file   string
	source string
	page   int
	format font.FileFormat
	text   string
}

// textReferences lists the .atxt and .utxt pages of block 0 of the sequence
// cases under refRoot, whose sources are under casesRoot.
func textReferences(t *testing.T, refRoot, casesRoot, prefix string) []textReference {
	t.Helper()
	var result []textReference
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
		pages := map[font.FileFormat]int{}
		for _, r := range meta.Blocks[0].Renders {
			if r.File == nil {
				continue
			}
			var format font.FileFormat
			switch filepath.Ext(*r.File) {
			case ".atxt":
				format = font.FormatATXT
			case ".utxt":
				format = font.FormatUTXT
			default:
				continue
			}
			text, err := os.ReadFile(filepath.Join(dir, *r.File))
			if err != nil {
				return err
			}
			result = append(result, textReference{name: prefix + filepath.ToSlash(rel), file: file,
				source: frontMatterBody(string(src)), page: pages[format], format: format, text: string(text)})
			pages[format]++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(result, func(a, b textReference) int {
		return strings.Compare(fmt.Sprint(a.name, a.format, a.page), fmt.Sprint(b.name, b.format, b.page))
	})
	return result
}

func exportText(d core.Diagram, page int, format font.FileFormat) (string, export.ImageData, error) {
	sd, ok := d.(*SequenceDiagram)
	if !ok {
		return "", nil, fmt.Errorf("not a sequence diagram: %T", d)
	}
	var buf bytes.Buffer
	data, err := sd.ExportDiagram(&buf, page, export.NewFileFormatOption(format))
	return buf.String(), data, err
}

func renderText(t *testing.T, source string, format font.FileFormat) (string, error) {
	t.Helper()
	d, _, err := parseCase(source)
	if err != nil {
		return "", err
	}
	got, _, err := exportText(d, 0, format)
	return got, err
}

func checkTextReferences(t *testing.T, refs []textReference) {
	identical := 0
	for _, ref := range refs {
		key := fmt.Sprintf("%s/%s/%d", ref.name, ref.format, ref.page)
		d, err := parseFile(ref.file, ref.source)
		var got string
		if err == nil {
			got, _, err = exportText(d, ref.page, ref.format)
		}
		switch {
		case err != nil:
			t.Errorf("%s: %v", key, err)
		case got != ref.text:
			t.Errorf("%s: text differs from the reference\ngot:\n%s\nwant:\n%s", key, got, ref.text)
		default:
			identical++
		}
	}
	t.Logf("%d of %d reference text pages identical", identical, len(refs))
}

func TestExportTextCuratedReferences(t *testing.T) {
	refs := textReferences(t, curatedReferenceRoot, conformanceRoot, "")
	if len(refs) == 0 {
		t.Fatal("no curated sequence text reference")
	}
	checkTextReferences(t, refs)
}

func TestExportTextVegaReferences(t *testing.T) {
	if _, err := os.Stat(vegaReferenceRoot); err != nil {
		t.Skip("Vega references not fetched (make reference)")
	}
	checkTextReferences(t, textReferences(t, vegaReferenceRoot, vegaRoot, "vega/"))
}

// The expected texts below are the reference jar's outputs.

func TestExportTextGroupsAndNotes(t *testing.T) {
	src := "@startuml\nAlice -> Bob : hello\nnote right: a note\nalt\nBob -> Alice : x\nend\n" +
		"group\nAlice -> Alice\nend\n@enduml\n"
	want := map[font.FileFormat]string{
		font.FormatATXT: "" +
			",-----.   ,---.\n" +
			"|Alice|   |Bob|\n" +
			"`--+--'   `-+-'\n" +
			"   | hello  |   ,--------!. \n" +
			"   |------->|   | a note |_\\\n" +
			"   |        |   `----------'\n" +
			" ,------------.\n" +
			" ! alt null / !\n" +
			" !_________/| !\n" +
			" ! |   x    | !\n" +
			" ! |<-------| !\n" +
			" ! |        | !\n" +
			" ~~~~~~~~~~~~~~\n" +
			" ,--------. |\n" +
			" ! group /! |\n" +
			" !______/ ! |\n" +
			" ! |----. ! |\n" +
			" ! |    | ! |\n" +
			" ! |<---' ! |\n" +
			" ~~~~~~~~~~ |\n" +
			",--+--.   ,-+-.\n" +
			"|Alice|   |Bob|\n" +
			"`-----'   `---'\n",
		font.FormatUTXT: "" +
			"┌─────┐   ┌───┐\n" +
			"│Alice│   │Bob│\n" +
			"└──┬──┘   └─┬─┘\n" +
			"   │ hello  │   ╔══════════╗\n" +
			"   │───────>│   ║ a note  ░║\n" +
			"   │        │   ╚══════════╝\n" +
			" ╔══════════╤═╗\n" +
			" ║ alt null │ ║\n" +
			" ╟──────────┘ ║\n" +
			" ║ │   x    │ ║\n" +
			" ║ │<───────│ ║\n" +
			" ║ │        │ ║\n" +
			" ╚══════════╧═╝\n" +
			" ╔═══════╤╗ │\n" +
			" ║ group │║ │\n" +
			" ╟───────┘║ │\n" +
			" ║ │────┐│║ │\n" +
			" ║ │    ││║ │\n" +
			" ║ │<───┘│║ │\n" +
			" ╚═══════╧╝ │\n" +
			"┌──┴──┐   ┌─┴─┐\n" +
			"│Alice│   │Bob│\n" +
			"└─────┘   └───┘\n",
	}
	for format, w := range want {
		got, err := renderText(t, src, format)
		if err != nil {
			t.Fatal(err)
		}
		if got != w {
			t.Errorf("%s:\ngot:\n%s\nwant:\n%s", format, got, w)
		}
	}
}

func TestExportTextTitleIsCentered(t *testing.T) {
	// The drawing reports a fixed width of 20 (upstream's placeholder), so
	// a short title is centered on 20 columns
	got, err := renderText(t, "@startuml\ntitle T\nAlice -> Bob : hello\n@enduml\n", font.FormatATXT)
	if err != nil {
		t.Fatal(err)
	}
	if want := "         T\n\n,-----.   ,---.\n"; !strings.HasPrefix(got, want) {
		t.Errorf("got:\n%s\nwant prefix:\n%s", got, want)
	}
}

func TestExportTextEveryPageIsTheWholeDiagram(t *testing.T) {
	d := parseModel(t, "Alice -> Bob : hello1\nnewpage\nAlice -> Bob : hello2")
	page0, data, err := exportText(d, 0, font.FormatUTXT)
	if err != nil {
		t.Fatal(err)
	}
	if data.Width() != 0 || data.Height() != 0 || data.Status() != 0 {
		t.Errorf("image data = %dx%d status %d, want upstream's ImageDataSimple.ok()", data.Width(), data.Height(),
			data.Status())
	}
	page1, _, err := exportText(d, 1, font.FormatUTXT)
	if err != nil {
		t.Fatal(err)
	}
	if page0 != page1 || !strings.Contains(page0, "hello1") || !strings.Contains(page0, "hello2") {
		t.Errorf("pages differ or are incomplete:\n%s\n%s", page0, page1)
	}
	if !strings.Contains(page0, "─ ─ ─│─ ─ ─ ─ ─│─ ─") {
		t.Errorf("no page break line:\n%s", page0)
	}
}

func TestExportTextUnsupportedElementIsAnError(t *testing.T) {
	// Upstream throws UnsupportedOperationException (and writes a crash
	// report) for the tiles without text drawing
	for src, class := range map[string]string{
		"Alice -> Bob\n...\nBob -> Alice": "DelayTile",
		"Alice -> Bob\n== d ==":           "DividerTile",
		"note over Alice : n":             "NoteTile",
		"ref over Alice : r":              "ReferenceTile",
		"Alice -> Bob\n|||":               "HSpaceTile",
		"Alice -> Bob\nnote top : n":      "CommunicationTileNoteTop",
		"Alice -> Bob\nnote bottom : n":   "CommunicationTileNoteBottom",
	} {
		d := parseModel(t, src)
		got, data, err := exportText(d, 0, font.FormatATXT)
		want := "java.lang.UnsupportedOperationException: net.sourceforge.plantuml.sequencediagram.teoz." + class
		if err == nil || !strings.Contains(err.Error(), want) || data != nil || got != "" {
			t.Errorf("%q: err = %v, output %q, want an error containing %q and no output", src, err, got, want)
		}
	}
}
