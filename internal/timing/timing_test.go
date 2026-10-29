// SPDX-License-Identifier: MIT
// Provenance: original

package timing_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/timing"
)

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*timing.TimingDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

// TestTimingWithoutDot checks the timing-diagram scenario "Timing diagram
// without Graphviz".
func TestTimingWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, "@startuml\nconcise \"Web User\" as WU\n@0\nWU is Idle\n@100\nWU is Waiting\n@enduml\n")
	var buf bytes.Buffer
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(">Waiting</text>")) {
		t.Errorf("the drawing has no state: %.400s", buf.String())
	}
}

// TestTimingDrawingFailure checks the timing-diagram scenario "Drawing
// failure": every time is below zero, which upstream's drawing crashes on.
func TestTimingDrawingFailure(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(),
		"@startuml\nconcise \"User\" as U\n@-20\nU is Idle\n@-10\nU is Busy\n@enduml\n")
	for _, format := range []font.FileFormat{font.FormatSVG, font.FormatPNG} {
		var buf bytes.Buffer
		if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format}); err == nil {
			t.Errorf("%v: no error", format)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}
