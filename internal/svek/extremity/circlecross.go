// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityCircleCross.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCircleCross.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// circleCross is upstream's ExtremityCircleCross: a crossed circle.
type circleCross struct {
	extremityBase
	px, py          float64
	dest            geom.XPoint2D
	backgroundColor color.HColor
}

const circleCrossRadius = 7.0

func newCircleCross(p1 geom.XPoint2D, backgroundColor color.HColor) *circleCross {
	return &circleCross{px: p1.X - circleCrossRadius, py: p1.Y - circleCrossRadius,
		dest: geom.NewXPoint2D(p1.X, p1.Y), backgroundColor: backgroundColor}
}

// SomePoint returns the center of the circle.
func (e *circleCross) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DrawU draws the circle and its two diagonals.
func (e *circleCross) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.Bg(e.backgroundColor))
	ug.Apply(klimt.UStrokeWithThickness(1.5)).
		Apply(klimt.NewUTranslate(e.dest.X-circleCrossRadius, e.dest.Y-circleCrossRadius)).
		Draw(klimt.UEllipseBuild(circleCrossRadius*2, circleCrossRadius*2))
	drawLine(ug, 0, 0, e.pointOnCircle(pi/4), e.pointOnCircle(pi+pi/4))
	drawLine(ug, 0, 0, e.pointOnCircle(-pi/4), e.pointOnCircle(pi-pi/4))
}

func (e *circleCross) pointOnCircle(angle float64) geom.XPoint2D {
	x := e.px + circleCrossRadius + float64(circleCrossRadius*jmath.Cos(angle))
	y := e.py + circleCrossRadius + float64(circleCrossRadius*jmath.Sin(angle))
	return geom.NewXPoint2D(x, y)
}

// FactoryCircleCross is upstream's ExtremityFactoryCircleCross.
type FactoryCircleCross struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactoryCircleCross returns the factory of crossed circles.
func NewFactoryCircleCross(backgroundColor color.HColor) *FactoryCircleCross {
	return &FactoryCircleCross{abstractFactory{"ExtremityFactoryCircleCross"}, backgroundColor}
}

// CreateTBRDrawableLegacy returns the crossed circle at p1.
func (f *FactoryCircleCross) CreateTBRDrawableLegacy(_, p1, _ geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	return newCircleCross(p1, f.backgroundColor)
}

// CreateUDrawable returns the crossed circle at p0.
func (f *FactoryCircleCross) CreateUDrawable(p0 geom.XPoint2D, _ float64, _ *geom.Side) klimt.UDrawable {
	return newCircleCross(p0, f.backgroundColor)
}
