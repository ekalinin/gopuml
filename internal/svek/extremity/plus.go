// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityPlus.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryPlus.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// plus is upstream's ExtremityPlus: a circled plus (a nested element).
type plus struct {
	extremityBase
	circle          *klimt.UEllipse
	px, py          float64
	angle           float64
	backgroundColor color.HColor
}

const plusRadius = 8.0

func newPlus(p1 geom.XPoint2D, angle float64, backgroundColor color.HColor) *plus {
	x := p1.X - plusRadius + float64(plusRadius*jmath.Sin(angle))
	y := p1.Y - plusRadius - float64(plusRadius*jmath.Cos(angle))
	return &plus{angle: angle, circle: klimt.UEllipseBuild(2*plusRadius, 2*plusRadius), px: x, py: y,
		backgroundColor: backgroundColor}
}

// SomePoint returns the corner of the circle.
func (e *plus) SomePoint() *geom.XPoint2D { return point(geom.NewXPoint2D(e.px, e.py)) }

// DecorationLength returns 16.
func (e *plus) DecorationLength() float64 { return 16 }

// DrawU draws the circle and the cross.
func (e *plus) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.Bg(e.backgroundColor)).Apply(klimt.NewUTranslate(e.px+0, e.py+0)).Draw(e.circle)
	drawLine(ug, 0, 0, e.pointOnCircle(e.angle-pi/2), e.pointOnCircle(e.angle+pi/2))
	drawLine(ug, 0, 0, e.pointOnCircle(e.angle), e.pointOnCircle(e.angle+pi))
}

func (e *plus) pointOnCircle(angle float64) geom.XPoint2D {
	x := e.px + plusRadius + float64(plusRadius*jmath.Cos(angle))
	y := e.py + plusRadius + float64(plusRadius*jmath.Sin(angle))
	return geom.NewXPoint2D(x, y)
}

// FactoryPlus is upstream's ExtremityFactoryPlus.
type FactoryPlus struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactoryPlus returns the factory of circled plus signs.
func NewFactoryPlus(backgroundColor color.HColor) *FactoryPlus {
	return &FactoryPlus{abstractFactory{"ExtremityFactoryPlus"}, backgroundColor}
}

// CreateUDrawable returns the decoration next to center.
func (f *FactoryPlus) CreateUDrawable(center geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newPlus(center, angle-pi/2, f.backgroundColor)
}

// CreateTBRDrawableLegacy returns the decoration next to p1.
func (f *FactoryPlus) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newPlus(p1, ortho, f.backgroundColor)
}
