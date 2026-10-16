// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityLineCrowfoot.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryLineCrowfoot.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// lineCrowfoot is upstream's ExtremityLineCrowfoot: a crow's foot with a bar
// (one or many).
type lineCrowfoot struct {
	extremityBase
	contact geom.XPoint2D
	angle   float64
}

const lineCrowfootHeight = 4.0

func newLineCrowfoot(p1 geom.XPoint2D, angle float64) *lineCrowfoot {
	return &lineCrowfoot{contact: geom.NewXPoint2D(p1.X, p1.Y), angle: manageround(angle + pi/2)}
}

// SomePoint returns the contact point.
func (e *lineCrowfoot) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 8.
func (e *lineCrowfoot) DecorationLength() float64 { return 8 }

// DrawU draws the three lines of the foot and the bar.
func (e *lineCrowfoot) DrawU(ug klimt.UGraphic) {
	const xWing = 8
	const yAperture = 6
	middle := geom.NewXPoint2D(0, 0)
	left := rotate(geom.NewXPoint2D(0, -yAperture), e.angle)
	base := rotate(geom.NewXPoint2D(-xWing, 0), e.angle)
	lineTop := rotate(geom.NewXPoint2D(-xWing-2, -lineCrowfootHeight), e.angle)
	lineBottom := rotate(geom.NewXPoint2D(-xWing-2, lineCrowfootHeight), e.angle)
	right := rotate(geom.NewXPoint2D(0, yAperture), e.angle)

	drawLine(ug, e.contact.X, e.contact.Y, base, left)
	drawLine(ug, e.contact.X, e.contact.Y, base, right)
	drawLine(ug, e.contact.X, e.contact.Y, base, middle)
	drawLine(ug, e.contact.X, e.contact.Y, lineTop, lineBottom)
}

// FactoryLineCrowfoot is upstream's ExtremityFactoryLineCrowfoot.
type FactoryLineCrowfoot struct{ abstractFactory }

// NewFactoryLineCrowfoot returns the factory of crow's feet with a bar.
func NewFactoryLineCrowfoot() *FactoryLineCrowfoot {
	return &FactoryLineCrowfoot{abstractFactory{"ExtremityFactoryLineCrowfoot"}}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryLineCrowfoot) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newLineCrowfoot(p0, angle)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryLineCrowfoot) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newLineCrowfoot(p1, ortho)
}
