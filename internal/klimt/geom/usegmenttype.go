// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/USegmentType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// USegmentType is the kind of a path segment.
type USegmentType int

// Segment types, in upstream's order.
const (
	SegMoveTo USegmentType = iota
	SegLineTo
	SegQuadTo
	SegCubicTo
	SegClose
	SegArcTo
)

// SegArcToValue is upstream's USegmentType.SEG_ARCTO_VALUE.
const SegArcToValue = 4321

var uSegmentTypeNames = []string{"SEG_MOVETO", "SEG_LINETO", "SEG_QUADTO", "SEG_CUBICTO", "SEG_CLOSE", "SEG_ARCTO"}

// String returns upstream's name, for example "SEG_MOVETO".
func (t USegmentType) String() string { return enumName("USegmentType", uSegmentTypeNames, t) }

// USegmentTypeValueOf returns the segment type named s, as Java's valueOf.
func USegmentTypeValueOf(s string) (USegmentType, bool) {
	return enumValueOf[USegmentType](uSegmentTypeNames, s)
}

// NbPoints returns the number of points of the segment. It panics for
// SegQuadTo and SegArcTo, as upstream throws UnsupportedOperationException.
func (t USegmentType) NbPoints() int {
	switch t {
	case SegMoveTo:
		return 1
	case SegLineTo:
		return 1
	case SegCubicTo:
		return 3
	case SegClose:
		return 0
	}
	panic("geom: NbPoints is not supported for " + t.String())
}
