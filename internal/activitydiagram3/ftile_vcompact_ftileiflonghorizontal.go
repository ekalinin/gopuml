// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileIfLongHorizontal.java
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

const ftileIfLongHorizontalXSeparation = 20

// FtileIfLongHorizontal is an "if" with "elseif": the diamonds in a row,
// each above its branch, the "else" branch on the right.
type FtileIfLongHorizontal struct {
	AbstractFtile
	tiles      []Ftile
	tile2      Ftile
	diamonds   []Ftile
	couples    []Ftile
	arrowColor *Rainbow
}

func newFtileIfLongHorizontal(diamonds []Ftile, inlabelSizes []float64, tiles []Ftile, tile2 Ftile,
	arrowColor *Rainbow) *FtileIfLongHorizontal {
	f := &FtileIfLongHorizontal{}
	f.initFtile(f, tiles[0].SkinParam())
	if len(diamonds) != len(tiles) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	for i, diamond := range diamonds {
		tmp := NewFtileAssemblySimple(diamond, tiles[i])
		f.couples = append(f.couples, FtileUtilsAddHorizontalMargin2(tmp, inlabelSizes[i], 0))
	}
	f.tile2 = tile2
	f.diamonds = slices.Clone(diamonds)
	f.tiles = slices.Clone(tiles)
	f.arrowColor = arrowColor
	return f
}

func ftileIfLongHorizontalAlignDiamonds(diamonds []Ftile, stringBounder font.StringBounder) []Ftile {
	maxOutY := ftileIfLongHorizontalMaxOutY(diamonds, stringBounder)
	var result []Ftile
	for _, diamond := range diamonds {
		missing := maxOutY - diamond.Geometry(stringBounder).OutY()
		result = append(result, FtileUtilsAddVerticalMargin(diamond, missing/2, 20))
	}
	return result
}

func ftileIfLongHorizontalMaxOutY(diamonds []Ftile, stringBounder font.StringBounder) float64 {
	maxOutY := 0.0
	for _, diamond := range diamonds {
		maxOutY = math.Max(maxOutY, diamond.Geometry(stringBounder).OutY())
	}
	return maxOutY
}

// Swimlanes implements Swimable.
func (f *FtileIfLongHorizontal) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if f.SwimlaneIn() != nil {
		result = result.add(f.SwimlaneIn())
	}
	for _, tile := range f.couples {
		result = result.addAll(tile.Swimlanes())
	}
	return result.addAll(f.tile2.Swimlanes())
}

// SwimlaneIn implements Swimable.
func (f *FtileIfLongHorizontal) SwimlaneIn() *Swimlane { return f.couples[0].SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileIfLongHorizontal) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

