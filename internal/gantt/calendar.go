// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/OpenClose.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/DayStatus.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/DaysAsDates.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/DayCalendarData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TimeRange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TimeRanges.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// DayStatus is upstream's DayStatus.
type DayStatus int

// The statuses of a day.
const (
	DayOpen DayStatus = iota + 1
	DayClose
)

// OpenClose is upstream's OpenClose: the open and closed days of the
// project, of a task or of a resource, and the working hours.
type OpenClose struct {
	weekdayStatus     [7]DayStatus
	dayStatus         map[jtime.LocalDate]DayStatus
	workingTimeRanges []*TimeRange
	offBefore         *jtime.LocalDate
	offAfter          *jtime.LocalDate
	cachedPlan        ngm.PiecewiseConstant
}

// NewOpenClose is new OpenClose: every day open.
func NewOpenClose() *OpenClose { return &OpenClose{dayStatus: map[jtime.LocalDate]DayStatus{}} }

// MutateMe is mutateMe: a copy with the days of the week and the dates of
// except on top, the working hours and the days off of this one.
func (oc *OpenClose) MutateMe(except *OpenClose) *OpenClose {
	result := NewOpenClose()
	result.weekdayStatus = oc.weekdayStatus
	for day, status := range oc.dayStatus {
		result.dayStatus[day] = status
	}
	result.offBefore = oc.offBefore
	result.offAfter = oc.offAfter
	result.workingTimeRanges = append(result.workingTimeRanges, oc.workingTimeRanges...)
	if except != nil {
		for i, status := range except.weekdayStatus {
			if status != 0 {
				result.weekdayStatus[i] = status
			}
		}
		for day, status := range except.dayStatus {
			result.dayStatus[day] = status
		}
	}
	return result
}

// DaysInWeek is daysInWeek: 7 minus the closed days of the week.
func (oc *OpenClose) DaysInWeek() int {
	result := 7
	for _, status := range oc.weekdayStatus {
		if status == DayClose {
			result--
		}
	}
	return result
}

// IsClosed is isClosed.
func (oc *OpenClose) IsClosed(day jtime.LocalDate) bool { return oc.localStatus(day) == DayClose }

func (oc *OpenClose) localStatus(day jtime.LocalDate) DayStatus {
	if oc.offBefore != nil && day.CompareTo(*oc.offBefore) < 0 {
		return DayClose
	}
	if oc.offAfter != nil && day.CompareTo(*oc.offAfter) > 0 {
		return DayClose
	}
	if status, ok := oc.dayStatus[day]; ok {
		return status
	}
	return oc.weekdayStatus[day.DayOfWeek().Ordinal()]
}

// CloseDayOfWeek is close(DayOfWeek).
func (oc *OpenClose) CloseDayOfWeek(day jtime.DayOfWeek) {
	oc.weekdayStatus[day.Ordinal()] = DayClose
	oc.cachedPlan = nil
}

// OpenDayOfWeek is open(DayOfWeek).
func (oc *OpenClose) OpenDayOfWeek(day jtime.DayOfWeek) {
	oc.weekdayStatus[day.Ordinal()] = DayOpen
	oc.cachedPlan = nil
}

// CloseDay is close(LocalDate).
func (oc *OpenClose) CloseDay(day jtime.LocalDate) {
	oc.dayStatus[day] = DayClose
	oc.cachedPlan = nil
}

// OpenDay is open(LocalDate).
func (oc *OpenClose) OpenDay(day jtime.LocalDate) {
	oc.dayStatus[day] = DayOpen
	oc.cachedPlan = nil
}

// LoadAtDUMMY is getLoadAtDUMMY: 0 on a closed day, 100 otherwise.
func (oc *OpenClose) LoadAtDUMMY(day jtime.LocalDate) int {
	if oc.IsClosed(day) {
		return 0
	}
	return 100
}

// SetOffBeforeDate is setOffBeforeDate.
func (oc *OpenClose) SetOffBeforeDate(day jtime.LocalDate) {
	oc.offBefore = &day
	oc.cachedPlan = nil
}

// SetOffAfterDate is setOffAfterDate.
func (oc *OpenClose) SetOffAfterDate(day jtime.LocalDate) {
	oc.offAfter = &day
	oc.cachedPlan = nil
}

// LastDayIfAny is getLastDayIfAny.
func (oc *OpenClose) LastDayIfAny() (jtime.LocalDate, bool) {
	if oc.offAfter == nil {
		return jtime.LocalDate{}, false
	}
	return *oc.offAfter, true
}

