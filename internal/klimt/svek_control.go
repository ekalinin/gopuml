// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/Control.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The dimensions of the control, boundary and entity heads.
const (
	svekHeadMargin = 4
	svekHeadRadius = 12
)

// Control is the head of a control participant: a circle with an arrow
// head on top. Upstream defines it in svek; it lives here because the skin
// components (skin.rose) and the symbols (decoration.symbol) draw it.
type Control struct {
	TextBlockDefaults
	symbolContext *Fashion
}

// NewControl returns the control head drawn with symbolContext.
func NewControl(symbolContext *Fashion) *Control { return &Control{symbolContext: symbolContext} }

// DrawU draws the circle and the arrow head.
func (c *Control) DrawU(ug UGraphic) {
	x := 0.0
	y := 0.0
	x += svekHeadMargin
	y += svekHeadMargin
	ug = c.symbolContext.Apply(ug)
	circle := UEllipseBuild(svekHeadRadius*2, svekHeadRadius*2)
	circle.SetDeltaShadow(c.symbolContext.DeltaShadow())
	ug.Apply(NewUTranslate(x, y)).Draw(circle)
	ug = ug.Apply(UStrokeSimple())

	ug = ug.Apply(Bg(c.symbolContext.ForeColor()))
	polygon := NewUPolygon()
	polygon.AddPoint(0, 0)
	const xWing = 6
	const yAperture = 5
	polygon.AddPoint(xWing, -yAperture)
	const xContact = 4
	polygon.AddPoint(xContact, 0)
	polygon.AddPoint(xWing, yAperture)
	polygon.AddPoint(0, 0)

	ug.Apply(NewUTranslate(x+svekHeadRadius-xContact, y)).Draw(polygon)
}

// CalculateDimension returns the size of the head with its margin.
func (c *Control) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(svekHeadRadius*2+2*svekHeadMargin, svekHeadRadius*2+2*svekHeadMargin)
}
