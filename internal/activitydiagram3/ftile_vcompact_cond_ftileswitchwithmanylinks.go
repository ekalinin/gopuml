// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileSwitchWithManyLinks.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

const ftileSwitchWithManyLinksMargin = 10.0

// FtileSwitchWithManyLinks is a switch with several cases: the arrows go
// from the corners of the test diamond to the cases and from the cases to
// the merge diamond.
type FtileSwitchWithManyLinks struct {
	FtileSwitchWithDiamonds
	arrowColor *Rainbow
}

// NewFtileSwitchWithManyLinks returns the switch of several cases.
func NewFtileSwitchWithManyLinks(tiles []Ftile, branches []*Branch, in *Swimlane, diamond1, diamond2 Ftile,
	stringBounder font.StringBounder, arrowColor *Rainbow) *FtileSwitchWithManyLinks {
	f := &FtileSwitchWithManyLinks{arrowColor: arrowColor}
	f.initFtileSwitchWithDiamonds(f, tiles, branches, in, diamond1, diamond2, stringBounder)
	return f
}

// ftileSwitchWithManyLinksConnectionHorizontalThenVertical is the arrow
// from a side corner of the test diamond to the first or last case.
type ftileSwitchWithManyLinksConnectionHorizontalThenVertical struct {
	AbstractConnection
	outer  *FtileSwitchWithManyLinks
	branch *Branch
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVertical) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.branch.TextBlockPositive(), c.ArrowHorizontalAlignment())
	snake.AddPoint(x1, y1)
	if c.isLast() && p1.X > p2.X {
		dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
		snake.AddPoint(x1+hexagonHalfSize, y1)
		snake.AddPoint(x1+hexagonHalfSize, y1+dimDiamond1.Height())
		snake.AddPoint(x2, y1+dimDiamond1.Height())
	} else {
		snake.AddPoint(x2, y1)
	}
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVertical) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
	var pt geom.XPoint2D
	switch {
	case c.Ftile2() == c.outer.tiles[0]:
		pt = dimDiamond1.PointD()
	case c.isLast():
		pt = dimDiamond1.PointB()
	default:
		uml.Throw(uml.IllegalStateException, "")
	}
	return c.outer.translateDiamond1(stringBounder).Translated(pt)
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVertical) isLast() bool {
	return c.Ftile2() == c.outer.tiles[len(c.outer.tiles)-1]
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVertical) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile2(), stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

// ftileSwitchWithManyLinksConnectionVerticalThenHorizontal is the arrow
// from the first or last case to a side corner of the merge diamond.
type ftileSwitchWithManyLinksConnectionVerticalThenHorizontal struct {
	AbstractConnection
	outer    *FtileSwitchWithManyLinks
	outLabel klimt.TextBlock
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontal) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	g := c.Ftile1().Geometry(stringBounder)
	if !g.HasPointOut() {
		return
	}
	p1 := c.p1(stringBounder)
	x1 := p1.X
	y1 := p1.Y
	dimDiamond2 := c.outer.diamond2.Geometry(stringBounder)
	translateDiamond2 := c.outer.translateDiamond2(stringBounder)
	ptA := translateDiamond2.Translated(dimDiamond2.PointA())
	ptB := translateDiamond2.Translated(dimDiamond2.PointB())
	ptD := translateDiamond2.Translated(dimDiamond2.PointD())
	skinParam := c.outer.SkinParam()
	var p2 geom.XPoint2D
	var arrow *klimt.UPolygon
	var direction klimt.Direction
	switch {
	case x1 < ptD.X:
		p2 = ptD
		arrow = skinParam.Arrows().AsToRight()
		direction = klimt.DirectionRight
	case x1 > ptB.X:
		p2 = ptB
		arrow = skinParam.Arrows().AsToLeft()
		direction = klimt.DirectionLeft
	default:
		p2 = ptA
		arrow = skinParam.Arrows().AsToDown()
		direction = klimt.DirectionDown
	}
	x2 := p2.X
	y2 := p2.Y
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, arrow).WithLabelVertical(c.outLabel,
		geom.VerticalAlignmentCenter)
	snake.AddPoint(x1, y1)
	if direction == klimt.DirectionLeft && x2 > x1-10 {
		snake.AddPoint(x1, y2-8)
		snake.AddPoint(x1+12, y2-8)
		snake.AddPoint(x1+12, y2)
	} else {
		snake.AddPoint(x1, y2)
	}
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontal) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile1(), stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

