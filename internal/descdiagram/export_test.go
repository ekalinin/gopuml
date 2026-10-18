// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

const useCases = "@startuml\nactor User\nusecase Login\nUser --> Login\n@enduml\n"

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*descdiagram.DescriptionDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

// TestDescriptionTxtUnsupported checks the description-diagram scenario
// "TXT for a description diagram": upstream's name of the type is
// "component".
func TestDescriptionTxtUnsupported(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), useCases)
	for _, format := range []font.FileFormat{font.FormatATXT, font.FormatUTXT} {
		var buf bytes.Buffer
		_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
		var unsupported *engine.UnsupportedFormatError
		if !errors.As(err, &unsupported) || !strings.HasSuffix(err.Error(), " output is not supported for component diagrams") {
			t.Errorf("%v: error %v", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}

// TestDescriptionWithoutDot checks the description-diagram scenario
// "Description diagram without dot": the Graphviz-not-found error diagram.
func TestDescriptionWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, useCases)
	var buf bytes.Buffer
	_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG})
	if !errors.Is(err, dot.ErrNotFound) {
		t.Fatalf("error %v, want dot.ErrNotFound", err)
	}
	for _, want := range []string{"Dot Executable: /nonexistent/graphviz/dot", "Cannot find Graphviz. You should try"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("the image does not say %q", want)
		}
	}
}
