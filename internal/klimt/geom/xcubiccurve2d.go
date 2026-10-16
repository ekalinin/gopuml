// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/XCubicCurve2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// XCubicCurve2D is a cubic Bezier curve from (X1, Y1) to (X2, Y2) with the
// control points (CtrlX1, CtrlY1) and (CtrlX2, CtrlY2).
//
// Upstream's class is mutable. Here it is a value: upstream's
// setCurve(double...) is a plain assignment of NewXCubicCurve2D, and
// subdivide(left, right) returns both halves.
type XCubicCurve2D struct {
	X1, Y1, CtrlX1, CtrlY1, CtrlX2, CtrlY2, X2, Y2 float64
}

// XCubicCurve2DNone returns the curve with all coordinates 0, which is also
// the zero value.
func XCubicCurve2DNone() XCubicCurve2D {
	return XCubicCurve2D{}
}

// NewXCubicCurve2D returns the curve with the given points.
func NewXCubicCurve2D(x1, y1, ctrlx1, ctrly1, ctrlx2, ctrly2, x2, y2 float64) XCubicCurve2D {
	return XCubicCurve2D{
		X1: x1, Y1: y1,
		CtrlX1: ctrlx1, CtrlY1: ctrly1,
		CtrlX2: ctrlx2, CtrlY2: ctrly2,
		X2: x2, Y2: y2,
	}
}

// P1 returns the start point.
func (c XCubicCurve2D) P1() XPoint2D { return XPoint2D{X: c.X1, Y: c.Y1} }

// P2 returns the end point.
func (c XCubicCurve2D) P2() XPoint2D { return XPoint2D{X: c.X2, Y: c.Y2} }

// setCurveFrom is the curve upstream's setCurve(XCubicCurve2D other) stores.
// Like upstream, it copies other's second control point into the end point
// as well.
func setCurveFrom(other XCubicCurve2D) XCubicCurve2D {
	return NewXCubicCurve2D(other.X1, other.Y1, other.CtrlX1, other.CtrlY1, other.CtrlX2, other.CtrlY2,
		other.CtrlX2, other.CtrlY2)
}

// Length returns the distance between the end points.
func (c XCubicCurve2D) Length() float64 {
	dx := c.X2 - c.X1
	dy := c.Y2 - c.Y1
	return math.Sqrt(float64(dx*dx) + float64(dy*dy))
}

// Subdivide splits the curve at its middle (t = 0.5) and returns both halves.
func (c XCubicCurve2D) Subdivide() (left, right XCubicCurve2D) {
	x1 := c.X1
	y1 := c.Y1
	ctrlx1 := c.CtrlX1
	ctrly1 := c.CtrlY1
	ctrlx2 := c.CtrlX2
	ctrly2 := c.CtrlY2
	x2 := c.X2
	y2 := c.Y2
	centerx := (ctrlx1 + ctrlx2) / 2.0
	centery := (ctrly1 + ctrly2) / 2.0
	ctrlx1 = (x1 + ctrlx1) / 2.0
	ctrly1 = (y1 + ctrly1) / 2.0
	ctrlx2 = (x2 + ctrlx2) / 2.0
	ctrly2 = (y2 + ctrly2) / 2.0
	ctrlx12 := (ctrlx1 + centerx) / 2.0
	ctrly12 := (ctrly1 + centery) / 2.0
	ctrlx21 := (ctrlx2 + centerx) / 2.0
	ctrly21 := (ctrly2 + centery) / 2.0
	centerx = (ctrlx12 + ctrlx21) / 2.0
	centery = (ctrly12 + ctrly21) / 2.0
	left = NewXCubicCurve2D(x1, y1, ctrlx1, ctrly1, ctrlx12, ctrly12, centerx, centery)
	right = NewXCubicCurve2D(centerx, centery, ctrlx21, ctrly21, ctrlx2, ctrly2, x2, y2)
	return left, right
}

// CtrlP1 returns the first control point.
func (c XCubicCurve2D) CtrlP1() XPoint2D { return XPoint2D{X: c.CtrlX1, Y: c.CtrlY1} }

// CtrlP2 returns the second control point.
func (c XCubicCurve2D) CtrlP2() XPoint2D { return XPoint2D{X: c.CtrlX2, Y: c.CtrlY2} }

// FlatnessSq returns the square of the largest distance from a control point
// to the segment between the end points.
func (c XCubicCurve2D) FlatnessSq() float64 {
	return max(XLine2DPtSegDistSq(c.X1, c.Y1, c.X2, c.Y2, c.CtrlX1, c.CtrlY1),
		XLine2DPtSegDistSq(c.X1, c.Y1, c.X2, c.Y2, c.CtrlX2, c.CtrlY2))
}

// Flatness returns the largest distance from a control point to the segment
// between the end points.
func (c XCubicCurve2D) Flatness() float64 {
	return math.Sqrt(c.FlatnessSq())
}
