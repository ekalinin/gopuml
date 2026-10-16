// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // the zones of %date must not depend on the system
	"unicode"
)

// javaZone is a java.util.TimeZone of the reference JVM: the offset rules of
// a zone and its display names in the en_US locale.
type javaZone struct {
	// loc holds the offset rules; nil for a SystemV zone with daylight
	// saving time, whose rules are computed by systemVOffset.
	loc *time.Location
	// sysvStd is the standard offset, in seconds, of a SystemV zone.
	sysvStd int
	// names are the short and long standard names, then the short and
	// long daylight names; nil when the zone is not a Java zone.
	names *[4]string
}

// javaZoneEntry is a line of javaTimeZoneData.
type javaZoneEntry struct {
	target string
	names  [4]string
}

var javaZones = sync.OnceValue(func() map[string]*javaZoneEntry {
	m := map[string]*javaZoneEntry{}
	for line := range strings.SplitSeq(strings.TrimSpace(javaTimeZoneData), "\n") {
		f := strings.Split(line, "|")
		m[f[0]] = &javaZoneEntry{target: f[1], names: [4]string{f[2], f[3], f[4], f[5]}}
	}
	return m
})

// lookupJavaZone is TimeZone.getTimeZone for the ids of
// TimeZone.getAvailableIDs; ok is false for any other id.
func lookupJavaZone(id string) (*javaZone, bool) {
	e := javaZones()[id]
	if e == nil {
		return nil, false
	}
	z := &javaZone{names: &e.names}
	target := e.target
	if target == "" {
		target = id
	}
	switch {
	case strings.HasPrefix(target, "+") || strings.HasPrefix(target, "-"):
		// a fixed offset, such as EST
		h, _ := strconv.Atoi(target[1:3])
		m, _ := strconv.Atoi(target[4:6])
		offset := h*3600 + m*60
		if target[0] == '-' {
			offset = -offset
		}
		z.loc = time.FixedZone(id, offset)
	case strings.HasPrefix(target, "SystemV/"):
		// SystemV/EST5 is fixed, SystemV/EST5EDT follows the SystemV rules
		name := target[len("SystemV/"):]
		i := strings.IndexFunc(name, unicode.IsDigit)
		j := i + strings.IndexFunc(name[i:], func(r rune) bool { return !unicode.IsDigit(r) })
		if j < i {
			j = len(name)
		}
		hours, _ := strconv.Atoi(name[i:j])
		z.sysvStd = -hours * 3600
		if j == len(name) {
			z.loc = time.FixedZone(id, z.sysvStd)
		}
	default:
		loc, err := time.LoadLocation(target)
		if err != nil {
			return nil, false
		}
		z.loc = loc
	}
	return z, true
}

// zoneOfLocation returns the Java zone of a Go location: the Java zone with
// the same name, or for time.Local the zone named by $TZ or /etc/localtime,
// or else a zone without Java names.
func zoneOfLocation(loc *time.Location) *javaZone {
	name := loc.String()
	if loc == time.Local {
		name = localZoneName()
	}
	if z, ok := lookupJavaZone(name); ok {
		z.loc = loc
		return z
	}
	return &javaZone{loc: loc}
}

// localZoneName is the IANA name of the local time zone, as Java finds it:
// $TZ, else the target of the /etc/localtime link.
var localZoneName = sync.OnceValue(func() string {
	if tz, ok := os.LookupEnv("TZ"); ok {
		return strings.TrimPrefix(tz, ":")
	}
	if target, err := os.Readlink("/etc/localtime"); err == nil {
		if _, after, ok := strings.Cut(target, "zoneinfo/"); ok {
			return after
		}
	}
	return "Local"
})

// at returns the offset from UTC in seconds and whether daylight saving time
// is in effect at the instant t, as Java's Calendar.DST_OFFSET tells it.
// Java has no negative saving: where the Go zone data has one (for example
// Europe/Dublin in winter), Java has standard time, and daylight saving time
// on the other side of the transition.
func (z *javaZone) at(t time.Time) (offset int, dst bool) {
	if z.loc == nil {
		return systemVOffset(t, z.sysvStd)
	}
	lt := t.In(z.loc)
	_, offset = lt.Zone()
	dst = lt.IsDST()
	start, end := lt.ZoneBounds()
	var neighbor time.Time
	switch {
	case !end.IsZero():
		neighbor = end
	case !start.IsZero():
		neighbor = start.Add(-time.Nanosecond)
	default:
		return offset, dst
	}
	nt := neighbor.In(z.loc)
	_, neighborOffset := nt.Zone()
	switch {
	case dst && !nt.IsDST() && offset < neighborOffset:
		dst = false
	case !dst && nt.IsDST() && neighborOffset < offset:
		dst = true
	}
	return offset, dst
}

// systemVOffset applies the rules of the SystemV zones with daylight saving
// time: from 2:00 on the last Sunday of April to 2:00 on the last Sunday of
// October, except in 1974 (from January 6 to the last Sunday of November)
// and 1975 (from February 23).
func systemVOffset(t time.Time, std int) (int, bool) {
	year := t.UTC().Add(time.Duration(std) * time.Second).Year()
	start := lastSunday(year, time.April)
	end := lastSunday(year, time.October)
	switch year {
	case 1974:
		start, end = time.Date(1974, time.January, 6, 0, 0, 0, 0, time.UTC), lastSunday(1974, time.November)
	case 1975:
		start = time.Date(1975, time.February, 23, 0, 0, 0, 0, time.UTC)
	}
	// the changes happen at 2:00 local time, standard time in spring and
	// daylight saving time in autumn
	startUTC := start.Add(2*time.Hour - time.Duration(std)*time.Second)
	endUTC := end.Add(time.Hour - time.Duration(std)*time.Second)
	if !t.Before(startUTC) && t.Before(endUTC) {
		return std + 3600, true
	}
	return std, false
}

func lastSunday(year int, month time.Month) time.Time {
	d := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return d.AddDate(0, 0, -int(d.Weekday()))
}

// displayName is TimeZone.getDisplayName(daylight, style, Locale.US). A
// zone without Java names uses the Go abbreviation when it is made of
// letters, else Java's custom form "GMT+hh:mm".
func (z *javaZone) displayName(t time.Time, dst, long bool) string {
	if z.names != nil {
		i := 0
		if dst {
			i = 2
		}
		if long {
			i++
		}
		return z.names[i]
	}
	name, offset := t.In(z.loc).Zone()
	if name != "" && strings.IndexFunc(name, func(r rune) bool { return !unicode.IsLetter(r) }) < 0 {
		return name
	}
	sign := byte('+')
	minutes := offset / 60
	if minutes < 0 {
		sign = '-'
		minutes = -minutes
	}
	var b strings.Builder
	b.WriteString("GMT")
	b.WriteByte(sign)
	sprintf0d(&b, minutes/60, 2)
	b.WriteByte(':')
	sprintf0d(&b, minutes%60, 2)
	return b.String()
}
