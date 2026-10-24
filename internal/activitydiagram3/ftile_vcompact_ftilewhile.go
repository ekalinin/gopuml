// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileWhile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// FtileWhile is a "while" loop: the test diamond above the body, the arrow
// back on the right and the exit on the left.
type FtileWhile struct {
	AbstractFtile
	whileBlock Ftile
	diamond1   Ftile
	specialOut Ftile
	backward   Ftile
	back1      klimt.TextBlock
}

// MyChildren implements Ftile.
func (f *FtileWhile) MyChildren() []Ftile {
	if f.specialOut == nil {
		return []Ftile{f.whileBlock, f.diamond1}
	}
	return []Ftile{f.whileBlock, f.diamond1, f.specialOut}
}

// Swimlanes implements Swimable.
func (f *FtileWhile) Swimlanes() SwimlaneSet {
	return SwimlaneSet(nil).addAll(f.whileBlock.Swimlanes()).add(f.SwimlaneIn())
}

// SwimlaneIn implements Swimable.
func (f *FtileWhile) SwimlaneIn() *Swimlane { return f.diamond1.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileWhile) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

func newFtileWhile(whileBlock, diamond1, specialOut, backward Ftile) *FtileWhile {
	f := &FtileWhile{whileBlock: whileBlock, diamond1: diamond1, specialOut: specialOut, backward: backward}
	f.initFtile(f, whileBlock.SkinParam())
	return f
}

