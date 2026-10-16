// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/MinFinder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// MinFinder tracks the smallest abscissa and ordinate seen. Create it with
// NewMinFinder: both minimums start at Double.MAX_VALUE.
type MinFinder struct {
	minX, minY float64
}

// NewMinFinder returns a finder that has seen nothing yet.
func NewMinFinder() *MinFinder {
	return &MinFinder{minX: math.MaxFloat64, minY: math.MaxFloat64}
}

// Manage takes (x, y) into account.
func (f *MinFinder) Manage(x, y float64) {
	if x < f.minX {
		f.minX = x
	}
	if y < f.minY {
		f.minY = y
	}
}

// ManagePoint takes p into account. It is upstream's manage(XPoint2D).
func (f *MinFinder) ManagePoint(p XPoint2D) {
	f.Manage(p.X, p.Y)
}

// ManageMinFinder takes the minimums of other into account. It is upstream's
// manage(MinFinder).
func (f *MinFinder) ManageMinFinder(other *MinFinder) {
	f.Manage(other.minX, other.minY)
}

// String returns upstream's format, for example "minX=1.0 minY=2.0".
func (f *MinFinder) String() string {
	return "minX=" + javaDoubleString(f.minX) + " minY=" + javaDoubleString(f.minY)
}

// MinX returns the smallest abscissa seen.
func (f *MinFinder) MinX() float64 { return f.minX }

// MinY returns the smallest ordinate seen.
func (f *MinFinder) MinY() float64 { return f.minY }
