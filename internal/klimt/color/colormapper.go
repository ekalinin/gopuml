// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/ColorMapper.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// ColorMapper turns simple colors into the colors actually drawn, for dark
// mode, monochrome output or reversed colors. Mappers are compared by
// identity, as upstream does.
type ColorMapper struct {
	fromColorSimple func(simple *HColorSimple) XColor
}

// FromColorSimple maps a simple color.
func (m *ColorMapper) FromColorSimple(simple *HColorSimple) XColor { return m.fromColorSimple(simple) }

// The mappers. Dark mode does not compute colors: it picks the dark variant
// given by WithDark, and keeps the color when there is none.
var (
	IdentityMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return s.AwtColor()
	}}
	TeaVMLightMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return s.AwtColor()
	}}
	TeaVMDarkMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return s.DarkSchemeTheme().(*HColorSimple).AwtColor()
	}}
	DarkModeMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return s.DarkSchemeTheme().(*HColorSimple).AwtColor()
	}}
	// LigthnessInverseMapper (upstream's spelling) reverses the HSLuv
	// lightness; it is "skinparam reversecolor dark".
	LigthnessInverseMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return Reversed(s.AwtColor())
	}}
	MonochromeMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return GrayScaleColor(s.AwtColor())
	}}
	MonochromeReverseMapper = &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return GrayScaleColorReverse(s.AwtColor())
	}}
)

// ReverseMapper returns a mapper drawing the negative of the colors with
// their channels reordered by order.
func ReverseMapper(order ColorOrder) *ColorMapper {
	return &ColorMapper{fromColorSimple: func(s *HColorSimple) XColor {
		return order.Reverse(s.AwtColor())
	}}
}
