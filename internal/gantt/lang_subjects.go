// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectGantt.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectToday.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectTask.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectResource.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectDaysAsDates.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectDayOfWeek.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectDayAsDate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectSeparator.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/SubjectWorkingHours.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/stereo/StereotypePattern.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// noStartingDate is the error of the verbs that need the start of the
// project.
const noStartingDate = "No starting date for the project"

// subjectGantt is upstream's SubjectGantt: "Project starts ...".
type subjectGantt struct{}

func (subjectGantt) subjectUbrex() klimt.UBrexPart {
	return klimt.UBrexOr(klimt.UBrexLeaf("project"), klimt.UBrexLeaf("gantt"))
}

func (subjectGantt) me(gantt *GanttDiagram, _ *klimt.UBrexMatcher) Failable[any] {
	return FailableOk[any](gantt)
}

func (subjectGantt) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbStarts, ignore(wordsZeroOrMore(wordOn, wordFor, wordThe, wordAt)),
			complementDate{dateOnlyAbsolute}, func(gantt *GanttDiagram, _, complement any) command.Result {
				return gantt.UpdateStartingPoint(complement.(jtime.LocalDate))
			}),
	}
}

// subjectToday is upstream's SubjectToday: "today is ...".
type subjectToday struct{}

func (subjectToday) subjectUbrex() klimt.UBrexPart { return klimt.UBrexLeaf("today") }

func (subjectToday) me(*GanttDiagram, *klimt.UBrexMatcher) Failable[any] {
	return FailableOk[any](Today{})
}

func (subjectToday) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbIsColored, nil, complementInColors{}, func(gantt *GanttDiagram, _, complement any) command.Result {
			gantt.SetTodayColors(complement.(*CenterBorderColor))
			return command.Ok()
		}),
		phrase(verbIs, nil, complementDate{dateAny}, func(gantt *GanttDiagram, _, complement any) command.Result {
			return gantt.SetToday(complement.(jtime.LocalDate))
		}),
	}
}

// stereotypeOptional is StereotypePattern.uoptional: an optional stereotype
// between spaces.
func stereotypeOptional(param string) klimt.UBrexPart {
	return klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexNamed(param, klimt.UBrexLeaf("<< 〇+「〤<>」>>")), klimt.UBrexSpaceZeroOrMore()))
}

// subjectTask is upstream's SubjectTask: "[name]", "then [name] as [id]
// on {resource}", or "it".
type subjectTask struct{}

func (subjectTask) subjectUbrex() klimt.UBrexPart {
	return klimt.UBrexOr(
		klimt.UBrexNamed("IT", klimt.UBrexLeaf(wordIt)),
		klimt.UBrexConcat(
			klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexNamed("THEN", klimt.UBrexLeaf(wordThen)),
				klimt.UBrexSpaceOneOrMore())),
			taskCodeUbrex("SUBJECT"),
			stereotypeOptional("STEREOTYPE"),
			klimt.UBrexOptional(klimt.UBrexConcat(wordsExactly(wordAs), klimt.UBrexSpaceOneOrMore(),
				taskCodeUbrex("SHORTNAME"))),
			klimt.UBrexOptional(klimt.UBrexConcat(wordsExactly(wordOn), klimt.UBrexSpaceOneOrMore(),
				klimt.UBrexNamed("RESOURCE", klimt.UBrexOneOrMore(klimt.UBrexConcat(klimt.UBrexLeaf("{〇+「〤{}」}"),
					klimt.UBrexSpaceZeroOrMore())))))))
}

func (subjectTask) me(gantt *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	var result Task
	if _, ok := arg.FindFirstValueByKey("IT"); ok {
		result = gantt.It()
		if result == nil {
			return FailableError[any]("Not sure what are you refering to?")
		}
	} else {
		subject, hasSubject := arg.FindFirstValueByKey("SUBJECT")
		shortName, hasShortName := arg.FindFirstValueByKey("SHORTNAME")
		_, then := arg.FindFirstValueByKey("THEN")
		stereotype, hasStereotype := arg.FindFirstValueByKey("STEREOTYPE")
		code := TaskCodeFromIDAndDisplay(shortName, hasShortName, subject, hasSubject)
		result = gantt.OrCreateTask(code, then)
		if hasStereotype {
			result.SetStereotype(klimt.StereotypeBuild(stereotype))
		}
		gantt.SetIt(result)
	}
	if resource, ok := arg.FindFirstValueByKey("RESOURCE"); ok {
		for _, part := range strings.FieldsFunc(resource, func(r rune) bool { return r == '{' || r == '}' }) {
			part = javaTrim(part)
			if part != "" && !gantt.AffectResource(result, part) {
				return FailableError[any]("Bad argument for resource")
			}
		}
	}
	return FailableOk[any](result)
}

