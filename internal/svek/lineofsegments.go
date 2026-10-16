// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/LineOfSegments.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"cmp"
	"slices"
)

// lineOfSegments spreads segments of a line so that they do not overlap,
// keeping their mean position (upstream's LineOfSegments).
type lineOfSegments struct {
	all []*segment
}

type segment struct {
	idx      int
	middle   float64
	halfSize float64
}

func (s *segment) overlap(other *segment) float64 {
	distance := other.middle - s.middle
	if distance < 0 {
		throwIllegalArgument("")
	}
	diff := distance - s.halfSize - other.halfSize
	if diff > 0 {
		return 0
	}
	return -diff
}

func (l *lineOfSegments) addSegment(x1, x2 float64) {
	l.all = append(l.all, &segment{idx: len(l.all), middle: (x1 + x2) / 2, halfSize: (x2 - x1) / 2})
}

func (l *lineOfSegments) mean() float64 {
	sum := 0.0
	for _, seg := range l.all {
		sum += seg.middle
	}
	return sum / float64(len(l.all))
}

func (l *lineOfSegments) solveOverlapsInternal() {
	if len(l.all) < 2 {
		return
	}
	// Collections.sort is stable.
	slices.SortStableFunc(l.all, func(a, b *segment) int { return compareDouble(a.middle, b.middle) })
	for range l.all {
		if !l.oneLoop() {
			return
		}
	}
}

func (l *lineOfSegments) oneLoop() bool {
	for i := len(l.all) - 2; i >= 0; i-- {
		seg1, seg2 := l.all[i], l.all[i+1]
		if overlap := seg1.overlap(seg2); overlap > 0 {
			for k := i + 1; k < len(l.all); k++ {
				l.all[k].middle += overlap
			}
			return true
		}
	}
	return false
}

// solveOverlaps returns the new start of each segment, by index.
func (l *lineOfSegments) solveOverlaps() []float64 {
	mean1 := l.mean()
	l.solveOverlapsInternal()
	mean2 := l.mean()
	if diff := mean1 - mean2; diff != 0 {
		for _, seg := range l.all {
			seg.middle += diff
		}
	}
	result := make([]float64, len(l.all))
	for _, seg := range l.all {
		result[seg.idx] = seg.middle - seg.halfSize
	}
	return result
}

// compareDouble is Double.compare: -0.0 is less than 0.0 and NaN is the
// greatest value.
func compareDouble(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	an, bn := a != a, b != b
	switch {
	case an && bn:
		return 0
	case an:
		return 1
	case bn:
		return -1
	}
	// a == b: order -0.0 before 0.0.
	return cmp.Compare(signBit(b), signBit(a))
}

func signBit(v float64) int {
	if 1/v < 0 {
		return 1
	}
	return 0
}
