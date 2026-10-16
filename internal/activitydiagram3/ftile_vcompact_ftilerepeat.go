// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileRepeat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// FtileRepeat is a "repeat" loop: the entry diamond (or start action), the
// body, the test diamond and the arrow back.
type FtileRepeat struct {
	AbstractFtile
	repeat   Ftile
	diamond1 Ftile
	diamond2 Ftile
	backward Ftile
	tbTest   klimt.TextBlock
}

// MyChildren implements Ftile.
func (f *FtileRepeat) MyChildren() []Ftile { return []Ftile{f.repeat, f.diamond1, f.diamond2} }

func newFtileRepeat(repeat, diamond1, diamond2 Ftile, tbTest klimt.TextBlock, backward Ftile) *FtileRepeat {
	f := &FtileRepeat{repeat: repeat, diamond1: diamond1, diamond2: diamond2, tbTest: tbTest, backward: backward}
	f.initFtile(f, repeat.SkinParam())
	return f
}

// SwimlaneIn implements Swimable.
func (f *FtileRepeat) SwimlaneIn() *Swimlane { return f.repeat.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileRepeat) SwimlaneOut() *Swimlane { return f.diamond2.SwimlaneOut() }

// Swimlanes implements Swimable.
func (f *FtileRepeat) Swimlanes() SwimlaneSet {
	result := SwimlaneSet(nil).addAll(f.repeat.Swimlanes()).addAll(f.diamond1.Swimlanes()).
		addAll(f.diamond2.Swimlanes())
	if f.backward != nil {
		result = result.addAll(f.backward.Swimlanes())
	}
	return result
}

