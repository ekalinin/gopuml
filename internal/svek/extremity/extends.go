// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityExtends.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryExtends.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// extends is upstream's ExtremityExtends: the hollow triangle of an
// inheritance.
type extends struct {
	extremityBase
	polygon *klimt.UPolygon
	fill    color.HColor
	contact geom.XPoint2D
}

func newExtends(p1 geom.XPoint2D, angle float64, backgroundColor color.HColor) *extends {
	e := &extends{polygon: klimt.NewUPolygon(), fill: backgroundColor, contact: geom.NewXPoint2D(p1.X, p1.Y)}
	angle = manageround(angle)
	e.polygon.AddPoint(0, 0)
	const xWing = 19.0    // 8 * 2.4
	const yAperture = 7.0 // 3 * 2.4
	e.polygon.AddPoint(-xWing, -yAperture)
	e.polygon.AddPoint(-xWing, yAperture)
	e.polygon.AddPoint(0, 0)
	e.polygon.Rotate(angle + pi/2)
	e.polygon = e.polygon.Translate(p1.X, p1.Y)
	return e
}

// SomePoint returns the contact point.
func (e *extends) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 18.
func (e *extends) DecorationLength() float64 { return 18 }

// DrawU draws the triangle filled with the background color.
func (e *extends) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.Bg(e.fill))
	ug.Draw(e.polygon)
}

// FactoryExtends is upstream's ExtremityFactoryExtends.
type FactoryExtends struct {
	abstractFactory
	backgroundColor color.HColor
}

// NewFactoryExtends returns the factory of inheritance triangles.
func NewFactoryExtends(backgroundColor color.HColor) Factory {
	return &FactoryExtends{abstractFactory{"ExtremityFactoryExtends"}, backgroundColor}
}

// CreateUDrawable returns the triangle at p0.
func (f *FactoryExtends) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newExtends(p0, angle-pi/2, f.backgroundColor)
}

// CreateTBRDrawableLegacy panics, as upstream throws
// UnsupportedOperationException.
func (f *FactoryExtends) CreateTBRDrawableLegacy(_, _, _ geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	unsupported()
	return nil
}
