// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/BezierUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// BezierUtilsEndingAngle returns the direction in which the curve arrives at
// its end point, in radians.
func BezierUtilsEndingAngle(left XCubicCurve2D) float64 {
	if left.CtrlP2().Equals(left.P2()) {
		return BezierUtilsAngle(left.P1(), left.P2())
	}
	return BezierUtilsAngle(left.CtrlP2(), left.P2())
}

// BezierUtilsStartingAngle returns the direction in which the curve leaves
// its start point, in radians.
func BezierUtilsStartingAngle(left XCubicCurve2D) float64 {
	if left.P1().Equals(left.CtrlP1()) {
		return BezierUtilsAngle(left.P1(), left.P2())
	}
	return BezierUtilsAngle(left.P1(), left.CtrlP1())
}

// BezierUtilsAngle returns the direction from p1 to p2 in radians, as
// Math.atan2. It panics when the points are equal, as upstream throws
// IllegalArgumentException.
func BezierUtilsAngle(p1, p2 XPoint2D) float64 {
	if p1.Equals(p2) {
		panic("geom: angle between equal points")
	}
	return atan2(p2.Y-p1.Y, p2.X-p1.X)
}

func bezierUtilsIsCutting(bez XCubicCurve2D, shape XRectangle2D) bool {
	contains1 := shape.Contains(bez.X1, bez.Y1)
	contains2 := shape.Contains(bez.X2, bez.Y2)
	return contains1 != contains2
}

// bezierUtilsShorten is upstream's private shorten, which upstream never
// calls. It returns the curve upstream leaves in bez.
func bezierUtilsShorten(bez XCubicCurve2D, shape XRectangle2D) XCubicCurve2D {
	contains1 := shape.Contains(bez.X1, bez.Y1)
	contains2 := shape.Contains(bez.X2, bez.Y2)
	if contains1 == contains2 {
		panic("geom: curve does not cut the shape")
	}
	if !contains1 {
		bez = NewXCubicCurve2D(bez.X2, bez.Y2, bez.CtrlX2, bez.CtrlY2, bez.CtrlX1, bez.CtrlY1, bez.X1, bez.Y1)
	}
	left, right := bezierUtilsSubdivide(bez, 0.5)
	if bezierUtilsIsCutting(left, shape) == bezierUtilsIsCutting(right, shape) {
		panic("geom: curve halves do not cut the shape once")
	}
	if bezierUtilsIsCutting(left, shape) {
		return setCurveFrom(left)
	}
	return setCurveFrom(right)
}

// bezierUtilsSubdivide is upstream's private subdivide, with its formulas:
// unlike XCubicCurve2D.Subdivide, only coef = 0.5 splits at the middle.
func bezierUtilsSubdivide(src XCubicCurve2D, coef float64) (left, right XCubicCurve2D) {
	coef1 := coef
	coef2 := 1 - coef
	centerxA := float64(src.CtrlX1*coef1) + float64(src.CtrlX2*coef2)
	centeryA := float64(src.CtrlY1*coef1) + float64(src.CtrlY2*coef2)

	x1 := src.X1
	y1 := src.Y1
	x2 := src.X2
	y2 := src.Y2
	ctrlx1 := float64(x1*coef1) + float64(src.CtrlX1*coef1)
	ctrly1 := float64(y1*coef1) + float64(src.CtrlY1*coef1)
	ctrlx2 := float64(x2*coef1) + float64(src.CtrlX2*coef1)
	ctrly2 := float64(y2*coef1) + float64(src.CtrlY2*coef1)

	ctrlx12 := float64(ctrlx1*coef1) + float64(centerxA*coef1)
	ctrly12 := float64(ctrly1*coef1) + float64(centeryA*coef1)
	ctrlx21 := float64(ctrlx2*coef1) + float64(centerxA*coef1)
	ctrly21 := float64(ctrly2*coef1) + float64(centeryA*coef1)
	centerxB := float64(ctrlx12*coef1) + float64(ctrlx21*coef1)
	centeryB := float64(ctrly12*coef1) + float64(ctrly21*coef1)
	left = NewXCubicCurve2D(x1, y1, ctrlx1, ctrly1, ctrlx12, ctrly12, centerxB, centeryB)
	right = NewXCubicCurve2D(centerxB, centeryB, ctrlx21, ctrly21, ctrlx2, ctrly2, x2, y2)
	return left, right
}

// bezierUtilsDistCurve is upstream's dist(XCubicCurve2D): the distance
// between the end points.
func bezierUtilsDistCurve(seg XCubicCurve2D) float64 {
	return XPoint2DDistance(seg.X1, seg.Y1, seg.X2, seg.Y2)
}

// bezierUtilsDistLine is upstream's dist(XLine2D): the length of the segment.
func bezierUtilsDistLine(seg XLine2D) float64 {
	return XPoint2DDistance(seg.X1, seg.Y1, seg.X2, seg.Y2)
}

// BezierUtilsMiddle returns the middle of seg. It is upstream's
// middle(XLine2D).
func BezierUtilsMiddle(seg XLine2D) XPoint2D {
	return XPoint2D{X: (seg.X1 + seg.X2) / 2, Y: (seg.Y1 + seg.Y2) / 2}
}

// BezierUtilsMiddlePoints returns the middle of p1 and p2. It is upstream's
// middle(XPoint2D, XPoint2D).
func BezierUtilsMiddlePoints(p1, p2 XPoint2D) XPoint2D {
	return XPoint2D{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
}

// BezierUtilsIntersect returns where orig crosses the border of shape, found
// by bisection down to a segment shorter than 0.1: the end of that segment
// outside shape. It panics when both ends of orig are inside shape or both
// are outside, as upstream throws IllegalArgumentException.
func BezierUtilsIntersect(orig XLine2D, shape XRectangle2D) XPoint2D {
	cp := orig
	contains1 := shape.Contains(cp.X1, cp.Y1)
	contains2 := shape.Contains(cp.X2, cp.Y2)
	if contains1 == contains2 {
		panic("geom: line does not cross the shape border")
	}
	for {
		m := cp.Middle()
		containsMiddle := shape.Contains(m.X, m.Y)
		if contains1 == containsMiddle {
			cp = cp.WithPoint1(m)
		} else {
			cp = cp.WithPoint2(m)
		}
		if bezierUtilsDistLine(cp) < 0.1 {
			if contains1 {
				return cp.P2()
			}
			if contains2 {
				return cp.P1()
			}
			panic("geom: illegal state in BezierUtils.intersect")
		}
	}
}

func bezierUtilsToRectangle(p Positionable) XRectangle2D {
	point := p.Position()
	dim := p.Size()
	return XRectangle2D{X: point.X, Y: point.Y, Width: dim.Width(), Height: dim.Height()}
}

// BezierUtilsIntersectPositionables reports whether the rectangles of p1 and
// p2 intersect. It is upstream's intersect(Positionable, Positionable).
func BezierUtilsIntersectPositionables(p1, p2 Positionable) bool {
	return bezierUtilsToRectangle(p1).Intersects(bezierUtilsToRectangle(p2))
}

// BezierUtilsCenter returns the center of the rectangle of p.
func BezierUtilsCenter(p Positionable) XPoint2D {
	return XPoint2D{X: bezierUtilsToRectangle(p).CenterX(), Y: bezierUtilsToRectangle(p).CenterY()}
}
