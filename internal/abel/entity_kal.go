// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/Entity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import "github.com/ekalinin/gopuml/internal/klimt"

// Kal is the qualifier box drawn at one end of a qualified association
// (upstream's svek.Kal, which the graph layout creates and draws). The
// entity only keeps them by side.
type Kal interface {
	Position() klimt.Direction
}

// AddKal adds a qualifier box on the side given by its position.
func (e *Entity) AddKal(kal Kal) {
	position := kal.Position()
	e.kals[position] = append(e.kals[position], kal)
}

// Kals returns the qualifier boxes of a side, in the order they were added.
func (e *Entity) Kals(position klimt.Direction) []Kal { return e.kals[position] }

// ClearKals removes the qualifier boxes. Upstream keeps the boxes of every
// layout of the diagram, so a diagram exported twice gets them twice; the
// layout calls ClearKals first so that each export has its own boxes.
func (e *Entity) ClearKals() { e.kals = [4][]Kal{} }
