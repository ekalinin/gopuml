// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeTick.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingFormat.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/DeduceFormat.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeAxisStategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"fmt"
	"strconv"
	"time"
	"unicode"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/tim"
)

// TimeTick is a time of the diagram with the format of its labels. Ticks
// are compared by value (10 and 10.0 are the same time) and have an
// identity, as upstream's TimeTick, which does not override equals.
type TimeTick struct {
	time   jmath.BigDecimal
	format *TimingFormat
}

func newTimeTick(t jmath.BigDecimal, format *TimingFormat) *TimeTick {
	return &TimeTick{time: t, format: format}
}

// Time is the time of the tick.
func (t *TimeTick) Time() jmath.BigDecimal { return t.time }

// CompareTo compares the times of two ticks, ignoring their scales.
func (t *TimeTick) CompareTo(other *TimeTick) int { return t.time.CompareTo(other.time) }

func (t *TimeTick) String() string { return t.format.formatTime(t.time) }

// TimingFormat formats the times of the axis: DECIMAL, HOUR, DATE, or a
// SimpleDateFormat given by "use date format". The three constants are
// compared by identity, as upstream does.
type TimingFormat struct {
	name string
	sdf  *tim.DateFormat
}

var (
	formatDecimal = &TimingFormat{name: "DECIMAL"}
	formatHour    = &TimingFormat{name: "HOUR"}
	formatDate    = &TimingFormat{name: "DATE"}
)

// newTimingFormat is TimingFormat.create(sdf).
func newTimingFormat(sdf *tim.DateFormat) *TimingFormat { return &TimingFormat{name: "SDF", sdf: sdf} }

// formatTime formats a time: HOUR, DATE and date formats need an integral
// time (longValueExact throws ArithmeticException otherwise), DECIMAL
// keeps the scale of the time (toPlainString).
func (f *TimingFormat) formatTime(t jmath.BigDecimal) string {
	if f == formatHour || f == formatDate || f.sdf != nil {
		return f.formatLong(t.LongValueExact())
	}
	return t.ToPlainString()
}

// formatLong formats a time in seconds: with the date format in the
// default time zone (the JVM's), as h:mm:ss with Java's int arithmetic, as
// MM/dd in GMT, or as the number.
func (f *TimingFormat) formatLong(t int64) string {
	if f.sdf != nil {
		return f.sdf.Format(time.UnixMilli(t*1000), time.Local)
	}
	if f == formatHour {
		s := int32(t) % 60
		m := int32(t/60) % 60
		h := int32(t / 3600)
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	if f == formatDate {
		date := jtime.DateOfEpochDay(floorDiv(t, 86400))
		return fmt.Sprintf("%02d/%02d", date.MonthValue(), date.DayOfMonth())
	}
	return strconv.FormatInt(t, 10)
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// createDate is TimingFormat.createDate: the epoch second of y/m/d at 00:00
// GMT, as a lenient GregorianCalendar computes it (a month above 12 and a
// day past the end of the month roll over), in the proleptic Gregorian
// calendar.
func createDate(yyyy, mm, dd int, format *TimingFormat) *TimeTick {
	month := int64(mm - 1)
	year := int64(yyyy) + floorDiv(month, 12)
	month -= floorDiv(month, 12) * 12
	epochDay := jtime.DateOf(int(year), int(month)+1, 1).EpochDay() + int64(dd) - 1
	return newTimeTick(jmath.BigDecimalOfInt64(epochDay*86400), format)
}

// TimeAxisStategy is how the time axis is drawn.
type TimeAxisStategy int

const (
	TimeAxisAutomatic TimeAxisStategy = iota
	TimeAxisHidden
	TimeAxisManual
)

// deduceFormat is upstream's DeduceFormat: the number format that the
// bounds of an analog player are written in.
type deduceFormat struct {
	forceSign         bool
	groupingSep       rune
	decimalSep        rune
	minFractionDigits int
	maxFractionDigits int
}

// mergeWith merges the formats of the two bounds; a nil other keeps f.
func (f *deduceFormat) mergeWith(other *deduceFormat) *deduceFormat {
	if other == nil {
		return f
	}
	mergedDecimal := f.decimalSep
	if mergedDecimal == 0 {
		mergedDecimal = other.decimalSep
	}
	var mergedGrouping rune
	switch {
	case f.groupingSep == 0:
		mergedGrouping = other.groupingSep
	case other.groupingSep == 0:
		mergedGrouping = f.groupingSep
	case f.groupingSep == other.groupingSep:
		mergedGrouping = f.groupingSep
	}
	if mergedGrouping != 0 && mergedDecimal != 0 && mergedGrouping == mergedDecimal {
		mergedGrouping = 0
	}
	return &deduceFormat{forceSign: f.forceSign || other.forceSign, groupingSep: mergedGrouping,
		decimalSep: mergedDecimal, minFractionDigits: min(f.minFractionDigits, other.minFractionDigits),
		maxFractionDigits: max(f.maxFractionDigits, other.maxFractionDigits)}
}

// decimalFormat is DeduceFormat.getDecimalFormat: the pattern "0" or
// "#,##0", with a "+" prefix when forced, the separators and the fraction
// digits.
func (f *deduceFormat) decimalFormat() *jfmt.DecimalFormat {
	df := jfmt.NewDecimalFormat()
	if f.decimalSep != 0 {
		df.DecimalSeparator = f.decimalSep
	}
	if f.groupingSep != 0 {
		df.GroupingSeparator = f.groupingSep
	}
	if f.forceSign {
		df.PositivePrefix = "+"
	}
	df.GroupingUsed = f.groupingSep != 0
	df.MinimumFractionDigits = f.minFractionDigits
	df.MaximumFractionDigits = max(f.maxFractionDigits, f.minFractionDigits)
	return df
}

// deduceFormatFrom is DeduceFormat.from: the sign, the separators and the
// fraction digits of a number as written; nil when it has more than two
// kinds of separators.
func deduceFormatFrom(number string) *deduceFormat {
	var nonDigit1, nonDigit2 struct {
		ch    rune
		count int
	}
	var sign rune
	if number != "" && (number[0] == '+' || number[0] == '-') {
		sign = rune(number[0])
		number = number[1:]
	}
	fractionDigits := 0
	for _, ch := range number {
		if unicode.IsDigit(ch) {
			if nonDigit1.count > 0 {
				fractionDigits++
			}
			continue
		}
		switch {
		case nonDigit1.count == 0:
			nonDigit1.ch, nonDigit1.count = ch, 1
			fractionDigits = 0
		case nonDigit1.ch == ch:
			nonDigit1.count++
		case nonDigit2.count == 0:
			nonDigit2.ch, nonDigit2.count = ch, 1
			fractionDigits = 0
		case nonDigit2.ch == ch:
			nonDigit2.count++
		default:
			return nil
		}
	}
	if nonDigit1.count > 0 && nonDigit2.count > 0 {
		return &deduceFormat{forceSign: sign == '+', groupingSep: nonDigit1.ch, decimalSep: nonDigit2.ch,
			minFractionDigits: fractionDigits, maxFractionDigits: fractionDigits}
	}
	return &deduceFormat{forceSign: sign == '+', decimalSep: nonDigit1.ch, minFractionDigits: fractionDigits,
		maxFractionDigits: fractionDigits}
}