// FtileWhileCreate returns the tile of a "while" loop; specialOut (the
// stop or end after the loop) and backward may be nil.
func FtileWhileCreate(outColor *LinkRendering, swimlane *Swimlane, whileBlock Ftile, test *klimt.Display,
	borderColor, backColor color.HColor, arrowColor *Rainbow, yes *klimt.Display, fontArrow *klimt.FontConfiguration,
	ftileFactory FtileFactory, conditionStyle klimt.ConditionStyle, fcTest *klimt.FontConfiguration,
	specialOut Instruction, backward Ftile, incoming1, incoming2 *LinkRendering) Ftile {
	yesTb := yes.Create(fontArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
	var testTb klimt.TextBlock
	if test.IsWhite() {
		testTb = klimt.TextBlockUtilsEmpty(0, 0)
	} else {
		testTb = test.Create(fcTest, whileBlock.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft),
			ftileFactory.SkinParam())
	}
	outTb := outColor.Display().Create(fontArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())

	var diamond1 Ftile
	switch conditionStyle {
	case klimt.ConditionStyleInsideHexagon:
		diamond1 = NewFtileDiamondInside(testTb, whileBlock.SkinParam(), backColor, borderColor, swimlane).
			WithNorth(yesTb).WithWest(outTb)
	case klimt.ConditionStyleInsideDiamond:
		diamond1 = NewFtileDiamondSquare(testTb, whileBlock.SkinParam(), backColor, borderColor, swimlane).
			WithNorth(yesTb).WithWest(outTb)
	case klimt.ConditionStyleEmptyDiamond:
		diamond1 = NewFtileDiamond(whileBlock.SkinParam(), backColor, borderColor, swimlane).WithNorth(testTb).
			WithSouth(yesTb).WithWest(outTb)
	default:
		uml.Throw(uml.IllegalStateException, "")
	}

	var special Ftile
	if specialOut != nil {
		special = specialOut.CreateFtile(ftileFactory)
	}
	result := newFtileWhile(whileBlock, diamond1, special, backward)

	dim := whileBlock.CalculateDimension(ftileFactory.StringBounder())
	result.back1 = incoming1.Display().Create(fontArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())

	var conns []Connection
	if dim.Width() == 0 || dim.Height() == 0 {
		conns = append(conns, result.newConnectionBackEmpty(incoming1.Rainbow()))
	} else {
		conns = append(conns, result.newConnectionIn(whileBlock.InLinkRendering().RainbowOr(arrowColor)))
		if backward == nil {
			conns = append(conns, result.newConnectionBackSimple(incoming1.Rainbow(), result.back1))
		} else {
			back2 := incoming2.Display().Create(fontArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
			conns = append(conns, result.newConnectionBackBackward1(incoming1.Rainbow(), result.back1))
			conns = append(conns, result.newConnectionBackBackward2(incoming2.Rainbow(), back2))
		}
	}
	if specialOut == nil {
		conns = append(conns, result.newConnectionOut(outColor.Rainbow()))
	} else {
		conns = append(conns, result.newConnectionOutSpecial(outColor.Rainbow()))
	}
	return FtileUtilsAddConnections(result, conns)
}

// ftileWhileConnectionIn is the arrow from the diamond into the body.
type ftileWhileConnectionIn struct {
	AbstractConnection
	w          *FtileWhile
	arrowColor *Rainbow
}

func (f *FtileWhile) newConnectionIn(arrowColor *Rainbow) *ftileWhileConnectionIn {
	return &ftileWhileConnectionIn{AbstractConnection: newAbstractConnection(f.diamond1, f.whileBlock), w: f,
		arrowColor: arrowColor}
}

func (c *ftileWhileConnectionIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileWhileConnectionIn) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateForWhile(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (c *ftileWhileConnectionIn) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	snake.AddPointPt(c.p1(stringBounder))
	snake.AddPointPt(c.p2(stringBounder))
	ug.Draw(snake)
}

func (c *ftileWhileConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).WithMerge(MergeLimited)
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := (mp1a.Y + mp2b.Y) / 2.0
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

// ftileWhileConnectionBackSimple is the arrow from the end of the body back
// to the diamond.
type ftileWhileConnectionBackSimple struct {
	AbstractConnection
	w              *FtileWhile
	endInlinkColor *Rainbow
	back           klimt.TextBlock
}

func (f *FtileWhile) newConnectionBackSimple(endInlinkColor *Rainbow,
	back klimt.TextBlock) *ftileWhileConnectionBackSimple {
	return &ftileWhileConnectionBackSimple{AbstractConnection: newAbstractConnection(f.whileBlock, f.diamond1), w: f,
		endInlinkColor: endInlinkColor, back: back}
}

func (c *ftileWhileConnectionBackSimple) p1(stringBounder font.StringBounder) (geom.XPoint2D, bool) {
	g := c.w.whileBlock.Geometry(stringBounder)
	if !g.HasPointOut() {
		return geom.XPoint2D{}, false
	}
	return c.w.translateForWhile(stringBounder).Translated(g.PointOut()), true
}

func (c *ftileWhileConnectionBackSimple) bottom(stringBounder font.StringBounder) float64 {
	g := c.w.whileBlock.Geometry(stringBounder)
	return c.w.translateForWhile(stringBounder).Dy() + g.Height()
}

func (c *ftileWhileConnectionBackSimple) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileWhileConnectionBackSimple) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := c.w.CalculateDimension(stringBounder)
	p1, ok := c.p1(stringBounder)
	if !ok {
		return
	}
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)

	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X + dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y2 := p2.Y + dimDiamond1.InY() + half

	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.endInlinkColor, skinParam.Arrows().AsToLeft()).
		EmphasizeDirection(klimt.DirectionUp).WithLabelVertical(c.back, geom.VerticalAlignmentBottom)
	snake.AddPoint(x1, y1)
	y1bis := math.Max(y1, c.bottom(stringBounder)) + hexagonHalfSize
	snake.AddPoint(x1, y1bis)
	xx := dimTotal.Width()
	snake.AddPoint(xx, y1bis)
	snake.AddPoint(xx, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
	ug.Apply(klimt.NewUTranslate(x1, y1bis)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
}

func (c *ftileWhileConnectionBackSimple) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.endInlinkColor, skinParam.Arrows().AsToLeft()).WithMerge(MergeLimited)
	dimTotal := c.w.CalculateDimension(stringBounder)
	ap1, ok := c.p1(stringBounder)
	if !ok {
		// upstream translates a null point
		uml.ThrowNPE("")
	}
	ap2 := c.p2(stringBounder)
	p1 := translate1.Translated(ap1)
	p2 := translate2.Translated(ap2)

	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)

	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X + dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y2 := p2.Y + dimDiamond1.InY() + half

	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y1+hexagonHalfSize)
	xx := math.Max(translate1.Dx(), translate2.Dx()) + dimTotal.Width()
	snake.AddPoint(xx, y1+hexagonHalfSize)
	snake.AddPoint(xx, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)

	ug.Apply(klimt.NewUTranslate(x1, y1+hexagonHalfSize)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))

	ug = ug.Apply(c.endInlinkColor.Color()).Apply(klimt.Bg(c.endInlinkColor.Color()))
	ug.Apply(klimt.NewUTranslate(xx, (y1+y2)/2)).Draw(skinParam.Arrows().AsToUp())
}

