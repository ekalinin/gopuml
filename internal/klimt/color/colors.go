// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/Colors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "strings"

// LineStyle is the line style a Colors holds. It stands for upstream's
// decoration LinkStyle, which is not ported in this package: Colors never
// sets a thickness, so the type of the LinkStyle is all it needs.
type LineStyle int

// Line styles. LineStyleNone is the absence of a style (null upstream).
const (
	LineStyleNone LineStyle = iota
	LineStyleNormal
	LineStyleDashed
	LineStyleDotted
	LineStyleBold
	LineStyleInvisible
)

// String is LinkStyle.toString, "DASHED(null)" for example, and "null" for
// LineStyleNone.
func (s LineStyle) String() string {
	switch s {
	case LineStyleNormal:
		return "NORMAL(null)"
	case LineStyleDashed:
		return "DASHED(null)"
	case LineStyleDotted:
		return "DOTTED(null)"
	case LineStyleBold:
		return "BOLD(null)"
	case LineStyleInvisible:
		return "INVISIBLE(null)"
	}
	return "null"
}

// Colors holds the colors given to an element by type (text, line, back,
// header, arrow), with an optional line style and shadowing. A type may be
// present with a nil color, when its value did not parse. Colors values are
// not modified once built: the methods return new values.
type Colors struct {
	colors    [numColorTypes]HColor
	present   [numColorTypes]bool
	lineStyle LineStyle
	shadowing *bool
}

// String is upstream's toString, for example
// "{BACK=[r=255,g=0,b=0,a=255] α=255} DASHED(null)".
func (c *Colors) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for t := range numColorTypes {
		if !c.present[t] {
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(t.String())
		b.WriteByte('=')
		if c.colors[t] == nil {
			b.WriteString("null")
		} else {
			b.WriteString(c.colors[t].String())
		}
	}
	b.WriteString("} ")
	b.WriteString(c.lineStyle.String())
	return b.String()
}

// EmptyColors returns a Colors without colors.
func EmptyColors() *Colors { return &Colors{} }

// IsEmpty reports whether no color type is present.
func (c *Colors) IsEmpty() bool {
	for _, p := range c.present {
		if p {
			return false
		}
	}
	return true
}

// copy keeps the colors and the line style, not the shadowing, as upstream.
func (c *Colors) copy() *Colors {
	return &Colors{colors: c.colors, present: c.present, lineStyle: c.lineStyle}
}

func (c *Colors) put(t ColorType, color HColor) {
	c.colors[t] = color
	c.present[t] = true
}

func (c *Colors) putAll(other *Colors) {
	for t := range numColorTypes {
		if other.present[t] {
			c.put(t, other.colors[t])
		}
	}
}

// MergeWith returns c with the colors of other added, and its line style
// when it has one. A nil other returns c.
func (c *Colors) MergeWith(other *Colors) *Colors {
	if other == nil {
		return c
	}
	result := c.copy()
	result.putAll(other)
	if other.lineStyle != LineStyleNone {
		result.lineStyle = other.lineStyle
	}
	return result
}

// NewColors parses the colors of a command, such as "#red",
// "#back:red;line:blue;text:green", "#line.dashed;header:yellow" or
// "#shadowing:false". The data is lower-cased and every '#' removed, then
// each ';'-separated part is either a color for mainType, a "type:color"
// pair whose color may not parse (giving a nil color), "shadowing:true" or
// "shadowing:false", or a line style such as "line.dotted". A plain color
// that does not parse is a *NoSuchColorError. An unknown type panics with an
// IllegalArgumentException RuntimeError, as upstream throws.
func NewColors(data string, set *HColorSet, mainType ColorType) (*Colors, error) {
	c := &Colors{}
	data = strings.ReplaceAll(javaLowerCase(data), "#", "")

	for _, s := range strings.FieldsFunc(data, func(r rune) bool { return r == ';' }) {
		name, value, found := strings.Cut(s, ":")
		if !found {
			if !strings.Contains(s, ".") {
				color, err := set.GetColor(s)
				if err != nil {
					return nil, err
				}
				c.put(mainType, color)
			}
			continue
		}
		if javaEqualsIgnoreCase(name, "shadowing") {
			shadowing := javaEqualsIgnoreCase(value, "true")
			c.shadowing = &shadowing
		} else {
			key := ColorTypeOf(name)
			c.put(key, set.GetColorOrNull(value))
		}
	}
	switch {
	case strings.Contains(data, "line.dashed"):
		c.lineStyle = LineStyleDashed
	case strings.Contains(data, "line.dotted"):
		c.lineStyle = LineStyleDotted
	case strings.Contains(data, "line.bold"):
		c.lineStyle = LineStyleBold
	}
	return c, nil
}

// Color returns the color of type key, nil when there is none.
func (c *Colors) Color(key ColorType) HColor { return c.colors[key] }

// FirstColor returns the color of key1, or of key2 when key1 has none.
func (c *Colors) FirstColor(key1, key2 ColorType) HColor {
	if result := c.Color(key1); result != nil {
		return result
	}
	return c.Color(key2)
}

// Add returns c with color as the color of t; a nil color returns c.
func (c *Colors) Add(t ColorType, color HColor) *Colors {
	if color == nil {
		return c
	}
	result := c.copy()
	result.put(t, color)
	return result
}

// LineStyle returns the line style, LineStyleNone when there is none.
func (c *Colors) LineStyle() LineStyle { return c.lineStyle }

// AddLegacyStroke returns c with the line style named s, ignoring case:
// "dashed", "dotted", "bold" or "hidden" (invisible), and normal for any
// other name.
func (c *Colors) AddLegacyStroke(s string) *Colors {
	result := c.copy()
	result.lineStyle = linkStyleFromString(strings.ToUpper(s))
	return result
}

// linkStyleFromString is upstream's LinkStyle.fromString1.
func linkStyleFromString(s string) LineStyle {
	switch {
	case javaEqualsIgnoreCase(s, "dashed"):
		return LineStyleDashed
	case javaEqualsIgnoreCase(s, "dotted"):
		return LineStyleDotted
	case javaEqualsIgnoreCase(s, "bold"):
		return LineStyleBold
	case javaEqualsIgnoreCase(s, "hidden"):
		return LineStyleInvisible
	}
	return LineStyleNormal
}

// Shadowing returns the shadowing given by "shadowing:true" or
// "shadowing:false"; ok is false when there is none (null upstream).
func (c *Colors) Shadowing() (shadowing, ok bool) {
	if c.shadowing == nil {
		return false, false
	}
	return *c.shadowing, true
}
