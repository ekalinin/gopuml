// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityExtendsLike.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryExtendsLike.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The triangle of the extends-like decorations. They are variables so that
// the static products below are rounded as Java's double arithmetic does.
var (
	extendsLikeXLen      = -19.0 // 8 * 2.4
	extendsLikeHalfWidth = 7.0   // 3 * 2.4

	redefinesXSuffix = extendsLikeXLen * 1.2
	definedByXSuffix = extendsLikeXLen * 1.3
)

const definedByDotHSize = 2.0

// extendsLikePoint is upstream's ExtremityExtendsLike.Point, a mutable point.
type extendsLikePoint struct{ x, y float64 }

func (p *extendsLikePoint) rotate(theta float64) {
	ct := jmath.Cos(theta)
	st := -jmath.Sin(theta)
	nx := float64(p.x*ct) - float64(p.y*st)
	p.y = float64(-p.x*st) - float64(p.y*ct)
	p.x = nx
}

func (p *extendsLikePoint) pos(pt geom.XPoint2D) klimt.UTranslate {
	return klimt.NewUTranslate(p.x+pt.X, p.y+pt.Y)
}

func (p *extendsLikePoint) translate(pt geom.XPoint2D) {
	p.x += pt.X
	p.y += pt.Y
}

// extendsLike is upstream's abstract ExtremityExtendsLike: the triangle of a
// redefinition or a definition, completed by the subclasses.
type extendsLike struct {
	extremityBase
	trig    *klimt.UPolygon
	back    klimt.UBackground
	contact geom.XPoint2D
}

func addTrigPoint(up *klimt.UPolygon, x, y, angle float64, porig geom.XPoint2D) {
	p := &extendsLikePoint{x, y}
	p.rotate(angle)
	p.translate(porig)
	up.AddPoint(p.x, p.y)
}

func newExtendsLike(porig geom.XPoint2D, angle float64, backgroundColor color.HColor) extendsLike {
	e := extendsLike{back: klimt.Bg(backgroundColor), contact: geom.NewXPoint2D(porig.X, porig.Y)}
	angle = manageround(angle)
	e.trig = klimt.NewUPolygon()
	e.trig.AddPointPt(porig)
	addTrigPoint(e.trig, extendsLikeXLen, -extendsLikeHalfWidth, angle, porig)
	addTrigPoint(e.trig, extendsLikeXLen, +extendsLikeHalfWidth, angle, porig)
	e.trig.AddPointPt(porig)
	return e
}

// SomePoint returns the contact point.
func (e *extendsLike) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DecorationLength returns 18.
func (e *extendsLike) DecorationLength() float64 { return 18 }

func (e *extendsLike) drawTriangle(ug klimt.UGraphic) { ug.Apply(e.back).Draw(e.trig) }

// redefines is upstream's ExtremityExtendsLike.Redefines: the triangle with
// a bar behind it.
type redefines struct {
	extendsLike
	pos klimt.UTranslate
	bar *klimt.ULine
}

func newRedefines(porig geom.XPoint2D, angle float64, backgroundColor color.HColor) *redefines {
	e := &redefines{extendsLike: newExtendsLike(porig, angle, backgroundColor)}
	p1 := &extendsLikePoint{redefinesXSuffix, -extendsLikeHalfWidth}
	p2 := &extendsLikePoint{redefinesXSuffix, +extendsLikeHalfWidth}
	p1.rotate(angle)
	p2.rotate(angle)
	e.bar = klimt.NewULine(p2.x-p1.x, p2.y-p1.y)
	e.pos = p1.pos(porig)
	return e
}

// DrawU draws the triangle and the bar.
func (e *redefines) DrawU(ug klimt.UGraphic) {
	e.drawTriangle(ug)
	ug.Apply(klimt.UStrokeWithThickness(2.0)).Apply(e.pos).Draw(e.bar)
}

// definedBy is upstream's ExtremityExtendsLike.DefinedBy: the triangle with
// two dots behind it.
type definedBy struct {
	extendsLike
	pos1, pos2 klimt.UTranslate
	dot        *klimt.UEllipse
}

func definedByDotPos(x, y, angle, size float64, porig geom.XPoint2D) klimt.UTranslate {
	p := &extendsLikePoint{x, y}
	p.rotate(angle)
	p.x -= size
	p.y -= size
	return p.pos(porig)
}

func newDefinedBy(porig geom.XPoint2D, angle float64, backgroundColor color.HColor) *definedBy {
	e := &definedBy{extendsLike: newExtendsLike(porig, angle, backgroundColor)}
	w := extendsLikeHalfWidth - definedByDotHSize
	e.pos1 = definedByDotPos(definedByXSuffix, -w, angle, definedByDotHSize, porig)
	e.pos2 = definedByDotPos(definedByXSuffix, +w, angle, definedByDotHSize, porig)
	s := definedByDotHSize + definedByDotHSize
	e.dot = klimt.UEllipseBuild(s, s)
	return e
}

// DrawU draws the triangle and the two dots, filled with the current color.
func (e *definedBy) DrawU(ug klimt.UGraphic) {
	e.drawTriangle(ug)
	if c := ug.Param().Color(); c != nil {
		ug = ug.Apply(klimt.Bg(c))
	}
	ug.Apply(e.pos1).Draw(e.dot)
	ug.Apply(e.pos2).Draw(e.dot)
}

// FactoryExtendsLike is upstream's ExtremityFactoryExtendsLike.
type FactoryExtendsLike struct {
	abstractFactory
	backgroundColor color.HColor
	definedBy       bool
}

// NewFactoryExtendsLike returns the factory of redefinition (definedBy
// false) or definition triangles.
func NewFactoryExtendsLike(backgroundColor color.HColor, definedBy bool) *FactoryExtendsLike {
	return &FactoryExtendsLike{abstractFactory{"ExtremityFactoryExtendsLike"}, backgroundColor, definedBy}
}

// CreateUDrawable returns the decoration at p0.
func (f *FactoryExtendsLike) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	if f.definedBy {
		return newDefinedBy(p0, angle, f.backgroundColor)
	}
	return newRedefines(p0, angle, f.backgroundColor)
}

// CreateTBRDrawableLegacy returns the decoration at p1.
func (f *FactoryExtendsLike) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2) + (pi / 2.0)
	if f.definedBy {
		return newDefinedBy(p1, ortho, f.backgroundColor)
	}
	return newRedefines(p1, ortho, f.backgroundColor)
}
