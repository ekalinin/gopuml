// SPDX-License-Identifier: MIT
// Provenance: original

package font

import "testing"

// TestDecodeFontSpec checks the name and style Java's Font.decode returns on
// the reference JVM.
func TestDecodeFontSpec(t *testing.T) {
	tests := []struct {
		spec, name string
		style      AwtStyle
	}{
		{"Arial", "Arial", AwtPlain},
		{"Arial-12", "Arial", AwtPlain},
		{"Arial-bold-12", "Arial", AwtBold},
		{"Arial-bold", "Arial", AwtBold},
		{"Monospaced Bold", "Monospaced", AwtBold},
		{"Times New Roman", "Times New Roman", AwtPlain},
		{"Foo Bar 12", "Foo Bar", AwtPlain},
		{"Foo 12", "Foo", AwtPlain},
		{"Foo-", "Foo", AwtPlain},
		{"Serif-", "Serif", AwtPlain},
		{"Serif ", "Serif", AwtPlain},
		{" Serif", " Serif", AwtPlain},
		{"Serif--12", "Serif", AwtPlain},
		{"Serif-bold-", "Serif", AwtBold},
		{"Serif-0", "Serif", AwtPlain},
		{"Serif--5", "Serif", AwtPlain},
		{"Serif-+5", "Serif", AwtPlain},
		{"Serif-bolditalic-12", "Serif", AwtBoldItalic},
		{"Serif ITALIC", "Serif", AwtItalic},
		{"Serif plain 10", "Serif", AwtPlain},
		{"Serif-x-12", "Serif-x", AwtPlain},
		{"-12", "-12", AwtPlain},
		{"12", "12", AwtPlain},
		{"-bold", "-bold", AwtPlain},
		{"Serif - bold", "Serif -", AwtBold},
		{"Courier New-Bold", "Courier New", AwtBold},
		{"DejaVu Sans Mono-BOLD-14", "DejaVu Sans Mono", AwtBold},
		{"a-b c", "a-b c", AwtPlain},
		{"a b-c", "a b-c", AwtPlain},
		{"Serif 12 bold", "Serif 12", AwtBold},
		{"Serif-12-bold", "Serif-12", AwtBold},
		{"", "", AwtPlain},
		{"-", "", AwtPlain},
		{" ", "", AwtPlain},
		{"--", "", AwtPlain},
		{"Serif-bold-bold", "Serif-bold", AwtBold},
		{"Serif bold bold", "Serif bold", AwtBold},
		{"Serif-italic-bold", "Serif-italic", AwtBold},
		{"Serif- bold", "Serif-", AwtBold},
		{"Serif -bold", "Serif ", AwtBold},
		{"Serif-bold 12", "Serif-bold", AwtPlain},
		{"Serif 99999999999", "Serif 99999999999", AwtPlain},
		{"Serif-1x", "Serif-1x", AwtPlain},
		{"Serif-١٢", "Serif", AwtPlain},
		{"Serif-BoLdItAlIc", "Serif", AwtBoldItalic},
		{"Serif--", "Serif", AwtPlain},
		{"Serif-bold--", "Serif-bold", AwtPlain},
		{"-Serif", "-Serif", AwtPlain},
		{"Serif-plain-", "Serif", AwtPlain},
		{"Serif plain", "Serif", AwtPlain},
		{"Serif  bold", "Serif ", AwtBold},
		{"Serif--bold", "Serif-", AwtBold},
		{"bold", "bold", AwtPlain},
		{"Serif\tbold", "Serif\tbold", AwtPlain},
	}
	for _, tt := range tests {
		name, style := decodeFontSpec(tt.spec)
		if name != tt.name || style != tt.style {
			t.Errorf("decodeFontSpec(%q) = %q, %v; want %q, %v", tt.spec, name, style, tt.name, tt.style)
		}
	}
}

func TestParseJavaInt(t *testing.T) {
	tests := []struct {
		s    string
		want int
		ok   bool
	}{
		{"12", 12, true},
		{"+5", 5, true},
		{"-5", -5, true},
		{"0", 0, true},
		{"2147483647", 2147483647, true},
		{"-2147483648", -2147483648, true},
		{"2147483648", 0, false},
		{"99999999999", 0, false},
		{"١٢", 12, true},
		{"１２", 12, true},
		{"", 0, false},
		{"+", 0, false},
		{"-", 0, false},
		{"1x", 0, false},
		{" 1", 0, false},
		{"1_0", 0, false},
		{"𝟏", 0, false}, // a digit outside the BMP is two UTF-16 units in Java
	}
	for _, tt := range tests {
		got, ok := parseJavaInt(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("parseJavaInt(%q) = %d, %v; want %d, %v", tt.s, got, ok, tt.want, tt.ok)
		}
	}
}

