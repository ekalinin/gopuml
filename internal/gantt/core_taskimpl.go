// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/solver/Solver.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
)

// TaskImpl is upstream's TaskImpl: a task or a milestone. Its start, end and
// plan are cached until its own attributes change, as upstream caches them:
// a later change of the calendar does not reach a task that was already
// solved.
type TaskImpl struct {
	abstractTask
	gantt           *GanttDiagram
	pausedDay       map[jtime.LocalDate]bool
	pausedDayOfWeek [7]bool
	solver          solver
	// resources are the resources in the order they were added, with their
	// percentages.
	resources   []*Resource
	percentages map[string]int
	diamond     bool
	cachedPlan  ngm.PiecewiseConstant
	cachedStart *TimePoint
	cachedEnd   *TimePoint
	completion  int
	note        *klimt.Display
	noteStereo  *klimt.Stereotype
	url         *klimt.URL
	colors      []*CenterBorderColor
}

// NewTaskImpl is new TaskImpl: a task of one day that starts on
// startingDay.
func NewTaskImpl(gantt *GanttDiagram, styleBuilder *klimt.StyleBuilder, code TaskCode, startingDay TimePoint,
	completion int) *TaskImpl {
	t := &TaskImpl{gantt: gantt, completion: completion, pausedDay: map[jtime.LocalDate]bool{},
		percentages: map[string]int{}}
	t.abstractTask = abstractTask{self: t, code: code, styleBuilder: styleBuilder}
	t.SetStart(startingDay)
	t.SetLoad(ngm.LoadOf(gantt.DurationOfDays(1)))
	return t
}

// DefaultPlan is getDefaultPlan: the calendar of the task.
func (t *TaskImpl) DefaultPlan() ngm.PiecewiseConstant {
	return t.gantt.LoadPlanableForTask(t.code.id)
}

func (t *TaskImpl) localPause() ngm.PiecewiseConstant {
	weekPattern := ngm.WeekdayOf(ngm.One)
	for _, day := range jtime.DaysOfWeek {
		if t.pausedDayOfWeek[day.Ordinal()] {
			weekPattern = weekPattern.With(day, ngm.Zero)
		}
	}
	var result ngm.PiecewiseConstant = weekPattern
	if len(t.pausedDay) > 0 {
		closedDays := map[jtime.LocalDate]ngm.Fraction{}
		for date := range t.pausedDay {
			closedDays[date] = ngm.Zero
		}
		result = ngm.CombinerProduct(weekPattern, ngm.SpecificDaysOfMap(ngm.One, closedDays))
	}
	return result
}

// AsPiecewiseConstant is asPiecewiseConstant: the share of work the task
// can do at each instant.
func (t *TaskImpl) AsPiecewiseConstant() ngm.PiecewiseConstant {
	if t.cachedPlan != nil {
		return t.cachedPlan
	}
	if len(t.resources) > 0 {
		t.cachedPlan = ngm.CombinerProduct(t.DefaultPlan(), t.allRessources())
	} else {
		t.cachedPlan = ngm.CombinerProduct(t.DefaultPlan(), t.localPause())
	}
	return t.cachedPlan
}

func (t *TaskImpl) isPaused(instant TimePoint) bool {
	return t.pausedDay[instant.Day()] || t.pausedDayOfWeek[instant.DayOfWeek().Ordinal()]
}

// LoadForResource is loadForResource: the percentage of res on the task at
// instant, 0 outside the task, on a paused day or a day off of res.
func (t *TaskImpl) LoadForResource(res *Resource, instant TimePoint) int {
	if t.IsAssignedTo(res) && instant.CompareTo(t.Start()) >= 0 && instant.CompareTo(endMinusOneDay(t)) <= 0 {
		if t.isPaused(instant) {
			return 0
		}
		if res.IsClosedAt(instant.Day()) {
			return 0
		}
		return t.percentages[res.name]
	}
	return 0
}

// AddPauseDay implements Task.
func (t *TaskImpl) AddPauseDay(pause jtime.LocalDate) {
	t.pausedDay[pause] = true
	t.invalidateCache()
}

