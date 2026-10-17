// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/FontConfiguration.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// fontStyleSet is a set of font.FontStyle, upstream's EnumSet<FontStyle>,
// iterated in enum order.
type fontStyleSet uint8

func (s fontStyleSet) with(style font.FontStyle) fontStyleSet    { return s | 1<<style }
func (s fontStyleSet) without(style font.FontStyle) fontStyleSet { return s &^ (1 << style) }
func (s fontStyleSet) contains(style font.FontStyle) bool        { return s&(1<<style) != 0 }

func (s fontStyleSet) each(f func(font.FontStyle)) {
	for style := font.FontStylePlain; style <= font.FontStyleBackcolor; style++ {
		if s.contains(style) {
			f(style)
		}
	}
}

// String is EnumSet.toString, for example "[ITALIC, BOLD]".
func (s fontStyleSet) String() string {
	var names []string
	s.each(func(style font.FontStyle) { names = append(names, style.String()) })
	return "[" + strings.Join(names, ", ") + "]"
}

// FontConfiguration is the font, colors and decorations of a text, as creole
// changes them along the text. It is immutable: the methods return new
// configurations.
type FontConfiguration struct {
	styles                   fontStyleSet
	currentFont              font.UFont
	motherFont               font.UFont
	motherColor              color.HColor
	currentColor             color.HColor
	extendedColor            color.HColor
	fontPosition             font.FontPosition
	svgAttributes            *SvgAttributes
	hyperlinkUnderlineStroke *UStroke
	hyperlinkColor           color.HColor
	tabSize                  int
}

// NewFontConfigurationTab is upstream's create(UFont, HColor, HColor,
// UStroke, int): the configuration of a font, a color, a hyperlink color, an
// optional hyperlink underline (nil for none) and a tab size.
func NewFontConfigurationTab(f font.UFont, c, hyperlinkColor color.HColor, hyperlinkUnderlineStroke *UStroke,
	tabSize int) *FontConfiguration {
	return &FontConfiguration{styles: stylesOf(f), currentFont: f, motherFont: f, motherColor: c, currentColor: c,
		fontPosition: font.FontPositionNormal, svgAttributes: SvgAttributesEmpty(), hyperlinkColor: hyperlinkColor,
		hyperlinkUnderlineStroke: hyperlinkUnderlineStroke, tabSize: tabSize}
}

// NewFontConfiguration is upstream's create(UFont, HColor, HColor, UStroke)
// with a tab size of 8.
func NewFontConfiguration(f font.UFont, c, hyperlinkColor color.HColor, hyperlinkUnderlineStroke *UStroke) *FontConfiguration {
	return NewFontConfigurationTab(f, c, hyperlinkColor, hyperlinkUnderlineStroke, 8)
}

// FontConfigurationBlackBlueTrue returns the configuration of a font in
// black (white in dark mode), with blue underlined hyperlinks.
func FontConfigurationBlackBlueTrue(f font.UFont) *FontConfiguration {
	return NewFontConfigurationTab(f, color.Black.WithDark(color.White), color.Blue, strokePtr(UStrokeSimple()), 8)
}

func stylesOf(f font.UFont) fontStyleSet {
	bold := f.FontFace().IsBold()
	italic := f.FontFace().IsItalic()
	var s fontStyleSet
	if bold && italic {
		return s.with(font.FontStyleItalic).with(font.FontStyleBold)
	}
	if bold {
		return s.with(font.FontStyleBold)
	}
	if italic {
		return s.with(font.FontStyleItalic)
	}
	return s
}

// FontConfigurationFromSkinParam is upstream's create(ISkinParam, FontParam,
// Stereotype): the configuration of a font parameter.
func FontConfigurationFromSkinParam(skinParam ISkinParam, fontParam FontParam, stereo *Stereotype) *FontConfiguration {
	return NewFontConfigurationTab(SkinParamUtilsFont(skinParam, fontParam, stereo),
		SkinParamUtilsFontColor(skinParam, fontParam, stereo), skinParam.HyperlinkColor(),
		skinParam.UseUnderlineForHyperlink(), skinParam.TabSize())
}

// FontConfigurationFromStyle is upstream's create(ISkinParam, Style,
// Colors): the configuration of a style, the text color of colors (which may
// be nil) taking precedence.
func FontConfigurationFromStyle(skinParam ISkinParam, style *Style, colors *color.Colors) *FontConfiguration {
	hyperlinkColor := style.Value(PNameHyperLinkColor).AsColor(skinParam.IHtmlColorSet())
	hyperlinkUnderlineStroke := skinParam.UseUnderlineForHyperlink()
	var c color.HColor
	if colors != nil {
		c = colors.Color(color.TypeText)
	}
	if c == nil {
		c = style.Value(PNameFontColor).AsColor(skinParam.IHtmlColorSet())
	}
	return NewFontConfigurationTab(style.UFont(), c, hyperlinkColor, hyperlinkUnderlineStroke, skinParam.TabSize())
}

