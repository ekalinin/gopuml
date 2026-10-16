// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityCircleCrowfoot.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCircleCrowfoot.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// circleCrowfoot is upstream's ExtremityCircleCrowfoot: a crow's foot with a
// circle (zero or many).
type circleCrowfoot struct {
	extremityBase
	contact geom.XPoint2D
	angle   float64
}

const circleCrowfootRadius = 4.0

func newCircleCrowfoot(p1 geom.XPoint2D, angle float64) *circleCrowfoot {
	return &circleCrowfoot{contact: geom.NewXPoint2D(p1.X, p1.Y), angle: manageround(angle + pi/2)}
}

// SomePoint returns the contact point.
func (e *circleCrowfoot) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 18.
func (e *circleCrowfoot) DecorationLength() float64 { return 18 }

// DrawU draws the three lines of the foot and the circle behind them.
func (e *circleCrowfoot) DrawU(ug klimt.UGraphic) {
	const xWing = 8
	const yAperture = 6
	middle := geom.NewXPoint2D(0, 0)
	left := rotate(geom.NewXPoint2D(0, -yAperture), e.angle)
	base := rotate(geom.NewXPoint2D(-xWing, 0), e.angle)
	right := rotate(geom.NewXPoint2D(0, yAperture), e.angle)
	circleBase := rotate(geom.NewXPoint2D(-xWing-circleCrowfootRadius-2, 0), e.angle)

	drawLine(ug, e.contact.X, e.contact.Y, base, left)
	drawLine(ug, e.contact.X, e.contact.Y, base, right)
	drawLine(ug, e.contact.X, e.contact.Y, base, middle)
	ug.Apply(klimt.NewUTranslate(e.contact.X+circleBase.X-circleCrowfootRadius,
		e.contact.Y+circleBase.Y-circleCrowfootRadius)).
		Draw(klimt.UEllipseBuild(2*circleCrowfootRadius, 2*circleCrowfootRadius))
}

// FactoryCircleCrowfoot is upstream's ExtremityFactoryCircleCrowfoot.
type FactoryCircleCrowfoot struct{ abstractFactory }

// NewFactoryCircleCrowfoot returns the factory of crow's feet with a circle.
func NewFactoryCircleCrowfoot() *FactoryCircleCrowfoot {
	return &FactoryCircleCrowfoot{abstractFactory{"ExtremityFactoryCircleCrowfoot"}}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryCircleCrowfoot) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newCircleCrowfoot(p0, angle)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryCircleCrowfoot) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newCircleCrowfoot(p1, ortho)
}