// AddPauseDayOfWeek implements Task.
func (t *TaskImpl) AddPauseDayOfWeek(pause jtime.DayOfWeek) {
	t.pausedDayOfWeek[pause.Ordinal()] = true
	t.invalidateCache()
}

func (t *TaskImpl) allRessources() ngm.PiecewiseConstant {
	contributions := make([]ngm.PiecewiseConstant, 0, len(t.resources))
	for _, res := range t.resources {
		percentage := ngm.SpecificDaysOf(ngm.NewFraction(int64(t.percentages[res.name]), 100))
		contributions = append(contributions, ngm.CombinerProduct(res.AsPiecewiseConstant(), percentage))
	}
	return ngm.CombinerProduct(t.localPause(), ngm.CombinerSum(contributions...))
}

// PrettyDisplay is getPrettyDisplay: the text of the task, followed by its
// resources and their percentages when they are not 100.
func (t *TaskImpl) PrettyDisplay() string {
	if len(t.resources) == 0 {
		return t.code.display
	}
	var b strings.Builder
	b.WriteString(t.code.display)
	b.WriteString(" ")
	for i, res := range t.resources {
		b.WriteString("{")
		b.WriteString(res.name)
		if percentage := t.percentages[res.name]; percentage != 100 {
			b.WriteString(":" + strconv.Itoa(percentage) + "%")
		}
		b.WriteString("}")
		if i < len(t.resources)-1 {
			b.WriteString(" ")
		}
	}
	return b.String()
}

// Start implements Moment: the start of the solver, moved to the first day
// the task can work on unless the task is a milestone.
func (t *TaskImpl) Start() TimePoint {
	if t.cachedStart != nil {
		return *t.cachedStart
	}
	allocation := ngm.AllocationOf(t.AsPiecewiseConstant())
	result := t.solver.data(allocation, TaskAttributeStart).(TimePoint)
	if !t.diamond {
		plan := t.AsPiecewiseConstant()
		for ngm.IsZeroOnDay(plan, result.Day()) {
			result = result.Increment()
		}
	}
	t.cachedStart = &result
	return result
}

// End implements Moment.
func (t *TaskImpl) End() TimePoint {
	if t.cachedEnd != nil {
		return *t.cachedEnd
	}
	result := t.solver.data(ngm.AllocationOf(t.AsPiecewiseConstant()), TaskAttributeEnd).(TimePoint)
	t.cachedEnd = &result
	return result
}

// Load implements Task.
func (t *TaskImpl) Load() ngm.Load {
	return t.solver.data(ngm.AllocationOf(t.AsPiecewiseConstant()), TaskAttributeLoad).(ngm.Load)
}

// SetLoad implements Task.
func (t *TaskImpl) SetLoad(load ngm.Load) {
	t.solver.setData(TaskAttributeLoad, load)
	t.invalidateCache()
}

// SetStart implements Task.
func (t *TaskImpl) SetStart(start TimePoint) {
	t.solver.setData(TaskAttributeStart, start)
	t.invalidateCache()
}

// SetEnd implements Task.
func (t *TaskImpl) SetEnd(end TimePoint) {
	t.solver.setData(TaskAttributeEnd, end)
	t.invalidateCache()
}

// SetColors implements Task.
func (t *TaskImpl) SetColors(colors ...*CenterBorderColor) {
	t.colors = colors
}

// AddResource implements Task.
func (t *TaskImpl) AddResource(resource *Resource, percentage int) {
	if _, ok := t.percentages[resource.name]; !ok {
		t.resources = append(t.resources, resource)
	}
	t.percentages[resource.name] = percentage
	t.invalidateCache()
}

func (t *TaskImpl) invalidateCache() {
	t.cachedPlan = nil
	t.cachedStart = nil
	t.cachedEnd = nil
}

// SetDiamond implements Task.
func (t *TaskImpl) SetDiamond(diamond bool) { t.diamond = diamond }

// IsDiamond implements Task.
func (t *TaskImpl) IsDiamond() bool { return t.diamond }

// SetCompletion implements Task.
func (t *TaskImpl) SetCompletion(completion int) { t.completion = completion }