// TestFamilyResolution checks the family, face name and PostScript name
// Java gives on the reference JVM.
func TestFamilyResolution(t *testing.T) {
	tests := []struct {
		name     string
		style    AwtStyle
		family   string
		table    string
		fontName string
		psName   string
	}{
		{"SansSerif", AwtPlain, "SansSerif", "SansSerif", "SansSerif.plain", "SansSerif.plain"},
		{"sansserif", AwtBold, "SansSerif", "SansSerif", "SansSerif.bold", "SansSerif.bold"},
		{"Serif", AwtItalic, "Serif", "Serif", "Serif.italic", "Serif.italic"},
		{"MONOSPACED", AwtBoldItalic, "Monospaced", "Monospaced", "Monospaced.bolditalic", "Monospaced.bolditalic"},
		{"Dialog", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"DialogInput", AwtBold, "DialogInput", "Monospaced", "DialogInput.bold", "DialogInput.bold"},
		{"dialoginput", AwtPlain, "DialogInput", "Monospaced", "DialogInput.plain", "DialogInput.plain"},
		{"Courier", AwtPlain, "Monospaced", "Monospaced", "Monospaced.plain", "Monospaced.plain"},
		{"courier", AwtBold, "Monospaced", "Monospaced", "Monospaced.bold", "Monospaced.bold"},
		{"Helvetica", AwtPlain, "SansSerif", "SansSerif", "SansSerif.plain", "SansSerif.plain"},
		{"TimesRoman", AwtPlain, "Serif", "Serif", "Serif.plain", "Serif.plain"},
		{"Arial", AwtBold, "Dialog", "Dialog", "Dialog.bold", "Dialog.bold"},
		{"Times", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"Times New Roman", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"Verdana", AwtItalic, "Dialog", "Dialog", "Dialog.italic", "Dialog.italic"},
		{"Courier New", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"monospace", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"sans-serif", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"", AwtPlain, "Dialog", "Dialog", "Dialog.plain", "Dialog.plain"},
		{"DejaVu Sans", AwtPlain, "DejaVu Sans", "SansSerif", "DejaVu Sans", "DejaVuSans"},
		{"DejaVu Sans", AwtBold, "DejaVu Sans", "SansSerif", "DejaVu Sans Bold", "DejaVuSans-Bold"},
		{"DejaVu Sans", AwtItalic, "DejaVu Sans", "SansSerif", "DejaVu Sans Oblique", "DejaVuSans-Oblique"},
		{"dejavu sans", AwtBoldItalic, "DejaVu Sans", "SansSerif", "DejaVu Sans Bold Oblique", "DejaVuSans-BoldOblique"},
		{"DejaVu Serif", AwtItalic, "DejaVu Serif", "Serif", "DejaVu Serif Italic", "DejaVuSerif-Italic"},
		{"DejaVu Serif", AwtBoldItalic, "DejaVu Serif", "Serif", "DejaVu Serif Bold Italic", "DejaVuSerif-BoldItalic"},
		{"DejaVu Sans Mono", AwtPlain, "DejaVu Sans Mono", "Monospaced", "DejaVu Sans Mono", "DejaVuSansMono"},
		{"DejaVu Sans Mono", AwtBoldItalic, "DejaVu Sans Mono", "Monospaced", "DejaVu Sans Mono Bold Oblique", "DejaVuSansMono-BoldOblique"},
	}
	for _, tt := range tests {
		f := NewAwtFont(tt.name, tt.style, 12)
		if f.Family() != tt.family || f.Metrics().Family != tt.table || f.FontName() != tt.fontName || f.PSName() != tt.psName {
			t.Errorf("NewAwtFont(%q, %v): family %q, table %q, font name %q, PS name %q; want %q, %q, %q, %q",
				tt.name, tt.style, f.Family(), f.Metrics().Family, f.FontName(), f.PSName(),
				tt.family, tt.table, tt.fontName, tt.psName)
		}
		if f.Style() != tt.style || f.Metrics().Style != tt.style.String() || f.Size() != 12 {
			t.Errorf("NewAwtFont(%q, %v): style %v, table style %s, size %d", tt.name, tt.style, f.Style(), f.Metrics().Style, f.Size())
		}
	}
}

