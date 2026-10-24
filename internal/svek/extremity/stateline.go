// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityStateLine1.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityStateLine2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// stateArrowPolygon builds the arrow head of the state line extremities.
func stateArrowPolygon(angle float64, center geom.XPoint2D) *klimt.UPolygon {
	polygon := klimt.NewUPolygon()
	polygon.AddPoint(0, 0)
	const xWing = 9
	const yAperture = 4
	polygon.AddPoint(-xWing, -yAperture)
	const xContact = 5
	polygon.AddPoint(-xContact, 0)
	polygon.AddPoint(-xWing, yAperture)
	polygon.AddPoint(0, 0)
	polygon.Rotate(angle)
	return polygon.Translate(center.X, center.Y)
}

// StateLine1 is upstream's ExtremityStateLine1: an arrow head in front of a
// crossed circle. Upstream does not use it.
type StateLine1 struct {
	extremityBase
	polygon *klimt.UPolygon
	dest    geom.XPoint2D
	angle   float64
}

const stateLine1Radius = 7.0

// NewStateLine1 returns the decoration at center along angle.
func NewStateLine1(angle float64, center geom.XPoint2D) *StateLine1 {
	e := &StateLine1{angle: manageround(angle), dest: geom.NewXPoint2D(center.X, center.Y)}
	e.polygon = stateArrowPolygon(e.angle, center)
	return e
}

// SomePoint returns the center of the circle.
func (e *StateLine1) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DrawU draws the head, the circle and its diagonals.
func (e *StateLine1) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.Bg(ug.Param().Color())).
		Apply(klimt.NewUTranslate(float64(-stateLine1Radius*jmath.Cos(e.angle)),
			float64(-stateLine1Radius*jmath.Sin(e.angle)))).Draw(e.polygon)
	ug = ug.Apply(klimt.Bg(color.White))
	ug.Apply(klimt.UStrokeWithThickness(1.5)).
		Apply(klimt.NewUTranslate(e.dest.X-stateLine1Radius, e.dest.Y-stateLine1Radius)).
		Draw(klimt.UEllipseBuild(stateLine1Radius*2, stateLine1Radius*2))
	stateDrawLine(ug, e.pointOnCircle(e.dest.X, e.dest.Y, pi/4), e.pointOnCircle(e.dest.X, e.dest.Y, pi+pi/4))
	stateDrawLine(ug, e.pointOnCircle(e.dest.X, e.dest.Y, -pi/4), e.pointOnCircle(e.dest.X, e.dest.Y, pi-pi/4))
}

func (e *StateLine1) pointOnCircle(centerX, centerY, angle float64) geom.XPoint2D {
	x := centerX + float64(stateLine1Radius*jmath.Cos(angle))
	y := centerY + float64(stateLine1Radius*jmath.Sin(angle))
	return geom.NewXPoint2D(x, y)
}

func stateDrawLine(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	ug.Apply(klimt.NewUTranslate(p1.X, p1.Y)).Draw(klimt.NewULine(dx, dy))
}

// StateLine2 is upstream's ExtremityStateLine2: an arrow head in front of a
// circle. Upstream does not use it.
type StateLine2 struct {
	extremityBase
	polygon *klimt.UPolygon
	dest    geom.XPoint2D
	angle   float64
}

const stateLine2Radius = 5.0

// NewStateLine2 returns the decoration at center along angle.
func NewStateLine2(angle float64, center geom.XPoint2D) *StateLine2 {
	e := &StateLine2{angle: manageround(angle), dest: geom.NewXPoint2D(center.X, center.Y)}
	e.polygon = stateArrowPolygon(e.angle, center)
	return e
}

// SomePoint returns the center of the circle.
func (e *StateLine2) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DrawU draws the head and the circle.
func (e *StateLine2) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.Bg(ug.Param().Color())).
		Apply(klimt.NewUTranslate(float64(-stateLine2Radius*jmath.Cos(e.angle)),
			float64(-stateLine2Radius*jmath.Sin(e.angle)))).Draw(e.polygon)
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.Bg(color.White)).
		Apply(klimt.NewUTranslate(e.dest.X-stateLine2Radius, e.dest.Y-stateLine2Radius)).
		Draw(klimt.UEllipseBuild(stateLine2Radius*2, stateLine2Radius*2))
}
