// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/ErrorUml.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/ErrorUmlType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/error/PSystemUnsupported.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/text"
)

// Diagram is the result of parsing one block.
type Diagram interface {
	// Description is upstream's short description, for example
	// "(2 participants)" or "(Error)".
	Description() string
}

// TypedDiagram is a diagram of a known type, as upstream's TitledDiagram.
type TypedDiagram interface {
	Diagram
	DiagramType() DiagramType
}

// ErrorType classifies an error, as upstream's ErrorUmlType.
type ErrorType int

// Error types.
const (
	SyntaxError ErrorType = iota
	ExecutionError
)

// ErrorUml is one error of a block.
type ErrorUml struct {
	Type  ErrorType
	Error string
	// Score ranks errors from different parsers; the best one is shown.
	Score int
	// Line is the offending line; its location may be nil.
	Line *text.StringLocated
	// AssumedType, when set, is appended to the message.
	AssumedType *DiagramType
}

// Message is the error text shown to users, with the assumed diagram type
// when known, for example "Syntax Error? (Assumed diagram type: sequence)".
func (e ErrorUml) Message() string {
	if e.AssumedType != nil {
		return e.Error + " (Assumed diagram type: " + e.AssumedType.HumanReadableName() + ")"
	}
	return e.Error
}

// Position is the 0-based position of the offending line in its source.
func (e ErrorUml) Position() int {
	if e.Line == nil || e.Line.Location == nil {
		return 0
	}
	return e.Line.Location.Position
}

// ErrorDiagram is a block that could not be turned into a diagram, as
// upstream's PSystemError: one error and the lines read before it.
type ErrorDiagram struct {
	// Source is the block source.
	Source []text.StringLocated
	// Trace holds the lines consumed before the error.
	Trace []text.StringLocated
	Error ErrorUml
	// NotSupportedYet marks the "not supported yet" error of a diagram
	// type or family this implementation does not parse yet.
	NotSupportedYet bool
}

// Description implements Diagram.
func (*ErrorDiagram) Description() string { return "(Error)" }

// Position is PSystemError.getLineLocation: the 0-based position of the
// last line of the trace, which upstream reports as the error line. It can
// differ from the line of the error itself, for example for multi-line
// commands.
func (d *ErrorDiagram) Position() int {
	if n := len(d.Trace); n > 0 && d.Trace[n-1].Location != nil {
		return d.Trace[n-1].Location.Position
	}
	return d.Error.Position()
}

// Score ranks errors from different parsers: parsers that consumed more
// lines before failing are preferred.
func (d *ErrorDiagram) Score() int { return len(d.Trace)*10 + d.Error.Score }

// UnsupportedDiagram is shown for a start directive upstream does not know.
type UnsupportedDiagram struct {
	Directive string
}

// Description implements Diagram.
func (*UnsupportedDiagram) Description() string { return "(Unsupported)" }

// Lines returns the text of the diagram, with creole markup. version is
// shown on the "Running on" line.
func (d *UnsupportedDiagram) Lines(version, license string) []string {
	return []string{
		"<b>Diagram not supported by this release of PlantUML",
		" ",
		`Sorry, but the following directive ""` + d.Directive + `"" is not recognized.`,
		" ",
		"Possible causes:",
		"- Typo in the directive or incorrect syntax.",
		"- The directive was added in a newer PlantUML release.",
		" ",
		"Suggested actions:",
		"- Check the directive spelling and syntax.",
		"- Upgrade PlantUML to the latest version.",
		"- Consult the documentation: https://plantuml.com",
		" ",
		"Running on " + version,
		"(License " + license + ")",
	}
}

// supportedTypes are the diagram types this implementation can parse.
var supportedTypes = []DiagramType{Sequence, Class, Activity, Packet, Gantt, MindMap, WBS, Salt}

// CheckFamily handles start directives whose family this implementation
// cannot parse. It returns nil when the block's types include a supported
// type, upstream's unsupported diagram for an unknown directive, and an error
// diagram stating that the family is not supported yet otherwise.
func CheckFamily(block *Block) Diagram {
	if len(block.Lines) == 0 {
		return nil
	}
	first := block.Lines[0]
	types := FindStartTypes(first.S)
	switch {
	case slices.Contains(types, Unknown):
		return &UnsupportedDiagram{Directive: first.S}
	case slices.ContainsFunc(types, func(t DiagramType) bool { return slices.Contains(supportedTypes, t) }):
		return nil
	}
	name := "unknown"
	if len(types) > 0 {
		name = types[0].HumanReadableName()
	}
	return &ErrorDiagram{
		Source: block.Lines,
		Trace:  []text.StringLocated{first},
		Error:  ErrorUml{Type: ExecutionError, Error: name + " diagrams are not supported yet", Line: &first},

		NotSupportedYet: true,
	}
}
