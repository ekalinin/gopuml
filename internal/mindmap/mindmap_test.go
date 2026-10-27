// SPDX-License-Identifier: MIT
// Provenance: original

package mindmap_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/mindmap"
)

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*mindmap.MindMapDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

const ideas = "@startmindmap\n* root\n** a\n*** b\n@endmindmap\n"

// TestMindMapWithoutDot checks the mindmap-diagram scenario "Mindmap
// diagram without Graphviz".
func TestMindMapWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, ideas)
	var buf bytes.Buffer
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(">root</text>")) {
		t.Errorf("the drawing has no root: %.400s", buf.String())
	}
}

// TestMindMapTxtUnsupported checks the mindmap-diagram scenario "TXT for a
// mindmap diagram".
func TestMindMapTxtUnsupported(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), ideas)
	for _, format := range []font.FileFormat{font.FormatATXT, font.FormatUTXT} {
		var buf bytes.Buffer
		_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
		var unsupported *engine.UnsupportedFormatError
		if !errors.As(err, &unsupported) || !strings.HasSuffix(err.Error(), " output is not supported for mindmap diagrams") {
			t.Errorf("%v: error %v", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}
