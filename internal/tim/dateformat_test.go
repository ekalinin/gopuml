// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"testing"
	"time"
)

func TestDateFormat(t *testing.T) {
	instant := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
	f, err := NewDateFormat("MMM yyyy")
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Format(instant, time.UTC); got != "Jul 2024" {
		t.Errorf("MMM yyyy in UTC = %q", got)
	}
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip(err)
	}
	day, err := NewDateFormat("dd")
	if err != nil {
		t.Fatal(err)
	}
	if got := day.Format(instant, ny); got != "14" {
		t.Errorf("dd in America/New_York = %q, want the previous day", got)
	}
	if _, err := NewDateFormat("yyyy-qq"); err == nil || err.Error() != "Illegal pattern character" {
		t.Errorf("yyyy-qq: error %v, want the error of the unknown letter", err)
	}
}
