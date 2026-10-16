// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// TransparentFillBehavior tells how the SVG writer fills a transparent
// background.
type TransparentFillBehavior int

// Transparent fill behaviors, in upstream's order.
const (
	WithFillNone TransparentFillBehavior = iota
	WithFillOpacity
)

// String returns upstream's constant name, for example "WITH_FILL_NONE".
func (b TransparentFillBehavior) String() string {
	if b == WithFillOpacity {
		return "WITH_FILL_OPACITY"
	}
	return "WITH_FILL_NONE"
}

// HColor is a color as diagrams use it: a simple color, or a color resolved
// later against a mapper (dark mode, monochrome) or a background (automatic
// colors, color schemes), or a gradient.
//
// The methods mirror upstream's abstract class. Kinds that do not override
// a method get upstream's default: ToColor, WithDark,
// WithTransparentFillBehavior and Opposite panic with an
// UnsupportedOperationException RuntimeError, the color transformations
// return the color itself, IsDark is true and IsTransparent is false.
type HColor interface {
	// ToColor resolves the color with mapper.
	ToColor(mapper *ColorMapper) XColor
	// ToRGB returns the resolved color as "#RRGGBB", ignoring alpha.
	ToRGB(mapper *ColorMapper) string
	// ToSvg returns the resolved color as SVG writes it, "#00000000" for a
	// transparent color.
	ToSvg(mapper *ColorMapper) string
	// Lighten increases the HSL lightness by ratio percent of itself.
	Lighten(ratio int) HColor
	// Darken decreases the HSL lightness by ratio percent of itself.
	Darken(ratio int) HColor
	// ReverseHsluv reverses the HSLuv lightness.
	ReverseHsluv() HColor
	// Reverse returns the RGB negative.
	Reverse() HColor
	IsDark() bool
	TransparentFillBehavior() TransparentFillBehavior
	// AsString is the form the preprocessor color functions return.
	AsString() string
	// DarkSchemeTheme returns the color to use in dark mode.
	DarkSchemeTheme() HColor
	// AppropriateColor resolves the color against the background back.
	AppropriateColor(back HColor) HColor
	WithDark(dark HColor) HColor
	WithTransparentFillBehavior(b TransparentFillBehavior) HColor
	// Opposite returns black or white, whichever contrasts with the color.
	Opposite() HColor
	IsTransparent() bool
	// Equals is upstream's equals: simple colors are equal when their
	// XColor values are; other kinds are only equal to themselves.
	Equals(other HColor) bool
	// String is upstream's toString.
	String() string
}

// hcolor holds upstream's default HColor behavior. self is the concrete
// color, so that the defaults that return "this" and the final methods
// ToRGB and ToSvg see the overriding methods, as Java's dispatch does.
type hcolor struct {
	self HColor
	name string // upstream's simple class name
}

func (h *hcolor) className() string { return "net.sourceforge.plantuml.klimt.color." + h.name }

func (h *hcolor) ToColor(*ColorMapper) XColor {
	throw(unsupportedOperation, "class "+h.className())
	return XColor{}
}

func (h *hcolor) ToRGB(mapper *ColorMapper) string {
	return ToHexRGBColor(h.self.ToColor(mapper).RGB())
}

func (h *hcolor) ToSvg(mapper *ColorMapper) string {
	if h.self.IsTransparent() {
		return "#00000000"
	}
	return h.self.ToColor(mapper).ToSvg()
}

func (h *hcolor) Lighten(int) HColor                  { return h.self }
func (h *hcolor) Darken(int) HColor                   { return h.self }
func (h *hcolor) ReverseHsluv() HColor                { return h.self }
func (h *hcolor) Reverse() HColor                     { return h.self }
func (h *hcolor) IsDark() bool                        { return true }
func (h *hcolor) AsString() string                    { return "?" + h.name }
func (h *hcolor) DarkSchemeTheme() HColor             { return h.self }
func (h *hcolor) AppropriateColor(back HColor) HColor { return h.self }
func (h *hcolor) IsTransparent() bool                 { return false }
func (h *hcolor) Equals(other HColor) bool            { return h.self == other }

func (h *hcolor) TransparentFillBehavior() TransparentFillBehavior { return WithFillNone }

func (h *hcolor) WithDark(HColor) HColor {
	throw(unsupportedOperation, "")
	return nil
}

func (h *hcolor) WithTransparentFillBehavior(TransparentFillBehavior) HColor {
	throw(unsupportedOperation, "")
	return nil
}

func (h *hcolor) Opposite() HColor {
	throw(unsupportedOperation, "")
	return nil
}

// String stands for Java's Object.toString, without the identity hash.
func (h *hcolor) String() string { return h.className() }
