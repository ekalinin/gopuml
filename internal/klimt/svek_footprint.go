// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/Footprint.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/ContainingEllipse.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/SmallestEnclosingCircle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/Circle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/YTransformer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/RotatedEllipse.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockInEllipse.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"reflect"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The types of this file are upstream's svek.image ellipse helpers, which
// compute the smallest ellipse around a text (use cases), and the working
// TextBlockInEllipse built on them. They live here because the use case
// symbol (decoration.symbol.USymbolUsecase) and the entity images (svek)
// both use them.

// Footprint collects the points of what a drawable draws. Upstream defines
// it in svek.image.
type Footprint struct {
	stringBounder font.StringBounder
}

// NewFootprint returns a footprint measuring with stringBounder.
func NewFootprint(stringBounder font.StringBounder) *Footprint {
	return &Footprint{stringBounder: stringBounder}
}

// footprintUGraphic is upstream's Footprint.MyUGraphic: it records the
// corners of the shapes drawn, in a list shared by the derived graphics.
type footprintUGraphic struct {
	UGraphicNo
	all *[]geom.XPoint2D
}

var _ UGraphic = (*footprintUGraphic)(nil)

// Apply returns a graphic sharing the points, translated. Changes other
// than colors, backgrounds, strokes and translations panic, as upstream
// throws UnsupportedOperationException.
func (g *footprintUGraphic) Apply(change UChange) UGraphic {
	tr := g.Translate()
	if t, ok := change.(UTranslate); ok {
		tr = g.Translate().Compose(t)
	}
	switch change.(type) {
	case UBackground, color.HColor, UStroke, UTranslate:
	default:
		throw(unsupportedOperation, "class "+reflect.TypeOf(change).String())
	}
	return &footprintUGraphic{UGraphicNo: NewUGraphicNo(g.StringBounder(), tr), all: g.all}
}

// Draw records the corners of shape; other shapes than texts, lines,
// images, paths, rectangles, ellipses and empty shapes panic, as upstream
// throws UnsupportedOperationException.
func (g *footprintUGraphic) Draw(shape UShape) {
	x := g.Translate().Dx()
	y := g.Translate().Dy()
	switch s := shape.(type) {
	case *UText:
		d := g.StringBounder().CalculateDimension(s.FontConfiguration().Font(), s.Text())
		y -= d.Height - 1.5
		g.addPoint(x, y)
		g.addPoint(x, y+d.Height)
		g.addPoint(x+d.Width, y)
		g.addPoint(x+d.Width, y+d.Height)
	case *UHorizontalLine:
		// Definitively a horizontal line
	case *ULine:
		// Probably a horizontal line
	case *UImage:
		g.addPoint(x, y)
		g.addPoint(x, y+float64(s.Height()))
		g.addPoint(x+float64(s.Width()), y)
		g.addPoint(x+float64(s.Width()), y+float64(s.Height()))
	case *UPath:
		g.addPoint(x+s.MinX(), y+s.MinY())
		g.addPoint(x+s.MaxX(), y+s.MaxY())
	case *URectangle:
		g.addPoint(x, y)
		g.addPoint(x+s.Width(), y+s.Height())
	case *UEllipse:
		g.addPoint(x, y)
		g.addPoint(x+s.Width(), y+s.Height())
	case *UEmpty:
		g.addPoint(x, y)
		g.addPoint(x+s.Width(), y+s.Height())
	default:
		throw(unsupportedOperation, "class "+reflect.TypeOf(shape).String())
	}
}

// ColorMapper returns the identity mapper.
func (g *footprintUGraphic) ColorMapper() *color.ColorMapper { return color.IdentityMapper }

func (g *footprintUGraphic) addPoint(x, y float64) { *g.all = append(*g.all, geom.NewXPoint2D(x, y)) }

// Ellipse returns the smallest ellipse of ratio alpha (height over width)
// around what drawable draws.
func (f *Footprint) Ellipse(drawable UDrawable, alpha float64) *ContainingEllipse {
	var all []geom.XPoint2D
	ug := &footprintUGraphic{UGraphicNo: NewUGraphicNo(f.stringBounder, UTranslateNone()), all: &all}
	drawable.DrawU(ug)
	circle := NewContainingEllipse(alpha)
	for _, pt := range all {
		circle.Append(pt)
	}
	return circle
}

// ContainingEllipse is the smallest ellipse of a given ratio around points.
// Upstream defines it in svek.image.
type ContainingEllipse struct {
	sec          smallestEnclosingCircle
	ytransformer yTransformer
	deltaShadow  float64
}

// NewContainingEllipse returns an empty ellipse whose height is coefY
// times its width.
func NewContainingEllipse(coefY float64) *ContainingEllipse {
	return &ContainingEllipse{ytransformer: newYTransformer(coefY)}
}

