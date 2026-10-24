// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HSLColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// HSLColor is a color in HSL (hue, saturation, luminance), taken by upstream
// from http://www.camick.com/java/source/HSLColor.java. The hue is an angle
// in degrees, red being 0, green 120 and blue 240; saturation and luminance
// are percentages. Changing the luminance gives the tones and shades of a
// color.
//
// As upstream, all the arithmetic is in float32. Products that are added are
// converted explicitly, which stops the compiler from fusing them into FMA
// instructions that Java never uses.
type HSLColor struct {
	rgb   XColor
	hsl   [3]float32
	alpha float32
}

// NewHSLColor returns the HSL form of an RGB color.
func NewHSLColor(rgb XColor) *HSLColor {
	return &HSLColor{rgb: rgb, hsl: HSLFromRGB(rgb), alpha: float32(rgb.Alpha()) / 255}
}

// NewHSLColorHSL returns an opaque color from its hue in degrees and its
// saturation and luminance in percent.
func NewHSLColorHSL(h, s, l float32) *HSLColor { return NewHSLColorHSLA(h, s, l, 1) }

// NewHSLColorHSLA returns a color from its hue in degrees, its saturation
// and luminance in percent and its alpha between 0 and 1.
func NewHSLColorHSLA(h, s, l, alpha float32) *HSLColor {
	return &HSLColor{hsl: [3]float32{h, s, l}, alpha: alpha, rgb: HSLToRGB(h, s, l, alpha)}
}

// AdjustHue returns the color with the hue set to degrees.
func (c *HSLColor) AdjustHue(degrees float32) XColor {
	return HSLToRGB(degrees, c.hsl[1], c.hsl[2], c.alpha)
}

// AdjustLuminance returns the color with the luminance set to percent.
func (c *HSLColor) AdjustLuminance(percent float32) XColor {
	return HSLToRGB(c.hsl[0], c.hsl[1], percent, c.alpha)
}

// AdjustSaturation returns the color with the saturation set to percent.
func (c *HSLColor) AdjustSaturation(percent float32) XColor {
	return HSLToRGB(c.hsl[0], percent, c.hsl[2], c.alpha)
}

// AdjustShade returns a darker color: the luminance decreased by percent of
// itself.
func (c *HSLColor) AdjustShade(percent float32) XColor {
	multiplier := (100.0 - percent) / 100.0
	l := max(0.0, c.hsl[2]*multiplier)
	return HSLToRGB(c.hsl[0], c.hsl[1], l, c.alpha)
}

// AdjustTone returns a lighter color: the luminance increased by percent of
// itself.
func (c *HSLColor) AdjustTone(percent float32) XColor {
	multiplier := (100.0 + percent) / 100.0
	l := min(100.0, c.hsl[2]*multiplier)
	return HSLToRGB(c.hsl[0], c.hsl[1], l, c.alpha)
}

// Alpha returns the alpha value, between 0 and 1.
func (c *HSLColor) Alpha() float32 { return c.alpha }

// Complementary returns the opaque color with the hue turned by 180
// degrees.
func (c *HSLColor) Complementary() XColor {
	hue := javaFRem(c.hsl[0]+180.0, 360.0)
	return HSLToRGB(hue, c.hsl[1], c.hsl[2], 1)
}

// Hue returns the hue in degrees.
func (c *HSLColor) Hue() float32 { return c.hsl[0] }

// HSL returns the hue, saturation and luminance.
func (c *HSLColor) HSL() [3]float32 { return c.hsl }

// Luminance returns the luminance in percent.
func (c *HSLColor) Luminance() float32 { return c.hsl[2] }

// RGB returns the RGB color.
func (c *HSLColor) RGB() XColor { return c.rgb }

// Saturation returns the saturation in percent.
func (c *HSLColor) Saturation() float32 { return c.hsl[1] }

// String is upstream's toString, with Java's float formatting.
func (c *HSLColor) String() string {
	return "HSLColor[h=" + javaFloatToString(c.hsl[0]) + ",s=" + javaFloatToString(c.hsl[1]) + ",l=" +
		javaFloatToString(c.hsl[2]) + ",alpha=" + javaFloatToString(c.alpha) + "]"
}

// HSLFromRGB converts an RGB color to hue, saturation and luminance.
func HSLFromRGB(color XColor) [3]float32 {
	rgb := color.RGBColorComponents()
	r, g, b := rgb[0], rgb[1], rgb[2]

	mn := min(r, min(g, b))
	mx := max(r, max(g, b))

	var h float32
	switch {
	case mx == mn:
		h = 0
	case mx == r:
		h = javaFRem(float32(60*(g-b))/(mx-mn)+360, 360)
	case mx == g:
		h = float32(60*(b-r))/(mx-mn) + 120
	case mx == b:
		h = float32(60*(r-g))/(mx-mn) + 240
	}

	l := (mx + mn) / 2

	var s float32
	switch {
	case mx == mn:
		s = 0
	case l <= .5:
		s = (mx - mn) / (mx + mn)
	default:
		s = (mx - mn) / (2 - mx - mn)
	}

	return [3]float32{h, s * 100, l * 100}
}

// HSLToRGB converts hue in degrees, saturation and luminance in percent and
// alpha between 0 and 1 to an RGB color. Saturation, luminance and alpha are
// clamped to their range.
func HSLToRGB(h, s, l, alpha float32) XColor {
	if s < 0 {
		s = 0
	}
	if s > 100 {
		s = 100
	}
	if l < 0 {
		l = 0
	}
	if l > 100 {
		l = 100
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	// Formula needs all values between 0 - 1.
	h = javaFRem(h, 360.0)
	h /= 360
	s /= 100
	l /= 100

	var q float32
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = (l + s) - float32(s*l)
	}

	p := float32(2*l) - q

	r := max(0, hueToRGB(p, q, h+1.0/3.0))
	g := max(0, hueToRGB(p, q, h))
	b := max(0, hueToRGB(p, q, h-1.0/3.0))

	r = min(r, 1.0)
	g = min(g, 1.0)
	b = min(b, 1.0)

	return XColorFromFloat(r, g, b, alpha)
}

func hueToRGB(p, q, h float32) float32 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	if 6*h < 1 {
		return p + float32(float32((q-p)*6)*h)
	}
	if 2*h < 1 {
		return q
	}
	if 3*h < 2 {
		return p + float32(float32((q-p)*6)*(2.0/3.0-h))
	}
	return p
}
