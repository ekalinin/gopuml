// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 e_asin.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)e_asin.c 1.3 95/01/18
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
// __ieee754_asin(x)
// Method :
//	Since  asin(x) = x + x^3/6 + x^5*3/40 + x^7*15/336 + ...
//	we approximate asin(x) on [0,0.5] by
//		asin(x) = x + x*x^2*R(x^2)
//	where
//		R(x^2) is a rational approximation of (asin(x)-x)/x^3
//	and its remez error is bounded by
//		|(asin(x)-x)/x^3 - R(x^2)| < 2^(-58.75)
//
//	For x in [0.5,1]
//		asin(x) = pi/2-2*asin(sqrt((1-x)/2))
//	Let y = (1-x), z = y/2, s := sqrt(z), and pio2_hi+pio2_lo=pi/2;
//	then for x>0.98
//		asin(x) = pi/2 - 2*(s+s*z*R(z))
//			= pio2_hi - (2*(s+s*z*R(z)) - pio2_lo)
//	For x<=0.98, let pio4_hi = pio2_hi/2, then
//		f = hi part of s;
//		c = sqrt(z) - f = (z-f*f)/(s+f) 	...f+c=sqrt(z)
//	and
//		asin(x) = pi/2 - 2*(s+s*z*R(z))
//			= pio4_hi+(pio4-2s)-(2s*z*R(z)-pio2_lo)
//			= pio4_hi+(pio4-2f)-(2s*z*R(z)-(pio2_lo+2c))
//
// Special cases:
//	if x is NaN, return x itself;
//	if |x|>1, return NaN with invalid signal.

import "math"

// Asin returns the arc sine of x, bit-identical to Java's StrictMath.asin.
// It is __ieee754_asin(x).
func Asin(x float64) float64 {
	const (
		one     float64 = 1.00000000000000000000e+00 // 0x3FF00000, 0x00000000
		huge    float64 = 1.000e+300
		pio2_hi float64 = 1.57079632679489655800e+00 // 0x3FF921FB, 0x54442D18
		pio2_lo float64 = 6.12323399573676603587e-17 // 0x3C91A626, 0x33145C07
		pio4_hi float64 = 7.85398163397448278999e-01 // 0x3FE921FB, 0x54442D18
		// coefficient for R(x^2)
		pS0 float64 = 1.66666666666666657415e-01  // 0x3FC55555, 0x55555555
		pS1 float64 = -3.25565818622400915405e-01 // 0xBFD4D612, 0x03EB6F7D
		pS2 float64 = 2.01212532134862925881e-01  // 0x3FC9C155, 0x0E884455
		pS3 float64 = -4.00555345006794114027e-02 // 0xBFA48228, 0xB5688F3B
		pS4 float64 = 7.91534994289814532176e-04  // 0x3F49EFE0, 0x7501B288
		pS5 float64 = 3.47933107596021167570e-05  // 0x3F023DE1, 0x0DFDF709
		qS1 float64 = -2.40339491173441421878e+00 // 0xC0033A27, 0x1C8A2D4B
		qS2 float64 = 2.02094576023350569471e+00  // 0x40002AE5, 0x9C598AC8
		qS3 float64 = -6.88283971605453293030e-01 // 0xBFE6066C, 0x1B8D0159
		qS4 float64 = 7.70381505559019352791e-02  // 0x3FB3B8C5, 0xB12E9282
	)
	var t, w, p, q, c, r, s float64
	var hx, ix int32
	hx = hi(x)
	ix = hx & 0x7fffffff
	if ix >= 0x3ff00000 { // |x|>= 1
		if uint32(ix-0x3ff00000)|lo(x) == 0 {
			// asin(1)=+-pi/2 with inexact
			return float64(x*pio2_hi) + float64(x*pio2_lo)
		}
		return (x - x) / (x - x) // asin(|x|>1) is NaN
	} else if ix < 0x3fe00000 { // |x|<0.5
		if ix < 0x3e400000 { // if |x| < 2**-27
			if huge+x > one {
				return x // return x with inexact if x!=0
			}
		} else {
			t = float64(x * x)
		}
		// In the C source the following lines follow an unbraced else
		// and so run for both branches above.
		p = float64(t * (pS0 + float64(t*(pS1+float64(t*(pS2+float64(t*(pS3+float64(t*(pS4+float64(t*pS5)))))))))))
		q = one + float64(t*(qS1+float64(t*(qS2+float64(t*(qS3+float64(t*qS4)))))))
		w = p / q
		return x + float64(x*w)
	}
	// 1> |x|>= 0.5
	w = one - math.Abs(x)
	t = float64(w * 0.5)
	p = float64(t * (pS0 + float64(t*(pS1+float64(t*(pS2+float64(t*(pS3+float64(t*(pS4+float64(t*pS5)))))))))))
	q = one + float64(t*(qS1+float64(t*(qS2+float64(t*(qS3+float64(t*qS4)))))))
	s = math.Sqrt(t)
	if ix >= 0x3FEF3333 { // if |x| > 0.975
		w = p / q
		t = pio2_hi - (float64(2.0*(s+float64(s*w))) - pio2_lo)
	} else {
		w = setLo(s, 0)
		c = (t - float64(w*w)) / (s + w)
		r = p / q
		p = float64(2.0*s*r) - (pio2_lo - float64(2.0*c))
		q = pio4_hi - float64(2.0*w)
		t = pio4_hi - (p - q)
	}
	if hx > 0 {
		return t
	}
	return -t
}
