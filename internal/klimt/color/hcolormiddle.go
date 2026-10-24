// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HColorMiddle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// HColorMiddle is the RGB average of two colors, resolved with the mapper.
type HColorMiddle struct {
	hcolor
	color1, color2 HColor
}

func newHColorMiddle(c1, c2 HColor) *HColorMiddle {
	c := &HColorMiddle{color1: c1, color2: c2}
	c.hcolor = hcolor{self: c, name: "HColorMiddle"}
	return c
}

// Color1 returns the first color.
func (c *HColorMiddle) Color1() HColor { return c.color1 }

// Color2 returns the second color.
func (c *HColorMiddle) Color2() HColor { return c.color2 }

// ToColor returns the opaque average of the two mapped colors, rounded down.
func (c *HColorMiddle) ToColor(mapper *ColorMapper) XColor {
	cc1 := c.color1.ToColor(mapper)
	cc2 := c.color2.ToColor(mapper)
	return NewXColor((cc1.Red()+cc2.Red())/2, (cc1.Green()+cc2.Green())/2, (cc1.Blue()+cc2.Blue())/2)
}
