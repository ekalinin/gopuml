// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Boundary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// boundaryLeft is the length of the horizontal bar of the boundary head.
const boundaryLeft = 17

// Boundary is the head of a boundary participant: a vertical bar joined to
// a circle. Upstream defines it in svek; it lives here because the skin
// components (skin.rose) and the symbols (decoration.symbol) draw it.
type Boundary struct {
	TextBlockDefaults
	symbolContext *Fashion
}

// NewBoundary returns the boundary head drawn with symbolContext.
func NewBoundary(symbolContext *Fashion) *Boundary { return &Boundary{symbolContext: symbolContext} }

// DrawU draws the bars and the circle.
func (b *Boundary) DrawU(ug UGraphic) {
	x := 0.0
	y := 0.0
	x += svekHeadMargin
	y += svekHeadMargin
	ug = b.symbolContext.Apply(ug)
	circle := UEllipseBuild(svekHeadRadius*2, svekHeadRadius*2)
	circle.SetDeltaShadow(b.symbolContext.DeltaShadow())

	// Upstream builds a first path that it never draws.
	path1 := UPathNone()
	path1.MoveTo(0, 0)
	path1.LineTo(0, svekHeadRadius*2)
	path1.SetDeltaShadow(b.symbolContext.DeltaShadow())

	path := UPathNone()
	path.MoveTo(0, 0)
	path.LineTo(0, svekHeadRadius*2)
	path.MoveTo(0, svekHeadRadius)
	path.LineTo(boundaryLeft, svekHeadRadius)
	path.SetDeltaShadow(b.symbolContext.DeltaShadow())
	ug.Apply(NewUTranslate(x, y)).Apply(Bg(color.None())).Draw(path)

	ug.Apply(NewUTranslate(x+boundaryLeft, y)).Draw(circle)
}

// CalculateDimension returns the size of the head with its margin.
func (b *Boundary) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(svekHeadRadius*2+boundaryLeft+2*svekHeadMargin, svekHeadRadius*2+2*svekHeadMargin)
}
