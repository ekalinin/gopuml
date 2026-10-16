// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileIfDown.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// FtileIfDown is an "if" whose only branch goes down below the diamond;
// the other branch is empty, or a stop drawn on the right.
type FtileIfDown struct {
	AbstractFtile
	thenBlock         Ftile
	diamond1          Ftile
	diamond2          Ftile
	optionalStop      Ftile
	conditionEndStyle klimt.ConditionEndStyle
	opale             klimt.TextBlock
}

// MyChildren implements Ftile.
func (f *FtileIfDown) MyChildren() []Ftile {
	if f.optionalStop == nil {
		return []Ftile{f.thenBlock, f.diamond1, f.diamond2}
	}
	return []Ftile{f.thenBlock, f.diamond1, f.diamond2, f.optionalStop}
}

// Swimlanes implements Swimable.
func (f *FtileIfDown) Swimlanes() SwimlaneSet {
	return SwimlaneSet(nil).addAll(f.thenBlock.Swimlanes()).add(f.SwimlaneIn())
}

// SwimlaneIn implements Swimable.
func (f *FtileIfDown) SwimlaneIn() *Swimlane { return f.diamond1.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileIfDown) SwimlaneOut() *Swimlane {
	if f.optionalStop == nil {
		return f.SwimlaneIn()
	}
	return f.thenBlock.SwimlaneOut()
}

func newFtileIfDown(thenBlock, diamond1, diamond2, optionalStop Ftile, conditionEndStyle klimt.ConditionEndStyle,
	notes []*PositionedNote) *FtileIfDown {
	f := &FtileIfDown{thenBlock: thenBlock, diamond1: diamond1, diamond2: diamond2, optionalStop: optionalStop,
		conditionEndStyle: conditionEndStyle}
	f.initFtile(f, thenBlock.SkinParam())
	if len(notes) == 1 {
		f.opale = FtileIfWithDiamondsCreateOpale(notes[0], f.SkinParam())
	} else {
		f.opale = klimt.TextBlockUtilsEmptyTextBlock()
	}
	return f
}

// FtileIfDownCreate returns the "if" of thenBlock below diamond1;
// optionalStop and elseColor may be nil.
func FtileIfDownCreate(diamond1, diamond2 Ftile, swimlane *Swimlane, thenBlock Ftile, arrowColor *Rainbow,
	conditionEndStyle klimt.ConditionEndStyle, ftileFactory FtileFactory, optionalStop Ftile, elseColor *Rainbow,
	notes []*PositionedNote) Ftile {
	if elseColor == nil {
		uml.ThrowNPE("")
	}
	elseColor = elseColor.WithDefault(arrowColor)
	d2 := diamond2
	if optionalStop != nil {
		d2 = NewFtileEmpty(ftileFactory.SkinParam(), 0, 0, nil)
	}
	result := newFtileIfDown(thenBlock, diamond1, d2, optionalStop, conditionEndStyle, notes)

	conns := []Connection{result.newConnectionIn(thenBlock.InLinkRendering().RainbowOr(arrowColor))}
	hasPointOut1 := thenBlock.Geometry(ftileFactory.StringBounder()).HasPointOut()
	if optionalStop == nil {
		if hasPointOut1 {
			switch conditionEndStyle {
			case klimt.ConditionEndStyleDiamond:
				if swimlane != nil && swimlane.IsSmallerThanAllOthers(thenBlock.Swimlanes()) {
					conns = append(conns, result.newConnectionElse1(elseColor))
					if d, ok := diamond1.(*FtileDiamondInside); ok {
						d.SwapEastWest()
					}
				} else {
					conns = append(conns, result.newConnectionElse2(elseColor, ftileIfDownElse2Normal))
				}
			case klimt.ConditionEndStyleHline:
				conns = append(conns, result.newConnectionElse2(elseColor, ftileIfDownElse2Hline),
					&ftileIfDownConnectionHline{AbstractConnection: newAbstractConnection(diamond1, d2), f: result,
						endInlinkColor: elseColor})
			}
		} else {
			conns = append(conns, result.newConnectionElse2(elseColor, ftileIfDownElse2NoDiamond))
		}
	} else {
		conns = append(conns, &ftileIfDownConnectionHorizontal{
			AbstractConnection: newAbstractConnection(diamond1, optionalStop), f: result, color: elseColor})
	}
	conns = append(conns, result.newConnectionOut(thenBlock.OutLinkRendering().RainbowOr(arrowColor)))
	return FtileUtilsAddConnections(result, conns)
}

