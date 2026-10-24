// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/GraphicPosition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// GraphicPosition is where a graphic is placed.
type GraphicPosition int

// Values of GraphicPosition, in upstream's order.
const (
	GraphicPositionBottom GraphicPosition = iota
	GraphicPositionBackgroundCornerBottomRight
	GraphicPositionBackgroundCornerTopRight
)

var graphicPositionNames = []string{"BOTTOM", "BACKGROUND_CORNER_BOTTOM_RIGHT", "BACKGROUND_CORNER_TOP_RIGHT"}

// String returns upstream's name, for example "BOTTOM".
func (v GraphicPosition) String() string { return enumName("GraphicPosition", graphicPositionNames, v) }

// GraphicPositionValueOf returns the value named exactly s, as Java's valueOf.
func GraphicPositionValueOf(s string) (GraphicPosition, bool) {
	return enumValueOf[GraphicPosition](graphicPositionNames, s)
}
