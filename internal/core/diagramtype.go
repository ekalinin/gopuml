// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/DiagramType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"strings"
	"unicode/utf8"

	"github.com/ekalinin/gopuml/internal/text"
)

// DiagramType is a kind of diagram, with upstream's names.
type DiagramType int

// Diagram types, in upstream's order.
const (
	Sequence DiagramType = iota
	State
	Class
	Object
	Activity
	Description
	Composite
	Timing
	Help
	BPM
	Ditaa
	Dot
	JCCKit
	Salt
	Flow
	Creole
	Math
	Latex
	Definition
	Gantt
	Chronology
	NwDiag
	MindMap
	WBS
	Wire
	JSON
	Git
	Board
	YAML
	HCL
	EBNF
	Regex
	Files
	ChenEER
	Chart
	Packet
	Sprites
	Crash
	Unknown
)

var diagramTypeNames = [...]string{
	"SEQUENCE", "STATE", "CLASS", "OBJECT", "ACTIVITY", "DESCRIPTION", "COMPOSITE", "TIMING", "HELP", "BPM",
	"DITAA", "DOT", "JCCKIT", "SALT", "FLOW", "CREOLE", "MATH", "LATEX", "DEFINITION", "GANTT", "CHRONOLOGY",
	"NWDIAG", "MINDMAP", "WBS", "WIRE", "JSON", "GIT", "BOARD", "YAML", "HCL", "EBNF", "REGEX", "FILES",
	"CHEN_EER", "CHART", "PACKET", "SPRITES", "CRASH", "UNKNOWN",
}

// String returns upstream's enum name, for example "SEQUENCE".
func (t DiagramType) String() string {
	if t < 0 || int(t) >= len(diagramTypeNames) {
		return "UNKNOWN"
	}
	return diagramTypeNames[t]
}

// HumanReadableName is the name used in error messages, for example
// "sequence" or "component".
func (t DiagramType) HumanReadableName() string {
	if t == Description {
		return "component"
	}
	return strings.ToLower(t.String())
}

// IsLegacyUML reports whether the type is one of the @startuml types.
func (t DiagramType) IsLegacyUML() bool {
	switch t {
	case Sequence, State, Class, Object, Activity, Description, Composite, Timing, Help, Sprites:
		return true
	}
	return false
}

// umlTypes are the types an @startuml block may have.
var umlTypes = []DiagramType{Sequence, State, Class, Object, Activity, Description, Composite, Timing, Help, Sprites}

// startKeywords maps the first letter of a start keyword to the keywords and
// their types, in upstream's matching order. Ditaa is not recognized in the
// MIT variant.
var startKeywords = map[byte][]struct {
	keyword string
	types   []DiagramType
}{
	'b': {{"bpm", []DiagramType{BPM}}, {"board", []DiagramType{Board}}},
	'c': {{"chart", []DiagramType{Chart}}, {"creole", []DiagramType{Creole}}, {"chronology", []DiagramType{Chronology}}, {"chen", []DiagramType{ChenEER}}, {"crash", []DiagramType{Crash}}},
	'd': {{"dot", []DiagramType{Dot}}, {"def", []DiagramType{Definition}}},
	'e': {{"ebnf", []DiagramType{EBNF}}},
	'f': {{"flow", []DiagramType{Flow}}, {"files", []DiagramType{Files}}},
	'g': {{"gantt", []DiagramType{Gantt}}, {"git", []DiagramType{Git}}},
	'h': {{"hcl", []DiagramType{HCL}}},
	'j': {{"json", []DiagramType{JSON}}},
	'l': {{"latex", []DiagramType{Latex}}},
	'm': {{"math", []DiagramType{Math}}, {"mindmap", []DiagramType{MindMap}}},
	'n': {{"nwdiag", []DiagramType{NwDiag}}},
	'p': {{"project", []DiagramType{Gantt}}, {"packetdiag", []DiagramType{Packet}}},
	'r': {{"regex", []DiagramType{Regex}}},
	's': {{"salt", []DiagramType{Salt}}, {"sprites", []DiagramType{Sprites}}},
	'u': {{"uml", umlTypes}},
	'w': {{"wire", []DiagramType{Wire}}, {"wbs", []DiagramType{WBS}}},
	'y': {{"yaml", []DiagramType{YAML}}},
}

// checkKeyword compares s[p:] with key, folding ASCII upper case letters.
func checkKeyword(key, s string, p int) bool {
	if p+len(key) > len(s) {
		return false
	}
	for i := range len(key) {
		c := s[p+i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != key[i] {
			return false
		}
	}
	return true
}

// FindStartTypes returns the diagram types a start line allows: nil when s
// is not a start line, [Unknown] for an unrecognized start keyword.
func FindStartTypes(s string) []DiagramType {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if text.IsJavaWhitespace(r) {
			i += size
			continue
		}
		if r != '@' && r != '\\' {
			return nil
		}
		pos := i + 1
		if len(s)-pos < 5 || !checkKeyword("start", s, pos) {
			return nil
		}
		p := pos + 5
		if p >= len(s) {
			return nil
		}
		first := s[p]
		if first >= 'A' && first <= 'Z' {
			first += 'a' - 'A'
		}
		for _, k := range startKeywords[first] {
			if checkKeyword(k.keyword, s, p) {
				return k.types
			}
		}
		return []DiagramType{Unknown}
	}
	return nil
}
