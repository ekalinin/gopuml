// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/time/TimePoint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/time/TimePointFormat.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/time/MonthUtils.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/time/DayOfWeekUtils.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/time/WeekNumberStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/PrintScale.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"
	"strings"
	"time"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// now is the clock of the current day (TimePoint.todayAtMidnight). Tests
// replace it.
var now = time.Now

// Locale is the part of java.util.Locale that gantt diagrams use: the
// language, lower case, as new Locale(language) keeps it.
type Locale string

// LocaleEnglish is Locale.ENGLISH, the default language of gantt diagrams.
const LocaleEnglish Locale = "en"

// NewLocale is new Locale(language): the language in ASCII lower case, with
// the obsolete codes that Java replaces.
func NewLocale(language string) Locale {
	b := []byte(language)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	switch l := string(b); l {
	case "iw":
		return "he"
	case "ji":
		return "yi"
	case "in":
		return "id"
	default:
		return Locale(l)
	}
}

// PrintScale is upstream's PrintScale: the unit of the calendar.
type PrintScale int

// The print scales, with their compression of a day.
const (
	PrintScaleDaily PrintScale = iota
	PrintScaleWeekly
	PrintScaleMonthly
	PrintScaleQuarterly
	PrintScaleYearly
)

var printScaleCompress = [...]int{1, 4, 15, 40, 60}

// DefaultScale is getDefaultScale: 1 divided by the compression.
func (p PrintScale) DefaultScale() float64 { return 1.0 / float64(printScaleCompress[p]) }

// PrintScaleFromString is PrintScale.fromString: the scale named by the
// first letter of value, which is case-sensitive, daily otherwise.
func PrintScaleFromString(value string) PrintScale {
	switch {
	case strings.HasPrefix(value, "w"):
		return PrintScaleWeekly
	case strings.HasPrefix(value, "m"):
		return PrintScaleMonthly
	case strings.HasPrefix(value, "q"):
		return PrintScaleQuarterly
	case strings.HasPrefix(value, "y"):
		return PrintScaleYearly
	}
	return PrintScaleDaily
}

// WeekNumberStrategy is upstream's WeekNumberStrategy: the first day of a
// week and the minimal number of days of the first week of a year.
type WeekNumberStrategy struct {
	FirstDayOfWeek         jtime.DayOfWeek
	MinimalDaysInFirstWeek int
}

const millisecondsPerDay = 1000 * 3600 * 24

// TimePoint is upstream's TimePoint: a date and a time of day, in UTC.
type TimePoint struct {
	dt jtime.LocalDateTime
}

// TimePointEpoch is TimePoint.epoch(): 1970-01-01, the start of a project
// without a start.
func TimePointEpoch() jtime.LocalDate { return jtime.Epoch }

// TimePointOf is TimePoint.of.
func TimePointOf(dt jtime.LocalDateTime) TimePoint { return TimePoint{dt} }

// TimePointOfStartOfDay is TimePoint.ofStartOfDay(LocalDate).
func TimePointOfStartOfDay(day jtime.LocalDate) TimePoint { return TimePoint{day.AtStartOfDay()} }

// TimePointOfStartOfDayYMD is TimePoint.ofStartOfDay(year, month,
// dayOfMonth), with Java's DateTimeException for an invalid date.
func TimePointOfStartOfDayYMD(year, month, dayOfMonth int) TimePoint {
	return TimePoint{jtime.DateTimeOfFields(year, month, dayOfMonth, 0, 0, 0)}
}

// TimePointOfStartOfDayMonthName is TimePoint.ofStartOfDay(year, month,
// dayOfMonth) with the name of the month.
func TimePointOfStartOfDayMonthName(year int, month string, dayOfMonth int) TimePoint {
	return TimePointOfStartOfDayYMD(year, MonthFromString(month).Value(), dayOfMonth)
}

// TimePointOfEndOfDayMinusOneSecond is TimePoint.ofEndOfDayMinusOneSecond.
func TimePointOfEndOfDayMinusOneSecond(day jtime.LocalDate) TimePoint {
	return TimePointOfStartOfDay(day.PlusDays(1)).MinusOneSecond()
}

// todayAtMidnight is TimePoint.todayAtMidnight: the start of the current
// day, in the local time zone.
func todayAtMidnight() TimePoint {
	t := now().Local()
	return TimePointOfStartOfDay(jtime.DateOf(t.Year(), int(t.Month()), t.Day()))
}

// LocalDateTime is toLocalDateTime.
func (t TimePoint) LocalDateTime() jtime.LocalDateTime { return t.dt }

