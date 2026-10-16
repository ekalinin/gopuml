// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityTriangle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryTriangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// triangle is upstream's ExtremityTriangle: a triangle of a given size,
// filled with the background color, the line color, or nothing.
type triangle struct {
	extremityBase
	polygon          *klimt.UPolygon
	fill             bool
	backgroundColor  color.HColor
	contact          geom.XPoint2D
	decorationLength int
}

func newTriangle(p1 geom.XPoint2D, angle float64, fill bool, backgroundColor color.HColor, xWing, yAperture,
	decorationLength int) *triangle {
	e := &triangle{polygon: klimt.NewUPolygon(), backgroundColor: backgroundColor, fill: fill,
		contact: geom.NewXPoint2D(p1.X, p1.Y), decorationLength: decorationLength}
	angle = manageround(angle)
	e.polygon.AddPoint(0, 0)
	e.polygon.AddPoint(float64(-xWing), float64(-yAperture))
	e.polygon.AddPoint(float64(-xWing), float64(yAperture))
	e.polygon.AddPoint(0, 0)
	e.polygon.Rotate(angle + pi/2)
	e.polygon = e.polygon.Translate(p1.X, p1.Y)
	return e
}

// SomePoint returns the contact point.
func (e *triangle) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns the length given to the factory.
func (e *triangle) DecorationLength() float64 { return float64(e.decorationLength) }

// DrawU draws the triangle.
func (e *triangle) DrawU(ug klimt.UGraphic) {
	if e.backgroundColor != nil {
		ug = ug.Apply(klimt.Bg(e.backgroundColor))
	} else if e.fill {
		ug = ug.Apply(klimt.ChangeBack(ug))
	}
	ug.Draw(e.polygon)
}

// FactoryTriangle is upstream's ExtremityFactoryTriangle.
type FactoryTriangle struct {
	abstractFactory
	backgroundColor  color.HColor
	xWing, yAperture int
	decorationLength int
}

// NewFactoryTriangle returns the factory of triangles; a nil background
// color fills them with the line color, or not at all for CreateUDrawable.
func NewFactoryTriangle(backgroundColor color.HColor, xWing, yAperture, decorationLength int) *FactoryTriangle {
	return &FactoryTriangle{abstractFactory{"ExtremityFactoryTriangle"}, backgroundColor, xWing, yAperture,
		decorationLength}
}

// CreateUDrawable returns the triangle at p0, not filled unless a background
// color was given.
func (f *FactoryTriangle) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newTriangle(p0, angle-pi/2, false, f.backgroundColor, f.xWing, f.yAperture, f.decorationLength)
}

// CreateTBRDrawableLegacy returns the triangle at p1, filled.
func (f *FactoryTriangle) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newTriangle(p1, ortho, true, f.backgroundColor, f.xWing, f.yAperture, f.decorationLength)
}
