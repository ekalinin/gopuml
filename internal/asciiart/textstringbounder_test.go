// SPDX-License-Identifier: MIT
// Provenance: original

package asciiart

import (
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

func TestTextStringBounder(t *testing.T) {
	b := NewTextStringBounder(font.FormatUTXT)
	if b.FileFormat() != font.FormatUTXT {
		t.Errorf("FileFormat = %v", b.FileFormat())
	}
	if b.MatchesProperty("SVG") {
		t.Error("MatchesProperty(SVG) is true")
	}
	f := font.Serif(14)
	for text, width := range map[string]float64{"": 0, "abc": 3, "日本": 2, "a😀": 3} {
		if got := b.CalculateDimension(f, text); got != (font.Dimension{Width: width, Height: 1}) {
			t.Errorf("CalculateDimension(%q) = %v, want width %v (UTF-16 units) and height 1", text, got, width)
		}
	}
}
