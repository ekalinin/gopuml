// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskGroup.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskSeparator.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/Resource.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

func unsupported() { uml.Throw(uml.UnsupportedOperationException, "") }

// TaskGroup is upstream's TaskGroup: the tasks between "group" and "end
// group", from the start of the first one to the end of the last one.
type TaskGroup struct {
	abstractTask
	parent   *TaskGroup
	children []Task
}

// NewTaskGroup is new TaskGroup.
func NewTaskGroup(parent *TaskGroup, styleBuilder *klimt.StyleBuilder, code TaskCode) *TaskGroup {
	g := &TaskGroup{parent: parent}
	g.abstractTask = abstractTask{self: g, code: code, styleBuilder: styleBuilder}
	return g
}

// Start implements Moment: the earliest start of the children, which fails
// with Java's UnsupportedOperationException for an empty group.
func (g *TaskGroup) Start() TimePoint {
	var min *TimePoint
	for _, child := range g.children {
		if min == nil || min.CompareTo(child.Start()) > 0 {
			s := child.Start()
			min = &s
		}
	}
	if min == nil {
		unsupported()
	}
	return *min
}

// End implements Moment: the latest end of the children, which fails with
// Java's UnsupportedOperationException for an empty group.
func (g *TaskGroup) End() TimePoint {
	var max *TimePoint
	for _, child := range g.children {
		if max == nil || max.CompareTo(child.End()) < 0 {
			e := child.End()
			max = &e
		}
	}
	if max == nil {
		unsupported()
	}
	return *max
}

func (g *TaskGroup) SetStart(TimePoint)                        { unsupported() }
func (g *TaskGroup) SetEnd(TimePoint)                          { unsupported() }
func (g *TaskGroup) SetColors(...*CenterBorderColor)           { unsupported() }
func (g *TaskGroup) AddResource(*Resource, int)                { unsupported() }
func (g *TaskGroup) Load() ngm.Load                            { unsupported(); return ngm.Load{} }
func (g *TaskGroup) SetLoad(ngm.Load)                          { unsupported() }
func (g *TaskGroup) SetDiamond(bool)                           { unsupported() }
func (g *TaskGroup) IsDiamond() bool                           { unsupported(); return false }
func (g *TaskGroup) SetCompletion(int)                         { unsupported() }
func (g *TaskGroup) SetURL(*klimt.URL)                         { unsupported() }
func (g *TaskGroup) AddPauseDay(jtime.LocalDate)               { unsupported() }
func (g *TaskGroup) AddPauseDayOfWeek(jtime.DayOfWeek)         { unsupported() }
func (g *TaskGroup) SetNote(*klimt.Display, *klimt.Stereotype) {}
func (g *TaskGroup) IsAssignedTo(*Resource) bool               { return false }

// AddTask is addTask.
func (g *TaskGroup) AddTask(child Task) { g.children = append(g.children, child) }

// Parent is getParent.
func (g *TaskGroup) Parent() *TaskGroup { return g.parent }

// TaskSeparator is upstream's TaskSeparator: a horizontal separator between
// the tasks, with an optional text.
type TaskSeparator struct {
	abstractTask
	// comment is the text, "" for none.
	comment    string
	hasComment bool
}

// NewTaskSeparator is new TaskSeparator: the separator number id, whose code
// is "##id".
func NewTaskSeparator(styleBuilder *klimt.StyleBuilder, comment string, hasComment bool, id int) *TaskSeparator {
	s := &TaskSeparator{comment: comment, hasComment: hasComment}
	s.abstractTask = abstractTask{self: s, code: TaskCodeFromID("##" + strconv.Itoa(id)), styleBuilder: styleBuilder}
	return s
}

// Name is getName: the text, ok false for none.
func (s *TaskSeparator) Name() (string, bool) { return s.comment, s.hasComment }

func (s *TaskSeparator) Start() TimePoint                          { unsupported(); return TimePoint{} }
func (s *TaskSeparator) End() TimePoint                            { unsupported(); return TimePoint{} }
func (s *TaskSeparator) SetStart(TimePoint)                        { unsupported() }
func (s *TaskSeparator) SetEnd(TimePoint)                          { unsupported() }
func (s *TaskSeparator) SetColors(...*CenterBorderColor)           { unsupported() }
func (s *TaskSeparator) AddResource(*Resource, int)                { unsupported() }
func (s *TaskSeparator) Load() ngm.Load                            { unsupported(); return ngm.Load{} }
func (s *TaskSeparator) SetLoad(ngm.Load)                          { unsupported() }
func (s *TaskSeparator) SetDiamond(bool)                           { unsupported() }
func (s *TaskSeparator) IsDiamond() bool                           { unsupported(); return false }
func (s *TaskSeparator) SetCompletion(int)                         { unsupported() }
func (s *TaskSeparator) SetURL(*klimt.URL)                         { unsupported() }
func (s *TaskSeparator) AddPauseDay(jtime.LocalDate)               { unsupported() }
func (s *TaskSeparator) AddPauseDayOfWeek(jtime.DayOfWeek)         { unsupported() }
func (s *TaskSeparator) SetNote(*klimt.Display, *klimt.Stereotype) {}
func (s *TaskSeparator) IsAssignedTo(*Resource) bool               { return false }

// Resource is upstream's Resource: a person or a team that works on tasks,
// with its own days off. Resources are equal by name.
type Resource struct {
	name      string
	draw      ResourceDraw
	openClose *OpenClose
}

// NewResource is new Resource.
func NewResource(name string) *Resource { return &Resource{name: name, openClose: NewOpenClose()} }

// String is toString: the name.
func (r *Resource) String() string { return r.name }

// Name is getName.
func (r *Resource) Name() string { return r.name }

// ResourceDraw is getResourceDraw.
func (r *Resource) ResourceDraw() ResourceDraw { return r.draw }

// SetTaskDraw is setTaskDraw.
func (r *Resource) SetTaskDraw(draw ResourceDraw) { r.draw = draw }

// IsClosedAt is isClosedAt.
func (r *Resource) IsClosedAt(day jtime.LocalDate) bool { return r.openClose.IsClosed(day) }

// AddCloseDay is addCloseDay(LocalDate).
func (r *Resource) AddCloseDay(day jtime.LocalDate) { r.openClose.CloseDay(day) }

// AddForceOnDay is addForceOnDay.
func (r *Resource) AddForceOnDay(day jtime.LocalDate) { r.openClose.OpenDay(day) }

// AddCloseDayOfWeek is addCloseDay(DayOfWeek).
func (r *Resource) AddCloseDayOfWeek(day jtime.DayOfWeek) { r.openClose.CloseDayOfWeek(day) }

// SetOffBeforeDate is setOffBeforeDate.
func (r *Resource) SetOffBeforeDate(day jtime.LocalDate) { r.openClose.SetOffBeforeDate(day) }

// SetOffAfterDate is setOffAfterDate.
func (r *Resource) SetOffAfterDate(day jtime.LocalDate) { r.openClose.SetOffAfterDate(day) }

// LastDayIfAny is getLastDayIfAny: the date after which the resource is
// off, ok false for none.
func (r *Resource) LastDayIfAny() (jtime.LocalDate, bool) { return r.openClose.LastDayIfAny() }

// AsPiecewiseConstant is asPiecewiseConstant: the days the resource works,
// without the calendar of the project.
func (r *Resource) AsPiecewiseConstant() ngm.PiecewiseConstant {
	return r.openClose.AsPiecewiseConstant()
}
