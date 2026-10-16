// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HColorSet.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import (
	"slices"
	"strings"
)

// HColorSet parses colors.
//
// A color is, after an optional '#':
//   - "transparent" or "background" (the transparent color) or "automatic",
//     ignoring case;
//   - a simple color: 1 hex digit (a gray, "#8" being #888888), 3 hex digits
//     (#RGB), 6 hex digits (#RRGGBB), 8 hex digits (#RRGGBBAA, alpha last), or
//     a color name, ignoring the case of ASCII letters;
//   - a scheme "?light:dark" or "?light:dark:transparent" of simple colors;
//   - a gradient of two simple colors separated by '-', '\\', '|' or '/',
//     each part with an optional '#'.
type HColorSet struct{}

var hcolorSet = &HColorSet{}

// Instance returns the color set, upstream's HColorSet.instance().
func Instance() *HColorSet { return hcolorSet }

// Names returns the color names in Java's String order (upper case before
// lower case).
func (set *HColorSet) Names() []string { return slices.Clone(colorNames) }

// GetColorOrWhite returns the color of s, or White when s is not a color.
func (set *HColorSet) GetColorOrWhite(s string) HColor {
	if result := set.parseColor(s); result != nil {
		return result
	}
	return White
}

// GetColorOrNull returns the color of s, or nil when s is not a color.
func (set *HColorSet) GetColorOrNull(s string) HColor { return set.parseColor(s) }

// GetColor returns the color of s, or a *NoSuchColorError when s is not a
// color.
func (set *HColorSet) GetColor(s string) (HColor, error) {
	if result := set.parseColor(s); result != nil {
		return result, nil
	}
	return nil, &NoSuchColorError{}
}

// parseColor returns nil when s is not a color. As upstream, a two-color
// scheme with an invalid color, such as "?red:nocolor", panics with an
// ArrayIndexOutOfBoundsException RuntimeError: upstream reads the third
// color that is not there.
func (set *HColorSet) parseColor(s string) HColor {
	s = strings.TrimPrefix(s, "#")

	if javaEqualsIgnoreCase(s, "transparent") || javaEqualsIgnoreCase(s, "background") {
		return None()
	}

	if javaEqualsIgnoreCase(s, "automatic") {
		return newHColorAutomagic()
	}

	if result := set.parseSimpleColor(s); result != nil {
		return newHColorSimple(result, nil, WithFillNone)
	}

	if rest, ok := strings.CutPrefix(s, "?"); ok {
		colors := javaSplitColon(rest)
		if len(colors) == 2 || len(colors) == 3 {
			col0 := set.parseSimpleColor(colors[0])
			col1 := set.parseSimpleColor(colors[1])
			if len(colors) == 2 && col0 != nil && col1 != nil {
				return NewHColorScheme(simpleOrNull(col0), simpleOrNull(col1), nil)
			}
			if len(colors) == 2 {
				throw(arrayIndexOutOfBounds, "Index 2 out of bounds for length 2")
			}
			if col2 := set.parseSimpleColor(colors[2]); col2 != nil {
				return NewHColorScheme(simpleOrNull(col0), simpleOrNull(col1), simpleOrNull(col2))
			}
		}
	}

	for i := range len(s) {
		c := s[i]
		if c == '-' || c == '\\' || c == '|' || c == '/' {
			col0 := set.parseSimpleColor(s[:i])
			col1 := set.parseSimpleColor(s[i+1:])
			if col0 != nil && col1 != nil {
				return Gradient(simpleOrNull(col0), simpleOrNull(col1), rune(c))
			}
		}
	}

	return nil
}

// simpleOrNull is HColors.simple on a possibly null XColor: a nil color gives
// a simple color that panics when used.
func simpleOrNull(c *XColor) HColor { return newHColorSimple(c, nil, WithFillNone) }

// javaSplitColon is Java's s.split(":"): trailing empty strings are removed,
// and a string without ':' gives itself.
func javaSplitColon(s string) []string {
	if !strings.Contains(s, ":") {
		return []string{s}
	}
	parts := strings.Split(s, ":")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// parseSimpleColor returns nil when s, after an optional '#', is neither 1,
// 3, 6 or 8 hex digits nor a color name.
func (set *HColorSet) parseSimpleColor(s string) *XColor {
	s = strings.TrimPrefix(s, "#")
	var c XColor
	switch len(s) {
	case 1:
		if d := hexNibble(s[0]); d >= 0 {
			v := d<<4 | d
			c = XColorFrom(v<<16 | v<<8 | v)
			return &c
		}
	case 3:
		r, g, b := hexNibble(s[0]), hexNibble(s[1]), hexNibble(s[2])
		if r >= 0 && g >= 0 && b >= 0 {
			c = NewXColor(r<<4|r, g<<4|g, b<<4|b)
			return &c
		}
	case 6:
		if rgb := parseHex24(s, 0); rgb >= 0 {
			c = XColorFrom(0xFF000000 | rgb)
			return &c
		}
	case 8:
		// https://forum.plantuml.net/11606/full-opacity-alpha-compositing-support-for-svg-and-png
		r, g, b, a := parseHexByte(s, 0), parseHexByte(s, 2), parseHexByte(s, 4), parseHexByte(s, 6)
		if r >= 0 && g >= 0 && b >= 0 && a >= 0 {
			c = NewXColorAlpha(r, g, b, a)
			return &c
		}
	}
	return ColorTrie.lookup(s)
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func parseHexByte(s string, off int) int {
	hi, lo := hexNibble(s[off]), hexNibble(s[off+1])
	if hi|lo < 0 {
		return -1
	}
	return hi<<4 | lo
}

func parseHex24(s string, off int) int {
	r, g, b := parseHexByte(s, off), parseHexByte(s, off+2), parseHexByte(s, off+4)
	if r|g|b < 0 {
		return -1
	}
	return r<<16 | g<<8 | b
}
