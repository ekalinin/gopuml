// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestXDimension2D(t *testing.T) {
	d := NewXDimension2D(10.1, 2.3)
	// Expected strings computed by upstream's toString.
	tests := []struct {
		name string
		got  XDimension2D
		want string
	}{
		{"Delta", d.Delta(0.7), "[10.799999999999999,3.0]"},
		{"DeltaWH", d.DeltaWH(0.1, 0.2), "[10.2,2.5]"},
		{"MergeTB", d.MergeTB(NewXDimension2D(3.3, 4.4)), "[10.1,6.7]"},
		{"MergeTB3", d.MergeTB3(NewXDimension2D(3.3, 4.4), NewXDimension2D(12.5, 0.1)), "[12.5,6.8]"},
		{"MergeLR", d.MergeLR(NewXDimension2D(3.3, 4.4)), "[13.399999999999999,4.4]"},
		{"AtLeast", d.AtLeast(5, 5), "[10.1,5.0]"},
		{"AtLeast width", d.AtLeast(20, 1), "[20.0,2.3]"},
		{"AtLeast same", d.AtLeast(1, 1), "[10.1,2.3]"},
		{"WithWidth", d.WithWidth(0.3), "[0.3,2.3]"},
		{"XDimension2DMergeLayoutT12B3", XDimension2DMergeLayoutT12B3(d, NewXDimension2D(0.2, 0.1), NewXDimension2D(7, 0.2)), "[10.1,2.6]"},
		{"XDimension2DMax", XDimension2DMax(d, NewXDimension2D(3, 7)), "[10.1,7.0]"},
		{"zero value", XDimension2D{}, "[0.0,0.0]"},
		{"small", NewXDimension2D(10, 2.5e-4), "[10.0,2.5E-4]"},
	}
	for _, tt := range tests {
		if got := tt.got.String(); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
		}
	}
	if d.Width() != 10.1 || d.Height() != 2.3 {
		t.Errorf("Width(), Height() = %v, %v", d.Width(), d.Height())
	}
	if d.DeltaWH(0, 0) != d {
		t.Error("DeltaWH(0, 0) changed the dimension")
	}
	mustPanic(t, "negative width", func() { NewXDimension2D(-1, 0) })
	mustPanic(t, "negative height", func() { NewXDimension2D(0, -1) })
	mustPanic(t, "NaN", func() { NewXDimension2D(math.NaN(), 0) })
	mustPanic(t, "Delta below zero", func() { d.Delta(-20) })
}
