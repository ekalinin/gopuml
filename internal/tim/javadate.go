// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// javaDateFormat is java.text.SimpleDateFormat in the en_US locale with the
// Gregorian calendar of that locale (weeks start on Sunday, the first week
// of a year or month needs one day). It supports every pattern letter of
// Java: G y Y M L w W D d F E u a H k K h m s S z Z X.
type javaDateFormat struct {
	fields []dateField
}

// dateField is a compiled pattern element: a letter repeated count times,
// or literal text when letter is 0.
type dateField struct {
	letter  byte
	count   int
	literal string
}

const javaPatternChars = "GyMdkHmsSEDFwWahKzZYuXL"

var (
	errUnterminatedQuote = errors.New("Unterminated quote")
	errIllegalPattern    = errors.New("Illegal pattern character")
	errISOFormat         = errors.New("invalid ISO 8601 format")
)

// newJavaDateFormat compiles a pattern like SimpleDateFormat.compile: ASCII
// letters are pattern letters, text between single quotes is literal and
// two single quotes are one quote. It fails like Java on an unknown letter,
// an unterminated quote and an X repeated more than three times.
func newJavaDateFormat(pattern string) (*javaDateFormat, error) {
	f := &javaDateFormat{}
	u := toUTF16(pattern)
	var literal []uint16
	inQuote := false
	var last byte
	count := 0
	flush := func() error {
		if count != 0 {
			if last == 'X' && count >= 4 {
				return errISOFormat
			}
			f.fields = append(f.fields, dateField{letter: last, count: count})
			count = 0
		}
		return nil
	}
	flushLiteral := func() {
		if len(literal) > 0 {
			f.fields = append(f.fields, dateField{literal: fromUTF16(literal)})
			literal = nil
		}
	}
	for i := 0; i < len(u); i++ {
		c := u[i]
		if c == '\'' {
			if i+1 < len(u) && u[i+1] == '\'' {
				i++
				if err := flush(); err != nil {
					return nil, err
				}
				literal = append(literal, '\'')
				continue
			}
			if err := flush(); err != nil {
				return nil, err
			}
			inQuote = !inQuote
			continue
		}
		if inQuote || !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			if err := flush(); err != nil {
				return nil, err
			}
			literal = append(literal, c)
			continue
		}
		if strings.IndexByte(javaPatternChars, byte(c)) < 0 {
			return nil, errIllegalPattern
		}
		if count != 0 && last != byte(c) {
			if err := flush(); err != nil {
				return nil, err
			}
		}
		flushLiteral()
		last = byte(c)
		count++
	}
	if inQuote {
		return nil, errUnterminatedQuote
	}
	if err := flush(); err != nil {
		return nil, err
	}
	flushLiteral()
	return f, nil
}

