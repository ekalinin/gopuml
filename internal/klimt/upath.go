// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/UPath.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/USegment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// UPath is a path of segments (move, line, cubic and arc). Its bounds follow
// the segment points.
type UPath struct {
	AbstractShadowable
	comment  string
	codeLine string
	segments []geom.USegment
	minmax   geom.MinMax

	ignoreForCompressionOnX bool
	ignoreForCompressionOnY bool
}

var _ UShapeIgnorableForCompression = (*UPath)(nil)

// NewUPath returns an empty path with a comment and a code line, written in
// the SVG output; both may be empty (null upstream).
func NewUPath(comment, codeLine string) *UPath {
	return &UPath{comment: comment, codeLine: codeLine, minmax: geom.MinMaxEmpty(false)}
}

// UPathNone returns an empty path without comment.
func UPathNone() *UPath { return NewUPath("", "") }

// Add appends a segment.
func (p *UPath) Add(coord []float64, pathType geom.USegmentType) {
	p.addInternal(geom.NewUSegment(coord, pathType))
}

// IsEmpty reports whether the path has no segment.
func (p *UPath) IsEmpty() bool { return len(p.segments) == 0 }

// Size returns the number of segments.
func (p *UPath) Size() int { return len(p.segments) }

func (p *UPath) addInternal(segment geom.USegment) {
	p.segments = append(p.segments, segment)
	coord := segment.Coord()
	if segment.SegmentType() == geom.SegArcTo {
		p.minmax = p.minmax.AddPointXY(coord[5], coord[6])
	} else {
		for i := 0; i < len(coord); i += 2 {
			p.minmax = p.minmax.AddPointXY(coord[i], coord[i+1])
		}
	}
}

func (p *UPath) lastCoord() []float64 {
	if len(p.segments) == 0 {
		return nil
	}
	return p.segments[len(p.segments)-1].Coord()
}

// IsInvisible reports whether the path only moves.
func (p *UPath) IsInvisible() bool {
	for _, segment := range p.segments {
		if segment.SegmentType() != geom.SegMoveTo {
			return false
		}
	}
	return true
}

// Translate returns the path moved by (dx, dy); the path itself when both
// are 0.
func (p *UPath) Translate(dx, dy float64) *UPath {
	if dx == 0 && dy == 0 {
		return p
	}
	result := NewUPath(p.comment, p.codeLine)
	for _, seg := range p.segments {
		result.addInternal(seg.Translate(dx, dy))
	}
	return result
}

// Rotate returns the path rotated by theta radians. Only move and line
// segments can be rotated; others panic as upstream throws.
func (p *UPath) Rotate(theta float64) *UPath {
	result := NewUPath(p.comment, p.codeLine)
	for _, seg := range p.segments {
		result.addInternal(rotateSegment(seg, theta))
	}
	return result
}

// Affine returns the path transformed; arcs get their radii scaled by scale
// and their rotation increased by angle.
func (p *UPath) Affine(transform *XAffineTransform, angle, scale float64) *UPath {
	result := NewUPath(p.comment, p.codeLine)
	for _, seg := range p.segments {
		result.addInternal(affineSegment(seg, transform, angle, scale))
	}
	return result
}

// rotateSegment is upstream's USegment.rotate.
func rotateSegment(seg geom.USegment, theta float64) geom.USegment {
	coord := seg.Coord()
	if len(coord) != 2 {
		throw(unsupportedOperation, "")
	}
	p1 := XAffineTransformRotateInstance(theta).Transform(geom.NewXPoint2D(coord[0], coord[1]))
	return geom.NewUSegment([]float64{p1.X, p1.Y}, seg.SegmentType())
}

// affineSegment is upstream's USegment.affine.
func affineSegment(seg geom.USegment, transform *XAffineTransform, angle, scale float64) geom.USegment {
	coord := seg.Coord()
	if seg.SegmentType() == geom.SegArcTo {
		p1 := transform.Transform(geom.NewXPoint2D(coord[5], coord[6]))
		largeArcFlag := coord[3]
		sweepFlag := coord[4]
		return geom.NewUSegment([]float64{coord[0] * scale, coord[1] * scale, coord[2] + angle, largeArcFlag,
			sweepFlag, p1.X, p1.Y}, seg.SegmentType())
	}
	if len(coord) != 2 {
		throw(unsupportedOperation, "")
	}
	p1 := transform.Transform(geom.NewXPoint2D(coord[0], coord[1]))
	return geom.NewUSegment([]float64{p1.X, p1.Y}, seg.SegmentType())
}

