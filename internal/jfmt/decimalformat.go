// SPDX-License-Identifier: MIT
// Provenance: original

package jfmt

import (
	"math"
	"math/big"
	"strings"
)

// DecimalFormat is java.text.DecimalFormat in the en_US locale with the
// patterns "0" and "#,##0" (a minimum of one integer digit, with or without
// grouping by 3), an optional positive prefix, settable separators, the
// minimum and maximum fraction digits, and the default HALF_EVEN rounding.
// The pattern has no negative subpattern, so a negative value is the minus
// sign followed by the positive prefix ("-+5").
type DecimalFormat struct {
	// PositivePrefix is the prefix of the pattern, "+" for "+0".
	PositivePrefix string
	// GroupingUsed groups the integer digits by 3 with GroupingSeparator.
	GroupingUsed      bool
	GroupingSeparator rune
	// DecimalSeparator is written before the fraction digits.
	DecimalSeparator                             rune
	MinimumFractionDigits, MaximumFractionDigits int
}

// NewDecimalFormat returns the format of the pattern "0" with the symbols of
// en_US: ',' groups, '.' separates the fraction, no fraction digits.
func NewDecimalFormat() *DecimalFormat {
	return &DecimalFormat{GroupingSeparator: ',', DecimalSeparator: '.'}
}

// Format is DecimalFormat.format(double). When x has no more fraction
// digits than the maximum, its shortest decimal digits are written, as the
// JDK does; otherwise the exact binary value of x is rounded HALF_EVEN, so
// 2.675 gives 2.67 and 0.005 gives 0.01 with two digits. A negative value
// that rounds to zero keeps its sign ("-0").
func (f *DecimalFormat) Format(x float64) string {
	prefix := f.PositivePrefix
	if x < 0 || (x == 0 && math.Signbit(x)) {
		prefix = "-" + prefix
	}
	switch {
	case math.IsNaN(x):
		return "NaN"
	case math.IsInf(x, 0):
		return prefix + "∞"
	}
	intPart, frac := f.digits(math.Abs(x))
	for len(frac) < f.MinimumFractionDigits {
		frac += "0"
	}
	var b strings.Builder
	b.WriteString(prefix)
	if f.GroupingUsed {
		for i, c := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				b.WriteRune(f.GroupingSeparator)
			}
			b.WriteRune(c)
		}
	} else {
		b.WriteString(intPart)
	}
	if frac != "" {
		b.WriteRune(f.DecimalSeparator)
		b.WriteString(frac)
	}
	return b.String()
}

// digits returns the integer digits (at least one) and the fraction
// digits, without trailing zeros, of x >= 0 at the maximum fraction digits.
func (f *DecimalFormat) digits(x float64) (string, string) {
	if x == 0 {
		return "0", ""
	}
	digits, exp := shortest(x)
	if fracDigits := len(digits) - exp; fracDigits <= f.MaximumFractionDigits {
		if exp <= 0 {
			return "0", strings.Repeat("0", -exp) + digits
		}
		if len(digits) <= exp {
			return digits + strings.Repeat("0", exp-len(digits)), ""
		}
		return digits[:exp], digits[exp:]
	}
	// Round x × 10^max to an integer, half to even, from the exact value.
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(f.MaximumFractionDigits)), nil)
	r := new(big.Rat).SetFloat64(x)
	r.Mul(r, new(big.Rat).SetInt(scale))
	q, rem := new(big.Int).QuoRem(r.Num(), r.Denom(), new(big.Int))
	if c := rem.Lsh(rem, 1).Cmp(r.Denom()); c > 0 || (c == 0 && q.Bit(0) == 1) {
		q.Add(q, big.NewInt(1))
	}
	ip, fp := new(big.Int).QuoRem(q, scale, new(big.Int))
	frac := ""
	if f.MaximumFractionDigits > 0 {
		s := fp.String()
		frac = strings.TrimRight(strings.Repeat("0", f.MaximumFractionDigits-len(s))+s, "0")
	}
	return ip.String(), frac
}
