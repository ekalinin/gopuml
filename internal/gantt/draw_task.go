// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/TaskDraw.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/AbstractTaskDraw.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/FingerPrint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/PathUtils.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/LabelStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/LabelPosition.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/CenterBorderColor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/sequencediagram"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Real is upstream's Real: a position solved after the layout.
type Real = sequencediagram.Real

// TaskDraw is upstream's TaskDraw: the drawing of a task, a milestone, a
// group or a separator.
type TaskDraw interface {
	klimt.UDrawable
	TrueRow() TaskDraw
	SetColorsAndCompletion(colors *CenterBorderColor, completion int, url *klimt.URL, note *klimt.Display,
		noteStereotype *klimt.Stereotype)
	// RealY is getY(StringBounder): the top of the row.
	RealY(stringBounder font.StringBounder) Real
	// SideY is getY(StringBounder, GSide): where an arrow touches a side.
	SideY(stringBounder font.StringBounder, side GSide) float64
	DrawTitle(ug klimt.UGraphic, labelStrategy LabelStrategy, colTitles, colBars float64)
	TitleWidth(stringBounder font.StringBounder) float64
	FullHeightTask(stringBounder font.StringBounder) float64
	HeightMax(stringBounder font.StringBounder) float64
	Task() Task
	FingerPrint(stringBounder font.StringBounder) *FingerPrint
	FingerPrintNote(stringBounder font.StringBounder) *FingerPrint
	LabelOverflow(stringBounder font.StringBounder, barsWidth float64) float64
	X(stringBounder font.StringBounder, side GSide, arrowType GArrowType) float64
}

// LabelPosition is upstream's LabelPosition.
type LabelPosition int

// The positions of the labels of the tasks.
const (
	LabelLegacy LabelPosition = iota
	LabelFirstColumn
	LabelLastColumn
)

// LabelStrategy is upstream's LabelStrategy: where the labels of the tasks
// are drawn. Upstream always draws them as LEGACY and LEFT.
type LabelStrategy struct {
	position  LabelPosition
	alignment geom.HorizontalAlignment
}

func (s LabelStrategy) titleInFirstColumn() bool { return s.position == LabelFirstColumn }

func (s LabelStrategy) titleInLastColumn() bool { return s.position == LabelLastColumn }

func (s LabelStrategy) rightAligned() bool { return s.alignment == geom.HorizontalAlignmentRight }

// FingerPrint is upstream's FingerPrint: the rectangle a task or a note
// takes.
type FingerPrint struct {
	x, y, width, height float64
}

// Overlap is overlap: how far the other rectangle must move down for the
// two not to overlap, 0 when they do not.
func (f *FingerPrint) Overlap(other *FingerPrint) float64 {
	if f.x >= other.x+other.width || other.x >= f.x+f.width {
		return 0
	}
	if f.y >= other.y+other.height || other.y >= f.y+f.height {
		return 0
	}
	return f.y + f.height - other.y
}

// pathUtoRight is PathUtils.UtoRight: a bar open on the left.
func pathUtoRight(width, height, round float64) *klimt.UPath {
	result := klimt.UPathNone()
	if round == 0 {
		result.MoveTo(0, 0)
		result.LineTo(width, 0)
		result.LineTo(width, height)
		result.LineTo(0, height)
	} else {
		halfRound := round / 2
		result.MoveTo(0, 0)
		result.LineTo(width-halfRound, 0)
		result.ArcToPoint(geom.NewXPoint2D(width, halfRound), halfRound, 0, 1)
		result.LineTo(width, height-halfRound)
		result.ArcToPoint(geom.NewXPoint2D(width-halfRound, height), halfRound, 0, 1)
		result.LineTo(0, height)
	}
	return result
}

