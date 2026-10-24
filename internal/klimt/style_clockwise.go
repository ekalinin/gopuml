// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/ClockwiseTopRightBottomLeft.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ClockwiseTopRightBottomLeft is a padding or margin: top, right, bottom and
// left sizes, as CSS. It is an immutable, comparable value.
type ClockwiseTopRightBottomLeft struct {
	top, right, bottom, left float64
}

// ClockwiseSame returns the same size on every side.
func ClockwiseSame(value float64) ClockwiseTopRightBottomLeft {
	return ClockwiseTopRightBottomLeft{top: value, right: value, bottom: value, left: value}
}

// ClockwiseNone returns zero on every side.
func ClockwiseNone() ClockwiseTopRightBottomLeft { return ClockwiseTopRightBottomLeft{} }

var numbersOnly = regexp.MustCompile(`^[0-9 ]+$`)

// ClockwiseRead parses one to four integers separated by spaces, as CSS
// ("5", "5 10", "5 10 15", "5 10 15 20"). Anything else gives zero sizes.
func ClockwiseRead(value string) ClockwiseTopRightBottomLeft {
	if !numbersOnly.MatchString(value) {
		return ClockwiseNone()
	}
	split := strings.Fields(strings.TrimSpace(value))
	parse := func(s string) (float64, bool) {
		v, ok := parseIntOK(s)
		return float64(v), ok
	}
	values := make([]float64, len(split))
	for i, s := range split {
		v, ok := parse(s)
		if !ok {
			return ClockwiseNone()
		}
		values[i] = v
	}
	switch len(values) {
	case 1:
		v := values[0]
		return ClockwiseTopRightBottomLeft{v, v, v, v}
	case 2:
		a, b := values[0], values[1]
		return ClockwiseTopRightBottomLeft{a, b, a, b}
	case 3:
		a, b, c := values[0], values[1], values[2]
		return ClockwiseTopRightBottomLeft{a, b, c, b}
	case 4:
		return ClockwiseTopRightBottomLeft{values[0], values[1], values[2], values[3]}
	}
	return ClockwiseNone()
}

// parseIntOK is Java's Integer.parseInt reporting a NumberFormatException as
// false.
func parseIntOK(s string) (v int, ok bool) {
	defer func() {
		if recover() != nil {
			v, ok = 0, false
		}
	}()
	return javaParseInt(s), true
}

// ClockwiseMargin1Margin2 returns margin1 at the top and bottom and margin2
// at the right and left.
func ClockwiseMargin1Margin2(margin1, margin2 float64) ClockwiseTopRightBottomLeft {
	return ClockwiseTopRightBottomLeftOf(margin1, margin2, margin1, margin2)
}

// ClockwiseTopRightBottomLeftOf returns the four sizes.
func ClockwiseTopRightBottomLeftOf(top, right, bottom, left float64) ClockwiseTopRightBottomLeft {
	return ClockwiseTopRightBottomLeft{top: top, right: right, bottom: bottom, left: left}
}

// IncTop returns the sizes with the top increased by delta.
func (c ClockwiseTopRightBottomLeft) IncTop(delta float64) ClockwiseTopRightBottomLeft {
	return ClockwiseTopRightBottomLeft{top: c.top + delta, right: c.right, bottom: c.bottom, left: c.left}
}

// String is upstream's toString, for example "1.0:2.0:3.0:4.0".
func (c ClockwiseTopRightBottomLeft) String() string {
	return jstr(c.top) + ":" + jstr(c.right) + ":" + jstr(c.bottom) + ":" + jstr(c.left)
}

// Top returns the top size.
func (c ClockwiseTopRightBottomLeft) Top() float64 { return c.top }

// Right returns the right size.
func (c ClockwiseTopRightBottomLeft) Right() float64 { return c.right }

// Bottom returns the bottom size.
func (c ClockwiseTopRightBottomLeft) Bottom() float64 { return c.bottom }

// Left returns the left size.
func (c ClockwiseTopRightBottomLeft) Left() float64 { return c.left }

// ClockwiseMarginForDocument returns the margin of the document style.
func ClockwiseMarginForDocument(styleBuilder *StyleBuilder) ClockwiseTopRightBottomLeft {
	style := StyleSignatureBasicOf(SNameRoot, SNameDocument).MergedStyle(styleBuilder)
	if style == nil {
		throw(nullPointer, "")
	}
	return style.Margin()
}

// Translate returns the translation by (left, top).
func (c ClockwiseTopRightBottomLeft) Translate() UTranslate { return NewUTranslate(c.left, c.top) }

// Apply returns d grown by the sizes.
func (c ClockwiseTopRightBottomLeft) Apply(d geom.XDimension2D) geom.XDimension2D {
	return geom.NewXDimension2D(c.left+d.Width()+c.right, c.top+d.Height()+c.bottom)
}

// IsZero reports whether every size is zero.
func (c ClockwiseTopRightBottomLeft) IsZero() bool {
	return c.left == 0 && c.right == 0 && c.top == 0 && c.bottom == 0
}
