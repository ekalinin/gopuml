// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityHalfArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryHalfArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// halfArrow is upstream's ExtremityHalfArrow: one wing of an arrow head,
// above or below the line depending on the direction.
type halfArrow struct {
	extremityBase
	line      *klimt.ULine
	otherLine *klimt.ULine
	contact   geom.XPoint2D
}

func newHalfArrowCenter(p1 geom.XPoint2D, angle float64, center geom.XPoint2D, direction int) *halfArrow {
	angle = manageround(angle)
	const xWing = 9
	yAperture := 4 * direction
	other := rotate(geom.NewXPoint2D(-xWing, float64(-yAperture)), angle+pi/2)
	e := &halfArrow{contact: p1}
	e.line = klimt.NewULine(center.X-e.contact.X, center.Y-e.contact.Y)
	e.otherLine = klimt.NewULine(other.X, other.Y)
	return e
}

func newHalfArrow(p0 geom.XPoint2D, angle float64, direction int) *halfArrow {
	angle = manageround(angle)
	const xWing = 9
	yAperture := 4 * direction
	e := &halfArrow{contact: p0}
	other := rotate(geom.NewXPoint2D(-xWing, float64(-yAperture)), angle)
	other2 := rotate(geom.NewXPoint2D(-8, 0), angle)
	e.line = klimt.NewULine(other.X, other.Y)
	e.otherLine = klimt.NewULine(other2.X, other2.Y)
	return e
}

// SomePoint returns the contact point.
func (e *halfArrow) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DrawU draws the two lines when the first is longer than 2.
func (e *halfArrow) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.ChangeBack(ug))
	if e.line != nil && e.line.Length() > 2 {
		position := klimt.UTranslatePoint(e.contact)
		ug.Apply(position).Draw(e.line)
		ug.Apply(position).Draw(e.otherLine)
	}
}

// FactoryHalfArrow is upstream's ExtremityFactoryHalfArrow.
type FactoryHalfArrow struct {
	abstractFactory
	direction int
}

// NewFactoryHalfArrow returns the factory of half arrow heads; direction is
// 1 (up) or -1 (down).
func NewFactoryHalfArrow(direction int) *FactoryHalfArrow {
	return &FactoryHalfArrow{abstractFactory{"ExtremityFactoryHalfArrow"}, direction}
}

// CreateUDrawable returns the half arrow head at p0.
func (f *FactoryHalfArrow) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newHalfArrow(p0, angle, f.direction)
}

// CreateTBRDrawableLegacy returns the half arrow head at p1, with a line to
// the middle of p0 and p2.
func (f *FactoryHalfArrow) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	center := geom.NewXPoint2D((p0.X+p2.X)/2, (p0.Y+p2.Y)/2)
	return newHalfArrowCenter(p1, ortho, center, f.direction)
}
