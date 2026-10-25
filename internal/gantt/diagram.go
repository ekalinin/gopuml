// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttDiagram.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/Failable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/Completion.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/Today.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"regexp"
	"slices"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// GanttDiagram is upstream's GanttDiagram: a parsed gantt diagram.
type GanttDiagram struct {
	uml.Base

	modelData     *GanttModelData
	timeBounds    *TimeBoundsData
	scaleConfig   *TimeScaleConfigData
	weekConfig    *WeekConfigData
	dayCalendar   *DayCalendarData
	displayConfig *DisplayConfigData

	today             *TimePoint
	defaultCompletion int

	// it is the last task used as a subject, they the last resource.
	it           Task
	they         *Resource
	currentGroup *TaskGroup

	// displayedColumn are the columns of the table of tasks.
	displayedColumn map[GanttTaskTableColumn]bool
}

// resourceAssignmentPattern is upstream's RESOURCE_ASSIGNMENT_PATTERN, used
// with find.
var resourceAssignmentPattern = regexp.MustCompile(`([^:]+)(:(\d+))?`)

// NewGanttDiagram returns an empty gantt diagram for src; md5map holds the
// base64 images of the source.
func NewGanttDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *GanttDiagram {
	return &GanttDiagram{
		Base:              uml.NewBase(ctx, src, core.Gantt, md5map),
		modelData:         NewGanttModelData(),
		timeBounds:        NewTimeBoundsData(),
		scaleConfig:       &TimeScaleConfigData{factorScale: 1.0},
		weekConfig:        NewWeekConfigData(),
		dayCalendar:       NewDayCalendarData(),
		displayConfig:     &DisplayConfigData{showFootbox: true},
		defaultCompletion: 100,
		displayedColumn:   map[GanttTaskTableColumn]bool{ColumnStart: true, ColumnEnd: true, ColumnDuration: true},
	}
}

// Description implements core.Diagram.
func (d *GanttDiagram) Description() string { return "(Gantt)" }

// DefaultMargins is getDefaultMargins: none.
func (d *GanttDiagram) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return klimt.ClockwiseNone()
}

// ChangeLanguage is changeLanguage.
func (d *GanttDiagram) ChangeLanguage(lang string) command.Result {
	d.weekConfig.locale = NewLocale(lang)
	return command.Ok()
}

// SetWeekNumberStrategy is setWeekNumberStrategy.
func (d *GanttDiagram) SetWeekNumberStrategy(firstDayOfWeek jtime.DayOfWeek, minimalDaysInFirstWeek int) {
	d.weekConfig.weekNumberStrategy = WeekNumberStrategy{firstDayOfWeek, minimalDaysInFirstWeek}
}

// SetPrintScale is setPrintScale.
func (d *GanttDiagram) SetPrintScale(printScale PrintScale) { d.scaleConfig.printScale = printScale }

// SetFactorScale is setFactorScale.
func (d *GanttDiagram) SetFactorScale(factorScale float64) { d.scaleConfig.factorScale = factorScale }

// CheckFinalError implements command.Diagram: upstream's checkFinalError,
// the last day of the calendar and the limit of 20 years.
func (d *GanttDiagram) CheckFinalError() string {
	d.initMinMax()
	if d.timeBounds.DurationInDays() > 365*20 {
		return "Gantt diagrams cannot last more than 20 years"
	}
	return ""
}

func (d *GanttDiagram) initMinMax() { d.timeBounds.InitMinMax(d.modelData, d.dayCalendar) }

// CloseDayOfWeek is closeDayOfWeek: the day is closed for every task, task
// is ignored as upstream ignores it.
func (d *GanttDiagram) CloseDayOfWeek(day jtime.DayOfWeek, task string) {
	d.dayCalendar.OpenClose().CloseDayOfWeek(day)
}

