// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/Moveable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// Moveable is something that can be moved.
type Moveable interface {
	MoveDelta(deltaX, deltaY float64)
}
