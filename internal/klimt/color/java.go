// SPDX-License-Identifier: MIT
// Provenance: original

package color

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// RuntimeError is the panic value of the code paths where upstream throws an
// unchecked Java exception, such as UnsupportedOperationException from
// HColor.toColor on a color scheme. Class is the qualified name of the Java
// exception and Message its message, empty when upstream has none.
type RuntimeError struct {
	Class   string
	Message string
}

// Error formats the exception as Java's Throwable.toString does.
func (e *RuntimeError) Error() string {
	if e.Message == "" {
		return e.Class
	}
	return e.Class + ": " + e.Message
}

const (
	unsupportedOperation  = "java.lang.UnsupportedOperationException"
	illegalArgument       = "java.lang.IllegalArgumentException"
	nullPointer           = "java.lang.NullPointerException"
	arrayIndexOutOfBounds = "java.lang.ArrayIndexOutOfBoundsException"
	numberFormat          = "java.lang.NumberFormatException"
)

func throw(class, message string) { panic(&RuntimeError{Class: class, Message: message}) }

// javaD2I is Java's (int) cast of a double: truncation toward zero, NaN to
// zero and saturation at the int range.
func javaD2I(x float64) int {
	switch {
	case math.IsNaN(x):
		return 0
	case x >= math.MaxInt32:
		return math.MaxInt32
	case x <= math.MinInt32:
		return math.MinInt32
	}
	return int(x)
}

// javaRound is Java's Math.round(double): the closest long, ties rounding
// toward positive infinity, NaN to zero.
func javaRound(x float64) int64 {
	if math.IsNaN(x) {
		return 0
	}
	r := math.Floor(x)
	if x-r >= 0.5 {
		r++
	}
	switch {
	case r >= math.MaxInt64:
		return math.MaxInt64
	case r <= math.MinInt64:
		return math.MinInt64
	}
	return int64(r)
}

// javaRoundF is Java's Math.round(float). Adding 0.5 to the float widened to
// double is exact, so the floor gives the closest int, ties up.
func javaRoundF(x float32) int {
	if math.IsNaN(float64(x)) {
		return 0
	}
	return javaD2I(math.Floor(float64(x) + 0.5))
}

// javaFRem is Java's float remainder operator. The remainder is exact, so
// computing it in double gives the same float.
func javaFRem(x, y float32) float32 {
	return float32(math.Mod(float64(x), float64(y)))
}

// javaDoubleCompare is Java's Double.compare: -0.0 is less than 0.0 and NaN
// is greater than every other value and equal to itself.
func javaDoubleCompare(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	bits := func(x float64) int64 {
		if math.IsNaN(x) {
			return 0x7ff8000000000000
		}
		return int64(math.Float64bits(x))
	}
	ba, bb := bits(a), bits(b)
	switch {
	case ba == bb:
		return 0
	case ba < bb:
		return -1
	}
	return 1
}

// javaEqualsIgnoreCase is Java's String.equalsIgnoreCase: characters match
// when equal, when their upper cases are equal or when the lower cases of
// their upper cases are equal. Unlike strings.EqualFold, U+0131 (dotless i)
// and U+0130 (dotted capital I) match the ASCII i.
func javaEqualsIgnoreCase(a, b string) bool {
	ra, rb := []rune(a), []rune(b)
	if len(ra) != len(rb) {
		return false
	}
	for i, c1 := range ra {
		c2 := rb[i]
		if c1 == c2 {
			continue
		}
		u1, u2 := unicode.ToUpper(c1), unicode.ToUpper(c2)
		if u1 == u2 || unicode.ToLower(u1) == unicode.ToLower(u2) {
			continue
		}
		return false
	}
	return true
}

// javaLowerCase is Java's String.toLowerCase(Locale.ENGLISH) for the needs of
// color parsing: like strings.ToLower, except that U+0130 (dotted capital I)
// becomes "i" followed by U+0307, as in Java, and not a plain "i". The Greek
// final sigma rule is not applied; it cannot turn a string into a color name.
func javaLowerCase(s string) string {
	if !strings.ContainsRune(s, 'İ') {
		return strings.ToLower(s)
	}
	return strings.ToLower(strings.ReplaceAll(s, "İ", "i̇"))
}

// javaFloatToString is Java's Float.toString.
func javaFloatToString(f float32) string { return javaFloating(float64(f), 32) }

// javaDoubleToString is Java's Double.toString.
func javaDoubleToString(d float64) string { return javaFloating(d, 64) }

// javaFloating formats the shortest decimal that identifies v, in plain
// notation between 1e-3 and 1e7 and in computerized scientific notation
// otherwise, with at least one digit after the point.
func javaFloating(v float64, bitSize int) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case v == 0 && math.Signbit(v):
		return "-0.0"
	case v == 0:
		return "0.0"
	}
	neg, digits, exp := shortestDigits(v, bitSize)
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	if abs := math.Abs(v); abs >= 1e-3 && abs < 1e7 {
		writePlain(&b, digits, exp, 1)
		return b.String()
	}
	b.WriteString(digits[:1])
	b.WriteByte('.')
	if len(digits) > 1 {
		b.WriteString(digits[1:])
	} else {
		b.WriteByte('0')
	}
	b.WriteByte('E')
	b.WriteString(strconv.Itoa(exp))
	return b.String()
}

// shortestDigits returns the sign, the significant digits and the decimal
// exponent of the shortest decimal identifying v: v = d.ddd * 10^exp.
func shortestDigits(v float64, bitSize int) (neg bool, digits string, exp int) {
	s := strconv.FormatFloat(v, 'e', -1, bitSize)
	if s[0] == '-' {
		neg, s = true, s[1:]
	}
	mant, e, _ := strings.Cut(s, "e")
	exp, _ = strconv.Atoi(e)
	return neg, strings.Replace(mant, ".", "", 1), exp
}

// writePlain writes digits * 10^(exp+1-len(digits)) in plain notation with
// at least minFrac fraction digits.
func writePlain(b *strings.Builder, digits string, exp, minFrac int) {
	var intPart, frac string
	if exp >= 0 {
		if len(digits) <= exp+1 {
			intPart = digits + strings.Repeat("0", exp+1-len(digits))
		} else {
			intPart, frac = digits[:exp+1], digits[exp+1:]
		}
	} else {
		intPart, frac = "0", strings.Repeat("0", -exp-1)+digits
	}
	if len(frac) < minFrac {
		frac += strings.Repeat("0", minFrac-len(frac))
	}
	b.WriteString(intPart)
	b.WriteByte('.')
	b.WriteString(frac)
}

// javaFormatFixed is Java's String.format("%.<prec>f", v): Java's Formatter
// rounds the shortest decimal of v half up, where Go rounds the exact binary
// value, so 0.0078125 gives 0.007813 and not 0.007812.
func javaFormatFixed(v float64, prec int) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	}
	neg := math.Signbit(v)
	digits, exp := "0", 0
	if v != 0 {
		_, digits, exp = shortestDigits(v, 64)
	}
	var plain strings.Builder
	writePlain(&plain, digits, exp, prec)
	intPart, frac, _ := strings.Cut(plain.String(), ".")
	num := []byte(intPart + frac[:prec])
	if len(frac) > prec && frac[prec] >= '5' {
		i := len(num) - 1
		for ; i >= 0 && num[i] == '9'; i-- {
			num[i] = '0'
		}
		if i >= 0 {
			num[i]++
		} else {
			num = append([]byte{'1'}, num...)
		}
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	n := len(num) - prec
	b.Write(num[:n])
	if prec > 0 {
		b.WriteByte('.')
		b.Write(num[n:])
	}
	return b.String()
}
