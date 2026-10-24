// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/RectangleArea.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// RectangleArea is an immutable rectangle given by its minimum and maximum
// coordinates.
//
// Upstream's getPosition(), which returns a UTranslate, is not ported yet.
type RectangleArea struct {
	minX, minY, maxX, maxY float64
}

// NewRectangleArea returns the rectangle from (minX, minY) to (maxX, maxY).
func NewRectangleArea(minX, minY, maxX, maxY float64) RectangleArea {
	return RectangleArea{minX: minX, minY: minY, maxX: maxX, maxY: maxY}
}

// RectangleAreaBuild returns the smallest rectangle with corners pt1 and pt2.
func RectangleAreaBuild(pt1, pt2 XPoint2D) RectangleArea {
	return NewRectangleArea(min(pt1.X, pt2.X), min(pt1.Y, pt2.Y), max(pt1.X, pt2.X), max(pt1.Y, pt2.Y))
}

// Move returns the rectangle translated by (deltaX, deltaY).
func (r RectangleArea) Move(deltaX, deltaY float64) RectangleArea {
	return NewRectangleArea(r.minX+deltaX, r.minY+deltaY, r.maxX+deltaX, r.maxY+deltaY)
}

// Width returns maxX - minX.
func (r RectangleArea) Width() float64 { return r.maxX - r.minX }

// Height returns maxY - minY.
func (r RectangleArea) Height() float64 { return r.maxY - r.minY }

// Contains reports whether (x, y) is inside, the minimum sides included and
// the maximum sides excluded.
func (r RectangleArea) Contains(x, y float64) bool {
	return x >= r.minX && x < r.maxX && y >= r.minY && y < r.maxY
}

// Merge returns the smallest rectangle containing r and other.
func (r RectangleArea) Merge(other RectangleArea) RectangleArea {
	return NewRectangleArea(min(r.minX, other.minX), min(r.minY, other.minY),
		max(r.maxX, other.maxX), max(r.maxY, other.maxY))
}

// MergePoint returns the smallest rectangle containing r and point. It is
// upstream's merge(XPoint2D).
func (r RectangleArea) MergePoint(point XPoint2D) RectangleArea {
	x := point.X
	y := point.Y
	return NewRectangleArea(min(r.minX, x), min(r.minY, y), max(r.maxX, x), max(r.maxY, y))
}

// ContainsPoint reports whether p is inside. It is upstream's
// contains(XPoint2D).
func (r RectangleArea) ContainsPoint(p XPoint2D) bool {
	return r.Contains(p.X, p.Y)
}

// String returns upstream's format, for example "minX=0.0 maxX=10.0 minY=0.0
// maxY=5.0".
func (r RectangleArea) String() string {
	return "minX=" + javaDoubleString(r.minX) + " maxX=" + javaDoubleString(r.maxX) + " minY=" +
		javaDoubleString(r.minY) + " maxY=" + javaDoubleString(r.maxY)
}

// MinX returns the minimum abscissa.
func (r RectangleArea) MinX() float64 { return r.minX }

// MinY returns the minimum ordinate.
func (r RectangleArea) MinY() float64 { return r.minY }

// MaxX returns the maximum abscissa.
func (r RectangleArea) MaxX() float64 { return r.maxX }

// MaxY returns the maximum ordinate.
func (r RectangleArea) MaxY() float64 { return r.maxY }

// Intersection returns where bez crosses the border of the rectangle, found by
// halving the curve until its end points are less than 2 apart, with the
// starting angle of that piece. The boolean is false when both end points are
// on the same side of the border (upstream returns null). It panics when no
// half of a crossing curve crosses the border, as upstream throws
// IllegalStateException.
func (r RectangleArea) Intersection(bez XCubicCurve2D) (PointDirected, bool) {
	if r.Contains(bez.X1, bez.Y1) == r.Contains(bez.X2, bez.Y2) {
		return PointDirected{}, false
	}
	dist := bez.P1().Distance(bez.P2())
	if dist < 2 {
		angle := BezierUtilsStartingAngle(bez)
		return NewPointDirected(bez.P1(), angle), true
	}
	left, right := bez.Subdivide()
	if int1, ok := r.Intersection(left); ok {
		return int1, true
	}
	if int2, ok := r.Intersection(right); ok {
		return int2, true
	}
	panic("geom: illegal state in RectangleArea.getIntersection")
}

// PointCenter returns the center.
func (r RectangleArea) PointCenter() XPoint2D {
	return XPoint2D{X: (r.minX + r.maxX) / 2, Y: (r.minY + r.maxY) / 2}
}

// WithMinX returns the rectangle with minX set to d.
func (r RectangleArea) WithMinX(d float64) RectangleArea {
	return NewRectangleArea(d, r.minY, r.maxX, r.maxY)
}

// WithMaxX returns the rectangle with maxX set to d.
func (r RectangleArea) WithMaxX(d float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY, d, r.maxY)
}

// AddMaxX returns the rectangle with d added to maxX.
func (r RectangleArea) AddMaxX(d float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY, r.maxX+d, r.maxY)
}

// AddMaxY returns the rectangle with d added to maxY.
func (r RectangleArea) AddMaxY(d float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY, r.maxX, r.maxY+d)
}

// AddMinX returns the rectangle with d added to minX.
func (r RectangleArea) AddMinX(d float64) RectangleArea {
	return NewRectangleArea(r.minX+d, r.minY, r.maxX, r.maxY)
}

// AddMinY returns the rectangle with d added to minY.
func (r RectangleArea) AddMinY(d float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY+d, r.maxX, r.maxY)
}

// WithMinY returns the rectangle with minY set to d.
func (r RectangleArea) WithMinY(d float64) RectangleArea {
	return NewRectangleArea(r.minX, d, r.maxX, r.maxY)
}

// WithMaxY returns the rectangle with maxY set to d.
func (r RectangleArea) WithMaxY(d float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY, r.maxX, d)
}

// Delta returns the rectangle with m1 added to maxX and m2 added to maxY.
func (r RectangleArea) Delta(m1, m2 float64) RectangleArea {
	return NewRectangleArea(r.minX, r.minY, r.maxX+m1, r.maxY+m2)
}

// Dimension returns the size. It panics for an inverted rectangle, as
// upstream does.
func (r RectangleArea) Dimension() XDimension2D {
	return NewXDimension2D(r.maxX-r.minX, r.maxY-r.minY)
}

// ClosestSide returns the side nearest to pt, preferring north, then south,
// east and west on ties. The boolean is false when no side is nearest, which
// only happens with NaN (upstream returns null).
func (r RectangleArea) ClosestSide(pt XPoint2D) (Side, bool) {
	distNorth := math.Abs(r.minY - pt.Y)
	distSouth := math.Abs(r.maxY - pt.Y)
	distWest := math.Abs(r.minX - pt.X)
	distEast := math.Abs(r.maxX - pt.X)
	if isSmallerThan(distNorth, distWest, distEast, distSouth) {
		return SideNorth, true
	}
	if isSmallerThan(distSouth, distNorth, distWest, distEast) {
		return SideSouth, true
	}
	if isSmallerThan(distEast, distNorth, distWest, distSouth) {
		return SideEast, true
	}
	if isSmallerThan(distWest, distNorth, distEast, distSouth) {
		return SideWest, true
	}
	return 0, false
}

func isSmallerThan(value, a, b, c float64) bool {
	return value <= a && value <= b && value <= c
}