type ftileIfDownConnectionHorizontal struct {
	AbstractConnection
	f     *FtileIfDown
	color *Rainbow
}

func (c *ftileIfDownConnectionHorizontal) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToRight())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfDownConnectionHorizontal) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.Ftile1().Geometry(stringBounder)
	p := geom.NewXPoint2D(dimDiamond1.Width(), (dimDiamond1.InY()+dimDiamond1.OutY())/2)
	return c.f.translateDiamond1(stringBounder).Translated(p)
}

func (c *ftileIfDownConnectionHorizontal) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimStop := c.Ftile2().CalculateDimension(stringBounder)
	p := geom.NewXPoint2D(0, dimStop.Height()/2)
	return c.f.translateOptionalStop(stringBounder).Translated(p)
}

type ftileIfDownConnectionIn struct {
	AbstractConnection
	f          *FtileIfDown
	arrowColor *Rainbow
}

func (f *FtileIfDown) newConnectionIn(arrowColor *Rainbow) *ftileIfDownConnectionIn {
	return &ftileIfDownConnectionIn{AbstractConnection: newAbstractConnection(f.diamond1, f.thenBlock), f: f,
		arrowColor: arrowColor}
}

func (c *ftileIfDownConnectionIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.f.translateDiamond1(stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileIfDownConnectionIn) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.f.translateForThen(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (c *ftileIfDownConnectionIn) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(c.p1(stringBounder))
	snake.AddPointPt(c.p2(stringBounder))
	ug.Draw(snake)
}

func (c *ftileIfDownConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := (mp1a.Y + mp2b.Y) / 2.0
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

type ftileIfDownConnectionOut struct {
	AbstractConnection
	f          *FtileIfDown
	arrowColor *Rainbow
}

func (f *FtileIfDown) newConnectionOut(arrowColor *Rainbow) *ftileIfDownConnectionOut {
	return &ftileIfDownConnectionOut{AbstractConnection: newAbstractConnection(f.thenBlock, f.diamond2), f: f,
		arrowColor: arrowColor}
}

func (c *ftileIfDownConnectionOut) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.f.translateForThen(stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileIfDownConnectionOut) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.f.translateDiamond2(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (c *ftileIfDownConnectionOut) p2hline(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.Ftile2().Geometry(stringBounder)
	x := dimDiamond2.Width()
	half := (dimDiamond2.OutY() - dimDiamond2.InY()) / 2
	return c.f.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond2.InY()+half))
}

func (c *ftileIfDownConnectionOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	if !c.Ftile1().Geometry(stringBounder).HasPointOut() {
		return
	}
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(c.p1(stringBounder))
	switch c.f.conditionEndStyle {
	case klimt.ConditionEndStyleDiamond:
		snake.AddPointPt(c.p2(stringBounder))
	case klimt.ConditionEndStyleHline:
		snake.AddPointPt(c.p2hline(stringBounder))
	}
	ug.Draw(snake)
}

func (c *ftileIfDownConnectionOut) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	if !c.Ftile1().Geometry(stringBounder).HasPointOut() {
		return
	}
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := (mp1a.Y + mp2b.Y) / 2.0
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

type ftileIfDownConnectionElse1 struct {
	AbstractConnection
	f              *FtileIfDown
	endInlinkColor *Rainbow
}

func (f *FtileIfDown) newConnectionElse1(endInlinkColor *Rainbow) *ftileIfDownConnectionElse1 {
	return &ftileIfDownConnectionElse1{AbstractConnection: newAbstractConnection(f.diamond1, f.diamond2), f: f,
		endInlinkColor: endInlinkColor}
}

func (c *ftileIfDownConnectionElse1) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.f.diamond1.Geometry(stringBounder)
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	return c.f.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(0, dimDiamond1.InY()+half))
}

