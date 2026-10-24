// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/ColorOrder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "strings"

// ColorOrder is a permutation of the red, green and blue channels, used by
// "skinparam reversecolor".
type ColorOrder int

// Channel orders, in upstream's order.
const (
	OrderRGB ColorOrder = iota
	OrderRBG
	OrderGRB
	OrderGBR
	OrderBRG
	OrderBGR
)

var colorOrderNames = [...]string{"RGB", "RBG", "GRB", "GBR", "BRG", "BGR"}

// String returns upstream's constant name, for example "RGB".
func (o ColorOrder) String() string {
	if o < 0 || int(o) >= len(colorOrderNames) {
		return "ColorOrder(?)"
	}
	return colorOrderNames[o]
}

// Color returns an opaque color whose red, green and blue channels are the
// channels of c taken in the order o.
func (o ColorOrder) Color(c XColor) XColor {
	switch o {
	case OrderRGB:
		return NewXColor(c.Red(), c.Green(), c.Blue())
	case OrderRBG:
		return NewXColor(c.Red(), c.Blue(), c.Green())
	case OrderGRB:
		return NewXColor(c.Green(), c.Red(), c.Blue())
	case OrderGBR:
		return NewXColor(c.Green(), c.Blue(), c.Red())
	case OrderBRG:
		return NewXColor(c.Blue(), c.Red(), c.Green())
	case OrderBGR:
		return NewXColor(c.Blue(), c.Green(), c.Red())
	}
	panic(&RuntimeError{Class: "java.lang.IllegalStateException"})
}

// Reverse returns the negative of the reordered color.
func (o ColorOrder) Reverse(c XColor) XColor {
	c = o.Color(c)
	return NewXColor(255-c.Red(), 255-c.Green(), 255-c.Blue())
}

// ColorOrderFromString returns the order named s, ignoring case; ok is false
// when there is none (null upstream).
func ColorOrderFromString(s string) (o ColorOrder, ok bool) {
	upper := strings.ToUpper(s)
	for i, name := range colorOrderNames {
		if name == upper {
			return ColorOrder(i), true
		}
	}
	return 0, false
}