// FtileRepeatCreate returns the tile of a "repeat" loop; entry (the start
// action), backward and endRepeatLinkColor may be nil.
func FtileRepeatCreate(swimlane, swimlaneOut *Swimlane, entry, repeat Ftile, test, yes, out *klimt.Display,
	borderColor, diamondColor1, diamondColor2 color.HColor, arrowColor, endRepeatLinkColor *Rainbow,
	conditionStyle klimt.ConditionStyle, spriteContainer klimt.ISkinSimple, fcDiamond, fcArrow *klimt.FontConfiguration,
	backward Ftile, noOut bool, incoming1, incoming2 *LinkRendering) Ftile {
	fontConfiguration1 := fcArrow
	if conditionStyle == klimt.ConditionStyleInsideHexagon {
		fontConfiguration1 = fcDiamond
	}
	var tbTest klimt.TextBlock
	if klimt.DisplayIsNull(test) || test.IsWhite() {
		tbTest = klimt.TextBlockUtilsEmpty(0, 0)
	} else {
		tbTest = test.Create(fontConfiguration1, repeat.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft),
			spriteContainer)
	}
	yesTb := yes.Create(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer)
	outTb := out.Create(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer)

	diamond1 := entry
	if entry == nil {
		diamond1 = NewFtileDiamond(repeat.SkinParam(), diamondColor1, borderColor, swimlane)
	}

	var result *FtileRepeat
	switch conditionStyle {
	case klimt.ConditionStyleInsideHexagon:
		var diamond2 Ftile
		switch {
		case noOut && klimt.DisplayIsNull(test):
			diamond2 = NewFtileEmpty(repeat.SkinParam(), 0, 0, nil)
		case ftileRepeatBackwardExitsOnLeft(backward, swimlaneOut):
			// The loop-back arrow exits towards the backward activity, on the
			// left when it lives in a lane left of the diamond: its label is
			// west of the diamond
			diamond2 = NewFtileDiamondInside(tbTest, repeat.SkinParam(), diamondColor2, borderColor, swimlaneOut).
				WithWest(yesTb).WithSouth(outTb)
		default:
			diamond2 = NewFtileDiamondInside(tbTest, repeat.SkinParam(), diamondColor2, borderColor, swimlaneOut).
				WithEast(yesTb).WithSouth(outTb)
		}
		result = newFtileRepeat(repeat, diamond1, diamond2, klimt.TextBlockUtilsEmpty(0, 0), backward)
	case klimt.ConditionStyleEmptyDiamond:
		diamond2 := NewFtileDiamond(repeat.SkinParam(), diamondColor2, borderColor, swimlane).WithEast(tbTest)
		result = newFtileRepeat(repeat, diamond1, diamond2, tbTest, backward)
	case klimt.ConditionStyleInsideDiamond:
		diamond2 := NewFtileDiamondSquare(tbTest, repeat.SkinParam(), diamondColor2, borderColor, swimlane).
			WithEast(yesTb).WithSouth(outTb)
		result = newFtileRepeat(repeat, diamond1, diamond2, klimt.TextBlockUtilsEmpty(0, 0), backward)
	default:
		uml.Throw(uml.IllegalStateException, "")
	}

	var conns []Connection
	in1 := repeat.InLinkRendering().Display()
	var tbin1 klimt.TextBlock
	if in1 != nil {
		tbin1 = in1.Create7(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer, klimt.CreoleModeSimpleLine)
	}
	conns = append(conns, result.newConnectionIn(repeat.InLinkRendering().RainbowOr(arrowColor), tbin1))

	var incomingText klimt.TextBlock
	if incoming1 != nil && incoming1.Display() != nil {
		incomingText = incoming1.Display().Create7(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer,
			klimt.CreoleModeSimpleLine)
	}

	switch {
	case backward != nil:
		rainbow1 := incoming1.RainbowOr(arrowColor)
		conns = append(conns, result.newConnectionBackBackward1(rainbow1, incomingText))
		var backArrowLabel klimt.TextBlock
		if incoming2 != nil && incoming2.Display() != nil {
			backArrowLabel = incoming2.Display().Create(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer)
		}
		rainbow2 := incoming2.RainbowOr(arrowColor)
		conns = append(conns, result.newConnectionBackBackward2(rainbow2, backArrowLabel))
	case swimlane == nil || swimlane == swimlaneOut:
		if swimlane != nil && swimlane.IsSmallerThanAllOthers(repeat.Swimlanes()) {
			conns = append(conns, result.newConnectionBackSimple1(incoming1.RainbowOr(arrowColor), incomingText))
		} else {
			conns = append(conns, result.newConnectionBackSimple2(incoming1.RainbowOr(arrowColor), incomingText))
		}
	default:
		conns = append(conns, result.newConnectionBackComplex1(incoming1.RainbowOr(arrowColor)))
	}

	out1 := repeat.OutLinkRendering().Display()
	var tbout1 klimt.TextBlock
	if out1 != nil {
		tbout1 = out1.Create7(fcArrow, geom.HorizontalAlignmentLeft, spriteContainer, klimt.CreoleModeSimpleLine)
	}
	if endRepeatLinkColor == nil {
		uml.ThrowNPE("")
	}
	tmpColor := endRepeatLinkColor.WithDefault(arrowColor)
	conns = append(conns, result.newConnectionOut(tmpColor, tbout1))
	return FtileUtilsAddConnections(result, conns)
}

// ftileRepeatBackwardExitsOnLeft reports whether the loop-back arrow leaves
// on the left: the backward activity lives in a lane ordered before the
// diamond's lane.
func ftileRepeatBackwardExitsOnLeft(backward Ftile, swimlaneOut *Swimlane) bool {
	if backward == nil {
		return false
	}
	swimlaneBackward := backward.SwimlaneIn()
	if swimlaneBackward == nil || swimlaneOut == nil {
		return false
	}
	return swimlaneBackward.CompareTo(swimlaneOut) < 0
}

// ftileRepeatConnectionIn is the arrow from the entry into the body.
type ftileRepeatConnectionIn struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	tbin       klimt.TextBlock
}

func (f *FtileRepeat) newConnectionIn(arrowColor *Rainbow, tbin klimt.TextBlock) *ftileRepeatConnectionIn {
	return &ftileRepeatConnectionIn{AbstractConnection: newAbstractConnection(f.diamond1, f.repeat), r: f,
		arrowColor: arrowColor, tbin: tbin}
}

func (c *ftileRepeatConnectionIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.Ftile1().Geometry(stringBounder).Translate(c.r.translateDiamond1(stringBounder)).PointOut()
}

func (c *ftileRepeatConnectionIn) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.Ftile2().Geometry(stringBounder).Translate(c.r.translateForRepeat(stringBounder)).PointIn()
}

