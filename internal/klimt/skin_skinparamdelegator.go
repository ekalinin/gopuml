// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamDelegator.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamBackcolored.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamBackcoloredReference.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamColors.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamForceColor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamForecolored.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// SkinParamDelegator forwards every method to other skin parameters; the
// types below embed it and replace some methods.
//
// Upstream's shadowingForNote calls itself and overflows the stack; per
// design D16 (upstream crashes are fixed) it forwards like the other
// methods.
type SkinParamDelegator struct {
	ISkinParam
}

// NewSkinParamDelegator returns the delegator to skinParam.
func NewSkinParamDelegator(skinParam ISkinParam) SkinParamDelegator {
	return SkinParamDelegator{ISkinParam: skinParam}
}

// String is upstream's toString, without Java's identity hash.
func (d SkinParamDelegator) String() string { return "SkinParamDelegator" }

// Font forwards with inPackageTitle false, as upstream does.
func (d SkinParamDelegator) Font(stereotype *Stereotype, inPackageTitle bool, fontParam ...FontParam) font.UFont {
	return d.ISkinParam.Font(stereotype, false, fontParam...)
}

// SkinParamBackcolored forces the background colors of an element (and of
// the diagram), and optionally its clickable colors.
type SkinParamBackcolored struct {
	SkinParamDelegator
	backColorElement color.HColor
	backColorGeneral color.HColor
	forceClickage    bool
	forced           map[ColorParam]color.HColor
}

// NewSkinParamBackcolored returns the parameters with background colors
// forced; either color may be nil.
func NewSkinParamBackcolored(skinParam ISkinParam, backColorElement, backColorGeneral color.HColor,
	forceClickage bool) *SkinParamBackcolored {
	return &SkinParamBackcolored{SkinParamDelegator: NewSkinParamDelegator(skinParam),
		backColorElement: backColorElement, backColorGeneral: backColorGeneral, forceClickage: forceClickage,
		forced: map[ColorParam]color.HColor{}}
}

// String is upstream's toString.
func (s *SkinParamBackcolored) String() string {
	return s.SkinParamDelegator.String() + " " + colorString(s.backColorElement) + " " +
		colorString(s.backColorGeneral)
}

// BackgroundColor returns the forced general background, if any.
func (s *SkinParamBackcolored) BackgroundColor() color.HColor {
	if s.backColorGeneral != nil {
		return s.backColorGeneral
	}
	return s.ISkinParam.BackgroundColor()
}

// HtmlColor returns the forced element background for background
// parameters, the clickable color when forced, the colors forced by
// ForceColor, and else the delegate's.
func (s *SkinParamBackcolored) HtmlColor(param ColorParam, stereotype *Stereotype, clickable bool) color.HColor {
	if param.IsBackground() && s.backColorElement != nil {
		return s.backColorElement
	}
	if s.forceClickage {
		if c1 := s.ISkinParam.HtmlColor(param, stereotype, true); c1 != nil {
			return c1
		}
	}
	if forcedColor := s.forced[param]; forcedColor != nil {
		return forcedColor
	}
	return s.ISkinParam.HtmlColor(param, stereotype, clickable)
}

// ForceColor forces the color of a parameter.
func (s *SkinParamBackcolored) ForceColor(param ColorParam, c color.HColor) { s.forced[param] = c }

// SkinParamBackcoloredReference forces the background colors of a sequence
// reference.
type SkinParamBackcoloredReference struct {
	SkinParamDelegator
	sequenceReferenceHeaderBackground color.HColor
	sequenceReferenceBackground       color.HColor
}

// NewSkinParamBackcoloredReference returns the parameters with the
// reference backgrounds forced; either may be nil.
func NewSkinParamBackcoloredReference(skinParam ISkinParam, sequenceReferenceHeaderBackground,
	sequenceReferenceBackground color.HColor) *SkinParamBackcoloredReference {
	return &SkinParamBackcoloredReference{SkinParamDelegator: NewSkinParamDelegator(skinParam),
		sequenceReferenceHeaderBackground: sequenceReferenceHeaderBackground,
		sequenceReferenceBackground:       sequenceReferenceBackground}
}

