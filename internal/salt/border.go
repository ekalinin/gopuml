// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementBorder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// elementBorder is a border layout: an element at the top, the bottom, the
// right, the left and the center.
type elementBorder struct {
	north, south, east, west, center element
}

func newElementBorder() *elementBorder {
	return &elementBorder{north: &elementEmpty{}, south: &elementEmpty{}, east: &elementEmpty{},
		west: &elementEmpty{}, center: &elementEmpty{}}
}

func (e *elementBorder) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	stringBounder := ug.StringBounder()
	dimNorth := e.north.preferredDimension(stringBounder, 0, 0)
	dimSouth := e.south.preferredDimension(stringBounder, 0, 0)
	dimEast := e.east.preferredDimension(stringBounder, 0, 0)
	dimWest := e.west.preferredDimension(stringBounder, 0, 0)
	pA := geom.NewXPoint2D(dimWest.Width(), dimNorth.Height())
	pB := geom.NewXPoint2D(dimOf(dimToUse, "getWidth").Width()-dimEast.Width(), dimNorth.Height())
	pC := geom.NewXPoint2D(dimWest.Width(), dimToUse.Height()-dimSouth.Height())
	e.north.drawU(ug, zIndex, dimToUse)
	e.south.drawU(ug.Apply(klimt.UTranslateDy(pC.Y)), zIndex, dimToUse)
	e.west.drawU(ug.Apply(klimt.UTranslateDy(pA.Y)), zIndex, dimToUse)
	e.east.drawU(ug.Apply(klimt.NewUTranslate(pB.X, pB.Y)), zIndex, dimToUse)
	e.center.drawU(ug.Apply(klimt.NewUTranslate(pA.X, pA.Y)), zIndex, dimToUse)
}

func (e *elementBorder) preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	dimNorth := e.north.preferredDimension(stringBounder, x, y)
	dimSouth := e.south.preferredDimension(stringBounder, x, y)
	dimEast := e.east.preferredDimension(stringBounder, x, y)
	dimWest := e.west.preferredDimension(stringBounder, x, y)
	dimCenter := e.center.preferredDimension(stringBounder, x, y)
	width := max(dimNorth.Width(), dimSouth.Width(), dimWest.Width()+dimCenter.Width()+dimEast.Width())
	height := dimNorth.Height() + max(dimWest.Height(), dimCenter.Height(), dimEast.Height()) + dimSouth.Height()
	return geom.NewXDimension2D(width, height)
}

func (e *elementBorder) mayDrawBeyondPreferredDimension() bool {
	return e.north.mayDrawBeyondPreferredDimension() || e.south.mayDrawBeyondPreferredDimension() ||
		e.east.mayDrawBeyondPreferredDimension() || e.west.mayDrawBeyondPreferredDimension() ||
		e.center.mayDrawBeyondPreferredDimension()
}
