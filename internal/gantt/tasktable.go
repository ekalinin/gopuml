// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttTaskTableColumn.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttTaskTable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"fmt"
	"math"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// GanttTaskTableColumn is upstream's GanttTaskTableColumn: a column of the
// table of tasks.
type GanttTaskTableColumn int

// The columns of the table of tasks, in upstream's order.
const (
	ColumnTask GanttTaskTableColumn = iota
	ColumnStart
	ColumnEnd
	ColumnDuration
)

// header is header: the title of the column, in the language.
func (c GanttTaskTableColumn) header(locale Locale) string {
	switch c {
	case ColumnTask:
		return i18nTask(locale)
	case ColumnStart:
		return i18nStart(locale)
	case ColumnEnd:
		return i18nEnd(locale)
	}
	return i18nDuration(locale)
}

// valueOf is valueOf: the cell of a task.
func (c GanttTaskTableColumn) valueOf(task Task, context *columnContext) string {
	switch c {
	case ColumnTask:
		return task.Code().Display()
	case ColumnStart:
		return context.formatDayWithTime(task.Start())
	case ColumnEnd:
		return context.formatEnd(task.End())
	}
	return i18nDurationHumanReadable(context.locale,
		jtime.Between(task.Start().LocalDateTime(), task.End().LocalDateTime()))
}

// columnContext is upstream's GanttTaskTableColumn.Context: how the days of
// the table are written.
type columnContext struct {
	locale       Locale
	relativeMode bool
	minDay       jtime.LocalDate
}

func (c *columnContext) formatDay(point TimePoint) string {
	if c.relativeMode {
		return i18nDayNumber(c.locale, c.relativeDayNum(point))
	}
	return point.StringShort(c.locale)
}

// formatDayWithTime is formatDay, followed by the time of day when it is
// not midnight, the seconds when they are not zero.
func (c *columnContext) formatDayWithTime(point TimePoint) string {
	day := c.formatDay(point)
	dt := point.LocalDateTime()
	hour, minute, second := dt.Hour(), dt.Minute(), dt.Second()
	if hour == 0 && minute == 0 && second == 0 {
		return day
	}
	if second == 0 {
		return day + " " + fmt.Sprintf("%02d:%02d", hour, minute)
	}
	return day + " " + fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// formatEnd writes an end at midnight as the day before, and an end with a
// time of day as it is.
func (c *columnContext) formatEnd(end TimePoint) string {
	dt := end.LocalDateTime()
	if dt.Hour() == 0 && dt.Minute() == 0 && dt.Second() == 0 {
		return c.formatDay(end.MinusOneSecond())
	}
	return c.formatDayWithTime(end)
}

func (c *columnContext) relativeDayNum(point TimePoint) int {
	minAbs := TimePointOfStartOfDay(c.minDay).AbsoluteDayNum()
	return point.AbsoluteDayNum() - minAbs + 1
}

// cellPadding is the space on each side of a cell of the table.
const cellPadding = 5

// GanttTaskTable is upstream's GanttTaskTable: the table of the tasks, left
// of the calendar.
type GanttTaskTable struct {
	modelData       *GanttModelData
	timeBounds      *TimeBoundsData
	drawRegistry    *TaskDrawRegistryData
	timelineStyle   *TimelineStyleData
	locale          Locale
	headerHeight    float64
	context         *columnContext
	displayedColumn []GanttTaskTableColumn
	// columnEdges are the left of each column, then the right of the last
	// one.
	columnEdges []float64
}

// NewGanttTaskTable is new GanttTaskTable.
func NewGanttTaskTable(modelData *GanttModelData, timeBounds *TimeBoundsData, drawRegistry *TaskDrawRegistryData,
	timelineStyle *TimelineStyleData, headerHeight float64, locale Locale, stringBounder font.StringBounder,
	displayedColumn []GanttTaskTableColumn) *GanttTaskTable {
	t := &GanttTaskTable{modelData: modelData, timeBounds: timeBounds, drawRegistry: drawRegistry,
		timelineStyle: timelineStyle, headerHeight: headerHeight, locale: locale, displayedColumn: displayedColumn}
	t.context = &columnContext{locale, timeBounds.MinDay() == TimePointEpoch(), timeBounds.MinDay()}
	t.columnEdges = make([]float64, len(displayedColumn)+1)
	for i, column := range displayedColumn {
		t.columnEdges[i+1] = t.columnEdges[i] + t.columnWidth(stringBounder, column)
	}
	return t
}

func (t *GanttTaskTable) columnWidth(stringBounder font.StringBounder, column GanttTaskTableColumn) float64 {
	width := t.widthOf(stringBounder, column.header(t.locale))
	for _, task := range t.modelData.Tasks() {
		if t.isDrawable(task) {
			width = math.Max(width, t.widthOf(stringBounder, column.valueOf(task, t.context)))
		}
	}
	return width + 2*cellPadding
}

// Width is getWidth.
func (t *GanttTaskTable) Width() float64 { return t.columnEdges[len(t.columnEdges)-1] }

// DrawU is drawU: the grid, the headers and a row per task.
func (t *GanttTaskTable) DrawU(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	stringBounder := ug.StringBounder()
	t.drawGrid(ug.Apply(t.timelineStyle.LineColor()), totalHeightWithoutFooter)
	for i, column := range t.displayedColumn {
		t.drawCell(ug, t.columnEdges[i], t.headerHeight/2, column.header(t.locale))
	}
	for _, task := range t.modelData.Tasks() {
		if !t.isDrawable(task) {
			continue
		}
		draw := t.drawRegistry.TaskDraw(task)
		yCenter := draw.RealY(stringBounder).CurrentValue() + draw.FullHeightTask(stringBounder)/2
		for i, column := range t.displayedColumn {
			t.drawCell(ug, t.columnEdges[i], yCenter, column.valueOf(task, t.context))
		}
	}
}

func (t *GanttTaskTable) drawGrid(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	width := t.Width()
	ug.Draw(klimt.ULineH(width))
	ug.Apply(klimt.UTranslateDy(t.headerHeight)).Draw(klimt.ULineH(width))
	ug.Apply(klimt.UTranslateDy(totalHeightWithoutFooter)).Draw(klimt.ULineH(width))
	for _, x := range t.columnEdges {
		ug.Apply(klimt.UTranslateDx(x)).Draw(klimt.ULineV(totalHeightWithoutFooter))
	}
}

func (t *GanttTaskTable) drawCell(ug klimt.UGraphic, xLeft, yCenter float64, text string) {
	block := t.textBlock(text)
	dim := block.CalculateDimension(ug.StringBounder())
	block.DrawU(ug.Apply(klimt.NewUTranslate(xLeft+cellPadding, yCenter-dim.Height()/2)))
}

func (t *GanttTaskTable) isDrawable(task Task) bool {
	if _, ok := task.(*TaskImpl); !ok {
		return false
	}
	return !t.timeBounds.IsHidden(task) && t.drawRegistry.TaskDraw(task) != nil
}

func (t *GanttTaskTable) widthOf(stringBounder font.StringBounder, text string) float64 {
	return t.textBlock(text).CalculateDimension(stringBounder).Width()
}

func (t *GanttTaskTable) textBlock(text string) klimt.TextBlock {
	style := ganttStyleQuery(klimt.SNameTimeline).MergedStyle(t.timelineStyle.skinParam.CurrentStyleBuilder())
	return textOf(t.timelineStyle.Pragma(), text, style.FontConfiguration(t.timelineStyle.ColorSet(), nil))
}
