// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/ColorUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "math"

// GrayScale returns the perceived brightness of c, 0 to 255, with the YIQ
// equation from http://24ways.org/2010/calculating-color-contrast.
func GrayScale(c XColor) int { return grayScaleFromRGB(c.Red(), c.Green(), c.Blue()) }

func grayScaleFromRGB(red, green, blue int) int { return grayScaleInternal(red, green, blue) / 1000 }

func grayScaleInternal(red, green, blue int) int { return red*299 + green*587 + blue*114 }

// GrayScaleFromRGB is GrayScale of a Java RGB int.
func GrayScaleFromRGB(rgb int32) int {
	return grayScaleFromRGB(int(rgb&0x00FF0000)>>16, int(rgb&0x0000FF00)>>8, int(rgb&0x000000FF))
}

// GrayScaleColor returns the opaque gray of the brightness of c.
func GrayScaleColor(c XColor) XColor {
	g := GrayScale(c)
	return NewXColor(g, g, g)
}

// GrayScaleColorReverse returns the opaque gray of the reversed brightness
// of c.
func GrayScaleColorReverse(c XColor) XColor {
	g := 255 - GrayScale(c)
	return NewXColor(g, g, g)
}

// Distance returns the YIQ-weighted distance between two colors, not divided
// by 1000.
func Distance(c1, c2 XColor) int {
	return grayScaleInternal(abs(c1.Red()-c2.Red()), abs(c1.Green()-c2.Green()), abs(c1.Blue()-c2.Blue()))
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// This code is still experimental, as upstream says. Many thanks to Alexei
// Boronine for the idea. Some pointers: https://www.hsluv.org/
// https://www.kuon.ch/post/2020-03-08-hsluv/
// https://www.boronine.com/2012/03/26/Color-Spaces-for-Human-Beings/

// toHsluv converts c to HSLuv. Upstream divides the channels by 256, not 255.
func toHsluv(c XColor) [3]float64 {
	return RgbToHsluv([3]float64{float64(c.Red()) / 256.0, float64(c.Green()) / 256.0, float64(c.Blue()) / 256.0})
}

// fromHsluv converts HSLuv to an opaque color, truncating the channels.
func fromHsluv(h, s, l float64) XColor {
	rgb := HsluvToRgb([3]float64{h, s, l})
	return NewXColor(to255(rgb[0]), to255(rgb[1]), to255(rgb[2]))
}

// ReverseHsluv moves the HSLuv lightness of c by half the range and pulls it
// a quarter of the way toward 50.
func ReverseHsluv(c XColor) XColor {
	hsluv := toHsluv(c)
	l := math.Mod(hsluv[2]+50, 100)
	l += float64(0.25 * (50 - l))
	return fromHsluv(hsluv[0], hsluv[1], l)
}

// GrayToColor raises the HSLuv lightness of c by coef of the remaining range.
func GrayToColor(coef float64, c XColor) XColor {
	hsluv := toHsluv(c)
	l := hsluv[2]
	l = l + float64((100-l)*coef)
	return fromHsluv(hsluv[0], hsluv[1], l)
}

// Reversed reverses the HSLuv lightness of c: it moves by 50 for a
// saturation strictly between 40 and 60, and is mirrored around 50
// otherwise.
func Reversed(c XColor) XColor {
	hsluv := toHsluv(c)
	h, s, l := hsluv[0], hsluv[1], hsluv[2]
	if s > 40 && s < 60 {
		if l > 50 {
			l -= 50
		} else if l < 50 {
			l += 50
		}
	} else {
		l = 100 - l
	}
	return fromHsluv(h, s, l)
}

func to255(value float64) int {
	result := javaD2I(float64(255 * value))
	return min(max(result, 0), 255)
}
