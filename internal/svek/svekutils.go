// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/SvekUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ArithmeticStrategy.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ArithmeticStrategySum.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/DotMode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Point2DFunction.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/YDelta.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ColorSequence.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/EmptySvgException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// svekGetValue is SvekUtils.getValue: the number of the attribute varName
// found after starting in svg.
func svekGetValue(svg string, starting int, varName string) float64 {
	varNameString := varName + "=\""
	p1 := indexFrom(svg, varNameString, starting)
	if p1 == -1 {
		throwIllegalState("")
	}
	p1 += len(varNameString)
	p2 := indexFrom(svg, "\"", p1)
	return javaParseDouble(svg[p1:p2])
}

// svekGetMinXY is SvekUtils.getMinXY: the smallest x and the smallest y of
// the points. It panics on an empty list, as upstream.
func svekGetMinXY(points []geom.XPoint2D) geom.XPoint2D {
	if len(points) == 0 {
		throwIndexOutOfBounds(0, 0)
	}
	minx, miny := points[0].X, points[0].Y
	for _, p := range points[1:] {
		if p.X < minx {
			minx = p.X
		}
		if p.Y < miny {
			miny = p.Y
		}
	}
	return geom.NewXPoint2D(minx, miny)
}

// svekGetMaxXY is SvekUtils.getMaxXY.
func svekGetMaxXY(points []geom.XPoint2D) geom.XPoint2D {
	if len(points) == 0 {
		throwIndexOutOfBounds(0, 0)
	}
	maxx, maxy := points[0].X, points[0].Y
	for _, p := range points[1:] {
		if p.X > maxx {
			maxx = p.X
		}
		if p.Y > maxy {
			maxy = p.Y
		}
	}
	return geom.NewXPoint2D(maxx, maxy)
}

// println is SvekUtils.println.
func println(sb *strings.Builder) { sb.WriteByte('\n') }

// PixelToInches is SvekUtils.pixelToInches: String.format("%6.6f") of the
// size in inches.
func PixelToInches(pixel float64) string {
	return jfmt.FormatFixed(pixel/72.0, 6)
}

// arithmeticStrategy is upstream's ArithmeticStrategy.
type arithmeticStrategy interface {
	eat(v float64)
	result() float64
}

// arithmeticStrategySum is upstream's ArithmeticStrategySum.
type arithmeticStrategySum struct{ sum float64 }

func (a *arithmeticStrategySum) eat(v float64)   { a.sum += v }
func (a *arithmeticStrategySum) result() float64 { return a.sum }

// DotMode is upstream's DotMode.
type DotMode int

// Dot modes, in upstream's order.
const (
	DotModeNormal DotMode = iota
	DotModeNoLeftRightAndXLabel
)

// point2DFunction is upstream's Point2DFunction: the conversion of dot's
// coordinates.
type point2DFunction interface {
	apply(pt geom.XPoint2D) geom.XPoint2D
}

// yDelta is upstream's YDelta: y becomes y * factor + delta.
type yDelta struct{ factor, delta float64 }

func newYDelta(delta float64) yDelta { return yDelta{factor: 1, delta: delta} }

func (y yDelta) apply(pt geom.XPoint2D) geom.XPoint2D {
	return geom.NewXPoint2D(pt.X, float64(pt.Y*y.factor)+y.delta)
}

// colorSequence is upstream's ColorSequence: the colors that identify the
// nodes, edges and clusters in dot's output, 2, 3, 4...
type colorSequence struct{ cpt int32 }

func newColorSequence() *colorSequence { return &colorSequence{cpt: 1} }

func (c *colorSequence) value() int32 {
	c.cpt++
	return c.cpt
}

// hexColor is XColor.toHexRGBColor.
func hexColor(c int32) string { return color.ToHexRGBColor(c) }

// indexFrom is Java's String.indexOf(s, from) on byte offsets (dot's SVG
// output is ASCII).
func indexFrom(s, substr string, from int) int {
	if from < 0 {
		from = 0
	}
	if from > len(s) {
		return -1
	}
	i := strings.Index(s[from:], substr)
	if i < 0 {
		return -1
	}
	return from + i
}

// javaParseDouble is Double.parseDouble for the numbers of dot's output.
// It panics with NumberFormatException on bad input.
func javaParseDouble(s string) float64 {
	t := strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
	t = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(t, "d"), "D"), "f"), "F")
	v, err := strconv.ParseFloat(t, 64)
	if err != nil || strings.ContainsAny(t, "xXpP_") || strings.EqualFold(t, "inf") ||
		strings.EqualFold(t, "+inf") || strings.EqualFold(t, "-inf") || strings.EqualFold(t, "nan") {
		throwNumberFormat(s)
	}
	return v
}

func itoa32(i int32) string { return strconv.Itoa(int(i)) }

// errEmptySvg is upstream's EmptySvgException.
var errEmptySvg = &emptySvgError{}

type emptySvgError struct{}

func (*emptySvgError) Error() string { return "net.sourceforge.plantuml.svek.EmptySvgException" }

func itoa64(i int64) string { return strconv.FormatInt(i, 10) }
