// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/ParallelBuilderSplit.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ParallelBuilderSplit draws a split: a thin line above and below the
// branches.
type ParallelBuilderSplit struct {
	AbstractParallelFtilesBuilder
}

// NewParallelBuilderSplit returns the builder of a split; colors may be
// nil.
func NewParallelBuilderSplit(skinParam klimt.ISkinParam, stringBounder font.StringBounder, all []Ftile,
	colors *color.Colors) *ParallelBuilderSplit {
	return &ParallelBuilderSplit{newAbstractParallelFtilesBuilder(skinParam, stringBounder, all, colors)}
}

// StyleSignature is the style signature of the thin lines: the arrows.
func (b *ParallelBuilderSplit) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameArrow)
}

// Build returns the split around inner.
func (b *ParallelBuilderSplit) Build(inner Ftile) Ftile {
	step1 := b.doStep1(inner)
	return b.doStep2(inner, step1)
}

func (b *ParallelBuilderSplit) doStep1(inner Ftile) Ftile {
	result := inner
	var conns []Connection
	style := b.StyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	thinColor := RainbowFromStyle(style, b.SkinParam().IHtmlColorSet())
	thin := NewFtileThinSplit(b.SkinParam(), b.thin1Color(thinColor), b.list99[0].SwimlaneIn())
	x := 0.0
	first := 0.0
	last := 0.0
	for _, tmp := range b.list99 {
		dim := tmp.Geometry(b.StringBounder())
		if first == 0 {
			first = x + dim.Left()
		}
		last = x + dim.Left()
		inLinkRendering := tmp.InLinkRendering()
		rainbow := inLinkRendering.RainbowOr(RainbowFromStyle(style, b.SkinParam().IHtmlColorSet()))
		conns = append(conns, newParallelBuilderSplitConnectionIn(b, thin, tmp, x, rainbow))
		x += dim.Width()
	}
	result = FtileUtilsAddConnections(result, conns)
	g := result.Geometry(b.StringBounder())
	if last < g.Left() {
		last = g.Left()
	}
	if first > g.Left() {
		first = g.Left()
	}
	thin.SetGeom(first, last, result.CalculateDimension(b.StringBounder()).Width())
	return NewFtileAssemblySimple(thin, result)
}

func (b *ParallelBuilderSplit) thin1Color(thinColor *Rainbow) color.HColor {
	style := b.StyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	for _, tmp := range b.list99 {
		inLinkRendering := tmp.InLinkRendering()
		rainbow := inLinkRendering.RainbowOr(RainbowFromStyle(style, b.SkinParam().IHtmlColorSet()))
		if !rainbow.IsInvisible() {
			return thinColor.Color()
		}
	}
	return nil
}

func (b *ParallelBuilderSplit) hasOut() bool {
	for _, tmp := range b.list99 {
		if tmp.Geometry(b.StringBounder()).HasPointOut() {
			return true
		}
	}
	return false
}

func (b *ParallelBuilderSplit) doStep2(inner, result Ftile) Ftile {
	g := result.Geometry(b.StringBounder())
	if !b.hasOut() {
		return NewFtileKilled(result)
	}
	style := b.StyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	inLinkRendering := result.InLinkRendering()
	thinColor := inLinkRendering.RainbowOr(RainbowFromStyle(style, b.SkinParam().IHtmlColorSet()))
	out := NewFtileThinSplit(b.SkinParam(), thinColor.Color(), b.swimlaneOutForStep2())
	result = NewFtileAssemblySimple(result, out)
	var conns []Connection
	x := 0.0
	first := 0.0
	last := 0.0
	for _, tmp := range b.list99 {
		translate0 := klimt.UTranslateDy(1.5)
		dim := tmp.Geometry(b.StringBounder())
		if dim.HasPointOut() {
			if first == 0 {
				first = x + dim.Left()
			}
			last = x + dim.Left()
		}
		outLinkRendering := tmp.OutLinkRendering()
		rainbow := outLinkRendering.RainbowOr(RainbowFromStyle(style, b.SkinParam().IHtmlColorSet()))
		if tmp.Geometry(b.StringBounder()).HasPointOut() {
			conns = append(conns, newParallelBuilderSplitConnectionOut(b, translate0, tmp, out, x, rainbow,
				b.heightOfMiddle(inner)))
		}
		x += dim.Width()
	}
	if last < g.Left() {
		last = g.Left()
	}
	if first > g.Left() {
		first = g.Left()
	}
	out.SetGeom(first, last, g.Width())
	return FtileUtilsAddConnections(result, conns)
}

// parallelBuilderSplitConnectionIn is the arrow from the upper line into a
// branch.
type parallelBuilderSplitConnectionIn struct {
	AbstractConnection
	outer      *ParallelBuilderSplit
	x          float64
	arrowColor *Rainbow
	label      *klimt.Display
}

func newParallelBuilderSplitConnectionIn(outer *ParallelBuilderSplit, ftile1, ftile2 Ftile, x float64,
	arrowColor *Rainbow) *parallelBuilderSplitConnectionIn {
	return &parallelBuilderSplitConnectionIn{AbstractConnection: newAbstractConnection(ftile1, ftile2), outer: outer,
		label: ftile2.InLinkRendering().Display(), x: x, arrowColor: arrowColor}
}

func (c *parallelBuilderSplitConnectionIn) DrawU(ug klimt.UGraphic) {
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

func (c *parallelBuilderSplitConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
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

// parallelBuilderSplitConnectionOut is the arrow from a branch to the
// lower line.
type parallelBuilderSplitConnectionOut struct {
	AbstractConnection
	outer      *ParallelBuilderSplit
	x          float64
	arrowColor *Rainbow
	height     float64
	label      *klimt.Display
	translate0 klimt.UTranslate
}

func newParallelBuilderSplitConnectionOut(outer *ParallelBuilderSplit, translate0 klimt.UTranslate, ftile1,
	ftile2 Ftile, x float64, arrowColor *Rainbow, height float64) *parallelBuilderSplitConnectionOut {
	return &parallelBuilderSplitConnectionOut{AbstractConnection: newAbstractConnection(ftile1, ftile2), outer: outer,
		translate0: translate0, label: ftile1.OutLinkRendering().Display(), x: x, arrowColor: arrowColor,
		height: height}
}

func (c *parallelBuilderSplitConnectionOut) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	g := c.Ftile1().Geometry(c.outer.StringBounder())
	if !g.HasPointOut() {
		return
	}
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	p1 := c.translate0.Translated(geom.NewXPoint2D(g.Left(), g.OutY()))
	p2 := c.translate0.Translated(geom.NewXPoint2D(g.Left(), c.height))
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *parallelBuilderSplitConnectionOut) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	g := c.Ftile1().Geometry(c.outer.StringBounder())
	if !g.HasPointOut() {
		return
	}
	p1 := c.translate0.Translated(geom.NewXPoint2D(g.Left(), g.OutY()))
	p2 := c.translate0.Translated(geom.NewXPoint2D(g.Left(), c.height))
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := mp2b.Y - 14
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}
