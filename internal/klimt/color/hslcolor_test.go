// SPDX-License-Identifier: MIT
// Provenance: original

package color

import (
	"math"
	"testing"
)

// The values of the preprocessor color functions below are derived by hand
// from upstream's float arithmetic, as explained next to each case, and were
// checked on the reference JVM.

func TestDarkenLighten(t *testing.T) {
	red := Instance().GetColorOrNull("red")
	tests := []struct {
		name string
		got  HColor
		want string
	}{
		// Red is h=0, s=100, l=50. Darken 20: l = 50 - 50*0.2 = 40, so
		// q = 0.4f*(1+1) = 0.8f, p = 2*0.4f - q = 0; red = q, giving
		// round(0.8f*255) = 204 = 0xCC.
		{"darken(red, 20)", red.Darken(20), "#CC0000"},
		// Lighten 20: l = 60, q = (0.6f+1) - 0.6f = 1, p = 2*0.6f - 1 =
		// 0.20000005f; green and blue are p, round(51.0000121) = 51 = 0x33.
		{"lighten(red, 20)", red.Lighten(20), "#FF3333"},
		// Darken -30 is lighten 30: l = 65, p = 2*0.65f - 1 = 0.29999995f.
		// Green is p: round(76.4999878) = 76 = 0x4C. Blue is p plus
		// (q-p)*6*(2/3f - 0.6666666f), about 2.5e-7: round(76.50005) = 77 =
		// 0x4D. Float arithmetic matters: in exact arithmetic both are 0x4D.
		{"darken(red, -30)", red.Darken(-30), "#FF4C4D"},
		{"darken(red, 10)", red.Darken(10), "#E60000"},
		{"lighten(red, 10)", red.Lighten(10), "#FF1A1A"},
		{"lighten(red, 150)", red.Lighten(150), "#FFFFFF"}, // l clamped to 100
		{"darken(red, 150)", red.Darken(150), "#000000"},   // l clamped to 0
		// Gray #808080: s = 0 and l = 50.196083; darken 50 gives
		// l = 25.098041, and every channel is round(0.25098041f*255) = 64.
		{"darken(#808080, 50)", Instance().GetColorOrNull("#808080").Darken(50), "#404040"},
		// The result is always opaque: the alpha is dropped.
		{"lighten(#FF000080, 0)", Instance().GetColorOrNull("#FF000080").Lighten(0), "#FF0000"},
		// Non-simple colors are returned as is.
		{"darken(red-blue, 10)", Instance().GetColorOrNull("red-blue").Darken(10), "?HColorGradient"},
	}
	for _, tt := range tests {
		if got := tt.got.AsString(); got != tt.want {
			t.Errorf("%s = %s, want %s", tt.name, got, tt.want)
		}
	}
}

