// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/ULine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ULine is a line segment from the current position to (dx, dy) away.
type ULine struct {
	AbstractShadowable
	dx float64
	dy float64
}

var _ UShapeSized = (*ULine)(nil)

// ULineCreate returns the line from p1 to p2.
func ULineCreate(p1, p2 geom.XPoint2D) *ULine { return NewULine(p2.X-p1.X, p2.Y-p1.Y) }

// NewULine returns the line of a vector.
func NewULine(dx, dy float64) *ULine { return &ULine{dx: dx, dy: dy} }

// Rotate returns the line rotated by theta radians, the line itself for 0.
func (l *ULine) Rotate(theta float64) *ULine {
	if theta == 0 {
		return l
	}
	tmp := XAffineTransformRotateInstance(theta).Transform(geom.NewXPoint2D(l.dx, l.dy))
	return NewULine(tmp.X, tmp.Y)
}

// ULineH returns a horizontal line.
func ULineH(dx float64) *ULine { return NewULine(dx, 0) }

// ULineV returns a vertical line.
func ULineV(dy float64) *ULine { return NewULine(0, dy) }

// String is upstream's toString.
func (l *ULine) String() string { return "ULine dx=" + jstr(l.dx) + " dy=" + jstr(l.dy) }

// DX returns the horizontal extent.
func (l *ULine) DX() float64 { return l.dx }

// DY returns the vertical extent.
func (l *ULine) DY() float64 { return l.dy }

// Length returns the length.
func (l *ULine) Length() float64 { return math.Sqrt(float64(l.dx*l.dx) + float64(l.dy*l.dy)) }

// Width returns the horizontal extent.
func (l *ULine) Width() float64 { return l.dx }

// Height returns the vertical extent.
func (l *ULine) Height() float64 { return l.dy }
