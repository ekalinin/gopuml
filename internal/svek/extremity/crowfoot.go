// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityCrowfoot.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCrowfoot.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// crowfoot is upstream's ExtremityCrowfoot: a crow's foot (many). When the
// side of the entity is known, the two outer lines end on that side.
type crowfoot struct {
	extremityBase
	contact geom.XPoint2D
	angle   float64
	side    *geom.Side
}

func newCrowfoot(p1 geom.XPoint2D, angle float64, side *geom.Side) *crowfoot {
	return &crowfoot{contact: geom.NewXPoint2D(p1.X, p1.Y), angle: manageround(angle + pi/2), side: side}
}

// SomePoint returns the contact point.
func (e *crowfoot) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 8.
func (e *crowfoot) DecorationLength() float64 { return 8 }

// DrawU draws the three lines of the foot.
func (e *crowfoot) DrawU(ug klimt.UGraphic) {
	const xWing = 8
	const yAperture = 8
	middle := geom.NewXPoint2D(0, 0)
	left := rotate(geom.NewXPoint2D(0, -yAperture), e.angle)
	base := rotate(geom.NewXPoint2D(-xWing, 0), e.angle)
	right := rotate(geom.NewXPoint2D(0, yAperture), e.angle)

	if e.side != nil && (*e.side == geom.SideWest || *e.side == geom.SideEast) {
		left = geom.NewXPoint2D(middle.X, left.Y)
		right = geom.NewXPoint2D(middle.X, right.Y)
	}
	if e.side != nil && (*e.side == geom.SideSouth || *e.side == geom.SideNorth) {
		left = geom.NewXPoint2D(left.X, middle.Y)
		right = geom.NewXPoint2D(right.X, middle.Y)
	}

	drawLine(ug, e.contact.X, e.contact.Y, base, left)
	drawLine(ug, e.contact.X, e.contact.Y, base, right)
	drawLine(ug, e.contact.X, e.contact.Y, base, middle)
}

// FactoryCrowfoot is upstream's ExtremityFactoryCrowfoot.
type FactoryCrowfoot struct{ abstractFactory }

// NewFactoryCrowfoot returns the factory of crow's feet.
func NewFactoryCrowfoot() *FactoryCrowfoot {
	return &FactoryCrowfoot{abstractFactory{"ExtremityFactoryCrowfoot"}}
}

// CreateTBRDrawableLegacy returns the crow's foot at p1.
func (f *FactoryCrowfoot) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, side *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newCrowfoot(p1, ortho, side)
}

// CreateUDrawable returns the crow's foot at p0.
func (f *FactoryCrowfoot) CreateUDrawable(p0 geom.XPoint2D, angle float64, side *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newCrowfoot(p0, angle, side)
}
