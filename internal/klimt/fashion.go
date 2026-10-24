// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/Fashion.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/color"

// Fashion is the look of a shape: colors, stroke, shadow and corners. It is
// immutable; the With methods return new values.
type Fashion struct {
	backColor      color.HColor
	foreColor      color.HColor
	stroke         UStroke
	deltaShadow    float64
	roundCorner    float64
	diagonalCorner float64
}

// NewFashion returns the fashion of two colors, either of which may be nil,
// with the simple stroke and no shadow or corner.
func NewFashion(backColor, foreColor color.HColor) *Fashion {
	return &Fashion{backColor: backColor, foreColor: foreColor, stroke: UStrokeSimple()}
}

// String is upstream's toString, without Java's identity hash.
func (f *Fashion) String() string {
	return "Fashion backColor=" + colorString(f.backColor) + " foreColor=" + colorString(f.foreColor)
}

// Apply applies the colors and the stroke to ug.
func (f *Fashion) Apply(ug UGraphic) UGraphic {
	ug = f.ApplyColors(ug)
	ug = f.ApplyStroke(ug)
	return ug
}

// ApplyColors applies the colors to ug, a nil color being transparent.
func (f *Fashion) ApplyColors(ug UGraphic) UGraphic {
	if f.foreColor == nil {
		ug = ug.Apply(color.None())
	} else {
		ug = ug.Apply(f.foreColor)
	}
	if f.backColor == nil {
		ug = ug.Apply(Bg(color.None()))
	} else {
		ug = ug.Apply(Bg(f.backColor))
	}
	return ug
}

// ApplyStroke applies the stroke to ug.
func (f *Fashion) ApplyStroke(ug UGraphic) UGraphic { return ug.Apply(f.stroke) }

func (f *Fashion) with(change func(c *Fashion)) *Fashion {
	c := *f
	change(&c)
	return &c
}

// WithShadow returns the fashion with another shadow size.
func (f *Fashion) WithShadow(deltaShadow float64) *Fashion {
	return f.with(func(c *Fashion) { c.deltaShadow = deltaShadow })
}

// WithDeltaShadow returns the fashion with another shadow size.
func (f *Fashion) WithDeltaShadow(deltaShadow float64) *Fashion { return f.WithShadow(deltaShadow) }

// WithStroke returns the fashion with another stroke.
func (f *Fashion) WithStroke(stroke UStroke) *Fashion {
	return f.with(func(c *Fashion) { c.stroke = stroke })
}

// WithBackColor returns the fashion with another background color.
func (f *Fashion) WithBackColor(backColor color.HColor) *Fashion {
	return f.with(func(c *Fashion) { c.backColor = backColor })
}

// WithForeColor returns the fashion with another foreground color.
func (f *Fashion) WithForeColor(foreColor color.HColor) *Fashion {
	return f.with(func(c *Fashion) { c.foreColor = foreColor })
}

// WithCorner returns the fashion with other corners.
func (f *Fashion) WithCorner(roundCorner, diagonalCorner float64) *Fashion {
	return f.with(func(c *Fashion) { c.roundCorner, c.diagonalCorner = roundCorner, diagonalCorner })
}

// BackColor returns the background color, possibly nil.
func (f *Fashion) BackColor() color.HColor { return f.backColor }

// ForeColor returns the foreground color, possibly nil.
func (f *Fashion) ForeColor() color.HColor { return f.foreColor }

// Stroke returns the stroke.
func (f *Fashion) Stroke() UStroke { return f.stroke }

// IsShadowing reports whether the shadow size is positive.
func (f *Fashion) IsShadowing() bool { return f.deltaShadow > 0 }

// DeltaShadow returns the shadow size.
func (f *Fashion) DeltaShadow() float64 { return f.deltaShadow }

// RoundCorner returns the round corner size.
func (f *Fashion) RoundCorner() float64 { return f.roundCorner }

// DiagonalCorner returns the diagonal corner size.
func (f *Fashion) DiagonalCorner() float64 { return f.diagonalCorner }

// colorString is Java's string concatenation of a color: "null" for nil.
func colorString(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.String()
}
