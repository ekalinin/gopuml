// SPDX-License-Identifier: MIT
// Provenance: original

package color

import (
	"math"
	"testing"
)

// The expected strings are Java 21's.

func TestJavaRound(t *testing.T) {
	for x, want := range map[float64]int64{
		0.49999999999999994: 0, // floor(x + 0.5) would give 1
		0.5:                 1,
		-0.5:                0,
		-0.5000000000000001: -1,
		2.5:                 3,
		-2.5:                -2,
		math.NaN():          0,
		1e300:               math.MaxInt64,
	} {
		if got := javaRound(x); got != want {
			t.Errorf("javaRound(%v) = %d, want %d", x, got, want)
		}
	}
	if javaRoundF(127.5) != 128 || javaRoundF(-0.5) != 0 || javaD2I(-1.9) != -1 || javaD2I(math.Inf(1)) != math.MaxInt32 {
		t.Error("float conversions")
	}
}

func TestJavaFloatingToString(t *testing.T) {
	floats := map[float32]string{
		0.1: "0.1", 1: "1.0", 100: "100.0", 1e7: "1.0E7", 1.2345678e7: "1.2345678E7", 0.001: "0.001",
		0.0009: "9.0E-4", 1e-5: "1.0E-5", 123456.7: "123456.7", -2.5: "-2.5", 0: "0.0", 1.0 / 3: "0.33333334",
	}
	for f, want := range floats {
		if got := javaFloatToString(f); got != want {
			t.Errorf("Float.toString(%v) = %s, want %s", f, got, want)
		}
	}
	doubles := map[float64]string{
		0.1: "0.1", 1e7: "1.0E7", 1e21: "1.0E21", 1.0 / 3: "0.3333333333333333", 0.0078125: "0.0078125",
		2.5e-7: "2.5E-7", -1e-4: "-1.0E-4",
	}
	for d, want := range doubles {
		if got := javaDoubleToString(d); got != want {
			t.Errorf("Double.toString(%v) = %s, want %s", d, got, want)
		}
	}
	if got := javaFloatToString(float32(math.Copysign(0, -1))); got != "-0.0" {
		t.Errorf("-0.0 = %s", got)
	}
}

func TestJavaFormatFixed(t *testing.T) {
	for v, want := range map[float64]string{
		0.0078125:                   "0.007813", // half up on the shortest decimal
		0.15:                        "0.150000",
		123.4565:                    "123.456500",
		2.5e-7:                      "0.000000",
		1e21:                        "1000000000000000000000.000000",
		float64(float32(123456.7)):  "123456.703125",
		float64(float32(1.0 / 3.0)): "0.333333",
		0.9999999:                   "1.000000",
		math.Copysign(0, -1):        "-0.000000",
	} {
		if got := javaFormatFixed(v, 6); got != want {
			t.Errorf("format(%v) = %s, want %s", v, got, want)
		}
	}
}

func TestJavaStrings(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"TRANSPARENT", "transparent", true},
		{"tran\u017Fparent", "transparent", true}, // long s
		{"bac\u212Aground", "background", true},   // Kelvin sign
		{"automat\u0131c", "automatic", true},     // dotless i
		{"automat\u0130c", "automatic", true},     // dotted capital I
		{"automati", "automatic", false},
	} {
		if got := javaEqualsIgnoreCase(c.a, c.b); got != c.want {
			t.Errorf("equalsIgnoreCase(%q, %q) = %v", c.a, c.b, got)
		}
	}
	if got := javaLowerCase("WH\u0130TE"); got != "whi\u0307te" {
		t.Errorf("toLowerCase = %q", got)
	}
	if javaDoubleCompare(math.Copysign(0, -1), 0) != -1 || javaDoubleCompare(math.NaN(), math.Inf(1)) != 1 ||
		javaDoubleCompare(math.NaN(), math.NaN()) != 0 {
		t.Error("Double.compare")
	}
}
