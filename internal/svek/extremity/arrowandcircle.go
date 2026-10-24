// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityArrowAndCircle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryArrowAndCircle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// arrowAndCircle is upstream's ExtremityArrowAndCircle: an arrow head in
// front of a circle.
type arrowAndCircle struct {
	extremityBase
	polygon         *klimt.UPolygon
	contact         geom.XPoint2D
	dest            geom.XPoint2D
	backgroundColor color.HColor
}

const arrowAndCircleRadius = 5.0

func newArrowAndCircle(p1 geom.XPoint2D, angle float64, backgroundColor color.HColor) *arrowAndCircle {
	e := &arrowAndCircle{polygon: klimt.NewUPolygon()}
	angle = manageround(angle)
	e.polygon.AddPoint(0, 0)
	e.backgroundColor = backgroundColor
	e.dest = geom.NewXPoint2D(p1.X, p1.Y)
	const xWing = 9
	const yAperture = 4
	e.polygon.AddPoint(-xWing, -yAperture)
	const xContact = 5
	e.polygon.AddPoint(-xContact, 0)
	e.polygon.AddPoint(-xWing, yAperture)
	e.polygon.AddPoint(0, 0)
	e.polygon.Rotate(angle + pi/2)
	e.polygon = e.polygon.Translate(p1.X+float64(arrowAndCircleRadius*jmath.Sin(angle)),
		p1.Y-float64(arrowAndCircleRadius*jmath.Cos(angle)))
	e.contact = geom.NewXPoint2D(p1.X-float64(xContact*jmath.Cos(angle+pi/2)),
		p1.Y-float64(xContact*jmath.Sin(angle+pi/2)))
	return e
}

// SomePoint returns the contact point.
func (e *arrowAndCircle) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DrawU draws the head filled with the current color, then the circle.
func (e *arrowAndCircle) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.ChangeBack(ug)).Draw(e.polygon)
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.Bg(e.backgroundColor)).
		Apply(klimt.NewUTranslate(e.dest.X-arrowAndCircleRadius, e.dest.Y-arrowAndCircleRadius)).
		Draw(klimt.UEllipseBuild(arrowAndCircleRadius*2, arrowAndCircleRadius*2))
}

// FactoryArrowAndCircle is upstream's ExtremityFactoryArrowAndCircle.
type FactoryArrowAndCircle struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactoryArrowAndCircle returns the factory of arrow heads with a circle.
func NewFactoryArrowAndCircle(backgroundColor color.HColor) *FactoryArrowAndCircle {
	return &FactoryArrowAndCircle{abstractFactory{"ExtremityFactoryArrowAndCircle"}, backgroundColor}
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryArrowAndCircle) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newArrowAndCircle(p1, ortho, f.backgroundColor)
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryArrowAndCircle) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newArrowAndCircle(p0, angle-pi/2, f.backgroundColor)
}
