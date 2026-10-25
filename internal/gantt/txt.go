// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"fmt"
	"io"

	"github.com/ekalinin/gopuml/internal/asciiverse"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// exportTxt is upstream's GanttDiagram.exportTxt: the table of the tasks
// with their start, end and duration, drawn in ASCII for TXT and UTXT
// alike, without the title and the other annotations. A panic is returned
// as an error and nothing is written.
func (d *GanttDiagram) exportTxt(w io.Writer) (data export.ImageData, err error) {
	var plan *asciiverse.InfinitePlan
	func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = fmt.Errorf("drawing crashed: %w", e)
				} else {
					err = fmt.Errorf("drawing crashed: %v", r)
				}
			}
		}()
		d.initMinMax()
		locale := d.weekConfig.locale
		var drawableTasks []Task
		for _, task := range d.modelData.Tasks() {
			if d.isDrawable(task) {
				drawableTasks = append(drawableTasks, task)
			}
		}
		// One header row plus one row per drawable task; four columns.
		table := asciiverse.NewATable(len(drawableTasks)+1, 4).WithLeftPadding(1).WithRightPadding(1)
		table.SetCellContent(0, 0, klimt.DisplayCreate(i18nTask(locale)))
		table.SetCellContent(0, 1, klimt.DisplayCreate(i18nStart(locale)))
		table.SetCellContent(0, 2, klimt.DisplayCreate(i18nEnd(locale)))
		table.SetCellContent(0, 3, klimt.DisplayCreate(i18nDuration(locale)))
		for i, task := range drawableTasks {
			row := i + 1
			table.SetCellContent(row, 0, klimt.DisplayCreate(task.Code().Display()))
			table.SetCellContent(row, 1, klimt.DisplayCreate(d.startOf(task, locale)))
			table.SetCellContent(row, 2, klimt.DisplayCreate(d.endOf(task, locale)))
			table.SetCellContent(row, 3, klimt.DisplayCreate(d.durationOf(task, locale)))
		}
		plan = asciiverse.NewInfinitePlan(font.FormatATXT)
		table.AsciiDraw(plan)
	}()
	if err != nil {
		return nil, err
	}
	if err := plan.ExportTxt(w); err != nil {
		return nil, err
	}
	return export.ImageDataOK(), nil
}

// isDrawable is isDrawable: the tasks and milestones within "print
// between".
func (d *GanttDiagram) isDrawable(task Task) bool {
	if _, ok := task.(*TaskImpl); !ok {
		return false
	}
	return !d.timeBounds.IsHidden(task)
}

// startOf is startOf: the day of the start, or "Day N" without a project
// start.
func (d *GanttDiagram) startOf(task Task, locale Locale) string {
	if d.isRelativeMode() {
		return i18nDayNumber(locale, d.relativeDayNum(task.Start()))
	}
	return task.Start().StringShort(locale)
}

// endOf is endOf: the day before the end.
func (d *GanttDiagram) endOf(task Task, locale Locale) string {
	if d.isRelativeMode() {
		return i18nDayNumber(locale, d.relativeDayNum(endMinusOneDay(task)))
	}
	return endMinusOneDay(task).StringShort(locale)
}

// durationOf is durationOf: the time from the start to the end, closed
// days included.
func (d *GanttDiagram) durationOf(task Task, locale Locale) string {
	return i18nDurationHumanReadable(locale, jtime.Between(task.Start().LocalDateTime(), task.End().LocalDateTime()))
}
