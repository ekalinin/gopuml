// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/awt/XColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import (
	"fmt"
	"math"
	"strconv"
)

// XColor is an sRGB color with an alpha channel, 8 bits per channel. It is
// upstream's replacement for java.awt.Color and compares with ==.
type XColor struct {
	red, green, blue, alpha uint8
}

// Basic colors.
var (
	XColorBlack = NewXColor(0, 0, 0)
	XColorWhite = NewXColor(255, 255, 255)
	XColorRed   = NewXColor(255, 0, 0)
	XColorBlue  = NewXColor(0, 0, 255)
)

// NewXColor returns an opaque color. Only the low 8 bits of each channel
// are kept.
func NewXColor(r, g, b int) XColor { return NewXColorAlpha(r, g, b, 255) }

// NewXColorAlpha returns a color with an alpha channel, 0 being fully
// transparent. Only the low 8 bits of each channel are kept.
func NewXColorAlpha(r, g, b, a int) XColor {
	return XColor{red: uint8(r & 0xFF), green: uint8(g & 0xFF), blue: uint8(b & 0xFF), alpha: uint8(a & 0xFF)}
}

// XColorFrom returns the opaque color of a 0xRRGGBB value; the high byte is
// ignored.
func XColorFrom(rgb int) XColor {
	return NewXColorAlpha((rgb>>16)&0xFF, (rgb>>8)&0xFF, rgb&0xFF, 255)
}

// XColorFromFloat returns the color of channels between 0 and 1. It panics
// with an IllegalArgumentException RuntimeError when a channel is not finite
// or out of range, as upstream throws.
func XColorFromFloat(r, g, b, a float32) XColor {
	for _, v := range [...]float32{r, g, b, a} {
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) || v < 0 || v > 1 {
			throw(illegalArgument, "")
		}
	}
	return NewXColorAlpha(javaRoundF(r*255)&0xFF, javaRoundF(g*255)&0xFF, javaRoundF(b*255)&0xFF,
		javaRoundF(a*255)&0xFF)
}

// Alpha returns the alpha channel, 0 being fully transparent.
func (c XColor) Alpha() int { return int(c.alpha) }

// Red returns the red channel.
func (c XColor) Red() int { return int(c.red) }

// Green returns the green channel.
func (c XColor) Green() int { return int(c.green) }

// Blue returns the blue channel.
func (c XColor) Blue() int { return int(c.blue) }

// RGB returns the color as a Java int 0xAARRGGBB, negative when alpha is
// 128 or more.
func (c XColor) RGB() int32 {
	return int32(c.alpha)<<24 | int32(c.red)<<16 | int32(c.green)<<8 | int32(c.blue)
}

// RGBColorComponents returns the red, green and blue channels divided by
// 255, in float as upstream.
func (c XColor) RGBColorComponents() [3]float32 {
	return [3]float32{float32(c.red) / 255, float32(c.green) / 255, float32(c.blue) / 255}
}

// HashCode is upstream's hashCode: the RGB value.
func (c XColor) HashCode() int32 { return c.RGB() }

// String is upstream's toString, for example "[r=255,g=0,b=0,a=255]".
func (c XColor) String() string {
	return "[r=" + strconv.Itoa(c.Red()) + ",g=" + strconv.Itoa(c.Green()) + ",b=" + strconv.Itoa(c.Blue()) +
		",a=" + strconv.Itoa(c.Alpha()) + "]"
}

// ToSvg returns the SVG form: "#RRGGBB" in upper case, followed by the alpha
// as "AA" when the color is not opaque, and "#00000000" when it is fully
// transparent. Red is "#FF0000": shortening to "#F00" is done later by the
// SVG writer (SvgGraphics.shortenColor upstream), not here.
func (c XColor) ToSvg() string {
	if c.alpha == 0 {
		return "#00000000"
	}
	if c.alpha != 255 {
		return fmt.Sprintf("#%02X%02X%02X%02X", c.red, c.green, c.blue, c.alpha)
	}
	return fmt.Sprintf("#%02X%02X%02X", c.red, c.green, c.blue)
}

// ToHexRGBColor formats the low 24 bits of a Java RGB int as "#RRGGBB" in
// upper case.
func ToHexRGBColor(rgb int32) string { return fmt.Sprintf("#%06X", rgb&0xFFFFFF) }
