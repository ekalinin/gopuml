// SPDX-License-Identifier: MIT
// Provenance: original

package jtime

import "strconv"

const (
	nanosPerSecond = 1_000_000_000
	secondsPerDay  = 86_400
	nanosPerDay    = secondsPerDay * nanosPerSecond
)

// LocalTime is java.time.LocalTime: a time of day, to the nanosecond.
type LocalTime struct {
	nanoOfDay int64
}

// The first and the last time of a day.
var (
	Midnight = LocalTime{0}
	MaxTime  = LocalTime{nanosPerDay - 1}
)

// TimeOf is LocalTime.of(hour, minute, second), with Java's validation and
// messages.
func TimeOf(hour, minute, second int) LocalTime {
	checkValue("HourOfDay", 0, 23, int64(hour))
	checkValue("MinuteOfHour", 0, 59, int64(minute))
	checkValue("SecondOfMinute", 0, 59, int64(second))
	return LocalTime{(int64(hour)*3600 + int64(minute)*60 + int64(second)) * nanosPerSecond}
}

// TimeOfNanoOfDay is LocalTime.ofNanoOfDay.
func TimeOfNanoOfDay(nanoOfDay int64) LocalTime {
	checkValue("NanoOfDay", 0, nanosPerDay-1, nanoOfDay)
	return LocalTime{nanoOfDay}
}

// Hour is getHour.
func (t LocalTime) Hour() int { return int(t.nanoOfDay / (3600 * nanosPerSecond)) }

// Minute is getMinute.
func (t LocalTime) Minute() int { return int(t.nanoOfDay / (60 * nanosPerSecond) % 60) }

// Second is getSecond.
func (t LocalTime) Second() int { return int(t.nanoOfDay / nanosPerSecond % 60) }

// Nano is getNano.
func (t LocalTime) Nano() int { return int(t.nanoOfDay % nanosPerSecond) }

// SecondOfDay is toSecondOfDay.
func (t LocalTime) SecondOfDay() int { return int(t.nanoOfDay / nanosPerSecond) }

// NanoOfDay is toNanoOfDay.
func (t LocalTime) NanoOfDay() int64 { return t.nanoOfDay }

// CompareTo is compareTo: negative, zero or positive.
func (t LocalTime) CompareTo(other LocalTime) int {
	switch {
	case t.nanoOfDay < other.nanoOfDay:
		return -1
	case t.nanoOfDay > other.nanoOfDay:
		return 1
	}
	return 0
}

// IsBefore is isBefore.
func (t LocalTime) IsBefore(other LocalTime) bool { return t.nanoOfDay < other.nanoOfDay }

// IsAfter is isAfter.
func (t LocalTime) IsAfter(other LocalTime) bool { return t.nanoOfDay > other.nanoOfDay }

// String is toString: HH:mm, with :ss when the seconds or the nanoseconds
// are not zero, and the fraction of the second in groups of three digits.
func (t LocalTime) String() string {
	s := pad(t.Hour(), 2) + ":" + pad(t.Minute(), 2)
	sec, nano := t.Second(), t.Nano()
	if sec > 0 || nano > 0 {
		s += ":" + pad(sec, 2)
		switch {
		case nano == 0:
		case nano%1_000_000 == 0:
			s += "." + strconv.Itoa(nano/1_000_000 + 1000)[1:]
		case nano%1000 == 0:
			s += "." + strconv.Itoa(nano/1000 + 1_000_000)[1:]
		default:
			s += "." + strconv.Itoa(nano + nanosPerSecond)[1:]
		}
	}
	return s
}

// LocalDateTime is java.time.LocalDateTime: a date and a time of day
// without a time zone.
type LocalDateTime struct {
	date LocalDate
	time LocalTime
}

// The extreme date-times.
var (
	MinDateTime = LocalDateTime{MinDate, Midnight}
	MaxDateTime = LocalDateTime{MaxDate, MaxTime}
)

// DateTimeOf is LocalDateTime.of(date, time).
func DateTimeOf(date LocalDate, time LocalTime) LocalDateTime { return LocalDateTime{date, time} }

// DateTimeOfFields is LocalDateTime.of(year, month, dayOfMonth, hour,
// minute, second), which validates the date, then the time.
func DateTimeOfFields(year, month, dayOfMonth, hour, minute, second int) LocalDateTime {
	date := DateOf(year, month, dayOfMonth)
	return LocalDateTime{date, TimeOf(hour, minute, second)}
}

// LocalDate is toLocalDate.
func (dt LocalDateTime) LocalDate() LocalDate { return dt.date }

// LocalTime is toLocalTime.
func (dt LocalDateTime) LocalTime() LocalTime { return dt.time }

// Year is getYear.
func (dt LocalDateTime) Year() int { return dt.date.year }

