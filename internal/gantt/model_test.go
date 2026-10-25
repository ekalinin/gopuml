// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/text"
)

// The expected dates of these tests come from the TXT output of the
// reference jar for the same lines.

func newTestDiagram(t *testing.T) *GanttDiagram {
	t.Helper()
	src := core.NewUmlSource([]text.StringLocated{{S: "@startgantt"}, {S: "@endgantt"}}, nil, true)
	return NewGanttDiagram(core.NewRenderContext(), src, nil)
}

func date(y, m, d int) jtime.LocalDate { return jtime.DateOf(y, m, d) }

// span returns the start and the end of a task as the TXT table writes
// them: the day of the start and the day before the end.
func span(task Task) string {
	return task.Start().Day().String() + " " + endMinusOneDay(task).Day().String()
}

func task(d *GanttDiagram, id string) Task { return d.OrCreateTask(TaskCodeFromID(id), false) }

func lasts(d *GanttDiagram, t Task, days int) { t.SetLoad(ngm.LoadOf(d.DurationOfDays(days))) }

func TestScheduling(t *testing.T) {
	t.Run("weeks of open days", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.CloseDayOfWeek(jtime.Saturday, "")
		d.CloseDayOfWeek(jtime.Sunday, "")
		a := task(d, "A")
		lasts(d, a, 2*d.DaysInWeek())
		if got := span(a); got != "2020-01-06 2020-01-17" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("delays", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.CloseDayOfWeek(jtime.Saturday, "")
		d.CloseDayOfWeek(jtime.Sunday, "")
		a := task(d, "A")
		lasts(d, a, 5)
		end := NewTaskInstant(a, TaskAttributeEnd)
		b := task(d, "B")
		b.SetStart(end.WithDelta(2, IgnoreCalendar, d.DefaultPlan()).InstantPrecise())
		c := task(d, "C")
		c.SetStart(end.WithDelta(2, DoNotCountCloseDay, d.DefaultPlan()).InstantPrecise())
		if got := span(b) + " " + span(c); got != "2020-01-13 2020-01-13 2020-01-14 2020-01-14" {
			t.Errorf("B and C %s", got)
		}
	})
	t.Run("start before the project start", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2024, 1, 1))
		a := task(d, "A")
		a.SetStart(TimePointOfStartOfDay(date(2023, 12, 25)))
		if got := span(a); got != "2024-01-01 2024-01-01" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("start, end and load", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		a := task(d, "A")
		a.SetStart(TimePointOfStartOfDay(date(2020, 1, 8)))
		a.SetEnd(TimePointOfStartOfDay(date(2020, 1, 11)))
		lasts(d, a, 1)
		if got := span(a); got != "2020-01-10 2020-01-10" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("cached end", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2024, 1, 1))
		a := task(d, "A")
		lasts(d, a, 10)
		b := task(d, "B")
		b.SetStart(NewTaskInstant(a, TaskAttributeEnd).InstantPrecise())
		c := task(d, "C")
		lasts(d, c, 10)
		d.CloseDayOfWeek(jtime.Saturday, "")
		d.CloseDayOfWeek(jtime.Sunday, "")
		if got := span(a) + " " + span(c); got != "2024-01-01 2024-01-10 2024-01-01 2024-01-12" {
			t.Errorf("A and C %s", got)
		}
	})
	t.Run("work frozen before working hours", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		a := task(d, "A")
		lasts(d, a, 2)
		d.AddWorkingTimeRange(NewTimeRange("9:00", "17:00"))
		if got := i18nDurationHumanReadable(LocaleEnglish, jtime.Between(a.Start().LocalDateTime(), a.End().LocalDateTime())); got != "5 days, 17 hours" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("work after working hours", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.AddWorkingTimeRange(NewTimeRange("9:00", "17:00"))
		a := task(d, "A")
		lasts(d, a, 2)
		if got := i18nDurationHumanReadable(LocaleEnglish, jtime.Between(a.Start().LocalDateTime(), a.End().LocalDateTime())); got != "1 day, 17 hours" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("resources", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.Resource("Alice").AddCloseDayOfWeek(jtime.Monday)
		a := task(d, "A")
		d.AffectResource(a, "Alice")
		lasts(d, a, 3)
		b := task(d, "B")
		d.AffectResource(b, "Bob:50")
		lasts(d, b, 2)
		if got := span(a) + " " + span(b); got != "2020-01-07 2020-01-09 2020-01-06 2020-01-09" {
			t.Errorf("A and B %s", got)
		}
		if got := b.(*TaskImpl).PrettyDisplay(); got != "B {Bob:50%}" {
			t.Errorf("B %q", got)
		}
		if d.AffectResource(b, "Carol:0") {
			t.Error("a resource at 0% is affected")
		}
	})
	t.Run("pause", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		a := task(d, "A")
		lasts(d, a, 3)
		a.AddPauseDay(date(2020, 1, 7))
		if got := span(a); got != "2020-01-06 2020-01-09" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("opened date on a closed day", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.CloseDayOfWeek(jtime.Saturday, "")
		d.CloseDayOfWeek(jtime.Sunday, "")
		d.OpenDayAsDate(date(2020, 1, 11), "")
		a := task(d, "A")
		lasts(d, a, 8)
		if got := span(a); got != "2020-01-06 2020-01-14" {
			t.Errorf("A %s", got)
		}
	})
	t.Run("day opened for one task", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		d.CloseDayOfWeek(jtime.Saturday, "")
		d.CloseDayOfWeek(jtime.Sunday, "")
		d.OpenDayOfWeek(jtime.Saturday, "Foo")
		bar := task(d, "Bar")
		lasts(d, bar, 6)
		foo := task(d, "Foo")
		lasts(d, foo, 6)
		if got := span(bar) + " " + span(foo); got != "2020-01-06 2020-01-13 2020-01-06 2020-01-11" {
			t.Errorf("Bar and Foo %s", got)
		}
	})
	t.Run("days as dates", func(t *testing.T) {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2020, 1, 6))
		var got []string
		for _, day := range NewDaysAsDatesCount(d, date(2020, 1, 6), 2).Dates() {
			got = append(got, day.String())
		}
		if strings.Join(got, " ") != "2020-01-06 2020-01-07 2020-01-08" {
			t.Errorf("dates %v", got)
		}
	})
}

