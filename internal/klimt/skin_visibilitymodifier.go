// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/VisibilityModifier.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/Padder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// VisibilityModifier is the visibility of a class member, drawn as an icon.
type VisibilityModifier int

// Visibilities, in upstream's order.
const (
	VisibilityPrivateField VisibilityModifier = iota
	VisibilityProtectedField
	VisibilityPackagePrivateField
	VisibilityPublicField
	VisibilityPrivateMethod
	VisibilityProtectedMethod
	VisibilityPackagePrivateMethod
	VisibilityPublicMethod
	VisibilityIEMandatory
	visibilityCount
)

var visibilityData = [...]struct {
	name       string
	unicode    rune
	foreground ColorParam
	background ColorParam
	hasBack    bool
}{
	{"PRIVATE_FIELD", charPrivateField, ColorParamIconPrivate, 0, false},
	{"PROTECTED_FIELD", charProtectedField, ColorParamIconProtected, 0, false},
	{"PACKAGE_PRIVATE_FIELD", charPackagePrivateField, ColorParamIconPackage, 0, false},
	{"PUBLIC_FIELD", charPublicField, ColorParamIconPublic, 0, false},
	{"PRIVATE_METHOD", charPrivateMethod, ColorParamIconPrivate, ColorParamIconPrivateBackground, true},
	{"PROTECTED_METHOD", charProtectedMethod, ColorParamIconProtected, ColorParamIconProtectedBackground, true},
	{"PACKAGE_PRIVATE_METHOD", charPackagePrivateMethod, ColorParamIconPackage, ColorParamIconPackageBackground,
		true},
	{"PUBLIC_METHOD", charPublicMethod, ColorParamIconPublic, ColorParamIconPublicBackground, true},
	{"IE_MANDATORY", charIEMandatory, ColorParamIconIEMandatory, ColorParamIconIEMandatory, true},
}

// String returns upstream's constant name.
func (v VisibilityModifier) String() string { return visibilityData[v].name }

// VisibilityRegexForVisibilityCharacter returns the pattern of the
// visibility characters.
func VisibilityRegexForVisibilityCharacter() string { return "[-#+~]" }

// UDrawable returns the icon as a drawable.
func (v VisibilityModifier) UDrawable(size int, foregroundColor, backgoundColor color.HColor) UDrawable {
	return udrawableFunc(func(ug UGraphic) { v.drawWithGroup(ug, size, foregroundColor, backgoundColor, 0, 0) })
}

type udrawableFunc func(ug UGraphic)

func (f udrawableFunc) DrawU(ug UGraphic) { f(ug) }

// visibilityBlock is the anonymous block of upstream's getUBlock.
type visibilityBlock struct {
	TextBlockDefaults
	v                      VisibilityModifier
	size                   int
	foreground, background color.HColor
	withInvisibleRectangle bool
}

func (b *visibilityBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(float64(b.size+1), float64(b.size+1))
}

func (b *visibilityBlock) InnerPosition(string, font.StringBounder) *geom.XRectangle2D { return nil }

func (b *visibilityBlock) DrawU(ug UGraphic) {
	if b.withInvisibleRectangle {
		ug.Apply(color.None()).Draw(URectangleBuild(float64(b.size*2), float64(b.size)))
	}
	b.v.drawWithGroup(ug, b.size, b.foreground, b.background, 0, 0)
}

// UBlock returns the icon as a block of size+1, with an invisible
// rectangle of twice the size when withInvisibleRectangle is set.
func (v VisibilityModifier) UBlock(size int, foregroundColor, backgoundColor color.HColor,
	withInvisibleRectangle bool) TextBlock {
	return &visibilityBlock{v: v, size: size, foreground: foregroundColor, background: backgoundColor,
		withInvisibleRectangle: withInvisibleRectangle}
}

func (v VisibilityModifier) drawWithGroup(ug UGraphic, size int, foregroundColor, backgoundColor color.HColor,
	x, y float64) {
	group := NewUGroup()
	group.Put(UGroupDataVisibilityModifier, v.String())
	ug.StartGroup(group)
	v.drawInternal(ug, size, foregroundColor, backgoundColor, x, y)
	ug.CloseGroup()
}

func (v VisibilityModifier) drawInternal(ug UGraphic, size int, foregroundColor, backgoundColor color.HColor,
	x, y float64) {
	if backgoundColor == nil {
		ug = ug.Apply(Bg(color.None()))
	} else {
		ug = ug.Apply(Bg(backgoundColor))
	}
	ug = ug.Apply(foregroundColor)
	size = ensureEven(size)
	switch v {
	case VisibilityPackagePrivateField, VisibilityPackagePrivateMethod:
		drawVisibilityTriangle(ug, size, x, y)
	case VisibilityPrivateField, VisibilityPrivateMethod:
		drawVisibilitySquare(ug, size, x, y)
	case VisibilityProtectedField, VisibilityProtectedMethod:
		drawVisibilityDiamond(ug, size, x, y)
	case VisibilityPublicField, VisibilityPublicMethod, VisibilityIEMandatory:
		drawVisibilityCircle(ug, size, x, y)
	default:
		throw(illegalState, "")
	}
}