func (c *ftileIfDownConnectionElse1) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.f.diamond2.Geometry(stringBounder)
	half := (dimDiamond2.OutY() - dimDiamond2.InY()) / 2
	return c.f.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(0, dimDiamond2.InY()+half))
}

func (c *ftileIfDownConnectionElse1) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	if !c.f.Geometry(stringBounder).HasPointOut() {
		return
	}
	p2 := c.p2(stringBounder)
	x1, y1, x2, y2 := p1.X, p1.Y, p2.X, p2.Y
	xmin := min(x1-hexagonHalfSize, c.f.translateForThen(stringBounder).Dx())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.endInlinkColor, c.f.SkinParam().Arrows().AsToRight()).
		EmphasizeDirection(klimt.DirectionDown)
	snake.AddPoint(x1, y1)
	snake.AddPoint(xmin, y1)
	snake.AddPoint(xmin, y2)
	snake.AddPoint(x2, y2)
	ug.Apply(klimt.NewUTranslate(x2, y2-hexagonHalfSize)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
	ug.Draw(snake)
}

// ftileIfDownElse2Kind tells ConnectionElse2 from its subclasses
// ConnectionElseHline and ConnectionElseNoDiamond.
type ftileIfDownElse2Kind int

const (
	ftileIfDownElse2Normal ftileIfDownElse2Kind = iota
	ftileIfDownElse2Hline
	ftileIfDownElse2NoDiamond
)

type ftileIfDownConnectionElse2 struct {
	AbstractConnection
	f              *FtileIfDown
	endInlinkColor *Rainbow
	kind           ftileIfDownElse2Kind
}

func (f *FtileIfDown) newConnectionElse2(endInlinkColor *Rainbow, kind ftileIfDownElse2Kind) *ftileIfDownConnectionElse2 {
	return &ftileIfDownConnectionElse2{AbstractConnection: newAbstractConnection(f.diamond1, f.diamond2), f: f,
		endInlinkColor: endInlinkColor, kind: kind}
}

func (c *ftileIfDownConnectionElse2) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.f.diamond1.Geometry(stringBounder)
	x := dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	return c.f.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond1.InY()+half))
}

func (c *ftileIfDownConnectionElse2) p2(stringBounder font.StringBounder) geom.XPoint2D {
	if c.kind == ftileIfDownElse2NoDiamond {
		return c.f.Geometry(stringBounder).PointOut()
	}
	dimDiamond2 := c.f.diamond2.Geometry(stringBounder)
	x := dimDiamond2.Width()
	half := (dimDiamond2.OutY() - dimDiamond2.InY()) / 2
	return c.f.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond2.InY()+half))
}

func (c *ftileIfDownConnectionElse2) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	if !c.f.Geometry(stringBounder).HasPointOut() {
		return
	}
	p2 := c.p2(stringBounder)
	x1, y1, x2, y2 := p1.X, p1.Y, p2.X, p2.Y
	thenGeom := c.f.thenBlock.Geometry(stringBounder)
	xmax := max(x1+hexagonHalfSize, c.f.translateForThen(stringBounder).Dx()+thenGeom.Width())
	if c.kind == ftileIfDownElse2Hline {
		snake := SnakeCreateEnd(c.f.SkinParam(), c.endInlinkColor, c.f.SkinParam().Arrows().AsToDown())
		snake.AddPoint(x1, y1)
		snake.AddPoint(xmax, y1)
		snake.AddPoint(xmax, y2)
		ug.Apply(klimt.NewUTranslate(xmax, y2-hexagonHalfSize)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
		ug.Draw(snake)
		return
	}
	snake := SnakeCreateEnd(c.f.SkinParam(), c.endInlinkColor, c.f.SkinParam().Arrows().AsToLeft()).
		EmphasizeDirection(klimt.DirectionDown)
	snake.AddPoint(x1, y1)
	snake.AddPoint(xmax, y1)
	snake.AddPoint(xmax, y2)
	snake.AddPoint(x2, y2)
	ug.Apply(klimt.NewUTranslate(x2, y2-hexagonHalfSize)).Draw(klimt.NewUEmpty(5, hexagonHalfSize))
	ug.Draw(snake)
}

