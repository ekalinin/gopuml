// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// The expected values of these tests were printed by upstream's classes of
// the reference jar.

// rec is a UGraphic that records the lines, rectangles and texts drawn,
// with the translation they are drawn at.
type rec struct {
	klimt.UGraphic
	log    *[]string
	dx, dy float64
}

func newRec() *rec { return &rec{UGraphic: klimt.NewUGraphicNull(), log: new([]string)} }

func (r *rec) StringBounder() font.StringBounder { return font.NewStringBounderSvg(nil) }

func (r *rec) Apply(change klimt.UChange) klimt.UGraphic {
	c := *r
	if t, ok := change.(klimt.UTranslate); ok {
		c.dx += t.Dx()
		c.dy += t.Dy()
	}
	return &c
}

func d2s(d float64) string { return jfmt.DoubleToString(d) }

func (r *rec) Draw(shape klimt.UShape) {
	at := d2s(r.dx) + "," + d2s(r.dy) + " "
	switch s := shape.(type) {
	case *klimt.ULine:
		*r.log = append(*r.log, at+"line "+d2s(s.DX())+"x"+d2s(s.DY()))
	case *klimt.URectangle:
		*r.log = append(*r.log, at+"rect "+d2s(s.Width())+"x"+d2s(s.Height()))
	case *klimt.UText:
		*r.log = append(*r.log, at+"text "+s.Text())
	}
}

func (r *rec) texts() []string {
	var result []string
	for _, l := range *r.log {
		if _, text, ok := strings.Cut(l, " text "); ok {
			result = append(result, text)
		}
	}
	return result
}

func TestTimeScales(t *testing.T) {
	jan6 := TimePointOfStartOfDayYMD(2020, 1, 6)
	daily := NewTimeScaleDaily(16, jan6, 2.0/3.0, nil)
	oc := NewOpenClose()
	oc.CloseDayOfWeek(jtime.Saturday)
	oc.CloseDayOfWeek(jtime.Sunday)
	hc := NewTimeScaleDailyHideClosed(16, jan6, 1.0, oc)
	weekly := NewTimeScaleWink(16, 0.25, PrintScaleWeekly)
	for _, tc := range []struct{ got, want string }{
		{d2s(NewTimeScaleWink(16, 1.0, PrintScaleDaily).Position(jan6)), "292272.0"},
		{fmt.Sprint(d2s(weekly.Position(jan6)), " ", weekly.IsBreaking(jan6), " ",
			weekly.IsBreaking(TimePointOfStartOfDayYMD(2020, 1, 8))), "73068.0 false true"},
		{fmt.Sprint(d2s(daily.Position(jan6)), " ", d2s(daily.Position(TimePointOfStartOfDayYMD(2020, 1, 9))), " ",
			d2s(daily.Position(TimePointOf(jtime.DateTimeOfFields(2020, 1, 7, 10, 30, 46))))),
			"0.0 32.0 15.339012345677475"},
		{fmt.Sprint(d2s(hc.Position(TimePointOfStartOfDayYMD(2020, 1, 10))), " ",
			d2s(hc.Position(TimePointOfStartOfDayYMD(2020, 1, 11))), " ",
			d2s(hc.Position(TimePointOfStartOfDayYMD(2020, 1, 13))), " ",
			d2s(hc.Width(TimePointOfStartOfDayYMD(2020, 1, 12)))), "64.0 80.0 80.0 0.0"},
	} {
		if tc.got != tc.want {
			t.Errorf("%q, want %q", tc.got, tc.want)
		}
	}
	if got := panicMessage(t, func() { hc.Position(TimePointOfStartOfDayYMD(2020, 1, 5)) }); got != "java.lang.IllegalArgumentException" {
		t.Errorf("%q", got)
	}
	if got := panicMessage(t, func() { hc.Position(TimePointOf(jtime.DateTimeOfFields(2020, 1, 14, 10, 0, 0))) }); got != `java.lang.NullPointerException: Cannot invoke "java.lang.Integer.intValue()" because the return value of "java.util.Map.get(Object)" is null` {
		t.Errorf("%q", got)
	}
	var ranges TimeRanges
	ranges.Add(NewTimeRange("9:00", "12:00"))
	ranges.Add(NewTimeRange("14:00", "18:00"))
	wh := &WorkingHours{&ranges, 120, 0.2}
	if got := fmt.Sprint(d2s(wh.WorkingDurationForOneDay(false)), " ", d2s(wh.WorkingDurationForOneDay(true)), " ",
		d2s(wh.PartialDayWidth(jtime.TimeOf(10, 30, 46))), " ", d2s(wh.PartialDayWidth(jtime.TimeOf(13, 0, 0))), " ",
		d2s(wh.PartialDayWidth(jtime.MaxTime)), " ", d2s(wh.PartialDayWidth(jtime.Midnight))); got != "52.0 24.0 16.56388888888889 25.0 51.999722222222225 0.0" {
		t.Errorf("working hours %s", got)
	}
	variant := &VariantTimeScale{jtime.DateOf(2020, 1, 6), jtime.DateOf(2020, 1, 10), wh, oc}
	if got := fmt.Sprint(d2s(variant.Position(TimePointOf(jtime.DateTimeOfFields(2020, 1, 8, 10, 30, 46)))), " ",
		d2s(variant.Position(TimePointOfStartOfDayYMD(2020, 1, 13))), " ",
		d2s(variant.Width(TimePointOfStartOfDayYMD(2020, 1, 11)))); got != "120.5638888888889 308.0 24.0" {
		t.Errorf("variant %s", got)
	}
}

