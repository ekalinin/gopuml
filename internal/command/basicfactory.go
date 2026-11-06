// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/PSystemBasicFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// SyntaxErrorMessage is upstream's error message for a line that a
// BasicFactory does not accept.
const SyntaxErrorMessage = "Syntax Error?"

// BasicFactory parses blocks of one diagram type line by line, without
// commands, as upstream's PSystemBasicFactory: InitDiagram starts the
// diagram and ExecuteLine adds each line to it. The functions hold no state
// of their own, as concurrent renders share the factory.
type BasicFactory struct {
	Type core.DiagramType
	// InitDiagram returns the diagram before the first line, or nil.
	InitDiagram func(ctx *core.RenderContext, src *core.UmlSource, startLine string) core.Diagram
	// ExecuteLine returns the diagram with the line, or nil when the line is
	// a syntax error. system is the diagram of the previous lines, or nil.
	ExecuteLine func(ctx *core.RenderContext, src *core.UmlSource, system core.Diagram, line string) core.Diagram
}

// DiagramType implements core.Factory.
func (f *BasicFactory) DiagramType() core.DiagramType { return f.Type }

// CreateSystem implements core.Factory. The noise that directly follows the
// start line and the empty lines that come first are skipped. A source of
// two lines fails with Empty description at its end line, and a line that
// ExecuteLine refuses fails with Syntax Error?. When no line made a diagram,
// as for a block of blank lines, it returns nil (upstream throws an
// IllegalStateException while merging the errors).
func (f *BasicFactory) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	src = src.RemoveInitialNoise()
	it := newIterator(src.Lines())
	start := it.next()
	system := f.InitDiagram(ctx, src, start.S)
	first := true
	for it.hasNext() {
		s := it.next()
		if first && text.Trin(s.S) == "" {
			continue
		}
		first = false
		if core.IsEndDirective(s.S) {
			if src.TotalLineCount() == 2 && !src.IsStartDef() {
				return f.newError(src, EmptyDescription, s, it.trace())
			}
			return system
		}
		if system = f.ExecuteLine(ctx, src, system, s.S); system == nil {
			return f.newError(src, SyntaxErrorMessage, s, it.trace())
		}
	}
	return system
}

func (f *BasicFactory) newError(src *core.UmlSource, msg string, line text.StringLocated, trace []text.StringLocated) *core.ErrorDiagram {
	t := f.Type
	return &core.ErrorDiagram{
		Source: src.Lines(),
		Trace:  trace,
		Error:  core.ErrorUml{Type: core.SyntaxError, Error: msg, Line: &line, AssumedType: &t},
	}
}