func TestModelErrors(t *testing.T) {
	d := newTestDiagram(t)
	if got := panicMessage(t, func() { NewTaskGroup(nil, d.CurrentStyleBuilder(), TaskCodeFromID("G")).Start() }); got != "java.lang.UnsupportedOperationException" {
		t.Errorf("empty group: %q", got)
	}
	if got := panicMessage(t, func() { NewTimeRange("8", "9:00") }); got != "java.lang.IllegalArgumentException: Invalid time: 8" {
		t.Errorf("time range: %q", got)
	}
	r := NewTimeRange(" 8 : 05 ", "1_0:00")
	if got := r.String(); got != "08:05-null" {
		t.Errorf("time range %q", got)
	}
	if got := panicMessage(t, func() { r.Duration() }); got != "java.lang.NullPointerException: temporal" {
		t.Errorf("duration: %q", got)
	}
	task(d, "A")
	if res := d.UpdateStartingPoint(date(2020, 1, 6)); res.Err() != "Starting point must be set before task definition" {
		t.Errorf("starting point: %q", res.Err())
	}
	if res := d.EndGroup(); res.Err() != "No group to be closed" {
		t.Errorf("end group: %q", res.Err())
	}
	if res := newTestDiagram(t).AddNote(nil, nil); res.Err() != "No task defined" {
		t.Errorf("note: %q", res.Err())
	}
	for _, tc := range []struct {
		days int
		want string
	}{{7301, ""}, {7302, "Gantt diagrams cannot last more than 20 years"}} {
		d := newTestDiagram(t)
		d.UpdateStartingPoint(date(2000, 1, 1))
		lasts(d, task(d, "A"), tc.days)
		if got := d.CheckFinalError(); got != tc.want {
			t.Errorf("%d days: %q", tc.days, got)
		}
	}
}

func TestDiagramModel(t *testing.T) {
	d := newTestDiagram(t)
	d.UpdateStartingPoint(date(2020, 1, 6))
	d.AddGroup(TaskCodeFromID("G"))
	a := task(d, "A")
	lasts(d, a, 2)
	b := d.OrCreateTask(TaskCodeFromID("B"), true)
	d.EndGroup()
	d.AddSeparator("Next", true)
	var ids []string
	for _, t := range d.modelData.Tasks() {
		ids = append(ids, t.Code().ID())
	}
	if got := strings.Join(ids, " "); got != "G A B ##3" {
		t.Errorf("tasks %q", got)
	}
	if got := span(b) + " " + span(d.ExistingTask("G")); got != "2020-01-08 2020-01-08 2020-01-06 2020-01-08" {
		t.Errorf("then and group %s", got)
	}
	if len(d.modelData.Constraints()) != 1 {
		t.Error("then adds no arrow")
	}
	if got := d.ThenDate().Day().String(); got != "2020-01-06" {
		t.Errorf("then date %s", got)
	}
	d.ColorDay(date(2020, 1, 20), nil)
	d.NameDay(date(2020, 1, 9), "Sprint")
	d.NameDay(date(2020, 1, 10), "Sprint")
	if got := d.ThenDate().Day().String(); got != "2020-01-20" {
		t.Errorf("then date %s", got)
	}
	if m := d.ExistingMoment("SPRINT"); m == nil || m.Start().Day() != date(2020, 1, 9) || m.End().Day() != date(2020, 1, 11) {
		t.Errorf("named days %v", m)
	}
	defer func(f func() time.Time) { now = f }(now)
	now = func() time.Time { return time.Date(2021, 3, 4, 12, 0, 0, 0, time.Local) }
	if got := newTestDiagram(t).Today(); got != date(2021, 3, 4) {
		t.Errorf("today %v", got)
	}
}