func TestTimeHeaders(t *testing.T) {
	header := func(lines ...string) TimeHeader {
		d := mustParse(t, lines...)
		d.initMinMax()
		return d.newTimeHeader(&TimelineStyleData{d, d.ColorSet(), d.SkinParam()})
	}
	sb := font.NewStringBounderSvg(nil)
	for _, tc := range []struct {
		lines                []string
		header, footer, full string
	}{
		{[]string{projectStart, "[A] lasts 3 days"}, "39.0", "40.0", "39.0"},
		{[]string{projectStart, "2020-01-06 to 2020-01-06 are named [N]", "[A] lasts 3 days"}, "39.0", "40.0", "55.0"},
		{[]string{projectStart, "printscale weekly", "[A] lasts 3 days"}, "27.0", "16.0", "27.0"},
		{[]string{projectStart, "printscale monthly", "[A] lasts 3 days"}, "30.0", "31.0", "30.0"},
		{[]string{projectStart, "printscale quarterly", "[A] lasts 3 days"}, "31.0", "31.0", "31.0"},
		{[]string{projectStart, "printscale yearly", "[A] lasts 3 days"}, "17.0", "17.0", "17.0"},
		{[]string{"[A] lasts 3 days"}, "16.0", "16.0", "16.0"},
		{[]string{projectStart, "from 9:00 to 17:00 are working hours", "[A] lasts 3 days"}, "39.0", "0.0", "39.0"},
	} {
		h := header(tc.lines...)
		got := []string{d2s(h.TimeHeaderHeight(sb)), d2s(h.TimeFooterHeight(sb)), d2s(h.FullHeaderHeight(sb))}
		if want := []string{tc.header, tc.footer, tc.full}; !slices.Equal(got, want) {
			t.Errorf("%q: heights %v, want %v", tc.lines, got, want)
		}
	}
	r := newRec()
	header("Project starts 2024-12-23", "printscale weekly", "[A] lasts 20 days").DrawTimeHeader(r, 100)
	if got := strings.Join(r.texts(), " "); got != "52 53 2 Dec Jan" {
		t.Errorf("weekly labels %q", got)
	}
}
