// SPDX-License-Identifier: MIT
// Provenance: original

package command

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// passDiagram requires three passes and logs the commands it runs.
type passDiagram struct {
	Base
	log []string
}

func (*passDiagram) Description() string { return "(passes)" }

func (*passDiagram) RequiredPasses() []ParserPass { return []ParserPass{One, Two, Three} }

// passBlock is a multi-line command: "begin" ... "end", logged as "block".
type passBlock struct{}

func (passBlock) IsValid(lines BlocLines) Control { return section{}.IsValid(lines) }

func (passBlock) Execute(d *passDiagram, _ BlocLines, pass ParserPass) Result {
	d.log = append(d.log, fmt.Sprintf("%d:block", pass))
	return Ok()
}

func (passBlock) EligibleFor(p ParserPass) bool { return p == One }

// limitedBlock is a passBlock that accepts few lines.
type limitedBlock struct{ passBlock }

func (limitedBlock) MaxLines() int { return 2 }

func wordCommand(word string) *SingleLine[*passDiagram] {
	return &SingleLine[*passDiagram]{
		Pattern: regex.NewConcat(regex.Start(), regex.NewLeaf(word), regex.End()),
		Exec: func(d *passDiagram, _ *text.LineLocation, _ *regex.Result, pass ParserPass) Result {
			d.log = append(d.log, fmt.Sprintf("%d:%s", pass, word))
			return Ok()
		},
	}
}

// TestInPasses checks that a wrapped single-line command and a wrapped
// multi-line command run only in their passes, while the other commands
// keep their own.
func TestInPasses(t *testing.T) {
	f := &Factory[*passDiagram]{
		Type: core.State,
		Commands: func() []Command[*passDiagram] {
			return []Command[*passDiagram]{
				wordCommand("first"),
				InPasses(wordCommand("second"), Two),
				InPasses(passBlock{}, Three),
			}
		},
		NewDiagram: func(*core.RenderContext, *core.UmlSource) *passDiagram { return &passDiagram{} },
	}
	lines := core.BuildBlocks("@startuml\nbegin\nx\nend\nsecond\nfirst\n@enduml\n", "string", nil)[0].Lines
	d, ok := f.CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, false)).(*passDiagram)
	if !ok {
		t.Fatalf("not parsed: %#v", d)
	}
	if want := []string{"0:first", "1:second", "2:block"}; !slices.Equal(d.log, want) {
		t.Errorf("log %v, want %v", d.log, want)
	}
}

// TestInPassesMaxLines checks that the wrapper keeps the line limit of the
// wrapped command.
func TestInPassesMaxLines(t *testing.T) {
	if got := InPasses(passBlock{}, Two).(MaxLinesCommand).MaxLines(); got != math.MaxInt {
		t.Errorf("unlimited command: %d", got)
	}
	if got := InPasses(limitedBlock{}, Two).(MaxLinesCommand).MaxLines(); got != 2 {
		t.Errorf("limited command: %d", got)
	}
}