// pathUtoLeft is PathUtils.UtoLeft: a bar open on the right.
func pathUtoLeft(width, height, round float64) *klimt.UPath {
	result := klimt.UPathNone()
	if round == 0 {
		result.MoveTo(width, height)
		result.LineTo(0, height)
		result.LineTo(0, 0)
		result.LineTo(width, 0)
	} else {
		halfRound := round / 2
		result.MoveTo(width, height)
		result.LineTo(halfRound, height)
		result.ArcToPoint(geom.NewXPoint2D(0, height-halfRound), halfRound, 0, 1)
		result.LineTo(0, halfRound)
		result.ArcToPoint(geom.NewXPoint2D(halfRound, 0), halfRound, 0, 1)
		result.LineTo(width, 0)
	}
	return result
}

// apply is CenterBorderColor.apply: the fill, then the border or the fill
// as the line color.
func (c *CenterBorderColor) apply(ug klimt.UGraphic) klimt.UGraphic {
	if !c.IsOk() {
		uml.Throw(uml.IllegalStateException, "")
	}
	ug = ug.Apply(klimt.Bg(c.center))
	if c.border == nil {
		return ug.Apply(c.center)
	}
	return ug.Apply(c.border)
}

// abstractTaskDraw is upstream's AbstractTaskDraw: what the drawings of
// tasks, milestones and groups share.
type abstractTaskDraw struct {
	// styleQuery is getStyleQuery of the kind of drawing.
	styleQuery     func() klimt.StyleSignatureBasic
	shapeHeight    func(stringBounder font.StringBounder) float64
	title          func() klimt.TextBlock
	colors         *CenterBorderColor
	completion     int
	url            *klimt.URL
	note           *klimt.Display
	noteStereotype *klimt.Stereotype
	timeScale      TimeScale
	y              Real
	prettyDisplay  string
	start          TimePoint
	styleBuilder   *klimt.StyleBuilder
	task           Task
	toTaskDraw     *TaskDrawRegistryData
	skinParam      klimt.ISkinParam
	colorSet       *color.HColorSet
}

func newAbstractTaskDraw(colorSet *color.HColorSet, timeScale TimeScale, y Real, prettyDisplay string,
	start TimePoint, task Task, toTaskDraw *TaskDrawRegistryData, styleBuilder *klimt.StyleBuilder,
	skinParam klimt.ISkinParam) abstractTaskDraw {
	return abstractTaskDraw{completion: 100, timeScale: timeScale, y: y, prettyDisplay: prettyDisplay, start: start,
		styleBuilder: styleBuilder, task: task, toTaskDraw: toTaskDraw, skinParam: skinParam, colorSet: colorSet}
}

// ganttStyleQuery is StyleQueries.GANTTDIAG with names.
func ganttStyleQuery(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameGanttDiagram}, names...)...)
}

// SetColorsAndCompletion implements TaskDraw.
func (t *abstractTaskDraw) SetColorsAndCompletion(colors *CenterBorderColor, completion int, url *klimt.URL,
	note *klimt.Display, noteStereotype *klimt.Stereotype) {
	t.colors = colors
	t.completion = completion
	t.url = url
	t.note = note
	t.noteStereotype = noteStereotype
}

// lineColor is getLineColor: between the colors of an unstarted and a
// started task, by the completion.
func (t *abstractTaskDraw) lineColor() color.HColor {
	unstarted := ganttStyleQuery(klimt.SNameTask, klimt.SNameUnstarted).MergedStyle(t.styleBuilder).
		Value(klimt.PNameLineColor).AsColor(t.colorSet)
	regular := t.style().Value(klimt.PNameLineColor).AsColor(t.colorSet)
	return color.Unlinear(unstarted, regular, t.completion)
}

// backgroundColor is getBackgroundColor.
func (t *abstractTaskDraw) backgroundColor() color.HColor {
	unstarted := ganttStyleQuery(klimt.SNameTask, klimt.SNameUnstarted).MergedStyle(t.styleBuilder).
		Value(klimt.PNameBackGroundColor).AsColor(t.colorSet)
	regular := t.style().Value(klimt.PNameBackGroundColor).AsColor(t.colorSet)
	return color.Unlinear(unstarted, regular, t.completion)
}

