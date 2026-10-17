// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/DateFunction.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Now.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"time"

	"github.com/ekalinin/gopuml/internal/text"
)

// currentTime is System.currentTimeMillis: Options.Now, or the clock.
func (c *TContext) currentTime() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// defaultZone is TimeZone.getDefault: Options.Location, or time.Local.
func (c *TContext) defaultZone() *javaZone {
	if c.location != nil {
		return zoneOfLocation(c.location)
	}
	return zoneOfLocation(time.Local)
}

// dateFunction is %date() (Date.toString of the current time) and
// %date(format[, seconds[, zone]]): the current time, or the instant given
// in seconds since the epoch, formatted with a SimpleDateFormat pattern in
// the default or the given time zone.
type dateFunction struct{ simpleReturnFunction }

func (dateFunction) Signature() TFunctionSignature { return NewTFunctionSignature("%date", 3) }

func (dateFunction) CanCover(nbArg int, _ []string) bool { return nbArg >= 0 && nbArg <= 3 }

func (dateFunction) ExecuteReturnFunction(context *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	if len(values) == 0 {
		return TValueFromString(javaDateToString(context.currentTime(), context.defaultZone())), nil
	}
	format := values[0].String()
	now := context.currentTime()
	if len(values) >= 2 {
		now = time.Unix(int64(values[1].ToInt()), 0)
	}
	zone := context.defaultZone()
	if len(values) == 3 {
		tzID := values[2].String()
		// SimpleDateFormat#setTimeZone silently falls back to GMT for unknown ids,
		// so validate explicitly and fail loudly to help users diagnose typos.
		z, ok := lookupJavaZone(tzID)
		if !ok {
			return nil, newEaterException("Unknown time zone: "+tzID, location)
		}
		zone = z
	}
	formatter, err := newJavaDateFormat(format)
	if err != nil {
		return nil, newEaterException("Bad date pattern", location)
	}
	return TValueFromString(formatter.format(now, zone)), nil
}

// now is %now(): the current time in seconds since the epoch, as an int
// (wrapping after 2038 as in Java).
type now struct{ simpleReturnFunction }

func (now) Signature() TFunctionSignature { return NewTFunctionSignature("%now", 0) }

func (now) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (now) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, _ []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromInt(int32(context.currentTime().UnixMilli() / 1000)), nil
}
