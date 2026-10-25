// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstantWeekday.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstantSpecificDays.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstantTimeWindow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/PiecewiseConstantHours.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package ngm

import (
	"maps"
	"slices"

	"github.com/ekalinin/gopuml/internal/jtime"
)

// Weekday is upstream's PiecewiseConstantWeekday: a value for each day of
// the week, on segments of one day.
type Weekday struct {
	workloadByDay [7]Fraction
}

// WeekdayOf is PiecewiseConstantWeekday.of: the same value every day.
func WeekdayOf(sameWorkload Fraction) *Weekday {
	w := &Weekday{}
	for i := range w.workloadByDay {
		w.workloadByDay[i] = sameWorkload
	}
	return w
}

// With is with: a copy with another value for day.
func (w *Weekday) With(day jtime.DayOfWeek, newWorkload Fraction) *Weekday {
	c := *w
	c.workloadByDay[day.Ordinal()] = newWorkload
	return &c
}

// SegmentAt returns the day that holds instant. Walking backward, midnight
// belongs to the day before.
func (w *Weekday) SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment {
	midnight := instant.LocalDate().AtStartOfDay()
	if direction == Forward {
		return ForwardSegment(midnight, midnight.PlusDays(1), w.workloadByDay[midnight.DayOfWeek().Ordinal()])
	}
	end := midnight
	if instant.LocalTime() == jtime.Midnight {
		end = midnight.MinusDays(1)
	}
	return BackwardSegment(end.PlusDays(1), end, w.workloadByDay[end.DayOfWeek().Ordinal()])
}

// SpecificDays is upstream's PiecewiseConstantSpecificDays: a value for
// some dates and a default value for the others, on segments of one day.
type SpecificDays struct {
	dayToFraction map[jtime.LocalDate]Fraction
	defaultValue  Fraction
}

// SpecificDaysOf is PiecewiseConstantSpecificDays.of(value).
func SpecificDaysOf(value Fraction) *SpecificDays {
	return &SpecificDays{defaultValue: value}
}

// SpecificDaysOfMap is PiecewiseConstantSpecificDays.of(defaultValue,
// dayToFraction).
func SpecificDaysOfMap(defaultValue Fraction, dayToFraction map[jtime.LocalDate]Fraction) *SpecificDays {
	return &SpecificDays{maps.Clone(dayToFraction), defaultValue}
}

// WithDay is withDay: a copy with another value for day.
func (s *SpecificDays) WithDay(day jtime.LocalDate, value Fraction) *SpecificDays {
	m := maps.Clone(s.dayToFraction)
	if m == nil {
		m = map[jtime.LocalDate]Fraction{}
	}
	m[day] = value
	return &SpecificDays{m, s.defaultValue}
}

func (s *SpecificDays) value(day jtime.LocalDate) Fraction {
	if v, ok := s.dayToFraction[day]; ok {
		return v
	}
	return s.defaultValue
}

// SegmentAt returns the day that holds instant. Walking backward, midnight
// belongs to the day before.
func (s *SpecificDays) SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment {
	if direction == Forward {
		day := instant.LocalDate()
		start := day.AtStartOfDay()
		return ForwardSegment(start, start.PlusDays(1), s.value(day))
	}
	day := instant.LocalDate()
	if instant.LocalTime() == jtime.Midnight {
		day = day.MinusDays(1)
	}
	end := day.AtStartOfDay()
	return BackwardSegment(end.PlusDays(1), end, s.value(day))
}

// TimeWindow is upstream's PiecewiseConstantTimeWindow: one from offBefore
// to offAfter, zero before and after.
type TimeWindow struct {
	offBefore, offAfter jtime.LocalDateTime
}

// NewTimeWindow is new PiecewiseConstantTimeWindow, which fails with Java's
// IllegalArgumentException unless offBefore is before offAfter.
func NewTimeWindow(offBefore, offAfter jtime.LocalDateTime) *TimeWindow {
	if !offBefore.IsBefore(offAfter) {
		throw(IllegalArgumentException, "offAfter must be after offBefore")
	}
	return &TimeWindow{offBefore, offAfter}
}

// SegmentAt returns the segment from instant to the next bound of the
// window in direction, or to LocalDateTime.MAX or MIN outside it.
func (w *TimeWindow) SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment {
	if direction == Forward {
		switch {
		case instant.IsBefore(w.offBefore):
			return ForwardSegment(instant, w.offBefore, Zero)
		case !instant.IsBefore(w.offAfter):
			return ForwardSegment(instant, jtime.MaxDateTime, Zero)
		}
		return ForwardSegment(instant, w.offAfter, One)
	}
	switch {
	case instant.IsAfter(w.offAfter):
		return BackwardSegment(instant, w.offAfter, Zero)
	case !instant.IsAfter(w.offBefore):
		return BackwardSegment(instant, jtime.MinDateTime, Zero)
	}
	return BackwardSegment(instant, w.offBefore, One)
}

