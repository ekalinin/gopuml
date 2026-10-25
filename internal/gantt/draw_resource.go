// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/ResourceDraw.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// ResourceDraw is upstream's ResourceDraw: the drawing of a resource below
// the tasks.
type ResourceDraw interface {
	klimt.UDrawable
	Height(stringBounder font.StringBounder) float64
	Y() float64
}
