// SPDX-License-Identifier: MIT
// Provenance: original

package jfmt

import (
	"math"
	"testing"
)

// Expected values printed by Java 21 (the reference image).
func TestFormatFixed(t *testing.T) {
	tests := []struct {
		x         float64
		two, four string
	}{
		{0.125, "0.13", "0.1250"},
		{1.005, "1.01", "1.0050"},
		{2.675, "2.68", "2.6750"},
		{0.00005, "0.00", "0.0001"},
		{1.00005, "1.00", "1.0001"},
		{0.03125, "0.03", "0.0313"},
		{12.03125, "12.03", "12.0313"},
		{3.14159265, "3.14", "3.1416"},
		{0.000049999, "0.00", "0.0000"},
		{1e-5, "0.00", "0.0000"},
		{123456.78905, "123456.79", "123456.7891"},
		{0.1 + 0.2, "0.30", "0.3000"},
		{9.99995, "10.00", "10.0000"},
		{-0.03125, "-0.03", "-0.0313"},
		{-2.5e-5, "-0.00", "-0.0000"},
		{0, "0.00", "0.0000"},
		{math.Copysign(0, -1), "-0.00", "-0.0000"},
		{1e21, "1000000000000000000000.00", "1000000000000000000000.0000"},
		{999.995, "1000.00", "999.9950"},
	}
	for _, tt := range tests {
		if got := FormatFixed(tt.x, 2); got != tt.two {
			t.Errorf("FormatFixed(%v, 2) = %q, want %q", tt.x, got, tt.two)
		}
		if got := FormatFixed(tt.x, 4); got != tt.four {
			t.Errorf("FormatFixed(%v, 4) = %q, want %q", tt.x, got, tt.four)
		}
	}
	if FormatFixed(math.NaN(), 2) != "NaN" || FormatFixed(math.Inf(-1), 2) != "-Infinity" || FormatFixed(2.5, 0) != "3" {
		t.Error("special values")
	}
}
