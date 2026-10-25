// SPDX-License-Identifier: MIT
// Provenance: original

package jtime

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// panicMessage runs f and returns the Java exception it panics with, as
// Throwable.toString prints it, or "" when it does not panic.
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

func TestEpochDay(t *testing.T) {
	for _, tc := range []struct {
		date LocalDate
		day  int64
	}{
		{Epoch, 0},
		{DateOf(1969, 12, 31), -1},
		{DateOf(2000, 2, 29), 11016},
		{DateOf(2000, 3, 1), 11017},
		{DateOf(2020, 1, 6), 18267},
		{DateOf(1900, 3, 1), -25508},
		{MinDate, -365243219162},
		{MaxDate, 365241780471},
	} {
		if got := tc.date.EpochDay(); got != tc.day {
			t.Errorf("%v: epoch day %d, want %d", tc.date, got, tc.day)
		}
		if got := DateOfEpochDay(tc.day); got != tc.date {
			t.Errorf("day %d: %v, want %v", tc.day, got, tc.date)
		}
	}
	// Every day of 400 years round-trips and follows the previous one.
	prev := DateOf(1899, 12, 31)
	for day := prev.EpochDay() + 1; day < DateOf(2300, 1, 1).EpochDay(); day++ {
		d := DateOfEpochDay(day)
		if d.EpochDay() != day || !d.IsAfter(prev) || d.DayOfWeek() != prev.DayOfWeek().Plus(1) {
			t.Fatalf("day %d: %v after %v", day, d, prev)
		}
		prev = d
	}
}

func TestDateFields(t *testing.T) {
	d := DateOf(2024, 12, 30)
	if d.DayOfWeek() != Monday || d.DayOfYear() != 365 || d.Month() != December || d.LengthOfMonth() != 31 {
		t.Errorf("%v: %v, day %d of the year", d, d.DayOfWeek(), d.DayOfYear())
	}
	if got := DateOf(2019, 12, 31).PlusDays(1); got != DateOf(2020, 1, 1) {
		t.Errorf("plusDays across a year: %v", got)
	}
	if got := DateOf(2020, 3, 1).MinusDays(1); got != DateOf(2020, 2, 29) {
		t.Errorf("minusDays in a leap year: %v", got)
	}
	if got := DaysBetween(DateOf(2020, 1, 6), DateOf(2019, 12, 31)); got != -6 {
		t.Errorf("days between %d", got)
	}
	for _, tc := range []struct {
		date LocalDate
		want string
	}{
		{DateOf(2020, 1, 6), "2020-01-06"},
		{DateOf(12345, 1, 1), "+12345-01-01"},
		{DateOf(-5, 1, 1), "-0005-01-01"},
		{DateOf(-12345, 1, 1), "-12345-01-01"},
		{DateOf(5, 1, 1), "0005-01-01"},
	} {
		if got := tc.date.String(); got != tc.want {
			t.Errorf("%q, want %q", got, tc.want)
		}
	}
	if Sunday.Plus(1) != Monday || Monday.Plus(-1) != Sunday || Wednesday.Plus(-15) != Tuesday {
		t.Error("DayOfWeek.plus")
	}
}

