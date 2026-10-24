// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityNotNavigable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryNotNavigable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// notNavigable is upstream's ExtremityNotNavigable: a cross on the line.
type notNavigable struct {
	extremityBase
	path    *klimt.UPath
	contact geom.XPoint2D
}

func newNotNavigable(p1 geom.XPoint2D, angle float64) *notNavigable {
	e := &notNavigable{path: klimt.UPathNone(), contact: geom.NewXPoint2D(p1.X, p1.Y)}
	angle = manageround(angle)
	const size = 4.0
	const move = 5.0
	e.path.MoveTo(-size, 0)
	e.path.LineTo(size, 2*size)
	e.path.MoveTo(size, 0)
	e.path.LineTo(-size, 2*size)
	e.path = e.path.Translate(0, move)
	e.path = e.path.Rotate(angle + pi)
	e.path = e.path.Translate(p1.X, p1.Y)
	return e
}

// SomePoint returns the contact point.
func (e *notNavigable) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 8.
func (e *notNavigable) DecorationLength() float64 { return 8 }

// DrawU draws the cross.
func (e *notNavigable) DrawU(ug klimt.UGraphic) { ug.Draw(e.path) }

// FactoryNotNavigable is upstream's ExtremityFactoryNotNavigable.
type FactoryNotNavigable struct{ abstractFactory }

// NewFactoryNotNavigable returns the factory of not navigable crosses.
func NewFactoryNotNavigable() *FactoryNotNavigable {
	return &FactoryNotNavigable{abstractFactory{"ExtremityFactoryNotNavigable"}}
}

// CreateUDrawable returns the cross at p0.
func (f *FactoryNotNavigable) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newNotNavigable(p0, angle-pi/2)
}

// CreateTBRDrawableLegacy returns the cross at p1.
func (f *FactoryNotNavigable) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newNotNavigable(p1, ortho)
}