// addInstantConstraint adds the constraint of a start or end given by
// another task, with the color and style of "with ... link" when colors is
// not nil.
func addInstantConstraint(gantt *GanttDiagram, when *TaskInstant, task Task, attribute TaskAttribute,
	colors *CenterBorderColor) {
	var c color.HColor
	if colors != nil {
		c = colors.Center()
	}
	link := NewGanttConstraint(gantt.CurrentStyleBuilder(), when, NewTaskInstant(task, attribute), c)
	if colors != nil {
		link.ApplyStyle(colors.Style(), colors.Style() != "")
	}
	gantt.AddContraint(link)
}

func (subjectTask) verbPhrases() []*verbPhraseAction {
	theOnAt := func() *klimt.UBrexPart { return ignore(wordsZeroOrMore(wordThe, wordOn, wordAt)) }
	pauses := ignore(wordsZeroOrMore(wordThe, wordOn, wordAt, wordFrom))
	addPauses := func(gantt *GanttDiagram, subject, complement any) command.Result {
		for _, day := range complement.(DaysAsDates).Dates() {
			subject.(Task).AddPauseDay(day)
		}
		return command.Ok()
	}
	return []*verbPhraseAction{
		phrase(verbRequires, ignore(wordsZeroOrMore(wordOn, wordFor, wordThe, wordAt)), complementDuration{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetLoad(complement.(ngm.Load))
				return command.Ok()
			}),
		phrase(verbStarts, theOnAt(), complementTimePoint{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				if gantt.MinDay() == TimePointEpoch() {
					return command.Error(noStartingDate)
				}
				subject.(Task).SetStart(complement.(TimePoint))
				return command.Ok()
			}),
		phrase(verbStarts, nil, pairOfSomething{complementTaskInstant{}, complementWithColorLink{}},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				pair := complement.([2]any)
				when := pair[0].(*TaskInstant)
				task := subject.(Task)
				task.SetStart(when.InstantPrecise())
				if when.IsTask() {
					addInstantConstraint(gantt, when, task, TaskAttributeStart, pair[1].(*CenterBorderColor))
				}
				return command.Ok()
			}),
		phrase(verbStarts, nil, complementTaskInstant{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				when := complement.(*TaskInstant)
				task := subject.(Task)
				task.SetStart(when.InstantPrecise())
				if when.IsTask() {
					addInstantConstraint(gantt, when, task, TaskAttributeStart, nil)
				}
				return command.Ok()
			}),
		phrase(verbStarts, theOnAt(), complementDate{dateOnlyRelative},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetStart(TimePointOfStartOfDay(complement.(jtime.LocalDate)))
				return command.Ok()
			}),
		phrase(verbStarts, theOnAt(), complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				if gantt.MinDay() == TimePointEpoch() {
					return command.Error(noStartingDate)
				}
				subject.(Task).SetStart(TimePointOfStartOfDay(complement.(jtime.LocalDate)))
				return command.Ok()
			}),
		phrase(verbIsColored, nil, complementInColors{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetColors(complement.(*CenterBorderColor))
				return command.Ok()
			}),
		phrase(verbIsColored, ignore(wordsExactly2(wordFor, wordCompletion)), complementInColorsFromTo{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetColors(complement.([]*CenterBorderColor)...)
				return command.Ok()
			}),
		phrase(verbHappens, theOnAt(), complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				task := subject.(Task)
				task.SetLoad(ngm.LoadOf(jtime.DurationOfDays(1)))
				task.SetStart(TimePointOfStartOfDay(complement.(jtime.LocalDate)))
				task.SetDiamond(true)
				return command.Ok()
			}),
		phrase(verbHappens, theOnAt(), complementTaskInstant{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				task := subject.(Task)
				task.SetLoad(ngm.LoadOf(jtime.DurationOfDays(1)))
				task.SetDiamond(true)
				when := complement.(*TaskInstant)
				if when.Attribute() == TaskAttributeEnd {
					task.SetStart(when.InstantPrecise().Decrement())
				} else {
					task.SetStart(when.InstantPrecise())
				}
				return command.Ok()
			}),
		phrase(verbOccurs, nil, complementFromTo{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				task := subject.(Task)
				names := complement.(twoNames)
				from := gantt.ExistingTask(names.name1)
				if from == nil {
					return command.Error("No such " + names.name1 + " task")
				}
				to := gantt.ExistingTask(names.name2)
				if to == nil {
					return command.Error("No such " + names.name2 + " task")
				}
				task.SetStart(from.Start())
				task.SetEnd(to.End())
				gantt.AddContraint(NewGanttConstraint(gantt.CurrentStyleBuilder(),
					NewTaskInstant(from, TaskAttributeStart), NewTaskInstant(task, TaskAttributeStart), nil))
				gantt.AddContraint(NewGanttConstraint(gantt.CurrentStyleBuilder(),
					NewTaskInstant(to, TaskAttributeEnd), NewTaskInstant(task, TaskAttributeEnd), nil))
				return command.Ok()
			}),
		phrase(verbEnds, theOnAt(), complementTimePoint{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				if gantt.MinDay() == TimePointEpoch() {
					return command.Error(noStartingDate)
				}
				subject.(Task).SetEnd(complement.(TimePoint))
				return command.Ok()
			}),
		phrase(verbEnds, nil, complementTaskInstant{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				when := complement.(*TaskInstant)
				task := subject.(Task)
				task.SetEnd(when.InstantPrecise())
				gantt.AddContraint(NewGanttConstraint(gantt.CurrentStyleBuilder(), when,
					NewTaskInstant(task, TaskAttributeEnd), nil))
				return command.Ok()
			}),
		phrase(verbEnds, theOnAt(), complementDate{dateOnlyRelative},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetEnd(TimePointOfStartOfDay(complement.(jtime.LocalDate)).Increment())
				return command.Ok()
			}),
		phrase(verbEnds, theOnAt(), complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				if gantt.MinDay() == TimePointEpoch() {
					return command.Error(noStartingDate)
				}
				subject.(Task).SetEnd(TimePointOfStartOfDay(complement.(jtime.LocalDate)).Increment())
				return command.Ok()
			}),
		phrase(verbDisplayOnSameRowAs, nil, complementNamed{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				task2 := gantt.ExistingTask(complement.(string))
				if task2 == nil {
					// upstream concatenates the missing task
					return command.Error("No such task null")
				}
				subject.(Task).PutInSameRowAs(task2)
				return command.Ok()
			}),
		phrase(verbIs, nil, complementDeleted{}, func(gantt *GanttDiagram, subject, _ any) command.Result {
			return gantt.DeleteTask(subject.(Task))
		}),
		phrase(verbIs, nil, complementCompleted{}, func(gantt *GanttDiagram, subject, complement any) command.Result {
			subject.(Task).SetCompletion(complement.(Completion).completion)
			return command.Ok()
		}),
		phrase(verbPauses, pauses, complementIntervals{}, addPauses),
		phrase(verbPauses, pauses, complementIntervalsSmart{}, addPauses),
		phrase(verbPauses, pauses, complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).AddPauseDay(complement.(jtime.LocalDate))
				return command.Ok()
			}),
		phrase(verbPauses, pauses, complementDayOfWeek{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).AddPauseDayOfWeek(complement.(jtime.DayOfWeek))
				return command.Ok()
			}),
		phrase(verbLinksTo, nil, complementURL{}, func(gantt *GanttDiagram, subject, complement any) command.Result {
			subject.(Task).SetURL(complement.(*klimt.URL))
			return command.Ok()
		}),
		phrase(verbIsDisplayedAs, nil, complementAnything{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(Task).SetDisplay(complement.(string))
				return command.Ok()
			}),
	}
}

