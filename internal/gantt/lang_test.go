// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The behavior these tests expect was checked with the reference jar.

// parse builds the gantt diagram of the lines with the factory. It returns
// the diagram, or the message and the line of its error, the last line of
// the trace, as upstream reports it.
func parse(t *testing.T, lines ...string) (d *GanttDiagram, errMsg, errLine string) {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@startgantt\n" + strings.Join(lines, "\n") + "\n@endgantt\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	if e, ok := result.(*core.ErrorDiagram); ok {
		line := ""
		if n := len(e.Trace); n > 0 {
			line = e.Trace[n-1].S
		}
		return nil, e.Error.Error, line
	}
	return result.(*GanttDiagram), "", ""
}

// mustParse is parse for lines without error.
func mustParse(t *testing.T, lines ...string) *GanttDiagram {
	t.Helper()
	d, msg, line := parse(t, lines...)
	if d == nil {
		t.Fatalf("%q on %q", msg, line)
	}
	return d
}

// spans returns "id start end" for each task, as the TXT table writes
// the days.
func spans(d *GanttDiagram) string {
	var result []string
	for _, task := range d.modelData.Tasks() {
		if _, ok := task.(*TaskImpl); !ok {
			continue
		}
		result = append(result, task.Code().ID()+" "+span(task))
	}
	return strings.Join(result, ", ")
}

const projectStart = "Project starts 2020-01-06"

func TestDates(t *testing.T) {
	for _, tc := range []struct {
		line, want string
	}{
		{"Project starts the 5th of december 2018", "2018-12-05"},
		{"Project starts 5 jan 2020", "2020-01-05"},
		{"Project starts 5-dec-2018", "2018-12-05"},
		{"Project starts December 5th, 2018", "2018-12-05"},
		{"project starts on january 5, 2020", "2020-01-05"},
		{"Project starts 2018/12/29", "2018-12-29"},
		{"Project starts 2018-12-29", "2018-12-29"},
		{"Project starts 2020.1.5", "2020-01-05"},
		{"PROJECT STARTS the 6th of january 2020", "2020-01-06"},
		{"Project starts 5 dec 18", "0018-12-05"},
	} {
		if got := mustParse(t, tc.line).MinDay().String(); got != tc.want {
			t.Errorf("%q: %s, want %s", tc.line, got, tc.want)
		}
	}
	for _, line := range []string{"Project starts 1jan2020", "Project starts 5th of décembre 2020", "monday is closed",
		"\t[A] lasts 5 days", "[A]\tlasts 5 days", "[A] foo", "[T] <<s>> as [X] lasts 2 days",
		"[T] ends before [X]'s start"} {
		if _, msg, errLine := parse(t, line); msg != "Syntax Error?" || errLine != line {
			t.Errorf("%q: %q on %q", line, msg, errLine)
		}
	}
	if _, msg, _ := parse(t, "Project starts 2020-02-30"); msg != "You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.time.DateTimeException: Invalid date 'FEBRUARY 30'" {
		t.Errorf("invalid date: %q", msg)
	}
}

