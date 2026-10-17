// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UText.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UText is a single line of text drawn with its baseline at the current
// position.
type UText struct {
	text        string
	font        *FontConfiguration
	orientation int
}

func newUText(text string, f *FontConfiguration, orientation int) *UText {
	text = strings.ReplaceAll(text, string(JawsBlockE1Newline), "↵")
	text = strings.ReplaceAll(text, string(JawsBlockE1Breakline), "⏎")
	return &UText{text: text, font: f, orientation: orientation}
}

// UTextBuild returns the text in a font configuration. The newline and
// break line markers of the preprocessor are shown as U+21B5 and U+23CE.
func UTextBuild(text string, f *FontConfiguration) *UText { return newUText(text, f, 0) }

// WithOrientation returns the text with another orientation in degrees.
func (t *UText) WithOrientation(orientation int) *UText { return newUText(t.text, t.font, orientation) }

// String is upstream's toString.
func (t *UText) String() string { return "UText[" + t.text + "]" }

// Text returns the text.
func (t *UText) Text() string { return t.text }

// FontConfiguration returns the font configuration.
func (t *UText) FontConfiguration() *FontConfiguration { return t.font }

// Descent returns the descent of the text.
func (t *UText) Descent(stringBounder font.StringBounder) float64 {
	return stringBounder.Descent(t.font.Font(), t.text)
}

// Orientation returns the orientation in degrees.
func (t *UText) Orientation() int { return t.orientation }

// CalculateDimension returns the size of the text.
func (t *UText) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return dim(stringBounder.CalculateDimension(t.font.Font(), t.text))
}