// subjectResource is upstream's SubjectResource: "{name}", or "she", "he"
// and "they".
type subjectResource struct{}

func (subjectResource) subjectUbrex() klimt.UBrexPart {
	return klimt.UBrexOr(klimt.UBrexNamed("THEY", klimt.UBrexLeaf("【she┇he┇they】")),
		klimt.UBrexConcat(klimt.UBrexLeaf("{"), klimt.UBrexNamed("RESOURCE", klimt.UBrexLeaf("〇+「〤{}」")),
			klimt.UBrexLeaf("}")))
}

func (subjectResource) me(gantt *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	if _, ok := arg.FindFirstValueByKey("THEY"); ok {
		they := gantt.They()
		if they == nil {
			return FailableError[any]("Not sure who are you refering to?")
		}
		return FailableOk[any](they)
	}
	result := gantt.Resource(firstValue(arg, "RESOURCE"))
	gantt.SetThey(result)
	return FailableOk[any](result)
}

func (subjectResource) verbPhrases() []*verbPhraseAction {
	words := func() *klimt.UBrexPart {
		return ignore(wordsZeroOrMore(wordFrom, wordOn, wordFor, wordThe, wordAt))
	}
	return []*verbPhraseAction{
		phrase(verbWorksOn, nil, complementTask{}, func(gantt *GanttDiagram, subject, complement any) command.Result {
			complement.(Task).AddResource(subject.(*Resource), 100)
			return command.Ok()
		}),
		phrase(verbIsOff, ignore(klimt.UBrexConcat(klimt.UBrexLeaf(wordBefore), wordsZeroOrMore(wordThe))),
			complementDate{dateAny}, func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(*Resource).SetOffBeforeDate(complement.(jtime.LocalDate))
				return command.Ok()
			}),
		phrase(verbIsOff, ignore(klimt.UBrexConcat(klimt.UBrexLeaf(wordAfter), wordsZeroOrMore(wordThe))),
			complementDate{dateAny}, func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(*Resource).SetOffAfterDate(complement.(jtime.LocalDate))
				return command.Ok()
			}),
		phrase(verbIsOff, words(), complementIntervals{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				for _, when := range complement.(DaysAsDates).Dates() {
					subject.(*Resource).AddCloseDay(when)
				}
				return command.Ok()
			}),
		phrase(verbIsOff, words(), complementDayOfWeek{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(*Resource).AddCloseDayOfWeek(complement.(jtime.DayOfWeek))
				return command.Ok()
			}),
		phrase(verbIsOff, words(), complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(*Resource).AddCloseDay(complement.(jtime.LocalDate))
				return command.Ok()
			}),
		phrase(verbIsOn, words(), complementIntervals{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				for _, when := range complement.(DaysAsDates).Dates() {
					subject.(*Resource).AddForceOnDay(when)
				}
				return command.Ok()
			}),
		phrase(verbIsOn, words(), complementDate{dateAny},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				subject.(*Resource).AddForceOnDay(complement.(jtime.LocalDate))
				return command.Ok()
			}),
	}
}

