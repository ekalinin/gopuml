// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/PSystemCommandFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/PSystemAbstractFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/ProtectedCommand.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/EmbeddedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// UpstreamVersion appears in crash messages, as upstream's version does.
const UpstreamVersion = "1.2026.9"

// EmptyDescription is upstream's error message for a diagram without content.
const EmptyDescription = "Empty description"

// Factory parses blocks of one diagram type with a list of commands. The
// first command, in order, that accepts the next lines is executed.
type Factory[D Diagram] struct {
	Type core.DiagramType
	// Commands builds the command list once.
	Commands func() []Command[D]
	// NewDiagram creates the empty diagram that commands fill.
	NewDiagram func(ctx *core.RenderContext, src *core.UmlSource) D
	// FirstTokens, when set, are the first tokens of the commands
	// (FirstTokensOf), computed by the caller once for all its factories.
	FirstTokens [][]string

	once   sync.Once
	cmds   []Command[D]
	tokens []commandTokens
}

// DiagramType implements core.Factory.
func (f *Factory[D]) DiagramType() core.DiagramType { return f.Type }

func (f *Factory[D]) commands() []Command[D] {
	f.once.Do(func() {
		f.cmds = f.Commands()
		// What each command declared, asked once rather than once per line.
		f.tokens = make([]commandTokens, len(f.cmds))
		for i, cmd := range f.cmds {
			if f.FirstTokens != nil {
				f.tokens[i] = commandTokens{tokens: f.FirstTokens[i]}
			} else {
				f.tokens[i] = commandTokens{tokens: FirstTokensOf(cmd)}
			}
		}
	})
	return f.cmds
}

func (f *Factory[D]) newError(src *core.UmlSource, kind core.ErrorType, msg string, score int, line text.StringLocated, trace []text.StringLocated) *core.ErrorDiagram {
	t := f.Type
	return &core.ErrorDiagram{
		Source: src.Lines(),
		Trace:  trace,
		Error:  core.ErrorUml{Type: kind, Error: msg, Score: score, Line: &line, AssumedType: &t},
	}
}

// CreateSystem implements core.Factory.
func (f *Factory[D]) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	it := newIterator(src.Lines())
	start := it.next()
	if !core.IsStartDirective(start.S) {
		panic("command: source does not begin with a start directive")
	}
	if src.IsEmpty() {
		if it.hasNext() {
			it.next()
		}
		return f.newError(src, core.SyntaxError, EmptyDescription, 0, start, it.trace)
	}

	var sys Diagram = f.NewDiagram(ctx, src)
	for _, pass := range sys.RequiredPasses() {
		sys.StartingPass(pass)
		for it.hasNext() {
			if core.IsEndDirective(it.peek().S) {
				it = newIterator(src.Lines())
				it.next()
				break
			}
			next, failed := f.executeFewLines(sys, src, it, pass)
			if failed != nil {
				return failed
			}
			sys = next
		}
	}
	return f.finalize(sys, src, it)
}

func (f *Factory[D]) finalize(sys Diagram, src *core.UmlSource, it *iterator) core.Diagram {
	if msg := sys.CheckFinalError(); msg != "" {
		loc := it.next()
		return f.newError(src, core.ExecutionError, msg, 0, loc, it.trace)
	}
	if src.TotalLineCount() == 2 {
		loc := it.next()
		return f.newError(src, core.SyntaxError, EmptyDescription, 0, loc, it.trace)
	}
	sys.MakeDiagramReady()
	if sys.IsIncomplete() {
		return nil
	}
	return sys
}

// executeFewLines runs the command that accepts the next lines. It returns
// the diagram to continue with, or the error that ends parsing.
func (f *Factory[D]) executeFewLines(sys Diagram, src *core.UmlSource, it *iterator, pass ParserPass) (Diagram, *core.ErrorDiagram) {
	cmd, lines, ok := f.candidate(it)
	if !ok {
		line := it.peek()
		it.next()
		return nil, f.newError(src, core.SyntaxError, "Syntax Error?", 0, line, it.trace)
	}
	if !cmd.EligibleFor(pass) {
		return sys, nil
	}
	var result Result
	if del, isDel := sys.(Delegator); isDel {
		result = del.Delegate(func(target Diagram) Result {
			return cmd.Execute(asD[D](target), lines, pass)
		})
	} else {
		result = execute(cmd, asD[D](sys), lines, pass)
	}
	if !result.IsOK() {
		first, _ := lines.First()
		return nil, f.newError(src, core.ExecutionError, result.Err(), result.Score(), first, it.trace)
	}
	if result.NewDiagram != nil {
		return result.NewDiagram, nil
	}
	return sys, nil
}

