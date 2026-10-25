// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/Moment.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/MomentImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/Task.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/AbstractTask.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskCode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskAttribute.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/GSide.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/GArrowType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/core/TaskInstant.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttConstraintMode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/CenterBorderColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Moment is upstream's Moment: something with a start and an end, a task
// or the days of a name.
type Moment interface {
	Start() TimePoint
	End() TimePoint
}

// endMinusOneDay is Moment.getEndMinusOneDayTOBEREMOVED: the day before the
// end, at the same time.
func endMinusOneDay(m Moment) TimePoint { return m.End().Decrement() }

// MomentImpl is upstream's MomentImpl.
type MomentImpl struct {
	start, end TimePoint
}

// Start implements Moment.
func (m *MomentImpl) Start() TimePoint { return m.start }

// End implements Moment.
func (m *MomentImpl) End() TimePoint { return m.end }

// Task is upstream's Task: a task, a group or a separator. The methods that
// upstream's TaskGroup and TaskSeparator do not support panic with Java's
// UnsupportedOperationException.
type Task interface {
	Moment
	Code() TaskCode
	Load() ngm.Load
	SetLoad(load ngm.Load)
	SetStart(start TimePoint)
	SetEnd(end TimePoint)
	SetColors(colors ...*CenterBorderColor)
	AddResource(resource *Resource, percentage int)
	SetDiamond(diamond bool)
	IsDiamond() bool
	SetCompletion(completion int)
	SetURL(url *klimt.URL)
	PutInSameRowAs(row Task)
	Row() Task
	AddPauseDay(pause jtime.LocalDate)
	AddPauseDayOfWeek(pause jtime.DayOfWeek)
	SetNote(note *klimt.Display, stereotype *klimt.Stereotype)
	StyleBuilder() *klimt.StyleBuilder
	IsAssignedTo(resource *Resource) bool
	SetDisplay(displayString string)
	// DisplayString is the text of "is displayed as", ok false for none.
	DisplayString() (displayString string, ok bool)
	Stereotype() *klimt.Stereotype
	SetStereotype(stereotype *klimt.Stereotype)
}

// abstractTask is upstream's AbstractTask: the part of the tasks, groups and
// separators they share.
type abstractTask struct {
	self          Task
	code          TaskCode
	styleBuilder  *klimt.StyleBuilder
	row           Task
	displayString *string
	stereotype    *klimt.Stereotype
}

// PutInSameRowAs implements Task: the task is drawn on the row of row.
func (t *abstractTask) PutInSameRowAs(row Task) {
	if t.self != row {
		t.row = row
	}
}

// Row implements Task.
func (t *abstractTask) Row() Task { return t.row }

// Code implements Task.
func (t *abstractTask) Code() TaskCode { return t.code }

// StyleBuilder implements Task: the styles of the time the task was
// created.
func (t *abstractTask) StyleBuilder() *klimt.StyleBuilder { return t.styleBuilder }

// SetDisplay implements Task.
func (t *abstractTask) SetDisplay(displayString string) { t.displayString = &displayString }

// DisplayString implements Task.
func (t *abstractTask) DisplayString() (string, bool) {
	if t.displayString == nil {
		return "", false
	}
	return *t.displayString, true
}

// Stereotype implements Task.
func (t *abstractTask) Stereotype() *klimt.Stereotype { return t.stereotype }

// SetStereotype implements Task.
func (t *abstractTask) SetStereotype(stereotype *klimt.Stereotype) { t.stereotype = stereotype }

// TaskCode is upstream's TaskCode: the id of a task, which identifies it,
// and the text it is drawn with.
type TaskCode struct {
	id, display string
}

// TaskCodeFromID is TaskCode.fromId.
func TaskCodeFromID(id string) TaskCode { return TaskCode{id, id} }

// TaskCodeFromIDAndDisplay is TaskCode.fromIdAndDisplay; an empty id or
// display stands for null.
func TaskCodeFromIDAndDisplay(id string, hasID bool, display string, hasDisplay bool) TaskCode {
	if !hasID {
		id = display
	}
	if !hasDisplay {
		return TaskCodeFromID(id)
	}
	return TaskCode{id, display}
}

// Display is getDisplay.
func (c TaskCode) Display() string { return c.display }

// ID is getId.
func (c TaskCode) ID() string { return c.id }

// TaskAttribute is upstream's TaskAttribute.
type TaskAttribute int

// The attributes of a task.
const (
	TaskAttributeStart TaskAttribute = iota
	TaskAttributeEnd
	TaskAttributeLoad
)

var taskAttributeNames = [...]string{"START", "END", "LOAD"}

// String is the name of the enum constant.
func (a TaskAttribute) String() string { return taskAttributeNames[a] }

