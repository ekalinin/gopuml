// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileIfWithLinks.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"
	"slices"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// FtileIfWithLinks is an "if" with diamonds and the arrows between the
// diamonds and the branches.
type FtileIfWithLinks struct {
	FtileIfWithDiamonds
	conditionEndStyle klimt.ConditionEndStyle
	arrowColor        *Rainbow
}

// NewFtileIfWithLinks returns the "if" of the diamonds and branches; the
// arrows are added by AddLinks.
func NewFtileIfWithLinks(diamond1, tile1, tile2, diamond2 Ftile, in *Swimlane, arrowColor *Rainbow,
	conditionEndStyle klimt.ConditionEndStyle, stringBounder font.StringBounder,
	notes []*PositionedNote) *FtileIfWithLinks {
	f := &FtileIfWithLinks{}
	f.initFtileIfWithDiamonds(f, diamond1, tile1, tile2, diamond2, in, stringBounder, notes)
	f.arrowColor = arrowColor
	f.conditionEndStyle = conditionEndStyle
	if arrowColor.Size() == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return f
}

func (f *FtileIfWithLinks) yDeltaForLabels(stringBounder font.StringBounder) float64 {
	if d, ok := f.diamond2.(*FtileDiamond); ok {
		return d.WestEastLabelHeight(stringBounder)
	}
	return 0
}

// translateOfTile is the translate of the inner classes: the position of
// tile1 or tile2.
func (f *FtileIfWithLinks) translateOfTile(tile Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	switch tile {
	case f.tile1:
		return f.translateBranch1(stringBounder)
	case f.tile2:
		return f.translateBranch2(stringBounder)
	}
	uml.Throw(uml.IllegalStateException, "")
	return klimt.UTranslate{}
}

type ftileIfWithLinksConnectionHorizontalThenVertical struct {
	AbstractConnection
	f          *FtileIfWithLinks
	color      *Rainbow
	usingArrow *klimt.UPolygon
}

func (f *FtileIfWithLinks) newConnectionHorizontalThenVertical(tile Ftile,
	branch *Branch) *ftileIfWithLinksConnectionHorizontalThenVertical {
	c := &ftileIfWithLinksConnectionHorizontalThenVertical{AbstractConnection: newAbstractConnection(f.diamond1, tile),
		f: f}
	c.color = branch.InColor(f.arrowColor)
	if c.color.Size() == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if !branch.IsEmpty() {
		c.usingArrow = f.SkinParam().Arrows().AsToDown()
	}
	return c
}

func (c *ftileIfWithLinksConnectionHorizontalThenVertical) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.usingArrow)
	snake.AddPoint(p1.X, p1.Y)
	snake.AddPoint(p2.X, p1.Y)
	snake.AddPoint(p2.X, p2.Y)
	ug.Draw(snake)
}

func (c *ftileIfWithLinksConnectionHorizontalThenVertical) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.f.diamond1.Geometry(stringBounder)
	var pt geom.XPoint2D
	switch c.Ftile2() {
	case c.f.tile1:
		pt = dimDiamond1.PointD()
	case c.f.tile2:
		pt = dimDiamond1.PointB()
	default:
		uml.Throw(uml.IllegalStateException, "")
	}
	return c.f.translateDiamond1(stringBounder).Translated(pt)
}

func (c *ftileIfWithLinksConnectionHorizontalThenVertical) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.f.translateOfTile(c.Ftile2(), stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (c *ftileIfWithLinksConnectionHorizontalThenVertical) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	originalDirection := klimt.DirectionLeftOrRight(p1, p2)
	p1 = translate1.Translated(p1)
	p2 = translate2.Translated(p2)
	newDirection := klimt.DirectionLeftOrRight(p1, p2)
	if originalDirection != newDirection {
		delta := hexagonHalfSize
		if originalDirection == klimt.DirectionRight {
			delta = -hexagonHalfSize
		}
		dimDiamond1 := c.f.diamond1.CalculateDimension(stringBounder)
		small := SnakeCreate(c.f.SkinParam(), c.color)
		small.AddPointPt(p1)
		small.AddPoint(p1.X+delta, p1.Y)
		small.AddPoint(p1.X+delta, p1.Y+float64(dimDiamond1.Height()*.75))
		ug.Draw(small)
		p1 = small.Last()
	}
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.usingArrow).WithMerge(MergeLimited)
	snake.AddPointPt(p1)
	snake.AddPoint(p2.X, p1.Y)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

