// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HColors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// Colors used by the diagrams.
var (
	Black     = mustParse("#000000")
	White     = mustParse("#FFFFFF")
	RedLight  = mustParse("#FEF6F3")
	RedDark   = mustParse("#CD0A0A")
	Red       = mustParse("#FF0000")
	Green     = mustParse("#00FF00")
	Blue      = mustParse("#0000FF")
	Gray      = mustParse("#808080")
	LightGray = mustParse("#C0C0C0")
	MyYellow  = mustParse("#FEFECE")
	MyRed     = mustParse("#A80036")
	MyGreen   = mustParse("#33FF02")

	ColC82930 = mustParse("#C82930")
	ColF24D5C = mustParse("#F24D5C")
	Col1963A0 = mustParse("#1963A0")
	Col4177AF = mustParse("#4177AF")
	ColB38D22 = mustParse("#B38D22")
	ColFFFF44 = mustParse("#FFFF44")
	Col038048 = mustParse("#038048")
	Col84BE84 = mustParse("#84BE84")
	ColDDDDDD = mustParse("#DDDDDD")
	ColEEEEEE = mustParse("#EEEEEE")
	ColFBFB77 = mustParse("#FBFB77")
	ColADD1B2 = mustParse("#ADD1B2")
	ColA9DCDF = mustParse("#A9DCDF")
	ColE3664A = mustParse("#E3664A")
	ColEB937F = mustParse("#EB937F")
	ColB4A7E5 = mustParse("#B4A7E5")
	Col527BC6 = mustParse("#527BC6")
	ColD1DBEF = mustParse("#D1DBEF")
	ColD7E0F2 = mustParse("#D7E0F2")
	Col989898 = mustParse("#989898")
	ColBBBBBB = mustParse("#BBBBBB")
)

// mustParse stands for upstream's getColorOrWhite on constants that always
// parse; going through GetColorOrWhite would make White depend on itself.
func mustParse(s string) HColor {
	c := hcolorSet.parseColor(s)
	if c == nil {
		panic("color: invalid constant " + s)
	}
	return c
}

var transparent = NewHColorSimple(NewXColorAlpha(0, 0, 0, 0))

// Transparent returns the transparent color. It is a single value, so
// comparing with == works.
func Transparent() HColor { return transparent }

// TransparentWith returns the transparent color with a fill behavior.
func TransparentWith(fillBehavior TransparentFillBehavior) HColor {
	return transparent.WithTransparentFillBehavior(fillBehavior)
}

// None returns the transparent color.
func None() HColor { return transparent }

// NoGradient returns the first color of a gradient, and any other color
// unchanged.
func NoGradient(c HColor) HColor {
	switch g := c.(type) {
	case *HColorGradient:
		return g.Color1()
	case *HColorLinearGradient:
		return g.Color1()
	}
	return c
}

// Unlinear moves from color1 to color2 by the cube of completion percent in
// HSL. It returns color1 at 0 and color2 at 100, and color1 when either is
// not a simple color.
func Unlinear(color1, color2 HColor, completion int) HColor {
	if completion == 0 {
		return color1
	}
	if completion == 100 {
		return color2
	}
	s1, ok1 := color1.(*HColorSimple)
	s2, ok2 := color2.(*HColorSimple)
	if ok1 && ok2 {
		return unlinearSimple(s1, s2, completion)
	}
	return color1
}

// Middle returns the RGB average of two colors.
func Middle(c1, c2 HColor) HColor { return newHColorMiddle(c1, c2) }

// Gradient returns a gradient from color1 to color2.
func Gradient(color1, color2 HColor, policy rune) *HColorGradient {
	return newHColorGradient(color1, color2, policy)
}

// Simple returns the simple color of c.
func Simple(c XColor) HColor { return NewHColorSimple(c) }
