// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PointAndAngle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// PointAndAngle is an immutable point with an angle.
type PointAndAngle struct {
	pt    XPoint2D
	angle float64
}

// NewPointAndAngle returns pt with angle.
func NewPointAndAngle(pt XPoint2D, angle float64) PointAndAngle {
	return PointAndAngle{pt: pt, angle: angle}
}

// Pt returns the point.
func (p PointAndAngle) Pt() XPoint2D { return p.pt }

// Angle returns the angle.
func (p PointAndAngle) Angle() float64 { return p.angle }

// X returns the abscissa of the point.
func (p PointAndAngle) X() float64 { return p.pt.X }

// Y returns the ordinate of the point.
func (p PointAndAngle) Y() float64 { return p.pt.Y }