// TaskAttributeFromString is TaskAttribute.fromString: the attribute named
// by value in any case, which fails with Java's IllegalArgumentException
// otherwise.
func TaskAttributeFromString(value string) TaskAttribute {
	upper := strings.ToUpper(value)
	for i, name := range taskAttributeNames {
		if name == upper {
			return TaskAttribute(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException,
		"No enum constant net.sourceforge.plantuml.gantt.core.TaskAttribute."+upper)
	return 0
}

// GSide is upstream's GSide: the side of a task an arrow touches.
type GSide int

// The sides of a task.
const (
	GSideLeft GSide = iota
	GSideRight
	GSideTopLeft
	GSideTopRight
	GSideBottomLeft
	GSideBottomRight
)

// IsBottom is isBottom.
func (s GSide) IsBottom() bool { return s == GSideBottomLeft || s == GSideBottomRight }

// IsTop is isTop.
func (s GSide) IsTop() bool { return s == GSideTopLeft || s == GSideTopRight }

// ReverseBottomTop is reverseBottomTop, which fails with Java's
// IllegalStateException for a side that is not at the bottom.
func (s GSide) ReverseBottomTop() GSide {
	switch s {
	case GSideBottomLeft:
		return GSideTopLeft
	case GSideBottomRight:
		return GSideTopRight
	}
	uml.Throw(uml.IllegalStateException, "")
	return 0
}

// GArrowType is upstream's GArrowType.
type GArrowType int

// The directions of an arrow from the point of view of a task.
const (
	GArrowIncoming GArrowType = iota
	GArrowOutgoing
)

// GanttConstraintMode is upstream's GanttConstraintMode: whether a delay
// counts the closed days.
type GanttConstraintMode int

// The modes of a delay.
const (
	IgnoreCalendar GanttConstraintMode = iota
	DoNotCountCloseDay
)

// TaskInstant is upstream's TaskInstant: the start or the end of a moment,
// with a delay in days.
type TaskInstant struct {
	task      Moment
	attribute TaskAttribute
	delta     int
	mode      GanttConstraintMode
	calendar  ngm.PiecewiseConstant
}

// NewTaskInstant is new TaskInstant(task, attribute), which fails with
// Java's IllegalArgumentException for the load.
func NewTaskInstant(task Moment, attribute TaskAttribute) *TaskInstant {
	return newTaskInstant(task, attribute, 0, IgnoreCalendar, nil)
}

func newTaskInstant(task Moment, attribute TaskAttribute, delta int, mode GanttConstraintMode,
	calendar ngm.PiecewiseConstant) *TaskInstant {
	if attribute != TaskAttributeStart && attribute != TaskAttributeEnd {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &TaskInstant{task, attribute, delta, mode, calendar}
}

// WithDelta is withDelta: the same instant with a delay of newDelta days,
// counted in calendar when mode skips the closed days.
func (ti *TaskInstant) WithDelta(newDelta int, mode GanttConstraintMode, calendar ngm.PiecewiseConstant) *TaskInstant {
	return newTaskInstant(ti.task, ti.attribute, newDelta, mode, calendar)
}

func (ti *TaskInstant) manageDelta(value TimePoint) TimePoint {
	if ti.delta > 0 {
		for added := 0; added < ti.delta; {
			value = value.Increment()
			if ti.mode == DoNotCountCloseDay && ngm.IsZeroOnDay(ti.calendar, value.Day()) {
				continue
			}
			added++
		}
	}
	for i := 0; i < -ti.delta; i++ {
		value = value.Decrement()
	}
	return value
}

// InstantPrecise is getInstantPrecise: the start or the end of the moment,
// after the delay.
func (ti *TaskInstant) InstantPrecise() TimePoint {
	if ti.attribute == TaskAttributeStart {
		return ti.manageDelta(ti.task.Start())
	}
	return ti.manageDelta(ti.task.End())
}

// Moment is getMoment.
func (ti *TaskInstant) Moment() Moment { return ti.task }

// IsTask is isTask: whether the moment is a task rather than named days.
func (ti *TaskInstant) IsTask() bool {
	_, ok := ti.task.(Task)
	return ok
}

// Attribute is getAttribute.
func (ti *TaskInstant) Attribute() TaskAttribute { return ti.attribute }

// SameRowAs is sameRowAs: whether the two tasks are drawn on the same row
// because one was put on the row of the other.
func (ti *TaskInstant) SameRowAs(dest *TaskInstant) bool {
	if ti.IsTask() && dest.IsTask() {
		t1, t2 := ti.task.(Task), dest.task.(Task)
		if t1 == t2.Row() || t2 == t1.Row() {
			return true
		}
	}
	return false
}

// CenterBorderColor is upstream's CenterBorderColor: the fill and border
// colors of a task, and the style of a link.
type CenterBorderColor struct {
	center, border color.HColor
	// style is the style of a link ("dotted", ...), "" for none.
	style string
}

// NewCenterBorderColor is new CenterBorderColor(center, border, style); a
// nil border draws the border in the center color.
func NewCenterBorderColor(center, border color.HColor, style string) *CenterBorderColor {
	return &CenterBorderColor{center, border, style}
}

// IsOk is isOk: whether there is a center color.
func (c *CenterBorderColor) IsOk() bool { return c.center != nil }

// Center is getCenter.
func (c *CenterBorderColor) Center() color.HColor { return c.center }

// Border returns the border color, nil for none.
func (c *CenterBorderColor) Border() color.HColor { return c.border }

// Style is getStyle.
func (c *CenterBorderColor) Style() string { return c.style }

// UnlinearTo is unlinearTo: the colors completion percent of the way to
// other.
func (c *CenterBorderColor) UnlinearTo(other *CenterBorderColor, completion int) *CenterBorderColor {
	return &CenterBorderColor{color.Unlinear(c.center, other.center, completion),
		color.Unlinear(c.border, other.border, completion), c.style}
}
