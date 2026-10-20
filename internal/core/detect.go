// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandArchimate.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandRobustConcise.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandClock.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandBinary.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandAnalog.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandAtTime.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
)

// unsupportedType recognizes an @startuml diagram type this implementation
// cannot parse yet, from the leading part of its characteristic commands.
// Keywords that sequence, class or activity diagrams also accept are left
// out, so that their syntax errors are not reported as another type.
type unsupportedType struct {
	name     string
	patterns []*regexp.Regexp
}

func patterns(p ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(p))
	for i, s := range p {
		out[i] = regex.MustCompile(s)
	}
	return out
}

// unsupportedTypes are tried in upstream's factory order.
var unsupportedTypes = []unsupportedType{
	{"archimate", patterns(
		`^archimate[%s]`,
	)},
	{"timing", patterns(
		`^(compact[%s]+)?(robust|concise)[%s]+`,
		`^(compact[%s]+)?clock[%s]+.*[%s]with[%s]+period[%s]`,
		`^(compact[%s]+)?(binary|analog)[%s]+`,
		`^@[+-]?[0-9]`,
	)},
}

// detectUnsupportedUmlType returns a "not supported yet" error when an
// @startuml block uses the characteristic syntax of a diagram type this
// implementation does not parse yet, or nil.
func detectUnsupportedUmlType(src *UmlSource) Diagram {
	lines := src.Lines()
	for _, t := range unsupportedTypes {
		for i, l := range lines {
			s := text.Trin(l.S)
			for _, p := range t.patterns {
				if !p.MatchString(s) {
					continue
				}
				line := l
				return &ErrorDiagram{
					Source: lines,
					Trace:  lines[:i+1],
					Error:  ErrorUml{Type: ExecutionError, Error: t.name + " diagrams are not supported yet", Line: &line},

					NotSupportedYet: true,
				}
			}
		}
	}
	return nil
}
