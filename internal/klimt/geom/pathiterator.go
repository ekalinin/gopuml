// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PathIterator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// The constants of upstream's PathIterator, with the values of
// java.awt.geom.PathIterator.
const (
	PathIteratorWindEvenOdd = 0
	PathIteratorWindNonZero = 1
	PathIteratorSegMoveTo   = 0
	PathIteratorSegLineTo   = 1
	PathIteratorSegQuadTo   = 2
	PathIteratorSegCubicTo  = 3
	PathIteratorSegClose    = 4
)

// PathIterator walks the segments of a path. Upstream's class only declares
// the methods, which all throw UnsupportedOperationException.
type PathIterator interface {
	IsDone() bool
	Next()
	WindingRule() int
	// CurrentSegment stores the points of the current segment in coord and
	// returns its type, one of the PathIteratorSeg constants.
	CurrentSegment(coord []float64) int
}
