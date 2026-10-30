// SPDX-License-Identifier: MIT
// Provenance: original

package salt_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/salt"
)

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*salt.PSystemSalt); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

func render(t *testing.T, block *engine.Block, format font.FileFormat) ([]byte, error) {
	t.Helper()
	var buf bytes.Buffer
	_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
	return buf.Bytes(), err
}

// TestSaltWithoutDot checks the salt-diagram scenario "Salt diagram without
// Graphviz".
func TestSaltWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, "@startsalt\n{\nLogin | \"me\"\n[OK]\n}\n@endsalt\n")
	data, err := render(t, block, font.FormatSVG)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(">OK</text>")) {
		t.Errorf("the drawing has no button: %.400s", data)
	}
}

// failsToDraw requires a salt diagram to return an error and no drawing in
// every format.
func failsToDraw(t *testing.T, src string, formats ...font.FileFormat) {
	t.Helper()
	block := parseBlock(t, core.NewRenderContext(), src)
	for _, format := range formats {
		data, err := render(t, block, format)
		if err == nil {
			t.Errorf("%v: no error", format)
		}
		if len(data) != 0 {
			t.Errorf("%v: %d bytes written", format, len(data))
		}
	}
}

// TestSaltContentOutsideAGroup checks the salt-diagram scenario "Content
// outside a group".
func TestSaltContentOutsideAGroup(t *testing.T) {
	failsToDraw(t, "@startsalt\nhello\n@endsalt\n", font.FormatSVG, font.FormatPNG)
}

// TestSaltUnclosedGroup checks the salt-diagram scenario "Unclosed group".
func TestSaltUnclosedGroup(t *testing.T) {
	failsToDraw(t, "@startsalt\n{\n[OK]\n@endsalt\n", font.FormatSVG)
}

// TestSaltUnknownImage checks the salt-diagram scenario "Unknown image".
func TestSaltUnknownImage(t *testing.T) {
	failsToDraw(t, "@startsalt\n{\n<<nothere>>\n}\n@endsalt\n", font.FormatSVG)
}

// TestSaltUnknownMenu checks the salt-diagram scenario "Unknown menu".
func TestSaltUnknownMenu(t *testing.T) {
	failsToDraw(t, "@startsalt\n{* File | Edit\n View | Zoom }\n@endsalt\n", font.FormatSVG)
}

// TestSaltMenuListOrder checks the salt-diagram scenario "Several menus with
// items": the lists are drawn in the order of the entries.
func TestSaltMenuListOrder(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), "@startsalt\n{\n{* File | Edit\n File | New\n Edit | Undo }\n}\n@endsalt\n")
	data, err := render(t, block, font.FormatSVG)
	if err != nil {
		t.Fatal(err)
	}
	svg := string(data)
	newAt, undoAt := strings.Index(svg, ">New</text>"), strings.Index(svg, ">Undo</text>")
	if newAt < 0 || undoAt < 0 || newAt > undoAt {
		t.Errorf("New at %d and Undo at %d, want the list of File first", newAt, undoAt)
	}
}