// SetURL implements Task.
func (t *TaskImpl) SetURL(url *klimt.URL) { t.url = url }

// URL is getUrl.
func (t *TaskImpl) URL() *klimt.URL { return t.url }

// Colors is getColors: the colors of the task, between the two colors of
// "is colored for completion" by its completion, nil for none.
func (t *TaskImpl) Colors() *CenterBorderColor {
	switch len(t.colors) {
	case 0:
		return nil
	case 1:
		return t.colors[0]
	}
	return t.colors[0].UnlinearTo(t.colors[1], t.completion)
}

// Completion is getCompletion.
func (t *TaskImpl) Completion() int { return t.completion }

// AllPaused is getAllPaused: the paused dates and the days of the task on a
// paused day of the week, in order.
func (t *TaskImpl) AllPaused() []jtime.LocalDate {
	result := map[jtime.LocalDate]bool{}
	maps.Copy(result, t.pausedDay)
	for _, dayOfWeek := range jtime.DaysOfWeek {
		if !t.pausedDayOfWeek[dayOfWeek.Ordinal()] {
			continue
		}
		start, end := t.Start(), endMinusOneDay(t)
		for current := start; current.CompareTo(end) <= 0; current = current.Increment() {
			if current.DayOfWeek() == dayOfWeek {
				result[current.Day()] = true
			}
		}
	}
	return slices.SortedFunc(maps.Keys(result), jtime.LocalDate.CompareTo)
}

// SetNote implements Task.
func (t *TaskImpl) SetNote(note *klimt.Display, stereotype *klimt.Stereotype) {
	t.note = note
	t.noteStereo = stereotype
}

// Note is getNote.
func (t *TaskImpl) Note() *klimt.Display { return t.note }

// NoteStereotype is getNoteStereotype.
func (t *TaskImpl) NoteStereotype() *klimt.Stereotype { return t.noteStereo }

// IsAssignedTo implements Task.
func (t *TaskImpl) IsAssignedTo(res *Resource) bool {
	_, ok := t.percentages[res.name]
	return ok
}

// solver is upstream's Solver: it keeps the two attributes of a task set
// last (start, end and load) and computes the third one.
type solver struct {
	attributes []TaskAttribute
	values     []any
}

func (s *solver) setData(attribute TaskAttribute, value any) {
	if i := slices.Index(s.attributes, attribute); i >= 0 {
		previous := s.values[i]
		s.attributes = slices.Delete(s.attributes, i, i+1)
		s.values = slices.Delete(s.values, i, i+1)
		// A start never moves before a start that is still kept.
		if attribute == TaskAttributeStart && previous.(TimePoint).CompareTo(value.(TimePoint)) > 0 {
			value = previous
		}
	}
	s.attributes = append(s.attributes, attribute)
	s.values = append(s.values, value)
	if len(s.attributes) > 2 {
		s.attributes = s.attributes[1:]
		s.values = s.values[1:]
	}
}

func (s *solver) get(attribute TaskAttribute) (any, bool) {
	if i := slices.Index(s.attributes, attribute); i >= 0 {
		return s.values[i], true
	}
	return nil, false
}

func (s *solver) data(allocation ngm.Allocation, attribute TaskAttribute) any {
	if result, ok := s.get(attribute); ok {
		return result
	}
	switch attribute {
	case TaskAttributeEnd:
		start, _ := s.get(TaskAttributeStart)
		load, _ := s.get(TaskAttributeLoad)
		task := ngm.TaskWithFixedTotalEffort(allocation, start.(TimePoint).LocalDateTime(), load.(ngm.Load))
		return TimePointOf(task.End())
	case TaskAttributeStart:
		end, _ := s.get(TaskAttributeEnd)
		load, _ := s.get(TaskAttributeLoad)
		task := ngm.TaskWithFixedTotalEffort(allocation, end.(TimePoint).LocalDateTime(), load.(ngm.Load))
		task.SetEnd(end.(TimePoint).LocalDateTime())
		return TimePointOf(task.Start())
	}
	return ngm.LoadOf(jtime.DurationOfDays(1))
}
