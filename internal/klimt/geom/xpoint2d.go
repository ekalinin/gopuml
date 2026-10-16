// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/XPoint2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// XPoint2D is an immutable point.
//
// Upstream's transform(XAffineTransform) is not ported yet.
type XPoint2D struct {
	X, Y float64
}

// NewXPoint2D returns the point (x, y).
func NewXPoint2D(x, y float64) XPoint2D {
	return XPoint2D{X: x, Y: y}
}

// String returns upstream's format, for example "(12.0,3.5)".
func (p XPoint2D) String() string {
	return "(" + javaDoubleString(p.X) + "," + javaDoubleString(p.Y) + ")"
}

// Equals compares coordinates with ==, like upstream's equals: NaN is never
// equal and -0.0 equals 0.0.
func (p XPoint2D) Equals(other XPoint2D) bool {
	return p.X == other.X && p.Y == other.Y
}

// HashCode is upstream's hashCode, the sum of Java's Double.hashCode of both
// coordinates.
func (p XPoint2D) HashCode() int32 {
	return javaDoubleHashCode(p.X) + javaDoubleHashCode(p.Y)
}

// Distance returns the distance to other.
func (p XPoint2D) Distance(other XPoint2D) float64 {
	px := other.X - p.X
	py := other.Y - p.Y
	return math.Sqrt(float64(px*px) + float64(py*py))
}

// DistanceSq returns the square of the distance to other.
func (p XPoint2D) DistanceSq(other XPoint2D) float64 {
	px := other.X - p.X
	py := other.Y - p.Y
	return float64(px*px) + float64(py*py)
}

// XPoint2DDistance returns the distance between (x1, y1) and (x2, y2).
func XPoint2DDistance(x1, y1, x2, y2 float64) float64 {
	x1 -= x2
	y1 -= y2
	return math.Sqrt(float64(x1*x1) + float64(y1*y1))
}

// DistanceXY returns the distance to (px, py). It is upstream's
// distance(double, double).
func (p XPoint2D) DistanceXY(px, py float64) float64 {
	px -= p.X
	py -= p.Y
	return math.Sqrt(float64(px*px) + float64(py*py))
}

// Move returns the point translated by (dx, dy).
func (p XPoint2D) Move(dx, dy float64) XPoint2D {
	return XPoint2D{X: p.X + dx, Y: p.Y + dy}
}

// MovePoint returns the point translated by delta. It is upstream's
// move(XPoint2D).
func (p XPoint2D) MovePoint(delta XPoint2D) XPoint2D {
	return XPoint2D{X: p.X + delta.X, Y: p.Y + delta.Y}
}
