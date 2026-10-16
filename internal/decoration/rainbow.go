// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/Rainbow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package decoration

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// Rainbow is the part of upstream's Rainbow that the links of graph
// diagrams use: the color of a line and of its arrow heads, read from the
// arrow style. (The activity diagrams have the full Rainbow with its line
// styles, in internal/activitydiagram3.)
type Rainbow struct {
	color, arrowHeadColor color.HColor
}

// RainbowBuild is Rainbow.build(Style, HColorSet): the line color of the
// style, and its head color, which is the line color when the style has
// none and transparent when it is set to a missing color. A style without
// line color gives the empty rainbow, whose colors are nil.
func RainbowBuild(style *klimt.Style, set *color.HColorSet) Rainbow {
	c := style.Value(klimt.PNameLineColor).AsColor(set)
	var colorHead color.HColor
	if head := style.Value(klimt.PNameHeadColor); isValueNull(head) {
		colorHead = c
	} else {
		colorHead = head.AsColor(set)
	}
	if colorHead == nil {
		colorHead = color.Transparent()
	}
	if c == nil {
		return Rainbow{}
	}
	return Rainbow{color: c, arrowHeadColor: colorHead}
}

func isValueNull(v klimt.Value) bool {
	_, ok := v.(*klimt.ValueNull)
	return ok
}

// Color returns the line color. Upstream throws IndexOutOfBoundsException
// for the empty rainbow; here it panics the same way.
func (r Rainbow) Color() color.HColor {
	if r.color == nil {
		emptyRainbow()
	}
	return r.color
}

// ArrowHeadColor returns the color of the arrow heads.
func (r Rainbow) ArrowHeadColor() color.HColor {
	if r.color == nil {
		emptyRainbow()
	}
	return r.arrowHeadColor
}

func emptyRainbow() {
	panic(&color.RuntimeError{Class: "java.lang.IndexOutOfBoundsException",
		Message: "Index 0 out of bounds for length 0"})
}