// FtileIfLongHorizontalCreate returns the tile of an "if" with "elseif".
func FtileIfLongHorizontalCreate(swimlane *Swimlane, backColor color.HColor, ftileFactory FtileFactory,
	conditionStyle klimt.ConditionStyle, thens []*Branch, branch2 *Branch, topInlinkRendering,
	afterEndwhile *LinkRendering, styleArrow, styleDiamond *klimt.Style) Ftile {
	requireLink(afterEndwhile)
	set := ftileFactory.SkinParam().IHtmlColorSet()
	borderColor := styleDiamond.Value(klimt.PNameLineColor).AsColor(set)
	arrowColor := RainbowFromStyle(styleArrow, set)
	fcTest := styleDiamond.FontConfiguration(set, nil)
	fcArrow := styleArrow.FontConfiguration(set, nil)

	var tiles []Ftile
	for _, branch := range thens {
		tiles = append(tiles, NewFtileMinWidthCentered(branch.Ftile(), 30))
	}
	tile2 := NewFtileMinWidthCentered(branch2.Ftile(), 30)

	var diamonds []Ftile
	var inlabelSizes []float64
	for _, branch := range thens {
		tb1 := branch.DisplayPositive().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
		lineBreak := styleDiamond.WrapWidth()
		tbTest := branch.LabelTest().Create0(fcTest,
			ftileFactory.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft), ftileFactory.SkinParam(),
			lineBreak, klimt.CreoleModeFull, nil, nil)
		diamondColor := branch.Color()
		if diamondColor == nil {
			diamondColor = backColor
		}
		diamond := NewFtileDiamondInside2(tbTest, branch.SkinParam(), diamondColor, borderColor, swimlane)
		if klimt.DisplayIsNull(branch.Inlabel()) {
			inlabelSizes = append(inlabelSizes, 0.0)
		} else {
			tbInlabel := branch.Inlabel().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
			inlabelSizes = append(inlabelSizes, tbInlabel.CalculateDimension(ftileFactory.StringBounder()).Width())
			diamond = diamond.WithWest(tbInlabel)
		}
		diamond = diamond.WithNorth(tb1)
		diamonds = append(diamonds, diamond)
	}

	tb2 := branch2.DisplayPositive().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
	last := len(diamonds) - 1
	diamonds[last] = diamonds[last].(*FtileDiamondInside2).WithEast(tb2)

	diamonds = ftileIfLongHorizontalAlignDiamonds(diamonds, ftileFactory.StringBounder())

	result := newFtileIfLongHorizontal(diamonds, inlabelSizes, tiles, tile2, arrowColor)
	var conns []Connection

	nbOut := 0
	for i := range thens {
		ftile := tiles[i]
		diam := diamonds[i]
		rainbowIn := thens[i].InColor(arrowColor)
		branch := thens[i]
		if branch.Ftile().Geometry(ftileFactory.StringBounder()).HasPointOut() {
			nbOut++
		}
		rainbowOut := branch.Out()
		var out2 klimt.TextBlock
		if branch.Special() != nil {
			out2 = branch.Special().Display().Create(fcTest, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
		}
		rainbow1 := rainbowIn
		if rainbowIn.Size() == 0 {
			rainbow1 = arrowColor
		}
		rainbow2 := rainbowOut
		if rainbowOut.Size() == 0 {
			rainbow2 = arrowColor
		}
		conns = append(conns,
			&ftileIfLongHorizontalConnectionVerticalIn{AbstractConnection: newAbstractConnection(diam, ftile),
				f: result, color: rainbow1},
			&ftileIfLongHorizontalConnectionVerticalOut{AbstractConnection: newAbstractConnection(ftile, nil),
				f: result, color: rainbow2, out2: out2})
	}

	topInColor := topInlinkRendering.RainbowOr(arrowColor)
	for i := 0; i < len(diamonds)-1; i++ {
		diam1 := diamonds[i]
		diam2 := diamonds[i+1]
		branch := thens[i+1]
		conns = append(conns, &ftileIfLongHorizontalConnectionHorizontal{
			AbstractConnection: newAbstractConnection(diam1, diam2), f: result,
			color: branch.InRainbow(arrowColor)})
	}
	conns = append(conns, &ftileIfLongHorizontalConnectionIn{
		AbstractConnection: newAbstractConnection(nil, result.diamonds[0]), f: result, arrowColor: topInColor})
	var out2 klimt.TextBlock
	if branch2.Special() != nil {
		out2 = branch2.Special().Display().Create(fcTest, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
	}
	rainbowOut := branch2.Out()
	rainbowIn := branch2.InColor(arrowColor)
	rainbow1 := rainbowIn
	if rainbowIn.Size() == 0 {
		rainbow1 = arrowColor
	}
	rainbow2 := rainbowOut
	if rainbowOut.Size() == 0 {
		rainbow2 = arrowColor
	}
	conns = append(conns,
		&ftileIfLongHorizontalConnectionLastElseIn{
			AbstractConnection: newAbstractConnection(result.diamonds[len(result.diamonds)-1], result.tile2),
			f:                  result, arrowColor: rainbow1},
		&ftileIfLongHorizontalConnectionLastElseOut{AbstractConnection: newAbstractConnection(result.tile2, nil),
			f: result, arrowColor: rainbow2, out2: out2, nbOut: nbOut})
	if nbOut > 0 {
		horizontalOutColor := afterEndwhile.RainbowOr(arrowColor)
		conns = append(conns, &ftileIfLongHorizontalConnectionHline{AbstractConnection: newAbstractConnection(nil, nil),
			f: result, arrowColor: horizontalOutColor})
	}
	return FtileUtilsAddConnections(result, conns)
}

type ftileIfLongHorizontalConnectionHorizontal struct {
	AbstractConnection
	f     *FtileIfLongHorizontal
	color *Rainbow
}

func (c *ftileIfLongHorizontalConnectionHorizontal) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToRight())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongHorizontalConnectionHorizontal) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.Ftile1().Geometry(stringBounder)
	p := geom.NewXPoint2D(dimDiamond1.Left()*2, ftileIfLongHorizontalYdiamontOutToLeft(dimDiamond1))
	return c.f.translateDiamond1(c.Ftile1(), stringBounder).Translated(p)
}

