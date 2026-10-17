// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/LinkStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package decoration

import (
	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// linkStyleType is upstream's LinkStyle.Type.
type linkStyleType int

const (
	linkStyleNormal linkStyleType = iota
	linkStyleDashed
	linkStyleDotted
	linkStyleBold
	linkStyleInvisible
)

var linkStyleTypeNames = [...]string{"NORMAL", "DASHED", "DOTTED", "BOLD", "INVISIBLE"}

// LinkStyle is the line style of a link, with an optional thickness. It is
// an immutable value.
type LinkStyle struct {
	typ          linkStyleType
	thickness    float64
	hasThickness bool
}

// String is upstream's toString, for example "DASHED(null)" or
// "NORMAL(2.0)".
func (s LinkStyle) String() string {
	thickness := "null"
	if s.hasThickness {
		thickness = jfmt.DoubleToString(s.thickness)
	}
	return linkStyleTypeNames[s.typ] + "(" + thickness + ")"
}

// IsNormal reports whether the style is the plain line.
func (s LinkStyle) IsNormal() bool { return s.typ == linkStyleNormal }

// IsInvisible reports whether the line is hidden.
func (s LinkStyle) IsInvisible() bool { return s.typ == linkStyleInvisible }

// LinkStyleNormal is upstream's LinkStyle.NORMAL().
func LinkStyleNormal() LinkStyle { return LinkStyle{typ: linkStyleNormal} }

// LinkStyleInvisible is upstream's LinkStyle.INVISIBLE().
func LinkStyleInvisible() LinkStyle { return LinkStyle{typ: linkStyleInvisible} }

// LinkStyleBold is upstream's LinkStyle.BOLD().
func LinkStyleBold() LinkStyle { return LinkStyle{typ: linkStyleBold} }

// LinkStyleDotted is upstream's LinkStyle.DOTTED().
func LinkStyleDotted() LinkStyle { return LinkStyle{typ: linkStyleDotted} }

// LinkStyleDashed is upstream's LinkStyle.DASHED().
func LinkStyleDashed() LinkStyle { return LinkStyle{typ: linkStyleDashed} }

// GoThickness returns the style with a thickness.
func (s LinkStyle) GoThickness(thickness float64) LinkStyle {
	return LinkStyle{typ: s.typ, thickness: thickness, hasThickness: true}
}

// Stroke3 is upstream's getStroke3: the stroke of the style.
func (s LinkStyle) Stroke3() klimt.UStroke {
	switch s.typ {
	case linkStyleDashed:
		return klimt.NewUStroke(7, 7, s.nonZeroThickness())
	case linkStyleDotted:
		return klimt.NewUStroke(1, 3, s.nonZeroThickness())
	case linkStyleBold:
		return klimt.UStrokeWithThickness(2)
	}
	return klimt.UStrokeWithThickness(s.nonZeroThickness())
}

// MuteStroke returns the stroke of a dashed, dotted or bold style, stroke
// otherwise.
func (s LinkStyle) MuteStroke(stroke klimt.UStroke) klimt.UStroke {
	if s.typ == linkStyleDashed || s.typ == linkStyleDotted || s.typ == linkStyleBold {
		return s.Stroke3()
	}
	return stroke
}

func (s LinkStyle) nonZeroThickness() float64 {
	if !s.hasThickness {
		return 1
	}
	return s.thickness
}

// IsThicknessOverrided reports whether a thickness is set.
func (s LinkStyle) IsThicknessOverrided() bool { return s.hasThickness }

// LinkStyleFromString1 is upstream's fromString1: the style named s, the
// normal style for any other text.
func LinkStyleFromString1(s string) LinkStyle {
	if result, ok := LinkStyleFromString2(s); ok {
		return result
	}
	return LinkStyleNormal()
}

// LinkStyleFromString2 is upstream's fromString2: "dashed", "dotted",
// "bold" and "hidden", ignoring case; ok is false for any other text.
func LinkStyleFromString2(s string) (LinkStyle, bool) {
	switch {
	case uml.EqualsIgnoreCase("dashed", s):
		return LinkStyleDashed(), true
	case uml.EqualsIgnoreCase("dotted", s):
		return LinkStyleDotted(), true
	case uml.EqualsIgnoreCase("bold", s):
		return LinkStyleBold(), true
	case uml.EqualsIgnoreCase("hidden", s):
		return LinkStyleInvisible(), true
	}
	return LinkStyle{}, false
}