func drawVisibilitySquare(ug UGraphic, size int, x, y float64) {
	ug.Apply(NewUTranslate(x+2, y+2)).Draw(URectangleBuild(float64(size-4), float64(size-4)))
}

func drawVisibilityCircle(ug UGraphic, size int, x, y float64) {
	ug.Apply(NewUTranslate(x+2, y+2)).Draw(UEllipseBuild(float64(size-4), float64(size-4)))
}

func ensureEven(n int) int {
	if n%2 == 1 {
		n--
	}
	return n
}

func drawVisibilityDiamond(ug UGraphic, size int, x, y float64) {
	poly := NewUPolygon()
	size -= 2
	s := float64(size)
	poly.AddPoint(s/2.0, 0)
	poly.AddPoint(s, s/2.0)
	poly.AddPoint(s/2.0, s)
	poly.AddPoint(0, s/2.0)
	ug.Apply(NewUTranslate(x+1, y)).Draw(poly)
}

func drawVisibilityTriangle(ug UGraphic, size int, x, y float64) {
	poly := NewUPolygon()
	size -= 2
	s := float64(size)
	poly.AddPoint(s/2.0, 1)
	poly.AddPoint(0, float64(size-1))
	poly.AddPoint(s, float64(size-1))
	ug.Apply(NewUTranslate(x+1, y)).Draw(poly)
}

// firstTwoChars returns the length and the first two UTF-16 units of s.
func firstTwoChars(s string) (n int, c0, c1 rune) {
	u := utf16.Encode([]rune(s))
	n = len(u)
	if n > 0 {
		c0 = rune(u[0])
	}
	if n > 1 {
		c1 = rune(u[1])
	}
	return
}

// VisibilityIsVisibilityCharacter reports whether s starts with a
// visibility character ('-', '#', '+', '~' or '*') not doubled, and is
// longer than 2.
func VisibilityIsVisibilityCharacter(s string) bool {
	n, c, c1 := firstTwoChars(s)
	if n <= 2 {
		return false
	}
	if c1 == c {
		return false
	}
	return c == '-' || c == '#' || c == '+' || c == '~' || c == '*'
}

// VisibilityByUnicode returns the visibility of a private-use character; ok
// is false when there is none.
func VisibilityByUnicode(c rune) (VisibilityModifier, bool) {
	for v := VisibilityModifier(0); v < visibilityCount; v++ {
		if visibilityData[v].unicode == c {
			return v, true
		}
	}
	return 0, false
}

// VisibilityReplaceVisibilityModifierByUnicodeChar replaces the visibility
// character of s by its private-use character.
func VisibilityReplaceVisibilityModifierByUnicodeChar(s string, isField bool) string {
	modifier, ok := VisibilityOf(s, isField)
	if !ok {
		return s
	}
	u := utf16.Encode([]rune(s))
	return string(visibilityData[modifier].unicode) + string(utf16.Decode(u[1:]))
}

// VisibilityOf returns the visibility given by the first character of s;
// ok is false when there is none.
func VisibilityOf(s string, isField bool) (VisibilityModifier, bool) {
	n, c, c1 := firstTwoChars(s)
	if n <= 2 {
		return 0, false
	}
	if c1 == c {
		return 0, false
	}
	if isField {
		return visibilityForField(c)
	}
	return visibilityForMethod(c)
}

func visibilityForField(c rune) (VisibilityModifier, bool) {
	switch c {
	case '-':
		return VisibilityPrivateField, true
	case '#':
		return VisibilityProtectedField, true
	case '+':
		return VisibilityPublicField, true
	case '~':
		return VisibilityPackagePrivateField, true
	case '*':
		return VisibilityIEMandatory, true
	}
	return 0, false
}

func visibilityForMethod(c rune) (VisibilityModifier, bool) {
	switch c {
	case '-':
		return VisibilityPrivateMethod, true
	case '#':
		return VisibilityProtectedMethod, true
	case '+':
		return VisibilityPublicMethod, true
	case '~':
		return VisibilityPackagePrivateMethod, true
	case '*':
		return VisibilityIEMandatory, true
	}
	return 0, false
}

// Foreground returns the color parameter of the icon border.
func (v VisibilityModifier) Foreground() ColorParam { return visibilityData[v].foreground }

// Background returns the color parameter of the icon fill; ok is false for
// fields (null upstream).
func (v VisibilityModifier) Background() (ColorParam, bool) {
	return visibilityData[v].background, visibilityData[v].hasBack
}

// XmiVisibility returns the XMI name of the visibility; IE_MANDATORY panics,
// as upstream throws IllegalStateException.
func (v VisibilityModifier) XmiVisibility() string {
	switch v {
	case VisibilityPublicField, VisibilityPublicMethod:
		return "public"
	case VisibilityPrivateField, VisibilityPrivateMethod:
		return "private"
	case VisibilityProtectedField, VisibilityProtectedMethod:
		return "protected"
	case VisibilityPackagePrivateField, VisibilityPackagePrivateMethod:
		return "package"
	}
	throw(illegalState, "")
	return ""
}

