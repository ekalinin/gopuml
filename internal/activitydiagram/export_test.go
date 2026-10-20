// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/activitydiagram"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

const activities = "@startuml\n(*) --> \"First step\"\n\"First step\" --> (*)\n@enduml\n"

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*activitydiagram.ActivityDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

// TestLegacyActivityTxtUnsupported checks the legacy-activity-diagram
// scenario "TXT for a legacy activity diagram".
func TestLegacyActivityTxtUnsupported(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), activities)
	for _, format := range []font.FileFormat{font.FormatATXT, font.FormatUTXT} {
		var buf bytes.Buffer
		_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
		var unsupported *engine.UnsupportedFormatError
		if !errors.As(err, &unsupported) || !strings.HasSuffix(err.Error(), " output is not supported for activity diagrams") {
			t.Errorf("%v: error %v", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}