// subjectDaysAsDates is upstream's SubjectDaysAsDates: "A to B", "D+N to
// D+M", "A and N days", "then N days".
type subjectDaysAsDates struct{}

func (subjectDaysAsDates) subjectUbrex() klimt.UBrexPart {
	b := klimt.UBrexConcat(ubrexYYYYMMDD("BYEAR1", "BMONTH1", "BDAY1"), wordsExactly(wordTo),
		klimt.UBrexSpaceOneOrMore(), ubrexYYYYMMDD("BYEAR2", "BMONTH2", "BDAY2"))
	e := klimt.UBrexConcat(klimt.UBrexLeaf("「dD」+"), klimt.UBrexNamed("ECOUNT1", klimt.UBrexLeaf("〇+〴d")),
		wordsExactly(wordTo), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("「dD」+"),
		klimt.UBrexNamed("ECOUNT2", klimt.UBrexLeaf("〇+〴d")))
	and := klimt.UBrexConcat(ubrexYYYYMMDD("BYEAR3", "BMONTH3", "BDAY3"), wordsExactly(wordAnd),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexNamed("COUNT_AND", klimt.UBrexLeaf("〇+〴d")),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("day〇?s"))
	then := klimt.UBrexConcat(klimt.UBrexLeaf("then"), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("COUNT_THEN", klimt.UBrexLeaf("〇+〴d")), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexLeaf("day〇?s"))
	return klimt.UBrexOr(b, e, and, then)
}