// OpenDayOfWeek is openDayOfWeek: for every task, or for one task.
func (d *GanttDiagram) OpenDayOfWeek(day jtime.DayOfWeek, task string) {
	if task == "" {
		d.dayCalendar.OpenClose().OpenDayOfWeek(day)
	} else {
		d.dayCalendar.OpenCloseForTask(task).OpenDayOfWeek(day)
	}
}

// CloseDayAsDate is closeDayAsDate: for every task, or for one task.
func (d *GanttDiagram) CloseDayAsDate(day jtime.LocalDate, task string) {
	if task == "" {
		d.dayCalendar.OpenClose().CloseDay(day)
	} else {
		d.dayCalendar.OpenCloseForTask(task).CloseDay(day)
	}
}

// OpenDayAsDate is openDayAsDate: for every task, or for one task.
func (d *GanttDiagram) OpenDayAsDate(day jtime.LocalDate, task string) {
	if task == "" {
		d.dayCalendar.OpenClose().OpenDay(day)
	} else {
		d.dayCalendar.OpenCloseForTask(task).OpenDay(day)
	}
}

// ThenDate is getThenDate: the start of "then N days", the last colored or
// named day, or the start of the project.
func (d *GanttDiagram) ThenDate() TimePoint {
	result := TimePointOfStartOfDay(d.timeBounds.MinDay())
	for _, day := range d.dayCalendar.ColorDays() {
		if day.CompareTo(result) > 0 {
			result = day
		}
	}
	for day := range d.dayCalendar.NameDays() {
		if day.CompareTo(result) > 0 {
			result = day
		}
	}
	return result
}

// ExistingTask is getExistingTask, nil for none.
func (d *GanttDiagram) ExistingTask(id string) Task { return d.modelData.Task(TaskCodeFromID(id)) }

// ForceTaskOrder is forceTaskOrder: task2 starts at the end of task1, with
// an arrow.
func (d *GanttDiagram) ForceTaskOrder(task1, task2 Task) *GanttConstraint {
	end1 := NewTaskInstant(task1, TaskAttributeEnd)
	task2.SetStart(end1.InstantPrecise())
	result := NewGanttConstraint(d.CurrentStyleBuilder(), end1, NewTaskInstant(task2, TaskAttributeStart), nil)
	d.AddContraint(result)
	return result
}

// OrCreateTask is getOrCreateTask: the task of code, created when it does
// not exist, starting at the end of the last task when linkedToPrevious.
func (d *GanttDiagram) OrCreateTask(code TaskCode, linkedToPrevious bool) Task {
	result := d.modelData.Task(code)
	if result == nil {
		var previous Task
		if linkedToPrevious {
			previous = d.lastCreatedTask()
		}
		result = NewTaskImpl(d, d.CurrentStyleBuilder(), code, TimePointOfStartOfDay(d.timeBounds.MinDay()),
			d.defaultCompletion)
		if d.currentGroup != nil {
			d.currentGroup.AddTask(result)
		}
		d.modelData.PutTask(code, result)
		if previous != nil {
			d.ForceTaskOrder(previous, result)
		}
	}
	return result
}

// LoadPlanableForTask is getLoadPlanableForTask: the calendar of the
// project with the days of the task on top.
func (d *GanttDiagram) LoadPlanableForTask(taskID string) ngm.PiecewiseConstant {
	return d.dayCalendar.OpenClose().MutateMe(d.dayCalendar.OpenCloseForTask(taskID)).AsPiecewiseConstant()
}

func (d *GanttDiagram) lastCreatedTask() Task {
	all := d.modelData.Tasks()
	for i := len(all) - 1; i >= 0; i-- {
		if t, ok := all[i].(*TaskImpl); ok {
			return t
		}
	}
	return nil
}

// AddSeparator is addSeparator; hasComment is false for "--".
func (d *GanttDiagram) AddSeparator(comment string, hasComment bool) {
	separator := NewTaskSeparator(d.CurrentStyleBuilder(), comment, hasComment, len(d.modelData.taskIDs))
	d.modelData.PutTask(separator.Code(), separator)
}

