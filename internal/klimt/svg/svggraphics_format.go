// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/SvgGraphics.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/jfmt"
)

// The number formatting of SvgGraphics, upstream's StringUtils.formatDecimal:
// the result of String.format(Locale.US, "%.<decimal>f", x), which on JDK 21
// rounds the shortest decimal representation of x half up
// (jfmt.FormatFixed), with trailing zeros trimmed. A negative value that
// rounds to zero gives "0", not "-0".

// Format formats a coordinate: the value times the scale, with the option's
// decimals, trailing zeros removed.
func (s *SvgGraphics) Format(xx float64) string {
	return formatNumber(xx, s.option.scale, s.option.decimal)
}

// FormatBoolean formats an arc flag: "0" for zero, else "1".
func (s *SvgGraphics) FormatBoolean(x float64) string { return formatBoolean(x) }

// FormatPercent formats a gradient position given as a fraction, with at
// least 2 decimals and a "%" suffix.
func (s *SvgGraphics) FormatPercent(value float64) string {
	return formatPercent(value, s.option.decimal)
}

// FormatOpacity formats an opacity clamped to [0, 1], with at least 2
// decimals.
func (s *SvgGraphics) FormatOpacity(value float64) string {
	return formatOpacity(value, s.option.decimal)
}

func formatNumber(xx, scale float64, decimal int) string {
	x := xx * scale
	if x == 0 {
		return "0"
	}
	return trimZeros(jfmt.FormatFixed(x, decimal))
}

func formatBoolean(x float64) string {
	if x == 0 {
		return "0"
	}
	return "1"
}

func formatPercent(value float64, decimal int) string {
	percent := value * 100.0
	if percent == 0 {
		return "0%"
	}
	return trimZeros(jfmt.FormatFixed(percent, max(decimal, 2))) + "%"
}

func formatOpacity(value float64, decimal int) string {
	if value <= 0 {
		return "0"
	}
	if value >= 1 {
		return "1"
	}
	return trimZeros(jfmt.FormatFixed(value, max(decimal, 2)))
}

// trimZeros removes trailing zeros after the decimal point, and the point
// when nothing is left after it; "-0" becomes "0".
func trimZeros(s string) string {
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		end := len(s) - 1
		for end > dot && s[end] == '0' {
			end--
		}
		if end == dot {
			end--
		}
		s = s[:end+1]
	}
	if s == "-0" {
		return "0"
	}
	return s
}
