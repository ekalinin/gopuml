// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/TaskDrawDiamond.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/TaskDrawGroup.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/TaskDrawSeparator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// TaskDrawDiamond is upstream's TaskDrawDiamond: the diamond of a
// milestone, or its text.
type TaskDrawDiamond struct {
	abstractTaskDraw
}

// NewTaskDrawDiamond is new TaskDrawDiamond.
func NewTaskDrawDiamond(colorSet *color.HColorSet, timeScale TimeScale, y Real, prettyDisplay string, start TimePoint,
	task Task, toTaskDraw *TaskDrawRegistryData, styleBuilder *klimt.StyleBuilder,
	skinParam klimt.ISkinParam) *TaskDrawDiamond {
	t := &TaskDrawDiamond{newAbstractTaskDraw(colorSet, timeScale, y, prettyDisplay, start, task, toTaskDraw,
		styleBuilder, skinParam)}
	t.styleQuery = func() klimt.StyleSignatureBasic { return ganttStyleQuery(klimt.SNameMilestone) }
	t.shapeHeight = t.ShapeHeight
	t.title = t.titleBlock
	return t
}

func (t *TaskDrawDiamond) yNotePosition(stringBounder font.StringBounder) float64 {
	margin := t.style().Margin()
	return margin.Top() + t.ShapeHeight(stringBounder) + margin.Bottom()
}

// HeightMax implements TaskDraw.
func (t *TaskDrawDiamond) HeightMax(stringBounder font.StringBounder) float64 {
	if t.note == nil {
		return t.FullHeightTask(stringBounder)
	}
	return t.yNotePosition(stringBounder) + t.opaleNote().CalculateDimension(stringBounder).Height()
}

// ShapeHeight is getShapeHeight: the higher of the text and the diamond.
func (t *TaskDrawDiamond) ShapeHeight(stringBounder font.StringBounder) float64 {
	return math.Max(t.title().CalculateDimension(stringBounder).Height(), t.diamondHeight())
}

// diamondHeight is the font size, made even.
func (t *TaskDrawDiamond) diamondHeight() float64 {
	result := int(float32(t.fontConfiguration().Font().Size2D()))
	if result%2 == 1 {
		result--
	}
	return float64(result)
}

// DrawTitle implements TaskDraw: the label right of the diamond.
func (t *TaskDrawDiamond) DrawTitle(ug klimt.UGraphic, labelStrategy LabelStrategy, colTitles, colBars float64) {
	style := t.style()
	margin, padding := style.Margin(), style.Padding()
	title := t.title()
	ug = ug.Apply(klimt.UTranslateDy(margin.Top()))
	titleDim := title.CalculateDimension(ug.StringBounder())
	h := (t.ShapeHeight(ug.StringBounder()) - titleDim.Height()) / 2
	var x float64
	switch {
	case labelStrategy.titleInFirstColumn():
		if labelStrategy.rightAligned() {
			x = colTitles - titleDim.Width() - margin.Right()
		} else {
			x = margin.Left()
		}
	case labelStrategy.titleInLastColumn():
		x = colBars + margin.Left()
	default:
		x1 := t.timeScale.Position(t.start)
		x2 := t.timeScale.Position(t.start) + t.timeScale.Width(t.start)
		width := t.diamondHeight()
		delta := x2 - x1 - width
		x = x2 - delta/2 + padding.Left()
	}
	title.DrawU(ug.Apply(klimt.NewUTranslate(x, h)))
}

func (t *TaskDrawDiamond) titleBlock() klimt.TextBlock {
	return textOf(t.pragma(), t.prettyDisplay, t.fontConfiguration())
}

// LabelOverflow implements TaskDraw.
func (t *TaskDrawDiamond) LabelOverflow(stringBounder font.StringBounder, barsWidth float64) float64 {
	titleDim := t.title().CalculateDimension(stringBounder)
	padding := t.style().Padding()
	x2 := t.timeScale.Position(t.start) + t.timeScale.Width(t.start)
	width := t.diamondHeight()
	delta := x2 - t.timeScale.Position(t.start) - width
	labelStart := x2 - delta/2 + padding.Left()
	labelEnd := labelStart + titleDim.Width()
	overflow := 0.0
	if labelEnd > barsWidth {
		overflow = labelEnd - barsWidth
	}
	if t.note != nil {
		noteEnd := t.timeScale.Position(t.start) + t.opaleNote().CalculateDimension(stringBounder).Width()
		if noteEnd-barsWidth > overflow {
			overflow = noteEnd - barsWidth
		}
	}
	return overflow
}

