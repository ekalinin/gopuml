// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/Rainbow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/HtmlColorAndStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The arrow colors of upstream's decoration package. Only the activity
// diagrams use them; the line styles are internal/decoration's LinkStyle.

// HtmlColorAndStyle is one part of a rainbow arrow: its color, the color
// of its head and its line style.
type HtmlColorAndStyle struct {
	arrowHeadColor color.HColor
	arrowColor     color.HColor
	style          decoration.LinkStyle
}

// NewHtmlColorAndStyle returns a part with a style; a nil head color is the
// arrow color. A nil arrow color panics, as upstream's requireNonNull.
func NewHtmlColorAndStyle(arrowColor color.HColor, style decoration.LinkStyle,
	arrowHeadColor color.HColor) *HtmlColorAndStyle {
	if arrowColor == nil {
		uml.ThrowNPE("")
	}
	if arrowHeadColor == nil {
		arrowHeadColor = arrowColor
	}
	return &HtmlColorAndStyle{arrowColor: arrowColor, arrowHeadColor: arrowHeadColor, style: style}
}

// String is upstream's toString: the arrow color and the style.
func (c *HtmlColorAndStyle) String() string { return c.arrowColor.String() + " " + c.style.String() }

// ArrowColor returns the color of the line.
func (c *HtmlColorAndStyle) ArrowColor() color.HColor { return c.arrowColor }

// ArrowHeadColor returns the color of the head.
func (c *HtmlColorAndStyle) ArrowHeadColor() color.HColor { return c.arrowHeadColor }

// Style returns the line style.
func (c *HtmlColorAndStyle) Style() decoration.LinkStyle { return c.style }

// arrowStyleSignature is HtmlColorAndStyle.getDefaultStyleDefinitionArrow.
func arrowStyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameArrow)
}

// HtmlColorAndStyleBuild is upstream's build: definition lists colors and
// styles separated by commas. The color is the last valid color, by default
// the line color of the arrow style; an unknown color gives a
// *color.NoSuchColorError.
func HtmlColorAndStyleBuild(skinParam klimt.ISkinParam, definition string) (*HtmlColorAndStyle, error) {
	style := arrowStyleSignature().MergedStyle(skinParam.CurrentStyleBuilder())
	arrowColor := style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	linkStyle := decoration.LinkStyleNormal()
	set := skinParam.IHtmlColorSet()
	for _, s := range uml.JavaSplit(definition, ",") {
		if tmpStyle := decoration.LinkStyleFromString1(s); !tmpStyle.IsNormal() {
			linkStyle = tmpStyle
			continue
		}
		tmpColor, err := set.GetColor(s)
		if err != nil {
			return nil, err
		}
		if tmpColor != nil {
			arrowColor = tmpColor
		}
	}
	return NewHtmlColorAndStyle(arrowColor, linkStyle, nil), nil
}

// Rainbow is the colors of an arrow: several parts are drawn side by side.
type Rainbow struct {
	colors                    []*HtmlColorAndStyle
	colorArrowSeparationSpace int
}

// String is upstream's toString: the list of the parts.
func (r *Rainbow) String() string {
	parts := make([]string, len(r.colors))
	for i, c := range r.colors {
		parts[i] = c.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// RainbowNone is upstream's none(): no color.
func RainbowNone() *Rainbow { return &Rainbow{} }

// RainbowFromColor returns a single part of color arrowColor, none for a
// nil color.
func RainbowFromColor(arrowColor, arrowHeadColor color.HColor) *Rainbow {
	if arrowColor == nil {
		return RainbowNone()
	}
	return RainbowOf(NewHtmlColorAndStyle(arrowColor, decoration.LinkStyleNormal(), arrowHeadColor))
}

// RainbowFromStyle is upstream's build(Style, HColorSet): the line color of
// a style, with its head color (transparent when the style gives none).
func RainbowFromStyle(style *klimt.Style, set *color.HColorSet) *Rainbow {
	c := style.Value(klimt.PNameLineColor).AsColor(set)
	var colorHead color.HColor
	if head := style.Value(klimt.PNameHeadColor); isValueNull(head) {
		colorHead = c
	} else {
		colorHead = head.AsColor(set)
	}
	if colorHead == nil {
		colorHead = color.Transparent()
	}
	return RainbowFromColor(c, colorHead)
}

func isValueNull(v klimt.Value) bool {
	_, ok := v.(*klimt.ValueNull)
	return ok
}

// WithDefault returns defaultColor when the rainbow has no part.
func (r *Rainbow) WithDefault(defaultColor *Rainbow) *Rainbow {
	if r.Size() == 0 {
		return defaultColor
	}
	return r
}

// RainbowOf is upstream's build(HtmlColorAndStyle): a single part.
func RainbowOf(c *HtmlColorAndStyle) *Rainbow {
	if c == nil {
		uml.ThrowNPE("")
	}
	return &Rainbow{colors: []*HtmlColorAndStyle{c}}
}

// RainbowBuild is upstream's build(ISkinParam, String, int): the parts of
// colorString, separated by semicolons; none for ok false. An unknown color
// gives a *color.NoSuchColorError.
func RainbowBuild(skinParam klimt.ISkinParam, colorString string, ok bool,
	colorArrowSeparationSpace int) (*Rainbow, error) {
	if !ok {
		return RainbowNone(), nil
	}
	result := &Rainbow{colorArrowSeparationSpace: colorArrowSeparationSpace}
	for _, s := range uml.JavaSplit(colorString, ";") {
		c, err := HtmlColorAndStyleBuild(skinParam, s)
		if err != nil {
			return nil, err
		}
		result.colors = append(result.colors, c)
	}
	return result, nil
}

// IsInvisible reports whether a part is hidden.
func (r *Rainbow) IsInvisible() bool {
	for _, s := range r.colors {
		if s.Style().IsInvisible() {
			return true
		}
	}
	return false
}

// Colors returns the parts; the slice must not be modified.
func (r *Rainbow) Colors() []*HtmlColorAndStyle { return r.colors }

// Color returns the color of the first part; it panics without parts, as
// upstream.
func (r *Rainbow) Color() color.HColor {
	if len(r.colors) == 0 {
		uml.ThrowIndexOutOfBounds(0, 0)
	}
	return r.colors[0].ArrowColor()
}

// ArrowHeadColor returns the head color of the first part; it panics
// without parts, as upstream.
func (r *Rainbow) ArrowHeadColor() color.HColor {
	if len(r.colors) == 0 {
		uml.ThrowIndexOutOfBounds(0, 0)
	}
	return r.colors[0].ArrowHeadColor()
}

// ColorArrowSeparationSpace returns the space between the parts.
func (r *Rainbow) ColorArrowSeparationSpace() int { return r.colorArrowSeparationSpace }

// Size returns the number of parts.
func (r *Rainbow) Size() int { return len(r.colors) }