// String is upstream's toString.
func (c *ContainingEllipse) String() string {
	return "ContainingEllipse " + jstr(c.Width()) + " " + jstr(c.Height())
}

// Append adds a point to enclose.
func (c *ContainingEllipse) Append(pt geom.XPoint2D) {
	pt = c.ytransformer.reversePoint2D(pt)
	c.sec.append(pt)
}

// AppendXY adds the point (x, y) to enclose.
func (c *ContainingEllipse) AppendXY(x, y float64) { c.Append(geom.NewXPoint2D(x, y)) }

// Width returns the width of the ellipse.
func (c *ContainingEllipse) Width() float64 { return 2 * c.sec.circle().radius }

// Height returns the height of the ellipse.
func (c *ContainingEllipse) Height() float64 { return 2 * c.sec.circle().radius * c.ytransformer.alpha }

// Center returns the center of the ellipse.
func (c *ContainingEllipse) Center() geom.XPoint2D {
	return c.ytransformer.point2D(c.sec.circle().center)
}

// AsUEllipse returns the ellipse as a shape, with the shadow.
func (c *ContainingEllipse) AsUEllipse() *UEllipse {
	ellipse := UEllipseBuild(c.Width(), c.Height())
	ellipse.SetDeltaShadow(c.deltaShadow)
	return ellipse
}

// SetDeltaShadow sets the shadow of the shapes AsUEllipse returns.
func (c *ContainingEllipse) SetDeltaShadow(deltaShadow float64) { c.deltaShadow = deltaShadow }

// smallestEnclosingCircle is upstream's SmallestEnclosingCircle (Welzl's
// algorithm over the points in insertion order).
type smallestEnclosingCircle struct {
	all          []geom.XPoint2D
	lastSolution *svekCircle
}

func (s *smallestEnclosingCircle) append(pt geom.XPoint2D) {
	found := false
	for _, p := range s.all {
		if p.Equals(pt) {
			found = true
			break
		}
	}
	if !found {
		s.all = append(s.all, pt)
	}
	s.lastSolution = nil
}

func (s *smallestEnclosingCircle) circle() *svekCircle {
	if s.lastSolution == nil {
		s.lastSolution = findSec(len(s.all), s.all, 0, append([]geom.XPoint2D(nil), s.all...))
	}
	return s.lastSolution
}

func findSec(n int, p []geom.XPoint2D, m int, b []geom.XPoint2D) *svekCircle {
	sec := newSvekCircle(geom.NewXPoint2D(0, 0))

	// Compute the Smallest Enclosing Circle defined by B
	switch m {
	case 1:
		sec = newSvekCircle(b[0])
	case 2:
		sec = newSvekCircle2(b[0], b[1])
	case 3:
		return svekCircle3(b[0], b[1], b[2])
	}

	// Check if all the points in p are enclosed
	for i := range n {
		if sec.isOutside(p[i]) {
			// Compute B <--- B union P[i].
			b[m] = p[i]
			// Recurse
			sec = findSec(i, p, m+1, b)
		}
	}
	return sec
}

// svekCircle is upstream's svek.image.Circle.
type svekCircle struct {
	center geom.XPoint2D
	radius float64
}

func newSvekCircle(center geom.XPoint2D) *svekCircle { return &svekCircle{center: center} }

func newSvekCircle2(p1, p2 geom.XPoint2D) *svekCircle {
	center := geom.NewXPoint2D((p1.X+p2.X)/2, (p1.Y+p2.Y)/2)
	return &svekCircle{center: center, radius: p1.Distance(center)}
}

// svekCircle3 is upstream's Circle.getCircle: the circle through three
// points.
func svekCircle3(p1, p2, p3 geom.XPoint2D) *svekCircle {
	if p3.Y != p2.Y {
		return newSvekCircle3(p1, p2, p3)
	}
	return newSvekCircle3(p2, p1, p3)
}

func newSvekCircle3(p1, p2, p3 geom.XPoint2D) *svekCircle {
	num1 := float64(float64(p3.X*p3.X)*(p1.Y-p2.Y)) +
		float64((float64(p1.X*p1.X)+float64((p1.Y-p2.Y)*(p1.Y-p3.Y)))*(p2.Y-p3.Y)) +
		float64(float64(p2.X*p2.X)*(-p1.Y+p3.Y))
	den1 := 2 * (float64(p3.X*(p1.Y-p2.Y)) + float64(p1.X*(p2.Y-p3.Y)) + float64(p2.X*(-p1.Y+p3.Y)))
	x := num1 / den1
	den2 := p3.Y - p2.Y
	y := (p2.Y+p3.Y)/2 - float64((p3.X-p2.X)/den2*(x-(p2.X+p3.X)/2))
	center := geom.NewXPoint2D(x, y)
	return &svekCircle{center: center, radius: center.Distance(p1)}
}

