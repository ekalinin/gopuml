// SPDX-License-Identifier: MIT
// Provenance: original

package timing

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/jmath"
)

func decimalTick(s string) *TimeTick { return newTimeTick(jmath.BigDecimalOf(s), formatDecimal) }

func rulerOf(times ...string) *TimingRuler {
	r := newTimingRuler(nil)
	for _, s := range times {
		r.addTime(decimalTick(s))
	}
	return r
}

// panicMessage returns the Java exception of a panic, "" for none.
func panicMessage(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = r.(error).Error()
		}
	}()
	f()
	return ""
}

func TestTimingFormats(t *testing.T) {
	if got := formatDecimal.formatTime(jmath.BigDecimalOf("2.50")); got != "2.50" {
		t.Errorf("DECIMAL of 2.50 = %q", got)
	}
	if got := formatHour.formatTime(jmath.BigDecimalOfInt64(4500)); got != "1:15:00" {
		t.Errorf("HOUR of 4500 = %q", got)
	}
	date := createDate(2024, 7, 15, formatDate)
	if got := date.String(); got != "07/15" {
		t.Errorf("DATE of 2024/07/15 = %q", got)
	}
	lenient := createDate(2019, 13, 40, formatDate)
	if got := lenient.time.LongValue(); got != 1581206400 {
		t.Errorf("2019/13/40 = %d, want 1581206400 (2020-02-09)", got)
	}
	if got := lenient.String(); got != "02/09" {
		t.Errorf("2019/13/40 = %q, want 02/09", got)
	}
	if got := panicMessage(func() { formatHour.formatTime(jmath.BigDecimalOf("2.5")) }); got !=
		"java.lang.ArithmeticException: Rounding necessary" {
		t.Errorf("HOUR of 2.5 panics with %q", got)
	}
}

func TestRulerKeepsFirstTick(t *testing.T) {
	r := rulerOf("10", "10.0")
	if len(r.times) != 1 || r.times[0].time.ToPlainString() != "10" {
		t.Errorf("times %v, want the first tick 10 only", r.times)
	}
}

func TestRulerTickUnit(t *testing.T) {
	if got := rulerOf("100", "300").getOptimalTickUnit(); got != 100 {
		t.Errorf("unit of 100 and 300 = %d", got)
	}
	if got := rulerOf("0", "0.5").getOptimalTickUnit(); got != -1 {
		t.Errorf("unit of 0 and 0.5 = %d, want -1", got)
	}
	// A unit of 1 would be 50050 pixels wide: round(1 + 50*1000/4000).
	if got := rulerOf("0", "1", "1000").getOptimalTickUnit(); got != 14 {
		t.Errorf("unit of 0, 1 and 1000 = %d, want 14", got)
	}
	r := rulerOf("0", "100", "300")
	r.scaleInPixels(100, 120)
	if got := r.getOptimalTickUnit(); got != 100 || r.tickIntervalInPixels != 120 {
		t.Errorf("forced unit %d and interval %d", got, r.tickIntervalInPixels)
	}
	if got := r.getWidth(); got != 480 {
		t.Errorf("width %v, want (300/100+1)*120", got)
	}
}

func TestRulerEnsureNotEmpty(t *testing.T) {
	r := rulerOf()
	if r.getWidth() != 100 {
		t.Errorf("width of an empty ruler %v, want 100", r.getWidth())
	}
	r.ensureNotEmpty()
	if len(r.times) != 1 || r.times[0].time.Signum() != 0 {
		t.Errorf("empty ruler: times %v, want 0", r.times)
	}
	r = rulerOf("-10", "10")
	r.ensureNotEmpty()
	if len(r.times) != 3 || r.times[1].time.Signum() != 0 {
		t.Errorf("-10 and 10: times %v, want 0 between them", r.times)
	}
	if got := panicMessage(func() { rulerOf("-20", "-10").getFirstPositiveOrZeroValue() }); got !=
		"java.lang.IllegalStateException" {
		t.Errorf("times below zero panic with %q", got)
	}
}

func TestRulerFormatOfLastTick(t *testing.T) {
	r := rulerOf("0")
	r.addTime(newTimeTick(jmath.BigDecimalOfInt64(900), formatHour))
	r.addTime(decimalTick("3600"))
	if r.format != formatHour {
		t.Error("the format of the labels is not the one of the last tick that is not DECIMAL")
	}
}

func TestDeduceFormat(t *testing.T) {
	f := deduceFormatFrom("0").mergeWith(deduceFormatFrom("5.25"))
	if f.minFractionDigits != 0 || f.maxFractionDigits != 2 || f.decimalSep != '.' || f.groupingSep != 0 {
		t.Errorf("format of 0 and 5.25 = %+v", *f)
	}
	df := f.decimalFormat()
	for _, tt := range []struct {
		v    float64
		want string
	}{{3, "3"}, {2.5, "2.5"}, {1.005, "1"}} {
		if got := df.Format(tt.v); got != tt.want {
			t.Errorf("Format(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}
