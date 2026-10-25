// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/TaskDrawRegistryData.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/sequencediagram"
)

// TaskDrawRegistryData is upstream's TaskDrawRegistryData: the drawings of
// the tasks in their order, their rows solved from a common origin.
type TaskDrawRegistryData struct {
	origin                   sequencediagram.RealOrigin
	tasks                    []Task
	draws                    map[Task]TaskDraw
	totalHeightWithoutFooter float64
}

// NewTaskDrawRegistryData is new TaskDrawRegistryData.
func NewTaskDrawRegistryData() *TaskDrawRegistryData {
	return &TaskDrawRegistryData{origin: sequencediagram.RealUtilsCreateOrigin(), draws: map[Task]TaskDraw{}}
}

// TotalHeightWithoutFooter is getTotalHeightWithoutFooter.
func (r *TaskDrawRegistryData) TotalHeightWithoutFooter() float64 { return r.totalHeightWithoutFooter }

// TaskDraw is getTaskDraw, nil for none.
func (r *TaskDrawRegistryData) TaskDraw(task Task) TaskDraw {
	if task == nil {
		return nil
	}
	return r.draws[task]
}

// Draws returns the drawings in the order of the tasks.
func (r *TaskDrawRegistryData) Draws() []TaskDraw {
	result := make([]TaskDraw, len(r.tasks))
	for i, task := range r.tasks {
		result[i] = r.draws[task]
	}
	return result
}

func (r *TaskDrawRegistryData) putTaskDraw(task Task, draw TaskDraw) {
	if _, ok := r.draws[task]; !ok {
		r.tasks = append(r.tasks, task)
	}
	r.draws[task] = draw
}

func (r *TaskDrawRegistryData) computeBottomY(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, td := range r.Draws() {
		result = math.Max(result, td.RealY(stringBounder).CurrentValue()+td.HeightMax(stringBounder))
	}
	return result
}

// resolveNoteOverlaps moves a task below the notes of the tasks before it
// that it would overlap.
func (r *TaskDrawRegistryData) resolveNoteOverlaps(stringBounder font.StringBounder) {
	var notes []TaskDraw
	for _, td := range r.Draws() {
		taskPrint := td.FingerPrint(stringBounder)
		fingerPrintNote := td.FingerPrintNote(stringBounder)
		if td.TrueRow() == nil {
			for _, note := range notes {
				otherNote := note.FingerPrintNote(stringBounder)
				if deltaY := otherNote.Overlap(taskPrint); deltaY > 0 {
					bottom := note.RealY(stringBounder).AddAtLeast(note.HeightMax(stringBounder))
					td.RealY(stringBounder).EnsureBiggerThan(bottom)
					r.origin.CompileNow()
				}
			}
		}
		if fingerPrintNote != nil {
			notes = append(notes, td)
		}
	}
}

func (r *TaskDrawRegistryData) createRegularTaskDraw(modelData *GanttModelData, timeBounds *TimeBoundsData,
	timelineStyle *TimelineStyleData, timeScale TimeScale, y Real, task Task, display string) TaskDraw {
	startForDrawing := timeBounds.StartForDrawing(task)
	endForDrawing := timeBounds.EndForDrawing(task)
	oddStart, oddEnd := false, false
	if _, ok := timeBounds.PrintStart(); ok {
		oddStart = TimePointOfStartOfDay(timeBounds.MinDay()).CompareTo(startForDrawing) == 0
		oddEnd = TimePointOfStartOfDay(timeBounds.MaxDay().PlusDays(1)).CompareTo(endForDrawing) == 0
	}
	return NewTaskDrawRegular(timelineStyle.ColorSet(), timeScale, y, display, startForDrawing, endForDrawing,
		oddStart, oddEnd, timelineStyle.skinParam, task, r, modelData.ConstraintsForTask(task), task.StyleBuilder())
}

// BuildTaskAndResourceDraws is buildTaskAndResourceDraws: the drawing of
// each task, one row after the other below the header, then the resources
// below them.
func (r *TaskDrawRegistryData) BuildTaskAndResourceDraws(stringBounder font.StringBounder, timeHeader TimeHeader,
	modelData *GanttModelData, displayConfig *DisplayConfigData, timeBounds *TimeBoundsData,
	timelineStyle *TimelineStyleData) {
	timeScale := timeHeader.TimeScale()
	headerTotalHeight := timeHeader.FullHeaderHeight(stringBounder)
	skinParam := timelineStyle.skinParam
	y := r.origin.AddFixed(headerTotalHeight)
	for _, task := range modelData.Tasks() {
		var draw TaskDraw
		switch t := task.(type) {
		case *TaskSeparator:
			name, ok := t.Name()
			draw = NewTaskDrawSeparator(name, ok, timeScale, y, timeBounds.MinDay(), timeBounds.MaxDay(),
				task.StyleBuilder(), skinParam)
		case *TaskGroup:
			draw = NewTaskDrawGroup(timelineStyle.ColorSet(), timeScale, y, t.Code().Display(),
				timeBounds.StartForDrawing(t), timeBounds.EndForDrawing(t), task, r, task.StyleBuilder(), skinParam)
		default:
			taskImpl := task.(*TaskImpl)
			taskLabel := taskImpl.PrettyDisplay()
			if displayConfig.hideResourceName {
				taskLabel = taskImpl.Code().Display()
			}
			if taskImpl.IsDiamond() {
				draw = NewTaskDrawDiamond(timelineStyle.ColorSet(), timeScale, y, taskLabel,
					timeBounds.StartForDrawing(taskImpl), taskImpl, r, task.StyleBuilder(), skinParam)
			} else {
				draw = r.createRegularTaskDraw(modelData, timeBounds, timelineStyle, timeScale, y, taskImpl, taskLabel)
			}
			draw.SetColorsAndCompletion(taskImpl.Colors(), taskImpl.Completion(), taskImpl.URL(), taskImpl.Note(),
				taskImpl.NoteStereotype())
		}
		if task.Row() == nil {
			y = y.AddAtLeast(draw.FullHeightTask(stringBounder))
		}
		r.putTaskDraw(task, draw)
	}
	r.origin.CompileNow()
	r.resolveNoteOverlaps(stringBounder)
	r.buildResourceDraws(stringBounder, timeScale, modelData, displayConfig, timeBounds, timelineStyle,
		headerTotalHeight)
}

func (r *TaskDrawRegistryData) buildResourceDraws(stringBounder font.StringBounder, timeScale TimeScale,
	modelData *GanttModelData, displayConfig *DisplayConfigData, timeBounds *TimeBoundsData,
	timelineStyle *TimelineStyleData, fullHeaderHeight float64) {
	yy := r.computeBottomY(stringBounder)
	if yy == 0 {
		yy = fullHeaderHeight
	} else if !displayConfig.hideResourceFootbox {
		for _, res := range modelData.Resources() {
			draw := &ResourceDrawNumbers{res, timeScale, yy, TimePointOfStartOfDay(timeBounds.MinDay()),
				TimePointOfEndOfDayMinusOneSecond(timeBounds.MaxDay()), modelData, timelineStyle}
			res.SetTaskDraw(draw)
			yy += draw.Height(stringBounder)
		}
	}
	r.totalHeightWithoutFooter = yy
}
