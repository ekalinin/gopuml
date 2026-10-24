// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileIfLongVertical.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

const (
	ftileIfLongVerticalYSeparation = 20
	ftileIfLongVerticalMarginy1    = 30
)

// FtileIfLongVertical is an "if" with "elseif" drawn vertically
// ("!pragma useVerticalIf on"): the diamonds in a column, each branch on
// its right.
type FtileIfLongVertical struct {
	AbstractFtile
	tiles       []Ftile
	tile2       Ftile
	diamonds    []Ftile
	lastDiamond Ftile
	arrowColor  *Rainbow
}

// MyChildren implements Ftile.
func (f *FtileIfLongVertical) MyChildren() []Ftile { return append(slices.Clone(f.tiles), f.tile2) }

func newFtileIfLongVertical(diamonds, tiles []Ftile, tile2 Ftile, arrowColor *Rainbow,
	lastDiamond Ftile) *FtileIfLongVertical {
	f := &FtileIfLongVertical{}
	f.initFtile(f, tiles[0].SkinParam())
	if len(diamonds) != len(tiles) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	f.lastDiamond = lastDiamond
	f.tile2 = tile2
	f.diamonds = slices.Clone(diamonds)
	f.tiles = slices.Clone(tiles)
	f.arrowColor = arrowColor
	return f
}

// Swimlanes implements Swimable.
func (f *FtileIfLongVertical) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if f.SwimlaneIn() != nil {
		result = result.add(f.SwimlaneIn())
	}
	for _, tile := range f.tiles {
		result = result.addAll(tile.Swimlanes())
	}
	return result.addAll(f.tile2.Swimlanes())
}

// SwimlaneIn implements Swimable.
func (f *FtileIfLongVertical) SwimlaneIn() *Swimlane { return f.tiles[0].SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileIfLongVertical) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

// FtileIfLongVerticalCreate returns the tile of an "if" with "elseif",
// drawn vertically.
func FtileIfLongVerticalCreate(swimlane *Swimlane, backColor color.HColor, ftileFactory FtileFactory,
	conditionStyle klimt.ConditionStyle, thens []*Branch, branch2 *Branch, topInlinkRendering,
	afterEndwhile *LinkRendering, styleArrow, styleDiamond *klimt.Style) Ftile {
	set := ftileFactory.SkinParam().IHtmlColorSet()
	fcArrow := styleArrow.FontConfiguration(set, nil)
	borderColor := styleDiamond.Value(klimt.PNameLineColor).AsColor(set)
	arrowColor := RainbowFromStyle(styleArrow, set)

	var diamonds []Ftile
	west := 10.0
	for _, branch := range thens {
		tb1 := branch.DisplayPositive().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
		tbTest := branch.LabelTest().Create(fcArrow,
			ftileFactory.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft), ftileFactory.SkinParam())
		diamond := NewFtileDiamondInside2(tbTest, branch.SkinParam(), backColor, borderColor, swimlane)
		diamond = diamond.WithEast(tb1)
		diamonds = append(diamonds, diamond)
		if !klimt.DisplayIsNull(branch.Inlabel()) {
			tbInlabel := branch.Inlabel().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
			west = math.Max(west, tbInlabel.CalculateDimension(ftileFactory.StringBounder()).Width())
		}
	}

	var tiles []Ftile
	for _, branch := range thens {
		tiles = append(tiles, NewFtileMargedWest(branch.Ftile(), west))
	}
	lastDiamond := NewFtileDiamond(tiles[0].SkinParam(), backColor, borderColor, swimlane)
	tile2 := NewFtileMinWidthCentered(branch2.Ftile(), 30)
	result := newFtileIfLongVertical(diamonds, tiles, tile2, arrowColor, lastDiamond)

	var conns []Connection
	for i := range thens {
		ftile := tiles[i]
		diam := diamonds[i]
		c := thens[i].InColor(arrowColor)
		if c == nil {
			c = arrowColor
		}
		conns = append(conns, &ftileIfLongVerticalConnectionVerticalIn{
			AbstractConnection: newAbstractConnection(diam, ftile), f: result, color: c})
	}
	for i := 0; i < len(diamonds)-1; i++ {
		branch := thens[i+1]
		var tbInlabel klimt.TextBlock
		if !klimt.DisplayIsNull(branch.Inlabel()) {
			tbInlabel = branch.Inlabel().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
		}
		conns = append(conns, &ftileIfLongVerticalConnectionVertical{
			AbstractConnection: newAbstractConnection(diamonds[i], diamonds[i+1]), f: result, color: arrowColor,
			label: tbInlabel})
	}
	conns = append(conns, &ftileIfLongVerticalConnectionThenOut{
		AbstractConnection: newAbstractConnection(tiles[0], lastDiamond), f: result, arrowColor: arrowColor})
	for i := 1; i < len(tiles); i++ {
		conns = append(conns, &ftileIfLongVerticalConnectionThenOutConnect{
			AbstractConnection: newAbstractConnection(tiles[i], lastDiamond), f: result, arrowColor: arrowColor})
	}
	topInColor := topInlinkRendering.RainbowOr(arrowColor)
	conns = append(conns, &ftileIfLongVerticalConnectionIn{
		AbstractConnection: newAbstractConnection(nil, result.diamonds[0]), f: result, arrowColor: topInColor})
	tb2 := branch2.DisplayPositive().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
	conns = append(conns,
		&ftileIfLongVerticalConnectionLastElse{
			AbstractConnection: newAbstractConnection(result.diamonds[len(result.diamonds)-1], result.tile2),
			f:                  result, arrowColor: topInColor, label: tb2},
		&ftileIfLongVerticalConnectionLastElseOut{
			AbstractConnection: newAbstractConnection(result.tile2, result.lastDiamond), f: result,
			arrowColor: arrowColor})
	return FtileUtilsAddConnections(result, conns)
}

