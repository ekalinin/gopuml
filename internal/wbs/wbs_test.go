// SPDX-License-Identifier: MIT
// Provenance: original

package wbs_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/wbs"
)

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*wbs.WBSDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

// TestWBSWithoutDot checks the wbs-diagram scenario "WBS diagram without
// Graphviz".
func TestWBSWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, "@startwbs\n* root\n** a\n*** b\n@endwbs\n")
	var buf bytes.Buffer
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(">root</text>")) {
		t.Errorf("the drawing has no root: %.400s", buf.String())
	}
}

// TestWBSWithoutNodes checks the wbs-diagram scenario "WBS diagram without
// nodes".
func TestWBSWithoutNodes(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), "@startwbs\ntitle Plan\n@endwbs\n")
	for _, format := range []font.FileFormat{font.FormatSVG, font.FormatATXT} {
		var buf bytes.Buffer
		if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format}); err == nil {
			t.Errorf("%v: no error", format)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}
