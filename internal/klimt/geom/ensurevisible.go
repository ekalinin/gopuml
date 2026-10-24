// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/EnsureVisible.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// EnsureVisible collects points that must be visible.
type EnsureVisible interface {
	EnsureVisible(x, y float64)
}
