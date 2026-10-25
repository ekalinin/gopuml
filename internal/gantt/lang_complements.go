// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/AbstractComplementTaskInstant.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementAnything.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementBeforeOrAfterOrAtTaskStartOrEnd.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementClose.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementCompleted.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementDate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementDayOfWeek.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementDeleted.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementDuration.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementIntervals.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementIntervalsSmart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementOpen.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementTimePoint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// taskCodeUbrex is SubjectTask.taskCode: a task between brackets, its name
// captured under name.
func taskCodeUbrex(name string) klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("["), klimt.UBrexNamed(name, klimt.UBrexLeaf("〇+「〤[]」")),
		klimt.UBrexLeaf("]"))
}

// The forms of upstream's ComplementDate.
type dateType int

const (
	dateAny dateType = iota
	dateOnlyRelative
	dateOnlyAbsolute
)

// complementDate is upstream's ComplementDate: an absolute date, or
// "N days after start", or "D+N".
type complementDate struct{ typ dateType }

func (c complementDate) complementUbrex() klimt.UBrexPart {
	days := newDayPattern("")
	switch c.typ {
	case dateOnlyAbsolute:
		return days.ubrex()
	case dateOnlyRelative:
		return klimt.UBrexOr(dateUbrexD(), dateUbrexE())
	}
	return klimt.UBrexOr(days.ubrex(), dateUbrexD(), dateUbrexE())
}

func dateUbrexD() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexNamed("DCOUNT", klimt.UBrexLeaf("〇+〴d")), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexLeaf("day〇?s"), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("after"),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("start"))
}

func dateUbrexE() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("「dD」+"), klimt.UBrexNamed("ECOUNT", klimt.UBrexLeaf("〇+〴d")))
}

func (c complementDate) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	if result, ok := newDayPattern("").day(arg); ok {
		return FailableOk[any](result)
	}
	if v, ok := arg.FindFirstValueByKey("DCOUNT"); ok {
		return FailableOk[any](gantt.MinDay().PlusDays(int64(parseInt(v))))
	}
	if v, ok := arg.FindFirstValueByKey("ECOUNT"); ok {
		return FailableOk[any](gantt.MinDay().PlusDays(int64(parseInt(v))))
	}
	uml.Throw(uml.IllegalStateException, "")
	return Failable[any]{}
}

// complementTimePoint is upstream's ComplementTimePoint: a date and a time
// of day, with T, t or a space between them.
type complementTimePoint struct{}

func (complementTimePoint) complementUbrex() klimt.UBrexPart {
	digits := klimt.UBrexLeaf("〇{1-2}〴d")
	seconds := klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf(":"), klimt.UBrexNamed("SECOND", digits)))
	return klimt.UBrexConcat(newDayPattern("").ubrex(), klimt.UBrexConcat(klimt.UBrexLeaf("「Tt∙」"),
		klimt.UBrexNamed("HOUR", digits), klimt.UBrexLeaf(":"), klimt.UBrexNamed("MINUTE", digits), seconds))
}

func (complementTimePoint) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	day, ok := newDayPattern("").day(arg)
	if !ok {
		uml.Throw(uml.IllegalStateException, "")
	}
	hour := parseInt(firstValue(arg, "HOUR"))
	minute := parseInt(firstValue(arg, "MINUTE"))
	second := 0
	if s, ok := arg.FindFirstValueByKey("SECOND"); ok {
		second = parseInt(s)
	}
	return FailableOk[any](TimePointOf(jtime.DateTimeOf(day, jtime.TimeOf(hour, minute, second))))
}

// complementDuration is upstream's ComplementDuration: numbers of hours,
// minutes, seconds, days, weeks or months, joined by "and" or commas.
type complementDuration struct{}

func (complementDuration) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(durationElement("0"),
		klimt.UBrexZeroOrMore(klimt.UBrexConcat(klimt.UBrexLeaf("【〇+「∙,」and〇+「∙,」┇〇+「∙,」】"), durationElement("1"))))
}

func durationElement(prefix string) klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexNamed("CNUM"+prefix, klimt.UBrexLeaf("〇+〴d")), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("CUNIT"+prefix, klimt.UBrexLeaf("【hour┇minute┇second┇day┇week┇month】")),
		klimt.UBrexLeaf("〇?s"))
}

func (complementDuration) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	// days, hours, minutes and seconds, as Java ints
	var totals [4]int32
	accumulateDuration(gantt, &totals, firstValue(arg, "CNUM0"), firstValue(arg, "CUNIT0"))
	cnum, cunit := arg.FindValuesByKey("CNUM1"), arg.FindValuesByKey("CUNIT1")
	for i := range cnum {
		accumulateDuration(gantt, &totals, cnum[i], cunit[i])
	}
	duration := gantt.DurationOfDaysHoursMinutesSeconds(int(totals[0]), int(totals[1]), int(totals[2]),
		int(totals[3]))
	return FailableOk[any](ngm.LoadOf(duration))
}