func (fc *FontConfiguration) copy() *FontConfiguration {
	c := *fc
	return &c
}

// Font returns the font: the current font with the styles applied, then the
// position (superscript, subscript).
func (fc *FontConfiguration) Font() font.UFont {
	result := fc.currentFont
	fc.styles.each(func(style font.FontStyle) { result = style.MutateFont(result) })
	return fc.fontPosition.Mute(result)
}

// FontFace returns the face of the current font.
func (fc *FontConfiguration) FontFace() font.UFontFace { return fc.currentFont.FontFace() }

// WithFontFace returns the configuration with the face of the current font
// changed; the zero face keeps the configuration.
func (fc *FontConfiguration) WithFontFace(face font.UFontFace) *FontConfiguration {
	if face.IsZero() {
		return fc
	}
	c := fc.copy()
	c.currentFont = fc.currentFont.WithFontFace(face)
	return c
}

// StringDebug is upstream's toStringDebug.
func (fc *FontConfiguration) StringDebug() string {
	return fc.Font().String() + " " + fc.styles.String()
}

// Equals is upstream's equals: same current font, styles, colors, hyperlink
// color and underline, position and tab size.
func (fc *FontConfiguration) Equals(other *FontConfiguration) bool {
	return fc.currentFont == other.currentFont && fc.styles == other.styles &&
		fc.currentColor.Equals(other.currentColor) && sameColor(fc.extendedColor, other.extendedColor) &&
		fc.hyperlinkColor.Equals(other.hyperlinkColor) &&
		sameStroke(fc.hyperlinkUnderlineStroke, other.hyperlinkUnderlineStroke) &&
		fc.fontPosition == other.fontPosition && fc.tabSize == other.tabSize
}

func sameColor(c1, c2 color.HColor) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 != nil && c2 != nil {
		return c1.Equals(c2)
	}
	return false
}

func sameStroke(s1, s2 *UStroke) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 != nil && s2 != nil {
		return *s1 == *s2
	}
	return false
}

// Mute returns the configuration with the text color of colors, if any.
func (fc *FontConfiguration) Mute(colors *color.Colors) *FontConfiguration {
	c := colors.Color(color.TypeText)
	if c == nil {
		return fc
	}
	return fc.ChangeColor(c)
}

// String is upstream's toString.
func (fc *FontConfiguration) String() string {
	return fc.styles.String() + " " + colorString(fc.currentColor)
}

// ForceFont returns the configuration of a stereotype: italic when newFont
// is nil, else that font (and color, when not nil).
func (fc *FontConfiguration) ForceFont(newFont *font.UFont, htmlColorForStereotype color.HColor) *FontConfiguration {
	if newFont == nil {
		return fc.Add(font.FontStyleItalic)
	}
	result := fc.copy()
	result.motherFont = *newFont
	result.currentFont = *newFont
	if htmlColorForStereotype != nil {
		result = result.ChangeColor(htmlColorForStereotype)
	}
	return result
}

// ChangeAttributes returns the configuration with more SVG attributes.
func (fc *FontConfiguration) ChangeAttributes(toBeAdded *SvgAttributes) *FontConfiguration {
	c := fc.copy()
	c.svgAttributes = fc.svgAttributes.AddAll(toBeAdded)
	return c
}

func (fc *FontConfiguration) withHyperlink() *FontConfiguration {
	c := fc.copy()
	c.currentColor = fc.hyperlinkColor
	return c
}

// ChangeColor returns the configuration with another text color.
func (fc *FontConfiguration) ChangeColor(newHtmlColor color.HColor) *FontConfiguration {
	c := fc.copy()
	c.currentColor = newHtmlColor
	return c
}

// ChangeExtendedColor returns the configuration with another decoration
// color (underline, wave, strike, background).
func (fc *FontConfiguration) ChangeExtendedColor(newExtendedColor color.HColor) *FontConfiguration {
	c := fc.copy()
	c.extendedColor = newExtendedColor
	return c
}

// ChangeSize returns the configuration with another size, truncated to an
// integer.
func (fc *FontConfiguration) ChangeSize(size float32) *FontConfiguration {
	c := fc.copy()
	c.currentFont = fc.currentFont.WithSize(size)
	return c
}

