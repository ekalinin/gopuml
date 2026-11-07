// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"regexp"

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

// unsupportedTypes are tried in upstream's factory order. No @startuml
// diagram type is reported as not supported yet in this release.
var unsupportedTypes []unsupportedType

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
