// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/UFontFace.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/UFontStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/UFontWeight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import (
	"strconv"
	"strings"
)

// UFontStyle is the italic axis of a font face.
type UFontStyle int

// The font styles.
const (
	StyleNormal UFontStyle = iota
	StyleItalic
)

// String returns upstream's enum name, "NORMAL" or "ITALIC".
func (s UFontStyle) String() string {
	if s == StyleItalic {
		return "ITALIC"
	}
	return "NORMAL"
}

// UFontWeight is a CSS font weight, from 100 to 900 in steps of 100. The
// zero value means no weight.
type UFontWeight int

// The CSS weights.
const (
	WeightThin       UFontWeight = 100
	WeightExtraLight UFontWeight = 200
	WeightLight      UFontWeight = 300
	WeightNormal     UFontWeight = 400
	WeightMedium     UFontWeight = 500
	WeightSemiBold   UFontWeight = 600
	WeightBold       UFontWeight = 700
	WeightExtraBold  UFontWeight = 800
	WeightBlack      UFontWeight = 900
)

// WeightFromCssValue returns the weight for a CSS numeric weight, clamped to
// 100-900 and rounded to the nearest hundred (halves round up).
func WeightFromCssValue(weight int) UFontWeight {
	switch {
	case weight < 100:
		return WeightThin
	case weight > 900:
		return WeightBlack
	}
	return UFontWeight((weight + 50) / 100 * 100)
}

// Weight returns the CSS weight.
func (w UFontWeight) Weight() int { return int(w) }

// IsBold reports whether the weight is 700 or more.
func (w UFontWeight) IsBold() bool { return w >= WeightBold }

// String returns upstream's representation, for example "UFontWeight[700]".
func (w UFontWeight) String() string {
	return "UFontWeight[" + strconv.Itoa(int(w)) + "]"
}

// UFontFace combines a font style (normal or italic) and a CSS weight
// (100-900), the two independent axes of the CSS model. It is a comparable
// value. The zero value stands for upstream's null face: functions that
// take a face treat it as missing.
type UFontFace struct {
	style  UFontStyle
	weight UFontWeight
}

// NormalFace returns the normal face (weight 400, not italic).
func NormalFace() UFontFace { return UFontFace{StyleNormal, WeightNormal} }

// BoldFace returns the bold face (weight 700, not italic).
func BoldFace() UFontFace { return UFontFace{StyleNormal, WeightBold} }

// ItalicFace returns the italic face (weight 400, italic).
func ItalicFace() UFontFace { return UFontFace{StyleItalic, WeightNormal} }

// BoldItalicFace returns the bold italic face (weight 700, italic).
func BoldItalicFace() UFontFace { return UFontFace{StyleItalic, WeightBold} }

// NewUFontFace returns the face with a style and a weight.
func NewUFontFace(style UFontStyle, weight UFontWeight) UFontFace {
	return UFontFace{style, weight}
}

// NewUFontFaceCss returns the face with a style and a CSS numeric weight.
func NewUFontFaceCss(style UFontStyle, cssWeight int) UFontFace {
	return UFontFace{style, WeightFromCssValue(cssWeight)}
}

// UFontFaceFromCssWeight parses a CSS font-weight: a number or one of the
// keywords normal, bold, lighter and bolder. The face has the normal style.
// It reports false when the value cannot be parsed.
func UFontFaceFromCssWeight(cssValue string) (UFontFace, bool) {
	trimmed := strings.ToLower(strings.TrimFunc(cssValue, func(r rune) bool { return r <= ' ' }))
	switch trimmed {
	case "normal":
		return NormalFace(), true
	case "bold":
		return BoldFace(), true
	case "lighter":
		return NewUFontFaceCss(StyleNormal, 300), true
	case "bolder":
		return NewUFontFaceCss(StyleNormal, 800), true
	}
	w, ok := parseJavaInt(trimmed)
	if !ok {
		return UFontFace{}, false
	}
	return NewUFontFaceCss(StyleNormal, w), true
}

// IsZero reports whether f is the zero face, upstream's null.
func (f UFontFace) IsZero() bool { return f == UFontFace{} }

// Style returns the style.
func (f UFontFace) Style() UFontStyle { return f.style }

// Weight returns the weight.
func (f UFontFace) Weight() UFontWeight { return f.weight }

// CssWeight returns the CSS numeric weight.
func (f UFontFace) CssWeight() int { return int(f.weight) }

// IsItalic reports whether the face is italic.
func (f UFontFace) IsItalic() bool { return f.style == StyleItalic }

// IsBold reports whether the weight is 700 or more.
func (f UFontFace) IsBold() bool { return f.weight >= WeightBold }

// WithStyle returns the face with another style and the same weight.
func (f UFontFace) WithStyle(style UFontStyle) UFontFace {
	return UFontFace{style, f.weight}
}

// WithWeight returns the face with another CSS weight and the same style.
func (f UFontFace) WithWeight(cssWeight int) UFontFace {
	return UFontFace{f.style, WeightFromCssValue(cssWeight)}
}

// WithFontWeight returns the face with another weight and the same style.
func (f UFontFace) WithFontWeight(weight UFontWeight) UFontFace {
	return UFontFace{f.style, weight}
}

// The values of Java's TextAttribute weight constants.
const (
	textAttributeWeightExtraLight float32 = 0.5
	textAttributeWeightLight      float32 = 0.75
	textAttributeWeightRegular    float32 = 1.0
	textAttributeWeightSemiBold   float32 = 1.25
	textAttributeWeightMedium     float32 = 1.5
	textAttributeWeightBold       float32 = 2.0
	textAttributeWeightHeavy      float32 = 2.25
	textAttributeWeightUltraBold  float32 = 2.75
)

// ToTextAttributeWeight maps the CSS weight to the value of the matching
// Java TextAttribute.WEIGHT constant.
func (f UFontFace) ToTextAttributeWeight() float32 {
	w := f.weight
	switch {
	case w <= 200:
		return textAttributeWeightExtraLight
	case w <= 300:
		return textAttributeWeightLight
	case w <= 400:
		return textAttributeWeightRegular
	case w <= 500:
		return textAttributeWeightMedium
	case w <= 600:
		return textAttributeWeightSemiBold
	case w <= 700:
		return textAttributeWeightBold
	case w <= 800:
		return textAttributeWeightHeavy
	}
	return textAttributeWeightUltraBold
}

// DeriveFont applies the face to a font, as upstream does through Java's
// TextAttribute.WEIGHT and POSTURE. Java makes a font bold when the weight
// attribute is at least WEIGHT_BOLD, and italic for POSTURE_OBLIQUE; the
// fonts have no intermediate weights.
func (f UFontFace) DeriveFont(base AwtFont) AwtFont {
	style := AwtPlain
	if f.ToTextAttributeWeight() >= textAttributeWeightBold {
		style |= AwtBold
	}
	if f.IsItalic() {
		style |= AwtItalic
	}
	return base.Derive(style, base.Size())
}

// ToCssWeightString returns the CSS font-weight for SVG output.
func (f UFontFace) ToCssWeightString() string { return strconv.Itoa(int(f.weight)) }

// ToCssStyleString returns the CSS font-style for SVG output.
func (f UFontFace) ToCssStyleString() string {
	if f.IsItalic() {
		return "italic"
	}
	return "normal"
}

// String returns upstream's representation, for example
// "UFontFace[NORMAL, weight=400]".
func (f UFontFace) String() string {
	return "UFontFace[" + f.style.String() + ", weight=" + strconv.Itoa(int(f.weight)) + "]"
}
