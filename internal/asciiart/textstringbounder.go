// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiart/TextStringBounder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiart

import "github.com/ekalinin/gopuml/internal/klimt/font"

// TextStringBounder measures text for the text formats (ATXT, UTXT): a text
// is one character high and as wide as its number of UTF-16 code units,
// whatever the font.
//
// Upstream also computes the code point count and Wcwidth.length of the
// text, then ignores them; they are not computed here.
type TextStringBounder struct {
	font.StringBounderDefaults
	fileFormat font.FileFormat
}

var _ font.StringBounder = (*TextStringBounder)(nil)

// NewTextStringBounder returns the measurer of a text format.
func NewTextStringBounder(fileFormat font.FileFormat) *TextStringBounder {
	return &TextStringBounder{fileFormat: fileFormat}
}

// FileFormat returns the format given to NewTextStringBounder.
func (b *TextStringBounder) FileFormat() font.FileFormat { return b.fileFormat }

// CalculateDimension returns the length of text in UTF-16 code units and a
// height of 1.
func (b *TextStringBounder) CalculateDimension(_ font.UFont, text string) font.Dimension {
	return font.Dimension{Width: float64(javaLength(text)), Height: 1}
}

// javaLength is Java's String.length: the number of UTF-16 code units.
func javaLength(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