func (c *ftileRepeatConnectionIn) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	c.drawSnake(ug, c.p1(stringBounder), c.p2(stringBounder))
}

func (c *ftileRepeatConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	c.drawSnake(ug, p1, p2)
}

func (c *ftileRepeatConnectionIn) drawSnake(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.tbin, c.ArrowHorizontalAlignment())
	snake.AddPointPt(p1)
	if p1.X != p2.X {
		my := (p1.Y + p2.Y) / 2
		snake.AddPointPt(geom.NewXPoint2D(p1.X, my))
		snake.AddPointPt(geom.NewXPoint2D(p2.X, my))
	}
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

// ftileRepeatConnectionOut is the arrow from the body to the test diamond.
type ftileRepeatConnectionOut struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	tbout      klimt.TextBlock
}

func (f *FtileRepeat) newConnectionOut(arrowColor *Rainbow, tbout klimt.TextBlock) *ftileRepeatConnectionOut {
	return &ftileRepeatConnectionOut{AbstractConnection: newAbstractConnection(f.repeat, f.diamond2), r: f,
		arrowColor: arrowColor, tbout: tbout}
}

func (c *ftileRepeatConnectionOut) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateForRepeat(stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileRepeatConnectionOut) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond2(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (c *ftileRepeatConnectionOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	if !c.Ftile1().Geometry(stringBounder).HasPointOut() {
		return
	}
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.tbout, c.ArrowHorizontalAlignment())
	snake.AddPointPt(c.p1(stringBounder))
	snake.AddPointPt(c.p2(stringBounder))
	ug.Draw(snake)
}

func (c *ftileRepeatConnectionOut) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	if !c.Ftile1().Geometry(stringBounder).HasPointOut() {
		return
	}
	skinParam := c.r.SkinParam()
	snake := SnakeCreate(skinParam, c.arrowColor)
	mp1a := translate1.Translated(c.p1(stringBounder))
	mp2b := translate2.Translated(c.p2(stringBounder))
	middle := (mp1a.Y + mp2b.Y) / 2.0
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	ug.Draw(snake)

	small := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.tbout, c.ArrowHorizontalAlignment())
	small.AddPoint(mp2b.X, middle)
	small.AddPointPt(mp2b)
	ug.Draw(small)
}

// ftileRepeatConnectionBackComplex1 is the arrow back when the test is in
// another lane than the start.
type ftileRepeatConnectionBackComplex1 struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
}

func (f *FtileRepeat) newConnectionBackComplex1(arrowColor *Rainbow) *ftileRepeatConnectionBackComplex1 {
	return &ftileRepeatConnectionBackComplex1{AbstractConnection: newAbstractConnection(f.diamond2, f.diamond1),
		r: f, arrowColor: arrowColor}
}

func (c *ftileRepeatConnectionBackComplex1) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackComplex1) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackComplex1) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	c.drawSnake(ug, c.p1(stringBounder), c.p2(stringBounder))
}

func (c *ftileRepeatConnectionBackComplex1) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	c.drawSnake(ug, p1, p2)
}

func (c *ftileRepeatConnectionBackComplex1) drawSnake(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	stringBounder := ug.StringBounder()
	dimRepeat := c.r.repeat.CalculateDimension(stringBounder)
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := c.r.diamond2.CalculateDimension(stringBounder)
	y1 := p1.Y + dimDiamond2.Height()/2
	x2 := p2.X + dimDiamond1.Width()
	y2 := p2.Y + dimDiamond1.Height()/2

	x1a := p1.X + dimDiamond2.Width()
	x1b := p1.X + dimDiamond2.Width()/2 + dimRepeat.Width()/2 + hexagonHalfSize

	skinParam := c.r.SkinParam()
	var snake *Snake
	if x2 < x1a {
		snake = SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToLeft()).
			EmphasizeDirection(klimt.DirectionUp)
		snake.AddPoint(x1a, y1)
		if x1a < x1b {
			snake.AddPoint(x1b, y1)
			snake.AddPoint(x1b, y2)
		} else {
			snake.AddPoint(x1a+10, y1)
			snake.AddPoint(x1a+10, y2)
		}
	} else {
		x2 = p2.X
		snake = SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToRight()).
			EmphasizeDirection(klimt.DirectionUp)
		snake.AddPoint(x1a, y1)
		threeQuarters := float64(x2*3) / 4
		middle := x1a/4 + threeQuarters
		snake.AddPoint(middle, y1)
		snake.AddPoint(middle, y2)
	}
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// ftileRepeatConnectionBackBackward1 is the arrow from the test diamond to
// the backward action.
type ftileRepeatConnectionBackBackward1 struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	tbback     klimt.TextBlock
}

