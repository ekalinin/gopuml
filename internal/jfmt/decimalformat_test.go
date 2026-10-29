// SPDX-License-Identifier: MIT
// Provenance: original

package jfmt

import (
	"math"
	"testing"
)

// The expected values come from the JDK (java.text.DecimalFormat with the
// symbols of Locale.US).

func TestDecimalFormat(t *testing.T) {
	upTo2 := &DecimalFormat{GroupingSeparator: ',', DecimalSeparator: '.', MaximumFractionDigits: 2}
	exactly2 := &DecimalFormat{GroupingSeparator: ',', DecimalSeparator: '.', MinimumFractionDigits: 2, MaximumFractionDigits: 2}
	upTo20 := &DecimalFormat{GroupingSeparator: ',', DecimalSeparator: '.', MaximumFractionDigits: 20}
	grouped := &DecimalFormat{GroupingUsed: true, GroupingSeparator: ' ', DecimalSeparator: ',', MaximumFractionDigits: 2}
	signed := &DecimalFormat{PositivePrefix: "+", GroupingSeparator: ',', DecimalSeparator: '.'}
	tests := []struct {
		f    *DecimalFormat
		x    float64
		want string
	}{
		{upTo2, 2.675, "2.67"},
		{upTo2, 0.125, "0.12"},
		{upTo2, 0.135, "0.14"},
		{upTo2, 1.005, "1"},
		{upTo2, -0.001, "-0"},
		{upTo2, 5.0, "5"},
		{upTo2, 2.5, "2.5"},
		{exactly2, 5, "5.00"},
		{exactly2, 0.005, "0.01"},
		{exactly2, 0.015, "0.01"},
		{exactly2, -1.005, "-1.00"},
		{exactly2, 1e-10, "0.00"},
		{upTo20, 0.1, "0.1"},
		{upTo20, 1e-7, "0.0000001"},
		{upTo20, 123456789012345678.0, "123456789012345680"},
		{upTo20, 5.25, "5.25"},
		{grouped, 1234567.891, "1 234 567,89"},
		{grouped, math.Copysign(0, -1), "-0"},
		{grouped, math.NaN(), "NaN"},
		{grouped, math.Inf(1), "∞"},
		{grouped, math.Inf(-1), "-∞"},
		{signed, 5, "+5"},
		{signed, -5, "-+5"},
		{NewDecimalFormat(), 7.5, "8"},
		{NewDecimalFormat(), 6.5, "6"},
	}
	for _, tt := range tests {
		if got := tt.f.Format(tt.x); got != tt.want {
			t.Errorf("%+v: Format(%v) = %q, want %q", *tt.f, tt.x, got, tt.want)
		}
	}
}

func TestParseDouble(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"1e3", 1000}, {" 2 ", 2}, {"1d", 1}, {"1F", 1}, {"0x1p3", 8}, {"0x1.8p1", 3},
		{"+5", 5}, {".5", 0.5}, {"5.", 5}, {"\t3\n", 3}, {"-2.5", -2.5},
		{"Infinity", math.Inf(1)}, {"-Infinity", math.Inf(-1)}, {"1e400", math.Inf(1)},
	}
	for _, tt := range tests {
		got, err := ParseDouble(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ParseDouble(%q) = %v, %v, want %v", tt.in, got, err, tt.want)
		}
	}
	for _, in := range []string{"NaN", "+NaN"} {
		if got, err := ParseDouble(in); err != nil || !math.IsNaN(got) {
			t.Errorf("ParseDouble(%q) = %v, %v, want NaN", in, got, err)
		}
	}
	invalid := []struct{ in, msg string }{
		{"inf", `For input string: "inf"`},
		{"1_0", `For input string: "1_0"`},
		{"abc", `For input string: "abc"`},
		{"1e", `For input string: "1e"`},
		{"0x1", `For input string: "0x1"`},
		{"Infinityx", `For input string: "Infinityx"`},
		{"", "empty String"},
		{"  ", "empty String"},
	}
	for _, tt := range invalid {
		if _, err := ParseDouble(tt.in); err == nil || err.Error() != tt.msg {
			t.Errorf("ParseDouble(%q) error %v, want %s", tt.in, err, tt.msg)
		}
	}
}