// TimeRange is a range of working hours, from start to end on the same
// day, an end at midnight meaning the end of the day.
type TimeRange struct {
	Start, End jtime.LocalTime
}

// hoursBoundary is an entry of the TreeMap of PiecewiseConstantHours.
type hoursBoundary struct {
	time  jtime.LocalTime
	value Fraction
}

// Hours is upstream's PiecewiseConstantHours: a value for the hours of
// every day, with boundaries at times of day.
type Hours struct {
	defaultValue Fraction
	// boundaries are sorted by time: the value from each time to the next.
	boundaries []hoursBoundary
}

// HoursOf is PiecewiseConstantHours.of.
func HoursOf(defaultValue Fraction) *Hours { return &Hours{defaultValue: defaultValue} }

// HoursFromWorkingHours is PiecewiseConstantHours.fromWorkingHours: one
// during the working hours and zero otherwise.
func HoursFromWorkingHours(workingTimeRanges []TimeRange) *Hours {
	result := HoursOf(Zero)
	for _, r := range workingTimeRanges {
		result = result.With(r, One)
	}
	return result
}

// With is with: a copy with fraction from the start of timeRange to its end,
// which fails with Java's IllegalArgumentException when the start is not
// before an end other than midnight.
func (h *Hours) With(timeRange TimeRange, fraction Fraction) *Hours {
	if timeRange.End != jtime.Midnight && timeRange.Start.CompareTo(timeRange.End) >= 0 {
		throw(IllegalArgumentException, "start must be strictly before end")
	}
	result := &Hours{h.defaultValue, slices.Clone(h.boundaries)}
	result.put(timeRange.Start, fraction)
	if timeRange.End != jtime.Midnight {
		result.put(timeRange.End, h.defaultValue)
	}
	return result
}

func (h *Hours) put(t jtime.LocalTime, value Fraction) {
	i, found := slices.BinarySearchFunc(h.boundaries, t, func(b hoursBoundary, t jtime.LocalTime) int {
		return b.time.CompareTo(t)
	})
	if found {
		h.boundaries[i].value = value
		return
	}
	h.boundaries = slices.Insert(h.boundaries, i, hoursBoundary{t, value})
}

// floor returns the index of the greatest boundary at or before t, or -1.
func (h *Hours) floor(t jtime.LocalTime) int {
	i := len(h.boundaries) - 1
	for i >= 0 && h.boundaries[i].time.IsAfter(t) {
		i--
	}
	return i
}

// lower returns the index of the greatest boundary before t, or -1.
func (h *Hours) lower(t jtime.LocalTime) int {
	i := len(h.boundaries) - 1
	for i >= 0 && !h.boundaries[i].time.IsBefore(t) {
		i--
	}
	return i
}

// higherKey returns the first boundary after t.
func (h *Hours) higherKey(t jtime.LocalTime) (jtime.LocalTime, bool) {
	for _, b := range h.boundaries {
		if b.time.IsAfter(t) {
			return b.time, true
		}
	}
	return jtime.LocalTime{}, false
}

// get is TreeMap.get.
func (h *Hours) get(t jtime.LocalTime) (Fraction, bool) {
	for _, b := range h.boundaries {
		if b.time == t {
			return b.value, true
		}
	}
	return Fraction{}, false
}

// SegmentAt returns the segment between two boundaries that holds instant.
// Walking backward, midnight belongs to the last segment of the day before.
func (h *Hours) SegmentAt(instant jtime.LocalDateTime, direction TimeDirection) Segment {
	if direction == Forward {
		dayStart := instant.LocalDate().AtStartOfDay()
		startKey := jtime.Midnight
		if i := h.floor(instant.LocalTime()); i >= 0 {
			startKey = h.boundaries[i].time
		}
		value, ok := h.get(startKey)
		if !ok {
			value = h.defaultValue
		}
		start := dayStart.WithTime(startKey)
		end := dayStart.PlusDays(1)
		if endKey, ok := h.higherKey(startKey); ok {
			end = dayStart.WithTime(endKey)
		}
		return ForwardSegment(start, end, value)
	}
	t := instant.LocalTime()
	if t == jtime.Midnight {
		dayStart := instant.LocalDate().MinusDays(1).AtStartOfDay()
		startKey, value := jtime.Midnight, h.defaultValue
		if n := len(h.boundaries); n > 0 {
			startKey, value = h.boundaries[n-1].time, h.boundaries[n-1].value
		}
		return BackwardSegment(dayStart.PlusDays(1), dayStart.WithTime(startKey), value)
	}
	dayStart := instant.LocalDate().AtStartOfDay()
	startKey, value := jtime.Midnight, h.defaultValue
	if i := h.lower(t); i >= 0 {
		startKey, value = h.boundaries[i].time, h.boundaries[i].value
	}
	start := dayStart.PlusDays(1)
	if endKey, ok := h.higherKey(startKey); ok {
		start = dayStart.WithTime(endKey)
	}
	return BackwardSegment(start, dayStart.WithTime(startKey), value)
}
