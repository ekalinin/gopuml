// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/emoji/ColorResolver.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/emoji/GrayLevelRange.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/color"

// GrayLevelRange is the range of gray levels of an SVG sprite, which a
// forced color is spread over. Upstream's package emoji joins this one
// (design D2).
type GrayLevelRange interface {
	MinGrayLevel() int
	MaxGrayLevel() int
}

// ColorResolver resolves the colors of an SVG sprite: the default color is
// the forced color, else the font color, else black; with a forced color,
// the sprite colors become gray levels (gray forced color) or tints of the
// forced color.
type ColorResolver struct {
	fontColor      *color.HColorSimple
	forcedColor    *color.HColorSimple
	grayLevelRange GrayLevelRange
}

// NewColorResolver returns the resolver. A color that is not simple panics,
// as upstream's cast to HColorSimple throws ClassCastException; nil stands
// for upstream's null.
func NewColorResolver(fontColor, forcedColor color.HColor, grayLevelRange GrayLevelRange) *ColorResolver {
	return &ColorResolver{fontColor: asHColorSimple(fontColor), forcedColor: asHColorSimple(forcedColor),
		grayLevelRange: grayLevelRange}
}

// asHColorSimple is upstream's cast (HColorSimple) c: nil stays nil, a color
// of another kind panics with a ClassCastException.
func asHColorSimple(c color.HColor) *color.HColorSimple {
	if c == nil {
		return nil
	}
	simple, ok := c.(*color.HColorSimple)
	if !ok {
		throw("java.lang.ClassCastException", "class "+javaClassName(c)+
			" cannot be cast to class net.sourceforge.plantuml.klimt.color.HColorSimple")
	}
	return simple
}

// javaClassName returns the class name of a color, as upstream's toString
// of the kinds that do not override it starts with it.
func javaClassName(c color.HColor) string {
	s := c.String()
	for i := 0; i < len(s); i++ {
		if s[i] == '@' || s[i] == ' ' {
			return s[:i]
		}
	}
	return s
}

// DefaultColor returns the forced color, else the font color, else black.
func (r *ColorResolver) DefaultColor() color.HColor {
	if r.fontColor == nil && r.forcedColor == nil {
		return color.Black
	}
	if r.forcedColor == nil {
		return r.fontColor
	}
	return r.forcedColor
}

// TrueColor returns the color of an SVG color code: none for "none"
// (ignoring case), else the color of the code (white when it is not one),
// turned to gray or to a tint of the forced color when there is one. A code
// that is a color but not a simple one (a gradient, "automatic") panics, as
// upstream's cast throws ClassCastException.
func (r *ColorResolver) TrueColor(code string) color.HColor {
	if javaEqualsIgnoreCase(code, "none") {
		return color.None()
	}

	result := asHColorSimple(color.Instance().GetColorOrWhite(code))

	if r.forcedColor == nil {
		return result
	}

	if r.forcedColor.IsGray() {
		return result.AsMonochrome()
	}

	return result.AsMonochromeWith(r.forcedColor, float64(r.grayLevelRange.MinGrayLevel()),
		float64(r.grayLevelRange.MaxGrayLevel()))
}