// ftileWhileConnectionBackBackward1 is the arrow from the end of the body
// to the backward action.
type ftileWhileConnectionBackBackward1 struct {
	AbstractConnection
	w              *FtileWhile
	endInlinkColor *Rainbow
	back           klimt.TextBlock
}

func (f *FtileWhile) newConnectionBackBackward1(endInlinkColor *Rainbow,
	back klimt.TextBlock) *ftileWhileConnectionBackBackward1 {
	return &ftileWhileConnectionBackBackward1{AbstractConnection: newAbstractConnection(f.whileBlock, f.backward),
		w: f, endInlinkColor: endInlinkColor, back: back}
}

func (c *ftileWhileConnectionBackBackward1) p1(stringBounder font.StringBounder) (geom.XPoint2D, bool) {
	g := c.w.whileBlock.Geometry(stringBounder)
	if !g.HasPointOut() {
		return geom.XPoint2D{}, false
	}
	return c.w.translateForWhile(stringBounder).Translated(g.PointOut()), true
}

func (c *ftileWhileConnectionBackBackward1) bottom(stringBounder font.StringBounder) float64 {
	g := c.w.whileBlock.Geometry(stringBounder)
	return c.w.translateForWhile(stringBounder).Dy() + g.Height()
}

func (c *ftileWhileConnectionBackBackward1) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dim := c.w.backward.Geometry(stringBounder)
	return c.w.translateBackward(stringBounder).Translated(geom.NewXPoint2D(dim.Left(), dim.OutY()))
}

func (c *ftileWhileConnectionBackBackward1) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1, ok := c.p1(stringBounder)
	if !ok {
		return
	}
	p2 := c.p2(stringBounder)
	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y

	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.endInlinkColor, skinParam.Arrows().AsToUp()).
		WithLabelVertical(c.back, geom.VerticalAlignmentBottom)
	snake.AddPoint(x1, y1)
	y1bis := math.Max(y1, c.bottom(stringBounder)) + hexagonHalfSize
	snake.AddPoint(x1, y1bis)
	snake.AddPoint(x2, y1bis)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
	ug.Apply(klimt.NewUTranslate(x1, y1bis)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
}

// ftileWhileConnectionBackBackward2 is the arrow from the backward action
// to the diamond.
type ftileWhileConnectionBackBackward2 struct {
	AbstractConnection
	w              *FtileWhile
	endInlinkColor *Rainbow
	back           klimt.TextBlock
}

func (f *FtileWhile) newConnectionBackBackward2(endInlinkColor *Rainbow,
	back klimt.TextBlock) *ftileWhileConnectionBackBackward2 {
	return &ftileWhileConnectionBackBackward2{AbstractConnection: newAbstractConnection(f.backward, f.diamond1),
		w: f, endInlinkColor: endInlinkColor, back: back}
}

func (c *ftileWhileConnectionBackBackward2) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dim := c.w.backward.Geometry(stringBounder)
	return c.w.translateBackward(stringBounder).Translated(geom.NewXPoint2D(dim.Left(), dim.InY()))
}

func (c *ftileWhileConnectionBackBackward2) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileWhileConnectionBackBackward2) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.endInlinkColor, skinParam.Arrows().AsToLeft()).
		WithLabel(c.back, c.ArrowHorizontalAlignment())

	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)

	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X + dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y2 := p2.Y + dimDiamond1.InY() + half

	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// ftileWhileConnectionBackEmpty is the arrow back to the diamond of a loop