type ftileIfWithLinksConnectionVerticalThenHorizontal struct {
	AbstractConnection
	f            *FtileIfWithLinks
	myArrowColor *Rainbow
	branchEmpty  bool
}

func (f *FtileIfWithLinks) newConnectionVerticalThenHorizontal(tile Ftile, myArrowColor *Rainbow,
	branchEmpty bool) *ftileIfWithLinksConnectionVerticalThenHorizontal {
	if myArrowColor == nil || myArrowColor.Size() == 0 {
		myArrowColor = f.arrowColor
	}
	return &ftileIfWithLinksConnectionVerticalThenHorizontal{
		AbstractConnection: newAbstractConnection(tile, f.diamond2), f: f, myArrowColor: myArrowColor,
		branchEmpty: branchEmpty}
}

func (c *ftileIfWithLinksConnectionVerticalThenHorizontal) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p1 := g.Translate(c.f.translateOfTile(c.Ftile1(), stringBounder)).PointOut()
	p2 := c.p2(stringBounder)
	x1, y1, x2, y2 := p1.X, p1.Y, p2.X, p2.Y
	arrow := c.f.SkinParam().Arrows().AsToLeft()
	if x2 > x1 {
		arrow = c.f.SkinParam().Arrows().AsToRight()
	}
	snake := SnakeCreateEnd(c.f.SkinParam(), c.myArrowColor, arrow)
	if c.branchEmpty {
		snake = snake.EmphasizeDirection(klimt.DirectionDown)
	}
	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileIfWithLinksConnectionVerticalThenHorizontal) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.f.diamond2.Geometry(stringBounder)
	var pt geom.XPoint2D
	switch c.Ftile1() {
	case c.f.tile1:
		pt = dimDiamond2.PointD()
	case c.f.tile2:
		pt = dimDiamond2.PointB()
	default:
		uml.Throw(uml.IllegalStateException, "")
	}
	return c.f.translateDiamond2(stringBounder).Translated(pt)
}

func (c *ftileIfWithLinksConnectionVerticalThenHorizontal) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p2 := c.p2(stringBounder)
	p1 := g.Translate(c.f.translateOfTile(c.Ftile1(), stringBounder)).PointOut()
	originalDirection := klimt.DirectionLeftOrRight(p1, p2)
	x1 := p1.X
	x2 := p2.X
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	newDirection := klimt.DirectionLeftOrRight(mp1a, mp2b)
	arrow := c.f.SkinParam().Arrows().AsToLeft()
	if x2 > x1 {
		arrow = c.f.SkinParam().Arrows().AsToRight()
	}
	delta := 1.5 * hexagonHalfSize
	if x2 > x1 {
		delta = -1.5 * hexagonHalfSize
	}
	if originalDirection == newDirection {
		mp2bc := geom.NewXPoint2D(mp2b.X+delta, mp2b.Y)
		snake := SnakeCreate(c.f.SkinParam(), c.myArrowColor).WithMerge(MergeLimited)
		middle := (mp1a.Y + mp2b.Y) / 2.0
		snake.AddPointPt(mp1a)
		snake.AddPoint(mp1a.X, middle)
		snake.AddPoint(mp2bc.X, middle)
		snake.AddPointPt(mp2bc)
		ug.Draw(snake)
		small := SnakeCreateEnd(c.f.SkinParam(), c.myArrowColor, arrow).WithMerge(MergeLimited)
		small.AddPointPt(mp2bc)
		small.AddPoint(mp2bc.X, mp2b.Y)
		small.AddPointPt(mp2b)
		ug.Draw(small)
	} else {
		mp2bb := geom.NewXPoint2D(mp2b.X+delta, mp2b.Y-1.5*hexagonHalfSize)
		snake := SnakeCreate(c.f.SkinParam(), c.myArrowColor).WithMerge(MergeLimited)
		snake.AddPointPt(mp1a)
		snake.AddPoint(mp1a.X, mp2bb.Y)
		snake.AddPointPt(mp2bb)
		ug.Draw(snake)
		small := SnakeCreateEnd(c.f.SkinParam(), c.myArrowColor, arrow).WithMerge(MergeLimited)
		small.AddPointPt(mp2bb)
		small.AddPoint(mp2bb.X, mp2b.Y)
		small.AddPointPt(mp2b)
		ug.Draw(small)
	}
}

