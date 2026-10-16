// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"strings"
	"unicode"
)

// AwtStyle is a java.awt.Font style: plain, or a combination of bold and
// italic, with Java's values.
type AwtStyle int

// The java.awt.Font styles.
const (
	AwtPlain      AwtStyle = 0
	AwtBold       AwtStyle = 1
	AwtItalic     AwtStyle = 2
	AwtBoldItalic AwtStyle = AwtBold | AwtItalic
)

// IsBold reports whether the style includes bold.
func (s AwtStyle) IsBold() bool { return s&AwtBold != 0 }

// IsItalic reports whether the style includes italic.
func (s AwtStyle) IsItalic() bool { return s&AwtItalic != 0 }

// String returns the style name Java uses in logical font names: "plain",
// "bold", "italic" or "bolditalic".
func (s AwtStyle) String() string {
	switch s & AwtBoldItalic {
	case AwtBold:
		return "bold"
	case AwtItalic:
		return "italic"
	case AwtBoldItalic:
		return "bolditalic"
	}
	return "plain"
}

// family is a font family as the reference JVM resolves it.
type family struct {
	// name is Java's Font.getFamily().
	name string
	// table is the embedded metrics table that measures the family.
	table string
	// logical families have font names like "SansSerif.bold". The other
	// ones are physical fonts with their own face names.
	logical bool
	// psName and italicName build the face names of physical fonts.
	psName, italicName string
}

var (
	familySansSerif  = &family{name: "SansSerif", table: "sansserif", logical: true}
	familySerif      = &family{name: "Serif", table: "serif", logical: true}
	familyMonospaced = &family{name: "Monospaced", table: "monospaced", logical: true}
	familyDialog     = &family{name: "Dialog", table: "dialog", logical: true}
	// DialogInput maps to the monospaced fonts: on the reference JVM its
	// coverage and advances equal Monospaced in every style.
	familyDialogInput = &family{name: "DialogInput", table: "monospaced", logical: true}
)

// families maps lower-case font names to the family Java resolves them to
// on the reference JVM (Linux, fonts-dejavu only). Any other name resolves
// to Dialog, as Java does for unknown fonts.
var families = map[string]*family{
	"sansserif":   familySansSerif,
	"serif":       familySerif,
	"monospaced":  familyMonospaced,
	"dialog":      familyDialog,
	"dialoginput": familyDialogInput,
	// Java's compatibility names for the old JDK 1.0 font names.
	"timesroman": familySerif,
	"helvetica":  familySansSerif,
	"courier":    familyMonospaced,
	// The physical DejaVu fonts behind the logical fonts. They have the
	// advances of the logical fonts but cover fewer code points, because
	// they lack the fallback fonts of the logical ones; the tables measure
	// them as the logical font, which is exact for the code points they
	// cover.
	"dejavu sans":      {name: "DejaVu Sans", table: "sansserif", psName: "DejaVuSans", italicName: "Oblique"},
	"dejavu serif":     {name: "DejaVu Serif", table: "serif", psName: "DejaVuSerif", italicName: "Italic"},
	"dejavu sans mono": {name: "DejaVu Sans Mono", table: "monospaced", psName: "DejaVuSansMono", italicName: "Oblique"},
}

// resolveFamily returns the family Java uses for a font name.
func resolveFamily(name string) *family {
	if f, ok := families[strings.ToLower(name)]; ok {
		return f
	}
	return familyDialog
}

// AwtFont stands in for the java.awt.Font that upstream gets from
// UFont.getUnderlayingFont: a family as the reference JVM resolves it, a
// style and an integer size. It measures text with the embedded metrics
// tables, exactly as Java2D does with fractional metrics.
//
// The zero value is not usable; build fonts with NewAwtFont or DecodeAwtFont.
type AwtFont struct {
	family  *family
	style   AwtStyle
	size    int
	metrics *Metrics
}

// NewAwtFont returns the font Java's new Font(name, style, size) gives on the
// reference JVM.
func NewAwtFont(name string, style AwtStyle, size int) AwtFont {
	return newAwtFont(resolveFamily(name), style, size)
}