// DrawU implements klimt.UDrawable: the note, then the diamond or the text
// of "is displayed as".
func (t *TaskDrawDiamond) DrawU(ug klimt.UGraphic) {
	x1 := t.timeScale.Position(t.start)
	t.drawNote(ug.Apply(klimt.NewUTranslate(x1, t.yNotePosition(ug.StringBounder()))))
	if t.url != nil {
		ug.StartURL(t.url)
	}
	margin := t.style().Margin()
	ug = ug.Apply(klimt.UTranslateDy(margin.Top()))
	ug = ug.Apply(klimt.UTranslateDx(x1))
	if displayString, ok := t.task.DisplayString(); !ok {
		x2 := t.timeScale.Position(t.start) + t.timeScale.Width(t.start)
		width := t.diamondHeight()
		delta := x2 - x1 - width
		ug = ug.Apply(klimt.UTranslateDx(delta / 2))
		t.applyColors(ug).Draw(t.diamond())
	} else {
		textOf(t.pragma(), displayString, t.fontConfiguration()).DrawU(ug)
	}
	if t.url != nil {
		ug.CloseURL()
	}
}

// FingerPrintNote implements TaskDraw: none.
func (t *TaskDrawDiamond) FingerPrintNote(font.StringBounder) *FingerPrint { return nil }

// FingerPrint implements TaskDraw, with the width and height of upstream.
func (t *TaskDrawDiamond) FingerPrint(stringBounder font.StringBounder) *FingerPrint {
	h := t.FullHeightTask(stringBounder)
	startPos := t.timeScale.Position(t.start)
	return &FingerPrint{startPos, t.RealY(stringBounder).CurrentValue(), startPos + h,
		t.RealY(stringBounder).CurrentValue() + h}
}

func (t *TaskDrawDiamond) diamond() *klimt.UPolygon {
	h := t.diamondHeight()
	result := klimt.NewUPolygon()
	result.AddPoint(h/2, 0)
	result.AddPoint(h, h/2)
	result.AddPoint(h/2, h)
	result.AddPoint(0, h/2)
	return result
}

// X implements TaskDraw.
func (t *TaskDrawDiamond) X(stringBounder font.StringBounder, side GSide, arrowType GArrowType) float64 {
	var x float64
	switch side {
	case GSideLeft:
		x = t.timeScale.Position(t.start)
	case GSideRight:
		x = t.timeScale.Position(t.start) + t.timeScale.Width(t.start)
	default:
		x = (t.timeScale.Position(t.start) + t.timeScale.Position(t.start) + t.timeScale.Width(t.start)) / 2
	}
	if arrowType == GArrowOutgoing {
		width := t.diamondHeight()
		if side == GSideLeft {
			x += width / 2
		} else if side == GSideRight {
			x -= width / 2
		}
	}
	return x
}

// SideY implements TaskDraw: the middle of the diamond.
func (t *TaskDrawDiamond) SideY(stringBounder font.StringBounder, side GSide) float64 {
	margin := t.style().Margin()
	y1 := margin.Top() + t.RealY(stringBounder).CurrentValue()
	y2 := y1 + t.ShapeHeight(stringBounder)
	if _, ok := t.task.DisplayString(); !ok {
		return y1 + t.diamondHeight()/2
	}
	return (y1 + y2) / 2
}

// TaskDrawGroup is upstream's TaskDrawGroup: the bracket of a group.
type TaskDrawGroup struct {
	abstractTaskDraw
	end TimePoint
}

// groupHeight is the height of the bracket of a group.
const groupHeight = 10

