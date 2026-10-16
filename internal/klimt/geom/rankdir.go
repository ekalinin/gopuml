// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/Rankdir.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// Rankdir is the main direction of a graph layout.
type Rankdir int

// Values of Rankdir, in upstream's order.
const (
	RankdirLeftToRight Rankdir = iota
	RankdirTopToBottom
)

var rankdirNames = []string{"LEFT_TO_RIGHT", "TOP_TO_BOTTOM"}

// String returns upstream's name, for example "LEFT_TO_RIGHT".
func (v Rankdir) String() string { return enumName("Rankdir", rankdirNames, v) }

// RankdirValueOf returns the value named exactly s, as Java's valueOf.
func RankdirValueOf(s string) (Rankdir, bool) {
	return enumValueOf[Rankdir](rankdirNames, s)
}
