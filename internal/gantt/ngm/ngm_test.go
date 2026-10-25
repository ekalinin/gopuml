// SPDX-License-Identifier: MIT
// Provenance: original

package ngm

import (
	"fmt"
	"testing"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The expected values of these tests were printed by upstream's classes of
// the reference jar.

func panicMessage(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*color.RuntimeError)
			if !ok {
				t.Fatalf("panic %v is not a Java exception", r)
			}
			msg = e.Error()
		}
	}()
	f()
	return ""
}

func dt(y, m, d, h, mi int) jtime.LocalDateTime { return jtime.DateTimeOfFields(y, m, d, h, mi, 0) }

func TestFraction(t *testing.T) {
	for _, tc := range []struct{ got, want string }{
		{NewFraction(6, -4).String(), "-3/2"},
		{NewFraction(0, 7).String(), "0"},
		{fmt.Sprint(NewFraction(7, 2).WholePart(), NewFraction(-7, 2).WholePart()), "3 -3"},
		{fmt.Sprint(NewFraction(1, 3).CompareTo(NewFraction(2, 6)), NewFraction(1, 3).CompareTo(NewFraction(1, 2))), "0 -1"},
		{NewFraction(1, 3).Add(NewFraction(1, 6)).String(), "1/2"},
		{NewFraction(1, 3).Subtract(NewFraction(1, 2)).String(), "-1/6"},
		{NewFraction(2, 3).Divide(NewFraction(4, 9)).String(), "3/2"},
	} {
		if tc.got != tc.want {
			t.Errorf("%q, want %q", tc.got, tc.want)
		}
	}
	if NewFraction(0, 7) != Zero || NewFraction(-4, -4) != One {
		t.Error("reduced fractions differ from the constants")
	}
	if got := panicMessage(t, func() { NewFraction(1, 0) }); got != "java.lang.IllegalArgumentException: Denominator cannot be zero." {
		t.Errorf("%q", got)
	}
}

func TestSegmentErrors(t *testing.T) {
	jan1, jan2, jan3, jan4 := dt(2020, 1, 1, 0, 0), dt(2020, 1, 2, 0, 0), dt(2020, 1, 3, 0, 0), dt(2020, 1, 4, 0, 0)
	for _, tc := range []struct {
		f    func()
		want string
	}{
		{func() { ForwardSegment(jan2, jan1, One) }, "end must be after start"},
		{func() { BackwardSegment(jan1, jan2, One) }, "end must be before start"},
		{func() { Intersection(nil, Product) }, "No segments to intersect"},
		{func() {
			Intersection([]Segment{ForwardSegment(jan1, jan2, One), BackwardSegment(jan2, jan1, One)}, Product)
		},
			"All segments must have the same direction"},
		{func() {
			Intersection([]Segment{ForwardSegment(jan1, jan2, One), ForwardSegment(jan3, jan4, One)}, Product)
		},
			"Segments do not overlap"},
		{func() { ForwardSegment(jan1, jan2, One).Split(jan3) }, "time must be within the segment bounds"},
		{func() { NewTimeWindow(jan1, jan1) }, "offAfter must be after offBefore"},
		{func() { HoursOf(Zero).With(TimeRange{jtime.TimeOf(9, 0, 0), jtime.TimeOf(9, 0, 0)}, One) },
			"start must be strictly before end"},
		{func() { LoadOf(jtime.DurationOfHours(-10)) }, "Effort cannot be negative: PT-10H"},
	} {
		if got := panicMessage(t, tc.f); got != "java.lang.IllegalArgumentException: "+tc.want {
			t.Errorf("%q, want %q", got, tc.want)
		}
	}
	got := Intersection([]Segment{ForwardSegment(jan1, jan3, NewFraction(1, 2)), ForwardSegment(jan2, jan4, NewFraction(2, 3))}, Product)
	if want := "FORWARD ]2020-01-02T00:00, 2020-01-03T00:00[ value=1/3"; got.String() != want {
		t.Errorf("%q, want %q", got, want)
	}
}