// Month is getMonth.
func (dt LocalDateTime) Month() Month { return dt.date.month }

// DayOfMonth is getDayOfMonth.
func (dt LocalDateTime) DayOfMonth() int { return dt.date.day }

// DayOfWeek is getDayOfWeek.
func (dt LocalDateTime) DayOfWeek() DayOfWeek { return dt.date.DayOfWeek() }

// Hour is getHour.
func (dt LocalDateTime) Hour() int { return dt.time.Hour() }

// Minute is getMinute.
func (dt LocalDateTime) Minute() int { return dt.time.Minute() }

// Second is getSecond.
func (dt LocalDateTime) Second() int { return dt.time.Second() }

// WithTime is with(LocalTime): the same date at another time.
func (dt LocalDateTime) WithTime(t LocalTime) LocalDateTime { return LocalDateTime{dt.date, t} }

// PlusDays is plusDays.
func (dt LocalDateTime) PlusDays(days int64) LocalDateTime {
	return LocalDateTime{dt.date.PlusDays(days), dt.time}
}

// MinusDays is minusDays.
func (dt LocalDateTime) MinusDays(days int64) LocalDateTime {
	return LocalDateTime{dt.date.MinusDays(days), dt.time}
}

// PlusSeconds is plusSeconds.
func (dt LocalDateTime) PlusSeconds(seconds int64) LocalDateTime { return dt.plus(seconds, 0) }

// MinusSeconds is minusSeconds.
func (dt LocalDateTime) MinusSeconds(seconds int64) LocalDateTime {
	if seconds == minLong {
		return dt.plus(maxLong, 0).plus(1, 0)
	}
	return dt.plus(-seconds, 0)
}

// Plus is plus(Duration).
func (dt LocalDateTime) Plus(d Duration) LocalDateTime { return dt.plus(d.seconds, int64(d.nanos)) }

// Minus is minus(Duration).
func (dt LocalDateTime) Minus(d Duration) LocalDateTime {
	r := dt
	if d.seconds != 0 {
		r = r.MinusSeconds(d.seconds)
	}
	if d.nanos != 0 {
		r = r.plus(0, -int64(d.nanos))
	}
	return r
}

// plus adds seconds and nanoseconds, the days to the date and the rest to
// the time of day, as plusWithOverflow does.
func (dt LocalDateTime) plus(seconds, nanos int64) LocalDateTime {
	if seconds == 0 && nanos == 0 {
		return dt
	}
	days := seconds/secondsPerDay + nanos/nanosPerDay
	rest := seconds%secondsPerDay*nanosPerSecond + nanos%nanosPerDay + dt.time.nanoOfDay
	days += floorDiv(rest, nanosPerDay)
	return LocalDateTime{dt.date.PlusDays(days), LocalTime{floorMod(rest, nanosPerDay)}}
}

// CompareTo is compareTo: negative, zero or positive.
func (dt LocalDateTime) CompareTo(other LocalDateTime) int {
	if c := dt.date.CompareTo(other.date); c != 0 {
		return c
	}
	return dt.time.CompareTo(other.time)
}

// IsBefore is isBefore.
func (dt LocalDateTime) IsBefore(other LocalDateTime) bool { return dt.CompareTo(other) < 0 }

// IsAfter is isAfter.
func (dt LocalDateTime) IsAfter(other LocalDateTime) bool { return dt.CompareTo(other) > 0 }

// EpochSecond is toEpochSecond(ZoneOffset.UTC).
func (dt LocalDateTime) EpochSecond() int64 {
	return dt.date.EpochDay()*secondsPerDay + int64(dt.time.SecondOfDay())
}

// EpochMilli is toInstant(ZoneOffset.UTC).toEpochMilli(), which fails with
// Java's ArithmeticException when the milliseconds overflow a long.
func (dt LocalDateTime) EpochMilli() int64 {
	seconds, nanos := dt.EpochSecond(), int64(dt.time.Nano())
	if seconds < 0 && nanos > 0 {
		return addExact(multiplyExact(seconds+1, 1000), nanos/1_000_000-1000)
	}
	return addExact(multiplyExact(seconds, 1000), nanos/1_000_000)
}

// DateTimeOfEpochSecond is LocalDateTime.ofEpochSecond(seconds, nanos,
// ZoneOffset.UTC).
func DateTimeOfEpochSecond(seconds int64, nanos int) LocalDateTime {
	return LocalDateTime{DateOfEpochDay(floorDiv(seconds, secondsPerDay)),
		LocalTime{floorMod(seconds, secondsPerDay)*nanosPerSecond + int64(nanos)}}
}

// String is toString, in ISO format: 2024-09-22T10:30:46.
func (dt LocalDateTime) String() string { return dt.date.String() + "T" + dt.time.String() }