type ftileIfLongVerticalConnectionIn struct {
	AbstractConnection
	f          *FtileIfLongVertical
	arrowColor *Rainbow
}

func (c *ftileIfLongVerticalConnectionIn) DrawU(ug klimt.UGraphic) {
	tr := c.f.translateDiamond(c.Ftile2(), ug.StringBounder())
	p2 := tr.Translated(c.Ftile2().Geometry(ug.StringBounder()).PointIn())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	p1 := c.f.calculateDimensionInternal(ug.StringBounder()).PointIn()
	snake.AddPointPt(p1)
	snake.AddPoint(p1.X, (p1.Y+p2.Y)/2)
	snake.AddPoint(p2.X, (p1.Y+p2.Y)/2)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfLongVerticalConnectionVerticalIn struct {
	AbstractConnection
	f     *FtileIfLongVertical
	color *Rainbow
}

func (c *ftileIfLongVerticalConnectionVerticalIn) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(p1)
	snake.AddPoint(p2.X, p1.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongVerticalConnectionVerticalIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.Ftile1().CalculateDimension(stringBounder)
	diamondWidth := dimDiamond1.Width()
	return c.f.translateDiamond(c.Ftile1(), stringBounder).Translated(
		geom.NewXPoint2D(diamondWidth, dimDiamond1.Height()/2))
}

func (c *ftileIfLongVerticalConnectionVerticalIn) p2(stringBounder font.StringBounder) geom.XPoint2D {
	p := c.Ftile2().Geometry(stringBounder).PointIn()
	return c.f.translate1(c.Ftile2(), stringBounder).Translated(p)
}

type ftileIfLongVerticalConnectionVertical struct {
	AbstractConnection
	f     *FtileIfLongVertical
	color *Rainbow
	label klimt.TextBlock
}

func (c *ftileIfLongVerticalConnectionVertical) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown()).WithLabelVertical(c.label,
		geom.VerticalAlignmentCenter)
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongVerticalConnectionVertical) p1(stringBounder font.StringBounder) geom.XPoint2D {
	p := c.Ftile1().Geometry(stringBounder).PointOut()
	return c.f.TranslateFor(c.Ftile1(), stringBounder).Translated(p)
}

func (c *ftileIfLongVerticalConnectionVertical) p2(stringBounder font.StringBounder) geom.XPoint2D {
	p := c.Ftile2().Geometry(stringBounder).PointIn()
	return c.f.TranslateFor(c.Ftile2(), stringBounder).Translated(p)
}

