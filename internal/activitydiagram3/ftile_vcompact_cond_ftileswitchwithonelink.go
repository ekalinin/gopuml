// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileSwitchWithOneLink.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileSwitchWithOneLink is a switch with a single case: straight arrows
// from the test diamond to the case and from the case to the merge
// diamond.
type FtileSwitchWithOneLink struct {
	FtileSwitchWithDiamonds
	arrowColor *Rainbow
}

// NewFtileSwitchWithOneLink returns the switch of a single case.
func NewFtileSwitchWithOneLink(tiles []Ftile, branches []*Branch, in *Swimlane, diamond1, diamond2 Ftile,
	stringBounder font.StringBounder, arrowColor *Rainbow) *FtileSwitchWithOneLink {
	f := &FtileSwitchWithOneLink{arrowColor: arrowColor}
	f.initFtileSwitchWithDiamonds(f, tiles, branches, in, diamond1, diamond2, stringBounder)
	return f
}

// ftileSwitchWithOneLinkConnectionVerticalTop is the arrow from the test
// diamond to the case.
type ftileSwitchWithOneLinkConnectionVerticalTop struct {
	AbstractConnection
	outer  *FtileSwitchWithOneLink
	branch *Branch
}

func (c *ftileSwitchWithOneLinkConnectionVerticalTop) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown()).
		WithLabel(c.branch.TextBlockPositive(), c.ArrowHorizontalAlignment())
	snake.AddPoint(x2, y1)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileSwitchWithOneLinkConnectionVerticalTop) p1(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond1 := c.outer.diamond1.Geometry(stringBounder)
	return c.outer.translateDiamond1(stringBounder).Translated(dimDiamond1.PointC())
}

func (c *ftileSwitchWithOneLinkConnectionVerticalTop) p2(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile2(), stringBounder).Translated(c.Ftile2().Geometry(stringBounder).PointIn())
}

// ftileSwitchWithOneLinkConnectionVerticalBottom is the arrow from the case
// to the merge diamond.
type ftileSwitchWithOneLinkConnectionVerticalBottom struct {
	AbstractConnection
	outer *FtileSwitchWithOneLink
}

func (c *ftileSwitchWithOneLinkConnectionVerticalBottom) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	p1 := c.p1(stringBounder)
	p2 := c.p2(stringBounder)
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y
	skinParam := c.outer.SkinParam()
	snake := SnakeCreateEnd(skinParam, c.outer.arrowColor, skinParam.Arrows().AsToDown())
	snake.AddPoint(x2, y1)
	snake.AddPoint(x2, y2)
	ug.Draw(snake)
}

func (c *ftileSwitchWithOneLinkConnectionVerticalBottom) p1(stringBounder font.StringBounder) geom.XPoint2D {
	return c.outer.translateOf(c.Ftile1(), stringBounder).Translated(c.Ftile1().Geometry(stringBounder).PointOut())
}

func (c *ftileSwitchWithOneLinkConnectionVerticalBottom) p2(stringBounder font.StringBounder) geom.XPoint2D {
	dimDiamond2 := c.outer.diamond2.Geometry(stringBounder)
	return c.outer.translateDiamond2(stringBounder).Translated(dimDiamond2.PointA())
}

// AddLinks returns the switch with its arrows.
func (f *FtileSwitchWithOneLink) AddLinks(stringBounder font.StringBounder) Ftile {
	var conns []Connection
	single := f.tiles[0]
	conns = append(conns, &ftileSwitchWithOneLinkConnectionVerticalTop{
		AbstractConnection: newAbstractConnection(f.diamond1, single), outer: f, branch: f.branches[0]})
	if single.Geometry(stringBounder).HasPointOut() {
		conns = append(conns, &ftileSwitchWithOneLinkConnectionVerticalBottom{
			AbstractConnection: newAbstractConnection(single, f.diamond2), outer: f})
	}
	return FtileUtilsAddConnections(f, conns)
}
