// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/PointDirected.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// PointDirected is an immutable point with a direction.
type PointDirected struct {
	x, y, angle float64
}

// NewPointDirected returns p with angle.
func NewPointDirected(p XPoint2D, angle float64) PointDirected {
	return PointDirected{x: p.X, y: p.Y, angle: angle}
}

// Point2D returns the point.
func (p PointDirected) Point2D() XPoint2D { return XPoint2D{X: p.x, Y: p.y} }

// Angle returns the direction.
func (p PointDirected) Angle() float64 { return p.angle }
