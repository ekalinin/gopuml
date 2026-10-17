// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/FontStack.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// The family names of Java's logical fonts used by upstream.
const (
	FamilyMonospace = "Monospaced"
	FamilySansSerif = "SansSerif"
	FamilySerif     = "Serif"
)

// FontStack is a font family definition, a comma-separated list of families
// such as "Arial, Helvetica". It is a comparable value: two stacks are equal
// when their definitions are, as in upstream.
//
// Upstream caches the stacks and their decoded fonts; decoding is cheap here,
// with the embedded metrics, so the fonts are decoded on use.
type FontStack struct {
	fullDefinition string
}

// BuildFontStack returns the stack of a family definition.
func BuildFontStack(fullDefinition string) FontStack {
	return FontStack{fullDefinition: fullDefinition}
}

// Fonts returns the fonts of the stack in order: each family, trimmed of
// whitespace and double quotes, decoded as Java's Font.decode does.
//
// Upstream splits with Java's String.split, which drops trailing empty
// families. When no family is left (a definition made of commas only),
// upstream fails with an index error; the stack then holds the font of an
// empty name, which is Dialog.
func (s FontStack) Fonts() []AwtFont {
	names := strings.Split(s.fullDefinition, ",")
	for len(names) > 1 && names[len(names)-1] == "" {
		names = names[:len(names)-1]
	}
	fonts := make([]AwtFont, 0, len(names))
	for _, name := range names {
		fonts = append(fonts, DecodeAwtFont(trimWhitespaceOrDoubleQuote(name)))
	}
	return fonts
}

// CanDisplayUpTo returns the byte index of the first code point of text the
// font at index cannot display, or -1.
func (s FontStack) CanDisplayUpTo(index int, text string) int {
	return s.Fonts()[index].CanDisplayUpTo(text)
}

// Font returns the font Java uses to draw text: the first font of the stack
// that can display the whole text, or else the first font, resized and with
// the face applied.
//
// Upstream skips the search when the text is null; an empty text selects
// the first font as well.
func (s FontStack) Font(text string, face UFontFace, size int) AwtFont {
	fonts := s.Fonts()
	base := fonts[0]
	if len(fonts) > 1 {
		for _, f := range fonts {
			if f.CanDisplayUpTo(text) == -1 {
				base = f
				break
			}
		}
	}
	return face.DeriveFont(base.Derive(AwtPlain, size))
}

func isWhitespaceOrDoubleQuote(r rune) bool {
	return text.IsJavaWhitespace(r) || r == '"'
}

func trimWhitespaceOrDoubleQuote(s string) string {
	return strings.TrimFunc(s, isWhitespaceOrDoubleQuote)
}

// String returns upstream's representation, "FontStack[<definition>]".
func (s FontStack) String() string { return "FontStack[" + s.fullDefinition + "]" }

// FullDefinition returns the family definition.
func (s FontStack) FullDefinition() string { return s.fullDefinition }

// SvgFamily returns the font-family for SVG output: the CSS generic family
// for the logical fonts, else the definition with double quotes replaced by
// single quotes.
func (s FontStack) SvgFamily() string {
	switch s.fullDefinition {
	case FamilySerif:
		return "serif"
	case FamilySansSerif:
		return "sans-serif"
	case FamilyMonospace:
		return "monospace"
	}
	return strings.ReplaceAll(s.fullDefinition, `"`, "'")
}
