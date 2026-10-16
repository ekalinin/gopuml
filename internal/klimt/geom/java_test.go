// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestJavaDoubleString(t *testing.T) {
	tests := []struct {
		bits uint64
		want string
	}{
		{0x0000000000000000, "0.0"},
		{0x8000000000000000, "-0.0"},
		{0x3ff0000000000000, "1.0"},
		{0xbff0000000000000, "-1.0"},
		{0x4028000000000000, "12.0"},
		{0x3fb999999999999a, "0.1"},
		{0x3fe0000000000000, "0.5"},
		{0x4059000000000000, "100.0"},
		{0x4132d68700000000, "1234567.0"},
		{0x416312cfe0000000, "9999999.0"},
		{0x416312d000000000, "1.0E7"},
		{0x416312d020000000, "1.0000001E7"},
		{0x41678c29dccccccd, "1.23456789E7"},
		{0x3f50624dd2f1a9fc, "0.001"},
		{0x3f50624dd2f1a9fd, "0.0010000000000000002"},
		{0x3f50624dd2f1a9fb, "9.999999999999998E-4"},
		{0x3f1a36e2eb1c432d, "1.0E-4"},
		{0x3ee4f8b588e368f1, "1.0E-5"},
		{0x405edd2f1a9fbe77, "123.456"},
		{0x3fe5555555555555, "0.6666666666666666"},
		{0x3fd5555555555555, "0.3333333333333333"},
		{0xbdf12e0be826d695, "-2.5E-10"},
		{0x7fefffffffffffff, "1.7976931348623157E308"},
		{0xffefffffffffffff, "-1.7976931348623157E308"},
		{0x0000000000000001, "4.9E-324"},
		{0x0000000000000002, "9.9E-324"},
		{0x0000000000000003, "1.5E-323"},
		{0x0000000000000014, "9.9E-323"},
		{0x0010000000000000, "2.2250738585072014E-308"},
		{0x44b52d02c7e14af6, "1.0E23"},
		{0x44c52d02c7e14af6, "2.0E23"},
		{0x4340000000000000, "9.007199254740992E15"},
		{0x4011666666666666, "4.35"},
		{0x3fd3333333333333, "0.3"},
		{0x3fd3333333333334, "0.30000000000000004"},
		{0x444b1ae4d6e2ef50, "1.0E21"},
		{0x4480f0cf064dd592, "1.0E22"},
		{0x0000000000000007, "3.5E-323"},
		{0x4059200000000000, "100.5"},
		{0x400921fb54442d18, "3.141592653589793"},
		{0x412e848000000000, "1000000.0"},
		{0x3f689374bc6a7efa, "0.003"},
		{0x40f0000000000000, "65536.0"},
		{0x3fd0000000000000, "0.25"},
		{0x44dfe185ca57c517, "6.02214076E23"},
		{0x7fd5555555555555, "5.992310449541053E307"},
		{0x3ff0000000000001, "1.0000000000000002"},
		{0x3fefffffffffffff, "0.9999999999999999"},
		{0x416312cfffffffff, "9999999.999999998"},
		{0x3f50624dd2f1a9fd, "0.0010000000000000002"},
		{0x0008000000000000, "1.1125369292536007E-308"},
	}
	for _, tt := range tests {
		if got := javaDoubleString(fb(tt.bits)); got != tt.want {
			t.Errorf("javaDoubleString(%#016x) = %q, want %q", tt.bits, got, tt.want)
		}
	}
	for _, tt := range []struct {
		d    float64
		want string
	}{
		{math.NaN(), "NaN"},
		{math.Inf(1), "Infinity"},
		{math.Inf(-1), "-Infinity"},
		{-12, "-12.0"},
		{-0.001, "-0.001"},
		{-1e7, "-1.0E7"},
		{1e-3 / 2, "5.0E-4"},
	} {
		if got := javaDoubleString(tt.d); got != tt.want {
			t.Errorf("javaDoubleString(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestJavaD2I(t *testing.T) {
	tests := []struct {
		d    float64
		want int32
	}{
		{0, 0},
		{2.9, 2},
		{-2.9, -2},
		{math.NaN(), 0},
		{3e9, math.MaxInt32},
		{-3e9, math.MinInt32},
		{math.Inf(1), math.MaxInt32},
		{math.Inf(-1), math.MinInt32},
		{2147483646.7, 2147483646},
		{-2147483648.5, math.MinInt32},
	}
	for _, tt := range tests {
		if got := javaD2I(tt.d); got != tt.want {
			t.Errorf("javaD2I(%v) = %d, want %d", tt.d, got, tt.want)
		}
	}
}

func TestJavaEqualsIgnoreCase(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"RIGHT", "right", true},
		{"RIGHT", "r\u0131ght", true}, // dotless i upper-cases to I
		{"RIGHT", "R\u0130ght", true}, // I with dot lower-cases to i
		{"LINK", "lin\u212a", true},   // Kelvin sign lower-cases to k
		{"LEFT", "LEFT ", false},
		{"LEFT", "LEF", false},
		{"", "", true},
		{"S", "\u017f", true}, // long s upper-cases to S
	}
	for _, tt := range tests {
		if got := javaEqualsIgnoreCase(tt.a, tt.b); got != tt.want {
			t.Errorf("javaEqualsIgnoreCase(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