// TestAwtFontMeasures checks values printed by the reference JVM for
// "Hello World iiiWWW 123 é中а" at size 13. Serif bold is taller than the
// other fonts.
func TestAwtFontMeasures(t *testing.T) {
	const text = "Hello World iiiWWW 123 é中а"
	tests := []struct {
		name          string
		style         AwtStyle
		width, height float64
	}{
		{"SansSerif", AwtPlain, 185.47216796875, 15.1328125},
		{"SansSerif", AwtBold, 208.0, 15.1328125},
		{"Serif", AwtPlain, 192.20068359375, 15.1328125},
		{"Serif", AwtBold, 211.59912109375, 15.2724609375},
		{"Monospaced", AwtPlain, 203.4931640625, 15.1328125},
		{"Monospaced", AwtBold, 203.4931640625, 15.1328125},
		{"Arial", AwtPlain, 185.47216796875, 15.1328125},
		{"Courier", AwtPlain, 203.4931640625, 15.1328125},
		{"DialogInput", AwtPlain, 203.4931640625, 15.1328125},
		{"Helvetica", AwtBold, 208.0, 15.1328125},
		{"TimesRoman", AwtBold, 211.59912109375, 15.2724609375},
	}
	for _, tt := range tests {
		f := NewAwtFont(tt.name, tt.style, 13)
		got := f.StringBounds(text)
		if got.Width != tt.width || got.Height != tt.height || f.Descent() != 3.06591796875 {
			t.Errorf("%s %v: got %v x %v descent %v, want %v x %v descent 3.06591796875",
				tt.name, tt.style, got.Width, got.Height, f.Descent(), tt.width, tt.height)
		}
	}

	f := NewAwtFont("SansSerif", AwtPlain, 100)
	if got := f.StringBounds("AV").Width; got != 136.81640625 {
		t.Errorf("AV: got %v, want 136.81640625 (no kerning)", got)
	}
	if got := f.StringBounds("a\tb").Width; got != 124.755859375 {
		t.Errorf("a<tab>b: got %v, want 124.755859375", got)
	}
	if got := f.StringBounds("AV😀").Width; got != 241.064453125 {
		t.Errorf("AV + emoji: got %v, want 241.064453125", got)
	}
	if got := f.StringBounds("").Height; got != 116.40625 {
		t.Errorf("empty text height: got %v, want 116.40625", got)
	}
	if got, want := f.Advance(0xE000), 60.009765625; got != want {
		t.Errorf("missing glyph advance: got %v, want %v", got, want)
	}
	if f.Ascent() != 92.822265625 || f.Descent() != 23.583984375 || f.Leading() != 0 {
		t.Errorf("line metrics: got %v %v %v", f.Ascent(), f.Descent(), f.Leading())
	}
}

func TestAwtFontCanDisplay(t *testing.T) {
	f := NewAwtFont("SansSerif", AwtPlain, 12)
	if !f.CanDisplay('a') || f.CanDisplay(0xE000) {
		t.Error("CanDisplay")
	}
	if got := f.CanDisplayUpTo("abc"); got != -1 {
		t.Errorf("CanDisplayUpTo(abc) = %d, want -1", got)
	}
	if got := f.CanDisplayUpTo(""); got != -1 {
		t.Errorf("CanDisplayUpTo(\"\") = %d, want -1", got)
	}
	if got := f.CanDisplayUpTo("éa\uE000b"); got != 3 {
		t.Errorf("CanDisplayUpTo = %d, want byte index 3", got)
	}
}

func TestAwtFontDeriveAndDecode(t *testing.T) {
	f := DecodeAwtFont("Courier New-Bold")
	if f.Family() != "Dialog" || f.Style() != AwtBold || f.Size() != 12 {
		t.Errorf("DecodeAwtFont: got %s %v %d", f.Family(), f.Style(), f.Size())
	}
	g := DecodeAwtFont("Monospaced Bold").Derive(AwtItalic, 20)
	if g.Family() != "Monospaced" || g.Style() != AwtItalic || g.Size() != 20 || g.Metrics().Style != "italic" {
		t.Errorf("Derive: got %s %v %d %s", g.Family(), g.Style(), g.Size(), g.Metrics().Style)
	}
	if !g.IsItalic() || g.IsBold() || !g.IsLogical() || NewAwtFont("DejaVu Sans", AwtPlain, 1).IsLogical() {
		t.Error("style or logical flags")
	}
	if NewAwtFont("Serif", AwtBoldItalic, 12) != NewAwtFont("serif", AwtBoldItalic, 12) {
		t.Error("fonts of the same family, style and size differ")
	}
}

func TestAwtStyleString(t *testing.T) {
	for style, want := range map[AwtStyle]string{AwtPlain: "plain", AwtBold: "bold", AwtItalic: "italic", AwtBoldItalic: "bolditalic"} {
		if got := style.String(); got != want {
			t.Errorf("%d: got %q, want %q", style, got, want)
		}
	}
}
