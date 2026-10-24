// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/SvgCommand.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/SvgCommandLetter.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/SvgCommandNumber.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/SvgPosition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"
	"unicode"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// SvgCommand is a token of SVG path data: a command letter or a number.
// Upstream's package openiconic joins this one (design D2).
type SvgCommand interface {
	ToSvg() string
}

// SvgCommandLetter is a path command letter (M, l, Z, ...).
type SvgCommandLetter struct {
	letter rune
}

var _ SvgCommand = SvgCommandLetter{}

// NewSvgCommandLetter returns the command of a letter.
func NewSvgCommandLetter(letter rune) SvgCommandLetter { return SvgCommandLetter{letter: letter} }

// String is upstream's toString: a space and the letter.
func (c SvgCommandLetter) String() string { return " " + string(c.letter) }

// ToSvg returns the letter.
func (c SvgCommandLetter) ToSvg() string { return string(c.letter) }

// ArgumentNumber returns the number of arguments of the command. An unknown
// letter panics, as upstream throws UnsupportedOperationException.
func (c SvgCommandLetter) ArgumentNumber() int {
	switch javaCharLower(c.letter) {
	case 'm', 'l', 't':
		return 2
	case 'h', 'v':
		return 1
	case 'z':
		return 0
	case 'c':
		return 6
	case 'q', 's':
		return 4
	case 'a':
		return 7
	}
	throw(unsupportedOperation, string(c.letter))
	return 0
}

// IsUpperCase reports whether the letter is upper case (absolute
// coordinates).
func (c SvgCommandLetter) IsUpperCase() bool { return unicode.IsUpper(c.letter) }

// Is reports whether the command is the letter r.
func (c SvgCommandLetter) Is(r rune) bool { return c.letter == r }

// Letter returns the letter.
func (c SvgCommandLetter) Letter() rune { return c.letter }

// Implicit returns the command that repeats after the arguments of this one:
// a line after a move ("l" after "m", "L" after "M"), the command itself
// otherwise.
func (c SvgCommandLetter) Implicit() SvgCommandLetter {
	// https://stackoverflow.com/questions/29251389/svg-path-spec-moveto-and-implicit-lineto
	if c.letter == 'm' {
		return NewSvgCommandLetter('l')
	}
	if c.letter == 'M' {
		return NewSvgCommandLetter('L')
	}
	return c
}

// SvgCommandNumber is a number of path data.
type SvgCommandNumber struct {
	number float64
}

var _ SvgCommand = SvgCommandNumber{}

// SvgCommandNumberParse parses a number with Java's Double.parseDouble; it
// panics with a NumberFormatException when the text is not a number.
func SvgCommandNumberParse(number string) SvgCommandNumber {
	return NewSvgCommandNumber(svgParseDouble(number))
}

// NewSvgCommandNumber returns the number.
func NewSvgCommandNumber(number float64) SvgCommandNumber { return SvgCommandNumber{number: number} }

// Add returns the sum of the two numbers.
func (n SvgCommandNumber) Add(other SvgCommandNumber) SvgCommandNumber {
	return NewSvgCommandNumber(n.Double() + other.Double())
}

// ToSvg returns the number with at most 4 decimals, upstream's
// StringUtils.formatDecimal(number, 4).
func (n SvgCommandNumber) ToSvg() string { return stringUtilsFormatDecimal(n.number, 4) }

// Double returns the value.
func (n SvgCommandNumber) Double() float64 { return n.number }

// SvgPosition is a point of path data.
type SvgPosition struct {
	x, y SvgCommandNumber
}

// NewSvgPosition returns the point (x, y).
func NewSvgPosition(x, y float64) SvgPosition {
	return SvgPosition{x: NewSvgCommandNumber(x), y: NewSvgCommandNumber(y)}
}

// NewSvgPositionOf returns the point of two numbers.
func NewSvgPositionOf(x, y SvgCommandNumber) SvgPosition { return SvgPosition{x: x, y: y} }

// String is upstream's toString: "x,y" with ToSvg numbers.
func (p SvgPosition) String() string { return p.x.ToSvg() + "," + p.y.ToSvg() }

// X returns the x number.
func (p SvgPosition) X() SvgCommandNumber { return p.x }

// Y returns the y number.
func (p SvgPosition) Y() SvgCommandNumber { return p.y }

// XDouble returns x.
func (p SvgPosition) XDouble() float64 { return p.x.Double() }

// YDouble returns y.
func (p SvgPosition) YDouble() float64 { return p.y.Double() }

// Add returns the sum of the two points.
func (p SvgPosition) Add(other SvgPosition) SvgPosition {
	return NewSvgPositionOf(p.x.Add(other.x), p.y.Add(other.y))
}

// Mirror returns the reflection of tobeMirrored through the point.
func (p SvgPosition) Mirror(tobeMirrored SvgPosition) SvgPosition {
	centerX := p.XDouble()
	centerY := p.YDouble()
	// x1+x2 = 2*xc
	x := float64(2*centerX) - tobeMirrored.XDouble()
	y := float64(2*centerY) - tobeMirrored.YDouble()
	return NewSvgPosition(x, y)
}

// Affine returns the point transformed by at.
func (p SvgPosition) Affine(at *XAffineTransform) geom.XPoint2D {
	return at.Transform(geom.NewXPoint2D(p.XDouble(), p.YDouble()))
}

// svgParseDouble is Java's Double.parseDouble: an empty text panics with
// Java's "empty String" message, any other text that is not a number with
// "For input string".
func svgParseDouble(s string) float64 {
	if javaTrim(s) == "" {
		throw(numberFormat, "empty String")
	}
	return javaParseDouble(s)
}

// javaTrim is Java's String.trim: the characters up to the space removed at
// both ends.
func javaTrim(s string) string { return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' }) }

// stringUtilsFormatDecimal is upstream's StringUtils.formatDecimal: "0" for
// zero, else String.format(Locale.US, "%.<decimal>f", x) without trailing
// zeros (the fast path of upstream gives the same digits).
func stringUtilsFormatDecimal(x float64, decimal int) string {
	if x == 0 {
		return "0"
	}
	s := jfmt.FormatFixed(x, decimal)
	if strings.IndexByte(s, '.') >= 0 {
		s = strings.TrimSuffix(strings.TrimRight(s, "0"), ".")
	}
	if s == "-0" {
		return "0"
	}
	return s
}
