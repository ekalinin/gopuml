// SPDX-License-Identifier: MIT
// Provenance: original

package color

import "testing"

// The expected values below were checked against upstream's classes run on
// the reference JVM; the derivations are given where they are short.

func TestMappers(t *testing.T) {
	mappers := []struct {
		name string
		m    *ColorMapper
	}{
		{"identity", IdentityMapper},
		{"dark mode", DarkModeMapper},
		{"lightness inverse", LigthnessInverseMapper},
		{"monochrome", MonochromeMapper},
		{"monochrome reverse", MonochromeReverseMapper},
		{"reverse RGB", ReverseMapper(OrderRGB)},
		{"reverse BGR", ReverseMapper(OrderBGR)},
	}
	tests := []struct {
		color string
		want  [7]string
	}{
		// Dark mode keeps colors without a dark variant. The lightness
		// inverse of black is the HSLuv white, whose red channel computes to
		// slightly below 1 and is truncated to 0xFE.
		{"white", [7]string{"#FFFFFF", "#FFFFFF", "#010101", "#FFFFFF", "#000000", "#000000", "#000000"}},
		{"black", [7]string{"#000000", "#000000", "#FEFFFF", "#000000", "#FFFFFF", "#FFFFFF", "#FFFFFF"}},
		// LightBlue (173, 216, 230): gray scale 204 = 0xCC, reversed 0x33;
		// RGB negative (82, 39, 25); in BGR order (230, 216, 173) whose
		// negative is (25, 39, 82).
		{"LightBlue", [7]string{"#ADD8E6", "#ADD8E6", "#3C5259", "#CCCCCC", "#333333", "#522719", "#192752"}},
		{"red", [7]string{"#FF0000", "#FF0000", "#E00000", "#4C4C4C", "#B3B3B3", "#00FFFF", "#FFFF00"}},
		{"#FEFECE", [7]string{"#FEFECE", "#FEFECE", "#757547", "#F8F8F8", "#070707", "#010131", "#310101"}},
		{"#181818", [7]string{"#181818", "#181818", "#E7E7E7", "#181818", "#E7E7E7", "#E7E7E7", "#E7E7E7"}},
		// Transparent colors are never mapped.
		{"transparent", [7]string{"#00000000", "#00000000", "#00000000", "#00000000", "#00000000", "#00000000", "#00000000"}},
	}
	for _, tt := range tests {
		c := Instance().GetColorOrNull(tt.color)
		for i, m := range mappers {
			if got := c.ToSvg(m.m); got != tt.want[i] {
				t.Errorf("%s with %s = %s, want %s", tt.color, m.name, got, tt.want[i])
			}
		}
	}
	// ToRGB ignores alpha and does not special-case transparent colors.
	if got := Instance().GetColorOrNull("#80FF0000").ToRGB(DarkModeMapper); got != "#80FF00" {
		t.Errorf("ToRGB = %s", got)
	}
}

func TestDarkVariant(t *testing.T) {
	white := Instance().GetColorOrNull("white")
	withDark := white.WithDark(Black)
	if got := withDark.ToSvg(DarkModeMapper); got != "#000000" {
		t.Errorf("dark mode of white with a dark variant = %s", got)
	}
	if got := withDark.ToSvg(IdentityMapper); got != "#FFFFFF" {
		t.Errorf("light mode = %s", got)
	}
	if got := withDark.String(); got != "WITHDARK [r=255,g=255,b=255,a=255] α=255" {
		t.Errorf("String = %s", got)
	}
	// Equality ignores the dark variant.
	if !withDark.Equals(white) || !white.Equals(withDark) {
		t.Error("dark variants must not change equality")
	}
	if got := Transparent().WithDark(Black).ToSvg(DarkModeMapper); got != "#00000000" {
		t.Errorf("transparent = %s", got)
	}
}

func TestEquality(t *testing.T) {
	set := Instance()
	if !Red.Equals(set.GetColorOrNull("red")) || Red.Equals(set.GetColorOrNull("red-blue")) {
		t.Error("simple colors are equal by value")
	}
	if !Transparent().Equals(set.GetColorOrNull("#00000000")) {
		t.Error("transparent equals #00000000")
	}
	if set.GetColorOrNull("#00000000").Equals(set.GetColorOrNull("#FF000000")) {
		t.Error("the channels of transparent colors count")
	}
	g := set.GetColorOrNull("red-blue")
	if !g.Equals(g) || g.Equals(set.GetColorOrNull("red-blue")) {
		t.Error("gradients are only equal to themselves")
	}
	if got := Red.(*HColorSimple).HashCode(); got != -65536 {
		t.Errorf("HashCode = %d", got)
	}
}

func TestOtherKinds(t *testing.T) {
	middle := Middle(Red, Blue)
	// (255+0)/2 = 127 and (0+255)/2 = 127.
	if got := middle.ToSvg(IdentityMapper); got != "#7F007F" {
		t.Errorf("middle = %s", got)
	}
	if middle.AsString() != "?HColorMiddle" || !middle.IsDark() || middle.Darken(10) != middle {
		t.Error("middle defaults")
	}
	tr := TransparentWith(WithFillOpacity)
	if tr.TransparentFillBehavior() != WithFillOpacity || tr.String() != "[r=0,g=0,b=0,a=0] α=0 transparent WITH_FILL_OPACITY" {
		t.Errorf("transparent with fill opacity: %s", tr)
	}
	if None() != Transparent() || Transparent().AsString() != "transparent" {
		t.Error("none")
	}
	assertPanics(t, "java.lang.UnsupportedOperationException", func() { middle.Opposite() })
	assertPanics(t, "java.lang.NullPointerException", func() { Gradient(nil, Red, '-') })
}

func TestLinearGradient(t *testing.T) {
	g := NewHColorLinearGradient(0, 0, 1, 1, false, SpreadPad, []Stop{
		NewStop(0.5, Red, 1), NewStop(-1, Blue, 0.5), NewStop(0.5, Green, 1), NewStop(2, White, 1),
		NewStop(1, Black, 1),
	})
	// Sorted by offset (stable), clamped, ties nudged by 0.000001 unless that
	// passes 1.
	want := []struct {
		offset float64
		color  string
	}{{0, "#0000FF"}, {0.5, "#FF0000"}, {0.5 + 0.000001, "#00FF00"}, {1, "#000000"}, {1, "#FFFFFF"}}
	stops := g.Stops()
	if len(stops) != len(want) {
		t.Fatalf("got %d stops", len(stops))
	}
	for i, s := range stops {
		if s.Offset() != want[i].offset || s.Color().AsString() != want[i].color {
			t.Errorf("stop %d = %v %s, want %v %s", i, s.Offset(), s.Color().AsString(), want[i].offset, want[i].color)
		}
	}
	if g.ToSvg(IdentityMapper) != "#0000FF" || !g.IsDark() || g.Opposite().AsString() != "#FFFFFF" ||
		g.AsString() != "?HColorLinearGradient" || NoGradient(g).AsString() != "#0000FF" {
		t.Error("the first stop stands for the gradient")
	}
	if g.SpreadMethod().String() != "PAD" || g.IsUserSpaceOnUse() || g.X2() != 1 {
		t.Error("accessors")
	}
}
