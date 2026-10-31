// SPDX-License-Identifier: MIT
// Provenance: original

package cheneer_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/cheneer"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// TestChenWithoutDot checks the chen-eer-diagram scenario "Chen diagram
// without dot": the Graphviz-not-found error diagram.
func TestChenWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	blocks, err := engine.Parse(context.Background(), rc,
		engine.Source{Text: "@startchen\nentity CUSTOMER {\n  Number\n}\n@endchen\n"})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*cheneer.ChenEerDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	var buf bytes.Buffer
	_, err = blocks[0].Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG})
	if !errors.Is(err, dot.ErrNotFound) {
		t.Fatalf("error %v, want dot.ErrNotFound", err)
	}
	for _, want := range []string{"Dot Executable: /nonexistent/graphviz/dot", "Cannot find Graphviz. You should try"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("the image does not say %q", want)
		}
	}
}
