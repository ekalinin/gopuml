// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import "math"

// Word access to the IEEE 754 representation of a float64, in place of
// fdlibm's __HI and __LO macros. Words are int32 for the high half and
// uint32 for the low half, matching how fdlibm uses them.

// hi returns the high 32-bit word of x, like __HI(x).
func hi(x float64) int32 { return int32(math.Float64bits(x) >> 32) }

// lo returns the low 32-bit word of x, like __LO(x).
func lo(x float64) uint32 { return uint32(math.Float64bits(x)) }

// setHi returns x with its high word replaced by h, like __HI(x) = h.
func setHi(x float64, h int32) float64 {
	return math.Float64frombits(uint64(uint32(h))<<32 | math.Float64bits(x)&0xffffffff)
}

// setLo returns x with its low word replaced by l, like __LO(x) = l.
func setLo(x float64, l uint32) float64 {
	return math.Float64frombits(math.Float64bits(x)&^0xffffffff | uint64(l))
}

// fromWords builds a float64 from its high and low words.
func fromWords(h int32, l uint32) float64 {
	return math.Float64frombits(uint64(uint32(h))<<32 | uint64(l))
}
