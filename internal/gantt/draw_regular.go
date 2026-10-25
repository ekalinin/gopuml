// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/TaskDrawRegular.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/RectangleTask.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"maps"
	"slices"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/sequencediagram"
	"github.com/ekalinin/gopuml/internal/uml"
)

// TaskDrawRegular is upstream's TaskDrawRegular: the bar of a task.
type TaskDrawRegular struct {
	abstractTaskDraw
	end         TimePoint
	oddStart    bool
	oddEnd      bool
	paused      []jtime.LocalDate
	constraints []*GanttConstraint
}

// NewTaskDrawRegular is new TaskDrawRegular: the paused days of the task
// and the days its calendar closes are drawn as gaps.
func NewTaskDrawRegular(colorSet *color.HColorSet, timeScale TimeScale, y Real, prettyDisplay string,
	start, end TimePoint, oddStart, oddEnd bool, skinParam klimt.ISkinParam, task Task,
	toTaskDraw *TaskDrawRegistryData, constraints []*GanttConstraint,
	styleBuilder *klimt.StyleBuilder) *TaskDrawRegular {
	t := &TaskDrawRegular{
		abstractTaskDraw: newAbstractTaskDraw(colorSet, timeScale, y, prettyDisplay, start, task, toTaskDraw,
			styleBuilder, skinParam),
		end: end, oddStart: oddStart, oddEnd: oddEnd, constraints: constraints}
	t.styleQuery = func() klimt.StyleSignatureBasic {
		return ganttStyleQuery(klimt.SNameTask).AddStereotypeOf(t.task.Stereotype())
	}
	t.shapeHeight = t.ShapeHeight
	t.title = t.titleBlock
	impl := task.(*TaskImpl)
	paused := map[jtime.LocalDate]bool{}
	for _, day := range impl.AllPaused() {
		paused[day] = true
	}
	defaultPlan := impl.DefaultPlan()
	for tmp := start; tmp.CompareTo(end) <= 0; tmp = tmp.Increment() {
		if ngm.IsZeroOnDay(defaultPlan, tmp.Day()) {
			paused[tmp.Day()] = true
		}
	}
	t.paused = slices.SortedFunc(maps.Keys(paused), jtime.LocalDate.CompareTo)
	return t
}

// ShapeHeight is getShapeHeight: the text with the padding.
func (t *TaskDrawRegular) ShapeHeight(stringBounder font.StringBounder) float64 {
	padding := t.style().Padding()
	return padding.Top() + t.title().CalculateDimension(stringBounder).Height() + padding.Bottom()
}

// DrawTitle implements TaskDraw: the label inside the bar when it fits,
// after it otherwise.
func (t *TaskDrawRegular) DrawTitle(ug klimt.UGraphic, labelStrategy LabelStrategy, colTitles, colBars float64) {
	title := t.title()
	dim := title.CalculateDimension(ug.StringBounder())
	style := t.style()
	margin, padding := style.Margin(), style.Padding()
	ug = ug.Apply(klimt.UTranslateDy(margin.Top() + padding.Top()))
	switch {
	case labelStrategy.titleInFirstColumn():
		if labelStrategy.rightAligned() {
			title.DrawU(ug.Apply(klimt.UTranslateDx(colTitles - dim.Width() - margin.Right())))
		} else {
			title.DrawU(ug.Apply(klimt.UTranslateDx(margin.Left())))
		}
		return
	case labelStrategy.titleInLastColumn():
		title.DrawU(ug.Apply(klimt.UTranslateDx(colBars + margin.Left())))
		return
	}
	pos1 := t.timeScale.Position(t.start) + 6
	pos2 := t.timeScale.Position(t.end) - 6
	var pos float64
	if pos2-pos1 > dim.Width() {
		pos = pos1
	} else {
		pos = t.outPosition(pos2)
	}
	title.DrawU(ug.Apply(klimt.UTranslateDx(pos)))
}

func (t *TaskDrawRegular) titleBlock() klimt.TextBlock {
	return textOf(t.pragma(), t.prettyDisplay, t.fontConfiguration())
}

func (t *TaskDrawRegular) outPosition(pos2 float64) float64 {
	if t.isThereRightArrow() {
		return pos2 + 18
	}
	return pos2 + 8
}

