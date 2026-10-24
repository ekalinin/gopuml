// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityCircleConnect.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCircleConnect.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// circleConnect is upstream's ExtremityCircleConnect: a circle with an arc
// around it (a ball and socket).
type circleConnect struct {
	extremityBase
	dest            geom.XPoint2D
	ortho           float64
	backgroundColor color.HColor
}

const (
	circleConnectRadius  = 6.0
	circleConnectRadius2 = 10.0
)

func newCircleConnect(p1 geom.XPoint2D, ortho float64, backgroundColor color.HColor) *circleConnect {
	return &circleConnect{dest: geom.NewXPoint2D(p1.X, p1.Y), ortho: ortho, backgroundColor: backgroundColor}
}

// SomePoint returns the center of the circle.
func (e *circleConnect) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DecorationLength returns 10.
func (e *circleConnect) DecorationLength() float64 { return 10 }

// DrawU draws the circle and the arc.
func (e *circleConnect) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.Bg(e.backgroundColor))
	ug.Apply(klimt.NewUTranslate(e.dest.X-circleConnectRadius, e.dest.Y-circleConnectRadius)).
		Draw(klimt.UEllipseBuild(circleConnectRadius*2, circleConnectRadius*2))

	deg := -e.ortho*180/pi + 90 - 45
	arc1 := klimt.NewUEllipse(2*circleConnectRadius2, 2*circleConnectRadius2, deg, 90)
	ug.Apply(klimt.NewUTranslate(e.dest.X-circleConnectRadius2, e.dest.Y-circleConnectRadius2)).Draw(arc1)
}

// FactoryCircleConnect is upstream's ExtremityFactoryCircleConnect.
type FactoryCircleConnect struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactoryCircleConnect returns the factory of balls and sockets.
func NewFactoryCircleConnect(backgroundColor color.HColor) *FactoryCircleConnect {
	return &FactoryCircleConnect{abstractFactory{"ExtremityFactoryCircleConnect"}, backgroundColor}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryCircleConnect) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newCircleConnect(p0, angle, f.backgroundColor)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryCircleConnect) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newCircleConnect(p1, ortho, f.backgroundColor)
}