func (t *abstractTaskDraw) fontConfiguration() *klimt.FontConfiguration {
	return t.style().FontConfiguration(t.colorSet, nil)
}

func (t *abstractTaskDraw) style() *klimt.Style { return t.styleQuery().MergedStyle(t.styleBuilder) }

// TitleWidth implements TaskDraw.
func (t *abstractTaskDraw) TitleWidth(stringBounder font.StringBounder) float64 {
	margin := t.style().Margin()
	return margin.Left() + t.title().CalculateDimension(stringBounder).Width() + margin.Right()
}

// FullHeightTask implements TaskDraw.
func (t *abstractTaskDraw) FullHeightTask(stringBounder font.StringBounder) float64 {
	margin := t.style().Margin()
	return margin.Top() + t.shapeHeight(stringBounder) + margin.Bottom()
}

// TrueRow implements TaskDraw: the drawing of the task this one is on the
// row of.
func (t *abstractTaskDraw) TrueRow() TaskDraw { return t.toTaskDraw.TaskDraw(t.task.Row()) }

// RealY implements TaskDraw.
func (t *abstractTaskDraw) RealY(stringBounder font.StringBounder) Real {
	if t.task.Row() == nil {
		return t.y
	}
	return t.TrueRow().RealY(stringBounder)
}

// Task implements TaskDraw.
func (t *abstractTaskDraw) Task() Task { return t.task }

// SideY implements TaskDraw.
func (t *abstractTaskDraw) SideY(stringBounder font.StringBounder, side GSide) float64 {
	margin := t.style().Margin()
	y1 := margin.Top() + t.RealY(stringBounder).CurrentValue()
	y2 := y1 + t.shapeHeight(stringBounder)
	if side.IsTop() {
		return y1
	}
	if side.IsBottom() {
		return y2
	}
	return (y1 + y2) / 2
}

// LabelOverflow is the default of TaskDraw.getLabelOverflow: 0.
func (t *abstractTaskDraw) LabelOverflow(font.StringBounder, float64) float64 { return 0 }

func (t *abstractTaskDraw) pragma() *klimt.Pragma { return t.skinParam.Pragma() }

// opaleNote is getOpaleNote: the note of the task.
func (t *abstractTaskDraw) opaleNote() *klimt.Opale {
	style := ganttStyleQuery(klimt.SNameNote).AddStereotypeOf(t.noteStereotype).MergedStyle(t.styleBuilder)
	fc := style.FontConfiguration(t.colorSet, nil)
	horizontalAlignment := style.Value(klimt.PNameHorizontalAlignment).AsHorizontalAlignment()
	sheet := t.skinParam.Sheet(fc, horizontalAlignment, klimt.CreoleModeFull).CreateSheet(t.note)
	sheet1 := klimt.NewSheetBlock1Padding(sheet, klimt.LineBreakStrategyNone, style.Padding())
	noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(t.colorSet)
	borderColor := style.Value(klimt.PNameLineColor).AsColor(t.colorSet)
	stroke := style.Stroke()
	return klimt.NewOpale(style.Shadowing(), borderColor, noteBackgroundColor, sheet1, false, &stroke)
}

func (t *abstractTaskDraw) drawNote(ug klimt.UGraphic) {
	if t.note == nil {
		return
	}
	t.opaleNote().DrawU(ug)
}

// applyColors is applyColors: the colors of the task, or of its style.
func (t *abstractTaskDraw) applyColors(ug klimt.UGraphic) klimt.UGraphic {
	if t.colors != nil && t.colors.IsOk() {
		return t.colors.apply(ug)
	}
	return ug.Apply(t.lineColor()).Apply(klimt.Bg(t.backgroundColor()))
}

// textOf is Display.getWithNewlines(pragma, text).create(fc, LEFT, ...).
func textOf(pragma *klimt.Pragma, text string, fc *klimt.FontConfiguration) klimt.TextBlock {
	return klimt.DisplayWithNewlines(pragma, text).Create(fc, geom.HorizontalAlignmentLeft, klimt.SpriteContainerEmpty{})
}
