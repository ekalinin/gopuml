// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/Segment.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/TimeDirection.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package ngm

import "github.com/ekalinin/gopuml/internal/jtime"

// TimeDirection is upstream's TimeDirection: the direction in which time is
// walked.
type TimeDirection int

// The directions of time.
const (
	Forward TimeDirection = iota
	Backward
)

// String is the name of the enum constant.
func (d TimeDirection) String() string {
	if d == Forward {
		return "FORWARD"
	}
	return "BACKWARD"
}

// AdjustBySeconds is adjustBySeconds: time moved by seconds in the
// direction.
func (d TimeDirection) AdjustBySeconds(time jtime.LocalDateTime, seconds int64) jtime.LocalDateTime {
	if d == Forward {
		return time.PlusSeconds(seconds)
	}
	return time.MinusSeconds(seconds)
}

// Segment is upstream's Segment: a constant value between two instants,
// both excluded, walked in a direction. A backward segment starts after
// its end.
type Segment struct {
	start, end jtime.LocalDateTime
	value      Fraction
	direction  TimeDirection
}

// ForwardSegment is Segment.forward, which fails with Java's
// IllegalArgumentException unless start is before end.
func ForwardSegment(start, end jtime.LocalDateTime, value Fraction) Segment {
	if !start.IsBefore(end) {
		throw(IllegalArgumentException, "end must be after start")
	}
	return Segment{start, end, value, Forward}
}

// BackwardSegment is Segment.backward, which fails with Java's
// IllegalArgumentException unless end is before start.
func BackwardSegment(start, end jtime.LocalDateTime, value Fraction) Segment {
	if !end.IsBefore(start) {
		throw(IllegalArgumentException, "end must be before start")
	}
	return Segment{start, end, value, Backward}
}

// TimeDirection is getTimeDirection.
func (s Segment) TimeDirection() TimeDirection { return s.direction }

// StartExclusive is startExclusive.
func (s Segment) StartExclusive() jtime.LocalDateTime { return s.start }

// EndExclusive is endExclusive.
func (s Segment) EndExclusive() jtime.LocalDateTime { return s.end }

// Value is getValue.
func (s Segment) Value() Fraction { return s.value }

// Split is split: the two segments on each side of time, which fails with
// Java's IllegalArgumentException when time is outside the segment.
func (s Segment) Split(time jtime.LocalDateTime) [2]Segment {
	if !s.Includes(time) {
		throw(IllegalArgumentException, "time must be within the segment bounds")
	}
	if s.direction == Forward {
		return [2]Segment{ForwardSegment(s.start, time, s.value), ForwardSegment(time, s.end, s.value)}
	}
	return [2]Segment{BackwardSegment(s.start, time, s.value), BackwardSegment(time, s.end, s.value)}
}

// Includes is includes: whether time is strictly inside the segment.
func (s Segment) Includes(time jtime.LocalDateTime) bool {
	if s.direction == Forward {
		return time.IsAfter(s.start) && time.IsBefore(s.end)
	}
	return time.IsBefore(s.start) && time.IsAfter(s.end)
}

// String is toString.
func (s Segment) String() string {
	return s.direction.String() + " ]" + s.start.String() + ", " + s.end.String() + "[ value=" + s.value.String()
}

// Intersection is Segment.intersection: the part common to all the
// segments, with their values combined by op in order. It fails with
// Java's IllegalArgumentException when there are no segments, when their
// directions differ, or when they do not overlap.
func Intersection(segments []Segment, op func(a, b Fraction) Fraction) Segment {
	if len(segments) == 0 {
		throw(IllegalArgumentException, "No segments to intersect")
	}
	if len(segments) == 1 {
		return segments[0]
	}
	direction := segments[0].direction
	start, end := segments[0].start, segments[0].end
	combined := segments[0].value
	for _, segment := range segments[1:] {
		if segment.direction != direction {
			throw(IllegalArgumentException, "All segments must have the same direction")
		}
		if direction == Forward {
			if segment.start.IsAfter(start) {
				start = segment.start
			}
			if segment.end.IsBefore(end) {
				end = segment.end
			}
		} else {
			if segment.start.IsBefore(start) {
				start = segment.start
			}
			if segment.end.IsAfter(end) {
				end = segment.end
			}
		}
		combined = op(combined, segment.value)
	}
	if direction == Forward {
		if !start.IsBefore(end) {
			throw(IllegalArgumentException, "Segments do not overlap")
		}
		return ForwardSegment(start, end, combined)
	}
	if !end.IsBefore(start) {
		throw(IllegalArgumentException, "Segments do not overlap")
	}
	return BackwardSegment(start, end, combined)
}

// ClampedStart is computeClampedStart: current, or the start of the
// segment when current is before it in the direction of the segment.
func (s Segment) ClampedStart(current jtime.LocalDateTime) jtime.LocalDateTime {
	if s.direction == Forward {
		if current.IsAfter(s.start) {
			return current
		}
		return s.start
	}
	if current.IsBefore(s.start) {
		return current
	}
	return s.start
}