func (c *ftileIfLongHorizontalConnectionHorizontal) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.Ftile2().Geometry(stringBounder)
	p := geom.NewXPoint2D(0, ftileIfLongHorizontalYdiamontOutToLeft(dimDiamond1))
	return c.f.translateDiamond1(c.Ftile2(), stringBounder).Translated(p)
}

func ftileIfLongHorizontalYdiamontOutToLeft(dimDiamond1 FtileGeometry) float64 {
	return (dimDiamond1.InY() + dimDiamond1.OutY()) / 2
}

type ftileIfLongHorizontalConnectionIn struct {
	AbstractConnection
	f          *FtileIfLongHorizontal
	arrowColor *Rainbow
}

func (c *ftileIfLongHorizontalConnectionIn) DrawU(ug klimt.UGraphic) {
	tr := c.f.translateDiamond1(c.Ftile2(), ug.StringBounder())
	p2 := tr.Translated(c.Ftile2().Geometry(ug.StringBounder()).PointIn())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	p1 := c.f.calculateDimensionInternal(ug.StringBounder()).PointIn()
	snake.AddPointPt(p1)
	snake.AddPoint(p2.X, p1.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfLongHorizontalConnectionLastElseIn struct {
	AbstractConnection
	f          *FtileIfLongHorizontal
	arrowColor *Rainbow
}

func (c *ftileIfLongHorizontalConnectionLastElseIn) DrawU(ug klimt.UGraphic) {
	p1 := c.p1(ug.StringBounder())
	tr2 := c.f.translate2(ug.StringBounder())
	p2 := tr2.Translated(c.Ftile2().Geometry(ug.StringBounder()).PointIn())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(p1)
	snake.AddPoint(p2.X, p1.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongHorizontalConnectionLastElseIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.Ftile1().Geometry(stringBounder)
	p := geom.NewXPoint2D(dimDiamond1.Left()*2, ftileIfLongHorizontalYdiamontOutToLeft(dimDiamond1))
	return c.f.translateDiamond1(c.Ftile1(), stringBounder).Translated(p)
}

type ftileIfLongHorizontalConnectionLastElseOut struct {
	AbstractConnection
	f          *FtileIfLongHorizontal
	arrowColor *Rainbow
	out2       klimt.TextBlock
	nbOut      int
}

func (c *ftileIfLongHorizontalConnectionLastElseOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	tr1 := c.f.translate2(stringBounder)
	dim := c.Ftile1().Geometry(stringBounder)
	if !dim.HasPointOut() {
		return
	}
	p1 := tr1.Translated(dim.PointOut())
	full := c.f.calculateDimensionInternal(stringBounder)
	totalHeight := full.Height()
	p2 := geom.NewXPoint2D(p1.X, totalHeight)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.arrowColor, c.f.SkinParam().Arrows().AsToDown()).WithLabel(c.out2,
		c.ArrowHorizontalAlignment())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	if c.nbOut == 0 {
		snake.AddPointPt(geom.NewXPoint2D(full.Left(), totalHeight))
	}
	ug.Draw(snake)
}

type ftileIfLongHorizontalConnectionVerticalIn struct {
	AbstractConnection
	f     *FtileIfLongHorizontal
	color *Rainbow
}

func (c *ftileIfLongHorizontalConnectionVerticalIn) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongHorizontalConnectionVerticalIn) p1(stringBounder font.StringBounder) geom.XPoint2D {
	p := c.Ftile1().Geometry(stringBounder).PointOut()
	return c.f.translateDiamond1(c.Ftile1(), stringBounder).Translated(p)
}

func (c *ftileIfLongHorizontalConnectionVerticalIn) p2(stringBounder font.StringBounder) geom.XPoint2D {
	p := c.Ftile2().Geometry(stringBounder).PointIn()
	return c.f.translate1(c.Ftile2(), stringBounder).Translated(p)
}

func (c *ftileIfLongHorizontalConnectionVerticalIn) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	p1 := c.p1(ug.StringBounder())
	p2 := c.p2(ug.StringBounder())
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown())
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	middle := mp1a.Y + 4
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