// AsPiecewiseConstant is asPiecewiseConstant: one on the open days and
// zero on the closed ones, an opened date winning over its closed day of
// the week, times the working hours, within the days off. The result is
// cached, and adding working hours does not clear the cache, as upstream.
func (oc *OpenClose) AsPiecewiseConstant() ngm.PiecewiseConstant {
	if oc.cachedPlan != nil {
		return oc.cachedPlan
	}
	weekPattern := ngm.WeekdayOf(ngm.One)
	for _, day := range jtime.DaysOfWeek {
		if oc.weekdayStatus[day.Ordinal()] == DayClose {
			weekPattern = weekPattern.With(day, ngm.Zero)
		}
	}
	var result ngm.PiecewiseConstant = weekPattern
	if len(oc.dayStatus) > 0 {
		closedDays := map[jtime.LocalDate]ngm.Fraction{}
		openDays := map[jtime.LocalDate]ngm.Fraction{}
		for date, status := range oc.dayStatus {
			if status == DayClose {
				closedDays[date] = ngm.Zero
			} else {
				openDays[date] = ngm.One
			}
		}
		specificDaysClosed := ngm.SpecificDaysOfMap(ngm.One, closedDays)
		specificDaysOpen := ngm.SpecificDaysOfMap(ngm.Zero, openDays)
		result = ngm.CombinerMax(specificDaysOpen, ngm.CombinerProduct(weekPattern, specificDaysClosed))
	}
	if len(oc.workingTimeRanges) > 0 {
		result = ngm.CombinerProduct(result, hoursFromWorkingHours(oc.workingTimeRanges))
	}
	switch {
	case oc.offBefore == nil && oc.offAfter == nil:
	case oc.offBefore == nil:
		result = ngm.CombinerProduct(result,
			ngm.NewTimeWindow(jtime.MinDateTime, oc.offAfter.PlusDays(1).AtStartOfDay()))
	case oc.offAfter == nil:
		result = ngm.CombinerProduct(result, ngm.NewTimeWindow(oc.offBefore.AtStartOfDay(), jtime.MaxDateTime))
	default:
		result = ngm.CombinerProduct(result,
			ngm.NewTimeWindow(oc.offBefore.AtStartOfDay(), oc.offAfter.PlusDays(1).AtStartOfDay()))
	}
	oc.cachedPlan = result
	return result
}

// AddWorkingTimeRange is addWorkingTimeRange, which keeps the cache.
func (oc *OpenClose) AddWorkingTimeRange(timeRange *TimeRange) {
	oc.workingTimeRanges = append(oc.workingTimeRanges, timeRange)
}

// hoursFromWorkingHours is PiecewiseConstantHours.fromWorkingHours, with the
// NullPointerException of a time range whose time did not parse.
func hoursFromWorkingHours(ranges []*TimeRange) ngm.PiecewiseConstant {
	result := ngm.HoursOf(ngm.Zero)
	for _, r := range ranges {
		switch {
		case r.end == nil:
			uml.ThrowNPE(`Cannot invoke "java.time.LocalTime.equals(Object)" because the return value of ` +
				`"net.sourceforge.plantuml.gantt.core.TimeRange.getEnd()" is null`)
		case r.start == nil && *r.end != jtime.Midnight:
			uml.ThrowNPE(`Cannot invoke "java.time.LocalTime.compareTo(java.time.LocalTime)" because the return ` +
				`value of "net.sourceforge.plantuml.gantt.core.TimeRange.getStart()" is null`)
		case r.start == nil:
			uml.ThrowNPE(`Cannot invoke "java.lang.Comparable.compareTo(Object)" because "k1" is null`)
		}
		result = result.With(ngm.TimeRange{Start: *r.start, End: *r.end}, ngm.One)
	}
	return result
}

// DaysAsDates is upstream's DaysAsDates: the dates from date1 to date2,
// both included.
type DaysAsDates struct {
	date1, date2 jtime.LocalDate
}

// NewDaysAsDates is new DaysAsDates(date1, date2).
func NewDaysAsDates(date1, date2 jtime.LocalDate) DaysAsDates { return DaysAsDates{date1, date2} }

// NewDaysAsDatesCount is new DaysAsDates(gantt, date1, count): from date1 to
// the day after the count-th open day, as upstream counts it.
func NewDaysAsDatesCount(gantt *GanttDiagram, date1 jtime.LocalDate, count int) DaysAsDates {
	tmp := date1
	for count > 0 {
		if gantt.IsOpen(tmp) {
			count--
		}
		tmp = tmp.PlusDays(1)
	}
	return DaysAsDates{date1, tmp}
}

