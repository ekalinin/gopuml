// SPDX-License-Identifier: MIT
// Provenance: fdlibm 5.3 e_atan2.c
// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.

package jmath

// @(#)e_atan2.c 1.3 95/01/18
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
// __ieee754_atan2(y,x)
// Method :
//	1. Reduce y to positive by atan2(y,x)=-atan2(-y,x).
//	2. Reduce x to positive by (if x and y are unexceptional):
//		ARG (x+iy) = arctan(y/x)   	   ... if x > 0,
//		ARG (x+iy) = pi - arctan[y/(-x)]   ... if x < 0,
//
// Special cases:
//
//	ATAN2((anything), NaN ) is NaN;
//	ATAN2(NAN , (anything) ) is NaN;
//	ATAN2(+-0, +(anything but NaN)) is +-0  ;
//	ATAN2(+-0, -(anything but NaN)) is +-pi ;
//	ATAN2(+-(anything but 0 and NaN), 0) is +-pi/2;
//	ATAN2(+-(anything but INF and NaN), +INF) is +-0 ;
//	ATAN2(+-(anything but INF and NaN), -INF) is +-pi;
//	ATAN2(+-INF,+INF ) is +-pi/4 ;
//	ATAN2(+-INF,-INF ) is +-3pi/4;
//	ATAN2(+-INF, (anything but,0,NaN, and INF)) is +-pi/2;
//
// Constants:
// The hexadecimal values are the intended ones for the following
// constants. The decimal values may be used, provided that the
// compiler will convert from decimal to binary accurately enough
// to produce the hexadecimal values shown.

import "math"

// Atan2 returns the angle of the point (x, y), bit-identical to Java's
// StrictMath.atan2. It is __ieee754_atan2(y, x).
func Atan2(y, x float64) float64 {
	const (
		tiny   float64 = 1.0e-300
		zero   float64 = 0.0
		pi_o_4 float64 = 7.8539816339744827900e-01 // 0x3FE921FB, 0x54442D18
		pi_o_2 float64 = 1.5707963267948965580e+00 // 0x3FF921FB, 0x54442D18
		pi     float64 = 3.1415926535897931160e+00 // 0x400921FB, 0x54442D18
		pi_lo  float64 = 1.2246467991473531772e-16 // 0x3CA1A626, 0x33145C07
	)
	var z float64
	var k, m, hx, hy, ix, iy int32
	var lx, ly uint32

	hx = hi(x)
	ix = hx & 0x7fffffff
	lx = lo(x)
	hy = hi(y)
	iy = hy & 0x7fffffff
	ly = lo(y)
	if (uint32(ix)|((lx|-lx)>>31)) > 0x7ff00000 ||
		(uint32(iy)|((ly|-ly)>>31)) > 0x7ff00000 { // x or y is NaN
		return x + y
	}
	if uint32(hx-0x3ff00000)|lx == 0 {
		return Atan(y) // x=1.0
	}
	m = ((hy >> 31) & 1) | ((hx >> 30) & 2) // 2*sign(x)+sign(y)

	// when y = 0
	if uint32(iy)|ly == 0 {
		switch m {
		case 0, 1:
			return y // atan(+-0,+anything)=+-0
		case 2:
			return pi + tiny // atan(+0,-anything) = pi
		case 3:
			return -pi - tiny // atan(-0,-anything) =-pi
		}
	}
	// when x = 0
	if uint32(ix)|lx == 0 {
		if hy < 0 {
			return -pi_o_2 - tiny
		}
		return pi_o_2 + tiny
	}

	// when x is INF
	if ix == 0x7ff00000 {
		if iy == 0x7ff00000 {
			switch m {
			case 0:
				return pi_o_4 + tiny // atan(+INF,+INF)
			case 1:
				return -pi_o_4 - tiny // atan(-INF,+INF)
			case 2:
				return 3.0*pi_o_4 + tiny // atan(+INF,-INF)
			case 3:
				return -3.0*pi_o_4 - tiny // atan(-INF,-INF)
			}
		} else {
			switch m {
			case 0:
				return zero // atan(+...,+INF)
			case 1:
				return math.Copysign(zero, -1) // atan(-...,+INF); a negated constant zero would be +0 in Go
			case 2:
				return pi + tiny // atan(+...,-INF)
			case 3:
				return -pi - tiny // atan(-...,-INF)
			}
		}
	}
	// when y is INF
	if iy == 0x7ff00000 {
		if hy < 0 {
			return -pi_o_2 - tiny
		}
		return pi_o_2 + tiny
	}

	// compute y/x
	k = (iy - ix) >> 20
	if k > 60 { // |y/x| >  2**60
		z = pi_o_2 + 0.5*pi_lo
	} else if hx < 0 && k < -60 { // |y|/x < -2**60
		z = 0.0
	} else { // safe to do y/x
		z = Atan(math.Abs(y / x))
	}
	switch m {
	case 0:
		return z // atan(+,+)
	case 1:
		return math.Float64frombits(math.Float64bits(z) ^ (1 << 63)) // atan(-,+): flip the sign bit
	case 2:
		return pi - (z - pi_lo) // atan(+,-)
	default: // case 3
		return (z - pi_lo) - pi // atan(-,-)
	}
}