// NewTaskDrawGroup is new TaskDrawGroup.
func NewTaskDrawGroup(colorSet *color.HColorSet, timeScale TimeScale, y Real, prettyDisplay string,
	start, end TimePoint, task Task, toTaskDraw *TaskDrawRegistryData, styleBuilder *klimt.StyleBuilder,
	skinParam klimt.ISkinParam) *TaskDrawGroup {
	t := &TaskDrawGroup{abstractTaskDraw: newAbstractTaskDraw(colorSet, timeScale, y, prettyDisplay, start, task,
		toTaskDraw, styleBuilder, skinParam), end: end}
	t.styleQuery = func() klimt.StyleSignatureBasic { return ganttStyleQuery(klimt.SNameTask) }
	t.shapeHeight = t.ShapeHeight
	t.title = t.titleBlock
	return t
}

// ShapeHeight is getShapeHeight: the text, and 2 more when it fits above
// the bracket.
func (t *TaskDrawGroup) ShapeHeight(stringBounder font.StringBounder) float64 {
	pos1 := t.timeScale.Position(t.start) + 6
	pos2 := t.timeScale.Position(t.end) + t.timeScale.Width(t.end) - 6
	dim := t.title().CalculateDimension(stringBounder)
	if pos2-pos1 > dim.Width() {
		return dim.Height() + 2
	}
	return dim.Height()
}

// DrawTitle implements TaskDraw: the name centered above the bracket when
// it fits, after it otherwise.
func (t *TaskDrawGroup) DrawTitle(ug klimt.UGraphic, labelStrategy LabelStrategy, colTitles, colBars float64) {
	title := t.title()
	stringBounder := ug.StringBounder()
	dim := title.CalculateDimension(stringBounder)
	pos1 := t.timeScale.Position(t.start) + 6
	pos2 := t.timeScale.Position(t.end) + t.timeScale.Width(t.end) - 6
	var pos, y float64
	if pos2-pos1 > dim.Width() {
		pos = pos1 + (pos2-pos1-dim.Width())/2
		y = 0
	} else {
		pos = pos2 + 6
		y = t.FullHeightTask(stringBounder) - dim.Height()
	}
	title.DrawU(ug.Apply(klimt.NewUTranslate(pos, y)))
}

func (t *TaskDrawGroup) titleBlock() klimt.TextBlock {
	return textOf(t.pragma(), t.prettyDisplay, t.fontConfiguration())
}

// LabelOverflow implements TaskDraw.
func (t *TaskDrawGroup) LabelOverflow(stringBounder font.StringBounder, barsWidth float64) float64 {
	dim := t.title().CalculateDimension(stringBounder)
	pos1 := t.timeScale.Position(t.start) + 6
	pos2 := t.timeScale.Position(t.end) + t.timeScale.Width(t.end) - 6
	if pos2-pos1 > dim.Width() {
		return 0
	}
	if labelEnd := pos2 + 6 + dim.Width(); labelEnd > barsWidth {
		return labelEnd - barsWidth
	}
	return 0
}

// DrawU implements klimt.UDrawable.
func (t *TaskDrawGroup) DrawU(ug klimt.UGraphic) { t.drawShape(t.applyColors(ug)) }

// HeightMax implements TaskDraw.
func (t *TaskDrawGroup) HeightMax(stringBounder font.StringBounder) float64 {
	return t.FullHeightTask(stringBounder)
}

// FingerPrint implements TaskDraw.
func (t *TaskDrawGroup) FingerPrint(stringBounder font.StringBounder) *FingerPrint {
	h := t.FullHeightTask(stringBounder)
	startPos := t.timeScale.Position(t.start)
	endPos := t.timeScale.Position(t.end) + t.timeScale.Width(t.end)
	return &FingerPrint{startPos, t.RealY(stringBounder).CurrentValue(), endPos - startPos, h}
}

// FingerPrintNote implements TaskDraw: none.
func (t *TaskDrawGroup) FingerPrintNote(font.StringBounder) *FingerPrint { return nil }