func (subjectDaysAsDates) me(gantt *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	if countAnd, ok := arg.FindFirstValueByKey("COUNT_AND"); ok {
		date3 := daysAsDatesDate(gantt, arg, "3")
		return FailableOk[any](NewDaysAsDatesCount(gantt, date3.Day(), parseInt(countAnd)))
	}
	if countThen, ok := arg.FindFirstValueByKey("COUNT_THEN"); ok {
		date3 := gantt.ThenDate()
		return FailableOk[any](NewDaysAsDatesCount(gantt, date3.Day(), parseInt(countThen)))
	}
	date1 := daysAsDatesDate(gantt, arg, "1")
	date2 := daysAsDatesDate(gantt, arg, "2")
	return FailableOk[any](NewDaysAsDates(date1.Day(), date2.Day()))
}

func daysAsDatesDate(gantt *GanttDiagram, arg *klimt.UBrexMatcher, suffix string) TimePoint {
	if d, ok := arg.FindFirstValueByKey("BDAY" + suffix); ok {
		day := parseInt(d)
		month := parseInt(firstValue(arg, "BMONTH"+suffix))
		year := parseInt(firstValue(arg, "BYEAR"+suffix))
		return TimePointOfStartOfDayYMD(year, month, day)
	}
	if e, ok := arg.FindFirstValueByKey("ECOUNT" + suffix); ok {
		return gantt.MinTimePoint().AddDays(parseInt(e))
	}
	uml.Throw(uml.IllegalStateException, "")
	return TimePoint{}
}

func (subjectDaysAsDates) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbIsOrAre, nil, complementClose, func(gantt *GanttDiagram, subject, complement any) command.Result {
			for _, d := range subject.(DaysAsDates).Dates() {
				gantt.CloseDayAsDate(d, complement.(string))
			}
			return command.Ok()
		}),
		phrase(verbIsOrAre, nil, complementOpen, func(gantt *GanttDiagram, subject, complement any) command.Result {
			for _, d := range subject.(DaysAsDates).Dates() {
				gantt.OpenDayAsDate(d, complement.(string))
			}
			return command.Ok()
		}),
		phrase(verbIsOrAre, nil, complementInColors2{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				c := complement.(*CenterBorderColor).Center()
				for _, d := range subject.(DaysAsDates).Dates() {
					gantt.ColorDay(d, c)
				}
				return command.Ok()
			}),
		phrase(verbIsOrAreNamed, nil, complementNamed{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				for _, d := range subject.(DaysAsDates).Dates() {
					gantt.NameDay(d, complement.(string))
				}
				return command.Ok()
			}),
	}
}

// subjectDayOfWeek is upstream's SubjectDayOfWeek: "saturday are closed".
type subjectDayOfWeek struct{}

func (subjectDayOfWeek) subjectUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed("SUBJECT", DayOfWeekUbrex())
}

func (subjectDayOfWeek) me(_ *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	return FailableOk[any](DayOfWeekFromString(firstValue(arg, "SUBJECT")))
}

func (subjectDayOfWeek) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbAre, nil, complementOpen, func(gantt *GanttDiagram, subject, complement any) command.Result {
			gantt.OpenDayOfWeek(subject.(jtime.DayOfWeek), complement.(string))
			return command.Ok()
		}),
		phrase(verbAre, nil, complementClose, func(gantt *GanttDiagram, subject, complement any) command.Result {
			gantt.CloseDayOfWeek(subject.(jtime.DayOfWeek), complement.(string))
			return command.Ok()
		}),
		phrase(verbIsOrAre, nil, complementInColors2{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				gantt.ColorDayOfWeek(subject.(jtime.DayOfWeek), complement.(*CenterBorderColor).Center())
				return command.Ok()
			}),
	}
}

// subjectDayAsDate is upstream's SubjectDayAsDate: "2020-01-06 is closed",
// "D+5", "T-1", "E+2".
type subjectDayAsDate struct{}

func (subjectDayAsDate) subjectUbrex() klimt.UBrexPart {
	e := klimt.UBrexConcat(klimt.UBrexNamed("ETYPE", klimt.UBrexLeaf("「dDtTeE」")),
		klimt.UBrexNamed("EOPERATION", klimt.UBrexLeaf("「-+」")), klimt.UBrexNamed("ECOUNT", klimt.UBrexLeaf("〇+〴d")))
	return klimt.UBrexOr(ubrexYYYYMMDD("BYEAR", "BMONTH", "BDAY"), e)
}

