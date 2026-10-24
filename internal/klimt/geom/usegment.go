// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/USegment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import (
	"slices"
	"strings"
)

// USegment is an immutable path segment: a type and its coordinates.
//
// Upstream's rotate(double) and affine(XAffineTransform, double, double) are
// not ported yet.
type USegment struct {
	coord    []float64
	pathType USegmentType
}

// NewUSegment returns a segment with a copy of coord.
func NewUSegment(coord []float64, pathType USegmentType) USegment {
	return USegment{coord: slices.Clone(coord), pathType: pathType}
}

// String returns upstream's format, the type followed by Java's
// Arrays.toString of the coordinates, for example "SEG_LINETO [1.0, 2.5]".
func (s USegment) String() string {
	parts := make([]string, len(s.coord))
	for i, c := range s.coord {
		parts[i] = javaDoubleString(c)
	}
	return s.pathType.String() + " [" + strings.Join(parts, ", ") + "]"
}

// Coord returns the coordinates. Like upstream, it does not copy them, so
// callers must not modify the slice.
func (s USegment) Coord() []float64 { return s.coord }

// SegmentType returns the type of the segment.
func (s USegment) SegmentType() USegmentType { return s.pathType }

// Translate returns the segment moved by (dx, dy): only the end point of an
// arc moves, all three points of a cubic curve move, and other segments must
// have exactly one point, or Translate panics as upstream throws
// UnsupportedOperationException.
func (s USegment) Translate(dx, dy float64) USegment {
	coord := s.coord
	if s.pathType == SegArcTo {
		return NewUSegment([]float64{coord[0], coord[1], coord[2], coord[3], coord[4], coord[5] + dx, coord[6] + dy},
			s.pathType)
	}
	if s.pathType == SegCubicTo {
		return NewUSegment([]float64{coord[0] + dx, coord[1] + dy, coord[2] + dx, coord[3] + dy,
			coord[4] + dx, coord[5] + dy}, s.pathType)
	}
	if len(coord) != 2 {
		panic("geom: translate is not supported for this segment")
	}
	return NewUSegment([]float64{coord[0] + dx, coord[1] + dy}, s.pathType)
}
