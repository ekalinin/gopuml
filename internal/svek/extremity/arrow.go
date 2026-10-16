// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Arrow is upstream's ExtremityArrow: a filled arrow head, with a line from
// its contact point to the middle of the head that Graphviz drew.
type Arrow struct {
	extremityBase
	polygon *klimt.UPolygon
	line    *klimt.ULine
	contact geom.XPoint2D
}

var _ Extremity = (*Arrow)(nil)

// NewArrowCenter returns the arrow head at p1 along angle, with the line to
// center (upstream's three-argument constructor).
func NewArrowCenter(p1 geom.XPoint2D, angle float64, center geom.XPoint2D) *Arrow {
	a := &Arrow{polygon: klimt.NewUPolygon()}
	angle = manageround(angle)
	xContact := float64(a.buildPolygon())
	a.polygon.Rotate(angle + pi/2)
	a.polygon = a.polygon.Translate(p1.X, p1.Y)
	a.contact = geom.NewXPoint2D(p1.X-float64(xContact*jmath.Cos(angle+pi/2)),
		p1.Y-float64(xContact*jmath.Sin(angle+pi/2)))
	a.line = klimt.NewULine(center.X-a.contact.X, center.Y-a.contact.Y)
	return a
}

// NewArrow returns the arrow head at p0 along angle, without line.
func NewArrow(p0 geom.XPoint2D, angle float64) *Arrow {
	a := &Arrow{polygon: klimt.NewUPolygon()}
	angle = manageround(angle)
	a.buildPolygon()
	a.polygon.Rotate(angle)
	a.polygon = a.polygon.Translate(p0.X, p0.Y)
	a.contact = p0
	return a
}

func (a *Arrow) buildPolygon() int {
	a.polygon.AddPoint(0, 0)
	const xWing = 9
	const yAperture = 4
	a.polygon.AddPoint(-xWing, -yAperture)
	const xContact = 5
	a.polygon.AddPoint(-xContact, 0)
	a.polygon.AddPoint(-xWing, yAperture)
	a.polygon.AddPoint(0, 0)
	return xContact
}

// SomePoint returns the contact point.
func (a *Arrow) SomePoint() *geom.XPoint2D { return point(a.contact) }

// DecorationLength returns 5.
func (a *Arrow) DecorationLength() float64 { return 5 }

// DrawU fills the head with the current color and draws the line when it is
// longer than 2.
func (a *Arrow) DrawU(ug klimt.UGraphic) {
	if c := ug.Param().Color(); c == nil {
		ug = ug.Apply(klimt.Bg(color.None()))
	} else {
		ug = ug.Apply(klimt.Bg(c))
	}
	ug.Draw(a.polygon)
	if a.line != nil && a.line.Length() > 2 {
		ug.Apply(klimt.UTranslatePoint(a.contact)).Draw(a.line)
	}
}

// DrawLineIfTransparent draws the line from the tip of the head to its
// contact point, used when the head color is transparent.
func (a *Arrow) DrawLineIfTransparent(ug klimt.UGraphic) {
	pt1 := polygonPoint(a.polygon, 0)
	pt2 := polygonPoint(a.polygon, 2)
	line := klimt.ULineCreate(pt1, pt2)
	ug.Apply(klimt.UTranslatePoint(pt1)).Draw(line)
}

// FactoryArrow is upstream's ExtremityFactoryArrow.
type FactoryArrow struct{ abstractFactory }

// NewFactoryArrow returns the factory of arrow heads.
func NewFactoryArrow() *FactoryArrow {
	return &FactoryArrow{abstractFactory{"ExtremityFactoryArrow"}}
}

// CreateUDrawable returns the arrow head at p0.
func (f *FactoryArrow) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return NewArrow(p0, angle)
}

// CreateTBRDrawableLegacy returns the arrow head at p1, with a line to the
// middle of p0 and p2.
func (f *FactoryArrow) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	center := geom.NewXPoint2D((p0.X+p2.X)/2, (p0.Y+p2.Y)/2)
	return NewArrowCenter(p1, ortho, center)
}
