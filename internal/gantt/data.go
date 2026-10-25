// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/GanttModelData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/TimeBoundsData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/WeekConfigData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/DisplayConfigData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/TimeScaleConfigData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/WeeklyHeaderStrategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/uml"
)

// GanttModelData is upstream's GanttModelData: the tasks and the resources
// in the order they were created, and the constraints.
type GanttModelData struct {
	constraints []*GanttConstraint
	taskIDs     []string
	tasks       map[string]Task
	resourceIDs []string
	resources   map[string]*Resource
}

// NewGanttModelData is new GanttModelData.
func NewGanttModelData() *GanttModelData {
	return &GanttModelData{tasks: map[string]Task{}, resources: map[string]*Resource{}}
}

// Tasks is getTasks: the tasks, groups and separators in the order they
// were created.
func (m *GanttModelData) Tasks() []Task {
	result := make([]Task, len(m.taskIDs))
	for i, id := range m.taskIDs {
		result[i] = m.tasks[id]
	}
	return result
}

// Resources is getResources: the resources in the order they were
// created.
func (m *GanttModelData) Resources() []*Resource {
	result := make([]*Resource, len(m.resourceIDs))
	for i, name := range m.resourceIDs {
		result[i] = m.resources[name]
	}
	return result
}

// Constraints is getConstraints.
func (m *GanttModelData) Constraints() []*GanttConstraint { return m.constraints }

// AddConstraint is addConstraint.
func (m *GanttModelData) AddConstraint(constraint *GanttConstraint) {
	m.constraints = append(m.constraints, constraint)
}

// PutTask is putTask: a task replaces the task of the same id at its place.
func (m *GanttModelData) PutTask(code TaskCode, task Task) {
	if _, ok := m.tasks[code.id]; !ok {
		m.taskIDs = append(m.taskIDs, code.id)
	}
	m.tasks[code.id] = task
}

// Task is getTask, nil for none.
func (m *GanttModelData) Task(code TaskCode) Task { return m.tasks[code.id] }

// PutResource is putResource.
func (m *GanttModelData) PutResource(name string, resource *Resource) {
	if _, ok := m.resources[name]; !ok {
		m.resourceIDs = append(m.resourceIDs, name)
	}
	m.resources[name] = resource
}

// Resource is getResource, nil for none.
func (m *GanttModelData) Resource(name string) *Resource { return m.resources[name] }

// LoadForResource is getLoadForResource: the load of res on the day i, the
// sum over the tasks. As upstream, a group fails with Java's
// ClassCastException.
func (m *GanttModelData) LoadForResource(res *Resource, i TimePoint) int {
	result := 0
	for _, task := range m.Tasks() {
		if _, ok := task.(*TaskSeparator); ok {
			continue
		}
		task2, ok := task.(*TaskImpl)
		if !ok {
			uml.Throw(uml.ClassCastException, "class net.sourceforge.plantuml.gantt.core.TaskGroup cannot be cast "+
				"to class net.sourceforge.plantuml.gantt.core.TaskImpl (net.sourceforge.plantuml.gantt.core.TaskGroup "+
				"and net.sourceforge.plantuml.gantt.core.TaskImpl are in unnamed module of loader 'app')")
		}
		result += task2.LoadForResource(res, i)
	}
	return result
}

// ConstraintsForTask is getConstraintsForTask.
func (m *GanttModelData) ConstraintsForTask(task Task) []*GanttConstraint {
	var result []*GanttConstraint
	for _, constraint := range m.constraints {
		if constraint.IsOn(task) {
			result = append(result, constraint)
		}
	}
	return result
}

// TimeBoundsData is upstream's TimeBoundsData: the first and the last day of
// the calendar, and the days of "print between".
type TimeBoundsData struct {
	minDay     jtime.LocalDate
	maxDay     *jtime.LocalDate
	printStart *jtime.LocalDate
	printEnd   *jtime.LocalDate
}

// NewTimeBoundsData is new TimeBoundsData: a project that starts on
// 1970-01-01.
func NewTimeBoundsData() *TimeBoundsData { return &TimeBoundsData{minDay: TimePointEpoch()} }

// MinDay is getMinDay.
func (b *TimeBoundsData) MinDay() jtime.LocalDate { return b.minDay }

// MaxDay is getMaxDay, which fails with Java's NullPointerException when
// it is not set.
func (b *TimeBoundsData) MaxDay() jtime.LocalDate {
	if b.maxDay == nil {
		uml.ThrowNPE("")
	}
	return *b.maxDay
}

// PrintStart is getPrintStart, ok false when there is none.
func (b *TimeBoundsData) PrintStart() (jtime.LocalDate, bool) {
	if b.printStart == nil {
		return jtime.LocalDate{}, false
	}
	return *b.printStart, true
}

// PrintEnd is getPrintEnd.
func (b *TimeBoundsData) PrintEnd() jtime.LocalDate { return *b.printEnd }

// SetMinDay is setMinDay.
func (b *TimeBoundsData) SetMinDay(minDay jtime.LocalDate) { b.minDay = minDay }

