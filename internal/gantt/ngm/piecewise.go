// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstant.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/AbstractPiecewiseConstant.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/Combiner.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstantUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package ngm

import "github.com/ekalinin/gopuml/internal/jtime"

// PiecewiseConstant is upstream's PiecewiseConstant: a function of time
// that is constant on segments.
type PiecewiseConstant interface {
	// SegmentAt returns the segment that holds instant, walked in
	// direction.
	SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment
}

// SegmentIterator is the iterator of upstream's
// AbstractPiecewiseConstant.iterateSegmentsFrom: the segments one after the
// other, without end. It fails with Java's IllegalStateException on the
// 10000th segment, as upstream detects an infinite loop.
type SegmentIterator struct {
	function  PiecewiseConstant
	direction TimeDirection
	current   jtime.LocalDateTime
	counter   int
}

// IterateSegmentsFrom is iterateSegmentsFrom.
func IterateSegmentsFrom(f PiecewiseConstant, instant jtime.LocalDateTime, direction TimeDirection) *SegmentIterator {
	return &SegmentIterator{function: f, direction: direction, current: instant, counter: 9999}
}

// HasNext is hasNext: always true.
func (it *SegmentIterator) HasNext() bool { return true }

// Next is next: the segment that holds the end of the previous one.
func (it *SegmentIterator) Next() Segment {
	if it.counter <= 0 {
		throw(IllegalStateException, "Infinite loop detected in PiecewiseConstant iteration")
	}
	it.counter--
	segment := it.function.SegmentAt(it.current, it.direction)
	it.current = segment.EndExclusive()
	return segment
}

// Combiner is upstream's Combiner: functions combined at each instant by an
// operation.
type Combiner struct {
	elements  []PiecewiseConstant
	operation func(a, b Fraction) Fraction
}

func combine(op func(a, b Fraction) Fraction, elements []PiecewiseConstant) PiecewiseConstant {
	if len(elements) == 1 {
		return elements[0]
	}
	return &Combiner{elements, op}
}

// CombinerSum is Combiner.sum.
func CombinerSum(elements ...PiecewiseConstant) PiecewiseConstant { return combine(Sum, elements) }

// CombinerMin is Combiner.min.
func CombinerMin(elements ...PiecewiseConstant) PiecewiseConstant { return combine(Min, elements) }

// CombinerMax is Combiner.max.
func CombinerMax(elements ...PiecewiseConstant) PiecewiseConstant { return combine(Max, elements) }

// CombinerProduct is Combiner.product.
func CombinerProduct(elements ...PiecewiseConstant) PiecewiseConstant {
	return combine(Product, elements)
}

// SegmentAt is the intersection of the segments of the functions at
// instant, which fails with Java's IllegalStateException when there are
// fewer than two functions.
func (c *Combiner) SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment {
	if len(c.elements) < 2 {
		throw(IllegalStateException, "At least two functions are required for combination")
	}
	segments := make([]Segment, len(c.elements))
	for i, e := range c.elements {
		segments[i] = e.SegmentAt(instant, direction)
	}
	return Intersection(segments, c.operation)
}

// IsZeroOnDay is PiecewiseConstantUtils.isZeroOnDay: whether the function
// is zero during the whole day.
func IsZeroOnDay(value PiecewiseConstant, day jtime.LocalDate) bool {
	startOfNextDay := day.PlusDays(1).AtStartOfDay()
	it := IterateSegmentsFrom(value, day.AtStartOfDay(), Forward)
	for it.HasNext() {
		segment := it.Next()
		if !segment.StartExclusive().IsBefore(startOfNextDay) {
			break
		}
		if !segment.Value().IsZero() {
			return false
		}
	}
	return true
}