// TestNoFusedMultiplyAdd uses values observed on the reference JVM that come
// out one unit off when the compiler fuses a product and an addition into an
// FMA instruction, as Go does on arm64 without the explicit conversions.
func TestNoFusedMultiplyAdd(t *testing.T) {
	tests := []struct {
		name string
		got  HColor
		want string
	}{
		{"lighten(#112222, 25)", Instance().GetColorOrNull("#112222").Lighten(25), "#152B2A"},
		{"lighten(#11EE99, 25)", Instance().GetColorOrNull("#11EE99").Lighten(25), "#4DF2B2"},
		{"lighten(#666699, 50)", Instance().GetColorOrNull("#666699").Lighten(50), "#B2B2CC"},
		{"lighten(#553355, -30)", Instance().GetColorOrNull("#553355").Lighten(-30), "#3B243C"},
		{"darken(#227733, 50)", Instance().GetColorOrNull("#227733").Darken(50), "#113C19"},
		{"darken(#335566, 10)", Instance().GetColorOrNull("#335566").Darken(10), "#2E4D5C"},
		{"darken(#3399CC, 50)", Instance().GetColorOrNull("#3399CC").Darken(50), "#194D66"},
		{"hsl(55, 102, 80)", Simple(NewHSLColorHSL(55, 102, 80).RGB()), "#FFF699"},
	}
	for _, tt := range tests {
		if got := tt.got.AsString(); got != tt.want {
			t.Errorf("%s = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestIsDarkReverse(t *testing.T) {
	// Gray scale of #808080: 128*(299+587+114)/1000 = 128, not below 128.
	if Instance().GetColorOrNull("#808080").IsDark() {
		t.Error("#808080 is light")
	}
	if !Instance().GetColorOrNull("#7F7F7F").IsDark() {
		t.Error("#7F7F7F is dark")
	}
	// LightBlue: (173*299 + 216*587 + 230*114)/1000 = 204.
	if got := GrayScale(XColorFrom(0xADD8E6)); got != 204 {
		t.Errorf("GrayScale(LightBlue) = %d", got)
	}
	if got := GrayScaleFromRGB(XColorFrom(0xADD8E6).RGB()); got != 204 {
		t.Errorf("GrayScaleFromRGB(LightBlue) = %d", got)
	}
	// %reverse_color: 255 minus each channel.
	if got := Instance().GetColorOrNull("#123456").Reverse().AsString(); got != "#EDCBA9" {
		t.Errorf("reverse = %s", got)
	}
	if got := Instance().GetColorOrNull("#808080").Opposite().AsString(); got != "#000000" {
		t.Errorf("opposite of #808080 = %s", got)
	}
	if got := Instance().GetColorOrNull("#7F7F7F").Opposite().AsString(); got != "#FFFFFF" {
		t.Errorf("opposite of #7F7F7F = %s", got)
	}
}

func TestHSLColor(t *testing.T) {
	tests := []struct {
		name string
		got  XColor
		want string
	}{
		// %hsl_color(120, 100, 50): h = 0.33333334f, q = 1, p = 0. Red uses
		// h+1/3 = 0.6666667f, where 3h = 2.0000002f is not below 2, so p.
		{"hsl(120,100,50)", NewHSLColorHSL(120, 100, 50).RGB(), "#00FF00"},
		{"hsl(0,100,50)", NewHSLColorHSL(0, 100, 50).RGB(), "#FF0000"},
		{"hsl(240,100,50)", NewHSLColorHSL(240, 100, 50).RGB(), "#0000FF"},
		// Hue is taken modulo 360, negative hues wrap through HueToRGB.
		{"hsl(-240,100,50)", NewHSLColorHSL(-240, 100, 50).RGB(), "#00FF00"},
		// Saturation and luminance are clamped.
		{"hsl(0,200,-5)", NewHSLColorHSL(0, 200, -5).RGB(), "#000000"},
		// %hsl_color(0, 100, 50, 50): alpha (float)(50/100.0) = 0.5 and
		// round(127.5) = 128.
		{"hsl(0,100,50,0.5)", NewHSLColorHSLA(0, 100, 50, 0.5).RGB(), "#FF000080"},
	}
	for _, tt := range tests {
		if got := tt.got.ToSvg(); got != tt.want {
			t.Errorf("%s = %s, want %s", tt.name, got, tt.want)
		}
	}
	if got := Simple(NewHSLColorHSLA(0, 100, 50, 0.5).RGB()).AsString(); got != "#80ff0000" {
		t.Errorf("asString = %s", got)
	}

	// #336699 is h=210, s=50.000008, l=40 in float.
	hsl := NewHSLColor(XColorFrom(0x336699))
	if got := hsl.String(); got != "HSLColor[h=210.0,s=50.000008,l=40.0,alpha=1.0]" {
		t.Errorf("String = %s", got)
	}
	for _, c := range []struct {
		name string
		got  XColor
		want string
	}{
		{"AdjustHue(30)", hsl.AdjustHue(30), "#996633"},
		{"AdjustLuminance(80)", hsl.AdjustLuminance(80), "#B3CCE6"},
		{"AdjustSaturation(10)", hsl.AdjustSaturation(10), "#5C6670"},
		{"AdjustShade(30)", hsl.AdjustShade(30), "#24476B"},
		{"AdjustTone(30)", hsl.AdjustTone(30), "#4785C2"},
		{"Complementary", hsl.Complementary(), "#996633"},
	} {
		if got := c.got.ToSvg(); got != c.want {
			t.Errorf("%s = %s, want %s", c.name, got, c.want)
		}
	}
	if got := HSLFromRGB(XColorFrom(0x0000FF)); got != [3]float32{240, 100, 50} {
		t.Errorf("HSLFromRGB(blue) = %v", got)
	}
	half := NewHSLColor(NewXColorAlpha(10, 20, 30, 128))
	if got := half.String(); got != "HSLColor[h=210.0,s=49.999992,l=7.8431377,alpha=0.5019608]" {
		t.Errorf("String = %s", got)
	}
	if got := half.AdjustTone(10).ToSvg(); got != "#0B162180" {
		t.Errorf("AdjustTone keeps alpha: %s", got)
	}
}

func TestUnlinear(t *testing.T) {
	// 50% is cubed to 0.125: the hue moves from 0 to 30, an orange.
	if got := Unlinear(Red, Blue, 50).AsString(); got != "#FF8000" {
		t.Errorf("Unlinear(50) = %s", got)
	}
	if Unlinear(Red, Blue, 0) != Red || Unlinear(Red, Blue, 100) != Blue {
		t.Error("Unlinear at 0 and 100 returns the colors")
	}
	if got := Unlinear(Red, Blue, 150).AsString(); got != "#0000FF" {
		t.Errorf("Unlinear(150) = %s", got)
	}
	g := Instance().GetColorOrNull("red-blue")
	if Unlinear(g, Blue, 50) != g {
		t.Error("Unlinear of a gradient returns it")
	}
}

func TestHsluv(t *testing.T) {
	// Black: linear RGB 0 gives Y = 0, so L = 0 and the HSLuv is (0, 0, 0).
	if got := RgbToHsluv([3]float64{0, 0, 0}); got != [3]float64{0, 0, 0} {
		t.Errorf("RgbToHsluv(black) = %v", got)
	}
	// L = 0 converts back to black whatever the hue and saturation.
	if got := HsluvToRgb([3]float64{123, 45, 0}); got != [3]float64{0, 0, 0} {
		t.Errorf("HsluvToRgb(L=0) = %v", got)
	}
	// %reverse_hsluv_color of black: l = (0+50) % 100 = 50, unchanged by
	// l += 0.25*(50-l). With s = 0 the result is the gray of L = 50:
	// Y = ((50+16)/116)^3 = 0.184187, and each channel is
	// 1.055*Y^(1/2.4) - 0.055 = 0.46633, truncated to (int)(118.91) = 118.
	if got := Instance().GetColorOrNull("black").ReverseHsluv().AsString(); got != "#767676" {
		t.Errorf("reverse_hsluv(black) = %s", got)
	}
	for _, c := range []struct{ in, want string }{
		{"white", "#767676"},
		{"red", "#530000"},
		{"LightBlue", "#445C64"},
		{"#A80036", "#FEA2AE"},
		{"#FEFECE", "#767648"},
		{"transparent", "#767676"}, // the channels of transparent are black
	} {
		if got := Instance().GetColorOrNull(c.in).ReverseHsluv().AsString(); got != c.want {
			t.Errorf("reverse_hsluv(%s) = %s, want %s", c.in, got, c.want)
		}
	}
	hsluv := HexToHsluv("#336699")
	want := [3]float64{246.942440261812, 78.45117445648185, 42.009163494482195}
	for i := range 3 {
		if math.Abs(hsluv[i]-want[i]) > 1e-9 {
			t.Errorf("HexToHsluv(#336699) = %v, want %v", hsluv, want)
			break
		}
	}
	if got := HsluvToHex([3]float64{250, 60, 40}); got != "#425f8b" {
		t.Errorf("HsluvToHex = %s", got)
	}
	if got := HpluvToHex([3]float64{250, 60, 40}); got != "#545f73" {
		t.Errorf("HpluvToHex = %s", got)
	}
	assertPanics(t, "java.lang.IllegalArgumentException: Illegal rgb value: 1.2",
		func() { RgbToHex([3]float64{1.2, 0, 0}) })
	assertPanics(t, `java.lang.NumberFormatException: For input string: "zz" under radix 16`,
		func() { HexToRgb("#zz0000") })
	// Integer.parseInt accepts a sign.
	if got := HexToRgb("#-1+f00"); got != [3]float64{-1 / 255.0, 15 / 255.0, 0} {
		t.Errorf("HexToRgb(#-1+f00) = %v", got)
	}
}

func TestGrayToColor(t *testing.T) {
	// asMonochrome maps grays onto the HSLuv tints of a color.
	tint := Col527BC6.(*HColorSimple)
	if got := Instance().GetColorOrNull("black").(*HColorSimple).AsMonochromeWith(tint, 20, 200).AsString(); got != "#4B71B6" {
		t.Errorf("AsMonochromeWith(black) = %s", got)
	}
	if got := GrayToColor(0.3, XColorBlack).ToSvg(); got != "#464646" {
		t.Errorf("GrayToColor(0.3, black) = %s", got)
	}
	if got := Instance().GetColorOrNull("LightBlue").(*HColorSimple).AsMonochrome().AsString(); got != "#CCCCCC" {
		t.Errorf("AsMonochrome = %s", got)
	}
}

func TestColorHSB(t *testing.T) {
	c := NewColorHSBFrom(XColorFrom(0x7F807F))
	// Saturation 1/128 = 0.0078125 is rounded half up from its shortest
	// decimal, as Java's Formatter does.
	if got := c.String(); got != "ColorHSB[a=FF r=7F g=80 b=7F / h=0.333333 s=0.007813 b=0.501961]" {
		t.Errorf("String = %s", got)
	}
	v := uint32(0x80020180)
	c = NewColorHSB(int32(v))
	if got := c.String(); got != "ColorHSB[a=80 r=02 g=01 b=80 / h=0.667979 s=0.992188 b=0.501961]" {
		t.Errorf("String = %s", got)
	}
	if c.Alpha() != 0x80 || c.RGB() != int32(v) {
		t.Errorf("Alpha = %d, RGB = %d", c.Alpha(), c.RGB())
	}
}
