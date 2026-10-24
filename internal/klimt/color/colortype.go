// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/ColorType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "strings"

// ColorType is the part of an element a color applies to.
type ColorType int

// Color types, in upstream's order.
const (
	TypeText ColorType = iota
	TypeLine
	TypeBack
	TypeHeader
	TypeArrow
	numColorTypes
)

var colorTypeNames = [numColorTypes]string{"TEXT", "LINE", "BACK", "HEADER", "ARROW"}

// String returns upstream's constant name, for example "BACK".
func (t ColorType) String() string {
	if t < 0 || t >= numColorTypes {
		return "ColorType(?)"
	}
	return colorTypeNames[t]
}

// ColorTypeOf is upstream's ColorType.getType: the type named by s up to its
// first dot, ignoring case, so "line.dashed" is TypeLine. As upstream's
// Enum.valueOf, it panics with an IllegalArgumentException RuntimeError when
// there is no such type.
func ColorTypeOf(s string) ColorType {
	if x := strings.IndexByte(s, '.'); x != -1 {
		s = s[:x]
	}
	upper := strings.ToUpper(s)
	for t, name := range colorTypeNames {
		if name == upper {
			return ColorType(t)
		}
	}
	throw(illegalArgument, "No enum constant net.sourceforge.plantuml.klimt.color.ColorType."+upper)
	return 0
}
