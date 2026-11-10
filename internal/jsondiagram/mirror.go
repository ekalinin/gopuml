// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/Mirror.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import "github.com/ekalinin/gopuml/internal/klimt/geom"

// Mirror turns the coordinates of the top-to-bottom layout into those of the
// drawing, from left to right: the distance from the top of the layout
// becomes the horizontal coordinate. Upstream also writes "BAD VALUE IN
// Mirror" on standard error for a value outside [0, max], which this
// release does not.
type Mirror struct {
	max float64
}

func (m Mirror) inv(v float64) float64 { return m.max - v }

func (m Mirror) invAndXYSwitch(pt geom.XPoint2D) geom.XPoint2D {
	return geom.NewXPoint2D(m.inv(pt.Y), pt.X)
}
