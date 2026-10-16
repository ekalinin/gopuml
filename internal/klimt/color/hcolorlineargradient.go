// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColorLinearGradient.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "slices"

// SpreadMethod tells how a linear gradient behaves outside its vector, as
// SVG's spreadMethod.
type SpreadMethod int

// Spread methods, in upstream's order.
const (
	// SpreadPad extends the edge colors past the end stops; it is SVG's
	// default and the zero value, standing for upstream's null.
	SpreadPad SpreadMethod = iota
	// SpreadReflect reflects the gradient back and forth.
	SpreadReflect
	// SpreadRepeat repeats the gradient in the same direction.
	SpreadRepeat
)

// String returns upstream's constant name, for example "PAD".
func (m SpreadMethod) String() string {
	switch m {
	case SpreadReflect:
		return "REFLECT"
	case SpreadRepeat:
		return "REPEAT"
	}
	return "PAD"
}

// Stop is a gradient stop: an offset along the gradient vector, a color and
// an opacity multiplying the color's alpha.
type Stop struct {
	offset  float64
	color   HColor
	opacity float64
}

// NewStop returns a stop; the owning gradient normalizes the offset.
func NewStop(offset float64, color HColor, opacity float64) Stop {
	return Stop{offset: offset, color: color, opacity: opacity}
}

// Offset returns the offset along the gradient vector, between 0 and 1.
func (s Stop) Offset() float64 { return s.offset }

// Color returns the color of the stop.
func (s Stop) Color() HColor { return s.color }

// Opacity returns the opacity multiplier, between 0 and 1.
func (s Stop) Opacity() float64 { return s.opacity }

// HColorLinearGradient is an SVG-like linear gradient with several stops,
// built by the SVG sprite parser. Where a single color is needed, it stands
// for its first stop color.
type HColorLinearGradient struct {
	hcolor
	x1, y1, x2, y2 float64
	userSpaceOnUse bool
	spreadMethod   SpreadMethod
	stops          []Stop
}

// NewHColorLinearGradient returns a gradient from (x1, y1) to (x2, y2), in
// user space when userSpaceOnUse is true and in the object bounding box
// otherwise. The stops are sorted by offset, offsets are clamped to 0..1, and
// equal offsets are nudged forward by 0.000001 while that stays within 1.
func NewHColorLinearGradient(x1, y1, x2, y2 float64, userSpaceOnUse bool, spreadMethod SpreadMethod,
	stops []Stop) *HColorLinearGradient {
	c := &HColorLinearGradient{x1: x1, y1: y1, x2: x2, y2: y2, userSpaceOnUse: userSpaceOnUse,
		spreadMethod: spreadMethod, stops: normalizeStops(stops)}
	c.hcolor = hcolor{self: c, name: "HColorLinearGradient"}
	return c
}

func normalizeStops(input []Stop) []Stop {
	sorted := slices.Clone(input)
	slices.SortStableFunc(sorted, func(a, b Stop) int { return javaDoubleCompare(a.offset, b.offset) })

	normalized := make([]Stop, 0, len(sorted))
	last := -1.0
	for _, stop := range sorted {
		offset := stop.offset
		if offset < 0.0 {
			offset = 0.0
		}
		if offset > 1.0 {
			offset = 1.0
		}
		if offset <= last {
			candidate := last + 0.000001
			if candidate > 1.0 {
				offset = last
			} else {
				offset = candidate
			}
		}
		last = offset
		normalized = append(normalized, Stop{offset: offset, color: stop.color, opacity: stop.opacity})
	}
	return normalized
}

// X1 returns the start X.
func (c *HColorLinearGradient) X1() float64 { return c.x1 }

// Y1 returns the start Y.
func (c *HColorLinearGradient) Y1() float64 { return c.y1 }

// X2 returns the end X.
func (c *HColorLinearGradient) X2() float64 { return c.x2 }

// Y2 returns the end Y.
func (c *HColorLinearGradient) Y2() float64 { return c.y2 }

// IsUserSpaceOnUse reports whether the coordinates are in user space rather
// than in the object bounding box.
func (c *HColorLinearGradient) IsUserSpaceOnUse() bool { return c.userSpaceOnUse }

// SpreadMethod returns the spread method.
func (c *HColorLinearGradient) SpreadMethod() SpreadMethod { return c.spreadMethod }

// Stops returns a copy of the normalized stops.
func (c *HColorLinearGradient) Stops() []Stop { return slices.Clone(c.stops) }

// Color1 returns the first stop color. It panics without stops, as upstream
// throws.
func (c *HColorLinearGradient) Color1() HColor { return c.stops[0].color }

// ToColor returns the first stop color, mapped.
func (c *HColorLinearGradient) ToColor(mapper *ColorMapper) XColor { return c.Color1().ToColor(mapper) }

// Opposite returns the opposite of the first stop color.
func (c *HColorLinearGradient) Opposite() HColor { return c.Color1().Opposite() }

// IsDark reports whether the first stop color is dark.
func (c *HColorLinearGradient) IsDark() bool { return c.Color1().IsDark() }

// IsTransparent reports whether the first stop color is transparent.
func (c *HColorLinearGradient) IsTransparent() bool { return c.Color1().IsTransparent() }
