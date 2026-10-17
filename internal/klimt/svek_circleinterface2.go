// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/CircleInterface2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The dimensions of CircleInterface2.
const (
	circleInterface2Margin = 1
	circleInterface2Radius = 8
)

// CircleInterface2 is the small circle of an interface. Upstream defines it
// in svek; it lives here because the interface symbol
// (decoration.symbol.USymbolInterface) draws it.
type CircleInterface2 struct {
	TextBlockDefaults
	backgroundColor color.HColor
	foregroundColor color.HColor
	deltaShadow     float64
}

// NewCircleInterface2 returns the circle with its colors and shadow.
func NewCircleInterface2(backgroundColor, foregroundColor color.HColor, deltaShadow float64) *CircleInterface2 {
	return &CircleInterface2{backgroundColor: backgroundColor, foregroundColor: foregroundColor,
		deltaShadow: deltaShadow}
}

// DrawU draws the circle after the margin.
func (c *CircleInterface2) DrawU(ug UGraphic) {
	x := 0.0
	y := 0.0
	x += circleInterface2Margin
	y += circleInterface2Margin
	ug = ug.Apply(Bg(c.backgroundColor)).Apply(c.foregroundColor)
	circle := UEllipseBuild(circleInterface2Radius*2, circleInterface2Radius*2)
	circle.SetDeltaShadow(c.deltaShadow)
	ug.Apply(NewUTranslate(x, y)).Draw(circle)
}

// CalculateDimension returns the circle with its margin.
func (c *CircleInterface2) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(circleInterface2Radius*2+2*circleInterface2Margin,
		circleInterface2Radius*2+2*circleInterface2Margin)
}
