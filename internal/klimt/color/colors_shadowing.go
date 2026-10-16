// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/Colors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// WithShadowing returns the colors with a shadowing. Upstream's
// applyStereotypeForNote sets the private shadowing field of the colors it
// returns; the notes of graph diagrams need it.
func (c *Colors) WithShadowing(shadowing bool) *Colors {
	result := c.copy()
	result.shadowing = &shadowing
	return result
}
