// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/SvgGraphics.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
)

// The number formatting of SvgGraphics. Upstream formats with
// String.format(Locale.US, "%.<decimal>f", x), which on JDK 21 rounds the
// shortest decimal representation of x half up (jfmt.FormatFixed), then
// trims trailing zeros. A negative value that rounds to zero keeps its sign
// ("-0"), only an exact zero (either sign) gives "0".

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
// when nothing is left after it.
func trimZeros(s string) string {
	dot := strings.IndexByte(s, '.')
	if dot < 0 {
		return s
	}
	end := len(s) - 1
	for end > dot && s[end] == '0' {
		end--
	}
	if end == dot {
		end--
	}
	return s[:end+1]
}
