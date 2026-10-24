// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityCircleLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCircleLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// circleLine is upstream's ExtremityCircleLine: a bar with a circle (zero
// or one).
type circleLine struct {
	extremityBase
	contact geom.XPoint2D
	angle   float64
}

func newCircleLine(p1 geom.XPoint2D, angle float64) *circleLine {
	return &circleLine{contact: geom.NewXPoint2D(p1.X, p1.Y), angle: manageround(angle + pi/2)}
}

// SomePoint returns the contact point.
func (e *circleLine) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 15.
func (e *circleLine) DecorationLength() float64 { return 15 }

// DrawU draws the line to the circle, the circle and the bar, sized by the
// current stroke.
func (e *circleLine) DrawU(ug klimt.UGraphic) {
	thickness := ug.Param().Stroke().Thickness()
	radius := 4 + thickness - 1
	lineHeight := 4 + thickness - 1
	const xWing = 4
	middle := geom.NewXPoint2D(0, 0)
	base := rotate(geom.NewXPoint2D(-xWing-radius-3, 0), e.angle)
	circleBase := rotate(geom.NewXPoint2D(-xWing-radius-3, 0), e.angle)
	lineTop := rotate(geom.NewXPoint2D(-xWing, -lineHeight), e.angle)
	lineBottom := rotate(geom.NewXPoint2D(-xWing, lineHeight), e.angle)

	drawLine(ug, e.contact.X, e.contact.Y, base, middle)
	stroke := klimt.UStrokeWithThickness(thickness)
	ug.Apply(klimt.NewUTranslate(e.contact.X+circleBase.X-radius, e.contact.Y+circleBase.Y-radius)).
		Apply(stroke).Draw(klimt.UEllipseBuild(2*radius, 2*radius))
	drawLine(ug.Apply(stroke), e.contact.X, e.contact.Y, lineTop, lineBottom)
}

// FactoryCircleLine is upstream's ExtremityFactoryCircleLine.
type FactoryCircleLine struct{ abstractFactory }

// NewFactoryCircleLine returns the factory of bars with a circle.
func NewFactoryCircleLine() *FactoryCircleLine {
	return &FactoryCircleLine{abstractFactory{"ExtremityFactoryCircleLine"}}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryCircleLine) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newCircleLine(p0, angle)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryCircleLine) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newCircleLine(p1, ortho)
}