// HtmlColor returns the forced reference backgrounds, else the delegate's.
func (s *SkinParamBackcoloredReference) HtmlColor(param ColorParam, stereotype *Stereotype,
	clickable bool) color.HColor {
	if param == ColorParamSequenceReferenceHeaderBackground && s.sequenceReferenceHeaderBackground != nil {
		return s.sequenceReferenceHeaderBackground
	}
	if param == ColorParamSequenceReferenceBackground && s.sequenceReferenceBackground != nil {
		return s.sequenceReferenceBackground
	}
	return s.ISkinParam.HtmlColor(param, stereotype, clickable)
}

// SkinParamColors applies the colors of a command (#back;line:...) over the
// parameters.
type SkinParamColors struct {
	SkinParamDelegator
	colors *color.Colors
}

// NewSkinParamColors returns the parameters with colors applied.
func NewSkinParamColors(skinParam ISkinParam, colors *color.Colors) *SkinParamColors {
	return &SkinParamColors{SkinParamDelegator: NewSkinParamDelegator(skinParam), colors: colors}
}

// ColorsOf returns the colors.
func (s *SkinParamColors) ColorsOf() *color.Colors { return s.colors }

// String is upstream's toString.
func (s *SkinParamColors) String() string {
	if s.colors == nil {
		return "SkinParamColors::null"
	}
	return "SkinParamColors::" + s.colors.String()
}

// Shadowing returns the shadowing of the colors, if given.
func (s *SkinParamColors) Shadowing(stereotype *Stereotype) bool {
	if shadowing, ok := s.colors.Shadowing(); ok {
		return shadowing
	}
	return s.ISkinParam.Shadowing(stereotype)
}

// FontHtmlColor returns the text color of the colors, if given.
func (s *SkinParamColors) FontHtmlColor(stereotype *Stereotype, param ...FontParam) color.HColor {
	if value := s.colors.Color(color.TypeText); value != nil {
		return value
	}
	return s.ISkinParam.FontHtmlColor(stereotype, param...)
}

// HtmlColor returns the color of the colors for the type of the parameter,
// if given.
func (s *SkinParamColors) HtmlColor(param ColorParam, stereotype *Stereotype, clickable bool) color.HColor {
	t, ok := param.ColorType()
	if !ok {
		return s.ISkinParam.HtmlColor(param, stereotype, clickable)
	}
	if value := s.colors.Color(t); value != nil {
		return value
	}
	return s.ISkinParam.HtmlColor(param, stereotype, clickable)
}

// SkinParamForceColor forces the color of one parameter.
type SkinParamForceColor struct {
	SkinParamDelegator
	colorParam ColorParam
	color      color.HColor
}

// NewSkinParamForceColor returns the parameters with one color forced.
func NewSkinParamForceColor(skinParam ISkinParam, colorParam ColorParam, c color.HColor) *SkinParamForceColor {
	return &SkinParamForceColor{SkinParamDelegator: NewSkinParamDelegator(skinParam), colorParam: colorParam,
		color: c}
}

// HtmlColor returns the forced color for its parameter.
func (s *SkinParamForceColor) HtmlColor(param ColorParam, stereotype *Stereotype, clickable bool) color.HColor {
	if s.colorParam == param {
		return s.color
	}
	return s.ISkinParam.HtmlColor(param, stereotype, clickable)
}

// SkinParamForecolored gives one color to every parameter.
type SkinParamForecolored struct {
	SkinParamDelegator
	forecolor color.HColor
}

// NewSkinParamForecolored returns the parameters with every color forced.
func NewSkinParamForecolored(skinParam ISkinParam, forecolor color.HColor) *SkinParamForecolored {
	return &SkinParamForecolored{SkinParamDelegator: NewSkinParamDelegator(skinParam), forecolor: forecolor}
}

// HtmlColor returns the forced color.
func (s *SkinParamForecolored) HtmlColor(ColorParam, *Stereotype, bool) color.HColor {
	return s.forecolor
}
