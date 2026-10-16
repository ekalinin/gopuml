// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"math"
	"slices"
	"testing"
)

func TestWeightFromCssValue(t *testing.T) {
	tests := map[int]UFontWeight{
		-5: WeightThin, 0: WeightThin, 99: WeightThin, 100: WeightThin, 149: WeightThin,
		150: WeightExtraLight, 400: WeightNormal, 449: WeightNormal, 450: WeightMedium,
		700: WeightBold, 849: WeightExtraBold, 850: WeightBlack, 900: WeightBlack, 5000: WeightBlack,
	}
	for in, want := range tests {
		if got := WeightFromCssValue(in); got != want {
			t.Errorf("WeightFromCssValue(%d) = %v, want %v", in, got, want)
		}
	}
	if !WeightBold.IsBold() || WeightSemiBold.IsBold() || WeightBold.Weight() != 700 {
		t.Error("IsBold or Weight")
	}
	if WeightBold.String() != "UFontWeight[700]" {
		t.Errorf("String: %q", WeightBold.String())
	}
}

func TestUFontFace(t *testing.T) {
	if f := NormalFace(); f.IsBold() || f.IsItalic() || f.CssWeight() != 400 || f.Style() != StyleNormal {
		t.Errorf("NormalFace: %v", f)
	}
	if f := BoldItalicFace(); !f.IsBold() || !f.IsItalic() || f.Weight() != WeightBold {
		t.Errorf("BoldItalicFace: %v", f)
	}
	if BoldFace() != NewUFontFace(StyleNormal, WeightBold) || ItalicFace() != NewUFontFaceCss(StyleItalic, 420) {
		t.Error("constructors")
	}
	if got := ItalicFace().WithWeight(650); got != NewUFontFace(StyleItalic, WeightBold) {
		t.Errorf("WithWeight: %v", got)
	}
	if got := BoldFace().WithStyle(StyleItalic); got != BoldItalicFace() {
		t.Errorf("WithStyle: %v", got)
	}
	if got := NormalFace().WithFontWeight(WeightLight); got.CssWeight() != 300 {
		t.Errorf("WithFontWeight: %v", got)
	}
	if got := BoldItalicFace().String(); got != "UFontFace[ITALIC, weight=700]" {
		t.Errorf("String: %q", got)
	}
	if BoldItalicFace().ToCssWeightString() != "700" || BoldItalicFace().ToCssStyleString() != "italic" ||
		NormalFace().ToCssStyleString() != "normal" {
		t.Error("CSS strings")
	}
	if !(UFontFace{}).IsZero() || NormalFace().IsZero() {
		t.Error("IsZero")
	}
}