func (t *TaskDrawRegular) isThereRightArrow() bool {
	for _, constraint := range t.constraints {
		if constraint.IsThereRightArrow(t.task) {
			return true
		}
	}
	return false
}

// LabelOverflow implements TaskDraw: how far the label or the note goes
// beyond the calendar.
func (t *TaskDrawRegular) LabelOverflow(stringBounder font.StringBounder, barsWidth float64) float64 {
	overflow := 0.0
	dim := t.title().CalculateDimension(stringBounder)
	pos1 := t.timeScale.Position(t.start) + 6
	pos2 := t.timeScale.Position(t.end) - 6
	if pos2-pos1 <= dim.Width() {
		labelEnd := t.outPosition(pos2) + dim.Width()
		if labelEnd > barsWidth {
			overflow = labelEnd - barsWidth
		}
	}
	if t.note != nil {
		noteEnd := t.timeScale.Position(t.start) + t.opaleNote().CalculateDimension(stringBounder).Width()
		if noteEnd-barsWidth > overflow {
			overflow = noteEnd - barsWidth
		}
	}
	return overflow
}

// DrawU implements klimt.UDrawable: the note, then the bar.
func (t *TaskDrawRegular) DrawU(ug klimt.UGraphic) {
	startPos := t.timeScale.Position(t.start)
	t.drawNote(ug.Apply(klimt.NewUTranslate(startPos, t.yNotePosition(ug.StringBounder()))))
	t.drawShape(ug)
}

func (t *TaskDrawRegular) yNotePosition(stringBounder font.StringBounder) float64 {
	margin := t.style().Margin()
	return margin.Top() + t.ShapeHeight(stringBounder) + margin.Bottom()
}

// HeightMax implements TaskDraw.
func (t *TaskDrawRegular) HeightMax(stringBounder font.StringBounder) float64 {
	if t.note == nil {
		return t.FullHeightTask(stringBounder)
	}
	return t.yNotePosition(stringBounder) + t.opaleNote().CalculateDimension(stringBounder).Height()
}

// FingerPrint implements TaskDraw.
func (t *TaskDrawRegular) FingerPrint(stringBounder font.StringBounder) *FingerPrint {
	h := t.FullHeightTask(stringBounder)
	startPos := t.timeScale.Position(t.start)
	endPos := t.timeScale.Position(t.end)
	return &FingerPrint{startPos, t.RealY(stringBounder).CurrentValue(), endPos - startPos, h}
}

// FingerPrintNote implements TaskDraw.
func (t *TaskDrawRegular) FingerPrintNote(stringBounder font.StringBounder) *FingerPrint {
	if t.note == nil {
		return nil
	}
	dim := t.opaleNote().CalculateDimension(stringBounder)
	startPos := t.timeScale.Position(t.start)
	return &FingerPrint{startPos, t.RealY(stringBounder).CurrentValue() + t.yNotePosition(stringBounder),
		dim.Width(), dim.Height()}
}

