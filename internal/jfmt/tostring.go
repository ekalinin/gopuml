// SPDX-License-Identifier: MIT
// Provenance: original

package jfmt

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// DoubleToString is Java's Double.toString (JDK 19 and later).
func DoubleToString(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	}
	return toString(v, 64)
}

// FloatToString is Java's Float.toString (JDK 19 and later).
func FloatToString(v float32) string {
	f := float64(v)
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	return toString(f, 32)
}

// toString returns the text of Java's Double.toString (bitSize 64)
// or Float.toString (bitSize 32) of the finite value v, as specified since
// JDK 19: the shortest decimal that rounds to v (at least two digits when a
// single digit would do), in plain notation with at least one fractional
// digit when 1e-3 <= |v| < 1e7 and in computerized scientific notation
// otherwise.
func toString(v float64, bitSize int) string {
	if v == 0 {
		if math.Signbit(v) {
			return "-0.0"
		}
		return "0.0"
	}
	var sb strings.Builder
	if v < 0 {
		sb.WriteByte('-')
		v = -v
	}
	digits, e := shortestJava(v, bitSize)
	n := len(digits)
	switch {
	case -3 <= e && e < 0:
		sb.WriteString("0.")
		sb.WriteString(strings.Repeat("0", -e-1))
		sb.WriteString(digits)
	case 0 <= e && e < 7:
		if n <= e+1 {
			sb.WriteString(digits)
			sb.WriteString(strings.Repeat("0", e+1-n))
			sb.WriteString(".0")
		} else {
			sb.WriteString(digits[:e+1])
			sb.WriteByte('.')
			sb.WriteString(digits[e+1:])
		}
	default:
		sb.WriteByte(digits[0])
		sb.WriteByte('.')
		if n == 1 {
			sb.WriteByte('0')
		} else {
			sb.WriteString(digits[1:])
		}
		sb.WriteByte('E')
		sb.WriteString(strconv.Itoa(e))
	}
	return sb.String()
}

// shortestJava returns the significant digits of the decimal that
// Java selects for the positive finite value v, without trailing zeros, and
// the exponent of its first digit: v ~ 0.d1d2... * 10^(e+1).
//
// Java takes the decimals of minimal length that round to v and picks the
// one closest to v, like strconv's shortest formatting. When that minimal
// length is 1, Java picks the closest among the decimals of length 1 or 2
// instead, so for example Double.MIN_VALUE is 4.9E-324 rather than 5E-324.
func shortestJava(v float64, bitSize int) (string, int) {
	digits, e := splitExponent(strconv.FormatFloat(v, 'e', -1, bitSize))
	if len(digits) > 1 {
		return digits, e
	}
	exact := new(big.Rat).SetFloat64(v)
	// Candidates are significand * 10^exp: the shortest decimal and the
	// two-digit decimals around v.
	type decimal struct {
		significand int64
		exp         int
	}
	first, _ := strconv.ParseInt(digits, 10, 64)
	candidates := []decimal{{first, e}}
	nearest, ne := splitExponent(strconv.FormatFloat(v, 'e', 1, bitSize))
	s, _ := strconv.ParseInt(nearest, 10, 64)
	for _, c := range []int64{s - 1, s, s + 1} {
		candidates = append(candidates, decimal{c, ne - 1})
	}
	var best decimal
	var bestDist *big.Rat
	for _, c := range candidates {
		text := strconv.FormatInt(c.significand, 10) + "e" + strconv.Itoa(c.exp)
		if f, err := strconv.ParseFloat(text, bitSize); err != nil || f != v {
			continue
		}
		r, _ := new(big.Rat).SetString(text)
		dist := r.Sub(r, exact).Abs(r)
		if bestDist == nil {
			best, bestDist = c, dist
			continue
		}
		switch cmp := dist.Cmp(bestDist); {
		case cmp < 0, cmp == 0 && c.significand%2 == 0 && best.significand%2 != 0:
			best, bestDist = c, dist
		}
	}
	text := strconv.FormatInt(best.significand, 10)
	e = best.exp + len(text) - 1
	return strings.TrimRight(text, "0"), e
}

// splitExponent splits strconv's 'e' format, such as "1.25e+06", into its
// digits without the point ("125") and its exponent (6).
func splitExponent(s string) (string, int) {
	mantissa, exp, _ := strings.Cut(s, "e")
	e, _ := strconv.Atoi(exp)
	return strings.Replace(mantissa, ".", "", 1), e
}
