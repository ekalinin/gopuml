// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/MinMax.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// MinMax is an immutable bounding box. The zero value is the box reduced to
// the origin, like MinMaxEmpty(true).
//
// Upstream's translate(UTranslate), drawGray(UGraphic) and draw(UGraphic,
// HColor, HColor) are not ported yet.
type MinMax struct {
	maxX, maxY, minX, minY float64
}

// DoesHorizontalCross reports whether the horizontal segment from pt1 to pt2
// crosses the whole box, with both ends strictly outside. It panics when the
// segment is not horizontal or is empty.
func (m MinMax) DoesHorizontalCross(pt1, pt2 XPoint2D) bool {
	if pt1.Y != pt2.Y {
		panic("geom: segment is not horizontal")
	}
	if pt1.X == pt2.X {
		panic("geom: empty segment")
	}
	y := pt1.Y
	if y < m.minY || y > m.maxY {
		return false
	}
	if pt1.X < m.minX && pt2.X > m.maxX {
		return true
	}
	if pt2.X < m.minX && pt1.X > m.maxX {
		return true
	}
	return false
}

// MinMaxEmpty returns the box reduced to the origin when initToZero is true,
// and otherwise the inverted box from (Double.MAX_VALUE, Double.MAX_VALUE) to
// (-Double.MAX_VALUE, -Double.MAX_VALUE) that any added point replaces.
func MinMaxEmpty(initToZero bool) MinMax {
	if initToZero {
		return newMinMax(0, 0, 0, 0)
	}
	return newMinMax(math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)
}

// String returns upstream's format, for example "(0.0,0.0)->(10.0,5.0)".
func (m MinMax) String() string {
	return "(" + javaDoubleString(m.minX) + "," + javaDoubleString(m.minY) + ")->(" +
		javaDoubleString(m.maxX) + "," + javaDoubleString(m.maxY) + ")"
}

// MinMaxFromMutable returns a snapshot of minmax.
func MinMaxFromMutable(minmax *MinMaxMutable) MinMax {
	return newMinMax(minmax.MinX(), minmax.MinY(), minmax.MaxX(), minmax.MaxY())
}

// newMinMax panics when a coordinate is NaN, as upstream throws
// IllegalArgumentException.
func newMinMax(minX, minY, maxX, maxY float64) MinMax {
	if math.IsNaN(minX) || math.IsNaN(maxX) || math.IsNaN(minY) || math.IsNaN(maxY) {
		panic("geom: NaN in MinMax")
	}
	return MinMax{minX: minX, minY: minY, maxX: maxX, maxY: maxY}
}

// AddPoint returns the box extended to contain pt.
func (m MinMax) AddPoint(pt XPoint2D) MinMax {
	return m.AddPointXY(pt.X, pt.Y)
}

// AddPointXY returns the box extended to contain (x, y). It is upstream's
// addPoint(double, double).
func (m MinMax) AddPointXY(x, y float64) MinMax {
	return newMinMax(min(x, m.minX), min(y, m.minY), max(x, m.maxX), max(y, m.maxY))
}

// AddMinMax returns the box extended to contain other.
func (m MinMax) AddMinMax(other MinMax) MinMax {
	return newMinMax(min(other.minX, m.minX), min(other.minY, m.minY), max(other.maxX, m.maxX),
		max(other.maxY, m.maxY))
}

// MinMaxFromMax returns the box from the origin extended to (maxX, maxY).
func MinMaxFromMax(maxX, maxY float64) MinMax {
	return MinMaxEmpty(true).AddPointXY(maxX, maxY)
}

// MinMaxFromDim returns the box from the origin to (width, height).
func MinMaxFromDim(dim XDimension2D) MinMax {
	return MinMaxFromMax(dim.Width(), dim.Height())
}

// MaxX returns the right side abscissa.
func (m MinMax) MaxX() float64 { return m.maxX }

// MaxY returns the bottom side ordinate.
func (m MinMax) MaxY() float64 { return m.maxY }

// MinX returns the left side abscissa.
func (m MinMax) MinX() float64 { return m.minX }

// MinY returns the top side ordinate.
func (m MinMax) MinY() float64 { return m.minY }

// Height returns maxY - minY.
func (m MinMax) Height() float64 { return m.maxY - m.minY }

// Width returns maxX - minX.
func (m MinMax) Width() float64 { return m.maxX - m.minX }

// Dimension returns the size of the box. It panics for an inverted box, as
// upstream does.
func (m MinMax) Dimension() XDimension2D {
	return NewXDimension2D(m.maxX-m.minX, m.maxY-m.minY)
}

// Enlarge returns the box with its right and bottom sides moved by dx and dy.
func (m MinMax) Enlarge(dx, dy float64) MinMax {
	return newMinMax(m.minX, m.minY, m.maxX+dx, m.maxY+dy)
}
