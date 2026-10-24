// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityCircle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryCircle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// circle is upstream's ExtremityCircle: a circle, filled with the line color
// or with the background color.
type circle struct {
	extremityBase
	dest            geom.XPoint2D
	fill            bool
	backgroundColor color.HColor
}

const circleRadius = 6.0

func newCircle(center geom.XPoint2D, fill bool, angle float64, backgroundColor color.HColor) *circle {
	return &circle{
		dest: geom.NewXPoint2D(center.X-float64(circleRadius*jmath.Cos(angle+pi/2)),
			center.Y-float64(circleRadius*jmath.Sin(angle+pi/2))),
		backgroundColor: backgroundColor,
		fill:            fill,
	}
}

// SomePoint returns the center of the circle.
func (e *circle) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DecorationLength returns 12.
func (e *circle) DecorationLength() float64 { return 12 }

// DrawU draws the circle.
func (e *circle) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5))
	if e.fill {
		ug = ug.Apply(klimt.ChangeBack(ug))
	} else {
		ug = ug.Apply(klimt.Bg(e.backgroundColor))
	}
	ug = ug.Apply(klimt.NewUTranslate(e.dest.X-circleRadius, e.dest.Y-circleRadius))
	ug.Draw(klimt.UEllipseBuild(circleRadius*2, circleRadius*2))
}

// FactoryCircle is upstream's ExtremityFactoryCircle.
type FactoryCircle struct {
	abstractFactory
	fill            bool
	backgroundColor color.HColor
}

// NewFactoryCircle returns the factory of circles.
func NewFactoryCircle(fill bool, backgroundColor color.HColor) *FactoryCircle {
	return &FactoryCircle{abstractFactory{"ExtremityFactoryCircle"}, fill, backgroundColor}
}

// CreateUDrawable returns the circle next to center.
func (f *FactoryCircle) CreateUDrawable(center geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	angle -= pi / 2
	return newCircle(center, f.fill, angle, f.backgroundColor)
}

// CreateTBRDrawableLegacy returns the circle next to p1.
func (f *FactoryCircle) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newCircle(p1, f.fill, ortho, f.backgroundColor)
}
