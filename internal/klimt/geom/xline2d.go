// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/XLine2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// XLine2D is an immutable line segment from (X1, Y1) to (X2, Y2).
//
// Upstream's drawU(UGraphic) is not ported yet.
type XLine2D struct {
	X1, Y1, X2, Y2 float64
}

// NewXLine2D returns the segment from (x1, y1) to (x2, y2).
func NewXLine2D(x1, y1, x2, y2 float64) XLine2D {
	return XLine2D{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// XLine2DLine returns the segment from p1 to p2.
func XLine2DLine(p1, p2 XPoint2D) XLine2D {
	return XLine2D{X1: p1.X, Y1: p1.Y, X2: p2.X, Y2: p2.Y}
}

// Middle returns the middle of the segment.
func (l XLine2D) Middle() XPoint2D {
	mx := (l.X1 + l.X2) / 2
	my := (l.Y1 + l.Y2) / 2
	return XPoint2D{X: mx, Y: my}
}

// P1 returns the start point.
func (l XLine2D) P1() XPoint2D { return XPoint2D{X: l.X1, Y: l.Y1} }

// P2 returns the end point.
func (l XLine2D) P2() XPoint2D { return XPoint2D{X: l.X2, Y: l.Y2} }

// WithPoint1 returns the segment with another start point.
func (l XLine2D) WithPoint1(other XPoint2D) XLine2D {
	return XLine2D{X1: other.X, Y1: other.Y, X2: l.X2, Y2: l.Y2}
}

// WithPoint2 returns the segment with another end point.
func (l XLine2D) WithPoint2(other XPoint2D) XLine2D {
	return XLine2D{X1: l.X1, Y1: l.Y1, X2: other.X, Y2: other.Y}
}

// XLine2DPtSegDistSq returns the square of the distance from (px, py) to the
// segment from (x1, y1) to (x2, y2), with the algorithm of
// java.awt.geom.Line2D.ptSegDistSq. It is 0 when the point is on the segment.
func XLine2DPtSegDistSq(x1, y1, x2, y2, px, py float64) float64 {
	// Adjust vectors relative to x1,y1
	// x2,y2 becomes relative vector from x1,y1 to end of segment
	x2 -= x1
	y2 -= y1
	// px,py becomes relative vector from x1,y1 to test point
	px -= x1
	py -= y1
	dotprod := float64(px*x2) + float64(py*y2)
	var projlenSq float64
	if dotprod <= 0.0 {
		// px,py is on the side of x1,y1 away from x2,y2
		// distance to segment is length of px,py vector
		// "length of its (clipped) projection" is now 0.0
		projlenSq = 0.0
	} else {
		// switch to backwards vectors relative to x2,y2
		// x2,y2 are already the negative of x1,y1=>x2,y2
		// to get px,py to be the negative of px,py=>x2,y2
		// the dot product of two negated vectors is the same
		// as the dot product of the two normal vectors
		px = x2 - px
		py = y2 - py
		dotprod = float64(px*x2) + float64(py*y2)
		if dotprod <= 0.0 {
			// px,py is on the side of x2,y2 away from x1,y1
			// distance to segment is length of (backwards) px,py vector
			// "length of its (clipped) projection" is now 0.0
			projlenSq = 0.0
		} else {
			// px,py is between x1,y1 and x2,y2
			// dotprod is the length of the px,py vector
			// projected on the x2,y2=>x1,y1 vector times the
			// length of the x2,y2=>x1,y1 vector
			projlenSq = float64(dotprod*dotprod) / (float64(x2*x2) + float64(y2*y2))
		}
	}
	// Distance to line is now the length of the relative point
	// vector minus the length of its projection onto the line
	// (which is zero if the projection falls outside the range
	// of the line segment).
	lenSq := float64(px*px) + float64(py*py) - projlenSq
	if lenSq < 0 {
		lenSq = 0
	}
	return lenSq
}

// Intersect returns the intersection of two segments. The boolean is false
// when they do not intersect (upstream returns null), including when they are
// parallel.
func (l XLine2D) Intersect(line2 XLine2D) (XPoint2D, bool) {
	s1x := l.X2 - l.X1
	s1y := l.Y2 - l.Y1

	s2x := line2.X2 - line2.X1
	s2y := line2.Y2 - line2.Y1

	s := (float64(-s1y*(l.X1-line2.X1)) + float64(s1x*(l.Y1-line2.Y1))) / (float64(-s2x*s1y) + float64(s1x*s2y))
	t := (float64(s2x*(l.Y1-line2.Y1)) - float64(s2y*(l.X1-line2.X1))) / (float64(-s2x*s1y) + float64(s1x*s2y))

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		return XPoint2D{X: l.X1 + float64(t*s1x), Y: l.Y1 + float64(t*s1y)}, true
	}
	return XPoint2D{}, false
}

// Angle returns the direction of the segment in radians, as Math.atan2.
func (l XLine2D) Angle() float64 {
	return atan2(l.Y2-l.Y1, l.X2-l.X1)
}