type ftileIfWithLinksConnectionVerticalThenHorizontalDirect struct {
	AbstractConnection
	f            *FtileIfWithLinks
	myArrowColor *Rainbow
	branchEmpty  bool
}

func (f *FtileIfWithLinks) newConnectionVerticalThenHorizontalDirect(tile Ftile, myArrowColor *Rainbow,
	branchEmpty bool) *ftileIfWithLinksConnectionVerticalThenHorizontalDirect {
	if myArrowColor == nil || myArrowColor.Size() == 0 {
		myArrowColor = f.arrowColor
	}
	return &ftileIfWithLinksConnectionVerticalThenHorizontalDirect{
		AbstractConnection: newAbstractConnection(tile, f.diamond2), f: f, myArrowColor: myArrowColor,
		branchEmpty: branchEmpty}
}

func (c *ftileIfWithLinksConnectionVerticalThenHorizontalDirect) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := c.f.calculateDimensionInternal(stringBounder)
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p1 := g.Translate(c.f.translateOfTile(c.Ftile1(), stringBounder)).PointOut()
	p2 := geom.NewXPoint2D(dimTotal.Left(), dimTotal.Height())
	x1, y1, x2, y2 := p1.X, p1.Y, p2.X, p2.Y
	snake := SnakeCreate(c.f.SkinParam(), c.myArrowColor)
	if c.branchEmpty {
		snake = snake.EmphasizeDirection(klimt.DirectionDown)
	}
	snake.AddPoint(x1, y1)
	snake.AddPoint(x1, y2)
	snake.AddPoint(x2, y2)
	snake.AddPoint(x2, dimTotal.Height())
	ug.Draw(snake)
}

func (c *ftileIfWithLinksConnectionVerticalThenHorizontalDirect) DrawTranslate(ug klimt.UGraphic, translate1,
	translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	dimTotal := c.f.calculateDimensionInternal(stringBounder)
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p1 := g.Translate(c.f.translateOfTile(c.Ftile1(), stringBounder)).PointOut()
	p2 := geom.NewXPoint2D(dimTotal.Left(), dimTotal.Height()-hexagonHalfSize)
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	snake := SnakeCreate(c.f.SkinParam(), c.myArrowColor).WithMerge(MergeLimited)
	x1 := mp1a.X
	x2 := mp2b.X
	y2 := mp2b.Y
	snake.AddPointPt(mp1a)
	snake.AddPoint(x1, y2)
	snake.AddPointPt(mp2b)
	snake.AddPoint(x2, dimTotal.Height())
	ug.Draw(snake)
}

// ftileIfWithLinksConnectionVerticalOut is copied upstream from
// FtileIfLongHorizontal, for ConditionEndStyle.HLINE.
type ftileIfWithLinksConnectionVerticalOut struct {
	AbstractConnection
	f     *FtileIfWithLinks
	color *Rainbow
	out2  klimt.TextBlock
}

func (c *ftileIfWithLinksConnectionVerticalOut) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p1 := g.Translate(c.f.translateOfTile(c.Ftile1(), stringBounder)).PointOut()
	totalHeight := c.f.calculateDimensionInternal(stringBounder).Height()
	p2 := geom.NewXPoint2D(p1.X, totalHeight)
	snake := SnakeCreateEnd(c.f.SkinParam(), c.color, c.f.SkinParam().Arrows().AsToDown()).WithLabel(c.out2,
		c.ArrowHorizontalAlignment())
	snake.AddPointPt(p1)
	snake.AddPointPt(p2)
	ug.Draw(snake)
}

// ftileIfWithLinksConnectionHline is copied upstream from
// FtileIfLongHorizontal, for ConditionEndStyle.HLINE.
type ftileIfWithLinksConnectionHline struct {
	AbstractConnection
	f          *FtileIfWithLinks
	arrowColor *Rainbow
}

func (c *ftileIfWithLinksConnectionHline) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	totalDim := c.f.calculateDimensionInternal(stringBounder)
	allTiles := []Ftile{c.f.tile1, c.f.tile2}
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

