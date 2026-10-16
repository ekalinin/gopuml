// SPDX-License-Identifier: MIT
// Provenance: original

// Package jfmt formats numbers exactly like the Java library calls that
// PlantUML uses for its output.
package jfmt

import (
	"math"
	"strconv"
	"strings"
)

// shortest returns the shortest decimal digits that round-trip x (no sign,
// no leading zeros) and the decimal exponent e such that x = 0.d1d2... * 10^e.
func shortest(x float64) (digits string, exp int) {
	s := strconv.FormatFloat(math.Abs(x), 'e', -1, 64) // d.ddddde±xx
	mantissa, e, _ := strings.Cut(s, "e")
	n, _ := strconv.Atoi(e)
	return strings.Replace(mantissa, ".", "", 1), n + 1
}

// FormatFixed is Java's String.format(Locale.US, "%.<prec>f", x): the
// shortest decimal representation of x, rounded half up to prec fractional
// digits. A negative value keeps its sign even when it rounds to zero.
func FormatFixed(x float64, prec int) string {
	switch {
	case math.IsNaN(x):
		return "NaN"
	case math.IsInf(x, 1):
		return "Infinity"
	case math.IsInf(x, -1):
		return "-Infinity"
	}
	neg := math.Signbit(x)
	var intPart, frac []byte
	if x == 0 {
		intPart = []byte("0")
	} else {
		digits, exp := shortest(x)
		// Position all digits around the decimal point.
		var all []byte
		point := exp
		if exp <= 0 {
			all = append([]byte(strings.Repeat("0", 1-exp)), digits...)
			point = 1
		} else {
			all = []byte(digits)
			for len(all) < exp {
				all = append(all, '0')
			}
		}
		intPart, frac = all[:point], all[point:]
		if len(frac) > prec {
			roundUp := frac[prec] >= '5'
			frac = frac[:prec]
			if roundUp {
				carry := true
				for i := len(frac) - 1; i >= 0 && carry; i-- {
					if frac[i] == '9' {
						frac[i] = '0'
					} else {
						frac[i]++
						carry = false
					}
				}
				for i := len(intPart) - 1; i >= 0 && carry; i-- {
					if intPart[i] == '9' {
						intPart[i] = '0'
					} else {
						intPart[i]++
						carry = false
					}
				}
				if carry {
					intPart = append([]byte("1"), intPart...)
				}
			}
		}
		intPart = []byte(strings.TrimLeft(string(intPart), "0"))
		if len(intPart) == 0 {
			intPart = []byte("0")
		}
	}
	for len(frac) < prec {
		frac = append(frac, '0')
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.Write(intPart)
	if prec > 0 {
		b.WriteByte('.')
		b.Write(frac)
	}
	return b.String()
}