func TestTaskSentences(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"times of day", []string{"Project starts 2024-09-22", "[A] starts 2024-09-22T10:30:46",
			"[B] starts 2024-09-22t10:30", "[C] starts 2024-09-22 11:30:46"},
			"A 2024-09-22 2024-09-22, B 2024-09-22 2024-09-22, C 2024-09-22 2024-09-22"},
		{"durations", []string{projectStart, "[A] lasts 1 day, and 2 days", "[B] requires 1 week and 2 days",
			"[C] lasts 1 month"}, "A 2020-01-06 2020-01-08, B 2020-01-06 2020-01-14, C 2020-01-06 2020-02-04"},
		{"chained", []string{projectStart, "[A] lasts 3 days", "[B] starts at [A]'s end and lasts 2 days"},
			"A 2020-01-06 2020-01-08, B 2020-01-09 2020-01-10"},
		{"ignored trailing text", []string{projectStart, "[A] lasts 5 days foo bar",
			"[B] lasts 5 days, and starts 2020-01-10", "[C] lasts 5 days and"},
			"A 2020-01-06 2020-01-10, B 2020-01-06 2020-01-10, C 2020-01-06 2020-01-10"},
		{"delays", []string{projectStart, "saturday are closed", "sunday are closed", "[X] lasts 5 days",
			"[A] starts 2 days after [X]'s end", "[B] starts 2 working days after [X]'s end",
			"[C] starts 1 week after [X]'s end", "[D] starts 1 week and 2 days after [X]'s end",
			"[E] starts 3 days before [X]'s end", "[F] starts 3 working days before [X]'s end"},
			"X 2020-01-06 2020-01-10, A 2020-01-13 2020-01-13, B 2020-01-14 2020-01-14, C 2020-01-16 2020-01-16, " +
				"D 2020-01-20 2020-01-20, E 2020-01-08 2020-01-08, F 2020-01-08 2020-01-08"},
		{"then", []string{projectStart, "[A] lasts 3 days", "then [B] lasts 2 days", "Then [C] lasts 1 day",
			"then [A] lasts 1 day"},
			"A 2020-01-06 2020-01-06, B 2020-01-09 2020-01-10, C 2020-01-11 2020-01-11"},
		{"milestones", []string{projectStart, "[A] lasts 4 days", "[M1] happens at [A]'s end",
			"[M2] happens 2020-01-15", "[M3] happens 2 days after [Sprint Dec 1]'s end"},
			"A 2020-01-06 2020-01-09, M1 2020-01-09 2020-01-09, M2 2020-01-15 2020-01-15, M3 2020-01-06 2020-01-06"},
		{"occurs", []string{projectStart, "[S] happens 2020-01-08", "[E] happens 2020-01-13",
			"[D] occurs from [S] to [E]"},
			"S 2020-01-08 2020-01-08, E 2020-01-13 2020-01-13, D 2020-01-08 2020-01-13"},
		{"ends", []string{projectStart, "[A] lasts 3 days", "[B] starts at [A]'s end and lasts 4 days",
			"[C] lasts 3 days and ends at [B]'s end", "[D] ends D+4", "[E] starts D+1 and ends on 2020-01-09",
			"[F] ends 2020-01-10 12:00"},
			"A 2020-01-06 2020-01-08, B 2020-01-09 2020-01-12, C 2020-01-10 2020-01-12, D 2020-01-10 2020-01-10, " +
				"E 2020-01-07 2020-01-09, F 2020-01-09 2020-01-09"},
		{"pauses", []string{projectStart, "[A] lasts 3 days", "[A] pauses on 2020-01-07", "[B] lasts 3 days",
			"[B] pauses from 2020-01-07 to 2020-01-08", "[C] lasts 3 days", "[C] pauses from 2020-01-07 to the 2020-01-08",
			"[D] lasts 3 days", "[D] pauses on tuesday", "[E] lasts 3 days", "[E] pauses on 2020-01-07 to [A]'s start"},
			"A 2020-01-06 2020-01-09, B 2020-01-06 2020-01-10, C 2020-01-06 2020-01-09, D 2020-01-06 2020-01-09, " +
				"E 2020-01-06 2020-01-08"},
		{"resources", []string{projectStart, "[A] on {Alice} lasts 2 days", "[B] on {Bob:50%} lasts 2 days",
			"{Carol} is off on monday", "[C] on {Carol} {Bob:50%} lasts 2 days", "they is off on 2020-01-07"},
			"A 2020-01-06 2020-01-07, B 2020-01-06 2020-01-09, C 2020-01-06 2020-01-07"},
		{"days on", []string{projectStart, "{Bob} is off on monday", "{Bob} is on 2020-01-13",
			"[G] on {Bob} lasts 7 days", "{Ann} is off on monday", "{Ann} is on D+7 to D+7", "[H] on {Ann} lasts 7 days"},
			"G 2020-01-07 2020-01-13, H 2020-01-07 2020-01-13"},
		{"off before and after", []string{projectStart, "{Alice} is off before 2020-01-08",
			"{Alice} is off 2020-01-09 to 2020-01-10", "[A] on {Alice} lasts 1 day"},
			"A 2020-01-09 2020-01-09"},
	} {
		d := mustParse(t, tc.lines...)
		if got := spans(d); got != tc.want {
			t.Errorf("%s:\n got %s\nwant %s", tc.name, got, tc.want)
		}
	}
}