func TestDateErrors(t *testing.T) {
	for _, tc := range []struct {
		f    func()
		want string
	}{
		{func() { DateOf(2020, 2, 30) }, "java.time.DateTimeException: Invalid date 'FEBRUARY 30'"},
		{func() { DateOf(2019, 2, 29) }, "java.time.DateTimeException: Invalid date 'February 29' as '2019' is not a leap year"},
		{func() { DateOf(2019, 4, 31) }, "java.time.DateTimeException: Invalid date 'APRIL 31'"},
		{func() { DateOf(2019, 13, 1) }, "java.time.DateTimeException: Invalid value for MonthOfYear (valid values 1 - 12): 13"},
		{func() { DateOf(2019, 0, 1) }, "java.time.DateTimeException: Invalid value for MonthOfYear (valid values 1 - 12): 0"},
		{func() { DateOf(2019, 1, 32) }, "java.time.DateTimeException: Invalid value for DayOfMonth (valid values 1 - 28/31): 32"},
		{func() { DateOf(2019, 1, 0) }, "java.time.DateTimeException: Invalid value for DayOfMonth (valid values 1 - 28/31): 0"},
		{func() { DateOf(1000000000, 1, 1) }, "java.time.DateTimeException: Invalid value for Year (valid values -999999999 - 999999999): 1000000000"},
		{func() { MaxDate.PlusDays(1) }, "java.time.DateTimeException: Invalid value for Year (valid values -999999999 - 999999999): 1000000000"},
		{func() { MaxDateTime.PlusSeconds(1) }, "java.time.DateTimeException: Invalid value for Year (valid values -999999999 - 999999999): 1000000000"},
		{func() { TimeOf(25, 0, 0) }, "java.time.DateTimeException: Invalid value for HourOfDay (valid values 0 - 23): 25"},
		{func() { TimeOf(8, 60, 0) }, "java.time.DateTimeException: Invalid value for MinuteOfHour (valid values 0 - 59): 60"},
		{func() { TimeOf(8, 0, 60) }, "java.time.DateTimeException: Invalid value for SecondOfMinute (valid values 0 - 59): 60"},
		{func() { DateTimeOfFields(2020, 2, 30, 25, 0, 0) }, "java.time.DateTimeException: Invalid date 'FEBRUARY 30'"},
		{func() { MaxDateTime.EpochMilli() }, "java.lang.ArithmeticException: long overflow"},
		{func() { WeekFieldsOf(Monday, 0) }, "java.lang.IllegalArgumentException: Minimal number of days is invalid"},
		{func() { WeekFieldsOf(Monday, 8) }, "java.lang.IllegalArgumentException: Minimal number of days is invalid"},
	} {
		if got := panicMessage(t, tc.f); got != tc.want {
			t.Errorf("%q, want %q", got, tc.want)
		}
	}
}

func TestDateTime(t *testing.T) {
	dt := DateTimeOfFields(2024, 9, 22, 10, 30, 46)
	for _, tc := range []struct {
		got, want string
	}{
		{dt.String(), "2024-09-22T10:30:46"},
		{DateTimeOfFields(2024, 9, 22, 10, 30, 0).String(), "2024-09-22T10:30"},
		{TimeOf(8, 0, 0).String(), "08:00"},
		{MaxTime.String(), "23:59:59.999999999"},
		{MinDateTime.String(), "-999999999-01-01T00:00"},
		{MaxDateTime.String(), "+999999999-12-31T23:59:59.999999999"},
		{TimeOfNanoOfDay(5_000_000).String(), "00:00:00.005"},
		{dt.PlusSeconds(86400*2 + 3600).String(), "2024-09-24T11:30:46"},
		{dt.MinusSeconds(11 * 3600).String(), "2024-09-21T23:30:46"},
		{dt.Plus(DurationOfHours(-34)).String(), "2024-09-21T00:30:46"},
		{dt.Minus(DurationOfSeconds(-1, 500_000_000)).String(), "2024-09-22T10:30:46.500"},
		{DateTimeOfEpochSecond(-1, 500_000_000).String(), "1969-12-31T23:59:59.500"},
	} {
		if tc.got != tc.want {
			t.Errorf("%q, want %q", tc.got, tc.want)
		}
	}
	if got := DateTimeOfEpochSecond(-1, 500_000_000).EpochMilli(); got != -500 {
		t.Errorf("epoch milli %d", got)
	}
	if got := DateTimeOfFields(2020, 1, 6, 0, 0, 0).EpochMilli(); got != 18267*86400000 {
		t.Errorf("epoch milli %d", got)
	}
	if !MinDateTime.IsBefore(dt) || !MaxDateTime.IsAfter(dt) || dt.CompareTo(dt) != 0 {
		t.Error("compareTo")
	}
}

