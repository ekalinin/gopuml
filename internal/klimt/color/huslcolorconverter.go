// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/HUSLColorConverter.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
)

// The HSLuv and HPLuv conversions, taken by upstream from
// https://github.com/hsluv/hsluv-java (Alexei Boronine). Tuples are RGB with
// channels between 0 and 1, CIE XYZ, CIE LUV, LCh (lightness, chroma, hue in
// degrees), and HSLuv/HPLuv (hue in degrees, saturation and lightness
// between 0 and 100).
//
// The arithmetic follows upstream operation by operation. Products that are
// added are converted explicitly, which stops the compiler from fusing them
// into FMA instructions that Java never uses.

var (
	huslM = [3][3]float64{
		{3.240969941904521, -1.537383177570093, -0.498610760293},
		{-0.96924363628087, 1.87596750150772, 0.041555057407175},
		{0.055630079696993, -0.20397695888897, 1.056971514242878},
	}
	huslMinv = [3][3]float64{
		{0.41239079926595, 0.35758433938387, 0.18048078840183},
		{0.21263900587151, 0.71516867876775, 0.072192315360733},
		{0.019330818715591, 0.11919477979462, 0.95053215224966},
	}
)

const (
	huslRefY    = 1.0
	huslRefU    = 0.19783000664283
	huslRefV    = 0.46831999493879
	huslKappa   = 903.2962962
	huslEpsilon = 0.0088564516
)

func huslBounds(L float64) [6][2]float64 {
	sub1 := jmath.Pow(L+16, 3) / 1560896
	sub2 := L / huslKappa
	if sub1 > huslEpsilon {
		sub2 = sub1
	}
	var result [6][2]float64
	for c := range 3 {
		m1, m2, m3 := huslM[c][0], huslM[c][1], huslM[c][2]
		for t := range 2 {
			top1 := (float64(284517*m1) - float64(94839*m3)) * sub2
			sum := float64(float64(838422*m3)+float64(769860*m2)) + float64(731718*m1)
			top2 := float64(sum*L*sub2) - float64(float64(769860*t)*L)
			bottom := float64((float64(632260*m3)-float64(126452*m2))*sub2) + float64(126452*t)
			result[c*2+t] = [2]float64{top1 / bottom, top2 / bottom}
		}
	}
	return result
}

func intersectLineLine(lineA, lineB [2]float64) float64 {
	return (lineA[1] - lineB[1]) / (lineB[0] - lineA[0])
}

func distanceFromPole(point [2]float64) float64 {
	// Math.pow(x, 2) is x*x on the JVM, as in Go's math.Pow.
	return math.Sqrt(float64(point[0]*point[0]) + float64(point[1]*point[1]))
}

// lengthOfRayUntilIntersect returns the length and whether it is >= 0.
func lengthOfRayUntilIntersect(theta float64, line [2]float64) (float64, bool) {
	length := line[1] / (jmath.Sin(theta) - float64(line[0]*jmath.Cos(theta)))
	return length, length >= 0
}

func maxSafeChromaForL(L float64) float64 {
	bounds := huslBounds(L)
	minimum := math.MaxFloat64
	for i := range 2 {
		m1, b1 := bounds[i][0], bounds[i][1]
		x := intersectLineLine([2]float64{m1, b1}, [2]float64{-1 / m1, 0})
		length := distanceFromPole([2]float64{x, b1 + float64(x*m1)})
		minimum = min(minimum, length)
	}
	return minimum
}

func maxChromaForLH(L, H float64) float64 {
	hrad := H / 360 * math.Pi * 2
	minimum := math.MaxFloat64
	for _, bound := range huslBounds(L) {
		if length, ok := lengthOfRayUntilIntersect(hrad, bound); ok {
			minimum = min(minimum, length)
		}
	}
	return minimum
}

func dotProduct(a, b [3]float64) float64 {
	sum := 0.0
	for i := range a {
		sum += float64(a[i] * b[i])
	}
	return sum
}

func huslRound(value float64, places int) float64 {
	n := jmath.Pow(10, float64(places))
	return float64(javaRound(float64(value*n))) / n
}

func fromLinear(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return float64(1.055*jmath.Pow(c, 1/float64(2.4))) - 0.055
}