func (c *ftileIfWithLinksConnectionHline) minmax(stringBounder font.StringBounder, width float64, allTiles []Ftile,
	intoSw *Swimlane, allSwimlanes []*Swimlane) (float64, float64) {
	current := slices.Index(allSwimlanes, intoSw)
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

func (c *ftileIfWithLinksConnectionHline) minmaxSimple(stringBounder font.StringBounder, width float64,
	allTiles []Ftile) (float64, float64) {
	minX := width / 2
	maxX := width / 2
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

// ftileIfHlineFirstSwimlane is getFirstSwimlane of the ConnectionHline
// classes: the first lane where a tile goes out.
func ftileIfHlineFirstSwimlane(stringBounder font.StringBounder, allTiles []Ftile, allSwimlanes []*Swimlane) int {
	for i := range allSwimlanes {
		if ftileIfHlineAtLeastOne(stringBounder, allSwimlanes[i], allTiles) {
			return i
		}
	}
	uml.Throw(uml.IllegalStateException, "")
	return 0
}

// ftileIfHlineLastSwimlane is getLastSwimlane of the ConnectionHline
// classes: the last lane where a tile goes out.
func ftileIfHlineLastSwimlane(stringBounder font.StringBounder, allTiles []Ftile, allSwimlanes []*Swimlane) int {
	for i := len(allSwimlanes) - 1; i >= 0; i-- {
		if ftileIfHlineAtLeastOne(stringBounder, allSwimlanes[i], allTiles) {
			return i
		}
	}
	uml.Throw(uml.IllegalStateException, "")
	return 0
}

func ftileIfHlineAtLeastOne(stringBounder font.StringBounder, intoSw *Swimlane, allTiles []Ftile) bool {
	for _, tmp := range allTiles {
		if tmp.Geometry(stringBounder).HasPointOut() && ftileIfHlineOutcomeInSwimlane(tmp, intoSw) {
			return true
		}
	}
	return false
}

func ftileIfHlineOutcomeInSwimlane(ftile Ftile, swimlane *Swimlane) bool {
	return ftile.SwimlaneOut() == swimlane && ftile.Swimlanes().Contains(swimlane)
}

// AddLinks returns the "if" with its arrows.
func (f *FtileIfWithLinks) AddLinks(branch1, branch2 *Branch, stringBounder font.StringBounder) Ftile {
	conns := []Connection{
		f.newConnectionHorizontalThenVertical(f.tile1, branch1),
		f.newConnectionHorizontalThenVertical(f.tile2, branch2),
	}
	hasPointOut1 := f.tile1.Geometry(stringBounder).HasPointOut()
	hasPointOut2 := f.tile2.Geometry(stringBounder).HasPointOut()
	switch f.conditionEndStyle {
	case klimt.ConditionEndStyleDiamond:
		switch {
		case hasPointOut1 && hasPointOut2:
			conns = append(conns, f.newConnectionVerticalThenHorizontal(f.tile1, branch1.Out(), branch1.IsEmpty()),
				f.newConnectionVerticalThenHorizontal(f.tile2, branch2.Out(), branch2.IsEmpty()))
		case hasPointOut1 && !hasPointOut2:
			conns = append(conns, f.newConnectionVerticalThenHorizontalDirect(f.tile1, branch1.Out(),
				branch1.IsEmpty()))
		case !hasPointOut1 && hasPointOut2:
			conns = append(conns, f.newConnectionVerticalThenHorizontalDirect(f.tile2, branch2.Out(),
				branch2.IsEmpty()))
		}
	case klimt.ConditionEndStyleHline:
		switch {
		case hasPointOut1 && hasPointOut2:
			conns = append(conns,
				&ftileIfWithLinksConnectionVerticalOut{AbstractConnection: newAbstractConnection(f.tile1, nil), f: f,
					color: f.arrowColor},
				&ftileIfWithLinksConnectionVerticalOut{AbstractConnection: newAbstractConnection(f.tile2, nil), f: f,
					color: f.arrowColor},
				&ftileIfWithLinksConnectionHline{AbstractConnection: newAbstractConnection(nil, nil), f: f,
					arrowColor: f.arrowColor})
		case hasPointOut1 && !hasPointOut2:
			// this is called when the "else" has a break statement
			conns = append(conns, f.newConnectionVerticalThenHorizontalDirect(f.tile1, branch1.Out(),
				branch1.IsEmpty()))
		case !hasPointOut1 && hasPointOut2:
			// this is called when the "if" has a break statement
			conns = append(conns, f.newConnectionVerticalThenHorizontalDirect(f.tile2, branch2.Out(),
				branch2.IsEmpty()))
		}
	}
	return FtileUtilsAddConnections(f, conns)
}
