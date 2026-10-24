// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/NotePosition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// NotePosition is the position of a note in a sequence diagram. Upstream
// defines it in sequencediagram; it lives here because Rose reads it and
// sequencediagram imports this package.
type NotePosition int

// Note positions, in upstream's order.
const (
	NotePositionLeft NotePosition = iota
	NotePositionRight
	NotePositionOver
	NotePositionOverSeveral
	NotePositionBottom
	NotePositionTop
)

var notePositionNames = [...]string{"LEFT", "RIGHT", "OVER", "OVER_SEVERAL", "BOTTOM", "TOP"}

// String returns upstream's constant name, for example "OVER_SEVERAL".
func (p NotePosition) String() string {
	if p < 0 || int(p) >= len(notePositionNames) {
		return "NotePosition(" + itoa(int(p)) + ")"
	}
	return notePositionNames[p]
}

// NotePositionValues returns the positions in upstream's order.
func NotePositionValues() []NotePosition {
	return []NotePosition{NotePositionLeft, NotePositionRight, NotePositionOver, NotePositionOverSeveral,
		NotePositionBottom, NotePositionTop}
}

// NotePositionValueOf is upstream's NotePosition.valueOf: the position of
// exactly that name. Another name panics, as Java throws
// IllegalArgumentException.
func NotePositionValueOf(name string) NotePosition {
	for i, n := range notePositionNames {
		if n == name {
			return NotePosition(i)
		}
	}
	panic(&color.RuntimeError{Class: illegalArgument,
		Message: "No enum constant net.sourceforge.plantuml.sequencediagram.NotePosition." + name})
}

// NotePositionDefaultLeft is upstream's NotePosition.defaultLeft: LEFT for a
// missing name (ok false, null upstream), else the position of the
// upper-cased name, which panics for an unknown name as upstream.
func NotePositionDefaultLeft(s string, ok bool) NotePosition {
	if !ok {
		return NotePositionLeft
	}
	return NotePositionValueOf(goUpperCase(s))
}

// goUpperCase is upstream's StringUtils.goUpperCase (toUpperCase with the
// English locale) for "ß" and the simple case mappings, which is enough
// for the names of an enum.
func goUpperCase(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "ß", "SS"))
}