func toLinear(c float64) float64 {
	if c > 0.04045 {
		return jmath.Pow((c+0.055)/(1+float64(0.055)), 2.4)
	}
	return c / 12.92
}

// rgbPrepare rounds the channels to 3 decimals and scales them to 0..255.
// It panics with an IllegalArgumentException RuntimeError for a channel out
// of range, as upstream throws.
func rgbPrepare(tuple [3]float64) [3]int {
	var results [3]int
	for i, ch := range tuple {
		rounded := huslRound(ch, 3)
		if rounded < -0.0001 || rounded > 1.0001 {
			throw(illegalArgument, "Illegal rgb value: "+javaDoubleToString(rounded))
		}
		results[i] = int(int32(javaRound(float64(rounded * 255))))
	}
	return results
}

// XyzToRgb converts CIE XYZ to RGB.
func XyzToRgb(tuple [3]float64) [3]float64 {
	return [3]float64{fromLinear(dotProduct(huslM[0], tuple)), fromLinear(dotProduct(huslM[1], tuple)),
		fromLinear(dotProduct(huslM[2], tuple))}
}

// RgbToXyz converts RGB to CIE XYZ.
func RgbToXyz(tuple [3]float64) [3]float64 {
	rgbl := [3]float64{toLinear(tuple[0]), toLinear(tuple[1]), toLinear(tuple[2])}
	return [3]float64{dotProduct(huslMinv[0], rgbl), dotProduct(huslMinv[1], rgbl), dotProduct(huslMinv[2], rgbl)}
}

func yToL(Y float64) float64 {
	if Y <= huslEpsilon {
		return (Y / huslRefY) * huslKappa
	}
	return float64(116*jmath.Pow(Y/huslRefY, 1.0/3.0)) - 16
}

func lToY(L float64) float64 {
	if L <= 8 {
		return huslRefY * L / huslKappa
	}
	return huslRefY * jmath.Pow((L+16)/116, 3)
}

// XyzToLuv converts CIE XYZ to CIE LUV.
func XyzToLuv(tuple [3]float64) [3]float64 {
	X, Y, Z := tuple[0], tuple[1], tuple[2]
	denominator := float64(X+float64(15*Y)) + float64(3*Z)
	varU := (4 * X) / denominator
	varV := (9 * Y) / denominator
	L := yToL(Y)
	if L == 0 {
		return [3]float64{}
	}
	U := 13 * L * (varU - huslRefU)
	V := 13 * L * (varV - huslRefV)
	return [3]float64{L, U, V}
}

// LuvToXyz converts CIE LUV to CIE XYZ.
func LuvToXyz(tuple [3]float64) [3]float64 {
	L, U, V := tuple[0], tuple[1], tuple[2]
	if L == 0 {
		return [3]float64{}
	}
	varU := U/(13*L) + huslRefU
	varV := V/(13*L) + huslRefV
	Y := lToY(L)
	X := 0 - (9*Y*varU)/(float64((varU-4)*varV)-float64(varU*varV))
	Z := (float64(float64(9*Y)-float64(15*varV*Y)) - float64(varV*X)) / (3 * varV)
	return [3]float64{X, Y, Z}
}

// LuvToLch converts CIE LUV to LCh.
func LuvToLch(tuple [3]float64) [3]float64 {
	L, U, V := tuple[0], tuple[1], tuple[2]
	C := math.Sqrt(float64(U*U) + float64(V*V))
	var H float64
	if !(C < 0.00000001) {
		Hrad := jmath.Atan2(V, U)
		// pi to more digits than they provide it in the stdlib
		H = (Hrad * 180.0) / 3.1415926535897932
		if H < 0 {
			H = 360 + H
		}
	}
	return [3]float64{L, C, H}
}

// LchToLuv converts LCh to CIE LUV.
func LchToLuv(tuple [3]float64) [3]float64 {
	L, C, H := tuple[0], tuple[1], tuple[2]
	Hrad := H / 360.0 * 2 * math.Pi
	U := jmath.Cos(Hrad) * C
	V := jmath.Sin(Hrad) * C
	return [3]float64{L, U, V}
}

// HsluvToLch converts HSLuv to LCh.
func HsluvToLch(tuple [3]float64) [3]float64 {
	H, S, L := tuple[0], tuple[1], tuple[2]
	if L > 99.9999999 {
		return [3]float64{100, 0, H}
	}
	if L < 0.00000001 {
		return [3]float64{0, 0, H}
	}
	C := maxChromaForLH(L, H) / 100 * S
	return [3]float64{L, C, H}
}

