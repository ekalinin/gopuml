// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/Side.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// Side is a side of a rectangle.
type Side int

// Values of Side, in upstream's order.
const (
	SideNorth Side = iota
	SideEast
	SideSouth
	SideWest
)

var sideNames = []string{"NORTH", "EAST", "SOUTH", "WEST"}

// String returns upstream's name, for example "NORTH".
func (v Side) String() string { return enumName("Side", sideNames, v) }

// SideValueOf returns the value named exactly s, as Java's valueOf.
func SideValueOf(s string) (Side, bool) {
	return enumValueOf[Side](sideNames, s)
}