// ftileSwitchWithManyLinksConnectionVerticalTop is the arrow from the test
// diamond to an inner case.
type ftileSwitchWithManyLinksConnectionVerticalTop struct {
	AbstractConnection
	outer  *FtileSwitchWithManyLinks
	branch *Branch
}

func (c *ftileSwitchWithManyLinksConnectionVerticalTop) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
	translateDiamond1 := c.outer.translateDiamond1(stringBounder)
	p1b := translateDiamond1.Translated(dimDiamond1.PointB())
	p1c := translateDiamond1.Translated(dimDiamond1.PointC())
	p1d := translateDiamond1.Translated(dimDiamond1.PointD())
	p2 := c.p2(stringBounder)
	x2 := p2.X
	y2 := p2.Y
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabelVertical(c.branch.TextBlockPositive(), geom.VerticalAlignmentCenter)
	if x2 < p1d.X-ftileSwitchWithManyLinksMargin || x2 > p1b.X+ftileSwitchWithManyLinksMargin {
		snake.AddPoint(x2, p1d.Y)
		snake.AddPoint(x2, y2)
	} else {
		x1 := p1c.X
		y1 := p1c.Y
		ym := (y1*2 + y2) / 3
		snake.AddPoint(x1, y1)
		snake.AddPoint(x1, ym)
		snake.AddPoint(x2, ym)
		snake.AddPoint(x2, y2)
	}
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionVerticalTop) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile2(), stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

// ftileSwitchWithManyLinksConnectionVerticalBottom is the arrow from an
// inner case to the merge diamond.
type ftileSwitchWithManyLinksConnectionVerticalBottom struct {
	AbstractConnection
	outer    *FtileSwitchWithManyLinks
	outLabel klimt.TextBlock
}

func (c *ftileSwitchWithManyLinksConnectionVerticalBottom) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	dimDiamond2 := c.outer.diamond2.Geometry(stringBounder)
	translateDiamond2 := c.outer.translateDiamond2(stringBounder)
	p2a := translateDiamond2.Translated(dimDiamond2.PointA())
	translateDiamond2.Translated(dimDiamond2.PointB())
	p2d := translateDiamond2.Translated(dimDiamond2.PointD())

	dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
	translateDiamond1 := c.outer.translateDiamond1(stringBounder)
	p1b := translateDiamond1.Translated(dimDiamond1.PointB())
	translateDiamond1.Translated(dimDiamond1.PointC())
	p1d := translateDiamond1.Translated(dimDiamond1.PointD())

	x1 := p1.X
	y1 := p1.Y
	x2 := p2a.X
	y2 := p2a.Y
	ym := (y1 + y2) / 2
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabelVertical(c.outLabel, geom.VerticalAlignmentCenter)
	if x1 < p1d.X-ftileSwitchWithManyLinksMargin || x1 > p1b.X+ftileSwitchWithManyLinksMargin {
		snake.AddPoint(x1, y1)
		snake.AddPoint(x1, p2d.Y)
	} else {
		snake.AddPoint(x1, y1)
		snake.AddPoint(x1, ym)
		snake.AddPoint(x2, ym)
		snake.AddPoint(x2, y2)
	}
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionVerticalBottom) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile1(), stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

// ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane is
// the arrow from the test diamond to a case in another lane.
type ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane struct {
	AbstractConnection
	outer  *FtileSwitchWithManyLinks
	branch *Branch
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown())
	snake.AddPointPt(c.p1(stringBounder))
	snake.AddPointPt(c.p2(stringBounder))
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane) DrawTranslate(ug klimt.UGraphic,
	translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	mp1a := translate1.Translated(p1)
	mp2b := translate2.Translated(p2)
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.branch.TextBlockPositive(), c.ArrowHorizontalAlignment())
	dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
	if mp1a.X > mp2b.X {
		snake.AddPoint(mp1a.X-(dimDiamond1.Width()/2), p1.Y-(dimDiamond1.Height()/2))
	} else {
		snake.AddPoint(mp1a.X+(dimDiamond1.Width()/2), p1.Y-(dimDiamond1.Height()/2))
	}
	snake.AddPoint(mp2b.X, p1.Y-(dimDiamond1.Height()/2))
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane) p1(
	stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateDiamond1(stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane) p2(
	stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile2(), stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

// ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane is
// the arrow from a case in another lane to the merge diamond.
type ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane struct {
	AbstractConnection
	outer *FtileSwitchWithManyLinks
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane) DrawU(klimt.UGraphic) {
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane) DrawTranslate(ug klimt.UGraphic,
	translate1, translate2 klimt.UTranslate) {
	stringBounder := ug.StringBounder()
	mp1a := translate1.Translated(c.p1(stringBounder))
	mp2b := translate2.Translated(c.p2(stringBounder))
	dimDiamond2 := c.outer.diamond2.Geometry(stringBounder)
	skinParam := c.outer.SkinParam()
	var arrow *klimt.UPolygon
	var direction klimt.Direction
	if mp1a.X > mp2b.X {
		arrow = skinParam.Arrows().AsToLeft()
		direction = klimt.DirectionLeft
	} else {
		arrow = skinParam.Arrows().AsToRight()
		direction = klimt.DirectionRight
	}
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, arrow)
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, mp2b.Y+(dimDiamond2.Height()/2))
	if direction == klimt.DirectionLeft {
		snake.AddPoint(mp2b.X+(dimDiamond2.Width()/2), mp2b.Y+(dimDiamond2.Height()/2))
	} else {
		snake.AddPoint(mp2b.X-(dimDiamond2.Width()/2), mp2b.Y+(dimDiamond2.Height()/2))
	}
	ug.Draw(snake)
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane) p1(
	stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile1(), stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane) p2(
	stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateDiamond2(stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

func (f *FtileSwitchWithManyLinks) differentSwimlane(ftile1, ftile2 Ftile) bool {
	var swimlane1, swimlane2 *Swimlane
	if ftile1 != nil {
		swimlane1 = ftile1.SwimlaneOut()
	}
	if ftile2 != nil {
		swimlane2 = ftile2.SwimlaneIn()
	}
	return swimlane1 != nil && swimlane2 != nil && swimlane1 != swimlane2
}

func (f *FtileSwitchWithManyLinks) ydelta1a(stringBounder font.StringBounder) float64 {
	maxV := 10.0
	for _, branch := range f.branches {
		maxV = math.Max(maxV, branch.TextBlockPositive().CalculateDimension(stringBounder).Height())
	}
	if f.mode == ftileSwitchWithDiamondsBigDiamond {
		diamondHeight := f.diamond1.CalculateDimension(stringBounder).Height()
		maxV += diamondHeight / 2
	}
	return maxV + 10
}

// AddLinks returns the switch with its arrows.
func (f *FtileSwitchWithManyLinks) AddLinks(stringBounder font.StringBounder) Ftile {
	var conns []Connection
	conns = f.addIngoingArrows(conns)
	conns = f.addOutgoingArrows(stringBounder, conns)
	return FtileUtilsAddConnections(f, conns)
}

func (f *FtileSwitchWithManyLinks) addIngoingArrows(conns []Connection) []Connection {
	last := len(f.tiles) - 1
	conns = append(conns, &ftileSwitchWithManyLinksConnectionHorizontalThenVertical{
		AbstractConnection: newAbstractConnection(f.diamond1, f.tiles[0]), outer: f, branch: f.branches[0]})
	conns = append(conns, &ftileSwitchWithManyLinksConnectionHorizontalThenVertical{
		AbstractConnection: newAbstractConnection(f.diamond1, f.tiles[last]), outer: f, branch: f.branches[last]})
	for i, tile := range f.tiles {
		if i > 0 && i < last {
			conns = append(conns, &ftileSwitchWithManyLinksConnectionVerticalTop{
				AbstractConnection: newAbstractConnection(f.diamond1, tile), outer: f, branch: f.branches[i]})
		}
		if f.differentSwimlane(f, tile) {
			conns = append(conns, &ftileSwitchWithManyLinksConnectionHorizontalThenVerticalCrossSwimlane{
				AbstractConnection: newAbstractConnection(f.diamond1, tile), outer: f, branch: f.branches[i]})
		}
	}
	return conns
}

func (f *FtileSwitchWithManyLinks) addOutgoingArrows(stringBounder font.StringBounder,
	conns []Connection) []Connection {
	lastOutgoingArrow := f.lastOutgoingArrow(stringBounder)
	if lastOutgoingArrow == -1 {
		return conns
	}
	firstOutgoingArrow := f.firstOutgoingArrow(stringBounder)
	if firstOutgoingArrow < len(f.tiles) {
		conns = append(conns, &ftileSwitchWithManyLinksConnectionVerticalThenHorizontal{
			AbstractConnection: newAbstractConnection(f.tiles[firstOutgoingArrow], f.diamond2), outer: f,
			outLabel: f.branches[firstOutgoingArrow].TextBlockSpecial()})
	}
	if lastOutgoingArrow > 0 {
		conns = append(conns, &ftileSwitchWithManyLinksConnectionVerticalThenHorizontal{
			AbstractConnection: newAbstractConnection(f.tiles[lastOutgoingArrow], f.diamond2), outer: f,
			outLabel: f.branches[lastOutgoingArrow].TextBlockSpecial()})
	}
	for i := firstOutgoingArrow + 1; i < lastOutgoingArrow; i++ {
		tile := f.tiles[i]
		if tile.Geometry(stringBounder).HasPointOut() {
			conns = append(conns, &ftileSwitchWithManyLinksConnectionVerticalBottom{
				AbstractConnection: newAbstractConnection(tile, f.diamond2), outer: f,
				outLabel: f.branches[i].TextBlockSpecial()})
		}
	}
	for _, tile := range f.tiles {
		origin := f.findLastWithPointOut(tile, stringBounder)
		if origin != nil && f.differentSwimlane(origin, f.diamond2) {
			conns = append(conns, &ftileSwitchWithManyLinksConnectionVerticalThenHorizontalCrossSwimlane{
				AbstractConnection: newAbstractConnection(origin, f.diamond2), outer: f})
		}
	}
	return conns
}

func (f *FtileSwitchWithManyLinks) findLastWithPointOut(tile Ftile, stringBounder font.StringBounder) Ftile {
	if tile.Geometry(stringBounder).HasPointOut() {
		return tile
	}
	for _, child := range tile.MyChildren() {
		if found := f.findLastWithPointOut(child, stringBounder); found != nil {
			return found
		}
	}
	return nil
}

func (f *FtileSwitchWithManyLinks) firstOutgoingArrow(stringBounder font.StringBounder) int {
	for i := 0; i < len(f.tiles)-1; i++ {
		tile := f.tiles[i]
		if tile.Geometry(stringBounder).HasPointOut() && !f.differentSwimlane(f, tile) {
			return i
		}
	}
	return len(f.tiles)
}

func (f *FtileSwitchWithManyLinks) lastOutgoingArrow(stringBounder font.StringBounder) int {
	for i := len(f.tiles) - 1; i >= 0; i-- {
		tile := f.tiles[i]
		if tile.Geometry(stringBounder).HasPointOut() && !f.differentSwimlane(f, tile) {
			return i
		}
	}
	return -1
}