// LchToHsluv converts LCh to HSLuv.
func LchToHsluv(tuple [3]float64) [3]float64 {
	L, C, H := tuple[0], tuple[1], tuple[2]
	if L > 99.9999999 {
		return [3]float64{H, 0, 100}
	}
	if L < 0.00000001 {
		return [3]float64{H, 0, 0}
	}
	S := C / maxChromaForLH(L, H) * 100
	return [3]float64{H, S, L}
}

// HpluvToLch converts HPLuv to LCh.
func HpluvToLch(tuple [3]float64) [3]float64 {
	H, S, L := tuple[0], tuple[1], tuple[2]
	if L > 99.9999999 {
		return [3]float64{100, 0, H}
	}
	if L < 0.00000001 {
		return [3]float64{0, 0, H}
	}
	C := maxSafeChromaForL(L) / 100 * S
	return [3]float64{L, C, H}
}

// LchToHpluv converts LCh to HPLuv.
func LchToHpluv(tuple [3]float64) [3]float64 {
	L, C, H := tuple[0], tuple[1], tuple[2]
	if L > 99.9999999 {
		return [3]float64{H, 0, 100}
	}
	if L < 0.00000001 {
		return [3]float64{H, 0, 0}
	}
	S := C / maxSafeChromaForL(L) * 100
	return [3]float64{H, S, L}
}

// RgbToHex formats RGB as "#rrggbb" in lower case. It panics with an
// IllegalArgumentException RuntimeError for a channel out of range.
func RgbToHex(tuple [3]float64) string {
	prepared := rgbPrepare(tuple)
	return fmt.Sprintf("#%02x%02x%02x", prepared[0], prepared[1], prepared[2])
}

// HexToRgb parses "#rrggbb". As upstream's Integer.parseInt, it panics with a
// NumberFormatException RuntimeError for a pair that is not hexadecimal, and
// a pair may start with a sign. A string shorter than 7 bytes panics.
func HexToRgb(hex string) [3]float64 {
	parse := func(s string) float64 {
		v, err := strconv.ParseInt(s, 16, 32)
		if err != nil {
			throw(numberFormat, `For input string: "`+s+`" under radix 16`)
		}
		return float64(v) / 255.0
	}
	return [3]float64{parse(hex[1:3]), parse(hex[3:5]), parse(hex[5:7])}
}

// LchToRgb converts LCh to RGB.
func LchToRgb(tuple [3]float64) [3]float64 { return XyzToRgb(LuvToXyz(LchToLuv(tuple))) }

// RgbToLch converts RGB to LCh.
func RgbToLch(tuple [3]float64) [3]float64 { return LuvToLch(XyzToLuv(RgbToXyz(tuple))) }

// HsluvToRgb converts HSLuv to RGB.
func HsluvToRgb(tuple [3]float64) [3]float64 { return LchToRgb(HsluvToLch(tuple)) }

// RgbToHsluv converts RGB to HSLuv.
func RgbToHsluv(tuple [3]float64) [3]float64 { return LchToHsluv(RgbToLch(tuple)) }

// HpluvToRgb converts HPLuv to RGB.
func HpluvToRgb(tuple [3]float64) [3]float64 { return LchToRgb(HpluvToLch(tuple)) }

// RgbToHpluv converts RGB to HPLuv.
func RgbToHpluv(tuple [3]float64) [3]float64 { return LchToHpluv(RgbToLch(tuple)) }

// HsluvToHex converts HSLuv to "#rrggbb".
func HsluvToHex(tuple [3]float64) string { return RgbToHex(HsluvToRgb(tuple)) }

// HpluvToHex converts HPLuv to "#rrggbb".
func HpluvToHex(tuple [3]float64) string { return RgbToHex(HpluvToRgb(tuple)) }

// HexToHsluv converts "#rrggbb" to HSLuv.
func HexToHsluv(s string) [3]float64 { return RgbToHsluv(HexToRgb(s)) }

// HexToHpluv converts "#rrggbb" to HPLuv.
func HexToHpluv(s string) [3]float64 { return RgbToHpluv(HexToRgb(s)) }