// AddGroup is addGroup.
func (d *GanttDiagram) AddGroup(code TaskCode) command.Result {
	group := NewTaskGroup(d.currentGroup, d.CurrentStyleBuilder(), code)
	if d.currentGroup != nil {
		d.currentGroup.AddTask(group)
	}
	d.currentGroup = group
	d.modelData.PutTask(group.Code(), group)
	return command.Ok()
}

// EndGroup is endGroup.
func (d *GanttDiagram) EndGroup() command.Result {
	if d.currentGroup == nil {
		return command.Error("No group to be closed")
	}
	d.currentGroup = d.currentGroup.Parent()
	return command.Ok()
}

// AddContraint is addContraint.
func (d *GanttDiagram) AddContraint(constraint *GanttConstraint) {
	d.modelData.AddConstraint(constraint)
}

// UpdateStartingPoint is updateStartingPoint: the start of the project,
// before any task.
func (d *GanttDiagram) UpdateStartingPoint(start jtime.LocalDate) command.Result {
	if len(d.modelData.taskIDs) > 0 {
		return command.Error("Starting point must be set before task definition")
	}
	d.timeBounds.SetMinDay(start)
	return command.Ok()
}

// MinDay is getMinDay.
func (d *GanttDiagram) MinDay() jtime.LocalDate { return d.timeBounds.MinDay() }

// MaxDay is getMaxDay: the last day of the tasks read so far.
func (d *GanttDiagram) MaxDay() jtime.LocalDate {
	d.initMinMax()
	return d.timeBounds.MaxDay()
}

// MinTimePoint is getMinTimePoint.
func (d *GanttDiagram) MinTimePoint() TimePoint { return TimePointOfStartOfDay(d.timeBounds.MinDay()) }

// MaxTimePoint is getMaxTimePoint.
func (d *GanttDiagram) MaxTimePoint() TimePoint {
	d.initMinMax()
	return TimePointOfStartOfDay(d.timeBounds.MaxDay())
}

// DaysInWeek is daysInWeek: the open days of a week, when the line is
// read.
func (d *GanttDiagram) DaysInWeek() int { return d.dayCalendar.OpenClose().DaysInWeek() }

// DaysInMonth is daysInMonth: 30.
func (d *GanttDiagram) DaysInMonth() int { return 30 }

// IsOpen is isOpen.
func (d *GanttDiagram) IsOpen(day jtime.LocalDate) bool { return d.dayCalendar.IsOpen(day) }

