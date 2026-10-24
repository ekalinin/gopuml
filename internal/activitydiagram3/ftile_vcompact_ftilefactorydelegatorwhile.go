// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorWhile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// FtileFactoryDelegatorWhile builds the tiles of "while" loops.
type FtileFactoryDelegatorWhile struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorWhile returns the delegator over factory.
func NewFtileFactoryDelegatorWhile(factory FtileFactory) *FtileFactoryDelegatorWhile {
	return &FtileFactoryDelegatorWhile{newFtileFactoryDelegator(factory)}
}

// CreateWhile implements FtileFactory.
func (d *FtileFactoryDelegatorWhile) CreateWhile(outColor *LinkRendering, swimlane *Swimlane, whileBlock Ftile,
	test, yes *klimt.Display, c color.HColor, specialOut Instruction, backward Ftile, incoming1,
	incoming2 *LinkRendering, currentStyleBuilder *klimt.StyleBuilder) Ftile {
	conditionStyle := d.SkinParam().ConditionStyle()

	styleArrow := d.DefaultStyleDefinitionArrow().MergedStyle(currentStyleBuilder)
	styleDiamond := d.DefaultStyleDefinitionDiamond().MergedStyle(currentStyleBuilder)
	borderColor := styleDiamond.Value(klimt.PNameLineColor).AsColor(d.SkinParam().IHtmlColorSet())
	backColor := c
	if c == nil {
		backColor = styleDiamond.Value(klimt.PNameBackGroundColor).AsColor(d.SkinParam().IHtmlColorSet())
	}
	arrowColor := RainbowFromStyle(styleArrow, d.SkinParam().IHtmlColorSet())
	fontArrow := styleArrow.FontConfiguration(d.SkinParam().IHtmlColorSet(), nil)
	fcTest := styleDiamond.FontConfiguration(d.SkinParam().IHtmlColorSet(), nil)

	incoming1 = ftileFactoryDelegatorWhileEnsureColor(incoming1, arrowColor)
	incoming2 = ftileFactoryDelegatorWhileEnsureColor(incoming2, arrowColor)
	outColor = ftileFactoryDelegatorWhileEnsureColor(outColor, arrowColor)

	result := FtileWhileCreate(outColor, swimlane, whileBlock, test, borderColor, backColor, arrowColor, yes,
		fontArrow, d.Factory(), conditionStyle, fcTest, specialOut, backward, incoming1, incoming2)

	weldingPoints := whileBlock.WeldingPoints()
	if len(weldingPoints) > 0 {
		genealogy := NewGenealogy(result)
		for _, w := range weldingPoints {
			ftileBreak, ok := w.(*FtileBreak)
			if !ok {
				uml.Throw(uml.ClassCastException, "")
			}
			result = FtileUtilsAddConnection(result, &ftileFactoryDelegatorWhileBreakConnection{
				genealogy: genealogy, ftileBreak: ftileBreak, skinParam: d.SkinParam(), arrowColor: arrowColor})
		}
	}
	return result
}

func ftileFactoryDelegatorWhileEnsureColor(link *LinkRendering, c *Rainbow) *LinkRendering {
	if link.Rainbow().Size() == 0 {
		return link.WithRainbow(c)
	}
	return link
}

// ftileFactoryDelegatorWhileBreakConnection is the arrow from a "break" to
// the exit of the loop (upstream's anonymous Connection).
type ftileFactoryDelegatorWhileBreakConnection struct {
	genealogy  *Genealogy
	ftileBreak *FtileBreak
	skinParam  klimt.ISkinParam
	arrowColor *Rainbow
}

func (c *ftileFactoryDelegatorWhileBreakConnection) DrawU(ug klimt.UGraphic) {
	tr1 := c.genealogy.Translate(c.ftileBreak, ug.StringBounder())
	snake := SnakeCreateEnd(c.skinParam, c.arrowColor, c.skinParam.Arrows().AsToLeft())
	snake.AddPoint(tr1.Dx(), tr1.Dy())
	snake.AddPoint(hexagonHalfSize, tr1.Dy())
	ug.Draw(snake)
}

func (c *ftileFactoryDelegatorWhileBreakConnection) Ftile1() Ftile { return c.ftileBreak }

func (c *ftileFactoryDelegatorWhileBreakConnection) Ftile2() Ftile { return nil }
