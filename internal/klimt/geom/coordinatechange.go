// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/CoordinateChange.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// CoordinateChange maps coordinates in the frame whose origin is a first
// point and whose first axis points to a second point back to the plane.
type CoordinateChange struct {
	x1, y1 float64

	vectUX, vectUY float64
	vectVX, vectVY float64
	len            float64
}

// CoordinateChangeCreate returns the change for the frame from p1 to p2.
func CoordinateChangeCreate(p1, p2 XPoint2D) CoordinateChange {
	return NewCoordinateChange(p1.X, p1.Y, p2.X, p2.Y)
}

// NewCoordinateChange returns the change for the frame from (x1, y1) to (x2,
// y2). It panics when both points are equal, as upstream throws
// IllegalArgumentException.
func NewCoordinateChange(x1, y1, x2, y2 float64) CoordinateChange {
	c := CoordinateChange{x1: x1, y1: y1}
	c.len = XPoint2DDistance(x1, y1, x2, y2)
	if c.len == 0 {
		panic("geom: coordinate change between equal points")
	}

	c.vectUX = (x2 - x1) / c.len
	c.vectUY = (y2 - y1) / c.len

	c.vectVX = -c.vectUY
	c.vectVY = c.vectUX
	return c
}

// TrueCoordinate returns the point at a along the first axis and b along the
// second one.
func (c CoordinateChange) TrueCoordinate(a, b float64) XPoint2D {
	x := float64(a*c.vectUX) + float64(b*c.vectVX)
	y := float64(a*c.vectUY) + float64(b*c.vectVY)
	return XPoint2D{X: c.x1 + x, Y: c.y1 + y}
}

// Length returns the distance between both points.
func (c CoordinateChange) Length() float64 { return c.len }
