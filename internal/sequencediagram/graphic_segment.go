// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/graphic/Segment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"math"
	"slices"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Segment is a vertical interval [Pos1, Pos2].
type Segment struct {
	pos1, pos2 float64
}

// NewSegment returns the interval; it panics when pos2 < pos1, as
// upstream.
func NewSegment(pos1, pos2 float64) Segment {
	if pos2 < pos1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return Segment{pos1: pos1, pos2: pos2}
}

// ContainsY reports whether y is in the interval.
func (s Segment) ContainsY(y float64) bool { return y >= s.pos1 && y <= s.pos2 }

// Contains reports whether other is inside the interval.
func (s Segment) Contains(other Segment) bool {
	return s.ContainsY(other.pos1) && s.ContainsY(other.pos2)
}

// String is upstream's toString.
func (s Segment) String() string {
	return jfmt.DoubleToString(s.pos1) + " - " + jfmt.DoubleToString(s.pos2)
}

// Length returns the length of the interval.
func (s Segment) Length() float64 { return s.pos2 - s.pos1 }

// Pos1 returns the start.
func (s Segment) Pos1() float64 { return s.pos1 }

// Pos2 returns the end.
func (s Segment) Pos2() float64 { return s.pos2 }

// Merge returns the smallest interval containing both.
func (s Segment) Merge(other Segment) Segment {
	return NewSegment(jmath.Min(s.pos1, other.pos1), jmath.Max(s.pos2, other.pos2))
}

// CutSegmentIfNeed returns the parts of the interval outside the delays.
func (s Segment) CutSegmentIfNeed(allDelays []Segment) []Segment {
	sortedDelay := slices.Clone(allDelays)
	// Collections.sort is stable; the comparator is (int) Math.signum.
	slices.SortStableFunc(sortedDelay, func(a, b Segment) int { return int(jmath.Signum(a.pos1 - b.pos1)) })
	var result2 []Segment
	pendingStart := s.pos1
	for _, pause := range sortedDelay {
		if math.Abs(pause.pos1-pendingStart) < 0.001 {
			pendingStart = pause.pos2
			continue
		}
		if pause.pos1 < pendingStart {
			continue
		}
		if pause.pos1 > s.pos2 {
			if pendingStart < s.pos2 {
				result2 = append(result2, NewSegment(pendingStart, s.pos2))
			}
			return result2
		}
		if s.Contains(pause) {
			result2 = append(result2, NewSegment(pendingStart, pause.pos1))
			pendingStart = pause.pos2
		}
	}
	if pendingStart < s.pos2 {
		result2 = append(result2, NewSegment(pendingStart, s.pos2))
	}
	return result2
}