// with an empty body.
type ftileWhileConnectionBackEmpty struct {
	AbstractConnection
	w              *FtileWhile
	endInlinkColor *Rainbow
}

func (f *FtileWhile) newConnectionBackEmpty(endInlinkColor *Rainbow) *ftileWhileConnectionBackEmpty {
	return &ftileWhileConnectionBackEmpty{AbstractConnection: newAbstractConnection(f.diamond1, f.diamond1), w: f,
		endInlinkColor: endInlinkColor}
}

func (c *ftileWhileConnectionBackEmpty) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(c.w.diamond1.Geometry(stringBounder).PointOut())
}

func (c *ftileWhileConnectionBackEmpty) bottom(stringBounder font.StringBounder) float64 {
	g := c.w.whileBlock.Geometry(stringBounder)
	return c.w.translateForWhile(stringBounder).Dy() + g.Height()
}

func (c *ftileWhileConnectionBackEmpty) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileWhileConnectionBackEmpty) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.endInlinkColor, skinParam.Arrows().AsToLeft()).
		EmphasizeDirection(klimt.DirectionUp)
	dimTotal := c.w.CalculateDimension(stringBounder)
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)

	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X + dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y2 := p2.Y + dimDiamond1.InY() + half

	snake.AddPoint(x1, y1)
	y1bis := math.Max(y1, c.bottom(stringBounder)) + hexagonHalfSize
	snake.AddPoint(x1, y1bis)
	xx := dimTotal.Width()
	snake.AddPoint(xx, y1bis)
	snake.AddPoint(xx, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)

	ug.Apply(klimt.NewUTranslate(x1, y1bis)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
}

// ftileWhileConnectionOut is the exit arrow, from the left of the diamond
// down to the bottom of the loop.
type ftileWhileConnectionOut struct {
	AbstractConnection
	w                  *FtileWhile
	afterEndwhileColor *Rainbow
}

func (f *FtileWhile) newConnectionOut(afterEndwhileColor *Rainbow) *ftileWhileConnectionOut {
	return &ftileWhileConnectionOut{AbstractConnection: newAbstractConnection(f.diamond1, nil), w: f,
		afterEndwhileColor: afterEndwhileColor}
}

func (c *ftileWhileConnectionOut) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileWhileConnectionOut) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimTotal := c.w.Geometry(stringBounder)
	return geom.NewXPoint2D(dimTotal.Left(), dimTotal.Height())
}

func (c *ftileWhileConnectionOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreate(skinParam, c.afterEndwhileColor).WithMerge(MergeLimited).
		EmphasizeDirection(klimt.DirectionDown)

	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)

	x1 := p1.X
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y1 := p1.Y + dimDiamond1.InY() + half
	x2 := p2.X
	y2 := p2.Y

	snake.AddPoint(x1, y1)
	snake.AddPoint(hexagonHalfSize, y1)
	snake.AddPoint(hexagonHalfSize, y2)
	ug.Draw(snake)

	snake2 := SnakeCreate(skinParam, c.afterEndwhileColor)
	snake2.AddPoint(hexagonHalfSize, y2)
	snake2.AddPoint(x2, y2)
	ug.Draw(snake2)
}

// ftileWhileConnectionOutSpecial is the exit arrow to the stop or end that
// follows the loop.
type ftileWhileConnectionOutSpecial struct {
	AbstractConnection
	w                  *FtileWhile
	afterEndwhileColor *Rainbow
}

func (f *FtileWhile) newConnectionOutSpecial(afterEndwhileColor *Rainbow) *ftileWhileConnectionOutSpecial {
	return &ftileWhileConnectionOutSpecial{AbstractConnection: newAbstractConnection(f.diamond1, f.specialOut),
		w: f, afterEndwhileColor: afterEndwhileColor}
}

func (c *ftileWhileConnectionOutSpecial) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, 0))
}

func (c *ftileWhileConnectionOutSpecial) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.w.translateForSpecial(stringBounder).Translated(c.w.specialOut.Geometry(stringBounder).PointIn())
}

