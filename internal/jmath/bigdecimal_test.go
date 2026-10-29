// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import (
	"errors"
	"math"
	"testing"
)

// The expected values come from the JDK (java.math.BigDecimal).

func TestBigDecimalOf(t *testing.T) {
	tests := []struct {
		in    string
		plain string
		scale int
	}{
		{"10.", "10", 0},
		{"1.50", "1.50", 2},
		{"007", "7", 0},
		{"-0", "0", 0},
		{"+.5", "0.5", 1},
		{"-.5", "-0.5", 1},
		{"00.10", "0.10", 2},
		{"1e5", "100000", -5},
		{"1E+2", "100", -2},
		{"1.5e-3", "0.0015", 4},
	}
	for _, tt := range tests {
		b := BigDecimalOf(tt.in)
		if got := b.ToPlainString(); got != tt.plain || b.Scale() != tt.scale {
			t.Errorf("BigDecimalOf(%q) = %s scale %d, want %s scale %d", tt.in, got, b.Scale(), tt.plain, tt.scale)
		}
	}
}

func panicOf(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	f()
	return nil
}

func TestBigDecimalOfInvalid(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1.2.3", "java.lang.NumberFormatException: Character array contains more than one decimal point."},
		{"+..", "java.lang.NumberFormatException: Character array contains more than one decimal point."},
		{"+.", "java.lang.NumberFormatException: No digits found."},
		{"-", "java.lang.NumberFormatException: No digits found."},
		{".", "java.lang.NumberFormatException: No digits found."},
		{"", "java.lang.NumberFormatException"},
		{"1e", "java.lang.NumberFormatException"},
		{"1x", `java.lang.NumberFormatException: Character x is neither a decimal digit number, decimal point, nor "e" notation exponential mark.`},
	}
	for _, tt := range tests {
		err := panicOf(func() { BigDecimalOf(tt.in) })
		var e *Exception
		if !errors.As(err, &e) || err.Error() != tt.want {
			t.Errorf("BigDecimalOf(%q) panics with %v, want %s", tt.in, err, tt.want)
		}
	}
}

func TestBigDecimalArithmetic(t *testing.T) {
	if got := BigDecimalOf("1.5").Add(BigDecimalOf("2.25")).ToPlainString(); got != "3.75" {
		t.Errorf("1.5 + 2.25 = %s", got)
	}
	if got := BigDecimalOf("12.5").Multiply(BigDecimalOfInt64(3)).ToPlainString(); got != "37.5" {
		t.Errorf("12.5 * 3 = %s", got)
	}
	if got := BigDecimalOf("1.50").Add(BigDecimalOfInt64(1)).ToPlainString(); got != "2.50" {
		t.Errorf("1.50 + 1 = %s, want the larger scale", got)
	}
	if got := BigDecimalOf("10").CompareTo(BigDecimalOf("10.0")); got != 0 {
		t.Errorf("10 compared with 10.0 = %d", got)
	}
	if got := BigDecimalOf("-1").CompareTo(BigDecimalOf("0.5")); got != -1 {
		t.Errorf("-1 compared with 0.5 = %d", got)
	}
	if got := BigDecimalOf("2.5").CompareTo(BigDecimalOfInt64(2)); got != 1 {
		t.Errorf("2.5 compared with 2 = %d", got)
	}
	if got := BigDecimalZero.Add(BigDecimalOf("0.1")).ToPlainString(); got != "0.1" {
		t.Errorf("ZERO + 0.1 = %s", got)
	}
	if BigDecimalOf("-3.00").Signum() != -1 || BigDecimalOf("0.000").Signum() != 0 || BigDecimalOf("4").Signum() != 1 {
		t.Error("Signum")
	}
}

func TestBigDecimalConversions(t *testing.T) {
	if got := BigDecimalOf("2.5").LongValue(); got != 2 {
		t.Errorf("longValue of 2.5 = %d", got)
	}
	if got := BigDecimalOf("-2.5").LongValue(); got != -2 {
		t.Errorf("longValue of -2.5 = %d", got)
	}
	if got := BigDecimalOf("-7.9").LongValue(); got != -7 {
		t.Errorf("longValue of -7.9 = %d", got)
	}
	if got := BigDecimalOf("1e30").LongValue(); got != 5076944270305263616 {
		t.Errorf("longValue of 1e30 = %d", got)
	}
	if got := BigDecimalOf("1E+2").LongValueExact(); got != 100 {
		t.Errorf("longValueExact of 1E+2 = %d", got)
	}
	if err := panicOf(func() { BigDecimalOf("2.5").LongValueExact() }); err == nil ||
		err.Error() != "java.lang.ArithmeticException: Rounding necessary" {
		t.Errorf("longValueExact of 2.5 panics with %v", err)
	}
	if err := panicOf(func() { BigDecimalOf("1e30").LongValueExact() }); err == nil ||
		err.Error() != "java.lang.ArithmeticException: Overflow" {
		t.Errorf("longValueExact of 1e30 panics with %v", err)
	}
	if got := BigDecimalOf("0.1").DoubleValue(); got != 0.1 {
		t.Errorf("doubleValue of 0.1 = %v", got)
	}
	if got := BigDecimalOf("12.5").DoubleValue(); got != 12.5 {
		t.Errorf("doubleValue of 12.5 = %v", got)
	}
}

func TestBigDecimalOfFloat64(t *testing.T) {
	tests := []struct {
		in    float64
		plain string
	}{
		{0.1, "0.1000000000000000055511151231257827021181583404541015625"},
		{2.5, "2.5"},
		{math.Copysign(0, -1), "0"},
		{1e20, "100000000000000000000"},
		{-3, "-3"},
	}
	for _, tt := range tests {
		if got := BigDecimalOfFloat64(tt.in).ToPlainString(); got != tt.plain {
			t.Errorf("BigDecimalOfFloat64(%v) = %s, want %s", tt.in, got, tt.plain)
		}
	}
	if got := BigDecimalOfFloat64(1e-3).Scale(); got != 60 {
		t.Errorf("scale of new BigDecimal(1e-3) = %d, want 60", got)
	}
	if got := BigDecimalOfFloat64(2.5).DoubleValue(); got != 2.5 {
		t.Errorf("round trip of 2.5 = %v", got)
	}
}
