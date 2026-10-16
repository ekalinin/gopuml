// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import "math"

// Constants used by java.lang.Math.toRadians and toDegrees.
const (
	degreesToRadians = 0.017453292519943295
	radiansToDegrees = 57.29577951308232
)

// Sqrt returns the correctly rounded square root of x, like Java's
// Math.sqrt and StrictMath.sqrt.
func Sqrt(x float64) float64 { return math.Sqrt(x) }

// ToRadians converts an angle in degrees to radians like Java's
// Math.toRadians.
func ToRadians(angdeg float64) float64 {
	// The conversion keeps the product from fusing with an addition in an
	// inlining caller.
	return float64(angdeg * degreesToRadians)
}

// ToDegrees converts an angle in radians to degrees like Java's
// Math.toDegrees.
func ToDegrees(angrad float64) float64 {
	return float64(angrad * radiansToDegrees)
}

// Round returns the closest int64 to x, with ties rounding toward positive
// infinity, like Java's Math.round(double). NaN gives 0, and values outside
// the int64 range saturate to math.MinInt64 or math.MaxInt64.
func Round(x float64) int64 {
	if math.IsNaN(x) {
		return 0
	}
	r := math.Floor(x)
	// x-r is exact except for x in (-0.5, 0), where it lies in (0.5, 1)
	// and rounding cannot take it below 0.5.
	if x-r >= 0.5 {
		r++
	}
	switch {
	case r >= 0x1p63:
		return math.MaxInt64
	case r <= -0x1p63:
		return math.MinInt64
	}
	return int64(r)
}

// RoundF returns the closest int32 to x, with ties rounding toward positive
// infinity, like Java's Math.round(float). NaN gives 0, and values outside
// the int32 range saturate to math.MinInt32 or math.MaxInt32.
func RoundF(x float32) int32 {
	if x != x {
		return 0
	}
	// d-r is exact except for x in (-0.5, 0), where it lies in (0.5, 1)
	// and rounding cannot take it below 0.5.
	d := float64(x)
	r := math.Floor(d)
	if d-r >= 0.5 {
		r++
	}
	switch {
	case r >= 0x1p31:
		return math.MaxInt32
	case r <= -0x1p31:
		return math.MinInt32
	}
	return int32(r)
}

// Rint returns the integer closest to x, with ties rounding to even, like
// Java's Math.rint. The sign of a zero result follows x.
func Rint(x float64) float64 {
	const twoToThe52 = 0x1p52
	a := math.Abs(x)
	if a < twoToThe52 {
		// Adding and removing 2**52 rounds a to an integer in the current
		// (round half to even) rounding mode.
		a = (twoToThe52 + a) - twoToThe52
	}
	return math.Copysign(a, x)
}

// Floor returns the largest integer value less than or equal to x, like
// Java's Math.floor.
func Floor(x float64) float64 { return math.Floor(x) }

// Ceil returns the smallest integer value greater than or equal to x, like
// Java's Math.ceil.
func Ceil(x float64) float64 { return math.Ceil(x) }

// Signum returns 1.0 or -1.0 with the sign of x, or x itself when x is zero
// or NaN, like Java's Math.signum.
func Signum(x float64) float64 {
	if x == 0 || x != x {
		return x
	}
	return math.Copysign(1, x)
}

// Abs returns the absolute value of x, like Java's Math.abs(double).
func Abs(x float64) float64 { return math.Abs(x) }

// Max returns the larger of a and b like Java's Math.max(double, double):
// NaN if either is NaN, and +0.0 is larger than -0.0.
func Max(a, b float64) float64 {
	switch {
	case a != a:
		return a
	case b != b:
		return b
	case a == 0 && b == 0:
		if math.Float64bits(a)>>63 != 0 {
			return b
		}
		return a
	case a >= b:
		return a
	default:
		return b
	}
}

// Min returns the smaller of a and b like Java's Math.min(double, double):
// NaN if either is NaN, and -0.0 is smaller than +0.0.
func Min(a, b float64) float64 {
	switch {
	case a != a:
		return a
	case b != b:
		return b
	case a == 0 && b == 0:
		if math.Float64bits(a)>>63 != 0 {
			return a
		}
		return b
	case a <= b:
		return a
	default:
		return b
	}
}
