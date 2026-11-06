// SPDX-License-Identifier: MIT
// Provenance: original

package directdot_test

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/directdot"
	"github.com/ekalinin/gopuml/internal/engine"
)

// parseDot preprocesses the first block of src and parses it with the dot
// factory.
func parseDot(t *testing.T, rc *core.RenderContext, src string) core.Diagram {
	t.Helper()
	blocks, err := engine.Parse(t.Context(), rc, engine.Source{Text: src})
	if err != nil {
		t.Fatal(err)
	}
	return directdot.Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Data, blocks[0].Raw, false))
}

func TestFactoryDotSource(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"graph", "@startdot\ndigraph G {\na -> b\n}\n@enddot\n", "digraph G {\na -> b\n}\n"},
		{"skin parameter", "@startdot\nskinparam monochrome true\ndigraph G {\na -> b\n}\n@enddot\n", "digraph G {\na -> b\n}\n"},
		{"preprocessor", "@startdot\n!define NODE alpha\ndigraph G {\n' a comment\nNODE -> b\n}\n@enddot\n", "digraph G {\nalpha -> b\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := parseDot(t, core.NewRenderContext(), tc.src)
			dot, ok := d.(*directdot.PSystemDot)
			if !ok {
				t.Fatalf("diagram %#v, want a *PSystemDot", d)
			}
			if got := dot.Data(); got != tc.want {
				t.Errorf("DOT source %q, want %q", got, tc.want)
			}
			if got := dot.Description(); got != "(Dot)" {
				t.Errorf("description %q, want (Dot)", got)
			}
			if _, typed := d.(core.TypedDiagram); typed {
				t.Error("a dot diagram must not be a core.TypedDiagram")
			}
		})
	}
}

func TestFactoryErrors(t *testing.T) {
	for src, want := range map[string]string{
		"@startdot\n@enddot\n":             "Empty description (Assumed diagram type: dot)",
		"@startdot\nfoo -> bar\n@enddot\n": "Syntax Error? (Assumed diagram type: dot)",
	} {
		e, ok := parseDot(t, core.NewRenderContext(), src).(*core.ErrorDiagram)
		if !ok {
			t.Fatalf("%q: not an error diagram", src)
		}
		if got := e.Error.Message(); got != want || e.Position() != 1 {
			t.Errorf("%q: %q at line %d, want %q at line 2", src, got, e.Position()+1, want)
		}
	}
}
