// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 e_exp.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)e_exp.c 1.6 04/04/22
//
// ====================================================
// Copyright (C) 2004 by Sun Microsystems, Inc. All rights reserved.
//
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================
//
// __ieee754_exp(x)
// Returns the exponential of x.
//
// Method
//   1. Argument reduction:
//      Reduce x to an r so that |r| <= 0.5*ln2 ~ 0.34658.
//	Given x, find r and integer k such that
//
//               x = k*ln2 + r,  |r| <= 0.5*ln2.
//
//      Here r will be represented as r = hi-lo for better
//	accuracy.
//
//   2. Approximation of exp(r) by a special rational function on
//	the interval [0,0.34658]:
//	Write
//	    R(r**2) = r*(exp(r)+1)/(exp(r)-1) = 2 + r*r/6 - r**4/360 + ...
//      We use a special Remes algorithm on [0,0.34658] to generate
// 	a polynomial of degree 5 to approximate R. The maximum error
//	of this polynomial approximation is bounded by 2**-59. In
//	other words,
//	    R(z) ~ 2.0 + P1*z + P2*z**2 + P3*z**3 + P4*z**4 + P5*z**5
//  	(where z=r*r, and the values of P1 to P5 are listed below)
//	and
//	    |                  5          |     -59
//	    | 2.0+P1*z+...+P5*z   -  R(z) | <= 2
//	    |                             |
//	The computation of exp(r) thus becomes
//                             2*r
//		exp(r) = 1 + -------
//		              R - r
//                                 r*R1(r)
//		       = 1 + r + ----------- (for better accuracy)
//		                  2 - R1(r)
//	where
//			         2       4             10
//		R1(r) = r - (P1*r  + P2*r  + ... + P5*r   ).
//
//   3. Scale back to obtain exp(x):
//	From step 1, we have
//	   exp(x) = 2^k * exp(r)
//
// Special cases:
//	exp(INF) is INF, exp(NaN) is NaN;
//	exp(-INF) is 0, and
//	for finite argument, only exp(0)=1 is exact.
//
// Accuracy:
//	according to an error analysis, the error is always less than
//	1 ulp (unit in the last place).
//
// Misc. info.
//	For IEEE double
//	    if x >  7.09782712893383973096e+02 then exp(x) overflow
//	    if x < -7.45133219101941108420e+02 then exp(x) underflow
//
// Constants:
// The hexadecimal values are the intended ones for the following
// constants. The decimal values may be used, provided that the
// compiler will convert from decimal to binary accurately enough
// to produce the hexadecimal values shown.

var halF = [2]float64{0.5, -0.5}

var ln2HI = [2]float64{
	6.93147180369123816490e-01,  // 0x3fe62e42, 0xfee00000
	-6.93147180369123816490e-01, // 0xbfe62e42, 0xfee00000
}

var ln2LO = [2]float64{
	1.90821492927058770002e-10,  // 0x3dea39ef, 0x35793c76
	-1.90821492927058770002e-10, // 0xbdea39ef, 0x35793c76
}

// Exp returns e**x, bit-identical to Java's StrictMath.exp. It is
// __ieee754_exp(x).
func Exp(x float64) float64 {
	const (
		one         float64 = 1.0
		o_threshold float64 = 7.09782712893383973096e+02  // 0x40862E42, 0xFEFA39EF
		u_threshold float64 = -7.45133219101941108420e+02 // 0xc0874910, 0xD52D3051
		invln2      float64 = 1.44269504088896338700e+00  // 0x3ff71547, 0x652b82fe
		P1          float64 = 1.66666666666666019037e-01  // 0x3FC55555, 0x5555553E
		P2          float64 = -2.77777777770155933842e-03 // 0xBF66C16C, 0x16BEBD93
		P3          float64 = 6.61375632143793436117e-05  // 0x3F11566A, 0xAF25DE2C
		P4          float64 = -1.65339022054652515390e-06 // 0xBEBBBD41, 0xC5D26BF1
		P5          float64 = 4.13813679705723846039e-08  // 0x3E663769, 0x72BEA4D0
	)
	// huge and twom1000 are variables so that huge*huge and
	// twom1000*twom1000 are evaluated in float64 arithmetic (to infinity
	// and zero) like the C code does.
	huge := 1.0e+300
	twom1000 := 9.33263618503218878990e-302 // 2**-1000=0x01700000,0
	var y, hi_, lo_, c, t float64           // hi_ and lo_ are hi and lo in the C source
	var k, xsb int32
	var hx uint32

	hx = uint32(hi(x))          // high word of x
	xsb = int32((hx >> 31) & 1) // sign bit of x
	hx &= 0x7fffffff            // high word of |x|

	// filter out non-finite argument
	if hx >= 0x40862E42 { // if |x|>=709.78...
		if hx >= 0x7ff00000 {
			if (hx&0xfffff)|lo(x) != 0 {
				return x + x // NaN
			}
			if xsb == 0 { // exp(+-inf)={inf,0}
				return x
			}
			return 0.0
		}
		if x > o_threshold {
			return float64(huge * huge) // overflow
		}
		if x < u_threshold {
			return float64(twom1000 * twom1000) // underflow
		}
	}

	// argument reduction
	if hx > 0x3fd62e42 { // if  |x| > 0.5 ln2
		if hx < 0x3FF0A2B2 { // and |x| < 1.5 ln2
			hi_ = x - ln2HI[xsb]
			lo_ = ln2LO[xsb]
			k = 1 - xsb - xsb
		} else {
			k = int32(float64(invln2*x) + halF[xsb])
			t = float64(k)
			hi_ = x - float64(t*ln2HI[0]) // t*ln2HI is exact here
			lo_ = float64(t * ln2LO[0])
		}
		x = hi_ - lo_
	} else if hx < 0x3e300000 { // when |x|<2**-28
		if huge+x > one {
			return one + x // trigger inexact
		}
	} else {
		k = 0
	}

	// x is now in primary range
	t = float64(x * x)
	c = x - float64(t*(P1+float64(t*(P2+float64(t*(P3+float64(t*(P4+float64(t*P5)))))))))
	if k == 0 {
		return one - ((x*c)/(c-2.0) - x)
	}
	y = one - ((lo_ - (x*c)/(2.0-c)) - hi_)
	if k >= -1021 {
		y = setHi(y, hi(y)+(k<<20)) // add k to y's exponent
		return y
	}
	y = setHi(y, hi(y)+((k+1000)<<20)) // add k to y's exponent
	return float64(y * twom1000)
}