// StringShort is toStringShort: the month and the day in the language.
func (t TimePoint) StringShort(locale Locale) string { return FormatMonthAndDay.Format(t, locale) }

// WeekOfYear is getWeekOfYear.
func (t TimePoint) WeekOfYear(strategy WeekNumberStrategy) int {
	return jtime.WeekFieldsOf(strategy.FirstDayOfWeek, strategy.MinimalDaysInFirstWeek).WeekOfYear(t.dt.LocalDate())
}

// Increment is increment: the next day, at the same time.
func (t TimePoint) Increment() TimePoint { return t.AddDays(1) }

// Decrement is decrement: the day before, at the same time.
func (t TimePoint) Decrement() TimePoint { return t.AddDays(-1) }

// AddDays is addDays.
func (t TimePoint) AddDays(nday int) TimePoint { return TimePoint{t.dt.PlusDays(int64(nday))} }

// AbsoluteDayNum is getAbsoluteDayNum: the number of days since
// 1970-01-01, rounded down.
func (t TimePoint) AbsoluteDayNum() int { return int(int32(floorDiv(t.Millis(), millisecondsPerDay))) }

// Millis is getMillis: the milliseconds since 1970-01-01.
func (t TimePoint) Millis() int64 { return t.dt.EpochMilli() }

// Year is year.
func (t TimePoint) Year() int { return t.dt.Year() }

// Quarter is quarter: Q1 to Q4.
func (t TimePoint) Quarter() string { return "Q" + strconv.Itoa((t.dt.Month().Ordinal()+3)/3) }

// String is toString: the date-time in ISO format.
func (t TimePoint) String() string { return t.dt.String() }

// CompareTo is compareTo.
func (t TimePoint) CompareTo(other TimePoint) int { return t.dt.CompareTo(other.dt) }

// DayOfMonth is getDayOfMonth.
func (t TimePoint) DayOfMonth() int { return t.dt.DayOfMonth() }

// Month is month.
func (t TimePoint) Month() jtime.Month { return t.dt.Month() }

// MonthYear is monthYear.
func (t TimePoint) MonthYear() jtime.YearMonth { return jtime.YearMonthFrom(t.dt.LocalDate()) }

// DayOfWeek is toDayOfWeek.
func (t TimePoint) DayOfWeek() jtime.DayOfWeek { return t.dt.DayOfWeek() }

// Day is toDay: the date.
func (t TimePoint) Day() jtime.LocalDate { return t.dt.LocalDate() }

// MinTimePoint is TimePoint.min.
func MinTimePoint(d1, d2 TimePoint) TimePoint {
	if d1.CompareTo(d2) <= 0 {
		return d1
	}
	return d2
}

// MaxTimePoint is TimePoint.max.
func MaxTimePoint(d1, d2 TimePoint) TimePoint {
	if d1.CompareTo(d2) >= 0 {
		return d1
	}
	return d2
}

// IncrementScale is increment(PrintScale): a week later for the weekly
// scale, a day later otherwise.
func (t TimePoint) IncrementScale(printScale PrintScale) TimePoint {
	if printScale == PrintScaleWeekly {
		return t.AddDays(7)
	}
	return t.Increment()
}

// MinusOneSecond is minusOneSecond.
func (t TimePoint) MinusOneSecond() TimePoint { return TimePoint{t.dt.MinusSeconds(1)} }

// TimePointFormat is upstream's TimePointFormat: the ways the calendar
// writes a time point.
type TimePointFormat int

// The formats of the calendar.
const (
	FormatDayOfWeekShort TimePointFormat = iota
	FormatDayOfWeekLong
	FormatMonthYearShort
	FormatMonthYearLong
	FormatMonthLong
	FormatMonthShort
	FormatYear
	FormatQuarter
	FormatDayOfMonth
	FormatMonthAndDay
)

// Format is format.
func (f TimePointFormat) Format(when TimePoint, locale Locale) string {
	switch f {
	case FormatDayOfMonth:
		return strconv.Itoa(when.DayOfMonth())
	case FormatQuarter:
		return when.Quarter()
	case FormatYear:
		return strconv.Itoa(when.Year())
	case FormatDayOfWeekLong:
		return dayOfWeekLong(when.DayOfWeek(), locale)
	case FormatDayOfWeekShort:
		return dayOfWeekShort(when.DayOfWeek(), locale)
	case FormatMonthLong:
		return monthLong(when.Month(), locale)
	case FormatMonthShort:
		return monthShort(when.Month(), locale)
	case FormatMonthYearShort:
		return monthShort(when.Month(), locale) + " " + strconv.Itoa(when.Year())
	case FormatMonthYearLong:
		return monthLong(when.Month(), locale) + " " + strconv.Itoa(when.Year())
	case FormatMonthAndDay:
		return monthAndDay(when, locale)
	}
	uml.Throw(uml.IllegalStateException, "")
	return ""
}