// ftileIfDownConnectionHline is copied upstream from FtileIfLongHorizontal,
// for ConditionEndStyle.HLINE.
type ftileIfDownConnectionHline struct {
	AbstractConnection
	f              *FtileIfDown
	endInlinkColor *Rainbow
}

func (c *ftileIfDownConnectionHline) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.f.diamond1.Geometry(stringBounder)
	x := dimDiamond1.Width()
	half := (dimDiamond1.OutY() - dimDiamond1.InY()) / 2
	return c.f.translateDiamond1(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond1.InY()+half))
}

func (c *ftileIfDownConnectionHline) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.f.diamond2.Geometry(stringBounder)
	x := dimDiamond2.Width()
	half := (dimDiamond2.OutY() - dimDiamond2.InY()) / 2
	return c.f.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond2.InY()+half))
}

// p3 is the bottom point of the omitted diamond.
func (c *ftileIfDownConnectionHline) p3(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.f.diamond2.Geometry(stringBounder)
	x := dimDiamond2.Width()
	return c.f.translateDiamond2(stringBounder).Translated(geom.NewXPoint2D(x, dimDiamond2.OutY()))
}

func (c *ftileIfDownConnectionHline) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	if !c.f.Geometry(stringBounder).HasPointOut() {
		return
	}
	p2 := c.p2(stringBounder)
	p3 := c.p3(stringBounder)
	x1 := p1.X
	x2, y2 := p2.X, p2.Y
	x3, y3 := p3.X, p3.Y
	thenGeom := c.f.thenBlock.Geometry(stringBounder)
	xmax := max(x1+hexagonHalfSize, c.f.translateForThen(stringBounder).Dx()+thenGeom.Width())
	snake := SnakeCreate(c.f.SkinParam(), c.endInlinkColor).WithMerge(MergeNone)
	snake.AddPoint(xmax, y2)
	snake.AddPoint(x2, y2)
	snake.AddPoint(x3, y3)
	ug.Draw(snake)
}

// DrawU implements klimt.TextBlock.
func (f *FtileIfDown) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	if !klimt.TextBlockUtilsIsEmpty(f.opale, stringBounder) {
		xOpale := f.translateDiamond1(stringBounder).Dx() - f.opale.CalculateDimension(stringBounder).Width()
		f.opale.DrawU(ug.Apply(klimt.UTranslateDx(xOpale)))
	}
	ug.Apply(f.translateForThen(stringBounder)).Draw(f.thenBlock)
	ug.Apply(f.translateDiamond1(stringBounder)).Draw(f.diamond1)
	if f.optionalStop == nil {
		ug.Apply(f.translateDiamond2(stringBounder)).Draw(f.diamond2)
	} else {
		ug.Apply(f.translateOptionalStop(stringBounder)).Draw(f.optionalStop)
	}
}

