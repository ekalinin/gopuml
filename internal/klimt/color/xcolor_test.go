// SPDX-License-Identifier: MIT
// Provenance: original

package color

import "testing"

func TestXColor(t *testing.T) {
	// Channels keep their low 8 bits.
	c := NewXColorAlpha(300, -1, 16, 128)
	if c != NewXColorAlpha(44, 255, 16, 128) {
		t.Errorf("got %v", c)
	}
	if got := c.String(); got != "[r=44,g=255,b=16,a=128]" {
		t.Errorf("String = %s", got)
	}
	// The Java int is negative when alpha is 128 or more.
	if got := c.RGB(); got != -2144534768 {
		t.Errorf("RGB = %d", got)
	}
	if got := c.ToSvg(); got != "#2CFF1080" {
		t.Errorf("ToSvg = %s", got)
	}
	if got := ToHexRGBColor(c.RGB()); got != "#2CFF10" {
		t.Errorf("ToHexRGBColor = %s", got)
	}
	if got := XColorFrom(0xFF123456).RGB(); got != -15584170 {
		t.Errorf("XColorFrom ignores the high byte: %d", got)
	}
	if got := NewXColorAlpha(1, 2, 3, 0).ToSvg(); got != "#00000000" {
		t.Errorf("transparent = %s", got)
	}
	if got := XColorFrom(0xADD8E6).ToSvg(); got != "#ADD8E6" {
		t.Errorf("LightBlue = %s", got)
	}
	// Math.round(0.002f * 255) = Math.round(0.51f) = 1.
	if got := XColorFromFloat(0.5, 0.25, 1, 0.002); got != NewXColorAlpha(128, 64, 255, 1) {
		t.Errorf("XColorFromFloat = %v", got)
	}
	assertPanics(t, "java.lang.IllegalArgumentException", func() { XColorFromFloat(1.5, 0, 0, 1) })
	if got := XColorFrom(0x336699).RGBColorComponents(); got != [3]float32{0.2, 0.4, 0.6} {
		t.Errorf("RGBColorComponents = %v", got)
	}
}

func TestSvgForms(t *testing.T) {
	// HColor.toSvg writes the full "#RRGGBB"; red becomes "#F00" only later,
	// in the SVG writer.
	for in, want := range map[string]string{
		"red":         "#FF0000",
		"LightBlue":   "#ADD8E6",
		"#abc":        "#AABBCC",
		"#FF000080":   "#FF000080",
		"transparent": "#00000000",
		"red-blue":    "#FF0000",
	} {
		if got := Instance().GetColorOrNull(in).ToSvg(IdentityMapper); got != want {
			t.Errorf("ToSvg(%s) = %s, want %s", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"red":       "#FF0000",
		"#FF000080": "#FF0000",
		"red-blue":  "#FF0000",
	} {
		if got := Instance().GetColorOrNull(in).ToRGB(IdentityMapper); got != want {
			t.Errorf("ToRGB(%s) = %s, want %s", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"red":       "[r=255,g=0,b=0,a=255] α=255",
		"#FF000080": "[r=255,g=0,b=0,a=128] α=128",
		"#80FF0000": "[r=128,g=255,b=0,a=0] α=0 transparent",
	} {
		if got := Instance().GetColorOrNull(in).String(); got != want {
			t.Errorf("String(%s) = %s, want %s", in, got, want)
		}
	}
}