// asD returns the diagram as the factory's diagram type.
func asD[D Diagram](sys Diagram) D {
	d, isD := sys.(D)
	if !isD {
		panic(fmt.Sprintf("command: diagram %T is not the factory's diagram type", sys))
	}
	return d
}

// execute runs a command, turning a panic into an error as upstream's
// ProtectedCommand does with exceptions.
func execute[D Diagram](cmd Command[D], d D, lines BlocLines, pass ParserPass) (r Result) {
	defer func() {
		if p := recover(); p != nil {
			r = Error(fmt.Sprintf("You should send a mail to plantuml@gmail.com with this log (V%s) %v", UpstreamVersion, p))
		}
	}()
	return cmd.Execute(d, lines, pass)
}

// candidate finds the first command that accepts the next line, or the
// next lines for a multi-line command, and advances past them. Only the
// commands that declared the first token of the line, or declared none, are
// tried, in order.
func (f *Factory[D]) candidate(it *iterator) (Command[D], BlocLines, bool) {
	single := Single(it.peek())
	token := it.peek().FirstToken()
	for i, cmd := range f.commands() {
		if !f.tokens[i].accepts(token) {
			continue
		}
		switch cmd.IsValid(single) {
		case OK:
			it.next()
			return cmd, single, true
		case OKPartial:
			cloned := it.clone()
			if lines, ok := multilineLines(cloned, cmd); ok {
				it.copyStateFrom(cloned)
				return cmd, lines, true
			}
		}
	}
	return nil, BlocLines{}, false
}

// MaxLinesCommand limits how many lines a multi-line command may span.
type MaxLinesCommand interface {
	MaxLines() int
}

func multilineLines[D Diagram](it *iterator, cmd Command[D]) (BlocLines, bool) {
	var lines BlocLines
	nb := 0
	for it.hasNext() {
		lines = addLineWithEmbedded(it, lines)
		switch cmd.IsValid(lines) {
		case NotOK:
			return BlocLines{}, false
		case OK:
			return lines, true
		}
		nb++
		if m, ok := cmd.(MaxLinesCommand); ok && nb > m.MaxLines() {
			return BlocLines{}, false
		}
	}
	return BlocLines{}, false
}

// addLineWithEmbedded adds the next line, and the whole {{ ... }} block when
// the line opens an embedded diagram.
func addLineWithEmbedded(it *iterator, lines BlocLines) BlocLines {
	l := it.next()
	lines = lines.Add(l)
	if EmbeddedType(l.S) == "" {
		return lines
	}
	nested := 1
	for it.hasNext() {
		s := it.next()
		lines = lines.Add(s)
		if EmbeddedType(s.S) != "" {
			nested++
		} else if text.Trin(s.S) == "}}" {
			nested--
			if nested == 0 {
				return lines
			}
		}
	}
	return lines
}

// embeddedTypes are the diagram types allowed after "{{".
var embeddedTypes = []string{
	"board", "creole", "chronology", "chen", "chart", "ditaa", "ebnf", "files", "gantt", "json",
	"mindmap", "nwdiag", "packetdiag", "regex", "salt", "uml", "wbs", "wire", "yaml",
}

// EmbeddedType returns the type of an embedded diagram opened by s, such as
// "uml" for "{{" or "json" for "{{json", or "" when s opens none.
func EmbeddedType(s string) string {
	trimmed := strings.TrimRightFunc(strings.TrimLeftFunc(s, text.IsJavaWhitespace), text.IsJavaWhitespace)
	rest, ok := strings.CutPrefix(trimmed, "{{")
	if !ok {
		return ""
	}
	if rest == "" {
		return "uml"
	}
	for _, t := range embeddedTypes {
		if rest == t {
			return t
		}
	}
	return ""
}
