// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremitySquare.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactorySquare.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// square is upstream's ExtremitySquare: a small square.
type square struct {
	extremityBase
	backgroundColor color.HColor
	dest            geom.XPoint2D
}

const squareRadius = 5.0

func newSquare(p1 geom.XPoint2D, backgroundColor color.HColor) *square {
	return &square{dest: geom.NewXPoint2D(p1.X, p1.Y), backgroundColor: backgroundColor}
}

// SomePoint returns the center of the square.
func (e *square) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DecorationLength returns 5.
func (e *square) DecorationLength() float64 { return 5 }

// DrawU draws the square.
func (e *square) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.Bg(e.backgroundColor)).
		Apply(klimt.NewUTranslate(e.dest.X-squareRadius, e.dest.Y-squareRadius)).
		Draw(klimt.URectangleBuild(squareRadius*2, squareRadius*2))
}

// FactorySquare is upstream's ExtremityFactorySquare.
type FactorySquare struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactorySquare returns the factory of squares.
func NewFactorySquare(backgroundColor color.HColor) *FactorySquare {
	return &FactorySquare{abstractFactory{"ExtremityFactorySquare"}, backgroundColor}
}

// CreateUDrawable returns the square at p0.
func (f *FactorySquare) CreateUDrawable(p0 geom.XPoint2D, _ float64, _ *geom.Side) klimt.UDrawable {
	return newSquare(p0, f.backgroundColor)
}

// CreateTBRDrawableLegacy returns the square at p1.
func (f *FactorySquare) CreateTBRDrawableLegacy(_, p1, _ geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	return newSquare(p1, f.backgroundColor)
}