// IsField reports whether the visibility is a field's.
func (v VisibilityModifier) IsField() bool {
	return v == VisibilityPublicField || v == VisibilityPrivateField || v == VisibilityProtectedField ||
		v == VisibilityPackagePrivateField
}

// StyleSignature returns the signature of the icon style.
func (v VisibilityModifier) StyleSignature() StyleSignatureBasic {
	switch v {
	case VisibilityIEMandatory:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameVisibilityIcon, SNameIEMandatory)
	case VisibilityPublicField, VisibilityPublicMethod:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameVisibilityIcon, SNamePublic)
	case VisibilityPrivateField, VisibilityPrivateMethod:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameVisibilityIcon, SNamePrivate)
	case VisibilityProtectedField, VisibilityProtectedMethod:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameVisibilityIcon, SNameProtected)
	case VisibilityPackagePrivateField, VisibilityPackagePrivateMethod:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameVisibilityIcon, SNamePackage)
	}
	throw(illegalState, "")
	return StyleSignatureBasic{}
}

// Padder is the margin, padding, colors and corner put around sequence
// messages ("skinparam SequenceMessagePadding" and friends).
type Padder struct {
	margin          ClockwiseTopRightBottomLeft
	padding         ClockwiseTopRightBottomLeft
	backgroundColor color.HColor
	borderColor     color.HColor
	roundCorner     float64
}

var padderNone = &Padder{}

// PadderNone returns the padder that changes nothing (compared by
// identity).
func PadderNone() *Padder { return padderNone }

// String is upstream's toString.
func (p *Padder) String() string {
	return p.margin.String() + "/" + p.padding.String() + "/" + colorString(p.borderColor) + "/" +
		colorString(p.backgroundColor)
}

func (p *Padder) with(change func(c *Padder)) *Padder {
	c := *p
	change(&c)
	return &c
}

// WithMargin returns the padder with another margin.
func (p *Padder) WithMargin(margin ClockwiseTopRightBottomLeft) *Padder {
	return p.with(func(c *Padder) { c.margin = margin })
}

// WithPadding returns the padder with another padding.
func (p *Padder) WithPadding(padding ClockwiseTopRightBottomLeft) *Padder {
	return p.with(func(c *Padder) { c.padding = padding })
}

// WithBackgroundColor returns the padder with another background color.
func (p *Padder) WithBackgroundColor(backgroundColor color.HColor) *Padder {
	return p.with(func(c *Padder) { c.backgroundColor = backgroundColor })
}

// WithBorderColor returns the padder with another border color.
func (p *Padder) WithBorderColor(borderColor color.HColor) *Padder {
	return p.with(func(c *Padder) { c.borderColor = borderColor })
}

// WithRoundCorner returns the padder with another corner.
func (p *Padder) WithRoundCorner(roundCorner float64) *Padder {
	return p.with(func(c *Padder) { c.roundCorner = roundCorner })
}

// Margin returns the margin.
func (p *Padder) Margin() ClockwiseTopRightBottomLeft { return p.margin }

// Padding returns the padding.
func (p *Padder) Padding() ClockwiseTopRightBottomLeft { return p.padding }

// BackgroundColor returns the background color, nil for none.
func (p *Padder) BackgroundColor() color.HColor { return p.backgroundColor }

// BorderColor returns the border color, nil for none.
func (p *Padder) BorderColor() color.HColor { return p.borderColor }

// padderBlock is the anonymous block of upstream's Padder.apply.
type padderBlock struct {
	TextBlockDefaults
	p    *Padder
	orig TextBlock
}

func (b *padderBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.p.padding.Apply(b.p.margin.Apply(b.orig.CalculateDimension(stringBounder)))
}

func (b *padderBlock) DrawU(ug UGraphic) {
	ug = ug.Apply(b.p.margin.Translate())
	ug2 := ug
	if b.p.borderColor == nil {
		ug2 = ug2.Apply(color.None())
	} else {
		ug2 = ug2.Apply(b.p.borderColor)
	}
	if b.p.backgroundColor == nil {
		ug2 = ug2.Apply(Bg(color.None()))
	} else {
		ug2 = ug2.Apply(Bg(b.p.backgroundColor))
	}
	originalDim := b.orig.CalculateDimension(ug.StringBounder())
	rect := URectangleBuildDim(b.p.padding.Apply(originalDim)).Rounded(b.p.roundCorner)
	ug2.Draw(rect)
	b.orig.DrawU(ug.Apply(b.p.padding.Translate()))
}

// Apply returns orig with the margin, padding and box; orig itself for
// PadderNone.
func (p *Padder) Apply(orig TextBlock) TextBlock {
	if p == padderNone {
		return orig
	}
	return &padderBlock{p: p, orig: orig}
}
