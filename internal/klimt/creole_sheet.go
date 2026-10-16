// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Sheet.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/SheetBuilder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Stripe.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/StripeStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/StripeStyleType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/CreoleContext.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Position.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Neutron.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/NeutronType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/atom/Atom.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/atom/AbstractAtom.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// SheetBuilder builds the sheet (the parsed creole) of a display. It is
// upstream's klimt/creole/SheetBuilder.
type SheetBuilder interface {
	CreateSheet(display *Display) *Sheet
}

// Sheet is the parsed creole of a display: a list of stripes (lines) and
// the horizontal alignment of the text.
type Sheet struct {
	stripes             []Stripe
	horizontalAlignment geom.HorizontalAlignment
}

// NewSheet returns an empty sheet; horizontalAlignment may be
// HorizontalAlignmentNull (null upstream).
func NewSheet(horizontalAlignment geom.HorizontalAlignment) *Sheet {
	return &Sheet{horizontalAlignment: horizontalAlignment}
}

// String is upstream's toString: the list of stripes.
func (s *Sheet) String() string {
	parts := make([]string, len(s.stripes))
	for i, st := range s.stripes {
		parts[i] = stripeString(st)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Add appends stripes.
func (s *Sheet) Add(stripes ...Stripe) { s.stripes = append(s.stripes, stripes...) }

// Stripes returns the stripes; the slice must not be modified.
func (s *Sheet) Stripes() []Stripe { return s.stripes }

// LastStripe returns the last stripe, nil for none.
func (s *Sheet) LastStripe() Stripe {
	if len(s.stripes) == 0 {
		return nil
	}
	return s.stripes[len(s.stripes)-1]
}

// HorizontalAlignment returns the alignment of the text,
// HorizontalAlignmentNull for none.
func (s *Sheet) HorizontalAlignment() geom.HorizontalAlignment { return s.horizontalAlignment }

// Stripe is a line of a sheet: atoms, the first one possibly a header
// (a bullet or a list number) that line breaks repeat as a blank.
type Stripe interface {
	// LHeader returns the header, nil for none.
	LHeader() Atom
	Atoms() []Atom
}

func stripeString(s Stripe) string {
	if st, ok := s.(interface{ String() string }); ok {
		return st.String()
	}
	return "Stripe"
}

// StripeStyleType is the kind of a creole line.
type StripeStyleType int

// The stripe style types, in upstream's order.
const (
	StripeStyleTypeNormal StripeStyleType = iota
	StripeStyleTypeHeading
	StripeStyleTypeListWithoutNumber
	StripeStyleTypeListWithNumber
	StripeStyleTypeHorizontalLine
	StripeStyleTypeTree
)

// String returns upstream's constant name.
func (t StripeStyleType) String() string {
	return enumName("StripeStyleType", []string{"NORMAL", "HEADING", "LIST_WITHOUT_NUMBER", "LIST_WITH_NUMBER",
		"HORIZONTAL_LINE", "TREE"}, int(t))
}

// StripeStyle is the style of a creole line: its kind, its level (order)
// and the character of a horizontal line.
type StripeStyle struct {
	typ   StripeStyleType
	order int
	style rune
}

// NewStripeStyle returns a stripe style.
func NewStripeStyle(typ StripeStyleType, order int, style rune) StripeStyle {
	return StripeStyle{typ: typ, order: order, style: style}
}

// Type returns the kind of line.
func (s StripeStyle) Type() StripeStyleType { return s.typ }

// Order returns the level of the line.
func (s StripeStyle) Order() int { return s.order }

// Style returns the character of a horizontal line.
func (s StripeStyle) Style() rune { return s.style }

// Header returns the bullet or list number of a list line, nil for other
// lines.
func (s StripeStyle) Header(fontConfiguration *FontConfiguration, context *CreoleContext) Atom {
	if s.typ == StripeStyleTypeListWithoutNumber {
		return NewBullet(fontConfiguration, s.order)
	}
	if s.typ == StripeStyleTypeListWithNumber {
		localNumber := context.LocalNumber(s.order)
		return AtomTextUtilsCreateListNumber(fontConfiguration, s.order, localNumber)
	}
	return nil
}

// CreoleContext numbers the items of the numbered lists of a text.
type CreoleContext struct {
	stack []int
}

// NewCreoleContext returns a context without numbered items.
func NewCreoleContext() *CreoleContext { return &CreoleContext{} }

// LocalNumber returns the number (from 0) of the next item of level order,
// restarting the deeper levels.
func (c *CreoleContext) LocalNumber(order int) int {
	c.ensureStackOk(order)
	n := c.stack[order]
	c.stack[order] = n + 1
	return n
}

func (c *CreoleContext) ensureStackOk(order int) {
	for len(c.stack) <= order {
		c.stack = append(c.stack, 0)
	}
	if len(c.stack) > order+1 {
		c.stack = c.stack[:order+1]
	}
}

// Position is the place of an atom: its top-left corner and size.
type Position struct {
	x, y float64
	dim  geom.XDimension2D
}

// NewPosition returns a position.
func NewPosition(x, y float64, dim geom.XDimension2D) Position {
	return Position{x: x, y: y, dim: dim}
}

// String is upstream's toString.
func (p Position) String() string {
	return "x=" + jstr(p.x) + " y=" + jstr(p.y) + " dim=" + p.dim.String()
}

// Align moves the position down so that its bottom is at height.
func (p Position) Align(height float64) Position { return p.TranslateY(height - p.dim.Height()) }

// MinY returns the top.
func (p Position) MinY() float64 { return p.y }

// MaxY returns the bottom.
func (p Position) MaxY() float64 { return p.y + p.Height() }

// Translate moves ug to the position.
func (p Position) Translate(ug UGraphic) UGraphic { return ug.Apply(NewUTranslate(p.x, p.y)) }

// TranslateY returns the position moved vertically.
func (p Position) TranslateY(dy float64) Position { return Position{x: p.x, y: p.y + dy, dim: p.dim} }

// TranslateX returns the position moved horizontally.
func (p Position) TranslateX(dx float64) Position { return Position{x: p.x + dx, y: p.y, dim: p.dim} }

// Update adds the bottom-right corner to minMax.
func (p Position) Update(minMax geom.MinMax) geom.MinMax {
	return minMax.AddPointXY(p.x+p.dim.Width(), p.y+p.dim.Height())
}

// DrawDebug draws the box of the position.
func (p Position) DrawDebug(ug UGraphic) {
	ug = ug.Apply(color.Black).Apply(Bg(color.None()))
	ug = ug.Apply(NewUTranslate(p.x, p.y))
	ug.Draw(URectangleBuildDim(p.dim))
}

// Height returns the height.
func (p Position) Height() float64 { return p.dim.Height() }

// Width returns the width.
func (p Position) Width() float64 { return p.dim.Width() }

// UTranslate returns the translation to the position.
func (p Position) UTranslate() UTranslate { return NewUTranslate(p.x, p.y) }

// NeutronType is the line-break class of a neutron.
type NeutronType int

// The neutron types, in upstream's order.
const (
	NeutronTypeUnbreakable NeutronType = iota
	NeutronTypeWhitespace
	NeutronTypeCjkIdeograph
	NeutronTypeZwspSeparator
	NeutronTypeUnknown
)

// String returns upstream's constant name.
func (t NeutronType) String() string {
	return enumName("NeutronType", []string{"UNBREAKABLE", "WHITESPACE", "CJK_IDEOGRAPH", "ZWSP_SEPARATOR",
		"UNKNOWN"}, int(t))
}

// Neutron is a piece of a stripe for line breaking: a word, a run of
// spaces, a CJK ideograph, another atom, or a break opportunity (ZWSP).
type Neutron struct {
	data    string
	hasData bool
	typ     NeutronType
	asAtom  Atom
}

// neutronZWSP is upstream's shared ZWSP neutron, a read-only value.
var neutronZWSP = &Neutron{typ: NeutronTypeZwspSeparator}

// NeutronCreate returns the neutron of an atom: a text atom takes the type
// of its first character.
func NeutronCreate(atom Atom) *Neutron {
	if at, ok := atom.(*AtomText); ok {
		text := at.Text()
		return &Neutron{data: text, hasData: true, typ: NeutronTypeFromChar(javaCharAt(text, 0)), asAtom: atom}
	}
	return &Neutron{typ: NeutronTypeUnknown, asAtom: atom}
}

// NeutronZwspSeparator returns the break opportunity.
func NeutronZwspSeparator() *Neutron { return neutronZWSP }

// String is upstream's toString.
func (n *Neutron) String() string {
	if n.typ == NeutronTypeZwspSeparator {
		return "ZWSP"
	}
	data := "null"
	if n.hasData {
		data = n.data
	}
	return n.typ.String() + "(" + data + ")"
}

// AsAtom returns the atom, nil for the ZWSP.
func (n *Neutron) AsAtom() Atom { return n.asAtom }

// Type returns the type.
func (n *Neutron) Type() NeutronType { return n.typ }

// isCjkOrJapanese reports whether ch belongs to a script commonly used in
// Chinese or Japanese writing systems, such as Kanji, Hiragana, Katakana, or
// Japanese punctuation. This is used to determine valid line-break
// positions.
func isCjkOrJapanese(ch rune) bool {
	return ch >= 0x4E00 && ch <= 0x9FFF || // CJK Unified Ideographs
		ch >= 0x3040 && ch <= 0x309F || // Hiragana
		ch >= 0x30A0 && ch <= 0x30FF || // Katakana
		ch >= 0x3000 && ch <= 0x303F || // CJK Symbols and Punctuation
		ch >= 0xFF65 && ch <= 0xFF9F // Halfwidth Katakana
}

// Width returns the width of the neutron, 0 for the ZWSP.
func (n *Neutron) Width(stringBounder font.StringBounder) float64 {
	if n.typ == NeutronTypeZwspSeparator {
		return 0
	}
	return n.asAtom.CalculateDimension(stringBounder).Width()
}

// NeutronTypeFromChar returns the type of a UTF-16 unit: whitespace, CJK
// ideograph or unbreakable.
func NeutronTypeFromChar(ch rune) NeutronType {
	if javaIsWhitespace(ch) {
		return NeutronTypeWhitespace
	}
	if isCjkOrJapanese(ch) {
		return NeutronTypeCjkIdeograph
	}
	return NeutronTypeUnbreakable
}

// javaIsWhitespace is Java's Character.isWhitespace for a UTF-16 unit: the
// space, line and paragraph separators but the no-break spaces, and the
// controls \t, \n, \u000B, \f, \r and \u001C to \u001F.
func javaIsWhitespace(ch rune) bool {
	switch {
	case ch >= '\t' && ch <= '\r', ch >= 0x1C && ch <= 0x1F:
		return true
	case ch == 0xA0 || ch == 0x2007 || ch == 0x202F:
		return false
	case ch == ' ', ch == 0x1680, ch >= 0x2000 && ch <= 0x200A, ch == 0x2028, ch == 0x2029, ch == 0x205F,
		ch == 0x3000:
		return true
	}
	return false
}

// Atom is a part of a creole line: a text block with a baseline offset
// (the starting altitude) that splits into neutrons for line breaking.
type Atom interface {
	TextBlock
	StartingAltitude(stringBounder font.StringBounder) float64
	Neutrons() []*Neutron
}

// atomNeutrons is upstream's AbstractAtom.getNeutrons: the atom as one
// neutron.
func atomNeutrons(a Atom) []*Neutron { return []*Neutron{NeutronCreate(a)} }

// AtomBase is upstream's AbstractAtom (a TextBlockMemoized): embed it in
// atoms for the default TextBlock methods and the dimension cache.
type AtomBase struct {
	TextBlockDefaults
	TextBlockMemoized
}

// utf16Of returns the UTF-16 units of s.
func utf16Of(s string) []uint16 { return utf16.Encode([]rune(s)) }

// stringOf returns the string of UTF-16 units.
func stringOf(u []uint16) string { return string(utf16.Decode(u)) }