func accumulateDuration(gantt *GanttDiagram, totals *[4]int32, num, unit string) {
	value := int32(parseInt(num))
	switch unit[0] {
	case 'H', 'h':
		totals[1] += value
	case 'D', 'd':
		totals[0] += value
	case 'W', 'w':
		totals[0] += value * int32(gantt.DaysInWeek())
	case 'S', 's':
		totals[3] += value
	case 'M', 'm':
		// "mi" is a minute, otherwise a month.
		if len(unit) >= 2 && (unit[1] == 'i' || unit[1] == 'I') {
			totals[2] += value
		} else {
			totals[0] += value * int32(gantt.DaysInMonth())
		}
	default:
		uml.Throw(uml.IllegalArgumentException, "unknown time unit: "+unit)
	}
}

// taskInstantComplement is upstream's AbstractComplementTaskInstant: the
// start or end of a task or of named days, with an optional delay.
func taskInstantComplement(gantt *GanttDiagram, arg captureLookup) Failable[*TaskInstant] {
	code := firstValue(arg, "COMPLEMENT_CODE_OTHER")
	startOrEnd := firstValue(arg, "COMPLEMENT_START_OR_END")
	task := gantt.ExistingMoment(code)
	if task == nil {
		return FailableError[*TaskInstant]("No such task " + code)
	}
	result := NewTaskInstant(task, TaskAttributeFromString(startOrEnd))
	if nb1, ok := arg.FindFirstValueByKey("COMPLEMENT_NB1"); ok {
		factor1 := int32(1)
		if strings.HasPrefix(firstValue(arg, "COMPLEMENT_DAY_OR_WEEK1"), "w") {
			factor1 = int32(gantt.DaysInWeek())
		}
		days1 := int32(parseInt(nb1)) * factor1
		days2 := int32(0)
		if nb2, ok := arg.FindFirstValueByKey("COMPLEMENT_NB2"); ok {
			factor2 := int32(1)
			if strings.HasPrefix(firstValue(arg, "COMPLEMENT_DAY_OR_WEEK2"), "w") {
				factor2 = int32(gantt.DaysInWeek())
			}
			days2 = int32(parseInt(nb2)) * factor2
		}
		delta := days1 + days2
		if uml.EqualsIgnoreCase("before", firstValue(arg, "COMPLEMENT_BEFORE_OR_AFTER")) {
			delta = -delta
		}
		_, working1 := arg.FindFirstValueByKey("COMPLEMENT_WORKING1")
		_, working2 := arg.FindFirstValueByKey("COMPLEMENT_WORKING2")
		mode := IgnoreCalendar
		if working1 || working2 {
			mode = DoNotCountCloseDay
		}
		result = result.WithDelta(int(delta), mode, gantt.DefaultPlan())
	}
	return FailableOk(result)
}

// durationBeforeOrAfter is the delay of a task instant: "2 days after",
// "1 week and 2 working days before".
func durationBeforeOrAfter() klimt.UBrexPart {
	return klimt.UBrexConcat(
		klimt.UBrexNamed("COMPLEMENT_NB1", klimt.UBrexLeaf("〇+〴d")), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexNamed("COMPLEMENT_WORKING1", klimt.UBrexLeaf("working")),
			klimt.UBrexSpaceOneOrMore())),
		klimt.UBrexNamed("COMPLEMENT_DAY_OR_WEEK1", klimt.UBrexLeaf("【day┇week】")), klimt.UBrexLeaf("〇?s"),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("and"),
			klimt.UBrexSpaceOneOrMore(), klimt.UBrexNamed("COMPLEMENT_NB2", klimt.UBrexLeaf("〇+〴d")),
			klimt.UBrexSpaceOneOrMore(),
			klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexNamed("COMPLEMENT_WORKING2", klimt.UBrexLeaf("working")),
				klimt.UBrexSpaceOneOrMore())),
			klimt.UBrexNamed("COMPLEMENT_DAY_OR_WEEK2", klimt.UBrexLeaf("【day┇week】")), klimt.UBrexLeaf("〇?s"))),
		klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("COMPLEMENT_BEFORE_OR_AFTER", wordsOneOf(wordBefore, wordAfter)))
}

// complementTaskInstant is upstream's
// ComplementBeforeOrAfterOrAtTaskStartOrEnd: "[T]'s start", "at [T]'s
// end", "2 days after [T]'s end".
type complementTaskInstant struct{}

func (complementTaskInstant) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(
		klimt.UBrexOptional(klimt.UBrexOr(klimt.UBrexLeaf(wordAt), klimt.UBrexLeaf(wordWith),
			klimt.UBrexLeaf(wordAfter), durationBeforeOrAfter())),
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexConcat(taskCodeUbrex("COMPLEMENT_CODE_OTHER"), klimt.UBrexLeaf("〴.s")),
		klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("COMPLEMENT_START_OR_END", wordsOneOf(wordStart, wordEnd)))
}