func (f *FtileRepeat) newConnectionBackBackward1(arrowColor *Rainbow,
	tbback klimt.TextBlock) *ftileRepeatConnectionBackBackward1 {
	return &ftileRepeatConnectionBackBackward1{AbstractConnection: newAbstractConnection(f.diamond2, f.backward),
		r: f, arrowColor: arrowColor, tbback: tbback}
}

func (c *ftileRepeatConnectionBackBackward1) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackBackward1) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dim := c.r.backward.Geometry(stringBounder)
	return c.r.translateBackward(stringBounder).Translated(geom.NewXPoint2D(dim.Left(), dim.OutY()))
}

func (c *ftileRepeatConnectionBackBackward1) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	c.drawSnake(ug, c.p1(stringBounder), c.p2(stringBounder))
}

func (c *ftileRepeatConnectionBackBackward1) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	c.drawSnake(ug, p1, p2)
}

func (c *ftileRepeatConnectionBackBackward1) drawSnake(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	dimDiamond2 := c.r.diamond2.CalculateDimension(ug.StringBounder())
	x2 := p2.X
	y2 := p2.Y
	// The loop-back arrow leaves the diamond on the side of the backward
	// activity
	diamondCenterX := p1.X + dimDiamond2.Width()/2
	x1 := p1.X + dimDiamond2.Width()
	if x2 < diamondCenterX {
		x1 = p1.X
	}
	y1 := p1.Y + dimDiamond2.Height()/2

	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToUp()).
		WithLabel(c.tbback, c.ArrowHorizontalAlignment())
	snake.AddPoint(x1, y1)
	snake.AddPoint(x2, y1)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// ftileRepeatConnectionBackBackward2 is the arrow from the backward action
// to the entry.
type ftileRepeatConnectionBackBackward2 struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	label      klimt.TextBlock
}

func (f *FtileRepeat) newConnectionBackBackward2(arrowColor *Rainbow,
	label klimt.TextBlock) *ftileRepeatConnectionBackBackward2 {
	return &ftileRepeatConnectionBackBackward2{AbstractConnection: newAbstractConnection(f.backward, f.diamond1),
		r: f, label: label, arrowColor: arrowColor}
}

func (c *ftileRepeatConnectionBackBackward2) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dim := c.r.backward.Geometry(stringBounder)
	return c.r.translateBackward(stringBounder).Translated(geom.NewXPoint2D(dim.Left(), dim.InY()))
}

func (c *ftileRepeatConnectionBackBackward2) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackBackward2) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)

	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X
	if x2 < x1 {
		x2 += dimDiamond1.Width()
	}
	y2 := p2.Y + dimDiamond1.Height()/2

	skinParam := c.r.SkinParam()
	arrow := skinParam.Arrows().AsToRight()
	if x2 < x1 {
		arrow = skinParam.Arrows().AsToLeft()
	}
	snake := SnakeCreateEnd(skinParam, c.arrowColor, arrow)
	if c.label != nil {
		snake = snake.WithLabel(c.label, c.ArrowHorizontalAlignment())
	}
	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileRepeatConnectionBackBackward2) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToLeft())
	if c.label != nil {
		snake = snake.WithLabel(c.label, c.ArrowHorizontalAlignment())
	}
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X + dimDiamond1.Width()
	y2 := p2.Y + dimDiamond1.Height()/2

	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// ftileRepeatConnectionBackSimple1 is the arrow back on the left, when the
// loop is in the leftmost of its lanes.
type ftileRepeatConnectionBackSimple1 struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	tbback     klimt.TextBlock
}

func (f *FtileRepeat) newConnectionBackSimple1(arrowColor *Rainbow,
	tbback klimt.TextBlock) *ftileRepeatConnectionBackSimple1 {
	return &ftileRepeatConnectionBackSimple1{AbstractConnection: newAbstractConnection(f.diamond2, f.diamond1), r: f,
		arrowColor: arrowColor, tbback: tbback}
}