// SetMaxDay is setMaxDay.
func (b *TimeBoundsData) SetMaxDay(maxDay jtime.LocalDate) { b.maxDay = &maxDay }

// SetPrintStart is setPrintStart.
func (b *TimeBoundsData) SetPrintStart(printStart jtime.LocalDate) { b.printStart = &printStart }

// SetPrintEnd is setPrintEnd.
func (b *TimeBoundsData) SetPrintEnd(printEnd jtime.LocalDate) { b.printEnd = &printEnd }

// StartForDrawing is getStartForDrawing: the start of the task, not before
// the first day of "print between".
func (b *TimeBoundsData) StartForDrawing(tmp Task) TimePoint {
	if b.printStart == nil {
		return tmp.Start()
	}
	return MaxTimePoint(TimePointOfStartOfDay(b.minDay), tmp.Start())
}

// EndForDrawing is getEndForDrawing: the end of the task, not after the
// last day of "print between".
func (b *TimeBoundsData) EndForDrawing(tmp Task) TimePoint {
	if b.printStart == nil {
		return tmp.End()
	}
	return MinTimePoint(TimePointOfStartOfDay(b.MaxDay().PlusDays(1)), tmp.End())
}

// IsHidden is isHidden: whether the task is outside "print between".
func (b *TimeBoundsData) IsHidden(task Task) bool {
	if _, ok := task.(*TaskSeparator); b.printStart == nil || ok {
		return false
	}
	if endMinusOneDay(task).CompareTo(TimePointOfStartOfDay(b.minDay)) < 0 {
		return true
	}
	return task.Start().CompareTo(TimePointOfEndOfDayMinusOneSecond(b.MaxDay())) > 0
}

// InitMinMax is initMinMax: the last day of the calendar, the last day of
// the tasks or of the colored and named days.
func (b *TimeBoundsData) InitMinMax(modelData *GanttModelData, dayCalendar *DayCalendarData) {
	tasks := modelData.Tasks()
	if len(tasks) == 0 {
		b.SetMaxDay(b.minDay)
	} else {
		b.maxDay = nil
		for _, task := range tasks {
			switch task.(type) {
			case *TaskSeparator, *TaskGroup:
				continue
			}
			tmp := task.End().MinusOneSecond()
			if b.maxDay == nil || b.maxDay.CompareTo(tmp.Day()) < 0 {
				b.SetMaxDay(tmp.Day())
			}
		}
	}
	for _, d := range dayCalendar.ColorDays() {
		if b.compareToMaxDay(d.Day()) > 0 {
			b.SetMaxDay(d.Day())
		}
	}
	for d := range dayCalendar.NameDays() {
		if b.compareToMaxDay(d.Day()) > 0 {
			b.SetMaxDay(d.Day())
		}
	}
}

// compareToMaxDay is day.compareTo(getMaxDay()), with the
// NullPointerException of upstream when there is no maximum (only groups
// and separators).
func (b *TimeBoundsData) compareToMaxDay(day jtime.LocalDate) int {
	if b.maxDay == nil {
		uml.ThrowNPE(`Cannot invoke "java.time.chrono.ChronoLocalDate.toEpochDay()" because "other" is null`)
	}
	return day.CompareTo(*b.maxDay)
}

// DurationInDays is durationInDays, which fails with Java's
// NullPointerException when there is no last day.
func (b *TimeBoundsData) DurationInDays() int64 {
	if b.maxDay == nil {
		uml.ThrowNPE("temporal")
	}
	return jtime.DaysBetween(b.minDay, *b.maxDay)
}

// WeeklyHeaderStrategy is upstream's WeeklyHeaderStrategy: what the weekly
// calendar writes for each week.
type WeeklyHeaderStrategy int

// The labels of the weeks.
const (
	WeekOfYear WeeklyHeaderStrategy = iota
	DayOfMonth
	FromN
)

// WeekConfigData is upstream's WeekConfigData: the language and the
// numbering of the weeks.
type WeekConfigData struct {
	locale               Locale
	weekNumberStrategy   WeekNumberStrategy
	weeklyHeaderStrategy WeeklyHeaderStrategy
	weekStartingNumber   int
}

// NewWeekConfigData is new WeekConfigData: English, weeks of ISO-8601
// (Monday, 4 days).
func NewWeekConfigData() *WeekConfigData {
	return &WeekConfigData{locale: LocaleEnglish, weekNumberStrategy: WeekNumberStrategy{jtime.Monday, 4}}
}

// TimeScaleConfigData is upstream's TimeScaleConfigData: the print scale,
// its zoom and hide closed.
type TimeScaleConfigData struct {
	printScale  PrintScale
	factorScale float64
	hideClosed  bool
}

// EffectiveScale is getEffectiveScale: the default scale of the print scale
// times the zoom.
func (c *TimeScaleConfigData) EffectiveScale() float64 {
	return c.printScale.DefaultScale() * c.factorScale
}

// DisplayConfigData is upstream's DisplayConfigData: the footbox and the
// names and footbox of the resources.
type DisplayConfigData struct {
	showFootbox         bool
	hideResourceName    bool
	hideResourceFootbox bool
}
