// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/awt/XAffineTransform.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/XPoint2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// XAffineTransform is an affine transformation matrix:
//
//	[ m00  m01  m02 ]   [ x ]   [ m00*x + m01*y + m02 ]
//	[ m10  m11  m12 ] * [ y ] = [ m10*x + m11*y + m12 ]
//	[  0    0    1  ]   [ 1 ]   [          1          ]
//
// Upstream keeps it in klimt.awt; it lives here because XPoint2D.transform,
// UTranslate, ULine, UPolygon and UPath need it and the geom package must
// not depend on this one. Upstream's toAffineTransform (java.awt) belongs to
// the raster driver.
type XAffineTransform struct {
	m00, m10, m01, m11, m02, m12 float64
}

// NewXAffineTransform returns the transform of the 6 matrix values.
func NewXAffineTransform(m00, m10, m01, m11, m02, m12 float64) *XAffineTransform {
	return &XAffineTransform{m00: m00, m10: m10, m01: m01, m11: m11, m02: m02, m12: m12}
}

// NewXAffineTransformFlat returns the transform of [m00, m10, m01, m11, m02,
// m12].
func NewXAffineTransformFlat(flat []float64) *XAffineTransform {
	return NewXAffineTransform(flat[0], flat[1], flat[2], flat[3], flat[4], flat[5])
}

// NewXAffineTransformIdentity returns the identity transform.
func NewXAffineTransformIdentity() *XAffineTransform { return NewXAffineTransform(1, 0, 0, 1, 0, 0) }

// Copy returns a copy of the transform.
func (t *XAffineTransform) Copy() *XAffineTransform {
	c := *t
	return &c
}

// XAffineTransformRotateInstance returns the rotation by thetaRadians.
func XAffineTransformRotateInstance(thetaRadians float64) *XAffineTransform {
	cos := jmath.Cos(thetaRadians)
	sin := jmath.Sin(thetaRadians)
	return NewXAffineTransform(cos, sin, -sin, cos, 0.0, 0.0)
}

// XAffineTransformTranslateInstance returns the translation by (tx, ty).
func XAffineTransformTranslateInstance(tx, ty float64) *XAffineTransform {
	return NewXAffineTransform(1.0, 0.0, 0.0, 1.0, tx, ty)
}

// XAffineTransformScaleInstance returns the scaling by (sx, sy).
func XAffineTransformScaleInstance(sx, sy float64) *XAffineTransform {
	return NewXAffineTransform(sx, 0.0, 0.0, sy, 0.0, 0.0)
}

// ScaleX returns m00.
func (t *XAffineTransform) ScaleX() float64 { return t.m00 }

// ScaleY returns m11.
func (t *XAffineTransform) ScaleY() float64 { return t.m11 }

// TranslateX returns m02.
func (t *XAffineTransform) TranslateX() float64 { return t.m02 }

// TranslateY returns m12.
func (t *XAffineTransform) TranslateY() float64 { return t.m12 }

// Scale concatenates a scaling to the transform.
func (t *XAffineTransform) Scale(sx, sy float64) {
	t.m00 *= sx
	t.m01 *= sx
	t.m02 *= sx
	t.m10 *= sy
	t.m11 *= sy
	t.m12 *= sy
}

// Transform returns src transformed. It is also upstream's
// XPoint2D.transform.
func (t *XAffineTransform) Transform(src geom.XPoint2D) geom.XPoint2D {
	x := float64(t.m00*src.X) + float64(t.m01*src.Y) + t.m02
	y := float64(t.m10*src.X) + float64(t.m11*src.Y) + t.m12
	return geom.NewXPoint2D(x, y)
}

// Concatenate sets the transform to [this] * [other].
func (t *XAffineTransform) Concatenate(other *XAffineTransform) {
	n00 := float64(t.m00*other.m00) + float64(t.m01*other.m10)
	n01 := float64(t.m00*other.m01) + float64(t.m01*other.m11)
	n02 := float64(t.m00*other.m02) + float64(t.m01*other.m12) + t.m02
	n10 := float64(t.m10*other.m00) + float64(t.m11*other.m10)
	n11 := float64(t.m10*other.m01) + float64(t.m11*other.m11)
	n12 := float64(t.m10*other.m02) + float64(t.m11*other.m12) + t.m12
	t.m00, t.m01, t.m02 = n00, n01, n02
	t.m10, t.m11, t.m12 = n10, n11, n12
}

// Translate concatenates a translation to the transform.
func (t *XAffineTransform) Translate(tx, ty float64) {
	t.m02 += float64(t.m00*tx) + float64(t.m01*ty)
	t.m12 += float64(t.m10*tx) + float64(t.m11*ty)
}

// Rotate concatenates a rotation by thetaRadians around (anchorX, anchorY).
func (t *XAffineTransform) Rotate(thetaRadians, anchorX, anchorY float64) {
	t.Translate(anchorX, anchorY)
	cos := jmath.Cos(thetaRadians)
	sin := jmath.Sin(thetaRadians)
	n00 := float64(t.m00*cos) + float64(t.m01*sin)
	n01 := float64(t.m00*-sin) + float64(t.m01*cos)
	n10 := float64(t.m10*cos) + float64(t.m11*sin)
	n11 := float64(t.m10*-sin) + float64(t.m11*cos)
	t.m00, t.m01, t.m10, t.m11 = n00, n01, n10, n11
	t.Translate(-anchorX, -anchorY)
}

// String is upstream's toString.
func (t *XAffineTransform) String() string {
	return "XAffineTransform[[" + jstr(t.m00) + ", " + jstr(t.m01) + ", " + jstr(t.m02) + "], [" +
		jstr(t.m10) + ", " + jstr(t.m11) + ", " + jstr(t.m12) + "]]"
}