// X implements TaskDraw.
func (t *TaskDrawRegular) X(stringBounder font.StringBounder, side GSide, arrowType GArrowType) float64 {
	var x float64
	switch side {
	case GSideLeft:
		x = t.timeScale.Position(t.start)
	case GSideRight:
		x = t.timeScale.Position(t.end)
	case GSideTopLeft, GSideBottomLeft:
		x = t.timeScale.Position(t.start) + 8
		if x > t.timeScale.Position(t.end) {
			x = (t.timeScale.Position(t.start) + t.timeScale.Position(t.start) + t.timeScale.Width(t.start)) / 2
		}
	case GSideTopRight, GSideBottomRight:
		x = t.timeScale.Position(t.end) - 8
		if x < t.timeScale.Position(t.start)+t.timeScale.Width(t.start) {
			x = t.timeScale.Position(t.end) - 8
		}
	default:
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if arrowType == GArrowOutgoing {
		margin := t.style().Margin()
		if side == GSideLeft {
			x += margin.Left()
		} else if side == GSideRight {
			x -= margin.Left()
		}
	}
	return x
}

func (t *TaskDrawRegular) drawShape(ug klimt.UGraphic) {
	ug = t.applyColors(ug)
	style := t.style()
	margin := style.Margin()
	startPos := t.timeScale.Position(t.start) + margin.Left()
	endPos := t.timeScale.Position(t.end) - margin.Right()
	if t.url != nil {
		ug.StartURL(t.url)
	}
	ug = ug.Apply(klimt.UTranslateDy(margin.Top()))
	round := style.Value(klimt.PNameRoundCorner).AsDouble()
	var off []sequencediagram.Segment
	for _, day := range t.paused {
		pause := TimePointOfStartOfDay(day)
		x1 := t.timeScale.Position(pause)
		x2 := t.timeScale.Position(pause) + t.timeScale.Width(pause)
		off = append(off, sequencediagram.NewSegment(x1, x2))
	}
	backUndone := ganttStyleQuery(klimt.SNameUndone).MergedStyle(t.styleBuilder).Value(klimt.PNameBackGroundColor).
		AsColor(t.colorSet)
	newRectangleTask(startPos, endPos, round, t.completion, off).draw(ug, t.ShapeHeight(ug.StringBounder()),
		backUndone, t.oddStart, t.oddEnd)
	if t.url != nil {
		ug.CloseURL()
	}
}

// rectangleTask is upstream's RectangleTask: a bar, cut by the paused days,
// filled by its completion.
type rectangleTask struct {
	segments   []sequencediagram.Segment
	round      float64
	completion int
}

func newRectangleTask(startPos, endPos, round float64, completion int,
	paused []sequencediagram.Segment) *rectangleTask {
	r := &rectangleTask{round: round, completion: completion}
	if startPos < endPos {
		r.segments = sequencediagram.NewSegment(startPos, endPos).CutSegmentIfNeed(paused)
	} else {
		r.segments = []sequencediagram.Segment{sequencediagram.NewSegment(startPos, startPos+1)}
	}
	return r
}

func draw2hlines(ug klimt.UGraphic, height float64, hline *klimt.ULine) {
	ug.Draw(hline)
	ug.Apply(klimt.UTranslateDy(height)).Draw(hline)
}

// fraction is a * completion / 100, in Java's order of operations.
func (r *rectangleTask) fraction(a float64, completion int) float64 {
	return a * float64(completion) / 100
}

func (r *rectangleTask) draw(ug klimt.UGraphic, height float64, documentBackground color.HColor,
	oddStart, oddEnd bool) {
	if r.round == 0 {
		r.drawWithoutRound(ug, height, documentBackground, oddStart, oddEnd)
		return
	}
	if len(r.segments) != 1 {
		r.drawWithRound(ug, height, documentBackground)
		return
	}
	segment := r.segments[0]
	width := segment.Length()
	partial := klimt.URectangleBuild(width, height).Rounded(r.round)
	if r.completion == 100 || r.completion == 0 {
		if r.completion == 0 {
			ug = ug.Apply(klimt.Bg(documentBackground))
		}
		switch {
		case oddStart && !oddEnd:
			ug.Apply(klimt.UTranslateDx(segment.Pos1())).Draw(pathUtoRight(width, height, r.round))
		case !oddStart && oddEnd:
			ug.Apply(klimt.UTranslateDx(segment.Pos1())).Draw(pathUtoLeft(width, height, r.round))
		default:
			ug.Apply(klimt.UTranslateDx(segment.Pos1())).Draw(partial)
		}
		return
	}
	x1 := r.fraction(width, r.completion)
	ug.Apply(color.None()).Apply(klimt.UTranslateDx(segment.Pos1())).Draw(pathUtoLeft(x1, height, r.round))
	ug.Apply(klimt.Bg(documentBackground)).Apply(color.None()).Apply(klimt.UTranslateDx(segment.Pos1() + x1)).
		Draw(pathUtoRight(r.fraction(width, 100-r.completion), height, r.round))
	ug.Apply(klimt.Bg(color.None())).Apply(klimt.UTranslateDx(segment.Pos1())).Draw(partial)
}

func (r *rectangleTask) drawWithRound(ug klimt.UGraphic, height float64, documentBackground color.HColor) {
	first := r.segments[0]
	ug.Apply(klimt.UTranslateDx(first.Pos1())).Draw(pathUtoLeft(first.Length(), height, r.round))
	for i := 1; i < len(r.segments)-1; i++ {
		segment := r.segments[i]
		r.drawPartly(r.fraction(segment.Length(), r.completion), ug, segment, height, documentBackground, i, false,
			false)
	}
	last := r.segments[len(r.segments)-1]
	ug.Apply(klimt.UTranslateDx(last.Pos1())).Draw(pathUtoRight(last.Length(), height, r.round))
	r.drawIntermediateDotted(ug, height)
}

func (r *rectangleTask) drawWithoutRound(ug klimt.UGraphic, height float64, documentBackground color.HColor,
	oddStart, oddEnd bool) {
	sumWidth := r.fullSegmentsLength()
	lim := sumWidth
	if r.completion != 100 {
		lim = r.fraction(sumWidth, r.completion)
	}
	if len(r.segments) == 1 && !oddStart && !oddEnd {
		widthCompletion := lim
		if r.completion == 100 {
			widthCompletion = -1
		}
		r.drawFull(widthCompletion, ug, r.segments[0], height, documentBackground)
		return
	}
	current := 0.0
	for i, segment := range r.segments {
		next := current + segment.Length()
		var widthCompletion float64
		switch {
		case lim >= next:
			widthCompletion = -1
		case current >= lim:
			widthCompletion = 0
		default:
			widthCompletion = lim - current
		}
		r.drawPartly(widthCompletion, ug, segment, height, documentBackground, i, !oddStart && i == 0,
			!oddEnd && i == len(r.segments)-1)
		current = next
	}
	r.drawIntermediateDotted(ug, height)
}

func (r *rectangleTask) fullSegmentsLength() float64 {
	result := 0.0
	for _, seg := range r.segments {
		result += seg.Length()
	}
	return result
}

func (r *rectangleTask) drawIntermediateDotted(ug klimt.UGraphic, height float64) {
	ug = ug.Apply(klimt.NewUStroke(2, 3, 1))
	for i := 0; i < len(r.segments)-1; i++ {
		v1 := r.segments[i].Pos2() + 3
		v2 := r.segments[i+1].Pos1() - 3
		if v2 > v1 {
			draw2hlines(ug.Apply(klimt.UTranslateDx(v1)), height, klimt.ULineH(v2-v1))
		}
	}
}

func (r *rectangleTask) drawFull(widthCompletion float64, ug klimt.UGraphic, segment sequencediagram.Segment,
	height float64, documentBackground color.HColor) {
	width := segment.Length()
	ug = ug.Apply(klimt.UTranslateDx(segment.Pos1()))
	r.drawBackgroundRect(widthCompletion, ug.Apply(color.None()), documentBackground, width, height)
	ug.Apply(klimt.Bg(color.None())).Draw(klimt.URectangleBuild(width, height))
}

func (r *rectangleTask) drawPartly(widthCompletion float64, ug klimt.UGraphic, segment sequencediagram.Segment,
	height float64, documentBackground color.HColor, i int, withStartVline, withEndVline bool) {
	if withStartVline && withEndVline {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	widthBack := segment.Length()
	if i != len(r.segments)-1 {
		widthBack++
	}
	ug = ug.Apply(klimt.UTranslateDx(segment.Pos1()))
	if widthBack > 0 {
		r.drawBackgroundRect(widthCompletion, ug.Apply(color.None()), documentBackground, widthBack, height)
	}
	ug = ug.Apply(klimt.Bg(color.None()))
	switch {
	case withStartVline:
		ug.Draw(pathUtoLeft(segment.Length(), height, 0))
	case withEndVline:
		ug.Apply(klimt.UTranslateDx(segment.Length() - segment.Length())).Draw(pathUtoRight(segment.Length(), height, 0))
	default:
		draw2hlines(ug, height, klimt.ULineH(segment.Length()))
	}
}

func (r *rectangleTask) drawBackgroundRect(widthCompletion float64, ug klimt.UGraphic,
	documentBackground color.HColor, width, height float64) {
	if widthCompletion == -1 || widthCompletion == 0 {
		if widthCompletion == 0 {
			ug = ug.Apply(klimt.Bg(documentBackground))
		}
		ug.Draw(klimt.URectangleBuild(width, height))
		return
	}
	ug.Draw(klimt.URectangleBuild(widthCompletion, height))
	ug.Apply(klimt.Bg(documentBackground)).Apply(klimt.UTranslateDx(widthCompletion)).
		Draw(klimt.URectangleBuild(width-widthCompletion, height))
}
