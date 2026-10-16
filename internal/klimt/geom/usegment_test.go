// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"slices"
	"testing"
)

func TestUSegment(t *testing.T) {
	coord := []float64{1, 2.5, math.Copysign(0, -1), 1e7}
	s := NewUSegment(coord, SegCubicTo)
	coord[0] = 99
	if got, want := s.String(), "SEG_CUBICTO [1.0, 2.5, -0.0, 1.0E7]"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got, want := NewUSegment(nil, SegClose).String(), "SEG_CLOSE []"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if s.SegmentType() != SegCubicTo {
		t.Errorf("SegmentType() = %v", s.SegmentType())
	}

	tests := []struct {
		s    USegment
		want []float64
	}{
		{NewUSegment([]float64{1, 2}, SegLineTo), []float64{11, 22}},
		{NewUSegment([]float64{1, 2, 3, 4, 5, 6}, SegCubicTo), []float64{11, 22, 13, 24, 15, 26}},
		{NewUSegment([]float64{1, 2, 3, 0, 1, 6, 7}, SegArcTo), []float64{1, 2, 3, 0, 1, 16, 27}},
	}
	for _, tt := range tests {
		got := tt.s.Translate(10, 20)
		if !slices.Equal(got.Coord(), tt.want) || got.SegmentType() != tt.s.SegmentType() {
			t.Errorf("%v.Translate = %v, want %v", tt.s, got, tt.want)
		}
	}
	mustPanic(t, "Translate of a close segment", func() { NewUSegment(nil, SegClose).Translate(1, 1) })
}