// MoveToPoint moves to pt.
func (p *UPath) MoveToPoint(pt geom.XPoint2D) { p.MoveTo(pt.X, pt.Y) }

// LineToPoint draws a line to pt.
func (p *UPath) LineToPoint(pt geom.XPoint2D) { p.LineTo(pt.X, pt.Y) }

// MoveTo moves to (x, y), unless the last segment already ends there.
func (p *UPath) MoveTo(x, y float64) {
	last := p.lastCoord()
	if last != nil && last[0] == x && last[1] == y {
		return
	}
	p.Add([]float64{x, y}, geom.SegMoveTo)
}

// LineTo draws a line to (x, y).
func (p *UPath) LineTo(x, y float64) { p.Add([]float64{x, y}, geom.SegLineTo) }

// CubicToPoints draws a cubic curve with control points p1 and p2 to pt.
func (p *UPath) CubicToPoints(p1, p2, pt geom.XPoint2D) {
	p.CubicTo(p1.X, p1.Y, p2.X, p2.Y, pt.X, pt.Y)
}

// CubicTo draws a cubic curve to (x2, y2).
func (p *UPath) CubicTo(ctrlx1, ctrly1, ctrlx2, ctrly2, x2, y2 float64) {
	p.Add([]float64{ctrlx1, ctrly1, ctrlx2, ctrly2, x2, y2}, geom.SegCubicTo)
}

// QuadTo draws a quadratic curve to (x2, y2), stored as a cubic curve whose
// control points are both the quadratic control point, as upstream does.
func (p *UPath) QuadTo(ctrlx, ctrly, x2, y2 float64) {
	p.Add([]float64{ctrlx, ctrly, ctrlx, ctrly, x2, y2}, geom.SegCubicTo)
}

// QuadToPoints is QuadTo with points.
func (p *UPath) QuadToPoints(ctrl, pt geom.XPoint2D) { p.QuadTo(ctrl.X, ctrl.Y, pt.X, pt.Y) }

// ArcTo draws an SVG elliptical arc to (x, y).
func (p *UPath) ArcTo(rx, ry, xAxisRotation, largeArcFlag, sweepFlag, x, y float64) {
	p.Add([]float64{rx, ry, xAxisRotation, largeArcFlag, sweepFlag, x, y}, geom.SegArcTo)
}

// ArcToPoint draws a circular arc of the given radius to pt.
func (p *UPath) ArcToPoint(pt geom.XPoint2D, radius, largeArcFlag, sweepFlag float64) {
	p.Add([]float64{radius, radius, 0, largeArcFlag, sweepFlag, pt.X, pt.Y}, geom.SegArcTo)
}

// ClosePath does nothing, as upstream.
func (p *UPath) ClosePath() {}

// MaxX returns the largest x of the points.
func (p *UPath) MaxX() float64 { return p.minmax.MaxX() }

// MaxY returns the largest y of the points.
func (p *UPath) MaxY() float64 { return p.minmax.MaxY() }

// MinX returns the smallest x of the points.
func (p *UPath) MinX() float64 { return p.minmax.MinX() }

// MinY returns the smallest y of the points.
func (p *UPath) MinY() float64 { return p.minmax.MinY() }

// String is upstream's toString: the list of segments.
func (p *UPath) String() string {
	parts := make([]string, len(p.segments))
	for i, s := range p.segments {
		parts[i] = s.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Segments returns the segments; the slice must not be modified.
func (p *UPath) Segments() []geom.USegment { return p.segments }

// Comment returns the comment, "" for none.
func (p *UPath) Comment() string { return p.comment }

// CodeLine returns the code line, "" for none.
func (p *UPath) CodeLine() string { return p.codeLine }

// SetIgnoreForCompressionOnX marks the path as ignored when compressing on x.
func (p *UPath) SetIgnoreForCompressionOnX() { p.ignoreForCompressionOnX = true }

// SetIgnoreForCompressionOnY marks the path as ignored when compressing on y.
func (p *UPath) SetIgnoreForCompressionOnY() { p.ignoreForCompressionOnY = true }

// DrawWhenCompressed draws nothing.
func (p *UPath) DrawWhenCompressed(ug UGraphic, mode CompressionMode) {}

// IsIgnoreForCompressionOn reports whether the path is ignored when
// compressing in mode.
func (p *UPath) IsIgnoreForCompressionOn(mode CompressionMode) bool {
	switch mode {
	case CompressionOnX:
		return p.ignoreForCompressionOnX
	case CompressionOnY:
		return p.ignoreForCompressionOnY
	}
	throw(illegalArgument, "")
	return false
}
