// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"math"
	"strconv"
	"testing"
)

func TestJavaDoubleToString(t *testing.T) {
	for _, tc := range doubleCases {
		v := math.Float64frombits(tc.bits)
		if got := javaFloatToString(v, 64); got != tc.java {
			t.Errorf("Double.toString(%v) = %q; want %q", v, got, tc.java)
		}
		jv, err := ValueOfDouble(v)
		if err != nil {
			t.Fatalf("ValueOfDouble(%v): %v", v, err)
		}
		if got := jv.String(); got != tc.json {
			t.Errorf("ValueOfDouble(%v) = %q; want %q", v, got, tc.json)
		}
	}
}

func TestJavaFloatToString(t *testing.T) {
	for _, tc := range floatCases {
		v := math.Float32frombits(tc.bits)
		if got := javaFloatToString(float64(v), 32); got != tc.java {
			t.Errorf("Float.toString(%v) = %q; want %q", v, got, tc.java)
		}
		jv, err := ValueOfFloat(v)
		if err != nil {
			t.Fatalf("ValueOfFloat(%v): %v", v, err)
		}
		if got := jv.String(); got != tc.json {
			t.Errorf("ValueOfFloat(%v) = %q; want %q", v, got, tc.json)
		}
	}
}

func TestValueOfDoubleReadable(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0.0, "0"},
		{math.Copysign(0, -1), "-0"},
		{1e7, "1.0E7"},
		{9999999.0, "9999999"},
		{0.001, "0.001"},
		{0.0009999, "9.999E-4"},
		{123456789012.0, "1.23456789012E11"},
		{1.0e-5, "1.0E-5"},
		{1.234e-5, "1.234E-5"},
		{-1.5e10, "-1.5E10"},
		{100.0, "100"},
		{100.25, "100.25"},
		{12345.678, "12345.678"},
		{math.SmallestNonzeroFloat64, "4.9E-324"},
		{math.MaxFloat64, "1.7976931348623157E308"},
		{1e23, "1.0E23"},
	}
	for _, tt := range tests {
		v, err := ValueOfDouble(tt.in)
		if err != nil {
			t.Fatalf("ValueOfDouble(%v): %v", tt.in, err)
		}
		if got := v.String(); got != tt.want {
			t.Errorf("ValueOfDouble(%v) = %q; want %q", tt.in, got, tt.want)
		}
		if !v.IsNumber() {
			t.Errorf("ValueOfDouble(%v) is not a number", tt.in)
		}
	}
}

func TestValueOfFloatReadable(t *testing.T) {
	tests := []struct {
		in   float32
		want string
	}{
		{0.1, "0.1"},
		{1.0 / 3, "0.33333334"},
		{1e7, "1.0E7"},
		{9999999.0, "9999999"},
		{16777217.0, "1.6777216E7"},
		{math.SmallestNonzeroFloat32, "1.4E-45"},
		{math.MaxFloat32, "3.4028235E38"},
	}
	for _, tt := range tests {
		v, err := ValueOfFloat(tt.in)
		if err != nil {
			t.Fatalf("ValueOfFloat(%v): %v", tt.in, err)
		}
		if got := v.String(); got != tt.want {
			t.Errorf("ValueOfFloat(%v) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestValueOfNotFinite(t *testing.T) {
	const want = "Infinite and NaN values not permitted in JSON"
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := ValueOfDouble(v); err == nil || err.Error() != want {
			t.Errorf("ValueOfDouble(%v) error = %v; want %q", v, err, want)
		}
		if _, err := ValueOfFloat(float32(v)); err == nil || err.Error() != want {
			t.Errorf("ValueOfFloat(%v) error = %v; want %q", v, err, want)
		}
	}
}

func TestValueOfOthers(t *testing.T) {
	tests := []struct {
		v    Value
		want string
	}{
		{ValueOfInt(0), "0"},
		{ValueOfInt(-5), "-5"},
		{ValueOfInt(math.MaxInt32), "2147483647"},
		{ValueOfLong(math.MinInt64), "-9223372036854775808"},
		{ValueOfString(""), `""`},
		{ValueOfString("a\"b"), `"a\"b"`},
		{ValueOfBool(true), "true"},
		{ValueOfBool(false), "false"},
		{Null, "null"},
		{True, "true"},
		{False, "false"},
	}
	for i, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("%d: String() = %q; want %q", i, got, tt.want)
		}
	}
	if ValueOfBool(true) != True || ValueOfBool(false) != False {
		t.Error("ValueOfBool does not return the True and False literals")
	}
}

func TestCutOffPointZero(t *testing.T) {
	for in, want := range map[string]string{"1.0": "1", "1.0E7": "1.0E7", "10.05": "10.05", "-0.0": "-0", "100.00": "100.00"} {
		if got := cutOffPointZero(in); got != want {
			t.Errorf("cutOffPointZero(%s) = %s; want %s", strconv.Quote(in), got, want)
		}
	}
}
