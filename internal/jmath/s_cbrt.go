// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 s_cbrt.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)s_cbrt.c 1.3 95/01/18
//
// ====================================================
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
//
// Developed at SunSoft, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================

// Cbrt returns the cube root of x, bit-identical to Java's StrictMath.cbrt.
func Cbrt(x float64) float64 {
	const (
		B1 uint32 = 715094163 // B1 = (682-0.03306235651)*2**20
		B2 uint32 = 696219795 // B2 = (664-0.03306235651)*2**20

		C float64 = 5.42857142857142815906e-01  // 19/35     = 0x3FE15F15, 0xF15F15F1
		D float64 = -7.05306122448979611050e-01 // -864/1225 = 0xBFE691DE, 0x2532C834
		E float64 = 1.41428571428571436819e+00  // 99/70     = 0x3FF6A0EA, 0x0EA0EA0F
		F float64 = 1.60714285714285720630e+00  // 45/28     = 0x3FF9B6DB, 0x6DB6DB6E
		G float64 = 3.57142857142857150787e-01  // 5/14      = 0x3FD6DB6D, 0xB6DB6DB7
	)
	var hx int32
	var r, s, w float64
	var t float64 = 0.0
	var sign uint32

	hx = hi(x)                     // high word of x
	sign = uint32(hx) & 0x80000000 // sign= sign(x)
	hx = int32(uint32(hx) ^ sign)
	if hx >= 0x7ff00000 {
		return x + x // cbrt(NaN,INF) is itself
	}
	if uint32(hx)|lo(x) == 0 {
		return x // cbrt(0) is itself
	}

	x = setHi(x, hx) // x <- |x|
	// rough cbrt to 5 bits
	if hx < 0x00100000 { // subnormal number
		t = setHi(t, 0x43500000) // set t= 2**54
		t = float64(t * x)
		t = setHi(t, int32(uint32(hi(t)/3)+B2))
	} else {
		t = setHi(t, int32(uint32(hx/3)+B1))
	}

	// new cbrt to 23 bits, may be implemented in single precision
	r = t * t / x
	s = C + float64(r*t)
	t = float64(t * (G + F/(s+E+D/s)))

	// chopped to 20 bits and make it larger than cbrt(x)
	t = setLo(t, 0)
	t = setHi(t, hi(t)+0x00000001)

	// one step newton iteration to 53 bits with error less than 0.667 ulps
	s = float64(t * t) // t*t is exact
	r = x / s
	w = t + t
	r = (r - t) / (w + r) // r-s is exact
	t = t + float64(t*r)

	// retore the sign bit
	t = setHi(t, int32(uint32(hi(t))|sign))
	return t
}