// Bigger returns the configuration with the size increased by delta.
func (fc *FontConfiguration) Bigger(delta float64) *FontConfiguration {
	return fc.ChangeSize(float32(float64(fc.currentFont.Size()) + delta))
}

// ChangeFontPosition returns the configuration with another position.
func (fc *FontConfiguration) ChangeFontPosition(fontPosition font.FontPosition) *FontConfiguration {
	c := fc.copy()
	c.fontPosition = fontPosition
	return c
}

// ChangeFamily returns the configuration with another font family, keeping
// the face and size.
func (fc *FontConfiguration) ChangeFamily(family string) *FontConfiguration {
	c := fc.copy()
	c.currentFont = font.Build(family, fc.currentFont.FontFace(), fc.currentFont.Size())
	return c
}

// ResetFont returns the configuration with the original font and color.
func (fc *FontConfiguration) ResetFont() *FontConfiguration {
	c := fc.copy()
	c.currentFont = fc.motherFont
	c.currentColor = fc.motherColor
	c.extendedColor = nil
	c.fontPosition = font.FontPositionNormal
	c.svgAttributes = SvgAttributesEmpty()
	return c
}

// Add returns the configuration with a style; the plain style clears the
// others first.
func (fc *FontConfiguration) Add(style font.FontStyle) *FontConfiguration {
	r := fc.styles
	if style == font.FontStylePlain {
		r = 0
	}
	c := fc.copy()
	c.styles = r.with(style)
	return c
}

// Italic adds the italic style.
func (fc *FontConfiguration) Italic() *FontConfiguration { return fc.Add(font.FontStyleItalic) }

// Bold adds the bold style.
func (fc *FontConfiguration) Bold() *FontConfiguration { return fc.Add(font.FontStyleBold) }

// Unbold removes the bold style.
func (fc *FontConfiguration) Unbold() *FontConfiguration { return fc.Remove(font.FontStyleBold) }

// Unitalic removes the italic style.
func (fc *FontConfiguration) Unitalic() *FontConfiguration { return fc.Remove(font.FontStyleItalic) }

// Underline adds the underline style.
func (fc *FontConfiguration) Underline() *FontConfiguration { return fc.Add(font.FontStyleUnderline) }

// Wave adds the wave style in a color.
func (fc *FontConfiguration) Wave(c color.HColor) *FontConfiguration {
	return fc.Add(font.FontStyleWave).ChangeExtendedColor(c)
}

// Hyperlink returns the configuration of a link: the hyperlink color,
// underlined when there is an underline stroke.
func (fc *FontConfiguration) Hyperlink() *FontConfiguration {
	if fc.hyperlinkUnderlineStroke != nil {
		return fc.Add(font.FontStyleUnderline).withHyperlink()
	}
	return fc.withHyperlink()
}

// Remove returns the configuration without a style.
func (fc *FontConfiguration) Remove(style font.FontStyle) *FontConfiguration {
	c := fc.copy()
	c.styles = fc.styles.without(style)
	return c
}

// AdjustColorForBackground returns the configuration with the text color
// resolved against the background of ug.
func (fc *FontConfiguration) AdjustColorForBackground(ug UGraphic) *FontConfiguration {
	backcolor := ug.Param().Backcolor()
	if backcolor.IsTransparent() {
		backcolor = ug.DefaultBackground()
	}
	textColor := fc.Color().AppropriateColor(backcolor)
	return fc.ChangeColor(textColor)
}

// Color returns the text color.
func (fc *FontConfiguration) Color() color.HColor { return fc.currentColor }

// ExtendedColor returns the decoration color, nil for none.
func (fc *FontConfiguration) ExtendedColor() color.HColor { return fc.extendedColor }

// ContainsStyle reports whether a style is set.
func (fc *FontConfiguration) ContainsStyle(style font.FontStyle) bool {
	return fc.styles.contains(style)
}

// Space returns the vertical shift of the position.
func (fc *FontConfiguration) Space() int { return fc.fontPosition.Space() }

// Attributes returns the SVG attributes.
func (fc *FontConfiguration) Attributes() *SvgAttributes { return fc.svgAttributes }

// Size2D returns the size of the current font.
func (fc *FontConfiguration) Size2D() float64 { return fc.currentFont.Size2D() }

// TabSize returns the tab size.
func (fc *FontConfiguration) TabSize() int { return fc.tabSize }

// UnderlineStroke returns the hyperlink underline, nil for none.
func (fc *FontConfiguration) UnderlineStroke() *UStroke { return fc.hyperlinkUnderlineStroke }