// Dates returns the dates in order.
func (d DaysAsDates) Dates() []jtime.LocalDate {
	var result []jtime.LocalDate
	for current := d.date1; current.CompareTo(d.date2) <= 0; current = current.PlusDays(1) {
		result = append(result, current)
	}
	return result
}

// TimeRange is upstream's TimeRange: working hours from a time to another.
// A time that does not parse is nil, as upstream's null, and fails where
// upstream fails.
type TimeRange struct {
	start, end *jtime.LocalTime
}

// NewTimeRange is new TimeRange(start, end) from "H:M" texts, which fails
// with Java's IllegalArgumentException without ":".
func NewTimeRange(start, end string) *TimeRange {
	return &TimeRange{parseTimeRangeTime(start), parseTimeRangeTime(end)}
}

func parseTimeRangeTime(time string) (result *jtime.LocalTime) {
	index := strings.Index(time, ":")
	if index == -1 {
		uml.Throw(uml.IllegalArgumentException, "Invalid time: "+time)
	}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*color.RuntimeError); !ok {
				panic(r)
			}
			result = nil
		}
	}()
	hour := uml.JavaParseInt(javaTrim(time[:index]))
	minute := uml.JavaParseInt(javaTrim(time[index+1:]))
	t := jtime.TimeOf(hour, minute, 0)
	return &t
}

// javaTrim is String.trim: the characters up to the space removed at both
// ends.
func javaTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

// Start is getStart, nil when it did not parse.
func (r *TimeRange) Start() *jtime.LocalTime { return r.start }

// End is getEnd, nil when it did not parse.
func (r *TimeRange) End() *jtime.LocalTime { return r.end }

// String is toString: 08:00-17:30.
func (r *TimeRange) String() string {
	s := func(t *jtime.LocalTime) string {
		if t == nil {
			return "null"
		}
		return t.String()
	}
	return s(r.start) + "-" + s(r.end)
}

// Duration is getDuration: Duration.between(start, end), with its
// NullPointerException for a time that did not parse.
func (r *TimeRange) Duration() jtime.Duration {
	if r.start == nil {
		uml.ThrowNPE(`Cannot invoke "java.time.temporal.Temporal.until(java.time.temporal.Temporal, ` +
			`java.time.temporal.TemporalUnit)" because "startInclusive" is null`)
	}
	if r.end == nil {
		uml.ThrowNPE("temporal")
	}
	return jtime.BetweenTimes(*r.start, *r.end)
}

// TimeRanges is upstream's TimeRanges: the working hours of the project.
type TimeRanges struct {
	timeRanges []*TimeRange
}

// HasWorkingTimeRanges is hasWorkingTimeRanges.
func (tr *TimeRanges) HasWorkingTimeRanges() bool { return len(tr.timeRanges) > 0 }

// Add is add.
func (tr *TimeRanges) Add(timeRange *TimeRange) { tr.timeRanges = append(tr.timeRanges, timeRange) }

// Ranges returns the ranges in order.
func (tr *TimeRanges) Ranges() []*TimeRange { return tr.timeRanges }

// TotalWorkingDuration is totalWorkingDuration: the sum of the durations of
// the ranges, which may be negative.
func (tr *TimeRanges) TotalWorkingDuration() jtime.Duration {
	working := jtime.ZeroDuration
	for _, r := range tr.timeRanges {
		working = working.Plus(r.Duration())
	}
	return working
}

// DurationOfDays is durationOfDays: totalDays days of work, of 24 hours or
// of the working hours.
func (tr *TimeRanges) DurationOfDays(totalDays int) jtime.Duration {
	if tr.HasWorkingTimeRanges() {
		return tr.TotalWorkingDuration().MultipliedBy(int64(totalDays))
	}
	return jtime.DurationOfDays(int64(totalDays))
}

// DayCalendarData is upstream's DayCalendarData: the calendar of the
// project, its colored and named days and its vertical separators.
type DayCalendarData struct {
	openClose               *OpenClose
	nameDays                map[TimePoint]string
	colorDaysToday          map[TimePoint]color.HColor
	colorDaysInternal       map[TimePoint]color.HColor
	colorDaysOfWeek         map[jtime.DayOfWeek]color.HColor
	verticalSeparatorBefore map[jtime.LocalDate]bool
	openCloseForTask        map[string]*OpenClose
	workingTimeRanges       TimeRanges
}

