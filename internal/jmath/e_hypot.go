// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 e_hypot.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)e_hypot.c 1.3 95/01/18
//
// ====================================================
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
//
// Developed at SunSoft, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================
//
// __ieee754_hypot(x,y)
//
// Method :
//	If (assume round-to-nearest) z=x*x+y*y
//	has error less than sqrt(2)/2 ulp, than
//	sqrt(z) has error less than 1 ulp (exercise).
//
//	So, compute sqrt(x*x+y*y) with some care as
//	follows to get the error below 1 ulp:
//
//	Assume x>y>0;
//	(if possible, set rounding to round-to-nearest)
//	1. if x > 2y  use
//		x1*x1+(y*y+(x2*(x+x1))) for x*x+y*y
//	where x1 = x with lower 32 bits cleared, x2 = x-x1; else
//	2. if x <= 2y use
//		t1*y1+((x-y)*(x-y)+(t1*y2+t2*y))
//	where t1 = 2x with lower 32 bits cleared, t2 = 2x-t1,
//	y1= y with lower 32 bits chopped, y2 = y-y1.
//
//	NOTE: scaling may be necessary if some argument is too
//	      large or too tiny
//
// Special cases:
//	hypot(x,y) is INF if x or y is +INF or -INF; else
//	hypot(x,y) is NAN if x or y is NAN.
//
// Accuracy:
//	hypot(x,y) returns sqrt(x^2+y^2) with error less
//	than 1 ulps (units in the last place)

import "math"

// Hypot returns sqrt(x*x+y*y) without undue overflow or underflow,
// bit-identical to Java's StrictMath.hypot, which Math.hypot calls. It is
// __ieee754_hypot(x, y) with the two changes of Java's port (FdLibm.Hypot):
// the special cases come first, and the arguments are ordered by their
// values rather than by their high words.
func Hypot(x, y float64) float64 {
	a := math.Abs(x)
	b := math.Abs(y)
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return math.Inf(1)
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return a + b
	}
	if b > a {
		a, b = b, a
	}
	ha := hi(a) // high word of a
	hb := hi(b) // high word of b
	if ha-hb > 0x3c00000 {
		return a + b // x/y > 2**60
	}
	k := int32(0)
	if ha > 0x5f300000 { // a>2**500
		// scale a and b by 2**-600
		ha -= 0x25800000
		hb -= 0x25800000
		k += 600
		a = setHi(a, ha)
		b = setHi(b, hb)
	}
	var t1, t2 float64
	if hb < 0x20b00000 { // b < 2**-500
		if hb <= 0x000fffff { // subnormal b or 0
			if b == 0 {
				return a
			}
			t1 = fromWords(0x7fd00000, 0) // t1=2^1022
			b *= t1
			a *= t1
			k -= 1022
		} else { // scale a and b by 2^600
			ha += 0x25800000 // a *= 2^600
			hb += 0x25800000 // b *= 2^600
			k -= 600
			a = setHi(a, ha)
			b = setHi(b, hb)
		}
	}
	// medium size a and b
	w := a - b
	if w > b {
		t1 = fromWords(ha, 0)
		t2 = a - t1
		w = math.Sqrt(float64(t1*t1) - (float64(b*(-b)) - float64(t2*(a+t1))))
	} else {
		a = a + a
		y1 := fromWords(hb, 0)
		y2 := b - y1
		t1 = fromWords(ha+0x00100000, 0)
		t2 = a - t1
		w = math.Sqrt(float64(t1*y1) - (float64(w*(-w)) - (float64(t1*y2) + float64(t2*b))))
	}
	if k != 0 {
		t1 = setHi(1.0, hi(1.0)+k<<20)
		return t1 * w
	}
	return w
}
