// SPDX-License-Identifier: MIT
// Provenance: original

package command

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// kvDiagram is a toy diagram of "key = value" lines and "begin" ... "end"
// sections, used to test the framework like upstream's command factories.
type kvDiagram struct {
	Base
	values   map[string]string
	sections [][]string
	final    string
}

func (*kvDiagram) Description() string { return "(kv)" }

func (d *kvDiagram) CheckFinalError() string { return d.final }

// section is a multi-line command: "begin" ... "end".
type section struct{}

func (section) IsValid(lines BlocLines) Control {
	if text.Trin(lines.At(0).S) != "begin" {
		return NotOK
	}
	if lines.Size() > 1 && text.Trin(lines.Last().S) == "end" {
		return OK
	}
	return OKPartial
}

func (section) Execute(d *kvDiagram, lines BlocLines, _ ParserPass) Result {
	d.sections = append(d.sections, lines.Strings())
	return Ok()
}

func (section) EligibleFor(p ParserPass) bool { return p == One }

func kvFactory() *Factory[*kvDiagram] {
	return &Factory[*kvDiagram]{
		Type: core.Sequence,
		Commands: func() []Command[*kvDiagram] {
			return []Command[*kvDiagram]{
				&SingleLine[*kvDiagram]{
					Pattern: regex.NewConcat(regex.Start(),
						regex.NewNamedLeaf(1, "KEY", `(\w+)`), regex.SpaceZeroOrMore(), regex.NewLeaf("="),
						regex.SpaceZeroOrMore(), regex.NewNamedLeaf(1, "VALUE", `(\w+)`), regex.End()),
					Exec: func(d *kvDiagram, _ *text.LineLocation, arg *regex.Result, _ ParserPass) Result {
						switch v := arg.Str("VALUE", 0); v {
						case "bad":
							return ErrorScore("bad value", 7)
						case "boom":
							panic("exploded")
						case "final":
							d.final = "final check failed"
						}
						d.values[arg.Str("KEY", 0)] = arg.Str("VALUE", 0)
						return Ok()
					},
				},
				section{},
			}
		},
		NewDiagram: func(*core.RenderContext, *core.UmlSource) *kvDiagram { return &kvDiagram{values: map[string]string{}} },
	}
}

func create(t *testing.T, src string) core.Diagram {
	t.Helper()
	lines := core.BuildBlocks(src, "string", nil)[0].Lines
	return kvFactory().CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, false))
}

func TestFactoryParses(t *testing.T) {
	d, ok := create(t, "@startuml\na = 1\n  b=2  \nbegin\nx\nend\n@enduml\n").(*kvDiagram)
	if !ok {
		t.Fatalf("want a diagram")
	}
	if d.values["a"] != "1" || d.values["b"] != "2" {
		t.Errorf("values = %v", d.values)
	}
	if len(d.sections) != 1 || strings.Join(d.sections[0], "|") != "begin|x|end" {
		t.Errorf("sections = %v", d.sections)
	}
}

func errorOf(t *testing.T, d core.Diagram) *core.ErrorDiagram {
	t.Helper()
	e, ok := d.(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("want an error diagram, got %#v", d)
	}
	return e
}

func TestFactorySyntaxError(t *testing.T) {
	e := errorOf(t, create(t, "@startuml\na = 1\nwhat is this\nb = 2\n@enduml\n"))
	if e.Error.Message() != "Syntax Error? (Assumed diagram type: sequence)" || e.Error.Type != core.SyntaxError {
		t.Errorf("message %q", e.Error.Message())
	}
	// Positions are 0-based: the offending line is the third line.
	if e.Error.Position() != 2 || len(e.Trace) != 3 || e.Score() != 30 {
		t.Errorf("position %d trace %d score %d", e.Error.Position(), len(e.Trace), e.Score())
	}
}

func TestFactoryExecutionError(t *testing.T) {
	e := errorOf(t, create(t, "@startuml\na = 1\nb = bad\n@enduml\n"))
	if e.Error.Message() != "bad value (Assumed diagram type: sequence)" || e.Error.Type != core.ExecutionError {
		t.Errorf("message %q", e.Error.Message())
	}
	if e.Error.Position() != 2 || e.Score() != 3*10+7 {
		t.Errorf("position %d score %d", e.Error.Position(), e.Score())
	}
}

func TestFactoryPanicIsReported(t *testing.T) {
	e := errorOf(t, create(t, "@startuml\na = boom\n@enduml\n"))
	if !strings.HasPrefix(e.Error.Error, "You should send a mail to plantuml@gmail.com with this log (V1.2026.9) exploded") {
		t.Errorf("message %q", e.Error.Error)
	}
}

func TestFactoryEmptyAndFinalErrors(t *testing.T) {
	// An empty diagram is reported at the start line.
	e := errorOf(t, create(t, "@startuml\n' only a comment\n\n@enduml\n"))
	if e.Error.Error != EmptyDescription || e.Error.Position() != 0 {
		t.Errorf("empty: %q at %d", e.Error.Error, e.Error.Position())
	}

	// The final check reports at the first content line, as upstream does
	// after rewinding its iterator.
	e = errorOf(t, create(t, "@startuml\na = 1\nb = final\n@enduml\n"))
	if e.Error.Error != "final check failed" || e.Error.Position() != 1 || e.Error.Type != core.ExecutionError {
		t.Errorf("final: %q at %d", e.Error.Error, e.Error.Position())
	}
}

func TestFactoryUnterminatedMultiline(t *testing.T) {
	e := errorOf(t, create(t, "@startuml\nbegin\nx\n@enduml\n"))
	if e.Error.Error != "Syntax Error?" || e.Error.Position() != 1 {
		t.Errorf("%q at %d", e.Error.Error, e.Error.Position())
	}
}

func TestEmbeddedType(t *testing.T) {
	for in, want := range map[string]string{"{{": "uml", "  {{json  ": "json", "{{uml": "uml", "{{foo": "", "{ {": "", "x{{": ""} {
		if got := EmbeddedType(in); got != want {
			t.Errorf("EmbeddedType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBlocLinesHelpers(t *testing.T) {
	b := NewBlocLines([]text.StringLocated{
		text.NewStringLocated("  a", nil), text.NewStringLocated("", nil), text.NewStringLocated("\t  b", nil),
	})
	if got := strings.Join(b.RemoveEmptyColumns().Strings(), "|"); got != "a|| b" {
		t.Errorf("RemoveEmptyColumns = %q", got)
	}
	if got := strings.Join(b.Trim().RemoveEmptyLines().Strings(), "|"); got != "a|b" {
		t.Errorf("Trim+RemoveEmptyLines = %q", got)
	}
	if got := strings.Join(b.RemoveStartingAndEnding("x", 1).Strings(), "|"); got != "x||\t  " {
		t.Errorf("RemoveStartingAndEnding = %q", got)
	}
	added := b.Add(text.NewStringLocated("c", nil))
	if b.Size() != 3 || added.Size() != 4 {
		t.Error("Add must not modify the original")
	}
}
