// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityParenthesis.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityParenthesis2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryParenthesis.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// useInterfaceEye2 is upstream's GlobalConfig.USE_INTERFACE_EYE2, a false
// constant.
const useInterfaceEye2 = false

// lollipopInterfaceEye2Size is upstream's EntityImageLollipopInterfaceEye2.SIZE.
const lollipopInterfaceEye2Size = 14.0

// parenthesis is upstream's ExtremityParenthesis: an arc (a socket).
type parenthesis struct {
	extremityBase
	dest  geom.XPoint2D
	ortho float64
}

const (
	parenthesisRadius2 = 9.0
	parenthesisAng     = 70.0
)

func newParenthesis(p1 geom.XPoint2D, ortho float64) *parenthesis {
	return &parenthesis{dest: geom.NewXPoint2D(p1.X, p1.Y), ortho: ortho}
}

// SomePoint returns the center of the arc.
func (e *parenthesis) SomePoint() *geom.XPoint2D { return point(e.dest) }

// DecorationLength returns 10.
func (e *parenthesis) DecorationLength() float64 { return 10 }

// DrawU draws the arc.
func (e *parenthesis) DrawU(ug klimt.UGraphic) {
	deg := -e.ortho*180/pi + 90 - parenthesisAng
	arc1 := klimt.NewUEllipse(2*parenthesisRadius2, 2*parenthesisRadius2, deg, 2*parenthesisAng)
	ug.Apply(klimt.UStrokeWithThickness(1.5)).
		Apply(klimt.NewUTranslate(e.dest.X-parenthesisRadius2, e.dest.Y-parenthesisRadius2)).Draw(arc1)
}

// parenthesis2 is upstream's ExtremityParenthesis2: the socket of the
// "interface eye 2" style, an arc centered beyond p1. Upstream only builds
// it when GlobalConfig.USE_INTERFACE_EYE2 is set, which is never.
type parenthesis2 struct {
	extremityBase
	contact geom.XPoint2D
	center  geom.XPoint2D
	ortho   float64
}

const parenthesis2Ang = 30.0

func newParenthesis2(contact geom.XPoint2D, ortho float64, p1 geom.XPoint2D) *parenthesis2 {
	e := &parenthesis2{contact: geom.NewXPoint2D(contact.X, contact.Y), ortho: ortho}
	dx := p1.X - contact.X
	dy := p1.Y - contact.Y
	distance1 := float64(jmath.Round(contact.Distance(p1)))
	length := float64(jmath.Round(distance1 + lollipopInterfaceEye2Size/2))
	e.center = geom.NewXPoint2D(contact.X+float64(dx/distance1*length), contact.Y+float64(dy/distance1*length))
	return e
}

// SomePoint returns the contact point.
func (e *parenthesis2) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DrawU draws the arc.
func (e *parenthesis2) DrawU(ug klimt.UGraphic) {
	deg := -e.ortho*180/pi + 90 - parenthesis2Ang
	size := float64(jmath.Round(e.contact.Distance(e.center)))
	circle := klimt.NewUEllipse(size*2, size*2, deg, 2*parenthesis2Ang)
	ug.Apply(klimt.NewUTranslate(e.center.X-size, e.center.Y-size)).Draw(circle)
}

// FactoryParenthesis is upstream's ExtremityFactoryParenthesis.
type FactoryParenthesis struct{ abstractFactory }

// NewFactoryParenthesis returns the factory of sockets.
func NewFactoryParenthesis() *FactoryParenthesis {
	return &FactoryParenthesis{abstractFactory{"ExtremityFactoryParenthesis"}}
}

// CreateUDrawable returns the socket at p0.
func (f *FactoryParenthesis) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newParenthesis(p0, angle-pi/2)
}

// CreateTBRDrawableLegacy returns the socket at p1.
func (f *FactoryParenthesis) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	if useInterfaceEye2 {
		center := geom.NewXPoint2D((p0.X+p2.X)/2, (p0.Y+p2.Y)/2)
		return newParenthesis2(center, ortho, p1)
	}
	return newParenthesis(p1, ortho)
}
