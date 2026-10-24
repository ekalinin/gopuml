// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/Positionable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// Positionable is a moveable item with a position and a size.
type Positionable interface {
	Moveable
	Size() XDimension2D
	Position() XPoint2D
}
