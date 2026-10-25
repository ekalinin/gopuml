// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/TimeScale.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/TimeScaleWink.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/TimeScaleDaily.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/TimeScaleCompressed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/TimeScaleDailyHideClosed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/timescale/VariantTimeScale.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/WorkingHours.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// TimeScale is upstream's TimeScale: the horizontal position of each
// instant.
type TimeScale interface {
	Position(instant TimePoint) float64
	Width(instant TimePoint) float64
	IsBreaking(instant TimePoint) bool
}

// TimeScaleWink is upstream's TimeScaleWink: positions from 1970-01-01,
// for a project without a start.
type TimeScaleWink struct {
	cellWidth  float64
	printScale PrintScale
}

// NewTimeScaleWink is new TimeScaleWink.
func NewTimeScaleWink(size, scale float64, printScale PrintScale) *TimeScaleWink {
	return &TimeScaleWink{size * scale, printScale}
}

// Position implements TimeScale: the milliseconds times the width of a
// day, in Java's order of operations.
func (s *TimeScaleWink) Position(instant TimePoint) float64 {
	wink := instant.Millis()
	return float64(wink) * s.cellWidth / float64(millisecondsPerDay)
}

// Width implements TimeScale.
func (s *TimeScaleWink) Width(TimePoint) float64 { return s.cellWidth }

// IsBreaking implements TimeScale: weekly, the last day of a week of days
// since 1970-01-01.
func (s *TimeScaleWink) IsBreaking(instant TimePoint) bool {
	if s.printScale == PrintScaleWeekly {
		num := instant.Millis() / millisecondsPerDay
		return num%7 == 6
	}
	return true
}

// TimeScaleDaily is upstream's TimeScaleDaily: positions from the first
// day.
type TimeScaleDaily struct {
	basic *TimeScaleWink
	delta float64
}

// NewTimeScaleDaily is new TimeScaleDaily; zeroDay is the first day of
// "print between", nil for none.
func NewTimeScaleDaily(size float64, startingDay TimePoint, scale float64, zeroDay *jtime.LocalDate) *TimeScaleDaily {
	s := &TimeScaleDaily{basic: NewTimeScaleWink(size, scale, PrintScaleDaily)}
	if zeroDay == nil {
		s.delta = s.basic.Position(startingDay)
	} else {
		s.delta = s.basic.Position(TimePointOfStartOfDay(*zeroDay))
	}
	return s
}

// Position implements TimeScale.
func (s *TimeScaleDaily) Position(instant TimePoint) float64 {
	return s.basic.Position(instant) - s.delta
}

// Width implements TimeScale.
func (s *TimeScaleDaily) Width(instant TimePoint) float64 { return s.basic.Width(instant) }

// IsBreaking implements TimeScale.
func (s *TimeScaleDaily) IsBreaking(TimePoint) bool { return true }

// TimeScaleCompressed is upstream's TimeScaleCompressed: the daily scale,
// breaking on Sundays.
type TimeScaleCompressed struct {
	daily *TimeScaleDaily
}

// NewTimeScaleCompressed is new TimeScaleCompressed.
func NewTimeScaleCompressed(size float64, calendar TimePoint, scale float64,
	zeroDay *jtime.LocalDate) *TimeScaleCompressed {
	return &TimeScaleCompressed{NewTimeScaleDaily(size, calendar, scale, zeroDay)}
}

// Position implements TimeScale.
func (s *TimeScaleCompressed) Position(instant TimePoint) float64 { return s.daily.Position(instant) }

// Width implements TimeScale.
func (s *TimeScaleCompressed) Width(instant TimePoint) float64 { return s.daily.Width(instant) }

// IsBreaking implements TimeScale.
func (s *TimeScaleCompressed) IsBreaking(instant TimePoint) bool {
	return instant.DayOfWeek() == jtime.Sunday
}

// TimeScaleDailyHideClosed is upstream's TimeScaleDailyHideClosed: the
// closed days have no width. As upstream, an instant that is not at
// midnight fails with Java's NullPointerException, and an instant before
// the first day with its IllegalArgumentException.
type TimeScaleDailyHideClosed struct {
	cellWidth   float64
	startingDay TimePoint
	biggest     TimePoint
	openClose   *OpenClose
	startingInt map[TimePoint]int
}

// NewTimeScaleDailyHideClosed is new TimeScaleDailyHideClosed.
func NewTimeScaleDailyHideClosed(size float64, startingDay TimePoint, scale float64,
	openClose *OpenClose) *TimeScaleDailyHideClosed {
	return &TimeScaleDailyHideClosed{cellWidth: size * scale, startingDay: startingDay, biggest: startingDay,
		openClose: openClose, startingInt: map[TimePoint]int{startingDay: 0}}
}

// Position implements TimeScale.
func (s *TimeScaleDailyHideClosed) Position(instant TimePoint) float64 {
	if instant.CompareTo(s.startingDay) < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.computeUpTo(instant)
	v, ok := s.startingInt[instant]
	if !ok {
		uml.ThrowNPE(`Cannot invoke "java.lang.Integer.intValue()" because the return value of ` +
			`"java.util.Map.get(Object)" is null`)
	}
	return float64(v) * s.cellWidth
}