func TestCalendarSentences(t *testing.T) {
	d := mustParse(t, projectStart, "saturday are closed", "sunday are closed for [T]", "saturday are open for [Foo]",
		"2020-01-08 to 2020-01-09 are closed", "D+10 to D+11 are opened", "2020-01-13 is colored in salmon",
		"friday are colored in lightblue", "2020-01-06 and 2 days are named [Sprint 1]",
		"then 2 days are colored in lightgreen and are named [Sprint 2]", "E+1 is closed", "D+1 is colored in red",
		"today is 2020-01-08 and is colored in #AAF", "separator just before 2020-01-08",
		"separator just after 2020-01-09", "[A] lasts 2 days", "separator just at [A]'s end", "[Z] lasts 8 days",
		"From 8:00 to 12:00 are working hours")
	// E+1 is the day after the last colored or named day when the line is
	// read, and "then 2 days" starts on the last colored day. Z keeps the
	// work of 8 days of 24 hours, done 4 hours a day, as upstream.
	if got := spans(d); got != "A 2020-01-06 2020-01-07, Z 2020-01-06 2020-03-15" {
		t.Errorf("tasks %s", got)
	}
	for day, open := range map[int]bool{6: true, 7: true, 8: false, 9: false, 10: true, 11: false, 12: false, 13: true,
		14: true, 16: false, 17: true} {
		if got := d.IsOpen(date(2020, 1, day)); got != open {
			t.Errorf("January %d: open %v", day, got)
		}
	}
	if d.dayCalendar.OpenCloseForTask("Foo").IsClosed(date(2020, 1, 11)) {
		t.Error("day opened for one task")
	}
	if d.dayCalendar.DayColor(TimePointOfStartOfDay(date(2020, 1, 13))) == nil ||
		d.dayCalendar.DayOfWeekColor(jtime.Friday) == nil || d.dayCalendar.DayColor(TimePointOfStartOfDay(date(2020, 1, 7))) == nil {
		t.Error("colored days")
	}
	if got := d.dayCalendar.DayColor(TimePointOfStartOfDay(date(2020, 1, 8))); got == nil || got.(*color.HColorSimple) == nil {
		t.Error("today")
	}
	names := map[string]int{}
	for _, name := range d.dayCalendar.NameDays() {
		names[name]++
	}
	if names["Sprint 1"] != 3 || names["Sprint 2"] != 3 {
		t.Errorf("named days %v", names)
	}
	for _, day := range []jtime.LocalDate{date(2020, 1, 8), date(2020, 1, 10)} {
		if !d.dayCalendar.HasSeparatorBefore(day) {
			t.Errorf("no separator before %v", day)
		}
	}
	if !d.dayCalendar.HasWorkingTimeRanges() {
		t.Error("working hours")
	}
}

func TestSentenceErrors(t *testing.T) {
	for _, tc := range []struct {
		lines        []string
		msg, errLine string
	}{
		{[]string{projectStart, "[A] lasts 3 days and foo"}, "No matching verb phrase", "[A] lasts 3 days and foo"},
		{[]string{projectStart, "[A] lasts 3 days and "}, "No matching verb phrase", "[A] lasts 3 days and "},
		{[]string{"[A] starts 2020-01-10"}, "No starting date for the project", "[A] starts 2020-01-10"},
		{[]string{"[A] lasts 3 days", projectStart}, "Starting point must be set before task definition",
			projectStart},
		{[]string{"[A] starts at [Nope]'s end"}, "No such task Nope", "[A] starts at [Nope]'s end"},
		{[]string{"[A] lasts 1 day", "{Alice} works on [A]"}, "No such task [A]", "{Alice} works on [A]"},
		{[]string{"it lasts 2 days"}, "Not sure what are you refering to?", "it lasts 2 days"},
		{[]string{"she is off on monday"}, "Not sure who are you refering to?", "she is off on monday"},
		{[]string{"[A] on {Bob:0} lasts 1 day"}, "Bad argument for resource", "[A] on {Bob:0} lasts 1 day"},
		{[]string{"[A] displays on same row as [B]"}, "No such task null", "[A] displays on same row as [B]"},
		{[]string{"[A] occurs from [B] to [C]"}, "No such B task", "[A] occurs from [B] to [C]"},
	} {
		if _, msg, errLine := parse(t, tc.lines...); msg != tc.msg || errLine != tc.errLine {
			t.Errorf("%q: %q on %q, want %q on %q", tc.lines, msg, errLine, tc.msg, tc.errLine)
		}
	}
}