func TestSegmentsAt(t *testing.T) {
	h := HoursFromWorkingHours([]TimeRange{{jtime.TimeOf(9, 0, 0), jtime.TimeOf(12, 0, 0)},
		{jtime.TimeOf(14, 0, 0), jtime.TimeOf(18, 0, 0)}})
	we := WeekdayOf(One).With(jtime.Saturday, Zero).With(jtime.Sunday, Zero)
	for _, tc := range []struct {
		f       PiecewiseConstant
		instant jtime.LocalDateTime
		dir     TimeDirection
		want    string
	}{
		{h, dt(2020, 1, 6, 0, 0), Forward, "FORWARD ]2020-01-06T00:00, 2020-01-06T09:00[ value=0"},
		{h, dt(2020, 1, 6, 10, 0), Forward, "FORWARD ]2020-01-06T09:00, 2020-01-06T12:00[ value=1"},
		{h, dt(2020, 1, 6, 18, 0), Forward, "FORWARD ]2020-01-06T18:00, 2020-01-07T00:00[ value=0"},
		{h, dt(2020, 1, 6, 0, 0), Backward, "BACKWARD ]2020-01-06T00:00, 2020-01-05T18:00[ value=0"},
		{h, dt(2020, 1, 6, 10, 0), Backward, "BACKWARD ]2020-01-06T12:00, 2020-01-06T09:00[ value=1"},
		{h, dt(2020, 1, 6, 12, 0), Backward, "BACKWARD ]2020-01-06T12:00, 2020-01-06T09:00[ value=1"},
		{h, dt(2020, 1, 6, 8, 0), Backward, "BACKWARD ]2020-01-06T09:00, 2020-01-06T00:00[ value=0"},
		{HoursOf(Zero).With(TimeRange{jtime.TimeOf(18, 0, 0), jtime.Midnight}, One), dt(2020, 1, 6, 0, 0), Backward,
			"BACKWARD ]2020-01-06T00:00, 2020-01-05T18:00[ value=1"},
		{we, dt(2020, 1, 6, 0, 0), Backward, "BACKWARD ]2020-01-06T00:00, 2020-01-05T00:00[ value=0"},
		{CombinerSum(SpecificDaysOf(NewFraction(1, 2)), SpecificDaysOf(NewFraction(1, 3))), dt(2020, 1, 6, 5, 0), Forward,
			"FORWARD ]2020-01-06T00:00, 2020-01-07T00:00[ value=5/6"},
	} {
		if got := tc.f.SegmentAt(tc.instant, tc.dir).String(); got != tc.want {
			t.Errorf("%v %v: %q, want %q", tc.instant, tc.dir, got, tc.want)
		}
	}
	if !IsZeroOnDay(we, jtime.DateOf(2020, 1, 11)) || IsZeroOnDay(we, jtime.DateOf(2020, 1, 10)) ||
		IsZeroOnDay(CombinerProduct(we, h), jtime.DateOf(2020, 1, 10)) {
		t.Error("isZeroOnDay")
	}
}

func TestIntegration(t *testing.T) {
	h := HoursFromWorkingHours([]TimeRange{{jtime.TimeOf(9, 0, 0), jtime.TimeOf(12, 0, 0)},
		{jtime.TimeOf(14, 0, 0), jtime.TimeOf(18, 0, 0)}})
	we := WeekdayOf(One).With(jtime.Saturday, Zero).With(jtime.Sunday, Zero)
	days := func(n int64) Load { return LoadOf(jtime.DurationOfDays(n)) }
	for _, tc := range []struct {
		name string
		got  func() jtime.LocalDateTime
		want string
	}{
		{"forward across a weekend", func() jtime.LocalDateTime {
			return NewLoadIntegrator(we, days(3)).ComputeEnd(dt(2020, 1, 9, 0, 0))
		}, "2020-01-14T00:00"},
		{"backward across a weekend", func() jtime.LocalDateTime {
			return NewLoadIntegrator(we, days(3)).ComputeStart(dt(2020, 1, 14, 0, 0))
		}, "2020-01-09T00:00"},
		{"forward with hours", func() jtime.LocalDateTime {
			return NewLoadIntegrator(CombinerProduct(we, h), LoadOf(jtime.DurationOfHours(10))).ComputeEnd(dt(2020, 1, 10, 10, 0))
		}, "2020-01-13T15:00"},
		{"backward with hours", func() jtime.LocalDateTime {
			return NewLoadIntegrator(CombinerProduct(we, h), LoadOf(jtime.DurationOfHours(10))).ComputeStart(dt(2020, 1, 13, 10, 0))
		}, "2020-01-09T16:00"},
		{"33 percent", func() jtime.LocalDateTime {
			return NewLoadIntegrator(CombinerProduct(we, SpecificDaysOf(NewFraction(33, 100))), days(1)).ComputeEnd(dt(2020, 1, 6, 0, 0))
		}, "2020-01-09T00:43:38"},
		{"time window", func() jtime.LocalDateTime {
			return NewLoadIntegrator(CombinerProduct(we, NewTimeWindow(dt(2020, 1, 8, 0, 0), jtime.MaxDateTime)), days(2)).ComputeEnd(dt(2020, 1, 6, 0, 0))
		}, "2020-01-10T00:00"},
		{"time window backward", func() jtime.LocalDateTime {
			return NewLoadIntegrator(CombinerProduct(we, NewTimeWindow(jtime.MinDateTime, dt(2020, 1, 8, 0, 0))), days(2)).ComputeStart(dt(2020, 1, 10, 0, 0))
		}, "2020-01-06T00:00"},
	} {
		if got := tc.got().String(); got != tc.want {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
	task := TaskWithFixedTotalEffort(AllocationOf(we), dt(2020, 1, 9, 0, 0), days(3))
	end := task.End()
	task.SetEnd(dt(2020, 1, 14, 0, 0))
	if got := fmt.Sprint(end, " ", task.Start(), " ", task.Duration()); got != "2020-01-14T00:00 2020-01-09T00:00 PT120H" {
		t.Errorf("task %q", got)
	}
	for _, f := range []func(){
		func() { NewLoadIntegrator(WeekdayOf(Zero), days(1)).ComputeEnd(dt(2020, 1, 6, 0, 0)) },
		func() { NewLoadIntegrator(we, days(7200)).ComputeEnd(dt(2020, 1, 6, 0, 0)) },
	} {
		if got := panicMessage(t, f); got != "java.lang.IllegalStateException: Infinite loop detected in PiecewiseConstant iteration" {
			t.Errorf("%q", got)
		}
	}
}
