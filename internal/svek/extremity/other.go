// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityOther.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// other is upstream's ExtremityOther: the polygon that Graphviz drew for an
// arrow head, drawn as is.
type other struct {
	extremityBase
	polygon *klimt.UPolygon
}

// NewOther returns the decoration that draws polygon.
func NewOther(polygon *klimt.UPolygon) Extremity { return &other{polygon: polygon} }

// DrawU draws the polygon.
func (e *other) DrawU(ug klimt.UGraphic) { ug.Draw(e.polygon) }

// SomePoint returns the first point of the polygon; it panics for an empty
// polygon, as upstream throws IndexOutOfBoundsException.
func (e *other) SomePoint() *geom.XPoint2D { return point(polygonPoint(e.polygon, 0)) }
