// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/StringBounder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringBounderSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringBounderAwt.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/SvgCharSizeHack.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/RichText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/StyledString.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import (
	"strings"
)

// Dimension is the width and height of a measured text, upstream's
// XDimension2D as returned by StringBounder.calculateDimension. It is kept
// minimal so that this package does not depend on the geometry package.
type Dimension struct {
	Width, Height float64
}

// StringBounder measures text for one output format.
type StringBounder interface {
	// MatchesProperty reports whether the bounder has a property such as
	// "SVG".
	MatchesProperty(propertyName string) bool
	// Descent returns the descent of font for text.
	Descent(font UFont, text string) float64
	// CalculateDimension returns the size of text drawn with font.
	CalculateDimension(font UFont, text string) Dimension
	// FileFormat returns the output format the bounder measures for.
	FileFormat() FileFormat
}

// StringBounderDefaults provides the default methods of upstream's
// StringBounder interface. Embed it in a StringBounder implementation.
type StringBounderDefaults struct{}

// MatchesProperty reports false.
func (StringBounderDefaults) MatchesProperty(propertyName string) bool { return false }

// Descent returns the font size divided by 4.5.
func (StringBounderDefaults) Descent(font UFont, text string) float64 {
	return font.Size2D() / 4.5
}

// SvgCharSizeHack rewrites a text before it is measured for SVG output.
type SvgCharSizeHack interface {
	TransformStringForSizeHack(s string) string
}

type noHack struct{}

func (noHack) TransformStringForSizeHack(s string) string { return s }

// NoHack is the SvgCharSizeHack that leaves texts unchanged.
var NoHack SvgCharSizeHack = noHack{}

// StringBounderSvg measures text for SVG output, as Java2D does, with the
// embedded metrics.
type StringBounderSvg struct {
	charSizeHack SvgCharSizeHack
}

var _ StringBounder = (*StringBounderSvg)(nil)

// NewStringBounderSvg returns an SVG bounder. A nil charSizeHack means
// NoHack.
func NewStringBounderSvg(charSizeHack SvgCharSizeHack) *StringBounderSvg {
	if charSizeHack == nil {
		charSizeHack = NoHack
	}
	return &StringBounderSvg{charSizeHack: charSizeHack}
}

// FileFormat returns FormatSVG.
func (b *StringBounderSvg) FileFormat() FileFormat { return FormatSVG }

// CalculateDimension returns the size of text after the char size hack.
func (b *StringBounderSvg) CalculateDimension(font UFont, text string) Dimension {
	text = b.charSizeHack.TransformStringForSizeHack(text)
	return JavaDimension(font, text)
}

// MatchesProperty reports whether propertyName is "SVG", in any case.
func (b *StringBounderSvg) MatchesProperty(propertyName string) bool {
	return strings.EqualFold("SVG", propertyName)
}

// Descent returns the descent of the font Java uses for text.
func (b *StringBounderSvg) Descent(font UFont, text string) float64 {
	return font.UnderlyingFont(text).Descent()
}

// StringBounderAwt measures text for PNG output, as Java2D does, with the
// embedded metrics.
type StringBounderAwt struct {
	StringBounderDefaults
}

var _ StringBounder = StringBounderAwt{}

// CalculateDimension returns the size of text. A text with the bold markers
// of upstream's rich text is measured part by part, the marked parts in bold.
func (StringBounderAwt) CalculateDimension(font UFont, text string) Dimension {
	if isRichText(text) {
		var width, height float64
		for _, s := range buildStyledStrings(text) {
			rect := JavaDimension(s.style.MutateFont(font), s.text)
			width += rect.Width
			height = max(height, rect.Height)
		}
		return Dimension{Width: width, Height: height}
	}
	return JavaDimension(font, text)
}

// Descent returns the descent of the font Java uses for text.
func (StringBounderAwt) Descent(font UFont, text string) float64 {
	return font.UnderlyingFont(text).Descent()
}

// FileFormat returns FormatPNG.
func (StringBounderAwt) FileFormat() FileFormat { return FormatPNG }

// The rich text markers of upstream's StringUtils.
const (
	boldStart = '\uEEF2'
	boldEnd   = '\uEEF1'
)

// isRichText reports whether text holds a rich text marker.
func isRichText(text string) bool {
	return strings.ContainsRune(text, boldStart) || strings.ContainsRune(text, boldEnd)
}

// styledString is a part of a rich text with its style.
type styledString struct {
	text  string
	style FontStyle
}

// buildStyledStrings splits a rich text into plain parts and bold parts
// between boldStart and boldEnd. A missing boldEnd extends the bold part to
// the end of the text.
//
// Upstream fails with an index error when a boldEnd comes before the first
// boldStart; here the boldEnd is looked for after the boldStart, and an
// earlier one stays in the plain part.
func buildStyledStrings(s string) []styledString {
	var result []styledString
	for s != "" {
		i1 := strings.IndexRune(s, boldStart)
		if i1 == -1 {
			result = append(result, styledString{s, FontStylePlain})
			break
		}
		if i1 > 0 {
			result = append(result, styledString{s[:i1], FontStylePlain})
		}
		s = s[i1+len(string(boldStart)):]
		i2 := strings.IndexRune(s, boldEnd)
		if i2 == -1 {
			result = append(result, styledString{s, FontStyleBold})
			s = ""
		} else {
			result = append(result, styledString{s[:i2], FontStyleBold})
			s = s[i2+len(string(boldEnd)):]
		}
	}
	return result
}