func newAwtFont(f *family, style AwtStyle, size int) AwtFont {
	style &= AwtBoldItalic
	return AwtFont{family: f, style: style, size: size, metrics: lookupMetrics(f.table, style)}
}

// DecodeAwtFont returns the font Java's Font.decode(spec) gives: spec is
// "name", "name-style", "name-size" or "name-style-size", or the same with
// spaces, where style is plain, bold, italic or bolditalic in any case. An
// unparsable size is ignored and the size is 12, Java's default.
func DecodeAwtFont(spec string) AwtFont {
	name, style := decodeFontSpec(spec)
	return NewAwtFont(name, style, 12)
}

// decodeFontSpec splits a Font.decode specification into the font name and
// style, with the results Java gives. The separator is '-' when the last
// hyphen comes after the last space, else ' '. The last token is a size when
// it is a Java integer; the token before the size, or the last token when
// there is no size, is a style when it names one. Otherwise it stays in the
// name. A separator left at the end of the name is dropped.
func decodeFontSpec(spec string) (string, AwtStyle) {
	sep := byte(' ')
	if strings.LastIndexByte(spec, '-') > strings.LastIndexByte(spec, ' ') {
		sep = '-'
	}
	n := len(spec)
	last := strings.LastIndexByte(spec, sep)
	// end is where the name and style end; styleSep is the separator
	// before the style token, or -1.
	end, styleSep := n, -1
	if last > 0 {
		end = last
		styleSep = strings.LastIndexByte(spec[:last], sep)
		if last+1 < n && !isJavaInt(spec[last+1:]) {
			end, styleSep = n, last
		}
	}
	if styleSep >= 0 {
		if style, ok := parseAwtStyle(spec[styleSep+1 : end]); ok {
			return spec[:styleSep], style
		}
	}
	name := spec[:end]
	if name != "" && name[len(name)-1] == sep {
		name = name[:len(name)-1]
	}
	return name, AwtPlain
}

// parseAwtStyle parses a Font.decode style name, in any case.
func parseAwtStyle(s string) (AwtStyle, bool) {
	switch strings.ToLower(s) {
	case "plain":
		return AwtPlain, true
	case "bold":
		return AwtBold, true
	case "italic":
		return AwtItalic, true
	case "bolditalic":
		return AwtBoldItalic, true
	}
	return AwtPlain, false
}

// isJavaInt reports whether Java's Integer.parseInt accepts s.
func isJavaInt(s string) bool {
	_, ok := parseJavaInt(s)
	return ok
}

// parseJavaInt parses s as Java's Integer.parseInt does: an optional sign,
// then decimal digits of any script (one UTF-16 unit each), with a value
// that fits in 32 bits.
func parseJavaInt(s string) (int, bool) {
	limit := int64(1<<31 - 1)
	negative := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		if s[0] == '-' {
			negative = true
			limit++
		}
		s = s[1:]
	}
	if s == "" {
		return 0, false
	}
	var v int64
	for _, r := range s {
		d, ok := decimalDigit(r)
		if !ok {
			return 0, false
		}
		v = v*10 + int64(d)
		if v > limit {
			return 0, false
		}
	}
	if negative {
		v = -v
	}
	return int(v), true
}

// decimalDigit returns the value of a BMP decimal digit of any script.
// Unicode encodes those digits in runs from zero to nine.
func decimalDigit(r rune) (int, bool) {
	if r > 0xffff || !unicode.Is(unicode.Nd, r) {
		return 0, false
	}
	for _, rg := range unicode.Nd.R16 {
		if lo, hi := rune(rg.Lo), rune(rg.Hi); r >= lo && r <= hi {
			return int(r-lo) % 10, true
		}
	}
	return 0, false
}

// Derive returns the font with another style and size, as Java's
// deriveFont(int style, float size).
func (f AwtFont) Derive(style AwtStyle, size int) AwtFont {
	return newAwtFont(f.family, style, size)
}

