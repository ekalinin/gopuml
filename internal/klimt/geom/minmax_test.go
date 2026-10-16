// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestMinMax(t *testing.T) {
	inverted := MinMaxEmpty(false)
	if got, want := inverted.String(),
		"(1.7976931348623157E308,1.7976931348623157E308)->(-1.7976931348623157E308,-1.7976931348623157E308)"; got != want {
		t.Errorf("MinMaxEmpty(false) = %q, want %q", got, want)
	}
	if MinMaxEmpty(true) != (MinMax{}) {
		t.Error("MinMaxEmpty(true) is not the zero value")
	}

	m := MinMaxEmpty(true).AddPointXY(1.5, -2).AddPoint(NewXPoint2D(1e7, 0.001))
	if got, want := m.String(), "(0.0,-2.0)->(1.0E7,0.001)"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	// The first point replaces an inverted box.
	m = inverted.AddPointXY(3, 4)
	if m.MinX() != 3 || m.MinY() != 4 || m.MaxX() != 3 || m.MaxY() != 4 {
		t.Errorf("first point = %v", m)
	}
	m = m.AddMinMax(MinMaxFromMax(-1, 10))
	if got, want := m.String(), "(-1.0,0.0)->(3.0,10.0)"; got != want {
		t.Errorf("AddMinMax = %q, want %q", got, want)
	}
	sameBits(t, "Width", m.Width(), 4)
	sameBits(t, "Height", m.Height(), 10)
	if got := m.Dimension(); got != NewXDimension2D(4, 10) {
		t.Errorf("Dimension() = %v", got)
	}
	if got := m.Enlarge(1, 2); got.String() != "(-1.0,0.0)->(4.0,12.0)" {
		t.Errorf("Enlarge = %v", got)
	}
	if got := MinMaxFromDim(NewXDimension2D(5, 6)); got.String() != "(0.0,0.0)->(5.0,6.0)" {
		t.Errorf("MinMaxFromDim = %v", got)
	}
	// Merging keeps -0.0 as the minimum and 0.0 as the maximum, as
	// Math.min and Math.max do.
	z := MinMaxEmpty(true).AddPointXY(math.Copysign(0, -1), math.Copysign(0, -1))
	if got := z.String(); got != "(-0.0,-0.0)->(0.0,0.0)" {
		t.Errorf("signed zeros = %q", got)
	}
	mustPanic(t, "AddPointXY(NaN)", func() { MinMaxEmpty(true).AddPointXY(math.NaN(), 0) })
	mustPanic(t, "Dimension of an inverted box", func() { inverted.Dimension() })
}

func TestMinMaxDoesHorizontalCross(t *testing.T) {
	m := MinMaxFromMax(10, 10)
	tests := []struct {
		pt1, pt2 XPoint2D
		want     bool
	}{
		{NewXPoint2D(-1, 5), NewXPoint2D(11, 5), true},
		{NewXPoint2D(11, 5), NewXPoint2D(-1, 5), true},
		{NewXPoint2D(0, 5), NewXPoint2D(11, 5), false},
		{NewXPoint2D(-1, 11), NewXPoint2D(11, 11), false},
		{NewXPoint2D(-1, 10), NewXPoint2D(11, 10), true},
	}
	for _, tt := range tests {
		if got := m.DoesHorizontalCross(tt.pt1, tt.pt2); got != tt.want {
			t.Errorf("DoesHorizontalCross(%v, %v) = %v, want %v", tt.pt1, tt.pt2, got, tt.want)
		}
	}
	mustPanic(t, "not horizontal", func() { m.DoesHorizontalCross(NewXPoint2D(0, 0), NewXPoint2D(1, 1)) })
	mustPanic(t, "empty", func() { m.DoesHorizontalCross(NewXPoint2D(1, 1), NewXPoint2D(1, 1)) })
}

func TestMinMaxMutable(t *testing.T) {
	m := MinMaxMutableEmpty(false)
	if got, want := m.String(),
		"X=1.7976931348623157E308 to -1.7976931348623157E308 and Y=1.7976931348623157E308 to -1.7976931348623157E308"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if !m.IsInfinity() {
		t.Error("IsInfinity() = false on an empty box")
	}
	m.AddPointXY(3, 4)
	m.AddPoint(NewXPoint2D(-0.5, 12.25))
	if got, want := m.String(), "X=-0.5 to 3.0 and Y=4.0 to 12.25"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if m.IsInfinity() {
		t.Error("IsInfinity() = true after adding points")
	}
	if got := m.Dimension().String(); got != "[3.5,8.25]" {
		t.Errorf("Dimension() = %q", got)
	}
	if got := MinMaxFromMutable(m).String(); got != "(-0.5,4.0)->(3.0,12.25)" {
		t.Errorf("MinMaxFromMutable = %q", got)
	}
	m.Reset()
	if m.MinX() != 0 || m.MinY() != 0 || m.MaxX() != 0 || m.MaxY() != 0 {
		t.Errorf("Reset() = %v", m)
	}
	f := MinMaxMutableFromMax(5, -6)
	if got := f.String(); got != "X=0.0 to 5.0 and Y=-6.0 to 0.0" {
		t.Errorf("MinMaxMutableFromMax = %q", got)
	}
	if *MinMaxMutableEmpty(true) != (MinMaxMutable{}) {
		t.Error("MinMaxMutableEmpty(true) is not the zero value")
	}
	mustPanic(t, "AddPointXY(NaN)", func() { m.AddPointXY(0, math.NaN()) })
	mustPanic(t, "MinMaxMutableFromMax(NaN)", func() { MinMaxMutableFromMax(math.NaN(), 0) })
}

func TestMinFinder(t *testing.T) {
	f := NewMinFinder()
	if got, want := f.String(), "minX=1.7976931348623157E308 minY=1.7976931348623157E308"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	f.Manage(3, 4)
	f.ManagePoint(NewXPoint2D(5, -1e-5))
	if got, want := f.String(), "minX=3.0 minY=-1.0E-5"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	g := NewMinFinder()
	g.Manage(-7, 100)
	f.ManageMinFinder(g)
	if f.MinX() != -7 || f.MinY() != -1e-5 {
		t.Errorf("ManageMinFinder = %v", f)
	}
	// NaN never replaces a minimum.
	f.Manage(math.NaN(), math.NaN())
	if f.MinX() != -7 || f.MinY() != -1e-5 {
		t.Errorf("Manage(NaN) = %v", f)
	}
}
