// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/MinMaxMutable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// MinMaxMutable is a bounding box that grows in place. The zero value is the
// box reduced to the origin, like MinMaxMutableEmpty(true).
type MinMaxMutable struct {
	maxX, maxY, minX, minY float64
}

// MinMaxMutableEmpty returns a box reduced to the origin when initToZero is
// true, and otherwise the inverted box from (Double.MAX_VALUE,
// Double.MAX_VALUE) to (-Double.MAX_VALUE, -Double.MAX_VALUE) that the first
// added point replaces.
func MinMaxMutableEmpty(initToZero bool) *MinMaxMutable {
	if initToZero {
		return newMinMaxMutable(0, 0, 0, 0)
	}
	return newMinMaxMutable(math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)
}

// IsInfinity reports whether no point was added to an inverted box.
func (m *MinMaxMutable) IsInfinity() bool {
	return m.minX == math.MaxFloat64
}

// String returns upstream's format, for example "X=0.0 to 10.0 and Y=0.0 to
// 5.0".
func (m *MinMaxMutable) String() string {
	return "X=" + javaDoubleString(m.minX) + " to " + javaDoubleString(m.maxX) + " and Y=" +
		javaDoubleString(m.minY) + " to " + javaDoubleString(m.maxY)
}

// newMinMaxMutable panics when a coordinate is NaN, as upstream throws
// IllegalArgumentException.
func newMinMaxMutable(minX, minY, maxX, maxY float64) *MinMaxMutable {
	if math.IsNaN(minX) || math.IsNaN(maxX) || math.IsNaN(minY) || math.IsNaN(maxY) {
		panic("geom: NaN in MinMaxMutable")
	}
	return &MinMaxMutable{minX: minX, minY: minY, maxX: maxX, maxY: maxY}
}

// AddPointXY extends the box to contain (x, y). It panics when x or y is NaN.
// It is upstream's addPoint(double, double).
func (m *MinMaxMutable) AddPointXY(x, y float64) {
	if math.IsNaN(x) {
		panic("geom: NaN x")
	}
	if math.IsNaN(y) {
		panic("geom: NaN y")
	}
	m.maxX = max(x, m.maxX)
	m.maxY = max(y, m.maxY)
	m.minX = min(x, m.minX)
	m.minY = min(y, m.minY)
}

// AddPoint extends the box to contain pt.
func (m *MinMaxMutable) AddPoint(pt XPoint2D) {
	m.AddPointXY(pt.X, pt.Y)
}

// MinMaxMutableFromMax returns a box from the origin extended to (maxX,
// maxY). It panics when maxX or maxY is NaN.
func MinMaxMutableFromMax(maxX, maxY float64) *MinMaxMutable {
	if math.IsNaN(maxX) {
		panic("geom: NaN maxX")
	}
	if math.IsNaN(maxY) {
		panic("geom: NaN maxY")
	}
	result := MinMaxMutableEmpty(true)
	result.AddPointXY(maxX, maxY)
	return result
}

// MaxX returns the right side abscissa.
func (m *MinMaxMutable) MaxX() float64 { return m.maxX }

// MaxY returns the bottom side ordinate.
func (m *MinMaxMutable) MaxY() float64 { return m.maxY }

// MinX returns the left side abscissa.
func (m *MinMaxMutable) MinX() float64 { return m.minX }

// MinY returns the top side ordinate.
func (m *MinMaxMutable) MinY() float64 { return m.minY }

// Dimension returns the size of the box. It panics for an inverted box, as
// upstream does.
func (m *MinMaxMutable) Dimension() XDimension2D {
	return NewXDimension2D(m.maxX-m.minX, m.maxY-m.minY)
}

// Reset reduces the box to the origin.
func (m *MinMaxMutable) Reset() {
	m.maxX = 0
	m.maxY = 0
	m.minX = 0
	m.minY = 0
}