func TestTaskAttributes(t *testing.T) {
	d := mustParse(t, "[Prototype design] as [P] lasts 5 days", "[P] is colored in Lavender/LightBlue",
		"[Prototype design] lasts 2 days", "[T] <<important>> lasts 2 days", "[C] lasts 2 days and links to [[https://example.com]]",
		"[D] lasts 2 days and is displayed as Build step", "it is 40% completed", "[E] lasts 1 day", "[E] is deleted",
		"[F] is colored for completion from red to green", "[M] displays on same row as [C]",
		"[X] is colored in nocolor")
	tasks := d.modelData.Tasks()
	if len(tasks) != 9 || tasks[0].Code().Display() != "Prototype design" || tasks[1].Code().ID() != "Prototype design" {
		t.Fatalf("tasks %d", len(tasks))
	}
	if c := tasks[0].(*TaskImpl).Colors(); c == nil || c.Center() == nil || c.Border() == nil {
		t.Error("colors")
	}
	if s := tasks[2].Stereotype(); s == nil {
		t.Error("stereotype")
	}
	if u := tasks[3].(*TaskImpl).URL(); u == nil || u.URL() != "https://example.com" {
		t.Error("link")
	}
	if s, ok := tasks[4].DisplayString(); !ok || s != "Build step" || tasks[4].(*TaskImpl).Completion() != 40 {
		t.Errorf("display %q, completion %d", s, tasks[4].(*TaskImpl).Completion())
	}
	if c := tasks[5].(*TaskImpl).Colors(); c == nil || c.Center() != color.White || c.Border() != color.Black {
		t.Error("deleted")
	}
	if len(tasks[6].(*TaskImpl).colors) != 2 {
		t.Error("completion colors")
	}
	if tasks[7].Row() != tasks[3] {
		t.Error("same row")
	}
	if c := tasks[8].(*TaskImpl).Colors(); c == nil || c.Center() != color.White {
		t.Error("an unknown color is white")
	}
	links := mustParse(t, "[A] lasts 1 day", "[B] starts at [A]'s end with red dotted link",
		"[C] starts at [A]'s end with blue link").modelData.Constraints()
	if len(links) != 2 || links[0].SpecificColor() == nil || links[0].LinkType.Style() != decorationDotted() ||
		links[1].LinkType.Style() == decorationDotted() {
		t.Error("links")
	}
}