func (complementTaskInstant) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	r := taskInstantComplement(gantt, arg)
	if r.IsFail() {
		return FailableError[any](r.Error())
	}
	return FailableOk[any](r.Get())
}

// complementIntervalsSmart is upstream's ComplementIntervalsSmart: a date
// to the start or end of a task.
type complementIntervalsSmart struct{}

func (complementIntervalsSmart) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(newDayPattern("1").ubrex(), wordsExactly(wordTo), klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOptional(durationBeforeOrAfter()), klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexConcat(taskCodeUbrex("COMPLEMENT_CODE_OTHER"), klimt.UBrexLeaf("〴.s")),
		klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("COMPLEMENT_START_OR_END", wordsOneOf(wordStart, wordEnd)))
}

func (complementIntervalsSmart) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	d1, _ := newDayPattern("1").day(arg)
	i2 := taskInstantComplement(gantt, arg)
	if i2.IsFail() {
		return FailableError[any](i2.Error())
	}
	end := i2.Get()
	precise := end.InstantPrecise()
	if end.Attribute() == TaskAttributeEnd {
		precise = precise.Decrement()
	}
	return FailableOk[any](NewDaysAsDates(d1, precise.Day()))
}

// complementIntervals is upstream's ComplementIntervals: "A to B" with two
// dates or two "D+N".
type complementIntervals struct{}

func (complementIntervals) complementUbrex() klimt.UBrexPart {
	b := klimt.UBrexConcat(newDayPattern("1").ubrex(), wordsExactly(wordTo), wordsZeroOrMore(wordThe),
		klimt.UBrexSpaceOneOrMore(), newDayPattern("2").ubrex())
	e := klimt.UBrexConcat(klimt.UBrexLeaf("「dD」+"), klimt.UBrexNamed("ECOUNT1", klimt.UBrexLeaf("〇+〴d")),
		wordsExactly(wordTo), wordsZeroOrMore(wordThe), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("「dD」+"),
		klimt.UBrexNamed("ECOUNT2", klimt.UBrexLeaf("〇+〴d")))
	return klimt.UBrexOr(b, e)
}

func (complementIntervals) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	if d1, ok := newDayPattern("1").day(arg); ok {
		d2, _ := newDayPattern("2").day(arg)
		return FailableOk[any](NewDaysAsDates(d1, d2))
	}
	if _, ok := arg.FindFirstValueByKey("ECOUNT1"); ok {
		day1 := parseInt(firstValue(arg, "ECOUNT1"))
		date1 := gantt.MinTimePoint().AddDays(day1)
		day2 := parseInt(firstValue(arg, "ECOUNT2"))
		date2 := gantt.MinTimePoint().AddDays(day2)
		return FailableOk[any](NewDaysAsDates(date1.Day(), date2.Day()))
	}
	uml.Throw(uml.IllegalStateException, "")
	return Failable[any]{}
}

// complementAnything is upstream's ComplementAnything: the rest of the
// line.
type complementAnything struct{}

func (complementAnything) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed("ANYTHING", klimt.UBrexLeaf("〇*〴."))
}

func (complementAnything) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](firstValue(arg, "ANYTHING"))
}

// complementOpenClose is upstream's ComplementClose and ComplementOpen:
// "closed" or "opened", optionally "for [task]".
type complementOpenClose struct {
	key, word string
}

var (
	complementClose = complementOpenClose{"CLOSED", "close〇?d"}
	complementOpen  = complementOpenClose{"OPEN", "open〇?e〇?d"}
)

func (c complementOpenClose) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed(c.key, klimt.UBrexConcat(klimt.UBrexLeaf(c.word),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf("∙for∙"), taskCodeUbrex("FOO")))))
}

// me returns the task between brackets, "" for none.
func (c complementOpenClose) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	value := firstValue(arg, c.key)
	if x := strings.Index(value, "["); x > 0 {
		y := strings.LastIndex(value, "]")
		return FailableOk[any](value[x+1 : y])
	}
	return FailableOk[any]("")
}

// complementCompleted is upstream's ComplementCompleted: "40% completed".
type complementCompleted struct{}

func (complementCompleted) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexNamed("COMPLEMENT", klimt.UBrexLeaf("〇+〴d")),
		klimt.UBrexLeaf("〇+〴W complete〇?d"))
}

func (complementCompleted) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](Completion{parseInt(firstValue(arg, "COMPLEMENT"))})
}

// complementDayOfWeek is upstream's ComplementDayOfWeek.
type complementDayOfWeek struct{}

func (complementDayOfWeek) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed("COMPLEMENT", DayOfWeekUbrex())
}

func (complementDayOfWeek) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](DayOfWeekFromString(firstValue(arg, "COMPLEMENT")))
}

// complementDeleted is upstream's ComplementDeleted.
type complementDeleted struct{}

func (complementDeleted) complementUbrex() klimt.UBrexPart { return klimt.UBrexLeaf("deleted") }

func (complementDeleted) me(*GanttDiagram, captureLookup) Failable[any] {
	return FailableOk[any](struct{}{})
}