// Date symbols of the en_US locale.
var (
	javaMonths        = [...]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	javaShortMonths   = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	javaWeekdays      = [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	javaShortWeekdays = [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
)

// format is SimpleDateFormat.format of the instant t in the time zone.
func (f *javaDateFormat) format(t time.Time, zone *javaZone) string {
	offset, dst := zone.at(t)
	c := newJavaCalendar(t, offset)
	var b strings.Builder
	for _, field := range f.fields {
		if field.letter == 0 {
			b.WriteString(field.literal)
			continue
		}
		count := field.count
		switch field.letter {
		case 'G':
			b.WriteString("AD")
		case 'y', 'Y':
			year := c.year
			if field.letter == 'Y' {
				year = c.weekYear
			}
			if count == 2 {
				zeroPaddingNumber(&b, year%100, 2)
			} else {
				zeroPaddingNumber(&b, year, count)
			}
		case 'M', 'L':
			switch {
			case count >= 4:
				b.WriteString(javaMonths[c.month])
			case count == 3:
				b.WriteString(javaShortMonths[c.month])
			default:
				zeroPaddingNumber(&b, c.month+1, count)
			}
		case 'E':
			if count >= 4 {
				b.WriteString(javaWeekdays[c.weekday])
			} else {
				b.WriteString(javaShortWeekdays[c.weekday])
			}
		case 'a':
			if c.hour < 12 {
				b.WriteString("AM")
			} else {
				b.WriteString("PM")
			}
		case 'k':
			zeroPaddingNumber(&b, cmpOr(c.hour, 24), count)
		case 'h':
			zeroPaddingNumber(&b, cmpOr(c.hour%12, 12), count)
		case 'H':
			zeroPaddingNumber(&b, c.hour, count)
		case 'K':
			zeroPaddingNumber(&b, c.hour%12, count)
		case 'd':
			zeroPaddingNumber(&b, c.day, count)
		case 'm':
			zeroPaddingNumber(&b, c.minute, count)
		case 's':
			zeroPaddingNumber(&b, c.second, count)
		case 'S':
			zeroPaddingNumber(&b, c.millis, count)
		case 'D':
			zeroPaddingNumber(&b, c.dayOfYear, count)
		case 'F':
			zeroPaddingNumber(&b, (c.day-1)/7+1, count)
		case 'w':
			zeroPaddingNumber(&b, c.weekOfYear, count)
		case 'W':
			zeroPaddingNumber(&b, c.weekOfMonth, count)
		case 'u':
			zeroPaddingNumber(&b, cmpOr(c.weekday, 7), count)
		case 'z':
			b.WriteString(zone.displayName(t, dst, count >= 4))
		case 'Z':
			value := offset / 60
			width := 4
			if value >= 0 {
				b.WriteByte('+')
			} else {
				width++
			}
			sprintf0d(&b, (value/60)*100+(value%60), width)
		case 'X':
			if offset == 0 {
				b.WriteByte('Z')
				break
			}
			value := offset / 60
			if value >= 0 {
				b.WriteByte('+')
			} else {
				b.WriteByte('-')
				value = -value
			}
			sprintf0d(&b, value/60, 2)
			if count == 1 {
				break
			}
			if count == 3 {
				b.WriteByte(':')
			}
			sprintf0d(&b, value%60, 2)
		}
	}
	return b.String()
}

// cmpOr returns v, or zero when v is 0 (the 1-based hours and the ISO day
// of the week).
func cmpOr(v, zero int) int {
	if v == 0 {
		return zero
	}
	return v
}

// zeroPaddingNumber writes a non-negative value with at least minDigits
// digits, as SimpleDateFormat does with its NumberFormat (which caps the
// minimum at 309 digits).
func zeroPaddingNumber(b *strings.Builder, value, minDigits int) {
	s := strconv.Itoa(value)
	for i := len(s); i < min(minDigits, 309); i++ {
		b.WriteByte('0')
	}
	b.WriteString(s)
}

// sprintf0d is CalendarUtils.sprintf0d: value with at least width
// characters, padded with zeros after the sign.
func sprintf0d(b *strings.Builder, value, width int) {
	d := value
	if d < 0 {
		b.WriteByte('-')
		d = -d
		width--
	}
	n := 10
	for i := 2; i < width; i++ {
		n *= 10
	}
	for i := 1; i < width && d < n; i++ {
		b.WriteByte('0')
		n /= 10
	}
	b.WriteString(strconv.Itoa(d))
}

// javaCalendar holds the fields of a GregorianCalendar in the en_US locale.
type javaCalendar struct {
	year, month, day                   int // month from 0
	hour, minute, second, millis       int
	weekday                            int // 0 for Sunday
	dayOfYear, weekOfYear, weekOfMonth int
	weekYear                           int
}

// newJavaCalendar returns the calendar fields of the instant t at the given
// offset from UTC, in seconds.
func newJavaCalendar(t time.Time, offset int) javaCalendar {
	local := t.UTC().Add(time.Duration(offset) * time.Second)
	c := javaCalendar{
		year:      local.Year(),
		month:     int(local.Month()) - 1,
		day:       local.Day(),
		hour:      local.Hour(),
		minute:    local.Minute(),
		second:    local.Second(),
		millis:    local.Nanosecond() / 1e6,
		weekday:   int(local.Weekday()),
		dayOfYear: local.YearDay(),
	}
	fixedDate := daysSinceEpoch(c.year, 1, 1) + int64(c.dayOfYear) - 1
	jan1 := daysSinceEpoch(c.year, 1, 1)
	c.weekOfYear = javaWeekNumber(jan1, fixedDate)
	if c.weekOfYear >= 52 {
		nextJan1 := daysSinceEpoch(c.year+1, 1, 1)
		nextJan1st := dayOfWeekOnOrBefore(nextJan1+6, 0)
		if nextJan1st-nextJan1 >= 1 && fixedDate >= nextJan1st-7 {
			// The first days form a week in which the date is included.
			c.weekOfYear = 1
		}
	}
	c.weekOfMonth = javaWeekNumber(daysSinceEpoch(c.year, c.month+1, 1), fixedDate)
	c.weekYear = c.year
	if c.month == 0 && c.weekOfYear >= 52 {
		c.weekYear--
	} else if c.month != 0 && c.weekOfYear == 1 {
		c.weekYear++
	}
	return c
}

// daysSinceEpoch is the number of days from 1970-01-01 to the date.
func daysSinceEpoch(year, month, day int) int64 {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Unix() / 86400
}

// dayOfWeekOnOrBefore is the last day on or before fixed that falls on the
// weekday dow (0 for Sunday).
func dayOfWeekOnOrBefore(fixed int64, dow int) int64 {
	weekday := int((fixed%7 + 7 + 4) % 7) // 1970-01-01 is a Thursday
	return fixed - int64((weekday-dow+7)%7)
}

// javaWeekNumber is GregorianCalendar.getWeekNumber with Sunday as the first
// day of the week and one minimal day in the first week.
func javaWeekNumber(fixedDay1, fixedDate int64) int {
	fixedDay1st := dayOfWeekOnOrBefore(fixedDay1+6, 0)
	if fixedDay1st-fixedDay1 >= 1 {
		fixedDay1st -= 7
	}
	n := fixedDate - fixedDay1st
	if n >= 0 {
		return int(n/7) + 1
	}
	return int((n-6)/7) + 1
}

// javaDateToString is Date.toString: "EEE MMM dd HH:mm:ss zzz yyyy".
func javaDateToString(t time.Time, zone *javaZone) string {
	f, _ := newJavaDateFormat("EEE MMM dd HH:mm:ss zzz yyyy")
	return f.format(t, zone)
}