type ftileIfLongHorizontalConnectionVerticalOut struct {
	AbstractConnection
	f     *FtileIfLongHorizontal
	color *Rainbow
	out2  klimt.TextBlock
}

func (c *ftileIfLongHorizontalConnectionVerticalOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	totalHeight := c.f.calculateDimensionInternal(stringBounder).Height()
	p1, ok := c.p1(stringBounder)
	if !ok {
		return
	}
	p2 := geom.NewXPoint2D(p1.X, totalHeight)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown()).WithLabel(c.out2,
		c.ArrowHorizontalAlignment())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

func (c *ftileIfLongHorizontalConnectionVerticalOut) p1(stringBounder font.StringBounder) (geom.XPoint2D, bool) {
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return geom.XPoint2D{}, false
	}
	return c.f.translate1(c.Ftile1(), stringBounder).Translated(g.PointOut()), true
}

type ftileIfLongHorizontalConnectionHline struct {
	AbstractConnection
	f          *FtileIfLongHorizontal
	arrowColor *Rainbow
}

func (c *ftileIfLongHorizontalConnectionHline) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	totalDim := c.f.calculateDimensionInternal(stringBounder)
	allTiles := append(slices.Clone(c.f.couples), c.f.tile2)
	var minX, maxX float64
	if interceptor, ok := ug.(*UGraphicInterceptorOneSwimlane); ok {
		allSwimlanes := interceptor.OrderedListOfAllSwimlanes()
		minX, maxX = c.minmax(stringBounder, totalDim.Width(), allTiles, interceptor.Swimlane(), allSwimlanes)
	} else {
		minX, maxX = c.minmaxSimple(stringBounder, totalDim.Width(), allTiles)
	}
	if math.IsNaN(minX) || math.IsNaN(maxX) {
		return
	}
	s := SnakeCreate(c.f.SkinParam(), c.arrowColor).WithMerge(MergeNone)
	s.AddPoint(minX, totalDim.Height())
	s.AddPoint(maxX, totalDim.Height())
	ug.Draw(s)
}

func (c *ftileIfLongHorizontalConnectionHline) leftOut(stringBounder font.StringBounder) (float64, bool) {
	dim := c.f.Geometry(stringBounder)
	if dim.HasPointOut() {
		return dim.Left(), true
	}
	return 0, false
}

func (c *ftileIfLongHorizontalConnectionHline) minmax(stringBounder font.StringBounder, width float64,
	allTiles []Ftile, intoSw *Swimlane, allSwimlanes []*Swimlane) (float64, float64) {
	current := slices.Index(allSwimlanes, intoSw)
	leftOut, ok := c.leftOut(stringBounder)
	if !ok {
		return math.NaN(), math.NaN()
	}
	if current == -1 {
		uml.Throw(uml.IllegalStateException, "")
	}
	first := ftileIfHlineFirstSwimlane(stringBounder, allTiles, allSwimlanes)
	last := ftileIfHlineLastSwimlane(stringBounder, allTiles, allSwimlanes)
	if current < first || current > last {
		return math.NaN(), math.NaN()
	}
	minX := width
	if current != first {
		minX = 0
	}
	maxX := 0.0
	if current != last {
		maxX = width
	}
	minX = math.Min(minX, leftOut)
	maxX = math.Max(maxX, leftOut)
	for _, tmp := range allTiles {
		if !tmp.Geometry(stringBounder).HasPointOut() {
			continue
		}
		if !ftileIfHlineOutcomeInSwimlane(tmp, intoSw) {
			continue
		}
		ut := c.f.TranslateFor(tmp, stringBounder)
		out := tmp.Geometry(stringBounder).Translate(*ut).Left()
		minX = math.Min(minX, out)
		maxX = math.Max(maxX, out)
	}
	return minX, maxX
}

