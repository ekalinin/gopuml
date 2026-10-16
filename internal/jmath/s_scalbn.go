// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 s_scalbn.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)s_scalbn.c 1.3 95/01/18
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
// scalbn (double x, int n)
// scalbn(x,n) returns x* 2**n  computed by  exponent
// manipulation rather than by actually performing an
// exponentiation or a multiplication.

import "math"

// scalbn(x, n) returns x*2**n computed by exponent manipulation.
func scalbn(x float64, n int32) float64 {
	const (
		two54  float64 = 1.80143985094819840000e+16 // 0x43500000, 0x00000000
		twom54 float64 = 5.55111512312578270212e-17 // 0x3C900000, 0x00000000
	)
	// huge and tiny are variables so that their products are evaluated in
	// float64 arithmetic (to infinity or zero) like the C code does.
	huge, tiny := 1.0e+300, 1.0e-300
	var k, hx int32
	var lx uint32
	hx = hi(x)
	lx = lo(x)
	k = (hx & 0x7ff00000) >> 20 // extract exponent
	if k == 0 {                 // 0 or subnormal x
		if lx|uint32(hx&0x7fffffff) == 0 {
			return x // +-0
		}
		x = float64(x * two54)
		hx = hi(x)
		k = ((hx & 0x7ff00000) >> 20) - 54
		if n < -50000 {
			return float64(tiny * x) // underflow
		}
	}
	if k == 0x7ff {
		return x + x // NaN or Inf
	}
	k = k + n
	if k > 0x7fe {
		return float64(huge * math.Copysign(huge, x)) // overflow
	}
	if k > 0 { // normal result
		return setHi(x, int32(uint32(hx)&0x800fffff)|(k<<20))
	}
	if k <= -54 {
		if n > 50000 { // in case integer overflow in n+k
			return float64(huge * math.Copysign(huge, x)) // overflow
		}
		return float64(tiny * math.Copysign(tiny, x)) // underflow
	}
	k += 54 // subnormal result
	x = setHi(x, int32(uint32(hx)&0x800fffff)|(k<<20))
	return float64(x * twom54)
}