func (c *svekCircle) isOutside(point geom.XPoint2D) bool {
	d := c.center.Distance(point)
	return d > c.radius
}

// yTransformer is upstream's YTransformer: it scales the y axis by alpha.
type yTransformer struct {
	alpha float64
}

func newYTransformer(alpha float64) yTransformer {
	if math.IsNaN(alpha) {
		return yTransformer{alpha: 1}
	}
	return yTransformer{alpha: alpha}
}

func (t yTransformer) point2D(pt geom.XPoint2D) geom.XPoint2D {
	return geom.NewXPoint2D(pt.X, pt.Y*t.alpha)
}

func (t yTransformer) reversePoint2D(pt geom.XPoint2D) geom.XPoint2D {
	return geom.NewXPoint2D(pt.X, pt.Y/t.alpha)
}

// RotatedEllipse is an ellipse rotated by beta radians. Upstream defines it
// in svek.image.
type RotatedEllipse struct {
	ellipse *UEllipse
	beta    float64
}

// NewRotatedEllipse returns the ellipse rotated by beta.
func NewRotatedEllipse(ellipse *UEllipse, beta float64) *RotatedEllipse {
	return &RotatedEllipse{ellipse: ellipse, beta: beta}
}

// A returns the semi-axis along x.
func (r *RotatedEllipse) A() float64 { return r.ellipse.Width() / 2 }

// B returns the semi-axis along y.
func (r *RotatedEllipse) B() float64 { return r.ellipse.Height() / 2 }

// Beta returns the rotation.
func (r *RotatedEllipse) Beta() float64 { return r.beta }

// Point returns the point of parameter theta.
func (r *RotatedEllipse) Point(theta float64) geom.XPoint2D {
	x := r.A() * jmath.Cos(theta)
	y := r.B() * jmath.Sin(theta)

	xp := float64(x*jmath.Cos(r.beta)) - float64(y*jmath.Sin(r.beta))
	yp := float64(x*jmath.Sin(r.beta)) + float64(y*jmath.Cos(r.beta))

	return geom.NewXPoint2D(xp, yp)
}

// OtherTheta returns the parameter of the other point of the ellipse with
// the same x as the point of parameter theta1.
func (r *RotatedEllipse) OtherTheta(theta1 float64) float64 {
	z := r.Point(theta1).X
	a := r.A() * jmath.Cos(r.beta)
	b := r.B() * jmath.Sin(r.beta)
	sum := 2 * a * z / (float64(a*a) + float64(b*b))
	other := sum - jmath.Cos(theta1)
	return -jmath.Acos(other)
}

// TextBlockInContainingEllipse is upstream's TextBlockInEllipse: a text in
// the smallest surrounding ellipse (use cases). TextBlockInEllipse of
// shape_textblockmisc.go is the stub of an earlier step; this type is the
// working one.
type TextBlockInContainingEllipse struct {
	TextBlockDefaults
	memo    TextBlockMemoized
	text    TextBlock
	ellipse *ContainingEllipse
}

// NewTextBlockInContainingEllipse returns the text in its ellipse, measured
// with stringBounder.
func NewTextBlockInContainingEllipse(text TextBlock, stringBounder font.StringBounder) *TextBlockInContainingEllipse {
	textDim := text.CalculateDimension(stringBounder)
	alpha := textDim.Height() / textDim.Width()
	if alpha < .2 {
		alpha = .2
	} else if alpha > .8 {
		alpha = .8
	}
	footprint := NewFootprint(stringBounder)
	return &TextBlockInContainingEllipse{text: text, ellipse: footprint.Ellipse(text, alpha)}
}

// UEllipse returns the ellipse grown by 6.
func (t *TextBlockInContainingEllipse) UEllipse() *UEllipse { return t.ellipse.AsUEllipse().Bigger(6) }

// DrawU draws the ellipse and the text centered in it.
func (t *TextBlockInContainingEllipse) DrawU(ug UGraphic) {
	sh := t.UEllipse()
	center := t.ellipse.Center()

	dx := sh.Width()/2 - center.X
	dy := sh.Height()/2 - center.Y

	ug.Draw(sh)

	t.text.DrawU(ug.Apply(NewUTranslate(dx, dy-2)))
}

// CalculateDimension returns the size of the ellipse, memoized.
func (t *TextBlockInContainingEllipse) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		return t.UEllipse().Dimension()
	})
}

// SetDeltaShadow sets the shadow of the ellipse.
func (t *TextBlockInContainingEllipse) SetDeltaShadow(deltaShadow float64) {
	t.ellipse.SetDeltaShadow(deltaShadow)
}