// Family returns the family name, as Java's Font.getFamily(): "SansSerif",
// "Serif", "Monospaced", "Dialog", "DialogInput" or a physical DejaVu family.
func (f AwtFont) Family() string { return f.family.name }

// IsLogical reports whether the font is one of Java's logical fonts rather
// than a physical font.
func (f AwtFont) IsLogical() bool { return f.family.logical }

// FontName returns the face name, as Java's Font.getFontName(): for example
// "SansSerif.bold" or "DejaVu Sans Bold Oblique".
func (f AwtFont) FontName() string {
	if f.family.logical {
		return f.family.name + "." + f.style.String()
	}
	name := f.family.name
	if f.style.IsBold() {
		name += " Bold"
	}
	if f.style.IsItalic() {
		name += " " + f.family.italicName
	}
	return name
}

// PSName returns the PostScript name, as Java's Font.getPSName(): for
// example "SansSerif.bold" or "DejaVuSans-BoldOblique".
func (f AwtFont) PSName() string {
	if f.family.logical {
		return f.FontName()
	}
	var suffix string
	if f.style.IsBold() {
		suffix = "Bold"
	}
	if f.style.IsItalic() {
		suffix += f.family.italicName
	}
	if suffix == "" {
		return f.family.psName
	}
	return f.family.psName + "-" + suffix
}

// Style returns the style.
func (f AwtFont) Style() AwtStyle { return f.style }

// IsBold reports whether the font is bold.
func (f AwtFont) IsBold() bool { return f.style.IsBold() }

// IsItalic reports whether the font is italic.
func (f AwtFont) IsItalic() bool { return f.style.IsItalic() }

// Size returns the point size.
func (f AwtFont) Size() int { return f.size }

// Metrics returns the table that measures the font.
func (f AwtFont) Metrics() *Metrics { return f.metrics }

// CanDisplay reports whether the font can display cp, as Java's
// Font.canDisplay.
func (f AwtFont) CanDisplay(cp rune) bool { return f.metrics.CanDisplay(cp) }

// CanDisplayUpTo returns the byte index of the first code point of text the
// font cannot display, or -1 when it can display all of them, as Java's
// Font.canDisplayUpTo (which returns a UTF-16 index).
func (f AwtFont) CanDisplayUpTo(text string) int {
	for i, cp := range text {
		if !f.metrics.CanDisplay(cp) {
			return i
		}
	}
	return -1
}

// scale converts a value measured at the table's reference size to the
// font's size. The division comes last so that the result is the exact
// value whenever it is representable, as Java's scaled metrics are.
func (f AwtFont) scale(v float32) float64 {
	return float64(v) * float64(f.size) / float64(f.metrics.RefSize)
}

// Advance returns the advance of cp at the font's size.
func (f AwtFont) Advance(cp rune) float64 {
	adv, _ := f.metrics.Advance(cp)
	return float64(float32(f.scale(adv)))
}

// StringBounds returns the width and height of
// FontMetrics.getStringBounds(text, g) on a Graphics2D with antialiasing and
// fractional metrics. Java returns a Rectangle2D.Float: the width is the sum
// of the advances of the code points, accumulated in float32 from left to
// right without kerning, and the height is ascent + descent + leading, which
// does not depend on the text.
func (f AwtFont) StringBounds(text string) Dimension {
	var width float32
	for _, cp := range text {
		adv, _ := f.metrics.Advance(cp)
		width += float32(f.scale(adv))
	}
	return Dimension{Width: float64(width), Height: float64(float32(f.scale(f.metrics.BoundsHeight)))}
}

// Ascent returns Java's LineMetrics.getAscent() at the font's size.
func (f AwtFont) Ascent() float64 { return float64(float32(f.scale(f.metrics.Ascent))) }

// Descent returns Java's LineMetrics.getDescent() at the font's size. Java's
// LineMetrics do not depend on the text.
func (f AwtFont) Descent() float64 { return float64(float32(f.scale(f.metrics.Descent))) }

// Leading returns Java's LineMetrics.getLeading() at the font's size.
func (f AwtFont) Leading() float64 { return float64(float32(f.scale(f.metrics.Leading))) }
