// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/VerticalPosition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// VerticalPosition is a vertical position.
type VerticalPosition int

// Values of VerticalPosition, in upstream's order.
const (
	VerticalPositionTop VerticalPosition = iota
	VerticalPositionBottom
)

var verticalPositionNames = []string{"TOP", "BOTTOM"}

// String returns upstream's name, for example "TOP".
func (v VerticalPosition) String() string {
	return enumName("VerticalPosition", verticalPositionNames, v)
}

// VerticalPositionValueOf returns the value named exactly s, as Java's valueOf.
func VerticalPositionValueOf(s string) (VerticalPosition, bool) {
	return enumValueOf[VerticalPosition](verticalPositionNames, s)
}
