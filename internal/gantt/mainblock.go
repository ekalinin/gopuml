// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttDiagramMainBlock.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttLayout.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ganttLayout is upstream's GanttLayout: the drawings of the tasks, laid
// out, and the sizes of the parts of the diagram.
type ganttLayout struct {
	titlesWidth, barsWidth, timelineWidth   float64
	totalHeight, headerHeight, footerHeight float64
}

func newGanttLayout(modelData *GanttModelData, displayConfig *DisplayConfigData, timeBounds *TimeBoundsData,
	timelineStyle *TimelineStyleData, drawRegistry *TaskDrawRegistryData, stringBounder font.StringBounder,
	timeHeader TimeHeader) *ganttLayout {
	drawRegistry.BuildTaskAndResourceDraws(stringBounder, timeHeader, modelData, displayConfig, timeBounds,
		timelineStyle)
	l := &ganttLayout{}
	scale := timeHeader.TimeScale()
	l.timelineWidth = scale.Position(TimePointOfStartOfDay(timeBounds.MaxDay()).AddDays(1)) -
		scale.Position(TimePointOfStartOfDay(timeBounds.MinDay()))
	maxLabelOverflow := 0.0
	for _, task := range modelData.Tasks() {
		if timeBounds.IsHidden(task) {
			continue
		}
		if draw := drawRegistry.TaskDraw(task); draw != nil {
			maxLabelOverflow = math.Max(maxLabelOverflow, draw.LabelOverflow(stringBounder, l.timelineWidth))
		}
	}
	l.barsWidth = l.timelineWidth + maxLabelOverflow
	l.headerHeight = timeHeader.TimeHeaderHeight(stringBounder)
	if displayConfig.showFootbox {
		l.footerHeight = timeHeader.TimeFooterHeight(stringBounder)
	}
	l.totalHeight = drawRegistry.TotalHeightWithoutFooter() + l.footerHeight
	return l
}

// ganttDiagramMainBlock is upstream's GanttDiagramMainBlock: the table, the
// calendar, the arrows, the tasks and the resources.
type ganttDiagramMainBlock struct {
	klimt.TextBlockDefaults
	memo          klimt.TextBlockMemoized
	timeHeader    TimeHeader
	layout        *ganttLayout
	modelData     *GanttModelData
	timeBounds    *TimeBoundsData
	displayConfig *DisplayConfigData
	timelineStyle *TimelineStyleData
	drawRegistry  *TaskDrawRegistryData
	taskTable     *GanttTaskTable
}

func newGanttDiagramMainBlock(d *GanttDiagram, timelineStyle *TimelineStyleData, timeHeader TimeHeader,
	stringBounder font.StringBounder) *ganttDiagramMainBlock {
	b := &ganttDiagramMainBlock{timeHeader: timeHeader, modelData: d.modelData, timeBounds: d.timeBounds,
		displayConfig: d.displayConfig, timelineStyle: timelineStyle, drawRegistry: d.drawRegistry}
	b.layout = newGanttLayout(d.modelData, d.displayConfig, d.timeBounds, timelineStyle, d.drawRegistry,
		stringBounder, timeHeader)
	b.taskTable = NewGanttTaskTable(d.modelData, d.timeBounds, d.drawRegistry, timelineStyle,
		timeHeader.FullHeaderHeight(stringBounder), d.weekConfig.locale, stringBounder, d.DisplayedColumn())
	return b
}

// DrawU implements klimt.TextBlock. A panic is not caught here: the export
// returns it as an error, where upstream draws a crash report.
func (b *ganttDiagramMainBlock) DrawU(ug klimt.UGraphic) {
	totalHeightWithoutFooter := b.drawRegistry.TotalHeightWithoutFooter()
	tableWidth := b.taskTable.Width()
	b.taskTable.DrawU(ug, totalHeightWithoutFooter)
	ug = ug.Apply(klimt.UTranslateDx(tableWidth))
	ugOrig := ug
	style := ganttStyleQuery(klimt.SNameTimeline).MergedStyle(b.timelineStyle.skinParam.CurrentStyleBuilder())
	back := style.Value(klimt.PNameBackGroundColor).AsColor(b.timelineStyle.ColorSet())
	if !back.IsTransparent() {
		fullWidth := b.layout.titlesWidth + b.layout.timelineWidth
		ug.Apply(klimt.Bg(back)).Draw(klimt.URectangleBuild(fullWidth, b.layout.headerHeight))
		if b.displayConfig.showFootbox {
			ug.Apply(klimt.Bg(back)).Apply(klimt.UTranslateDy(totalHeightWithoutFooter)).
				Draw(klimt.URectangleBuild(fullWidth, b.layout.footerHeight))
		}
	}
	b.timeHeader.DrawTimeHeader(ug, totalHeightWithoutFooter)
	b.drawConstraints(ug, b.timeHeader.TimeScale())
	b.drawTasksRect(ug)
	b.drawTasksTitle(ugOrig, b.layout.titlesWidth, b.layout.barsWidth)
	if !b.displayConfig.hideResourceFootbox {
		b.drawResources(ug)
	}
	if b.displayConfig.showFootbox {
		b.timeHeader.DrawTimeFooter(ug.Apply(klimt.UTranslateDy(totalHeightWithoutFooter)))
	}
}

// CalculateDimension implements klimt.TextBlock: the table, the calendar
// with the labels beyond it and a margin of 20, and the height of the
// layout.
func (b *ganttDiagramMainBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		const margin = 20
		width := b.taskTable.Width() + b.layout.titlesWidth + b.layout.barsWidth + margin
		return geom.NewXDimension2D(width, b.layout.totalHeight)
	})
}

func (b *ganttDiagramMainBlock) drawTasksRect(ug klimt.UGraphic) {
	for _, task := range b.modelData.Tasks() {
		if b.timeBounds.IsHidden(task) {
			continue
		}
		draw := b.drawRegistry.TaskDraw(task)
		draw.DrawU(ug.Apply(klimt.UTranslateDy(draw.RealY(ug.StringBounder()).CurrentValue())))
	}
}

func (b *ganttDiagramMainBlock) drawConstraints(ug klimt.UGraphic, timeScale TimeScale) {
	for _, constraint := range b.modelData.Constraints() {
		if _, ok := b.timeBounds.PrintStart(); ok && constraint.IsHidden(TimePointOfStartOfDay(b.timeBounds.MinDay()),
			TimePointOfEndOfDayMinusOneSecond(b.timeBounds.MaxDay())) {
			continue
		}
		constraint.UDrawable(b.timelineStyle.ColorSet(), timeScale, b.drawRegistry).DrawU(ug)
	}
}

func (b *ganttDiagramMainBlock) drawTasksTitle(ug klimt.UGraphic, colTitles, colBars float64) {
	labelStrategy := LabelStrategy{LabelLegacy, geom.HorizontalAlignmentLeft}
	for _, task := range b.modelData.Tasks() {
		if b.timeBounds.IsHidden(task) {
			continue
		}
		draw := b.drawRegistry.TaskDraw(task)
		draw.DrawTitle(ug.Apply(klimt.UTranslateDy(draw.RealY(ug.StringBounder()).CurrentValue())), labelStrategy,
			colTitles, colBars)
	}
}

func (b *ganttDiagramMainBlock) drawResources(ug klimt.UGraphic) {
	for _, res := range b.modelData.Resources() {
		draw := res.ResourceDraw()
		draw.DrawU(ug.Apply(klimt.UTranslateDy(draw.Y())))
	}
}