func (subjectDayAsDate) me(gantt *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	if d, ok := arg.FindFirstValueByKey("BDAY"); ok {
		day := parseInt(d)
		month := parseInt(firstValue(arg, "BMONTH"))
		year := parseInt(firstValue(arg, "BYEAR"))
		return FailableOk[any](jtime.DateOf(year, month, day))
	}
	if _, ok := arg.FindFirstValueByKey("ECOUNT"); ok {
		typ := strings.ToUpper(firstValue(arg, "ETYPE"))
		day := int32(parseInt(firstValue(arg, "ECOUNT")))
		if firstValue(arg, "EOPERATION") == "-" {
			day = -day
		}
		switch typ {
		case "D":
			return FailableOk[any](gantt.MinDay().PlusDays(int64(day)))
		case "T":
			return FailableOk[any](gantt.Today().PlusDays(int64(day)))
		case "E":
			return FailableOk[any](gantt.MaxDay().PlusDays(int64(day)))
		}
	}
	uml.Throw(uml.IllegalStateException, "")
	return Failable[any]{}
}

func (subjectDayAsDate) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbIsOrAre, nil, complementOpen, func(gantt *GanttDiagram, subject, complement any) command.Result {
			gantt.OpenDayAsDate(subject.(jtime.LocalDate), complement.(string))
			return command.Ok()
		}),
		phrase(verbIsOrAre, nil, complementClose, func(gantt *GanttDiagram, subject, complement any) command.Result {
			gantt.CloseDayAsDate(subject.(jtime.LocalDate), complement.(string))
			return command.Ok()
		}),
		phrase(verbIsOrAre, nil, complementInColors2{},
			func(gantt *GanttDiagram, subject, complement any) command.Result {
				gantt.ColorDay(subject.(jtime.LocalDate), complement.(*CenterBorderColor).Center())
				return command.Ok()
			}),
	}
}

// subjectSeparator is upstream's SubjectSeparator: "separator just before
// ...".
type subjectSeparator struct{}

func (subjectSeparator) subjectUbrex() klimt.UBrexPart { return klimt.UBrexLeaf("separator") }

func (subjectSeparator) me(gantt *GanttDiagram, _ *klimt.UBrexMatcher) Failable[any] {
	return FailableOk[any](gantt)
}

func (subjectSeparator) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbJust, ignore(klimt.UBrexLeaf(wordBefore)), complementDate{dateAny},
			func(gantt *GanttDiagram, _, complement any) command.Result {
				gantt.AddVerticalSeparatorBefore(complement.(jtime.LocalDate))
				return command.Ok()
			}),
		phrase(verbJust, ignore(klimt.UBrexLeaf(wordAfter)), complementDate{dateAny},
			func(gantt *GanttDiagram, _, complement any) command.Result {
				gantt.AddVerticalSeparatorBefore(complement.(jtime.LocalDate).PlusDays(1))
				return command.Ok()
			}),
		phrase(verbJust, nil, complementTaskInstant{}, func(gantt *GanttDiagram, _, complement any) command.Result {
			gantt.AddVerticalSeparatorBefore(complement.(*TaskInstant).InstantPrecise().Day())
			return command.Ok()
		}),
	}
}

// subjectWorkingHours is upstream's SubjectWorkingHours: "from 8:00 to
// 12:00 are working hours".
type subjectWorkingHours struct{}

func (subjectWorkingHours) subjectUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("from"), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("START", klimt.UBrexLeaf("〇+〴d:〇+〴d")), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("to"),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexNamed("END", klimt.UBrexLeaf("〇+〴d:〇+〴d")))
}

func (subjectWorkingHours) me(_ *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any] {
	return FailableOk[any](NewTimeRange(firstValue(arg, "START"), firstValue(arg, "END")))
}

func (subjectWorkingHours) verbPhrases() []*verbPhraseAction {
	return []*verbPhraseAction{
		phrase(verbAre, nil, complementWorkingHours{}, func(gantt *GanttDiagram, subject, _ any) command.Result {
			return gantt.AddWorkingTimeRange(subject.(*TimeRange))
		}),
	}
}
