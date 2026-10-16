// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"strings"
	"testing"
)

type doubleHack struct{}

func (doubleHack) TransformStringForSizeHack(s string) string { return s + s }

func TestStringBounderSvg(t *testing.T) {
	b := NewStringBounderSvg(nil)
	f := SansSerif(13)
	if got := b.CalculateDimension(f, ""); got != (Dimension{}) {
		t.Errorf("empty text: got %v, want 0x0", got)
	}
	one := b.CalculateDimension(f, "abc")
	if want := f.UnderlyingFont("abc").StringBounds("abc"); one != want || one.Width <= 0 {
		t.Errorf("got %v, want %v", one, want)
	}
	two := NewStringBounderSvg(doubleHack{}).CalculateDimension(f, "abc")
	if two.Width != 2*one.Width || two.Height != one.Height {
		t.Errorf("char size hack: got %v, want twice %v", two, one)
	}
	if !b.MatchesProperty("svg") || !b.MatchesProperty("SVG") || b.MatchesProperty("TIKZ") {
		t.Error("MatchesProperty")
	}
	if b.FileFormat() != FormatSVG {
		t.Errorf("FileFormat: %v", b.FileFormat())
	}
	if got := b.Descent(Build("Serif", BoldFace(), 24), "x"); got != 5.66015625 {
		t.Errorf("Descent: got %v, want 5.66015625", got)
	}
}

func TestStringBounderAwt(t *testing.T) {
	var b StringBounderAwt
	f := Serif(20)
	if b.FileFormat() != FormatPNG || b.MatchesProperty("SVG") {
		t.Error("FileFormat or MatchesProperty")
	}
	if got, want := b.CalculateDimension(f, "plain text"), JavaDimension(f, "plain text"); got != want {
		t.Errorf("plain text: got %v, want %v", got, want)
	}

	plainFont, boldFont := f, FontStyleBold.MutateFont(f)
	plain := func(s string) float64 { return JavaDimension(plainFont, s).Width }
	bold := func(s string) float64 { return JavaDimension(boldFont, s).Width }
	plainHeight, boldHeight := JavaDimension(plainFont, "x").Height, JavaDimension(boldFont, "x").Height
	tests := []struct {
		text          string
		width, height float64
	}{
		{"ab\uEEF2cd\uEEF1ef", plain("ab") + bold("cd") + plain("ef"), boldHeight},
		{"\uEEF2cd\uEEF1", bold("cd"), boldHeight},
		// An empty bold part measures 0x0.
		{"ab\uEEF2\uEEF1", plain("ab"), plainHeight},
		// Without a start marker, the end marker is measured as text.
		{"ab\uEEF1cd", plain("ab\uEEF1cd"), plainHeight},
		// A missing end marker extends the bold part to the end.
		{"x\uEEF2yz", plain("x") + bold("yz"), boldHeight},
	}
	for _, tt := range tests {
		got := b.CalculateDimension(f, tt.text)
		if got.Width != tt.width || got.Height != tt.height {
			t.Errorf("%q: got %v, want %v x %v", tt.text, got, tt.width, tt.height)
		}
	}
	if got := b.Descent(f, "x"); got != f.UnderlyingFont("x").Descent() {
		t.Errorf("Descent: %v", got)
	}
}

func TestBuildStyledStrings(t *testing.T) {
	got := buildStyledStrings("a\uEEF2b\uEEF1c\uEEF2d")
	var parts []string
	for _, s := range got {
		parts = append(parts, s.style.String()+"["+s.text+"]")
	}
	if want := "PLAIN[a] BOLD[b] PLAIN[c] BOLD[d]"; strings.Join(parts, " ") != want {
		t.Errorf("got %q, want %q", strings.Join(parts, " "), want)
	}
}

func TestStringBounderDefaults(t *testing.T) {
	var d StringBounderDefaults
	if d.MatchesProperty("SVG") {
		t.Error("MatchesProperty")
	}
	if got := d.Descent(SansSerif(9), "x"); got != 2 {
		t.Errorf("Descent: got %v, want 2", got)
	}
}