func TestUFontFaceFromCssWeight(t *testing.T) {
	tests := []struct {
		in   string
		want UFontFace
		ok   bool
	}{
		{"normal", NormalFace(), true},
		{" BOLD ", BoldFace(), true},
		{"lighter", NewUFontFaceCss(StyleNormal, 300), true},
		{"bolder", NewUFontFaceCss(StyleNormal, 800), true},
		{"550", NewUFontFaceCss(StyleNormal, 600), true},
		{"+700", BoldFace(), true},
		{"-1", NewUFontFaceCss(StyleNormal, 100), true},
		{"heavy", UFontFace{}, false},
		{"", UFontFace{}, false},
		{"7OO", UFontFace{}, false},
	}
	for _, tt := range tests {
		got, ok := UFontFaceFromCssWeight(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("UFontFaceFromCssWeight(%q) = %v, %v; want %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

// TestUFontFaceDeriveFont checks that Java makes a font bold from weight 700
// on, as TextAttribute.WEIGHT_BOLD, and that the reference JVM measures the
// intermediate weights as plain or bold.
func TestUFontFaceDeriveFont(t *testing.T) {
	base := NewAwtFont("SansSerif", AwtBoldItalic, 13)
	tests := []struct {
		face  UFontFace
		style AwtStyle
		attr  float32
	}{
		{NewUFontFaceCss(StyleNormal, 100), AwtPlain, 0.5},
		{NewUFontFaceCss(StyleNormal, 300), AwtPlain, 0.75},
		{NormalFace(), AwtPlain, 1.0},
		{NewUFontFaceCss(StyleNormal, 500), AwtPlain, 1.5},
		{NewUFontFaceCss(StyleNormal, 600), AwtPlain, 1.25},
		{BoldFace(), AwtBold, 2.0},
		{NewUFontFaceCss(StyleItalic, 800), AwtBoldItalic, 2.25},
		{NewUFontFaceCss(StyleItalic, 900), AwtBoldItalic, 2.75},
		{ItalicFace(), AwtItalic, 1.0},
	}
	for _, tt := range tests {
		if got := tt.face.ToTextAttributeWeight(); got != tt.attr {
			t.Errorf("%v: text attribute weight %v, want %v", tt.face, got, tt.attr)
		}
		f := tt.face.DeriveFont(base)
		if f.Style() != tt.style || f.Size() != 13 || f.Family() != "SansSerif" {
			t.Errorf("%v: got %s %v %d", tt.face, f.Family(), f.Style(), f.Size())
		}
	}
	// Values printed by the reference JVM for WEIGHT_SEMIBOLD and WEIGHT_HEAVY.
	const text = "Hello World iiiWWW 123 é中а"
	semi := NewUFontFaceCss(StyleNormal, 600).DeriveFont(base).StringBounds(text).Width
	heavy := NewUFontFaceCss(StyleNormal, 800).DeriveFont(base).StringBounds(text).Width
	if semi != 185.47216796875 || heavy != 208.0 {
		t.Errorf("semibold %v, heavy %v; want 185.47216796875, 208", semi, heavy)
	}
}

func TestUFont(t *testing.T) {
	f := Build("Arial, Helvetica", BoldFace(), 14)
	if f.Size() != 14 || f.Size2D() != 14 || f.FontFace() != BoldFace() || f.FontStack() != BuildFontStack("Arial, Helvetica") {
		t.Errorf("accessors: %v", f)
	}
	if Build("Arial, Helvetica", BoldFace(), 14) != f || Build("Arial", BoldFace(), 14) == f {
		t.Error("equality")
	}
	fonts := map[UFont]int{f: 1}
	if fonts[Build("Arial, Helvetica", BoldFace(), 14)] != 1 {
		t.Error("map key")
	}
	if got := Build("Serif", UFontFace{}, 10).FontFace(); got != NormalFace() {
		t.Errorf("zero face: got %v", got)
	}
	if got := f.WithFontFace(UFontFace{}); got != f {
		t.Error("WithFontFace(zero) changed the font")
	}
	if got := f.WithFontFace(ItalicFace()).FontFace(); got != ItalicFace() {
		t.Errorf("WithFontFace: %v", got)
	}
	sizes := map[float32]int{10.9: 10, -2.5: -2, float32(math.NaN()): 0, 1e20: math.MaxInt32, -1e20: math.MinInt32}
	for in, want := range sizes {
		if got := f.WithSize(in).Size(); got != want {
			t.Errorf("WithSize(%v) = %d, want %d", in, got, want)
		}
	}
	if got := ByDefault(30); got != SansSerif(12) {
		t.Errorf("ByDefault: %v", got)
	}
	if Serif(9).FontStack().FullDefinition() != "Serif" || Monospace(9).FontStack().FullDefinition() != "Monospaced" {
		t.Error("factories")
	}
}

func TestUFontString(t *testing.T) {
	tests := []struct {
		font UFont
		want string
	}{
		{SansSerif(12), "SansSerif.plain/12"},
		{Build("Arial", BoldFace(), 14), "Dialog.bold/14"},
		{Build("Courier", BoldItalicFace(), 9), "Monospaced.bolditalic/9"},
		{Build("DejaVu Sans", NormalFace(), 10), "DejaVu Sans.plain/10"},
		{Build("DejaVu Serif", BoldItalicFace(), 10), "DejaVu Serif Bold Italic.bolditalic/10"},
	}
	for _, tt := range tests {
		if got := tt.font.String(); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestUFontFamily(t *testing.T) {
	f := Build(`"DejaVu Sans", Arial`, BoldFace(), 12)
	if got := f.Family("x", ContextSVG); got != "'DejaVu Sans', Arial" {
		t.Errorf("SVG: %q", got)
	}
	if got := f.Family("x", ContextPDF); got != `"DejaVu Sans", Arial` {
		t.Errorf("PDF: %q", got)
	}
	if got := f.Family("x", ContextEPS); got != "DejaVuSans-Bold" {
		t.Errorf("EPS: %q", got)
	}
	if got := SansSerif(12).Family("x", ContextSVG); got != "sans-serif" {
		t.Errorf("SVG logical: %q", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("G2D: no panic")
		}
	}()
	f.Family("x", ContextG2D)
}

func TestFontStyleMutateFont(t *testing.T) {
	f := Build("Serif", NewUFontFaceCss(StyleNormal, 300), 12)
	if got := FontStyleBold.MutateFont(f).FontFace(); got != BoldFace() {
		t.Errorf("bold: %v", got)
	}
	if got := FontStyleItalic.MutateFont(f).FontFace(); got != NewUFontFaceCss(StyleItalic, 300) {
		t.Errorf("italic: %v", got)
	}
	if got := FontStylePlain.MutateFont(Build("Serif", BoldItalicFace(), 12)).FontFace(); got != NormalFace() {
		t.Errorf("plain: %v", got)
	}
	for _, s := range []FontStyle{FontStyleUnderline, FontStyleStrike, FontStyleWave, FontStyleBackcolor} {
		if got := s.MutateFont(f); got != f {
			t.Errorf("%v changed the font", s)
		}
	}
}

func TestFontStyle(t *testing.T) {
	if got := FontStyleStrike.Starters(false); !slices.Equal(got, []string{"<s", "<S", "<d", "<D"}) {
		t.Errorf("strike starters: %q", got)
	}
	if got := FontStyleBold.Starters(true); !slices.Equal(got, []string{"**"}) {
		t.Errorf("bold creole starters: %q", got)
	}
	if got := FontStyleBackcolor.Starters(true); !slices.Equal(got, []string{"<b", "<B"}) {
		t.Errorf("backcolor starters: %q", got)
	}
	for s := FontStylePlain; s <= FontStyleBackcolor; s++ {
		if s.RegexActivationPattern() == "" || s.RegexDeactivationPattern() == "" ||
			s.UbrexActivationPattern() == "" || s.UbrexDeactivationPattern() == "" {
			t.Errorf("%v: missing pattern", s)
		}
	}
	if FontStyleBold.RegexActivationPattern() != `\<[bB]\>` {
		t.Errorf("bold regex: %q", FontStyleBold.RegexActivationPattern())
	}
	if !FontStyleWave.CanHaveExtendedColor() || FontStyleBold.CanHaveExtendedColor() {
		t.Error("CanHaveExtendedColor")
	}
	if FontStyleStrike.UbrexCreoleSyntax() != "--" || FontStyleBackcolor.String() != "BACKCOLOR" {
		t.Error("creole syntax or name")
	}
	defer func() {
		if recover() == nil {
			t.Error("plain creole syntax: no panic")
		}
	}()
	FontStylePlain.UbrexCreoleSyntax()
}

func TestFontPosition(t *testing.T) {
	f := SansSerif(14)
	if got := FontPositionNormal.Mute(f); got != f {
		t.Errorf("normal: %v", got)
	}
	if got := FontPositionExposant.Mute(f).Size(); got != 11 {
		t.Errorf("exposant: size %d", got)
	}
	if got := FontPositionIndice.Mute(SansSerif(4)).Size(); got != 2 {
		t.Errorf("indice minimum: size %d", got)
	}
	if FontPositionExposant.Space() != -6 || FontPositionIndice.Space() != 3 || FontPositionNormal.Space() != 0 {
		t.Error("Space")
	}
	if FontPositionExposant.HtmlTag() != "sup" || FontPositionIndice.HtmlTag() != "sub" {
		t.Error("HtmlTag")
	}
}

func TestFileFormat(t *testing.T) {
	tests := []struct {
		format             FileFormat
		name, suffix, mime string
	}{
		{FormatSVG, "SVG", ".svg", "image/svg+xml"},
		{FormatPNG, "PNG", ".png", "image/png"},
		{FormatEPSText, "EPS_TEXT", ".eps", "application/postscript"},
		{FormatXMICustom, "XMI_CUSTOM", ".xmi_custom", "application/vnd.xmi+xml"},
		{FormatXMIArgo, "XMI_ARGO", ".xmi", "application/vnd.xmi+xml"},
		{FormatLATEXNoPreamble, "LATEX_NO_PREAMBLE", ".tex", "application/x-latex"},
		{FormatBraillePNG, "BRAILLE_PNG", ".braille.png", "image/png"},
		{FormatSVGDeterministic, "SVG_DETERMINISTIC", ".svg", "image/svg+xml"},
		{FormatUTXT, "UTXT", ".utxt", "text/plain;charset=UTF-8"},
		{FormatPNGEmpty, "PNG_EMPTY", ".png_empty", "image/png"},
	}
	for _, tt := range tests {
		if tt.format.String() != tt.name || tt.format.FileSuffix() != tt.suffix || tt.format.MimeType() != tt.mime {
			t.Errorf("%v: got %q %q %q", tt.name, tt.format.String(), tt.format.FileSuffix(), tt.format.MimeType())
		}
	}
	for name, want := range map[string]FileFormat{"svg": FormatSVG, "txt": FormatATXT, "eps": FormatLATEXDeterministic, "png": FormatPNG} {
		if got, ok := FileFormatFromCli(name); !ok || got != want {
			t.Errorf("FileFormatFromCli(%q) = %v, %v; want %v", name, got, ok, want)
		}
	}
	if _, ok := FileFormatFromCli("SVG"); ok {
		t.Error("FileFormatFromCli is case-sensitive")
	}
	if !FormatEPSText.IsEps() || FormatPDF.IsEps() || !FormatXMIStar.IsXmi() || FormatSVG.IsXmi() {
		t.Error("IsEps or IsXmi")
	}
	if !FormatPNG.DoesSupportMetadata() || FormatSVGDeterministic.DoesSupportMetadata() {
		t.Error("DoesSupportMetadata")
	}
	if FileFormat(99).String() != "FileFormat(99)" || FileFormat(-1).MimeType() != "" {
		t.Error("out of range format")
	}
}