func (s *TimeScaleDailyHideClosed) computeUpTo(dest TimePoint) {
	if dest.CompareTo(s.biggest) <= 0 {
		return
	}
	i := s.biggest
	current := s.startingInt[i]
	for {
		if s.Width(i) > 0 {
			current++
		}
		i = i.Increment()
		s.startingInt[i] = current
		if i.CompareTo(dest) >= 0 {
			break
		}
	}
}

// Width implements TimeScale: 0 on a closed day.
func (s *TimeScaleDailyHideClosed) Width(instant TimePoint) float64 {
	if s.openClose.IsClosed(instant.Day()) {
		return 0
	}
	return s.cellWidth
}

// IsBreaking implements TimeScale.
func (s *TimeScaleDailyHideClosed) IsBreaking(TimePoint) bool { return true }

// VariantTimeScale is upstream's VariantTimeScale: days drawn with their
// working hours.
type VariantTimeScale struct {
	min, max     jtime.LocalDate
	workingHours *WorkingHours
	openClose    *OpenClose
}

// Position implements TimeScale: the widths of the days before, then the
// part of the day.
func (s *VariantTimeScale) Position(instant TimePoint) float64 {
	dateOnly := instant.Day()
	tmp := 0.0
	for dd := s.min; dd.IsBefore(dateOnly); dd = dd.PlusDays(1) {
		tmp += s.workingHours.WorkingDurationForOneDay(s.openClose.IsClosed(dd))
	}
	return tmp + s.workingHours.PartialDayWidth(instant.LocalDateTime().LocalTime())
}

// Width implements TimeScale.
func (s *VariantTimeScale) Width(instant TimePoint) float64 {
	return s.workingHours.WorkingDurationForOneDay(s.openClose.IsClosed(instant.Day()))
}

// IsBreaking implements TimeScale.
func (s *VariantTimeScale) IsBreaking(TimePoint) bool { return false }

// WorkingHours is upstream's WorkingHours: the widths of the working and
// closed hours of a day.
type WorkingHours struct {
	workingTimeRanges *TimeRanges
	dayWidth          float64
	closeFactor       float64
}

// WorkingDurationForOneDay is getWorkingDurationForOneDay.
func (w *WorkingHours) WorkingDurationForOneDay(isClosed bool) float64 {
	if isClosed {
		return w.widthForClosedHours(jtime.DurationOfHours(24))
	}
	working := w.workingTimeRanges.TotalWorkingDuration()
	pause := jtime.DurationOfHours(24).Plus(working.Negated())
	return w.widthForWorkingHours(working) + w.widthForClosedHours(pause)
}

// PartialDayWidth is getPartialDayWidth: the width from midnight to pos.
func (w *WorkingHours) PartialDayWidth(pos jtime.LocalTime) float64 {
	working := jtime.ZeroDuration
	pause := jtime.ZeroDuration
	t := jtime.Midnight
	for _, r := range w.workingTimeRanges.Ranges() {
		start, end := rangeTime(r.start), rangeTime(r.end)
		if pos.IsBefore(start) {
			pause = pause.Plus(jtime.BetweenTimes(t, pos))
			return w.widthForWorkingHours(working) + w.widthForClosedHours(pause)
		} else if pos.IsBefore(end) {
			pause = pause.Plus(jtime.BetweenTimes(t, start))
			working = working.Plus(jtime.BetweenTimes(start, pos))
			return w.widthForWorkingHours(working) + w.widthForClosedHours(pause)
		}
		pause = pause.Plus(jtime.BetweenTimes(t, start))
		working = working.Plus(r.Duration())
		t = end
	}
	pause = pause.Plus(jtime.BetweenTimes(t, pos))
	return w.widthForWorkingHours(working) + w.widthForClosedHours(pause)
}

// rangeTime is a time of a range, with Java's NullPointerException for a
// time that did not parse.
func rangeTime(t *jtime.LocalTime) jtime.LocalTime {
	if t == nil {
		uml.ThrowNPE("")
	}
	return *t
}

func (w *WorkingHours) widthForWorkingHours(between jtime.Duration) float64 {
	return float64(between.Seconds()) * w.dayWidth / (24.0 * 3600.0)
}

func (w *WorkingHours) widthForClosedHours(between jtime.Duration) float64 {
	return w.closeFactor * w.widthForWorkingHours(between)
}

// DrawOneDay is drawOneDay: the closed hours of a day, as rectangles.
func (w *WorkingHours) DrawOneDay(ug klimt.UGraphic, height float64) {
	current := w.PartialDayWidth(jtime.Midnight)
	for _, r := range w.workingTimeRanges.Ranges() {
		start := w.PartialDayWidth(rangeTime(r.start))
		ug.Apply(klimt.UTranslateDx(current)).Draw(klimt.URectangleBuild(start-current, height))
		current = w.PartialDayWidth(rangeTime(r.end))
	}
	last := w.PartialDayWidth(jtime.MaxTime)
	ug.Apply(klimt.UTranslateDx(current)).Draw(klimt.URectangleBuild(last-current, height))
}