func TestCommands(t *testing.T) {
	d := mustParse(t, projectStart, "[A] lasts 1 day", "[B] lasts 1 day", "[A]->[B]", "[C]-[#red]->[D]",
		"A -[#blue,dashed]-> C", "[A] #red/#blue", "-- Phase 2 --", "--", "group [G]", "[E] lasts 1 day",
		"end group", "week starts on sunday and must have at least 1 day", "language DE",
		"printscale Weekly zoom 2", "print between 2020-01-10 and 2020-01-20", "note bottom", "a note", "end note",
		"hide footbox", "labels on First column", "hide resources names", "hide ressources footbox",
		"Show column duration", "show column task", "hide closed", "task default completion to 50", "[F] lasts 1 day")
	if got := spans(d); got != "A 2020-01-06 2020-01-06, B 2020-01-07 2020-01-07, C 2020-01-07 2020-01-07, "+
		"D 2020-01-07 2020-01-07, E 2020-01-06 2020-01-06, F 2020-01-06 2020-01-06" {
		t.Errorf("tasks %s", got)
	}
	if len(d.modelData.Constraints()) != 3 {
		t.Errorf("%d constraints", len(d.modelData.Constraints()))
	}
	var ids []string
	for _, task := range d.modelData.Tasks() {
		ids = append(ids, task.Code().ID())
	}
	if got := strings.Join(ids, " "); got != "A B C D ##4 ##5 G E F" {
		t.Errorf("tasks %q", got)
	}
	if d.weekConfig.weekNumberStrategy != (WeekNumberStrategy{jtime.Sunday, 1}) || d.weekConfig.locale != "de" {
		t.Error("week strategy or language")
	}
	if d.scaleConfig.printScale != PrintScaleDaily || d.scaleConfig.factorScale != 2 {
		t.Errorf("print scale %v zoom %v", d.scaleConfig.printScale, d.scaleConfig.factorScale)
	}
	if s, ok := d.timeBounds.PrintStart(); !ok || s != date(2020, 1, 10) || d.timeBounds.PrintEnd() != date(2020, 1, 20) {
		t.Error("print between")
	}
	if d.displayConfig.showFootbox || !d.displayConfig.hideResourceName || !d.displayConfig.hideResourceFootbox ||
		!d.scaleConfig.hideClosed || d.defaultCompletion != 50 {
		t.Error("display options")
	}
	if got := d.DisplayedColumn(); len(got) != 3 || got[0] != ColumnTask || got[2] != ColumnEnd {
		t.Errorf("columns %v", got)
	}
	if w := d.Warnings(); len(w) != 1 || w[0].AsSingleLine() != "'labels on First column' is no longer supported: this line is ignored" {
		t.Errorf("warnings %v", w)
	}
	if note := d.modelData.Tasks()[5].(*TaskSeparator); note == nil {
		t.Error("note on a separator")
	}
	d = mustParse(t, projectStart, "printscale weekly with week numbering from 3", "[A] lasts 1 day")
	if d.weekConfig.weeklyHeaderStrategy != FromN || d.weekConfig.weekStartingNumber != 3 ||
		d.scaleConfig.printScale != PrintScaleWeekly {
		t.Error("week numbering")
	}
	d = mustParse(t, projectStart, "projectscale weekly with calendar date", "[A] lasts 1 day", "note bottom",
		"text", "end note")
	if d.weekConfig.weeklyHeaderStrategy != DayOfMonth || d.modelData.Tasks()[0].(*TaskImpl).Note() == nil {
		t.Error("calendar date or note")
	}
	for _, tc := range []struct {
		lines        []string
		msg, errLine string
	}{
		{[]string{"A -> B"}, "No such task A", "A -> B"},
		{[]string{"[A] lasts 1 day", "[A] #nocolor"}, "No such color", "[A] #nocolor"},
		{[]string{"[A] lasts 1 day", "note bottom", "end note"}, "No note defined", "end note"},
		{[]string{"note bottom", "text", "end note"}, "No task defined", "end note"},
		{[]string{"footbox"}, "You should send a mail to plantuml@gmail.com with this log (V1.2026.9) " +
			`java.lang.NullPointerException: Cannot invoke "String.equalsIgnoreCase(String)" because the return value of "net.sourceforge.plantuml.regex.RegexResult.get(String, int)" is null`,
			"footbox"},
		{[]string{"task default completion to 101"}, "Completetion must between 0 and 100",
			"task default completion to 101"},
		{[]string{"end group"}, "No group to be closed", "end group"},
		{[]string{"-- Phase 2"}, "Syntax Error?", "-- Phase 2"},
	} {
		if _, msg, errLine := parse(t, tc.lines...); msg != tc.msg || errLine != tc.errLine {
			t.Errorf("%q: %q on %q, want %q on %q", tc.lines, msg, errLine, tc.msg, tc.errLine)
		}
	}
}

func decorationDotted() decoration.LinkStyle { return decoration.LinkStyleDotted() }