func (c *ftileRepeatConnectionBackSimple1) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackSimple1) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackSimple1) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToRight()).
		EmphasizeDirection(klimt.DirectionUp).WithLabel(c.tbback, c.ArrowHorizontalAlignment())
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := c.r.diamond2.CalculateDimension(stringBounder)
	x1 := p1.X
	y1 := p1.Y + dimDiamond2.Height()/2
	x2 := p2.X
	y2 := p2.Y + dimDiamond1.Height()/2

	snake.AddPoint(x1, y1)
	const xmin = -hexagonHalfSize
	snake.AddPoint(xmin, y1)
	snake.AddPoint(xmin, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileRepeatConnectionBackSimple1) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToLeft()).
		EmphasizeDirection(klimt.DirectionUp).WithLabel(c.tbback, c.ArrowHorizontalAlignment())
	dimRepeat := c.r.repeat.CalculateDimension(stringBounder)

	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := c.r.diamond2.CalculateDimension(stringBounder)
	x1 := p1.X
	y1 := p1.Y + dimDiamond2.Height()/2
	x2 := p2.X
	y2 := p2.Y + dimDiamond1.Height()/2

	snake.AddPoint(x1, y1)
	xmax := p1.X + dimDiamond2.Width()/2 + dimRepeat.Width()/2 + hexagonHalfSize
	snake.AddPoint(xmax, y1)
	snake.AddPoint(xmax, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// ftileRepeatConnectionBackSimple2 is the arrow back on the right.
type ftileRepeatConnectionBackSimple2 struct {
	AbstractConnection
	r          *FtileRepeat
	arrowColor *Rainbow
	tbback     klimt.TextBlock
}

func (f *FtileRepeat) newConnectionBackSimple2(arrowColor *Rainbow,
	tbback klimt.TextBlock) *ftileRepeatConnectionBackSimple2 {
	return &ftileRepeatConnectionBackSimple2{AbstractConnection: newAbstractConnection(f.diamond2, f.diamond1), r: f,
		arrowColor: arrowColor, tbback: tbback}
}

func (c *ftileRepeatConnectionBackSimple2) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackSimple2) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.r.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileRepeatConnectionBackSimple2) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.r.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToLeft()).
		EmphasizeDirection(klimt.DirectionUp).WithLabel(c.tbback, c.ArrowHorizontalAlignment())
	dimTotal := c.r.calculateDimensionInternal(stringBounder)
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := c.r.diamond2.CalculateDimension(stringBounder)
	x1 := p1.X + dimDiamond2.Width()
	y1 := p1.Y + dimDiamond2.Height()/2
	x2 := p2.X + dimDiamond1.Width()
	y2 := p2.Y + dimDiamond1.Height()/2

	snake.AddPoint(x1, y1)
	xmax := dimTotal.Width() - hexagonHalfSize
	snake.AddPoint(xmax, y1)
	snake.AddPoint(xmax, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileRepeatConnectionBackSimple2) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := translate1.Translated(c.p1(stringBounder))
	p2 := translate2.Translated(c.p2(stringBounder))
	dimDiamond1 := c.r.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := c.r.diamond2.CalculateDimension(stringBounder)
	x1 := p1.X + dimDiamond2.Width()
	y1 := p1.Y + dimDiamond2.Height()/2

	x2a := p2.X
	x2b := p2.X + dimDiamond1.Width()
	isOnA := x1 < (x2a+x2b)/2

	x2 := x2b
	if isOnA {
		x2 = x2a
	}
	y2 := p2.Y + dimDiamond1.Height()/2

	skinParam := c.r.SkinParam()
	arrow := skinParam.Arrows().AsToLeft()
	if isOnA {
		arrow = skinParam.Arrows().AsToRight()
	}
	snake := SnakeCreateEnd(skinParam, c.arrowColor, arrow).EmphasizeDirection(klimt.DirectionUp).
		WithLabel(c.tbback, c.ArrowHorizontalAlignment())

	snake.AddPoint(x1, y1)
	xmiddle := (x1 + x2) / 2
	snake.AddPoint(xmiddle, y1)
	snake.AddPoint(xmiddle, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// DrawU implements klimt.TextBlock.
func (f *FtileRepeat) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug.Apply(f.translateForRepeat(stringBounder)).Draw(f.repeat)
	ug.Apply(f.translateDiamond1(stringBounder)).Draw(f.diamond1)
	ug.Apply(f.translateDiamond2(stringBounder)).Draw(f.diamond2)
	if f.backward != nil {
		ug.Apply(f.translateBackward(stringBounder)).Draw(f.backward)
	}
}

