// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/ParallelBuilderMerge.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ParallelBuilderMerge draws "end merge": a black bar above the branches
// and a diamond below, which the branches join.
type ParallelBuilderMerge struct {
	AbstractParallelFtilesBuilder
}

// NewParallelBuilderMerge returns the builder of a merge.
func NewParallelBuilderMerge(skinParam klimt.ISkinParam, stringBounder font.StringBounder, all []Ftile,
	colors *color.Colors) *ParallelBuilderMerge {
	return &ParallelBuilderMerge{newAbstractParallelFtilesBuilder(skinParam, stringBounder, all, colors)}
}

// Build returns the merge around inner.
func (b *ParallelBuilderMerge) Build(inner Ftile) Ftile {
	step1 := b.doStep1(inner)
	return b.doStep2(inner, step1)
}

func (b *ParallelBuilderMerge) doStep1(inner Ftile) Ftile {
	result := inner
	var conns []Connection
	black := NewFtileBlackBlock(b.SkinParam(), b.list99[0].SwimlaneIn(), b.colors)
	x := 0.0
	for _, tmp := range b.list99 {
		dim := tmp.CalculateDimension(b.StringBounder())
		style := b.StyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
		def := RainbowFromStyle(style, b.SkinParam().IHtmlColorSet())
		rainbow := tmp.InLinkRendering().RainbowOr(def)
		conns = append(conns, newParallelBuilderMergeConnectionIn(b, black, tmp, x, rainbow))
		x += dim.Width()
	}
	result = FtileUtilsAddConnections(result, conns)
	black.SetBlackBlockDimension(result.CalculateDimension(b.StringBounder()).Width(),
		abstractParallelFtilesBuilderBarHeight)
	return NewFtileAssemblySimple(black, result)
}

func (b *ParallelBuilderMerge) doStep2(_, result Ftile) Ftile {
	style := b.StyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(b.SkinParam().IHtmlColorSet())
	backColor := style.Value(klimt.PNameBackGroundColor).AsColor(b.SkinParam().IHtmlColorSet())
	out := NewFtileDiamond(b.SkinParam(), backColor, borderColor, b.swimlaneOutForStep2())
	result = NewFtileAssemblySimple(result, out)
	var conns []Connection
	diamondTranslate := *result.TranslateFor(out, b.StringBounder())
	x := 0.0
	for _, tmp := range b.list99 {
		dim := tmp.CalculateDimension(b.StringBounder())
		translate0 := klimt.NewUTranslate(x, abstractParallelFtilesBuilderBarHeight)
		def := RainbowFromStyle(style, b.SkinParam().IHtmlColorSet())
		rainbow := tmp.OutLinkRendering().RainbowOr(def)
		if tmp.Geometry(b.StringBounder()).HasPointOut() {
			conns = append(conns, newParallelBuilderMergeConnectionHorizontalThenVertical(b, tmp, out, rainbow,
				translate0, diamondTranslate))
		}
		x += dim.Width()
	}
	return FtileUtilsAddConnections(result, conns)
}

// parallelBuilderMergeConnectionHorizontalThenVertical is the arrow from a
// branch to the diamond.
type parallelBuilderMergeConnectionHorizontalThenVertical struct {
	AbstractConnection
	outer            *ParallelBuilderMerge
	arrowColor       *Rainbow
	diamondTranslate klimt.UTranslate
	translate0       klimt.UTranslate
}

func newParallelBuilderMergeConnectionHorizontalThenVertical(outer *ParallelBuilderMerge, tile, diamond Ftile,
	arrowColor *Rainbow, translate0, diamondTranslate klimt.UTranslate) *parallelBuilderMergeConnectionHorizontalThenVertical {
	return &parallelBuilderMergeConnectionHorizontalThenVertical{
		AbstractConnection: newAbstractConnection(tile, diamond), outer: outer, arrowColor: arrowColor,
		diamondTranslate: diamondTranslate, translate0: translate0}
}

func (c *parallelBuilderMergeConnectionHorizontalThenVertical) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder, p1.X)
	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y
	skinParam := c.outer.SkinParam()
	arrival := c.ArrivalOnDiamond(stringBounder, p1.X)
	var endDecoration *klimt.UPolygon
	switch {
	case arrival.Dx() < 0:
		endDecoration = skinParam.Arrows().AsToRight()
	case arrival.Dx() > 0:
		endDecoration = skinParam.Arrows().AsToLeft()
	default:
		endDecoration = skinParam.Arrows().AsToDown()
	}
	snake := SnakeCreateEnd(skinParam, c.arrowColor, endDecoration)
	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *parallelBuilderMergeConnectionHorizontalThenVertical) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.translate0.Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *parallelBuilderMergeConnectionHorizontalThenVertical) p2(stringBounder font.StringBounder,
	startX float64) geom.XPoint2D {
	arrival := c.ArrivalOnDiamond(stringBounder, startX)
	return arrival.Translated(c.DiamondOut(stringBounder))
}

// DiamondOut returns the exit point of the diamond.
func (c *parallelBuilderMergeConnectionHorizontalThenVertical) DiamondOut(stringBounder font.StringBounder) geom.XPoint2D {
	return c.diamondTranslate.Translated(c.Ftile2().Geometry(stringBounder).PointOut())
}

// ArrivalOnDiamond returns where the arrow enters the diamond, from the
// exit point.
func (c *parallelBuilderMergeConnectionHorizontalThenVertical) ArrivalOnDiamond(stringBounder font.StringBounder,
	startX float64) klimt.UTranslate {
	result := c.DiamondOut(stringBounder)
	dim := c.Ftile2().CalculateDimension(stringBounder)
	a := result.X - dim.Width()/2
	b := result.X + dim.Width()/2
	switch {
	case startX < a:
		return klimt.NewUTranslate(-dim.Width()/2, -dim.Height()/2)
	case startX > b:
		return klimt.NewUTranslate(dim.Width()/2, -dim.Height()/2)
	}
	return klimt.NewUTranslate(0, -dim.Height())
}

// parallelBuilderMergeConnectionIn is the arrow from the bar into a
// branch.
type parallelBuilderMergeConnectionIn struct {
	AbstractConnection
	outer      *ParallelBuilderMerge
	x          float64
	arrowColor *Rainbow
	label      *klimt.Display
}

func newParallelBuilderMergeConnectionIn(outer *ParallelBuilderMerge, ftile1, ftile2 Ftile, x float64,
	arrowColor *Rainbow) *parallelBuilderMergeConnectionIn {
	return &parallelBuilderMergeConnectionIn{AbstractConnection: newAbstractConnection(ftile1, ftile2), outer: outer,
		label: ftile2.InLinkRendering().Display(), x: x, arrowColor: arrowColor}
}

func (c *parallelBuilderMergeConnectionIn) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	g := c.Ftile2().Geometry(c.outer.StringBounder())
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	snake.AddPoint(g.Left(), 0)
	snake.AddPoint(g.Left(), g.InY())
	ug.Draw(snake)
}

func (c *parallelBuilderMergeConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	g := c.Ftile2().Geometry(c.outer.StringBounder())
	p1 := geom.NewXPoint2D(g.Left(), 0)
	p2 := geom.NewXPoint2D(g.Left(), g.InY())
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := mp1a.Y + 4
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}
