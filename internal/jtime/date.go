// SPDX-License-Identifier: MIT
// Provenance: original

// Package jtime reproduces the java.time behavior that the gantt diagrams of
// upstream rely on: dates, times and date-times without a time zone, the
// days of the week and the months, durations, Java's week of year, and the
// DateTimeException and ArithmeticException messages of invalid values and
// overflows. The exceptions are panics with *color.RuntimeError, as for the
// other Java exceptions of the port. It follows the documented semantics of
// java.time, checked against a JDK.
package jtime

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// Java exception classes thrown by this package.
const (
	DateTimeException        = "java.time.DateTimeException"
	ArithmeticException      = "java.lang.ArithmeticException"
	IllegalArgumentException = "java.lang.IllegalArgumentException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

func checkValue(field string, min, max, value int64) {
	checkRange(field, min, max, strconv.FormatInt(min, 10)+" - "+strconv.FormatInt(max, 10), value)
}

// checkRange is ValueRange.checkValidValue, with the range as Java prints
// it ("1 - 28/31" for the day of the month).
func checkRange(field string, min, max int64, valid string, value int64) {
	if value < min || value > max {
		throw(DateTimeException, "Invalid value for "+field+" (valid values "+valid+"): "+strconv.FormatInt(value, 10))
	}
}

// DayOfWeek is java.time.DayOfWeek: Monday is 1 and Sunday is 7.
type DayOfWeek int

// The days of the week, in Java's order.
const (
	Monday DayOfWeek = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

// DaysOfWeek are DayOfWeek.values(), from Monday to Sunday.
var DaysOfWeek = []DayOfWeek{Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday}

var dayNames = [...]string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}

// Value is getValue: 1 for Monday to 7 for Sunday.
func (d DayOfWeek) Value() int { return int(d) }

// Ordinal is the position in DaysOfWeek: 0 for Monday.
func (d DayOfWeek) Ordinal() int { return int(d) - 1 }

// Name is the name of the enum constant, "MONDAY" to "SUNDAY".
func (d DayOfWeek) Name() string { return dayNames[d-1] }

// String is Java's toString, the name of the constant.
func (d DayOfWeek) String() string { return d.Name() }

// Plus is plus: the day of the week days later.
func (d DayOfWeek) Plus(days int64) DayOfWeek {
	return DayOfWeek(floorMod(int64(d.Ordinal())+days%7+7, 7) + 1)
}

// Month is java.time.Month: January is 1 and December is 12.
type Month int

// The months of the year.
const (
	January Month = iota + 1
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

// Months are Month.values(), from January to December.
var Months = []Month{January, February, March, April, May, June, July, August, September, October, November,
	December}

var monthNames = [...]string{"JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE", "JULY", "AUGUST",
	"SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER"}

// Value is getValue: 1 for January to 12 for December.
func (m Month) Value() int { return int(m) }

// Ordinal is the position in Months: 0 for January.
func (m Month) Ordinal() int { return int(m) - 1 }

// Name is the name of the enum constant, "JANUARY" to "DECEMBER".
func (m Month) Name() string { return monthNames[m-1] }

// String is Java's toString, the name of the constant.
func (m Month) String() string { return m.Name() }

// Length is length: the number of days of the month.
func (m Month) Length(leapYear bool) int {
	switch m {
	case February:
		if leapYear {
			return 29
		}
		return 28
	case April, June, September, November:
		return 30
	}
	return 31
}

// IsLeapYear is the proleptic Gregorian leap year rule of
// IsoChronology.isLeapYear.
func IsLeapYear(year int64) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// LocalDate is java.time.LocalDate: a date of the proleptic Gregorian
// calendar without a time zone.
type LocalDate struct {
	year  int
	month Month
	day   int
}

// The extreme dates and the epoch.
var (
	MinDate = LocalDate{-999999999, January, 1}
	MaxDate = LocalDate{999999999, December, 31}
	Epoch   = LocalDate{1970, January, 1}
)

// DateOf is LocalDate.of(year, month, dayOfMonth), with Java's validation
// and messages.
func DateOf(year, month, dayOfMonth int) LocalDate {
	checkValue("Year", -999999999, 999999999, int64(year))
	checkValue("MonthOfYear", 1, 12, int64(month))
	checkRange("DayOfMonth", 1, 31, "1 - 28/31", int64(dayOfMonth))
	if dayOfMonth > 28 {
		m := Month(month)
		if dayOfMonth > m.Length(IsLeapYear(int64(year))) {
			if dayOfMonth == 29 {
				throw(DateTimeException, "Invalid date 'February 29' as '"+strconv.Itoa(year)+"' is not a leap year")
			}
			throw(DateTimeException, "Invalid date '"+m.Name()+" "+strconv.Itoa(dayOfMonth)+"'")
		}
	}
	return LocalDate{year, Month(month), dayOfMonth}
}

// DateOfEpochDay is LocalDate.ofEpochDay: the date epochDay days after
// 1970-01-01.
func DateOfEpochDay(epochDay int64) LocalDate {
	// Days from 0000-03-01, in eras of 400 years (146097 days).
	z := epochDay + 719468
	era := floorDiv(z, 146097)
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	y := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	d := doy - (153*mp+2)/5 + 1
	m := mp + 3
	if m > 12 {
		m -= 12
	}
	if m <= 2 {
		y++
	}
	checkValue("Year", -999999999, 999999999, y)
	return LocalDate{int(y), Month(m), int(d)}
}

// EpochDay is toEpochDay: the number of days since 1970-01-01.
func (d LocalDate) EpochDay() int64 {
	y := int64(d.year)
	m := int64(d.month)
	if m <= 2 {
		y--
	}
	era := floorDiv(y, 400)
	yoe := y - era*400
	mp := (m + 9) % 12
	doy := (153*mp+2)/5 + int64(d.day) - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

// Year is getYear.
func (d LocalDate) Year() int { return d.year }

// Month is getMonth.
func (d LocalDate) Month() Month { return d.month }

// MonthValue is getMonthValue, 1 to 12.
func (d LocalDate) MonthValue() int { return int(d.month) }

// DayOfMonth is getDayOfMonth.
func (d LocalDate) DayOfMonth() int { return d.day }

// IsLeapYear is isLeapYear.
func (d LocalDate) IsLeapYear() bool { return IsLeapYear(int64(d.year)) }

// LengthOfMonth is lengthOfMonth.
func (d LocalDate) LengthOfMonth() int { return d.month.Length(d.IsLeapYear()) }

// DayOfYear is getDayOfYear, 1 for January 1.
func (d LocalDate) DayOfYear() int {
	return int(d.EpochDay()-LocalDate{d.year, January, 1}.EpochDay()) + 1
}

// DayOfWeek is getDayOfWeek.
func (d LocalDate) DayOfWeek() DayOfWeek {
	// 1970-01-01 is a Thursday.
	return DayOfWeek(floorMod(d.EpochDay()+3, 7) + 1)
}

// PlusDays is plusDays.
func (d LocalDate) PlusDays(days int64) LocalDate {
	if days == 0 {
		return d
	}
	return DateOfEpochDay(addExact(d.EpochDay(), days))
}

// MinusDays is minusDays.
func (d LocalDate) MinusDays(days int64) LocalDate {
	if days == minLong {
		return d.PlusDays(maxLong).PlusDays(1)
	}
	return d.PlusDays(-days)
}

// CompareTo is compareTo: negative, zero or positive.
func (d LocalDate) CompareTo(other LocalDate) int {
	switch {
	case d.year != other.year:
		return cmpInt(d.year, other.year)
	case d.month != other.month:
		return cmpInt(int(d.month), int(other.month))
	}
	return cmpInt(d.day, other.day)
}

// IsBefore is isBefore.
func (d LocalDate) IsBefore(other LocalDate) bool { return d.CompareTo(other) < 0 }

// IsAfter is isAfter.
func (d LocalDate) IsAfter(other LocalDate) bool { return d.CompareTo(other) > 0 }

// AtStartOfDay is atStartOfDay: midnight at the start of the date.
func (d LocalDate) AtStartOfDay() LocalDateTime { return LocalDateTime{d, Midnight} }

// AtTime is atTime.
func (d LocalDate) AtTime(t LocalTime) LocalDateTime { return LocalDateTime{d, t} }

// String is toString, in ISO format: 2020-01-06, +12345-01-01, -0005-01-01.
func (d LocalDate) String() string {
	var s string
	switch y := d.year; {
	case y < 0 && y > -1000:
		s = "-" + pad(-y, 4)
	case y >= 0 && y < 1000:
		s = pad(y, 4)
	case y > 9999:
		s = "+" + strconv.Itoa(y)
	default:
		s = strconv.Itoa(y)
	}
	return s + "-" + pad(int(d.month), 2) + "-" + pad(d.day, 2)
}

// DaysBetween is ChronoUnit.DAYS.between(start, end).
func DaysBetween(start, end LocalDate) int64 { return end.EpochDay() - start.EpochDay() }

// YearMonth is java.time.YearMonth.
type YearMonth struct {
	Year  int
	Month Month
}

// YearMonthFrom is YearMonth.from for a date.
func YearMonthFrom(d LocalDate) YearMonth { return YearMonth{d.year, d.month} }

// WeekFields is java.time.temporal.WeekFields: the first day of the week
// and the minimal number of days of the first week of a year.
type WeekFields struct {
	firstDayOfWeek DayOfWeek
	minimalDays    int
}

// WeekFieldsOf is WeekFields.of, which fails with Java's
// IllegalArgumentException when minimalDays is outside 1 to 7.
func WeekFieldsOf(firstDayOfWeek DayOfWeek, minimalDays int) WeekFields {
	if minimalDays < 1 || minimalDays > 7 {
		throw(IllegalArgumentException, "Minimal number of days is invalid")
	}
	return WeekFields{firstDayOfWeek, minimalDays}
}

// WeekOfYear is date.get(weekOfYear()): the weeks of a year start on the
// first day of the week, week 1 is the first week with at least the
// minimal number of days in the year, and the days before it are in week 0.
// The last days of a year stay in that year (week 52 or 53), unlike ISO
// weeks.
func (w WeekFields) WeekOfYear(date LocalDate) int {
	jan1 := LocalDate{date.year, January, 1}
	// Day of the week of January 1, 1 for the first day of the week.
	jan1Dow := int(floorMod(int64(jan1.DayOfWeek()-w.firstDayOfWeek), 7)) + 1
	// Day of the year of the first day of week 1, 1 or less when the week
	// that holds January 1 is week 1.
	week1 := 2 - jan1Dow
	if 8-jan1Dow < w.minimalDays {
		week1 += 7
	}
	return int(floorDiv(int64(date.DayOfYear()-week1), 7)) + 1
}

const (
	minLong = -1 << 63
	maxLong = 1<<63 - 1
)

func floorDiv(x, y int64) int64 {
	q := x / y
	if (x%y != 0) && ((x < 0) != (y < 0)) {
		q--
	}
	return q
}

func floorMod(x, y int64) int64 { return x - floorDiv(x, y)*y }

// addExact is Math.addExact for longs.
func addExact(x, y int64) int64 {
	r := x + y
	if (x^r)&(y^r) < 0 {
		throw(ArithmeticException, "long overflow")
	}
	return r
}

// multiplyExact is Math.multiplyExact for longs.
func multiplyExact(x, y int64) int64 {
	if x == 0 || y == 0 {
		return 0
	}
	r := x * y
	if r/y != x || (x == -1 && y == minLong) || (y == -1 && x == minLong) {
		throw(ArithmeticException, "long overflow")
	}
	return r
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func pad(v, width int) string {
	s := strconv.Itoa(v)
	for len(s) < width {
		s = "0" + s
	}
	return s
}