// AffectResource is affectResource: the resource of description ("name" or
// "name:percentage") added to the task. It returns false for a percentage
// of 0.
func (d *GanttDiagram) AffectResource(result Task, description string) bool {
	m := resourceAssignmentPattern.FindStringSubmatch(description)
	if m == nil {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	resource := d.Resource(m[1])
	percentage := 100
	if m[3] != "" {
		percentage = uml.JavaParseInt(m[3])
	}
	if percentage == 0 {
		return false
	}
	result.AddResource(resource, percentage)
	return true
}

// Resource is getResource: the resource of the name, created the first
// time.
func (d *GanttDiagram) Resource(resourceName string) *Resource {
	resource := d.modelData.Resource(resourceName)
	if resource == nil {
		resource = NewResource(resourceName)
	}
	d.modelData.PutResource(resourceName, resource)
	return resource
}

// ExistingMoment is getExistingMoment: the task of id, or the days named
// id without regard to case, nil for none.
func (d *GanttDiagram) ExistingMoment(id string) Moment {
	if task := d.ExistingTask(id); task != nil {
		return task
	}
	var start, end *TimePoint
	for day, name := range d.dayCalendar.NameDays() {
		if !uml.EqualsIgnoreCase(name, id) {
			continue
		}
		if start == nil || start.CompareTo(day) > 0 {
			s := day
			start = &s
		}
		if end == nil || end.CompareTo(day) < 0 {
			e := day
			end = &e
		}
	}
	if start == nil {
		return nil
	}
	return &MomentImpl{*start, end.Increment()}
}

// ColorDay is colorDay(LocalDate, HColor).
func (d *GanttDiagram) ColorDay(day jtime.LocalDate, c color.HColor) {
	d.dayCalendar.PutColorDay(TimePointOfStartOfDay(day), c)
}

// ColorDayOfWeek is colorDay(DayOfWeek, HColor).
func (d *GanttDiagram) ColorDayOfWeek(day jtime.DayOfWeek, c color.HColor) {
	d.dayCalendar.PutColorDayOfWeek(day, c)
}

// NameDay is nameDay.
func (d *GanttDiagram) NameDay(day jtime.LocalDate, name string) {
	d.dayCalendar.PutNameDay(TimePointOfStartOfDay(day), name)
}

// Today is getToday: the day of "today is", or the current day.
func (d *GanttDiagram) Today() jtime.LocalDate {
	if d.today == nil {
		t := todayAtMidnight()
		d.today = &t
	}
	return d.today.Day()
}

// SetTodayColors is setTodayColors: the color of the day of "today is", or
// of the current day.
func (d *GanttDiagram) SetTodayColors(colors *CenterBorderColor) {
	if d.today == nil {
		t := todayAtMidnight()
		d.today = &t
	}
	d.dayCalendar.PutColorDayToday(*d.today, colors.Center())
}

// SetToday is setToday.
func (d *GanttDiagram) SetToday(date jtime.LocalDate) command.Result {
	t := TimePointOfStartOfDay(date)
	d.today = &t
	return command.Ok()
}

// DeleteTask is deleteTask: the task is drawn white with a black border.
func (d *GanttDiagram) DeleteTask(task Task) command.Result {
	task.SetColors(NewCenterBorderColor(color.White, color.Black, ""))
	return command.Ok()
}

// SetPrintInterval is setPrintInterval.
func (d *GanttDiagram) SetPrintInterval(start, end jtime.LocalDate) {
	d.timeBounds.SetPrintStart(start)
	d.timeBounds.SetPrintEnd(end)
}

// AddNote is addNote: the note of the last task, group or separator.
func (d *GanttDiagram) AddNote(note *klimt.Display, stereotype *klimt.Stereotype) command.Result {
	tasks := d.modelData.Tasks()
	if len(tasks) == 0 {
		return command.Error("No task defined")
	}
	tasks[len(tasks)-1].SetNote(note, stereotype)
	return command.Ok()
}

// SetShowFootbox is setShowFootbox.
func (d *GanttDiagram) SetShowFootbox(footbox bool) { d.displayConfig.showFootbox = footbox }

// SetWeeklyHeaderStrategy is setWeeklyHeaderStrategy.
func (d *GanttDiagram) SetWeeklyHeaderStrategy(strategy WeeklyHeaderStrategy, weekStartingNumber int) {
	d.weekConfig.weeklyHeaderStrategy = strategy
	d.weekConfig.weekStartingNumber = weekStartingNumber
}

// HideResourceName is hideResourceName.
func (d *GanttDiagram) HideResourceName() command.Result {
	d.displayConfig.hideResourceName = true
	return command.Ok()
}

// HideResourceFootbox is hideResourceFootbox.
func (d *GanttDiagram) HideResourceFootbox() command.Result {
	d.displayConfig.hideResourceFootbox = true
	return command.Ok()
}

// AddVerticalSeparatorBefore is addVerticalSeparatorBefore.
func (d *GanttDiagram) AddVerticalSeparatorBefore(day jtime.LocalDate) {
	d.dayCalendar.AddSeparatorBefore(day)
}

// SetTaskDefaultCompletion is setTaskDefaultCompletion.
func (d *GanttDiagram) SetTaskDefaultCompletion(defaultCompletion int) {
	d.defaultCompletion = defaultCompletion
}

// SetIt is setIt.
func (d *GanttDiagram) SetIt(result Task) { d.it = result }

// It is getIt, nil for none.
func (d *GanttDiagram) It() Task { return d.it }

// They is getThey, nil for none.
func (d *GanttDiagram) They() *Resource { return d.they }

// SetThey is setThey.
func (d *GanttDiagram) SetThey(they *Resource) { d.they = they }

// SetHideClosed is setHideClosed.
func (d *GanttDiagram) SetHideClosed(hideClosed bool) { d.scaleConfig.hideClosed = hideClosed }

// DefaultPlan is getDefaultPlan.
func (d *GanttDiagram) DefaultPlan() ngm.PiecewiseConstant { return d.dayCalendar.DefaultPlan() }

// ColorSet is getIHtmlColorSet.
func (d *GanttDiagram) ColorSet() *color.HColorSet { return color.Instance() }

// AddWorkingTimeRange is addWorkingTimeRange.
func (d *GanttDiagram) AddWorkingTimeRange(timeRange *TimeRange) command.Result {
	d.dayCalendar.AddWorkingTimeRange(timeRange)
	return command.Ok()
}

// DurationOfDays is durationOfDays: totalDays days of work.
func (d *GanttDiagram) DurationOfDays(totalDays int) jtime.Duration {
	return d.dayCalendar.DurationOfDays(totalDays)
}

// DurationOfDaysAndHours is durationOfDaysAndHours.
func (d *GanttDiagram) DurationOfDaysAndHours(totalDays, totalHours int) jtime.Duration {
	return d.DurationOfDays(totalDays).Plus(jtime.DurationOfHours(int64(totalHours)))
}

// DurationOfDaysHoursMinutesSeconds is durationOfDaysHoursMinutesSeconds.
func (d *GanttDiagram) DurationOfDaysHoursMinutesSeconds(totalDays, totalHours, totalMinutes,
	totalSeconds int) jtime.Duration {
	return d.DurationOfDays(totalDays).PlusHours(int64(totalHours)).PlusMinutes(int64(totalMinutes)).
		PlusSeconds(int64(totalSeconds))
}

// isRelativeMode is isRelativeMode: whether the project has no start.
func (d *GanttDiagram) isRelativeMode() bool { return d.timeBounds.MinDay() == TimePointEpoch() }

// relativeDayNum is relativeDayNum: the number of the day of point, 1 for
// the start of the project.
func (d *GanttDiagram) relativeDayNum(point TimePoint) int {
	minAbs := TimePointOfStartOfDay(d.timeBounds.MinDay()).AbsoluteDayNum()
	return point.AbsoluteDayNum() - minAbs + 1
}

// HideOrShowColumn is hideOrShowColumn.
func (d *GanttDiagram) HideOrShowColumn(show bool, column GanttTaskTableColumn, ok bool) command.Result {
	if ok {
		d.displayedColumn[column] = show
	}
	return command.Ok()
}

// DisplayedColumn is getDisplayedColumn: the columns of the table, in
// upstream's order.
func (d *GanttDiagram) DisplayedColumn() []GanttTaskTableColumn {
	var result []GanttTaskTableColumn
	for _, c := range []GanttTaskTableColumn{ColumnTask, ColumnStart, ColumnEnd, ColumnDuration} {
		if d.displayedColumn[c] {
			result = append(result, c)
		}
	}
	return slices.Clip(result)
}

// Failable is upstream's Failable: a value or an error message.
type Failable[O any] struct {
	data  O
	error string
	fail  bool
}

// FailableOk is Failable.ok.
func FailableOk[O any](data O) Failable[O] { return Failable[O]{data: data} }

// FailableError is Failable.error.
func FailableError[O any](error string) Failable[O] { return Failable[O]{error: error, fail: true} }

// Get is get.
func (f Failable[O]) Get() O { return f.data }

// IsFail is isFail.
func (f Failable[O]) IsFail() bool { return f.fail }

// Error is getError.
func (f Failable[O]) Error() string { return f.error }

// Completion is upstream's Completion: a percentage of completion.
type Completion struct {
	completion int
}

// Today is upstream's Today: the subject "today".
type Today struct{}
