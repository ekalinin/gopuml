// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/export"
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
	case *klimt.UPath:
		*r.log = append(*r.log, at+"path "+s.String())
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

// laidOut parses the lines and lays the diagram out, as its SVG export
// does.
func laidOut(t *testing.T, lines ...string) *GanttDiagram {
	t.Helper()
	d := mustParse(t, lines...)
	d.TextBlock(0, svgFormat())
	return d
}

func drawOf(d *GanttDiagram, id string) TaskDraw { return d.drawRegistry.TaskDraw(d.ExistingTask(id)) }

// TestTaskDraws checks the drawing of the tasks of a diagram whose SVG is
// the one of the reference image.
func TestTaskDraws(t *testing.T) {
	sb := font.NewStringBounderSvg(nil)
	labels := LabelStrategy{LabelLegacy, 0}
	d := laidOut(t, projectStart, "[A] lasts 5 days", "[Long name] as [B] lasts 1 day", "[M] happens 2020-01-08",
		"[C] lasts 4 days and is 25% completed", "[P] lasts 3 days", "[P] pauses on 2020-01-07",
		"[N] lasts 5 days", "note bottom", "Long very very very Long Long note", "end note", "[After] lasts 3 days",
		"[R] on {Alice} lasts 2 days")
	for _, tc := range []struct {
		id, want string
	}{
		// inside the bar, 6 after its start, and after the end of a bar too short
		{"A", "6.0,12.21044921875 text A"},
		{"B", "18.0,12.21044921875 text Long name"},
		// right of the diamond, whose height is the font size made even
		{"M", "48.0,12.21044921875 text M"},
	} {
		r := newRec()
		drawOf(d, tc.id).DrawTitle(r, labels, 0, 0)
		if got := strings.Join(*r.log, "|"); got != tc.want {
			t.Errorf("title of %s: %q, want %q", tc.id, got, tc.want)
		}
	}
	if got := drawOf(d, "M").(*TaskDrawDiamond).diamondHeight(); got != 10 {
		t.Errorf("diamond height %v", got)
	}
	// The note of N moves the next row down.
	n, after := drawOf(d, "N"), drawOf(d, "After")
	if got, min := after.RealY(sb).CurrentValue(), n.RealY(sb).CurrentValue()+n.HeightMax(sb); got < min {
		t.Errorf("row after the note at %v, the note ends at %v", got, min)
	}
	for _, tc := range []struct {
		id, want string
	}{
		// the done part in the color of the task, the rest in the undone color
		{"C", "2.0,2.0 rect 15.0x12.8046875|17.0,2.0 rect 45.0x12.8046875|2.0,2.0 rect 60.0x12.8046875"},
		// two pieces around the paused day, joined by dotted lines
		{"P", "2.0,2.0 rect 15.0x12.8046875|2.0,2.0 path|32.0,2.0 rect 30.0x12.8046875|32.0,2.0 path|" +
			"19.0,2.0 line 10.0x0.0|19.0,14.8046875 line 10.0x0.0"},
	} {
		r := newRec()
		drawOf(d, tc.id).DrawU(r)
		var got []string
		for _, l := range *r.log {
			if i := strings.Index(l, " path "); i >= 0 {
				l = l[:i+5]
			}
			got = append(got, l)
		}
		if strings.Join(got, "|") != tc.want {
			t.Errorf("bar of %s: %q, want %q", tc.id, strings.Join(got, "|"), tc.want)
		}
	}
	r := newRec()
	d.ExistingTask("R")
	d.modelData.Resource("Alice").ResourceDraw().DrawU(r)
	if got := strings.Join(r.texts(), " "); got != "Alice 100 100" {
		t.Errorf("resource %q", got)
	}
}

func svgFormat() export.FileFormatOption {
	return export.NewFileFormatOptionMetadata(font.FormatSVG, true)
}
