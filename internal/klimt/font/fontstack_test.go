// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"slices"
	"testing"
)

func stackFamilies(s FontStack) []string {
	var families []string
	for _, f := range s.Fonts() {
		families = append(families, f.Family()+"."+f.Style().String())
	}
	return families
}

func TestFontStackFonts(t *testing.T) {
	tests := []struct {
		definition string
		want       []string
	}{
		{"SansSerif", []string{"SansSerif.plain"}},
		{"Arial, Helvetica", []string{"Dialog.plain", "SansSerif.plain"}},
		{`"Times New Roman", TimesRoman, serif`, []string{"Dialog.plain", "Serif.plain", "Serif.plain"}},
		{" \"DejaVu Sans Mono\" ,Courier", []string{"DejaVu Sans Mono.plain", "Monospaced.plain"}},
		{"Monospaced Bold, Serif-italic-12", []string{"Monospaced.bold", "Serif.italic"}},
		{"Courier,,", []string{"Monospaced.plain"}},
		{",Courier", []string{"Dialog.plain", "Monospaced.plain"}},
		{"", []string{"Dialog.plain"}},
		{",,", []string{"Dialog.plain"}},
		{" Serif\u00a0", []string{"Dialog.plain"}}, // U+00A0 is not Java whitespace
		{" Serif\t", []string{"Serif.plain"}},
	}
	for _, tt := range tests {
		if got := stackFamilies(BuildFontStack(tt.definition)); !slices.Equal(got, tt.want) {
			t.Errorf("%q: got %q, want %q", tt.definition, got, tt.want)
		}
	}
}

// firstMissing returns a code point that font a displays and font b does
// not.
func firstMissing(t *testing.T, a, b AwtFont) rune {
	t.Helper()
	for cp := rune(0); cp < 0x110000; cp++ {
		if a.CanDisplay(cp) && !b.CanDisplay(cp) {
			return cp
		}
	}
	t.Fatalf("%s displays nothing that %s cannot", a.FontName(), b.FontName())
	return 0
}

// TestFontStackFont checks that the stack picks the first family that can
// display the whole text, else the first family.
func TestFontStackFont(t *testing.T) {
	mono := NewAwtFont("Monospaced", AwtPlain, 12)
	sans := NewAwtFont("SansSerif", AwtPlain, 12)
	onlySans := string(firstMissing(t, sans, mono))

	stack := BuildFontStack("Courier, Arial")
	tests := []struct {
		text   string
		family string
	}{
		{"abc", "Monospaced"},
		{"", "Monospaced"},
		{"ab" + onlySans, "Dialog"},
		{"ab" + onlySans + "\uE000", "Monospaced"}, // no family displays U+E000
	}
	for _, tt := range tests {
		f := stack.Font(tt.text, BoldFace(), 17)
		if f.Family() != tt.family || f.Style() != AwtBold || f.Size() != 17 {
			t.Errorf("%q: got %s %v %d, want %s bold 17", tt.text, f.Family(), f.Style(), f.Size(), tt.family)
		}
	}
	if got := stack.CanDisplayUpTo(0, "ab"+onlySans); got != 2 {
		t.Errorf("CanDisplayUpTo(0) = %d, want 2", got)
	}
	if got := stack.CanDisplayUpTo(1, "ab"+onlySans); got != -1 {
		t.Errorf("CanDisplayUpTo(1) = %d, want -1", got)
	}

	// A single family is used whatever the text.
	single := BuildFontStack("Courier")
	if f := single.Font("ab"+onlySans, NormalFace(), 10); f.Family() != "Monospaced" {
		t.Errorf("single family: got %s", f.Family())
	}

	// The style written in the definition is dropped: the face decides.
	if f := BuildFontStack("Serif Bold").Font("x", ItalicFace(), 10); f.Style() != AwtItalic {
		t.Errorf("decoded style kept: got %v", f.Style())
	}
}

func TestFontStackSvgFamily(t *testing.T) {
	tests := map[string]string{
		"Serif":                    "serif",
		"SansSerif":                "sans-serif",
		"Monospaced":               "monospace",
		"sansserif":                "sansserif",
		"Arial":                    "Arial",
		`"Times New Roman", serif`: "'Times New Roman', serif",
	}
	for definition, want := range tests {
		if got := BuildFontStack(definition).SvgFamily(); got != want {
			t.Errorf("%q: got %q, want %q", definition, got, want)
		}
	}
}

func TestFontStackValue(t *testing.T) {
	a, b := BuildFontStack("Arial, Helvetica"), BuildFontStack("Arial, Helvetica")
	if a != b || a == BuildFontStack("Arial,Helvetica") {
		t.Error("stacks compare by definition")
	}
	if a.String() != "FontStack[Arial, Helvetica]" || a.FullDefinition() != "Arial, Helvetica" {
		t.Errorf("got %q, %q", a.String(), a.FullDefinition())
	}
}
