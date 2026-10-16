// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 e_log.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)e_log.c 1.3 95/01/18
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
// __ieee754_log(x)
// Return the logrithm of x
//
// Method :
//   1. Argument Reduction: find k and f such that
//			x = 2^k * (1+f),
//	   where  sqrt(2)/2 < 1+f < sqrt(2) .
//
//   2. Approximation of log(1+f).
//	Let s = f/(2+f) ; based on log(1+f) = log(1+s) - log(1-s)
//		 = 2s + 2/3 s**3 + 2/5 s**5 + .....,
//	     	 = 2s + s*R
//      We use a special Reme algorithm on [0,0.1716] to generate
// 	a polynomial of degree 14 to approximate R The maximum error
//	of this polynomial approximation is bounded by 2**-58.45. In
//	other words,
//		        2      4      6      8      10      12      14
//	    R(z) ~ Lg1*s +Lg2*s +Lg3*s +Lg4*s +Lg5*s  +Lg6*s  +Lg7*s
//  	(the values of Lg1 to Lg7 are listed in the program)
//	and
//	    |      2          14          |     -58.45
//	    | Lg1*s +...+Lg7*s    -  R(z) | <= 2
//	    |                             |
//	Note that 2s = f - s*f = f - hfsq + s*hfsq, where hfsq = f*f/2.
//	In order to guarantee error in log below 1ulp, we compute log
//	by
//		log(1+f) = f - s*(f - R)	(if f is not too large)
//		log(1+f) = f - (hfsq - s*(hfsq+R)).	(better accuracy)
//
//	3. Finally,  log(x) = k*ln2 + log(1+f).
//			    = k*ln2_hi+(f-(hfsq-(s*(hfsq+R)+k*ln2_lo)))
//	   Here ln2 is split into two floating point number:
//			ln2_hi + ln2_lo,
//	   where n*ln2_hi is always exact for |n| < 2000.
//
// Special cases:
//	log(x) is NaN with signal if x < 0 (including -INF) ;
//	log(+INF) is +INF; log(0) is -INF with signal;
//	log(NaN) is that NaN with no signal.
//
// Accuracy:
//	according to an error analysis, the error is always less than
//	1 ulp (unit in the last place).
//
// Constants:
// The hexadecimal values are the intended ones for the following
// constants. The decimal values may be used, provided that the
// compiler will convert from decimal to binary accurately enough
// to produce the hexadecimal values shown.

// Log returns the natural logarithm of x, bit-identical to Java's
// StrictMath.log. It is __ieee754_log(x).
func Log(x float64) float64 {
	const (
		ln2_hi float64 = 6.93147180369123816490e-01 // 3fe62e42 fee00000
		ln2_lo float64 = 1.90821492927058770002e-10 // 3dea39ef 35793c76
		two54  float64 = 1.80143985094819840000e+16 // 43500000 00000000
		Lg1    float64 = 6.666666666666735130e-01   // 3FE55555 55555593
		Lg2    float64 = 3.999999999940941908e-01   // 3FD99999 9997FA04
		Lg3    float64 = 2.857142874366239149e-01   // 3FD24924 94229359
		Lg4    float64 = 2.222219843214978396e-01   // 3FCC71C5 1D8E78AF
		Lg5    float64 = 1.818357216161805012e-01   // 3FC74664 96CB03DE
		Lg6    float64 = 1.531383769920937332e-01   // 3FC39A09 D078C69F
		Lg7    float64 = 1.479819860511658591e-01   // 3FC2F112 DF3E5244
	)
	// zero is a variable, as in the C source, so that -two54/zero and
	// (x-x)/zero are evaluated at run time.
	zero := 0.0
	var hfsq, f, s, z, R, w, t1, t2, dk float64
	var k, hx, i, j int32
	var lx uint32

	hx = hi(x) // high word of x
	lx = lo(x) // low  word of x

	k = 0
	if hx < 0x00100000 { // x < 2**-1022
		if uint32(hx&0x7fffffff)|lx == 0 {
			return -two54 / zero // log(+-0)=-inf
		}
		if hx < 0 {
			return (x - x) / zero // log(-#) = NaN
		}
		k -= 54
		x = float64(x * two54) // subnormal number, scale up x
		hx = hi(x)             // high word of x
	}
	if hx >= 0x7ff00000 {
		return x + x
	}
	k += (hx >> 20) - 1023
	hx &= 0x000fffff
	i = (hx + 0x95f64) & 0x100000
	x = setHi(x, hx|(i^0x3ff00000)) // normalize x or x/2
	k += (i >> 20)
	f = x - 1.0
	if (0x000fffff & (2 + hx)) < 3 { // |f| < 2**-20
		if f == zero {
			if k == 0 {
				return zero
			}
			dk = float64(k)
			return float64(dk*ln2_hi) + float64(dk*ln2_lo)
		}
		R = float64(f * f * (0.5 - float64(0.33333333333333333*f)))
		if k == 0 {
			return f - R
		}
		dk = float64(k)
		return float64(dk*ln2_hi) - ((R - float64(dk*ln2_lo)) - f)
	}
	s = f / (2.0 + f)
	dk = float64(k)
	z = float64(s * s)
	i = hx - 0x6147a
	w = float64(z * z)
	j = 0x6b851 - hx
	t1 = float64(w * (Lg2 + float64(w*(Lg4+float64(w*Lg6)))))
	t2 = float64(z * (Lg1 + float64(w*(Lg3+float64(w*(Lg5+float64(w*Lg7)))))))
	i |= j
	R = t2 + t1
	if i > 0 {
		hfsq = float64(0.5 * f * f)
		if k == 0 {
			return f - (hfsq - float64(s*(hfsq+R)))
		}
		return float64(dk*ln2_hi) - ((hfsq - (float64(s*(hfsq+R)) + float64(dk*ln2_lo))) - f)
	}
	if k == 0 {
		return f - float64(s*(f-R))
	}
	return float64(dk*ln2_hi) - ((float64(s*(f-R)) - float64(dk*ln2_lo)) - f)
}