func monthAndDay(when TimePoint, locale Locale) string {
	day := strconv.Itoa(when.DayOfMonth())
	switch locale {
	case "fr":
		return day + " " + FormatMonthShort.Format(when, locale)
	case "es":
		return day + " de " + FormatMonthShort.Format(when, locale)
	case "de":
		return day + ". " + FormatMonthShort.Format(when, locale)
	case "ja":
		return FormatMonthShort.Format(when, locale) + "/" + day + "日"
	case "ko":
		return FormatMonthShort.Format(when, locale) + " " + day + "일"
	case "ru":
		return day + " " + strings.ToLower(FormatMonthLong.Format(when, locale))
	case "zh":
		return FormatMonthShort.Format(when, locale) + day + "日"
	}
	return FormatMonthShort.Format(when, locale) + " " + day
}

// MonthRegexString is MonthUtils.getRegexString: JAN[a-z]*|FEB[a-z]*|...
func MonthRegexString() string {
	parts := make([]string, len(jtime.Months))
	for i, m := range jtime.Months {
		parts[i] = m.Name()[:3] + "[a-z]*"
	}
	return strings.Join(parts, "|")
}

// MonthUbrex is MonthUtils.getUbrex: the first three letters of a month in
// any case, then any letters.
func MonthUbrex() klimt.UBrexPart {
	alternatives := make([]klimt.UBrexPart, len(jtime.Months))
	for i, m := range jtime.Months {
		alternatives[i] = klimt.UBrexLeaf(prefixUbrex(m.Name()[:3]) + " 〇*〴le")
	}
	return klimt.UBrexOr(alternatives...)
}

// MonthFromString is MonthUtils.fromString: the month whose name starts
// with the first three letters of value, in any case.
func MonthFromString(value string) jtime.Month {
	value = javaSubstring(uml.UpperCase(value), 0, 3)
	for _, m := range jtime.Months {
		if strings.HasPrefix(m.Name(), value) {
			return m
		}
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}

// DayOfWeekRegexString is DayOfWeekUtils.getRegexString:
// MON[a-z]*|TUE[a-z]*|...
func DayOfWeekRegexString() string {
	parts := make([]string, len(jtime.DaysOfWeek))
	for i, d := range jtime.DaysOfWeek {
		parts[i] = d.Name()[:3] + "[a-z]*"
	}
	return strings.Join(parts, "|")
}

// DayOfWeekUbrex is DayOfWeekUtils.getUbrex: the first three letters of a
// day of the week in any case, then any letters.
func DayOfWeekUbrex() klimt.UBrexPart {
	alternatives := make([]klimt.UBrexPart, len(jtime.DaysOfWeek))
	for i, d := range jtime.DaysOfWeek {
		alternatives[i] = klimt.UBrexLeaf(prefixUbrex(d.Name()[:3]) + "〇*〴le")
	}
	return klimt.UBrexOr(alternatives...)
}

// DayOfWeekFromString is DayOfWeekUtils.fromString: the day of the week
// whose name starts with the first three letters of value, in any case.
func DayOfWeekFromString(value string) jtime.DayOfWeek {
	value = javaSubstring(uml.UpperCase(value), 0, 3)
	for _, d := range jtime.DaysOfWeek {
		if strings.HasPrefix(d.Name(), value) {
			return d
		}
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}

// prefixUbrex writes each letter of prefix as a UBrex set of its upper
// and lower case.
func prefixUbrex(prefix string) string {
	var b strings.Builder
	for _, c := range prefix {
		b.WriteString("「")
		b.WriteString(strings.ToUpper(string(c)))
		b.WriteString(strings.ToLower(string(c)))
		b.WriteString("」")
	}
	return b.String()
}

// javaSubstring is String.substring on UTF-16 code units, which fails with
// the StringIndexOutOfBoundsException of Java 21.
func javaSubstring(s string, begin, end int) string {
	u := utf16Encode(s)
	if begin < 0 || end > len(u) || begin > end {
		uml.Throw(uml.StringIndexOutOfBoundsException, "Range ["+strconv.Itoa(begin)+", "+strconv.Itoa(end)+
			") out of bounds for length "+strconv.Itoa(len(u)))
	}
	return utf16Decode(u[begin:end])
}

func floorDiv(x, y int64) int64 {
	q := x / y
	if (x%y != 0) && ((x < 0) != (y < 0)) {
		q--
	}
	return q
}
