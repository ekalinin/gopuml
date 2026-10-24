// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UTranslate.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/geom"

// UTranslate is a translation, the change that moves what is drawn next. It
// is an immutable value.
type UTranslate struct {
	dx float64
	dy float64
}

// NewUTranslate returns the translation by (dx, dy).
func NewUTranslate(dx, dy float64) UTranslate { return UTranslate{dx: dx, dy: dy} }

// UTranslateNone returns the null translation.
func UTranslateNone() UTranslate { return UTranslate{} }

// UTranslateDx returns the horizontal translation by dx.
func UTranslateDx(dx float64) UTranslate { return UTranslate{dx: dx} }

// UTranslateDy returns the vertical translation by dy.
func UTranslateDy(dy float64) UTranslate { return UTranslate{dy: dy} }

// UTranslatePoint returns the translation by the coordinates of p.
func UTranslatePoint(p geom.XPoint2D) UTranslate { return UTranslate{dx: p.X, dy: p.Y} }

// String is upstream's toString, for example "translate dx=1.0 dy=2.0".
func (t UTranslate) String() string { return "translate dx=" + jstr(t.dx) + " dy=" + jstr(t.dy) }

// Dx returns the horizontal offset.
func (t UTranslate) Dx() float64 { return t.dx }

// Dy returns the vertical offset.
func (t UTranslate) Dy() float64 { return t.dy }

// IsAlmostSame reports whether either offset is equal to other's, as
// upstream does (with an "or").
func (t UTranslate) IsAlmostSame(other UTranslate) bool { return t.dx == other.dx || t.dy == other.dy }

// Translated returns p moved by the translation.
func (t UTranslate) Translated(p geom.XPoint2D) geom.XPoint2D {
	return geom.NewXPoint2D(p.X+t.dx, p.Y+t.dy)
}

// Scaled returns the translation multiplied by scale.
func (t UTranslate) Scaled(scale float64) UTranslate {
	return UTranslate{dx: t.dx * scale, dy: t.dy * scale}
}

// Compose returns the sum of the two translations.
func (t UTranslate) Compose(other UTranslate) UTranslate {
	return UTranslate{dx: t.dx + other.dx, dy: t.dy + other.dy}
}

// Reverse returns the opposite translation.
func (t UTranslate) Reverse() UTranslate { return UTranslate{dx: -t.dx, dy: -t.dy} }

// Apply returns rect moved by the translation.
func (t UTranslate) Apply(rect geom.XRectangle2D) geom.XRectangle2D {
	return geom.NewXRectangle2D(rect.X+t.dx, rect.Y+t.dy, rect.Width, rect.Height)
}

// MultiplyBy returns the translation multiplied by v.
func (t UTranslate) MultiplyBy(v float64) UTranslate { return UTranslate{dx: t.dx * v, dy: t.dy * v} }

// Sym returns the translation with dx and dy swapped.
func (t UTranslate) Sym() UTranslate { return UTranslate{dx: t.dy, dy: t.dx} }

// Position returns the offsets as a point.
func (t UTranslate) Position() geom.XPoint2D { return geom.NewXPoint2D(t.dx, t.dy) }

// Rotate returns the translation rotated by angle radians.
func (t UTranslate) Rotate(angle float64) UTranslate {
	rotate := XAffineTransformRotateInstance(angle)
	return UTranslatePoint(rotate.Transform(geom.NewXPoint2D(t.dx, t.dy)))
}
