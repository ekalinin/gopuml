// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

// The expected values in these tests were computed by upstream's Java classes
// (plantuml-mit 1.2026.8 on Java 25) and are written as IEEE 754 bit
// patterns, so that comparisons are bit-exact.

func fb(bits uint64) float64 { return math.Float64frombits(bits) }

func cv(v ...float64) XCubicCurve2D {
	return NewXCubicCurve2D(v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7])
}

// sameBits reports whether got and want have the same bit pattern.
func sameBits(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Errorf("%s = %v (%#016x), want %v (%#016x)", name, got, math.Float64bits(got), want, math.Float64bits(want))
	}
}

// samePoint reports whether got and want have the same coordinates, bit for
// bit.
func samePoint(t *testing.T, name string, got, want XPoint2D) {
	t.Helper()
	sameBits(t, name+".X", got.X, want.X)
	sameBits(t, name+".Y", got.Y, want.Y)
}

// sameCurve reports whether got and want have the same coordinates, bit for
// bit.
func sameCurve(t *testing.T, name string, got, want XCubicCurve2D) {
	t.Helper()
	g := []float64{got.X1, got.Y1, got.CtrlX1, got.CtrlY1, got.CtrlX2, got.CtrlY2, got.X2, got.Y2}
	w := []float64{want.X1, want.Y1, want.CtrlX1, want.CtrlY1, want.CtrlX2, want.CtrlY2, want.X2, want.Y2}
	for i := range g {
		sameBits(t, name, g[i], w[i])
	}
}

// nearAngle checks an angle from Math.atan2, which Java computes with fdlibm
// as jmath.Atan2 does: got must have the bits of Java's value.
func nearAngle(t *testing.T, name string, got, javaWant float64) {
	t.Helper()
	sameBits(t, name, got, javaWant)
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", name)
		}
	}()
	f()
}
