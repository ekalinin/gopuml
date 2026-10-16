// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/ParallelBuilderFork.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// ParallelBuilderFork draws a fork: a black bar above and below the
// branches, the bar below possibly labeled.
type ParallelBuilderFork struct {
	AbstractParallelFtilesBuilder
	label    string
	hasLabel bool
	in       *Swimlane
	out      *Swimlane
}

// NewParallelBuilderFork returns the builder of a fork; label is the
// "{label}" of the bar below (hasLabel false for none).
func NewParallelBuilderFork(skinParam klimt.ISkinParam, stringBounder font.StringBounder, label string,
	hasLabel bool, in, out *Swimlane, all []Ftile, colors *color.Colors) *ParallelBuilderFork {
	return &ParallelBuilderFork{
		AbstractParallelFtilesBuilder: newAbstractParallelFtilesBuilder(skinParam, stringBounder, all, colors),
		label:                         label, hasLabel: hasLabel, in: in, out: out,
	}
}

// Build returns the fork around inner.
func (b *ParallelBuilderFork) Build(inner Ftile) Ftile {
	step1 := b.doStep1(inner)
	return b.doStep2(inner, step1)
}

func (b *ParallelBuilderFork) swimlaneOutForStep2() *Swimlane { return b.out }

func (b *ParallelBuilderFork) doStep1(middle Ftile) Ftile {
	result := middle
	var conns []Connection
	swimlaneBlack := b.in
	style := b.StyleSignatureArrow().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	black := NewFtileBlackBlock(b.SkinParam(), swimlaneBlack, b.colors)
	x := 0.0
	for _, tmp := range b.list99 {
		dim := tmp.CalculateDimension(b.StringBounder())
		def := RainbowFromStyle(style, b.SkinParam().IHtmlColorSet())
		rainbow := tmp.InLinkRendering().RainbowOr(def)
		conns = append(conns, newParallelBuilderForkConnectionIn(b, black, tmp, x, rainbow))
		x += dim.Width()
	}
	result = FtileUtilsAddConnections(result, conns)
	black.SetBlackBlockDimension(result.CalculateDimension(b.StringBounder()).Width(),
		abstractParallelFtilesBuilderBarHeight)
	return NewFtileAssemblySimple(black, result)
}

func (b *ParallelBuilderFork) justBeforeBar2(middle Ftile, _ font.StringBounder) float64 {
	return abstractParallelFtilesBuilderBarHeight + b.heightOfMiddle(middle)
}

func (b *ParallelBuilderFork) doStep2(middle, result Ftile) Ftile {
	swimlaneBlack := b.out
	out := NewFtileBlackBlock(b.SkinParam(), swimlaneBlack, b.colors)
	out.SetBlackBlockDimension(result.CalculateDimension(b.StringBounder()).Width(),
		abstractParallelFtilesBuilderBarHeight)
	if b.hasLabel {
		out.SetLabel(b.TextBlock(klimt.DisplayWithNewlines(b.SkinParam().Pragma(), b.label)))
	}
	result = NewFtileAssemblySimple(result, out)
	var conns []Connection
	style := b.StyleSignatureArrow().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	x := 0.0
	for _, tmp := range b.list99 {
		dim := tmp.CalculateDimension(b.StringBounder())
		def := RainbowFromStyle(style, b.SkinParam().IHtmlColorSet())
		rainbow := tmp.OutLinkRendering().RainbowOr(def)
		if tmp.Geometry(b.StringBounder()).HasPointOut() {
			conns = append(conns, newParallelBuilderForkConnectionOut(b, tmp, out, x, rainbow,
				b.justBeforeBar2(middle, b.StringBounder())))
		}
		x += dim.Width()
	}
	return FtileUtilsAddConnections(result, conns)
}

// parallelBuilderForkConnectionIn is the arrow from the upper bar into a
// branch.
type parallelBuilderForkConnectionIn struct {
	AbstractConnection
	outer      *ParallelBuilderFork
	x          float64
	arrowColor *Rainbow
	label      *klimt.Display
}

func newParallelBuilderForkConnectionIn(outer *ParallelBuilderFork, ftile1, ftile2 Ftile, x float64,
	arrowColor *Rainbow) *parallelBuilderForkConnectionIn {
	return &parallelBuilderForkConnectionIn{AbstractConnection: newAbstractConnection(ftile1, ftile2), outer: outer,
		label: ftile2.InLinkRendering().Display(), x: x, arrowColor: arrowColor}
}

func (c *parallelBuilderForkConnectionIn) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	geo2 := c.Ftile2().Geometry(c.outer.StringBounder())
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	p1 := geom.NewXPoint2D(geo2.Left(), 0)
	p2 := geom.NewXPoint2D(geo2.Left(), geo2.InY())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *parallelBuilderForkConnectionIn) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	geo2 := c.Ftile2().Geometry(c.outer.StringBounder())
	p1 := geom.NewXPoint2D(geo2.Left(), 0)
	p2 := geom.NewXPoint2D(geo2.Left(), geo2.InY())
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).IgnoreForCompression()
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

// parallelBuilderForkConnectionOut is the arrow from a branch to the lower
// bar.
type parallelBuilderForkConnectionOut struct {
	AbstractConnection
	outer          *ParallelBuilderFork
	x              float64
	arrowColor     *Rainbow
	label          *klimt.Display
	justBeforeBar2 float64
}

func newParallelBuilderForkConnectionOut(outer *ParallelBuilderFork, ftile1, ftile2 Ftile, x float64,
	arrowColor *Rainbow, justBeforeBar2 float64) *parallelBuilderForkConnectionOut {
	return &parallelBuilderForkConnectionOut{AbstractConnection: newAbstractConnection(ftile1, ftile2), outer: outer,
		justBeforeBar2: justBeforeBar2, label: ftile1.OutLinkRendering().Display(), x: x, arrowColor: arrowColor}
}

func (c *parallelBuilderForkConnectionOut) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	geo1 := c.Ftile1().Geometry(c.outer.StringBounder())
	if !geo1.HasPointOut() {
		return
	}
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown())
	if !klimt.DisplayIsNull(c.label) {
		snake = snake.WithLabel(c.outer.TextBlock(c.label), c.ArrowHorizontalAlignment())
	}
	p1 := geom.NewXPoint2D(geo1.Left(), abstractParallelFtilesBuilderBarHeight+geo1.OutY())
	p2 := geom.NewXPoint2D(geo1.Left(), c.justBeforeBar2)
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *parallelBuilderForkConnectionOut) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	ug = ug.Apply(klimt.UTranslateDx(c.x))
	g := c.Ftile1().Geometry(c.outer.StringBounder())
	if !g.HasPointOut() {
		return
	}
	p1 := geom.NewXPoint2D(g.Left(), abstractParallelFtilesBuilderBarHeight+g.OutY())
	p2 := geom.NewXPoint2D(g.Left(), c.justBeforeBar2)
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.arrowColor, skinParam.Arrows().AsToDown()).IgnoreForCompression()
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
