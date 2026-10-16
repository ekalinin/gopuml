// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"strconv"
	"unicode"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
)

// javaDoubleString formats d like Java's Double.toString; see jfmt.
func javaDoubleString(d float64) string { return jfmt.DoubleToString(d) }

// javaDoubleHashCode returns Java's Double.hashCode(d): the two halves of the
// canonical IEEE 754 bit pattern XOR-ed together, with all NaNs mapped to
// 0x7ff8000000000000 as Double.doubleToLongBits does.
func javaDoubleHashCode(d float64) int32 {
	bits := uint64(0x7ff8000000000000)
	if !math.IsNaN(d) {
		bits = math.Float64bits(d)
	}
	return int32(bits ^ bits>>32)
}

// javaD2I converts like Java's (int) cast of a double: truncation toward
// zero, NaN becomes 0 and out of range values saturate.
func javaD2I(d float64) int32 {
	switch {
	case math.IsNaN(d):
		return 0
	case d >= math.MaxInt32:
		return math.MaxInt32
	case d <= math.MinInt32:
		return math.MinInt32
	}
	return int32(d)
}

// javaEqualsIgnoreCase reports whether a and b are equal according to Java's
// String.equalsIgnoreCase: characters match when they are equal, when their
// upper case forms are equal, or when the lower case forms of those are equal.
// Unlike strings.EqualFold, "ı" (dotless i) and "İ" (I with dot)
// both match "I".
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

// atan2 is Java's Math.atan2. Java computes it with fdlibm, whose result can
// differ from Go's math.Atan2 in the last bit.
func atan2(y, x float64) float64 {
	return jmath.Atan2(y, x)
}

// enumName returns names[v], or "Type(v)" for a value outside the enum.
func enumName[T ~int](typ string, names []string, v T) string {
	if v < 0 || int(v) >= len(names) {
		return typ + "(" + strconv.Itoa(int(v)) + ")"
	}
	return names[v]
}

// enumValueOf is Java's Enum.valueOf: the constant whose name is exactly s.
func enumValueOf[T ~int](names []string, s string) (T, bool) {
	for i, name := range names {
		if name == s {
			return T(i), true
		}
	}
	return 0, false
}
