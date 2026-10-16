// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColorGradient.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// HColorGradient is a two-color gradient such as "red-blue". The policy is
// the separator that gave its direction: '-', '\\', '|' or '/'.
type HColorGradient struct {
	hcolor
	color1, color2 HColor
	policy         rune
}

// newHColorGradient flattens gradient arguments: the first color of a
// gradient color1 and the second color of a gradient color2 are used. It
// panics with a NullPointerException RuntimeError for a nil color.
func newHColorGradient(color1, color2 HColor, policy rune) *HColorGradient {
	if g, ok := color1.(*HColorGradient); ok {
		color1 = g.color1
	}
	if g, ok := color2.(*HColorGradient); ok {
		color2 = g.color2
	}
	if color1 == nil || color2 == nil {
		throw(nullPointer, "")
	}
	c := &HColorGradient{color1: color1, color2: color2, policy: policy}
	c.hcolor = hcolor{self: c, name: "HColorGradient"}
	return c
}

// Color1 returns the start color.
func (c *HColorGradient) Color1() HColor { return c.color1 }

// Color2 returns the end color.
func (c *HColorGradient) Color2() HColor { return c.color2 }

// Color returns the color at coeff, between 0 and 1, with the given alpha.
// Each channel moves from color1 by coeff of the difference, truncated
// toward zero. It panics with an IllegalArgumentException RuntimeError when
// coeff is out of range.
func (c *HColorGradient) Color(mapper *ColorMapper, coeff float64, alpha int) XColor {
	if coeff > 1 || coeff < 0 {
		throw(illegalArgument, "c="+javaDoubleToString(coeff))
	}
	c1 := c.color1.ToColor(mapper)
	c2 := c.color2.ToColor(mapper)

	vRed := javaD2I(coeff * float64(c2.Red()-c1.Red()))
	vGreen := javaD2I(coeff * float64(c2.Green()-c1.Green()))
	vBlue := javaD2I(coeff * float64(c2.Blue()-c1.Blue()))

	return NewXColorAlpha(c1.Red()+vRed, c1.Green()+vGreen, c1.Blue()+vBlue, alpha)
}

// RGB returns Color as a Java RGB int.
func (c *HColorGradient) RGB(mapper *ColorMapper, coeff float64, alpha int) int32 {
	return c.Color(mapper, coeff, alpha).RGB()
}

// Policy returns the separator of the gradient.
func (c *HColorGradient) Policy() rune { return c.policy }

// ToColor returns the start color.
func (c *HColorGradient) ToColor(mapper *ColorMapper) XColor { return c.color1.ToColor(mapper) }

// Opposite returns the opposite of the start color.
func (c *HColorGradient) Opposite() HColor { return c.color1.Opposite() }
