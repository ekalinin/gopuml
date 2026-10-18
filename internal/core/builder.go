// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/PSystemBuilder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/error/PSystemErrorUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/directdot/PSystemDotFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// Factory parses the blocks of one diagram type.
type Factory interface {
	DiagramType() DiagramType
	// CreateSystem returns the diagram, or an *ErrorDiagram when the source
	// is not a valid diagram of this type.
	CreateSystem(ctx *RenderContext, src *UmlSource) Diagram
}

// graphvizDotHeader matches a Graphviz graph header such as "digraph G {".
var graphvizDotHeader = regexp.MustCompile(`^[\t\n\x0B\f\r ]*(strict[\t\n\x0B\f\r ]+)?(di)?graph[\t\n\x0B\f\r ]+([_\p{L}][_\p{L}\p{N}]*|-?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)|"([^"\\]|\\")*")?[\t\n\x0B\f\r ]*\{[\t\n\x0B\f\r ]*$`)

// misuseChecks detect other diagram languages written in an @startuml block.
var misuseChecks = []struct {
	match   func(trimmed, line string) bool
	message string
}{
	{func(_, line string) bool { return graphvizDotHeader.MatchString(line) }, "This looks like a DOT diagram. Please use @startdot instead of @startuml."},
	{func(trimmed, _ string) bool { return trimmed == "ditaa" }, "This looks like a DITAA diagram. Please use @startditaa instead of @startuml."},
	{func(trimmed, _ string) bool { return trimmed == "salt" }, "This looks like a salt diagram. Please use @startsalt instead of @startuml."},
	{func(trimmed, _ string) bool { return trimmed == "nwdiag {" }, "This looks like a network diagram. Please use @startnwdiag instead of @startuml."},
}

// checkBasicError returns a diagram for blocks that no parser should see:
// unknown start directives and other languages in an @startuml block.
func checkBasicError(types []DiagramType, src *UmlSource) Diagram {
	lines := src.Lines()
	if slices.Contains(types, Unknown) {
		return &UnsupportedDiagram{Directive: lines[0].S}
	}
	if slices.Contains(types, Sequence) && len(lines) > 1 {
		second := lines[1]
		trimmed := strings.TrimFunc(second.S, func(r rune) bool { return r <= ' ' })
		for _, c := range misuseChecks {
			if c.match(trimmed, second.S) {
				seq := Sequence
				return &ErrorDiagram{
					Source: lines,
					Trace:  lines,
					Error:  ErrorUml{Type: ExecutionError, Error: c.message, Score: 100, Line: &second, AssumedType: &seq},
				}
			}
		}
	}
	return nil
}

// CreateDiagram turns a block into a diagram: the first factory, in order,
// that accepts the block wins; when all fail, the best scored error is
// returned. lines are the preprocessed block lines.
func CreateDiagram(ctx *RenderContext, lines, raw []text.StringLocated, factories []Factory) (result Diagram) {
	types := FindStartTypes(lines[0].S)
	src := NewUmlSource(lines, raw, slices.Contains(types, Sequence))
	if explain := src.ExplainTypes(); explain != nil {
		first := lines[0]
		return &ErrorDiagram{Source: lines, Trace: []text.StringLocated{first}, Error: ErrorUml{Type: ExecutionError, Error: "explain diagrams are not supported yet", Line: &first}, NotSupportedYet: true}
	}
	if d := checkBasicError(src.DiagramTypes(), src); d != nil {
		return d
	}
	if d := CheckFamily(&Block{Lines: lines}); d != nil {
		return d
	}

	var errors []*ErrorDiagram
	for _, f := range factories {
		if !slices.Contains(src.DiagramTypes(), f.DiagramType()) {
			continue
		}
		d := createSafely(ctx, f, src)
		switch d := d.(type) {
		case *ErrorDiagram:
			errors = append(errors, d)
		case *UnsupportedDiagram, nil:
		default:
			return d
		}
	}
	// No parser accepted the block: name the diagram type when the block
	// uses the syntax of an @startuml type that is not supported yet.
	if slices.Contains(src.DiagramTypes(), Sequence) {
		if d := detectUnsupportedUmlType(src); d != nil {
			return d
		}
	}
	if len(errors) == 0 {
		return &UnsupportedDiagram{Directive: lines[0].S}
	}
	return mergeErrors(errors)
}

// createSafely runs a factory and turns a panic into an execution error, as
// upstream does with exceptions.
func createSafely(ctx *RenderContext, f Factory, src *UmlSource) (d Diagram) {
	defer func() {
		if r := recover(); r != nil {
			first := src.Lines()[0]
			d = &ErrorDiagram{
				Source: src.Lines(),
				Trace:  src.Lines(),
				Error: ErrorUml{
					Type:  ExecutionError,
					Error: fmt.Sprintf("Fatal crash error, you should send a mail to plantuml@gmail.com with: %v", r),
					Line:  &first,
				},
			}
		}
	}()
	return f.CreateSystem(ctx, src)
}

// mergeErrors keeps the error with the highest score, the first one on ties.
func mergeErrors(errors []*ErrorDiagram) *ErrorDiagram {
	best := errors[0]
	for _, e := range errors[1:] {
		if best.Score() < e.Score() {
			best = e
		}
	}
	return best
}
