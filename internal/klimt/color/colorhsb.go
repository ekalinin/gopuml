// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/ColorHSB.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "fmt"

// ColorHSB is a color with its hue, saturation and brightness, as computed by
// java.awt.Color.RGBtoHSB.
type ColorHSB struct {
	value      int32
	hue        float32
	saturation float32
	brightness float32
}

// NewColorHSB returns the color of a Java ARGB int, alpha included.
func NewColorHSB(rgba int32) *ColorHSB {
	c := &ColorHSB{value: rgba}
	c.hue, c.saturation, c.brightness = rgbToHSB(c.Red(), c.Green(), c.Blue())
	return c
}

// NewColorHSBFrom returns the HSB form of other.
func NewColorHSBFrom(other XColor) *ColorHSB { return NewColorHSB(other.RGB()) }

// rgbToHSB is java.awt.Color.RGBtoHSB.
func rgbToHSB(r, g, b int) (hue, saturation, brightness float32) {
	cmax := max(r, g, b)
	cmin := min(r, g, b)

	brightness = float32(cmax) / 255.0
	if cmax != 0 {
		saturation = float32(cmax-cmin) / float32(cmax)
	}
	if saturation == 0 {
		return 0, saturation, brightness
	}
	redc := float32(cmax-r) / float32(cmax-cmin)
	greenc := float32(cmax-g) / float32(cmax-cmin)
	bluec := float32(cmax-b) / float32(cmax-cmin)
	switch {
	case r == cmax:
		hue = bluec - greenc
	case g == cmax:
		hue = 2.0 + redc - bluec
	default:
		hue = 4.0 + greenc - redc
	}
	hue = hue / 6.0
	if hue < 0 {
		hue = hue + 1.0
	}
	return hue, saturation, brightness
}

// Alpha returns the alpha channel.
func (c *ColorHSB) Alpha() int { return int(c.value>>24) & 0xFF }

// Red returns the red channel.
func (c *ColorHSB) Red() int { return int(c.value>>16) & 0xFF }

// Green returns the green channel.
func (c *ColorHSB) Green() int { return int(c.value>>8) & 0xFF }

// Blue returns the blue channel.
func (c *ColorHSB) Blue() int { return int(c.value) & 0xFF }

// RGB returns the Java ARGB int.
func (c *ColorHSB) RGB() int32 { return c.value }

// Hue returns the hue, between 0 and 1.
func (c *ColorHSB) Hue() float32 { return c.hue }

// Saturation returns the saturation, between 0 and 1.
func (c *ColorHSB) Saturation() float32 { return c.saturation }

// Brightness returns the brightness, between 0 and 1.
func (c *ColorHSB) Brightness() float32 { return c.brightness }

// String is upstream's toString, for example
// "ColorHSB[a=FF r=FF g=00 b=00 / h=0.000000 s=1.000000 b=1.000000]".
func (c *ColorHSB) String() string {
	return fmt.Sprintf("ColorHSB[a=%02X r=%02X g=%02X b=%02X / h=%s s=%s b=%s]", c.Alpha(), c.Red(), c.Green(),
		c.Blue(), javaFormatFixed(float64(c.hue), 6), javaFormatFixed(float64(c.saturation), 6),
		javaFormatFixed(float64(c.brightness), 6))
}
