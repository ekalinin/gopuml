// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/UFont.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/UFontContext.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/UFontFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import (
	"math"
	"strconv"
	"strings"
)

// UFontContext is the output a font family name is written for.
type UFontContext int

// The font contexts.
const (
	ContextEPS UFontContext = iota
	ContextSVG
	ContextG2D
	ContextTIKZ
	ContextPDF
)

// String returns upstream's enum name.
func (c UFontContext) String() string {
	switch c {
	case ContextEPS:
		return "EPS"
	case ContextSVG:
		return "SVG"
	case ContextG2D:
		return "G2D"
	case ContextTIKZ:
		return "TIKZ"
	case ContextPDF:
		return "PDF"
	}
	return "UFontContext(" + strconv.Itoa(int(c)) + ")"
}

// UFont is a font: a family stack, a face (style and CSS weight) and an
// integer size. It is a comparable value, equal to another font with the
// same stack, face and size, as in upstream.
type UFont struct {
	fontStack FontStack
	face      UFontFace
	size      int
}

// newUFont returns a font; the zero face means the normal face.
func newUFont(fontStack FontStack, face UFontFace, size int) UFont {
	if face.IsZero() {
		face = NormalFace()
	}
	return UFont{fontStack: fontStack, face: face, size: size}
}

// UnderlyingFont returns the font Java uses to draw text: the first family
// of the stack that can display text, with the face and size applied.
func (f UFont) UnderlyingFont(text string) AwtFont {
	return f.fontStack.Font(text, f.face, f.size)
}

// WithSize returns the font with another size, truncated to an integer as
// Java's (int) cast does.
func (f UFont) WithSize(size float32) UFont {
	return newUFont(f.fontStack, f.face, javaFloatToInt(size))
}

// javaFloatToInt converts like Java's (int) cast: toward zero, NaN to 0,
// and saturated to the int range.
func javaFloatToInt(v float32) int {
	switch {
	case math.IsNaN(float64(v)):
		return 0
	case v >= math.MaxInt32:
		return math.MaxInt32
	case v <= math.MinInt32:
		return math.MinInt32
	}
	return int(v)
}

// FontFace returns the face.
func (f UFont) FontFace() UFontFace { return f.face }

// WithFontFace returns the font with another face. The zero face, upstream's
// null, leaves the font unchanged.
func (f UFont) WithFontFace(face UFontFace) UFont {
	if face.IsZero() {
		return f
	}
	return newUFont(f.fontStack, face, f.size)
}

// FontStack returns the family stack.
func (f UFont) FontStack() FontStack { return f.fontStack }

// Size returns the size.
func (f UFont) Size() int { return f.size }

// Size2D returns the size as a float.
func (f UFont) Size2D() float64 { return float64(f.size) }

// Family returns the font family to write for an output: the family
// definition for PDF, the PostScript name of the font Java uses for EPS, and
// the SVG family for SVG. Upstream throws for the other contexts; this
// panics.
func (f UFont) Family(text string, context UFontContext) string {
	switch context {
	case ContextPDF:
		return f.fontStack.FullDefinition()
	case ContextEPS:
		return f.UnderlyingFont(text).PSName()
	case ContextSVG:
		return f.fontStack.SvgFamily()
	}
	panic("font: no family for context " + context.String())
}

// portableFontName returns the face name of the font Java uses, ending with
// its style as on Linux: "SansSerif.bold", "DejaVu Sans Bold.bold".
func (f UFont) portableFontName() string {
	font := f.UnderlyingFont("")
	name := font.FontName()
	suffix := "." + font.Style().String()
	if strings.HasSuffix(name, suffix) {
		return name
	}
	return name + suffix
}

// String returns upstream's representation: the face name of the font Java
// uses, a slash and the size, for example "SansSerif.plain/12".
func (f UFont) String() string {
	return f.portableFontName() + "/" + strconv.Itoa(f.size)
}

// Build returns the font of a family definition, a face and a size. The
// zero face means the normal face.
func Build(fullDefinition string, face UFontFace, fontSize int) UFont {
	return newUFont(BuildFontStack(fullDefinition), face, fontSize)
}

// Serif returns the normal Serif font of a size.
func Serif(size int) UFont { return Build(FamilySerif, NormalFace(), size) }

// SansSerif returns the normal SansSerif font of a size.
func SansSerif(size int) UFont { return Build(FamilySansSerif, NormalFace(), size) }

// Monospace returns the normal Monospaced font of a size.
func Monospace(size int) UFont { return Build(FamilyMonospace, NormalFace(), size) }

// ByDefault returns the default font. As in upstream, it is SansSerif 12
// whatever the size.
func ByDefault(size int) UFont { return SansSerif(12) }