// NewDayCalendarData is new DayCalendarData.
func NewDayCalendarData() *DayCalendarData {
	return &DayCalendarData{
		openClose:               NewOpenClose(),
		nameDays:                map[TimePoint]string{},
		colorDaysToday:          map[TimePoint]color.HColor{},
		colorDaysInternal:       map[TimePoint]color.HColor{},
		colorDaysOfWeek:         map[jtime.DayOfWeek]color.HColor{},
		verticalSeparatorBefore: map[jtime.LocalDate]bool{},
		openCloseForTask:        map[string]*OpenClose{},
	}
}

// IsOpen is isOpen(LocalDate).
func (c *DayCalendarData) IsOpen(day jtime.LocalDate) bool { return c.openClose.LoadAtDUMMY(day) > 0 }

// DayColor is getDayColor: the color of today, else of the date, nil for
// none.
func (c *DayCalendarData) DayColor(day TimePoint) color.HColor {
	if result, ok := c.colorDaysToday[day]; ok && result != nil {
		return result
	}
	return c.colorDaysInternal[day]
}

// DayOfWeekColor is getDayOfWeekColor, nil for none.
func (c *DayCalendarData) DayOfWeekColor(dayOfWeek jtime.DayOfWeek) color.HColor {
	return c.colorDaysOfWeek[dayOfWeek]
}

// DayName is getDayName: the name of the day, ok false for none.
func (c *DayCalendarData) DayName(day TimePoint) (string, bool) {
	name, ok := c.nameDays[day]
	return name, ok
}

// OpenClose is getOpenClose.
func (c *DayCalendarData) OpenClose() *OpenClose { return c.openClose }

// ColorDays is getColorDays: the colored dates, in no order.
func (c *DayCalendarData) ColorDays() []TimePoint {
	result := make([]TimePoint, 0, len(c.colorDaysInternal))
	for day := range c.colorDaysInternal {
		result = append(result, day)
	}
	return result
}

// NameDays is getNameDays: the named dates, in no order.
func (c *DayCalendarData) NameDays() map[TimePoint]string { return c.nameDays }

// HasSeparatorBefore is hasSeparatorBefore.
func (c *DayCalendarData) HasSeparatorBefore(day jtime.LocalDate) bool {
	return c.verticalSeparatorBefore[day]
}

// AddSeparatorBefore is addSeparatorBefore.
func (c *DayCalendarData) AddSeparatorBefore(day jtime.LocalDate) {
	c.verticalSeparatorBefore[day] = true
}

// PutNameDay is putNameDay.
func (c *DayCalendarData) PutNameDay(day TimePoint, name string) { c.nameDays[day] = name }

// PutColorDayToday is putColorDayToday.
func (c *DayCalendarData) PutColorDayToday(day TimePoint, col color.HColor) {
	c.colorDaysToday[day] = col
}

// PutColorDay is putColorDay.
func (c *DayCalendarData) PutColorDay(day TimePoint, col color.HColor) {
	c.colorDaysInternal[day] = col
}

// PutColorDayOfWeek is putColorDayOfWeek.
func (c *DayCalendarData) PutColorDayOfWeek(dow jtime.DayOfWeek, col color.HColor) {
	c.colorDaysOfWeek[dow] = col
}

// OpenCloseForTask is getOpenCloseForTask: the calendar of a task, created
// empty the first time.
func (c *DayCalendarData) OpenCloseForTask(task string) *OpenClose {
	if oc, ok := c.openCloseForTask[task]; ok {
		return oc
	}
	oc := NewOpenClose()
	c.openCloseForTask[task] = oc
	return oc
}

// DefaultPlan is getDefaultPlan.
func (c *DayCalendarData) DefaultPlan() ngm.PiecewiseConstant {
	return c.openClose.AsPiecewiseConstant()
}

// AddWorkingTimeRange is addWorkingTimeRange.
func (c *DayCalendarData) AddWorkingTimeRange(timeRange *TimeRange) {
	c.workingTimeRanges.Add(timeRange)
	c.openClose.AddWorkingTimeRange(timeRange)
}

// HasWorkingTimeRanges is hasWorkingTimeRanges.
func (c *DayCalendarData) HasWorkingTimeRanges() bool {
	return c.workingTimeRanges.HasWorkingTimeRanges()
}

// WorkingTimeRanges returns the working hours.
func (c *DayCalendarData) WorkingTimeRanges() *TimeRanges { return &c.workingTimeRanges }

// DurationOfDays is durationOfDays.
func (c *DayCalendarData) DurationOfDays(totalDays int) jtime.Duration {
	return c.workingTimeRanges.DurationOfDays(totalDays)
}
