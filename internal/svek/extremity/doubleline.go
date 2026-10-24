// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityDoubleLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryDoubleLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// doubleLine is upstream's ExtremityDoubleLine: two bars (exactly one).
type doubleLine struct {
	extremityBase
	contact geom.XPoint2D
	angle   float64
}

const doubleLineHeight = 4.0

func newDoubleLine(p1 geom.XPoint2D, angle float64) *doubleLine {
	return &doubleLine{contact: geom.NewXPoint2D(p1.X, p1.Y), angle: manageround(angle + pi/2)}
}

// SomePoint returns the contact point.
func (e *doubleLine) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 8.
func (e *doubleLine) DecorationLength() float64 { return 8 }

// DrawU draws the two bars and the line to them.
func (e *doubleLine) DrawU(ug klimt.UGraphic) {
	const xWing = 4
	firstLineTop := rotate(geom.NewXPoint2D(-xWing, -doubleLineHeight), e.angle)
	firstLineBottom := rotate(geom.NewXPoint2D(-xWing, doubleLineHeight), e.angle)
	secondLineTop := rotate(geom.NewXPoint2D(-xWing-3, -doubleLineHeight), e.angle)
	secondLineBottom := rotate(geom.NewXPoint2D(-xWing-3, doubleLineHeight), e.angle)
	middle := rotate(geom.NewXPoint2D(0, 0), e.angle)
	base := rotate(geom.NewXPoint2D(-xWing-4, 0), e.angle)

	drawLine(ug, e.contact.X, e.contact.Y, firstLineTop, firstLineBottom)
	drawLine(ug, e.contact.X, e.contact.Y, secondLineTop, secondLineBottom)
	drawLine(ug, e.contact.X, e.contact.Y, base, middle)
}

// FactoryDoubleLine is upstream's ExtremityFactoryDoubleLine.
type FactoryDoubleLine struct{ abstractFactory }

// NewFactoryDoubleLine returns the factory of double bars.
func NewFactoryDoubleLine() *FactoryDoubleLine {
	return &FactoryDoubleLine{abstractFactory{"ExtremityFactoryDoubleLine"}}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryDoubleLine) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newDoubleLine(p0, angle)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryDoubleLine) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newDoubleLine(p1, ortho)
}