type ftileIfLongVerticalConnectionLastElse struct {
	AbstractConnection
	f          *FtileIfLongVertical
	arrowColor *Rainbow
	label      klimt.TextBlock
}

func (c *ftileIfLongVerticalConnectionLastElse) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	tr1 := c.f.translateDiamond(c.Ftile1(), stringBounder)
	dimDiamond := c.Ftile1().Geometry(stringBounder)
	p1 := tr1.Translated(dimDiamond.PointOut())
	p2 := c.f.translate2(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown()).
		WithLabelVertical(c.label, geom.VerticalAlignmentCenter)
	snake.AddPointPt(p1)
	snake.AddPoint(p1.X, p2.Y-15)
	snake.AddPoint(p2.X, p2.Y-15)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfLongVerticalConnectionLastElseOut struct {
	AbstractConnection
	f          *FtileIfLongVertical
	arrowColor *Rainbow
}

func (c *ftileIfLongVerticalConnectionLastElseOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim1 := c.Ftile1().Geometry(stringBounder)
	if !dim1.HasPointOut() {
		return
	}
	p1 := c.f.translate2(stringBounder).Translated(dim1.PointOut())
	p2 := c.f.translateLastDiamond(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(p1)
	snake.AddPoint(p1.X, p2.Y-15)
	snake.AddPoint(p2.X, p2.Y-15)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfLongVerticalConnectionThenOut struct {
	AbstractConnection
	f          *FtileIfLongVertical
	arrowColor *Rainbow
}

func (c *ftileIfLongVerticalConnectionThenOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim1 := c.Ftile1().Geometry(stringBounder)
	if !dim1.HasPointOut() {
		return
	}
	p1 := c.f.translate1(c.Ftile1(), stringBounder).Translated(dim1.PointOut())
	dimLastDiamond := c.Ftile2().Geometry(stringBounder)
	p2 := c.f.translateLastDiamond(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
	p2 = klimt.NewUTranslate(dimLastDiamond.Width()/2, dimLastDiamond.Height()/2).Translated(p2)
	dimTotal := c.f.calculateDimensionInternal(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToLeft())
	snake.AddPointPt(p1)
	snake.AddPoint(p1.X, p1.Y+15)
	snake.AddPoint(dimTotal.Width(), p1.Y+15)
	snake.AddPoint(dimTotal.Width(), p2.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfLongVerticalConnectionThenOutConnect struct {
	AbstractConnection
	f          *FtileIfLongVertical
	arrowColor *Rainbow
}

func (c *ftileIfLongVerticalConnectionThenOutConnect) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim1 := c.Ftile1().Geometry(stringBounder)
	if !dim1.HasPointOut() {
		return
	}
	p1 := c.f.translate1(c.Ftile1(), stringBounder).Translated(dim1.PointOut())
	dimTotal := c.f.calculateDimensionInternal(stringBounder)
	p2 := geom.NewXPoint2D(dimTotal.Width(), p1.Y+15)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToRight())
	snake.AddPointPt(p1)
	snake.AddPoint(p1.X, p2.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

// TranslateFor implements Ftile.
func (f *FtileIfLongVertical) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	var t klimt.UTranslate
	switch {
	case child == f.tile2:
		t = f.translate2(stringBounder)
	case child == f.lastDiamond:
		t = f.translateLastDiamond(stringBounder)
	case slices.Contains(f.tiles, child):
		t = f.translate1(child, stringBounder)
	case slices.Contains(f.diamonds, child):
		t = f.translateDiamond(child, stringBounder)
	default:
		uml.Throw(uml.UnsupportedOperationException, "")
	}
	return &t
}

func (f *FtileIfLongVertical) translateDiamond(diamond Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	allDiamondsWidth := f.allDiamondsWidth(stringBounder)
	idx := slices.Index(f.diamonds, diamond)
	if idx == -1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	y1 := f.translateDy(idx, stringBounder)
	return klimt.NewUTranslate((allDiamondsWidth-diamond.CalculateDimension(stringBounder).Width())/2, y1)
}

func (f *FtileIfLongVertical) translateLastDiamond(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimLast := f.lastDiamond.Geometry(stringBounder)
	x := (dimTotal.Width() - dimLast.Width()) / 2
	return klimt.NewUTranslate(x, dimTotal.Height()-dimLast.Height())
}

func (f *FtileIfLongVertical) translate1(candidate Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	idx := slices.Index(f.tiles, candidate)
	if idx == -1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	y1 := f.translateDy(idx, stringBounder)
	diam := f.diamonds[idx].Geometry(stringBounder)
	dim1 := candidate.Geometry(stringBounder)
	dimTotal := f.calculateDimensionInternal(stringBounder)
	allDiamondsWidth := f.allDiamondsWidth(stringBounder)
	x := allDiamondsWidth + (dimTotal.Width()-allDiamondsWidth-dim1.Width())/2
	return klimt.NewUTranslate(x, y1+diam.Height())
}

func (f *FtileIfLongVertical) translateDy(idx int, stringBounder font.StringBounder) float64 {
	y1 := float64(ftileIfLongVerticalMarginy1)
	for i := 0; i < idx; i++ {
		dim1 := f.tiles[i].Geometry(stringBounder)
		diam := f.diamonds[i].Geometry(stringBounder)
		y1 += dim1.Height() + diam.Height() + ftileIfLongVerticalYSeparation
	}
	return y1
}

func (f *FtileIfLongVertical) translate2(stringBounder font.StringBounder) klimt.UTranslate {
	y1 := f.translateDy(len(f.tiles), stringBounder)
	dim2 := f.tile2.Geometry(stringBounder)
	dimTotal := f.calculateDimensionInternal(stringBounder)
	x := (dimTotal.Width() - dim2.Width()) / 2
	return klimt.NewUTranslate(x, y1)
}

// DrawU implements klimt.TextBlock.
func (f *FtileIfLongVertical) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	for _, tile1 := range f.tiles {
		ug.Apply(f.translate1(tile1, stringBounder)).Draw(tile1)
	}
	for _, diam := range f.diamonds {
		ug.Apply(f.translateDiamond(diam, stringBounder)).Draw(diam)
	}
	ug.Apply(f.translate2(stringBounder)).Draw(f.tile2)
	ug.Apply(f.translateLastDiamond(stringBounder)).Draw(f.lastDiamond)
}

func (f *FtileIfLongVertical) calculateDimensionInternal(stringBounder font.StringBounder) FtileGeometry {
	col1 := 0.0
	col2 := 0.0
	height := float64(ftileIfLongVerticalMarginy1)
	for i := range f.tiles {
		dim1 := f.tiles[i].Geometry(stringBounder)
		diamondDim := f.diamonds[i].Geometry(stringBounder)
		height += diamondDim.Height() + dim1.Height()
		col1 = math.Max(col1, diamondDim.Width())
		col2 = math.Max(col2, dim1.Width())
	}
	width := col1 + col2
	result := NewFtileGeometry(width, height, width/2, 0)
	dimTile2 := f.tile2.Geometry(stringBounder)
	result = result.AppendBottom(dimTile2)
	dimLastDiamond := f.lastDiamond.Geometry(stringBounder)
	const lastElseArrowHeight = 40
	result = result.AddDim(0, float64(ftileIfLongVerticalYSeparation*len(f.tiles))+lastElseArrowHeight+
		dimLastDiamond.Height())
	return NewFtileGeometryDim(result.Dim(), result.Width()/2, 0)
}

func (f *FtileIfLongVertical) allDiamondsWidth(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, diam := range f.diamonds {
		width = math.Max(width, diam.Geometry(stringBounder).Width())
	}
	return width
}

func (f *FtileIfLongVertical) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	all := append(slices.Clone(f.tiles), f.tile2)
	for _, tmp := range all {
		if tmp.Geometry(stringBounder).HasPointOut() {
			return NewFtileGeometryDimOut(dimTotal.Dim(), dimTotal.Width()/2, 0, dimTotal.Height())
		}
	}
	return NewFtileGeometryDim(dimTotal.Dim(), dimTotal.Width()/2, 0)
}
