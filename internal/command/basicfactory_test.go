// SPDX-License-Identifier: MIT
// Provenance: original

package command

import (
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
)

// okDiagram is a toy diagram of the lines that start with "ok".
type okDiagram struct{ lines []string }

func (*okDiagram) Description() string { return "(ok)" }

func okFactory() *BasicFactory {
	return &BasicFactory{
		Type:        core.Dot,
		InitDiagram: func(*core.RenderContext, *core.UmlSource, string) core.Diagram { return nil },
		ExecuteLine: func(_ *core.RenderContext, _ *core.UmlSource, system core.Diagram, line string) core.Diagram {
			if !strings.HasPrefix(line, "ok") {
				return nil
			}
			d := &okDiagram{}
			if system != nil {
				d.lines = slices.Clone(system.(*okDiagram).lines)
			}
			d.lines = append(d.lines, line)
			return d
		},
	}
}

func basicParse(src string) core.Diagram {
	lines := core.BuildBlocks(src, "string", nil)[0].Lines
	return okFactory().CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, false))
}

func TestBasicFactoryLines(t *testing.T) {
	for src, want := range map[string][]string{
		"@startdot\n\nok 1\nok 2\n@enddot\n":        {"ok 1", "ok 2"},
		"@startdot\nskinparam a b\nok 1\n@enddot\n": {"ok 1"},
	} {
		d, ok := basicParse(src).(*okDiagram)
		if !ok || !slices.Equal(d.lines, want) {
			t.Errorf("%q: diagram %#v, want the lines %q", src, d, want)
		}
	}
}

func TestBasicFactoryErrors(t *testing.T) {
	for src, want := range map[string]string{
		"@startdot\n@enddot\n":      "Empty description",
		"@startdot\nbad\n@enddot\n": "Syntax Error?",
	} {
		e, ok := basicParse(src).(*core.ErrorDiagram)
		if !ok {
			t.Fatalf("%q: not an error diagram", src)
		}
		if e.Error.Error != want || e.Error.Type != core.SyntaxError || e.Error.Score != 0 ||
			e.Error.Line.Location.Position != 1 || e.Error.AssumedType == nil || *e.Error.AssumedType != core.Dot {
			t.Errorf("%q: error %+v at %d, want %q at line 2 with the assumed type dot", src, e.Error, e.Error.Line.Location.Position, want)
		}
		if len(e.Trace) != 2 {
			t.Errorf("%q: trace of %d lines, want the start line and the failing one", src, len(e.Trace))
		}
	}
}

func TestBasicFactoryBlankBlock(t *testing.T) {
	// Upstream throws while merging the errors; no diagram is returned.
	if d := basicParse("@startdot\n   \n@enddot\n"); d != nil {
		t.Errorf("diagram %#v, want none", d)
	}
}