func (f *FtileRepeat) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	return NewFtileGeometryDimOut(dimTotal, f.left(stringBounder), 0, dimTotal.Height())
}

func (f *FtileRepeat) calculateDimensionInternal(stringBounder font.StringBounder) geom.XDimension2D {
	dimDiamond1 := f.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := f.diamond2.CalculateDimension(stringBounder)
	dimRepeat := f.repeat.CalculateDimension(stringBounder)

	w := f.tbTest.CalculateDimension(stringBounder).Width()

	width := f.left(stringBounder) + f.right(stringBounder)
	width = math.Max(width, w+2*hexagonHalfSize)
	if f.backward != nil {
		width += f.backward.CalculateDimension(stringBounder).Width()
	}
	height := dimDiamond1.Height() + dimRepeat.Height() + dimDiamond2.Height() + 8*hexagonHalfSize
	return geom.NewXDimension2D(width+2*hexagonHalfSize, height)
}

// TranslateFor implements Ftile.
func (f *FtileRepeat) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == f.repeat {
		t := f.translateForRepeat(stringBounder)
		return &t
	}
	if child == f.diamond1 {
		t := f.translateDiamond1(stringBounder)
		return &t
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (f *FtileRepeat) translateForRepeat(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := f.diamond2.CalculateDimension(stringBounder)
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimRepeat := f.repeat.CalculateDimension(stringBounder)
	space := dimTotal.Height() - dimDiamond1.Height() - dimDiamond2.Height() - dimRepeat.Height()
	y := dimDiamond1.Height() + space/2
	left := f.left(stringBounder)
	return klimt.NewUTranslate(left-f.repeat.Geometry(stringBounder).Left(), y)
}

func (f *FtileRepeat) translateDiamond1(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.CalculateDimension(stringBounder)
	left := f.left(stringBounder)
	return klimt.UTranslateDx(left - dimDiamond1.Width()/2)
}

func (f *FtileRepeat) translateBackward(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimBackward := f.backward.CalculateDimension(stringBounder)
	x := dimTotal.Width() - dimBackward.Width()
	y := (dimTotal.Height() - dimBackward.Height()) / 2
	return klimt.NewUTranslate(x, y)
}

func (f *FtileRepeat) translateDiamond2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimDiamond2 := f.diamond2.CalculateDimension(stringBounder)
	y2 := dimTotal.Height() - dimDiamond2.Height()
	left := f.left(stringBounder)
	return klimt.NewUTranslate(left-dimDiamond2.Width()/2, y2)
}

func (f *FtileRepeat) left(stringBounder font.StringBounder) float64 {
	dimDiamond1 := f.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := f.diamond2.CalculateDimension(stringBounder)
	left1 := f.repeat.Geometry(stringBounder).Left()
	left1 = math.Max(left1, dimDiamond1.Width()/2)
	left2 := f.repeat.Geometry(stringBounder).Left()
	left2 = math.Max(left2, dimDiamond2.Width()/2)
	return math.Max(left1, left2)
}

func (f *FtileRepeat) right(stringBounder font.StringBounder) float64 {
	dimDiamond1 := f.diamond1.CalculateDimension(stringBounder)
	dimDiamond2 := f.diamond2.CalculateDimension(stringBounder)
	dimRepeat := f.repeat.CalculateDimension(stringBounder)
	right1 := dimRepeat.Width() - f.repeat.Geometry(stringBounder).Left()
	right1 = math.Max(right1, dimDiamond1.Width()/2)
	right2 := dimRepeat.Width() - f.repeat.Geometry(stringBounder).Left()
	right2 = math.Max(right2, dimDiamond2.Width()/2)
	return math.Max(right1, right2)
}
