// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorRepeat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// FtileFactoryDelegatorRepeat builds the tiles of "repeat" loops.
type FtileFactoryDelegatorRepeat struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorRepeat returns the delegator over factory.
func NewFtileFactoryDelegatorRepeat(factory FtileFactory) *FtileFactoryDelegatorRepeat {
	return &FtileFactoryDelegatorRepeat{newFtileFactoryDelegator(factory)}
}

// ftileFactoryDelegatorRepeatColor is upstream's NoSuchColorException
// propagation of a color lookup.
func ftileFactoryDelegatorRepeatColor(c color.HColor, err error) color.HColor {
	if err != nil {
		throwNoSuchColor(err)
	}
	return c
}

// Repeat implements FtileFactory.
func (d *FtileFactoryDelegatorRepeat) Repeat(stereogroup1, stereogroup2 *klimt.Stereogroup, boxStyleIn BoxStyle,
	swimlane, swimlaneOut *Swimlane, startLabel *klimt.Display, repeat Ftile, test, yes, out *klimt.Display,
	backward Ftile, noOut bool, incoming1, incoming2 *LinkRendering, currentStyleBuilder *klimt.StyleBuilder) Ftile {
	if stereogroup1 == nil || stereogroup2 == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	conditionStyle := d.SkinParam().ConditionStyle()
	colorSet := d.SkinParam().IHtmlColorSet()

	styleArrow := d.DefaultStyleDefinitionArrow().MergedStyle(currentStyleBuilder)
	styleDiamond := d.DefaultStyleDefinitionDiamond().MergedStyle(currentStyleBuilder)

	borderColor := styleDiamond.Value(klimt.PNameLineColor).AsColor(colorSet)

	diamondColor1 := ftileFactoryDelegatorRepeatColor(stereogroup1.HColor(d.DefaultStyleDefinitionDiamond(),
		klimt.PNameBackGroundColor, currentStyleBuilder, colorSet))
	diamondColor2 := ftileFactoryDelegatorRepeatColor(stereogroup2.HColor(d.DefaultStyleDefinitionDiamond(),
		klimt.PNameBackGroundColor, currentStyleBuilder, colorSet))

	innerColors2, err := stereogroup2.InnerColors(colorSet)
	if err != nil {
		throwNoSuchColor(err)
	}
	if innerColors2.IsEmpty() && stereogroup2.IsEmpty() {
		diamondColor2 = diamondColor1
	}

	colorsEntry := d.DefaultStyleDefinitionActivity().MergedStyle(currentStyleBuilder).ToColors(colorSet).
		Add(color.TypeBack, diamondColor1)

	arrowColor := RainbowFromStyle(styleArrow, colorSet)
	fcDiamond := styleDiamond.FontConfiguration(colorSet, nil)
	fcArrow := styleArrow.FontConfiguration(colorSet, nil)

	endRepeatLinkRendering := repeat.OutLinkRendering()
	var endRepeatLinkColor *Rainbow
	if endRepeatLinkRendering != nil {
		endRepeatLinkColor = endRepeatLinkRendering.Rainbow()
	}

	entry := d.entry(swimlane, startLabel, colorsEntry, boxStyleIn, stereogroup1.BuildStereotype())

	result := FtileRepeatCreate(swimlane, swimlaneOut, entry, repeat, test, yes, out, borderColor, diamondColor1,
		diamondColor2, arrowColor, endRepeatLinkColor, conditionStyle, d.SkinParam(), fcDiamond, fcArrow, backward,
		noOut, incoming1, incoming2)

	weldingPoints := repeat.WeldingPoints()
	if len(weldingPoints) > 0 {
		diamondBreak := NewFtileDiamond(repeat.SkinParam(), diamondColor2, borderColor, swimlane)
		result = d.Assembly(FtileUtilsAddHorizontalMargin2(result, 10, 0), diamondBreak)
		genealogy := NewGenealogy(result)

		connections := make([]Connection, 0, len(weldingPoints))
		for i, w := range weldingPoints {
			ftileBreak, ok := w.(*FtileBreak)
			if !ok {
				uml.Throw(uml.ClassCastException, "")
			}
			connections = append(connections, &ftileFactoryDelegatorRepeatBreakConnection{genealogy: genealogy,
				ftileBreak: ftileBreak, diamondBreak: diamondBreak, first: i == 0, skinParam: d.SkinParam(),
				arrowColor: arrowColor})
		}
		result = FtileUtilsAddConnections(result, connections)
	}
	return result
}

func (d *FtileFactoryDelegatorRepeat) entry(swimlane *Swimlane, startLabel *klimt.Display, colors *color.Colors,
	boxStyleIn BoxStyle, stereotype *klimt.Stereotype) Ftile {
	if klimt.DisplayIsNull(startLabel) {
		return nil
	}
	return d.Activity(startLabel, swimlane, boxStyleIn, colors, stereotype, nil)
}

// ftileFactoryDelegatorRepeatBreakConnection is the arrow from a "break"
// to the diamond after the loop (upstream's anonymous Connection).
type ftileFactoryDelegatorRepeatBreakConnection struct {
	genealogy    *Genealogy
	ftileBreak   *FtileBreak
	diamondBreak *FtileDiamond
	first        bool
	skinParam    klimt.ISkinParam
	arrowColor   *Rainbow
}

func (c *ftileFactoryDelegatorRepeatBreakConnection) DrawU(ug klimt.UGraphic) {
	tr1 := c.genealogy.Translate(c.ftileBreak, ug.StringBounder())
	tr2 := c.genealogy.Translate(c.diamondBreak, ug.StringBounder())
	dimDiamond := c.diamondBreak.CalculateDimension(ug.StringBounder())

	var snake *Snake
	if c.first {
		snake = SnakeCreateEnd(c.skinParam, c.arrowColor, c.skinParam.Arrows().AsToRight())
		snake.AddPoint(tr1.Dx(), tr1.Dy())
		snake.AddPoint(0, tr1.Dy())
		snake.AddPoint(0, tr2.Dy()+dimDiamond.Height()/2)
		snake.AddPoint(tr2.Dx(), tr2.Dy()+dimDiamond.Height()/2)
	} else {
		snake = SnakeCreateEnd(c.skinParam, c.arrowColor, c.skinParam.Arrows().AsToLeft())
		snake.AddPoint(tr1.Dx(), tr1.Dy())
		snake.AddPoint(0, tr1.Dy())
	}
	ug.Draw(snake)
}

func (c *ftileFactoryDelegatorRepeatBreakConnection) Ftile1() Ftile { return c.ftileBreak }

func (c *ftileFactoryDelegatorRepeatBreakConnection) Ftile2() Ftile { return c.diamondBreak }