func (f *FtileIfDown) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimOpale := f.opale.CalculateDimension(stringBounder)
	geoDiamond1 := f.diamond1.Geometry(stringBounder)
	geoThen := f.thenBlock.Geometry(stringBounder)
	geoDiamond2 := f.diamond2.Geometry(stringBounder)
	g := geoDiamond1.AppendBottom(geoThen).AppendBottom(geoDiamond2)
	opaleWidth := dimOpale.Width()
	opaleHeight := dimOpale.Height()
	height := g.Height() + 3*hexagonHalfSize + max(hexagonHalfSize*1, f.southLabelHeight(stringBounder)) +
		opaleHeight
	supp := 0.0
	if opaleWidth > g.Left() {
		supp = opaleWidth - g.Left()
	}
	width := supp + g.Width() + hexagonHalfSize
	if f.optionalStop != nil {
		width += f.optionalStop.CalculateDimension(stringBounder).Width() + f.additionalWidth(stringBounder)
	}
	var result FtileGeometry
	if supp > 0 {
		result = NewFtileGeometryOut(width, height, opaleWidth+geoDiamond1.Left(), geoDiamond1.InY()+opaleHeight,
			height)
	} else {
		result = NewFtileGeometryOut(width, height, g.Left(), geoDiamond1.InY()+opaleHeight, height)
	}
	if !geoThen.HasPointOut() && f.optionalStop != nil {
		return result.WithoutPointOut()
	}
	return result
}

func (f *FtileIfDown) additionalWidth(stringBounder font.StringBounder) float64 {
	dimStop := f.optionalStop.Geometry(stringBounder)
	val1 := f.eastLabelWidth(stringBounder)
	stopWidth := dimStop.Width()
	return max(stopWidth, val1+stopWidth/2)
}

func (f *FtileIfDown) southLabelHeight(stringBounder font.StringBounder) float64 {
	switch d := f.diamond1.(type) {
	case *FtileDiamondInside:
		return d.SouthLabelHeight(stringBounder)
	case *FtileDiamond:
		return d.SouthLabelHeight(stringBounder)
	}
	return 0
}

func (f *FtileIfDown) eastLabelWidth(stringBounder font.StringBounder) float64 {
	switch d := f.diamond1.(type) {
	case *FtileDiamondInside:
		return d.EastLabelWidth(stringBounder)
	case *FtileDiamond:
		return d.EastLabelWidth(stringBounder)
	}
	return 0
}

// TranslateFor implements Ftile.
func (f *FtileIfDown) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	var t klimt.UTranslate
	switch {
	case child == f.thenBlock:
		t = f.translateForThen(stringBounder)
	case child == f.diamond1:
		t = f.translateDiamond1(stringBounder)
	case child == f.optionalStop:
		t = f.translateOptionalStop(stringBounder)
	case child == f.diamond2:
		t = f.translateDiamond2(stringBounder)
	default:
		uml.Throw(uml.UnsupportedOperationException, "")
	}
	return &t
}

func (f *FtileIfDown) translateForThen(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	dimDiamond2 := f.diamond2.Geometry(stringBounder)
	dimTotal := f.Geometry(stringBounder)
	dimThen := f.thenBlock.Geometry(stringBounder)
	opaleHeight := f.opale.CalculateDimension(stringBounder).Height()
	y := opaleHeight + dimDiamond1.Height() + (dimTotal.Height()-opaleHeight-dimDiamond1.Height()-
		dimDiamond2.Height()-dimThen.Height())/2
	x := dimTotal.Left() - dimThen.Left()
	return klimt.NewUTranslate(x, y)
}

func (f *FtileIfDown) translateDiamond1(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.Geometry(stringBounder)
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	y1 := f.opale.CalculateDimension(stringBounder).Height()
	x1 := dimTotal.Left() - dimDiamond1.Left()
	return klimt.NewUTranslate(x1, y1)
}

func (f *FtileIfDown) translateOptionalStop(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.Geometry(stringBounder)
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	dimStop := f.optionalStop.Geometry(stringBounder)
	labelNorth := dimDiamond1.InY()
	y1 := labelNorth + (dimDiamond1.Height()-labelNorth-dimStop.Height())/2
	x1 := dimTotal.Left() - dimDiamond1.Left() + dimDiamond1.Width() + f.additionalWidth(stringBounder)
	return klimt.NewUTranslate(x1, y1)
}

func (f *FtileIfDown) translateDiamond2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.Geometry(stringBounder)
	dimDiamond2 := f.diamond2.Geometry(stringBounder)
	y2 := dimTotal.Height() - dimDiamond2.Height()
	x2 := dimTotal.Left() - dimDiamond2.Left()
	return klimt.NewUTranslate(x2, y2)
}