func TestDuration(t *testing.T) {
	for _, tc := range []struct {
		d    Duration
		want string
	}{
		{DurationOfHours(25).PlusMinutes(1).PlusSeconds(1), "PT25H1M1S"},
		{DurationOfHours(-10), "PT-10H"},
		{ZeroDuration, "PT0S"},
		{DurationOfSeconds(-1, 500_000_000), "PT-0.5S"},
		{DurationOfSeconds(59, 0), "PT59S"},
		{DurationOfSeconds(0, 1), "PT0.000000001S"},
		{DurationOfDays(-1589934592), "PT-38158430208H"},
		{DurationOfSeconds(-90, 0), "PT-1M-30S"},
		{Between(DateTimeOfFields(2020, 1, 1, 0, 0, 0), MaxDateTime), "PT8765802293039H59M59.999999999S"},
		{BetweenTimes(TimeOf(18, 0, 0), MaxTime), "PT5H59M59.999999999S"},
		{DurationOfDays(2).MultipliedBy(-3), "PT-144H"},
		{DurationOfSeconds(-1, 500_000_000).Plus(DurationOfSeconds(0, 600_000_000)), "PT0.1S"},
	} {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("%q, want %q", got, tc.want)
		}
	}
	if d := BetweenTimes(TimeOf(18, 0, 0), MaxTime); d.Seconds() != 21599 {
		t.Errorf("seconds %d", d.Seconds())
	}
	if d := DurationOfSeconds(-1, 500_000_000); d.Seconds() != -1 || d.ToMillis() != -500 || !d.IsNegative() {
		t.Errorf("%v: seconds %d, millis %d", d, d.Seconds(), d.ToMillis())
	}
	if DurationOfDays(1).CompareTo(DurationOfHours(24)) != 0 || DurationOfHours(1).CompareTo(DurationOfSeconds(3600, 1)) >= 0 {
		t.Error("compareTo")
	}
	for _, tc := range []struct {
		f    func()
		want string
	}{
		{func() { DurationOfSeconds(maxLong, 0).ToMillis() }, "java.lang.ArithmeticException: long overflow"},
		{func() { DurationOfDays(1).MultipliedBy(maxLong) }, "java.lang.ArithmeticException: Exceeds capacity of Duration: 796899343984252629724800000000000"},
		{func() { DurationOfDays(maxLong) }, "java.lang.ArithmeticException: long overflow"},
	} {
		if got := panicMessage(t, tc.f); got != tc.want {
			t.Errorf("%q, want %q", got, tc.want)
		}
	}
}

func TestWeekOfYear(t *testing.T) {
	for _, tc := range []struct {
		date             string
		mon4, sun1, sat7 int
	}{
		{"2024-12-23", 52, 52, 51},
		{"2024-12-30", 53, 53, 52},
		{"2025-01-06", 2, 2, 1},
		{"2027-01-01", 0, 1, 0},
		{"2026-01-01", 1, 1, 0},
		{"2020-01-05", 1, 2, 1},
		{"2020-01-12", 2, 3, 2},
		{"2021-01-03", 0, 2, 1},
	} {
		var y, m, d int
		for i, p := range []*int{&y, &m, &d} {
			*p = atoi(tc.date[[]int{0, 5, 8}[i]:[]int{4, 7, 10}[i]])
		}
		date := DateOf(y, m, d)
		got := []int{WeekFieldsOf(Monday, 4).WeekOfYear(date), WeekFieldsOf(Sunday, 1).WeekOfYear(date),
			WeekFieldsOf(Saturday, 7).WeekOfYear(date)}
		if got[0] != tc.mon4 || got[1] != tc.sun1 || got[2] != tc.sat7 {
			t.Errorf("%s: weeks %v, want %d %d %d", tc.date, got, tc.mon4, tc.sun1, tc.sat7)
		}
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
