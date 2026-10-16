// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColorSimple.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import (
	"fmt"
	"strconv"
	"strings"
)

// HColorSimple is a plain color, with an optional variant for dark mode.
type HColorSimple struct {
	hcolor
	// color is nil only for the colors of a malformed three-color scheme
	// such as "?nocolor:red:blue", which upstream builds from a null color:
	// using them panics, where upstream throws a NullPointerException.
	color                   *XColor
	dark                    HColor
	transparentFillBehavior TransparentFillBehavior
}

// NewHColorSimple returns the simple color of c, upstream's
// HColorSimple.create.
func NewHColorSimple(c XColor) *HColorSimple { return newHColorSimple(&c, nil, WithFillNone) }

func newHColorSimple(c *XColor, dark HColor, b TransparentFillBehavior) *HColorSimple {
	s := &HColorSimple{color: c, dark: dark, transparentFillBehavior: b}
	s.hcolor = hcolor{self: s, name: "HColorSimple"}
	return s
}

// String is upstream's toString, for example
// "[r=255,g=0,b=0,a=255] α=255".
func (s *HColorSimple) String() string {
	var sb strings.Builder
	if s.dark != nil {
		sb.WriteString("WITHDARK ")
	}
	sb.WriteString(s.color.String())
	sb.WriteString(" α=")
	sb.WriteString(strconv.Itoa(s.color.Alpha()))
	if s.IsTransparent() {
		sb.WriteString(" transparent")
	}
	if s.transparentFillBehavior != WithFillNone {
		sb.WriteString(" " + s.transparentFillBehavior.String())
	}
	return sb.String()
}

// AsString returns "transparent" for a transparent color, "#RRGGBB" in upper
// case for an opaque one, and otherwise "#aarrggbb" in lower case with the
// alpha first.
func (s *HColorSimple) AsString() string {
	if s.IsTransparent() {
		return "transparent"
	}
	c := s.color
	if c.alpha == 255 {
		return fmt.Sprintf("#%02X%02X%02X", c.red, c.green, c.blue)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.alpha, c.red, c.green, c.blue)
}

// Equals reports whether other is a simple color with the same XColor; the
// dark variant and the fill behavior are ignored.
func (s *HColorSimple) Equals(other HColor) bool {
	o, ok := other.(*HColorSimple)
	if !ok || o == nil {
		return false
	}
	c := *s.color
	return o.color != nil && c == *o.color
}

// HashCode is upstream's hashCode: the RGB value of the color.
func (s *HColorSimple) HashCode() int32 { return s.color.HashCode() }

// Lighten increases the HSL luminance by ratio percent of itself. The result
// is opaque.
func (s *HColorSimple) Lighten(ratio int) HColor {
	hsl := HSLFromRGB(*s.color)
	hsl[2] = float32(float64(hsl[2]) + float64(float64(hsl[2])*(float64(ratio)/100.0)))
	return NewHColorSimple(NewHSLColorHSL(hsl[0], hsl[1], hsl[2]).RGB())
}

// Darken decreases the HSL luminance by ratio percent of itself. The result
// is opaque.
func (s *HColorSimple) Darken(ratio int) HColor {
	hsl := HSLFromRGB(*s.color)
	hsl[2] = float32(float64(hsl[2]) - float64(float64(hsl[2])*(float64(ratio)/100.0)))
	return NewHColorSimple(NewHSLColorHSL(hsl[0], hsl[1], hsl[2]).RGB())
}

// ReverseHsluv applies ReverseHsluv to the color.
func (s *HColorSimple) ReverseHsluv() HColor { return NewHColorSimple(ReverseHsluv(*s.color)) }

// Reverse returns the opaque RGB negative.
func (s *HColorSimple) Reverse() HColor { return NewHColorSimple(OrderRGB.Reverse(*s.color)) }

// IsDark reports whether the gray scale of the color is below 128.
func (s *HColorSimple) IsDark() bool { return GrayScale(*s.color) < 128 }

// IsTransparent reports whether alpha is 0.
func (s *HColorSimple) IsTransparent() bool { return s.color.Alpha() == 0 }

func (s *HColorSimple) TransparentFillBehavior() TransparentFillBehavior {
	return s.transparentFillBehavior
}

// AwtColor returns the color.
func (s *HColorSimple) AwtColor() XColor { return *s.color }

// AsMonochrome returns the gray of the color.
func (s *HColorSimple) AsMonochrome() HColor { return NewHColorSimple(GrayScaleColor(*s.color)) }

// AsMonochromeWith maps the gray of the color to a tint of
// colorForMonochrome. maxGray is unused, as upstream.
func (s *HColorSimple) AsMonochromeWith(colorForMonochrome *HColorSimple, minGray, maxGray float64) HColor {
	gray := GrayScaleColor(*s.color).Green()
	coef := (float64(gray) - minGray) / 256.0
	return NewHColorSimple(GrayToColor(coef, *colorForMonochrome.color))
}

// ToColor returns the color itself when it is transparent, and the color
// mapped by mapper otherwise.
func (s *HColorSimple) ToColor(mapper *ColorMapper) XColor {
	if s.IsTransparent() {
		return s.AwtColor()
	}
	return mapper.FromColorSimple(s)
}

// unlinearSimple moves from color1 to color2 in HSL, by the cube of
// completion percent, capped at 100.
func unlinearSimple(color1, color2 *HColorSimple, completionInt int) *HColorSimple {
	hsl1 := HSLFromRGB(*color1.color)
	hsl2 := HSLFromRGB(*color2.color)
	completionInt = min(completionInt, 100)
	completion := float32(float64(completionInt) / 100.0)
	completion = completion * completion * completion
	hsl := [3]float32{linear(completion, hsl1[0], hsl2[0]), linear(completion, hsl1[1], hsl2[1]),
		linear(completion, hsl1[2], hsl2[2])}
	return NewHColorSimple(NewHSLColorHSL(hsl[0], hsl[1], hsl[2]).RGB())
}

func linear(factor, x, y float32) float32 { return x + float32((y-x)*factor) }

// Opposite returns white for a dark color and black otherwise, by gray scale.
func (s *HColorSimple) Opposite() HColor {
	mono := GrayScaleColor(*s.color)
	grayScale := 0
	if 255-mono.Green() > 127 {
		grayScale = 255
	}
	return NewHColorSimple(NewXColor(grayScale, grayScale, grayScale))
}

// DistanceTo is Distance between the two colors.
func (s *HColorSimple) DistanceTo(other *HColorSimple) int { return Distance(*s.color, *other.color) }

// IsGray reports whether the red, green and blue channels are equal.
func (s *HColorSimple) IsGray() bool {
	c := s.color
	return c.red == c.green && c.green == c.blue
}

// WithDark returns the color with dark as its dark mode variant.
func (s *HColorSimple) WithDark(dark HColor) HColor {
	return newHColorSimple(s.color, dark, s.transparentFillBehavior)
}

// WithTransparentFillBehavior returns the color with another fill behavior.
func (s *HColorSimple) WithTransparentFillBehavior(b TransparentFillBehavior) HColor {
	return newHColorSimple(s.color, s.dark, b)
}

// DarkSchemeTheme returns the dark variant, or the color itself without one.
func (s *HColorSimple) DarkSchemeTheme() HColor {
	if s.dark == nil {
		return s
	}
	return s.dark
}
