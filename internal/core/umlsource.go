// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/core/UmlSource.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// UmlSource is the preprocessed source of one block, as parsers see it.
type UmlSource struct {
	lines []text.StringLocated
	raw   []text.StringLocated
}

// NewUmlSource builds the source of a block from its preprocessed lines.
// With checkEndingBackslash, lines ending with a backslash are joined with
// the next line, as upstream does for @startuml blocks.
func NewUmlSource(lines, raw []text.StringLocated, checkEndingBackslash bool) *UmlSource {
	src := &UmlSource{raw: raw}
	if !checkEndingBackslash {
		src.lines = append(src.lines, lines...)
		return src
	}
	var pending strings.Builder
	for _, l := range lines {
		if text.EndsWithBackslash(l.S) {
			pending.WriteString(l.S[:len(l.S)-1])
			continue
		}
		pending.WriteString(l.S)
		src.lines = append(src.lines, text.NewStringLocated(pending.String(), l.Location))
		pending.Reset()
	}
	return src
}

// Lines returns the source lines, from the start line to the end line.
func (s *UmlSource) Lines() []text.StringLocated { return s.lines }

// Raw returns the block lines before preprocessing.
func (s *UmlSource) Raw() []text.StringLocated { return s.raw }

// DiagramTypes returns the types allowed by the start line.
func (s *UmlSource) DiagramTypes() []DiagramType {
	if len(s.lines) == 0 {
		return nil
	}
	return FindStartTypes(s.lines[0].S)
}

// ExplainTypes returns the types of an @startexplain... block, or nil.
func (s *UmlSource) ExplainTypes() []DiagramType {
	if len(s.lines) == 0 || !strings.HasPrefix(s.lines[0].S, "@startexplain") {
		return nil
	}
	return FindStartTypes(strings.Replace(s.lines[0].S, "@startexplain", "@start", 1))
}

var commentLine = regexp.MustCompile(`^[\t\n\x0B\f\r ]*'.*$`)

// TotalLineCount is the number of source lines, start and end lines included.
func (s *UmlSource) TotalLineCount() int { return len(s.lines) }

// IsEmpty reports whether the block has no content besides the start and
// end lines, comment lines and blank lines.
func (s *UmlSource) IsEmpty() bool {
	for _, l := range s.lines {
		switch {
		case IsStartDirective(l.S), IsEndDirective(l.S), commentLine.MatchString(l.S):
			continue
		case text.Trin(l.S) != "":
			return false
		}
	}
	return true
}
