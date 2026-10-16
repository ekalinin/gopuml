// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColorScheme.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// HColorScheme is a "?light:dark[:transparent]" color: the color to use
// depends on the background it is drawn on.
type HColorScheme struct {
	hcolor
	colorForLight       HColor
	colorForDark        HColor
	colorForTransparent HColor
}

// NewHColorScheme returns a scheme; colorForTransparent may be nil.
func NewHColorScheme(colorForLight, colorForDark, colorForTransparent HColor) *HColorScheme {
	c := &HColorScheme{colorForLight: colorForLight, colorForDark: colorForDark, colorForTransparent: colorForTransparent}
	c.hcolor = hcolor{self: c, name: "HColorScheme"}
	return c
}

// AppropriateColor picks the color for the background back. On a
// transparent background without a color for it, it returns the light color
// with the dark one as its dark mode variant.
func (c *HColorScheme) AppropriateColor(back HColor) HColor {
	if back.IsTransparent() {
		if c.colorForTransparent != nil {
			return c.colorForTransparent
		}
		return c.colorForLight.WithDark(c.colorForDark)
	}
	if back.IsDark() {
		return c.colorForDark
	}
	return c.colorForLight
}