// X implements TaskDraw.
func (t *TaskDrawGroup) X(stringBounder font.StringBounder, side GSide, arrowType GArrowType) float64 {
	switch side {
	case GSideLeft:
		return t.timeScale.Position(t.start)
	case GSideRight:
		return t.timeScale.Position(t.end) + t.timeScale.Width(t.end)
	case GSideTopLeft, GSideBottomLeft:
		return (t.timeScale.Position(t.start) + t.timeScale.Position(t.start) + t.timeScale.Width(t.start)) / 2
	case GSideTopRight, GSideBottomRight:
		return (t.timeScale.Position(t.end) + t.timeScale.Position(t.end) + t.timeScale.Width(t.end)) / 2
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}

func (t *TaskDrawGroup) drawShape(ug klimt.UGraphic) {
	margin := t.style().Margin()
	startPos := t.timeScale.Position(t.start) + margin.Left()
	endPos := t.timeScale.Position(t.end) + t.timeScale.Width(t.end) - margin.Right()
	if t.url != nil {
		ug.StartURL(t.url)
	}
	ug = ug.Apply(klimt.UTranslateDy(t.FullHeightTask(ug.StringBounder()) - groupHeight))
	ug = ug.Apply(color.Black).Apply(klimt.Bg(color.Black))
	ug.Draw(t.shape(startPos, endPos))
	if t.url != nil {
		ug.CloseURL()
	}
}

func (t *TaskDrawGroup) shape(startPos, endPos float64) *klimt.UPath {
	rect := klimt.UPathNone()
	const thick = 2.0
	y1 := (groupHeight - thick) / 2
	y2 := groupHeight - (groupHeight-thick)/2
	const dx = 6.0
	rect.MoveTo(startPos, 0)
	rect.LineTo(startPos+dx, y1)
	rect.LineTo(endPos-dx, y1)
	rect.LineTo(endPos, 0)
	rect.LineTo(endPos, groupHeight)
	rect.LineTo(endPos-dx, y2)
	rect.LineTo(startPos+dx, y2)
	rect.LineTo(startPos, groupHeight)
	rect.LineTo(startPos, 0)
	rect.ClosePath()
	return rect
}

// TaskDrawSeparator is upstream's TaskDrawSeparator: a horizontal line
// across the calendar, with its text.
type TaskDrawSeparator struct {
	timeScale      TimeScale
	y              Real
	minDay, maxDay jtime.LocalDate
	name           string
	hasName        bool
	styleBuilder   *klimt.StyleBuilder
	colorSet       *color.HColorSet
	skinParam      klimt.ISkinParam
}

// NewTaskDrawSeparator is new TaskDrawSeparator; hasName is false for a
// separator without text.
func NewTaskDrawSeparator(name string, hasName bool, timeScale TimeScale, y Real, minDay, maxDay jtime.LocalDate,
	styleBuilder *klimt.StyleBuilder, skinParam klimt.ISkinParam) *TaskDrawSeparator {
	return &TaskDrawSeparator{timeScale: timeScale, y: y, minDay: minDay, maxDay: maxDay, name: name,
		hasName: hasName, styleBuilder: styleBuilder, colorSet: skinParam.IHtmlColorSet(), skinParam: skinParam}
}

// DrawTitle implements TaskDraw.
func (t *TaskDrawSeparator) DrawTitle(ug klimt.UGraphic, labelStrategy LabelStrategy, colTitles, colBars float64) {
	padding, margin := t.style().Padding(), t.style().Margin()
	dx := margin.Left() + padding.Left()
	dy := margin.Top() + padding.Top()
	x := 0.0
	if labelStrategy.titleInFirstColumn() {
		x = colTitles
	}
	t.title().DrawU(ug.Apply(klimt.NewUTranslate(x+dx, dy)))
}

// TitleWidth implements TaskDraw: 0.
func (t *TaskDrawSeparator) TitleWidth(font.StringBounder) float64 { return 0 }

func (t *TaskDrawSeparator) style() *klimt.Style {
	return ganttStyleQuery(klimt.SNameSeparator).MergedStyle(t.styleBuilder)
}

func (t *TaskDrawSeparator) title() klimt.TextBlock {
	if !t.hasName {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	return textOf(t.skinParam.Pragma(), t.name, t.style().FontConfiguration(t.colorSet, nil))
}

// DrawU implements klimt.UDrawable: the background, then the line on each
// side of the text.
func (t *TaskDrawSeparator) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	widthTitle := t.title().CalculateDimension(stringBounder).Width()
	start := t.timeScale.Position(TimePointOfStartOfDay(t.minDay))
	end := t.timeScale.Position(TimePointOfStartOfDay(t.maxDay.PlusDays(1)))
	padding, margin := t.style().Padding(), t.style().Margin()
	ug = ug.Apply(klimt.NewUTranslate(0, margin.Top()))
	backColor := t.style().Value(klimt.PNameBackGroundColor).AsColor(t.colorSet)
	if !backColor.IsTransparent() {
		height := padding.Top() + t.textHeight(stringBounder) + padding.Bottom()
		if height > 0 {
			ug.Apply(klimt.Bg(backColor)).Draw(klimt.URectangleBuild(end-start, height))
		}
	}
	lineColor := t.style().Value(klimt.PNameLineColor).AsColor(t.colorSet)
	ug = ug.Apply(lineColor)
	ug = ug.Apply(klimt.UTranslateDy(padding.Top() + t.textHeight(stringBounder)/2))
	if widthTitle == 0 {
		ug.Draw(klimt.ULineH(end - start))
		return
	}
	if padding.Left() > 1 {
		ug.Draw(klimt.ULineH(padding.Left()))
	}
	x1 := padding.Left() + margin.Left() + widthTitle + margin.Right()
	x2 := end - 1
	ug.Apply(klimt.UTranslateDx(x1)).Draw(klimt.ULineH(x2 - x1))
}

// FingerPrint implements TaskDraw, with the width and height of upstream.
func (t *TaskDrawSeparator) FingerPrint(stringBounder font.StringBounder) *FingerPrint {
	h := t.FullHeightTask(stringBounder)
	end := t.timeScale.Position(TimePointOfStartOfDay(t.maxDay.PlusDays(1)))
	return &FingerPrint{0, t.RealY(stringBounder).CurrentValue(), end, t.RealY(stringBounder).CurrentValue() + h}
}

// FingerPrintNote implements TaskDraw: none.
func (t *TaskDrawSeparator) FingerPrintNote(font.StringBounder) *FingerPrint { return nil }

// FullHeightTask implements TaskDraw.
func (t *TaskDrawSeparator) FullHeightTask(stringBounder font.StringBounder) float64 {
	padding, margin := t.style().Padding(), t.style().Margin()
	return margin.Top() + padding.Top() + t.textHeight(stringBounder) + padding.Bottom() + margin.Bottom()
}

func (t *TaskDrawSeparator) textHeight(stringBounder font.StringBounder) float64 {
	return t.title().CalculateDimension(stringBounder).Height()
}

// RealY implements TaskDraw.
func (t *TaskDrawSeparator) RealY(font.StringBounder) Real { return t.y }

// TrueRow implements TaskDraw: none.
func (t *TaskDrawSeparator) TrueRow() TaskDraw { return nil }

// SetColorsAndCompletion implements TaskDraw: nothing.
func (t *TaskDrawSeparator) SetColorsAndCompletion(*CenterBorderColor, int, *klimt.URL, *klimt.Display,
	*klimt.Stereotype) {
}

// Task implements TaskDraw, which fails with Java's
// UnsupportedOperationException.
func (t *TaskDrawSeparator) Task() Task {
	unsupported()
	return nil
}

// SideY implements TaskDraw, which fails with Java's
// UnsupportedOperationException.
func (t *TaskDrawSeparator) SideY(font.StringBounder, GSide) float64 {
	unsupported()
	return 0
}

// HeightMax implements TaskDraw.
func (t *TaskDrawSeparator) HeightMax(stringBounder font.StringBounder) float64 {
	return t.FullHeightTask(stringBounder)
}

// LabelOverflow implements TaskDraw: 0.
func (t *TaskDrawSeparator) LabelOverflow(font.StringBounder, float64) float64 { return 0 }

// X implements TaskDraw, which fails with Java's
// UnsupportedOperationException.
func (t *TaskDrawSeparator) X(font.StringBounder, GSide, GArrowType) float64 {
	unsupported()
	return 0
}
