// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HColorAutomagic.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// HColorAutomagic is the "automatic" color: black or white, whichever
// contrasts with the background it is drawn on. Each parse returns a new
// value, and values are only equal to themselves.
type HColorAutomagic struct {
	hcolor
}

func newHColorAutomagic() *HColorAutomagic {
	c := &HColorAutomagic{}
	c.hcolor = hcolor{self: c, name: "HColorAutomagic"}
	return c
}

// AppropriateColor returns the opposite of the background.
func (c *HColorAutomagic) AppropriateColor(back HColor) HColor { return back.Opposite() }
