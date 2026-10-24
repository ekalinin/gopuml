// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

const fields = "@startpacketdiag\n0-7: Source Port\n8-15: Destination Port\n@endpacketdiag\n"

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if d, ok := blocks[0].Diagram.(*core.ErrorDiagram); ok && d.NotSupportedYet {
		t.Skip("packet diagrams are not registered yet (task 3.1)")
	}
	if got := blocks[0].Diagram.Description(); got != "Packet Diagram" {
		t.Fatalf("diagram %T, description %q", blocks[0].Diagram, got)
	}
	return blocks[0]
}

// TestPacketTxtUnsupported checks the packet-diagram scenario "TXT for a
// packet diagram".
func TestPacketTxtUnsupported(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), fields)
	for _, format := range []font.FileFormat{font.FormatATXT, font.FormatUTXT} {
		var buf bytes.Buffer
		_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
		var unsupported *engine.UnsupportedFormatError
		if !errors.As(err, &unsupported) || !strings.HasSuffix(err.Error(), " output is not supported for packet diagrams") {
			t.Errorf("%v: error %v", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}

// TestPacketWithoutDot checks the packet-diagram scenario "Packet diagram
// without Graphviz".
func TestPacketWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, fields)
	var buf bytes.Buffer
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), ">Source Port</text>") {
		t.Errorf("the drawing has no field: %.400s", buf.String())
	}
}

// TestPacketReversedRange checks the packet-diagram scenario "Reversed
// range": upstream fails to draw a field of negative width and writes a
// crash report, so the render returns an error and no drawing.
func TestPacketReversedRange(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), "@startpacketdiag\n4-2: x\n@endpacketdiag\n")
	var buf bytes.Buffer
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err == nil {
		t.Fatal("the reversed range rendered without an error")
	}
	if buf.Len() != 0 {
		t.Errorf("%d bytes written", buf.Len())
	}
}
