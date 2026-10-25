// SPDX-License-Identifier: MIT
// Provenance: original

package jtime

import (
	"math/big"
	"strconv"
	"strings"
)

// Duration is java.time.Duration: seconds and a nanosecond adjustment from
// 0 to 999,999,999, so that a negative duration has a negative number of
// seconds and a positive adjustment.
type Duration struct {
	seconds int64
	nanos   int32
}

// ZeroDuration is Duration.ZERO.
var ZeroDuration = Duration{}

// DurationOfSeconds is Duration.ofSeconds(seconds, nanoAdjustment).
func DurationOfSeconds(seconds, nanoAdjustment int64) Duration {
	return Duration{addExact(seconds, floorDiv(nanoAdjustment, nanosPerSecond)),
		int32(floorMod(nanoAdjustment, nanosPerSecond))}
}

// DurationOfDays is Duration.ofDays.
func DurationOfDays(days int64) Duration { return Duration{multiplyExact(days, secondsPerDay), 0} }

// DurationOfHours is Duration.ofHours.
func DurationOfHours(hours int64) Duration { return Duration{multiplyExact(hours, 3600), 0} }

// DurationOfMinutes is Duration.ofMinutes.
func DurationOfMinutes(minutes int64) Duration { return Duration{multiplyExact(minutes, 60), 0} }

// Between is Duration.between for two date-times: the exact time from start
// to end, negative when end is before start.
func Between(start, end LocalDateTime) Duration {
	seconds := (end.date.EpochDay()-start.date.EpochDay())*secondsPerDay +
		int64(end.time.SecondOfDay()-start.time.SecondOfDay())
	return DurationOfSeconds(seconds, int64(end.time.Nano()-start.time.Nano()))
}

// BetweenTimes is Duration.between for two times of day.
func BetweenTimes(start, end LocalTime) Duration {
	return DurationOfSeconds(0, end.nanoOfDay-start.nanoOfDay)
}

// Seconds is getSeconds: the whole seconds, rounded down.
func (d Duration) Seconds() int64 { return d.seconds }

// Nano is getNano: the nanosecond adjustment, 0 to 999,999,999.
func (d Duration) Nano() int { return int(d.nanos) }

// IsZero is isZero.
func (d Duration) IsZero() bool { return d.seconds == 0 && d.nanos == 0 }

// IsNegative is isNegative.
func (d Duration) IsNegative() bool { return d.seconds < 0 }

// Plus is plus(Duration).
func (d Duration) Plus(other Duration) Duration { return d.plus(other.seconds, int64(other.nanos)) }

// PlusDays is plusDays.
func (d Duration) PlusDays(days int64) Duration { return d.plus(multiplyExact(days, secondsPerDay), 0) }

// PlusHours is plusHours.
func (d Duration) PlusHours(hours int64) Duration { return d.plus(multiplyExact(hours, 3600), 0) }

// PlusMinutes is plusMinutes.
func (d Duration) PlusMinutes(minutes int64) Duration { return d.plus(multiplyExact(minutes, 60), 0) }

// PlusSeconds is plusSeconds.
func (d Duration) PlusSeconds(seconds int64) Duration { return d.plus(seconds, 0) }

func (d Duration) plus(seconds, nanos int64) Duration {
	if seconds == 0 && nanos == 0 {
		return d
	}
	s := addExact(addExact(d.seconds, seconds), nanos/nanosPerSecond)
	return DurationOfSeconds(s, int64(d.nanos)+nanos%nanosPerSecond)
}

// Negated is negated.
func (d Duration) Negated() Duration { return d.MultipliedBy(-1) }

// MultipliedBy is multipliedBy, which fails with Java's ArithmeticException
// when the result does not fit.
func (d Duration) MultipliedBy(multiplicand int64) Duration {
	switch multiplicand {
	case 0:
		return ZeroDuration
	case 1:
		return d
	}
	nanos := new(big.Int).Mul(big.NewInt(d.seconds), big.NewInt(nanosPerSecond))
	nanos.Add(nanos, big.NewInt(int64(d.nanos)))
	nanos.Mul(nanos, big.NewInt(multiplicand))
	q, r := new(big.Int).QuoRem(nanos, big.NewInt(nanosPerSecond), new(big.Int))
	if q.BitLen() > 63 {
		throw(ArithmeticException, "Exceeds capacity of Duration: "+nanos.String())
	}
	return DurationOfSeconds(q.Int64(), r.Int64())
}

// CompareTo is compareTo: negative, zero or positive.
func (d Duration) CompareTo(other Duration) int {
	switch {
	case d.seconds < other.seconds:
		return -1
	case d.seconds > other.seconds:
		return 1
	}
	return cmpInt(int(d.nanos), int(other.nanos))
}

// ToMillis is toMillis, which fails with Java's ArithmeticException when
// the milliseconds overflow a long.
func (d Duration) ToMillis() int64 {
	seconds, nanos := d.seconds, int64(d.nanos)
	if seconds < 0 {
		seconds++
		nanos -= nanosPerSecond
	}
	return addExact(multiplyExact(seconds, 1000), nanos/1_000_000)
}

// String is toString, in ISO-8601 format: PT25H1M1S, PT-10H, PT0S,
// PT-0.5S.
func (d Duration) String() string {
	if d.IsZero() {
		return "PT0S"
	}
	total := d.seconds
	if d.seconds < 0 && d.nanos > 0 {
		total++
	}
	hours := total / 3600
	minutes := total % 3600 / 60
	secs := total % 60
	var b strings.Builder
	b.WriteString("PT")
	if hours != 0 {
		b.WriteString(strconv.FormatInt(hours, 10))
		b.WriteString("H")
	}
	if minutes != 0 {
		b.WriteString(strconv.FormatInt(minutes, 10))
		b.WriteString("M")
	}
	if secs == 0 && d.nanos == 0 && b.Len() > 2 {
		return b.String()
	}
	if d.seconds < 0 && d.nanos > 0 && secs == 0 {
		b.WriteString("-0")
	} else {
		b.WriteString(strconv.FormatInt(secs, 10))
	}
	if d.nanos > 0 {
		var frac string
		if d.seconds < 0 {
			frac = strconv.FormatInt(2*nanosPerSecond-int64(d.nanos), 10)
		} else {
			frac = strconv.FormatInt(int64(d.nanos)+nanosPerSecond, 10)
		}
		b.WriteString(".")
		b.WriteString(strings.TrimRight(frac[1:], "0"))
	}
	b.WriteString("S")
	return b.String()
}
