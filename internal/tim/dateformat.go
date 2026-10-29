// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import "time"

// DateFormat is java.text.SimpleDateFormat in the en_US locale, for the
// packages outside the preprocessor (the date format of timing diagrams).
type DateFormat struct {
	f *javaDateFormat
}

// NewDateFormat compiles a pattern as new SimpleDateFormat(pattern,
// Locale.US) does, and returns its error for an unknown pattern letter or
// an unterminated quote.
func NewDateFormat(pattern string) (*DateFormat, error) {
	f, err := newJavaDateFormat(pattern)
	if err != nil {
		return nil, err
	}
	return &DateFormat{f: f}, nil
}

// Format is SimpleDateFormat.format of the instant t in the time zone of
// loc, with the Java rules and names of that zone; time.Local is the zone
// of $TZ or /etc/localtime, as the JVM's default zone.
func (d *DateFormat) Format(t time.Time, loc *time.Location) string {
	return d.f.format(t, zoneOfLocation(loc))
}