func (c *ftileIfLongHorizontalConnectionHline) minmaxSimple(stringBounder font.StringBounder, width float64,
	allTiles []Ftile) (float64, float64) {
	leftOut, ok := c.leftOut(stringBounder)
	if !ok {
		return math.NaN(), math.NaN()
	}
	minX := width / 2
	maxX := width / 2
	minX = math.Min(minX, leftOut)
	maxX = math.Max(maxX, leftOut)
	for _, tmp := range allTiles {
		if !tmp.Geometry(stringBounder).HasPointOut() {
			continue
		}
		ut := c.f.TranslateFor(tmp, stringBounder)
		out := tmp.Geometry(stringBounder).Translate(*ut).Left()
		minX = math.Min(minX, out)
		maxX = math.Max(maxX, out)
	}
	return minX, maxX
}

// MyChildren implements Ftile.
func (f *FtileIfLongHorizontal) MyChildren() []Ftile { return append(slices.Clone(f.tiles), f.tile2) }

// TranslateFor implements Ftile.
func (f *FtileIfLongHorizontal) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	var t klimt.UTranslate
	switch {
	case child == f.tile2:
		t = f.translate2(stringBounder)
	case slices.Contains(f.couples, child):
		t = f.translateCouple1(child, stringBounder)
	case slices.Contains(f.tiles, child):
		t = f.translate1(child, stringBounder)
	default:
		uml.Throw(uml.UnsupportedOperationException, "")
	}
	return &t
}

func (f *FtileIfLongHorizontal) translate2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dim2 := f.tile2.CalculateDimension(stringBounder)
	x2 := dimTotal.Width() - dim2.Width()
	h := 0.0
	y2 := (dimTotal.Height()-h*2-dim2.Height())/2 + h
	return klimt.NewUTranslate(x2, y2)
}

func (f *FtileIfLongHorizontal) translateDiamond1(diamond Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	idx := slices.Index(f.diamonds, diamond)
	if idx == -1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	trCouple := f.translateCouple1(f.couples[idx], stringBounder)
	in := f.couples[idx].TranslateFor(diamond, stringBounder)
	return trCouple.Compose(*in)
}

func (f *FtileIfLongHorizontal) translate1(tile Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	idx := slices.Index(f.tiles, tile)
	if idx == -1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	trCouple := f.translateCouple1(f.couples[idx], stringBounder)
	in := f.couples[idx].TranslateFor(tile, stringBounder)
	return trCouple.Compose(*in)
}

func (f *FtileIfLongHorizontal) translateCouple1(candidate Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	x1 := 0.0
	for _, couple := range f.couples {
		dim1 := couple.Geometry(stringBounder)
		if couple == candidate {
			return klimt.NewUTranslate(x1, 25)
		}
		x1 += dim1.Width() + ftileIfLongHorizontalXSeparation
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}

// DrawU implements klimt.TextBlock.
func (f *FtileIfLongHorizontal) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	for _, couple := range f.couples {
		ug.Apply(f.translateCouple1(couple, stringBounder)).Draw(couple)
	}
	ug.Apply(f.translate2(stringBounder)).Draw(f.tile2)
}

func (f *FtileIfLongHorizontal) calculateDimensionInternal(stringBounder font.StringBounder) FtileGeometry {
	maxOutY := ftileIfLongHorizontalMaxOutY(f.diamonds, stringBounder)
	result := geom.NewXDimension2D(0, 0)
	for _, couple := range f.couples {
		result = result.MergeLR(couple.CalculateDimension(stringBounder))
	}
	dimTile2 := f.tile2.CalculateDimension(stringBounder)
	dimTile2 = dimTile2.DeltaWH(0, f.diamondsHeight(stringBounder)/2)
	result = result.MergeLR(dimTile2)
	result = result.DeltaWH(float64(ftileIfLongHorizontalXSeparation*len(f.couples)), math.Max(100, maxOutY))
	return NewFtileGeometryDim(result, result.Width()/2, 0)
}

func (f *FtileIfLongHorizontal) diamondsHeight(stringBounder font.StringBounder) float64 {
	height := 0.0
	for _, diamond := range f.diamonds {
		height = math.Max(height, diamond.Geometry(stringBounder).Height())
	}
	return height
}

func (f *FtileIfLongHorizontal) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	all := append(slices.Clone(f.tiles), f.tile2)
	for _, tmp := range all {
		if tmp.Geometry(stringBounder).HasPointOut() {
			return NewFtileGeometryDimOut(dimTotal.Dim(), dimTotal.Width()/2, 0, dimTotal.Height())
		}
	}
	return NewFtileGeometryDim(dimTotal.Dim(), dimTotal.Width()/2, 0)
}