func (c *ftileWhileConnectionOutSpecial) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.w.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.afterEndwhileColor, skinParam.Arrows().AsToDown())

	dimDiamond1 := c.w.diamond1.Geometry(stringBounder)
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)

	x1 := p1.X
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y1 := p1.Y + dimDiamond1.InY() + half
	x2 := p2.X
	y2 := p2.Y

	snake.AddPoint(x1, y1)
	snake.AddPoint(x2, y1)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

// DrawU implements klimt.TextBlock.
func (f *FtileWhile) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug.Apply(f.translateForWhile(stringBounder)).Draw(f.whileBlock)
	ug.Apply(f.translateDiamond1(stringBounder)).Draw(f.diamond1)
	if f.specialOut != nil {
		ug.Apply(f.translateForSpecial(stringBounder)).Draw(f.specialOut)
	}
	if f.backward != nil {
		ug.Apply(f.translateBackward(stringBounder)).Draw(f.backward)
	}
}

func (f *FtileWhile) translateBackward(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionFtile(stringBounder)
	dimBackward := f.backward.CalculateDimension(stringBounder)
	x := dimTotal.Width() - dimBackward.Width()
	y := (dimTotal.Height() - dimBackward.Height()) / 2
	return klimt.NewUTranslate(x, y)
}

func (f *FtileWhile) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	geoDiamond1 := f.diamond1.Geometry(stringBounder)
	geoWhile := f.whileBlock.Geometry(stringBounder)
	if diff := -geoWhile.Width(); diff > 0 {
		geoWhile = geoWhile.AddMarginX(diff / 2)
	}
	g := geoDiamond1.AppendBottom(geoWhile)
	height := g.Height() + 4*hexagonHalfSize + f.suppHeightForLabel(stringBounder)
	const dx = 2 * hexagonHalfSize
	backwardWidth := 0.0
	if f.backward != nil {
		backwardWidth += f.backward.CalculateDimension(stringBounder).Width()
	}
	return NewFtileGeometryOut(f.xDeltaBecauseSpecial(stringBounder)+g.Width()+dx+hexagonHalfSize+backwardWidth,
		height, f.xDeltaBecauseSpecial(stringBounder)+g.Left()+dx, geoDiamond1.InY(), height)
}

func (f *FtileWhile) suppHeightForLabel(stringBounder font.StringBounder) float64 {
	if f.back1 != nil {
		return f.back1.CalculateDimension(stringBounder).Height()
	}
	return 0
}

func (f *FtileWhile) xDeltaBecauseSpecial(stringBounder font.StringBounder) float64 {
	if f.specialOut == nil {
		return 0
	}
	return f.specialOut.CalculateDimension(stringBounder).Width()
}

// TranslateFor implements Ftile.
func (f *FtileWhile) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == f.whileBlock {
		t := f.translateForWhile(stringBounder)
		return &t
	}
	if child == f.diamond1 {
		t := f.translateDiamond1(stringBounder)
		return &t
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (f *FtileWhile) translateForWhile(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	dimTotal := f.Geometry(stringBounder)
	dimWhile := f.whileBlock.Geometry(stringBounder)
	y := dimDiamond1.Height() + (dimTotal.Height()-dimDiamond1.Height()-dimWhile.Height()-
		f.suppHeightForLabel(stringBounder))/2
	x := dimTotal.Left() - dimWhile.Left()
	return klimt.NewUTranslate(x, y)
}

func (f *FtileWhile) translateDiamond1(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.Geometry(stringBounder)
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	const y1 = 0.0
	x1 := dimTotal.Left() - dimDiamond1.Left()
	return klimt.NewUTranslate(x1, y1)
}

func (f *FtileWhile) translateForSpecial(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	y1 := math.Max(3*half, 4*hexagonHalfSize)
	xWhile := f.translateForWhile(stringBounder).Dx() - hexagonHalfSize
	xDiamond := f.translateDiamond1(stringBounder).Dx()
	x1 := math.Min(xWhile, xDiamond) - f.xDeltaBecauseSpecial(stringBounder)
	return klimt.NewUTranslate(x1, y1)
}
