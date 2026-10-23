// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"testing"
	"time"
)

func TestUndecoratedWait(t *testing.T) {
	at := func(min, sec int) time.Time { return time.Date(2026, 10, 23, 12, min, sec, 0, time.UTC) }
	// India is 5:30 ahead of UTC: 12:45 there is 07:15 UTC, a decorated
	// minute, while minute 45 is not one. Upstream counts minutes from the
	// Unix time, as UTC.
	india := time.FixedZone("IST", 5*3600+30*60)
	window := 20 * time.Second
	for _, tc := range []struct {
		name string
		now  time.Time
		want time.Duration
	}{
		{"plain minute", at(16, 0), 0},
		{"window ends before a decorated minute", at(14, 30), 0},
		{"window reaches minute 15", at(14, 50), 70 * time.Second},
		{"in minute 1", at(1, 10), 50 * time.Second},
		{"in minute 13", at(13, 0), time.Minute},
		{"window reaches minute 30", at(29, 45), 75 * time.Second},
		{"across the hour", at(59, 50), 0},
		{"window reaches minute 1 of the next hour", at(0, 50), 70 * time.Second},
		{"Unix minutes, not local ones", time.Date(2026, 10, 23, 12, 45, 0, 0, india), time.Minute},
	} {
		if got := undecoratedWait(tc.now, window); got != tc.want {
			t.Errorf("%s: undecoratedWait(%s) = %s, want %s", tc.name, tc.now.Format(time.TimeOnly+" MST"), got, tc.want)
		}
	}
}
