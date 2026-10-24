// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorSwitch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileFactoryDelegatorSwitch draws the "switch" blocks.
type FtileFactoryDelegatorSwitch struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorSwitch returns the delegator over factory.
func NewFtileFactoryDelegatorSwitch(factory FtileFactory) *FtileFactoryDelegatorSwitch {
	return &FtileFactoryDelegatorSwitch{newFtileFactoryDelegator(factory)}
}

// CreateSwitch implements FtileFactory: the cases between two diamonds,
// linked by arrows.
func (d *FtileFactoryDelegatorSwitch) CreateSwitch(swimlane *Swimlane, branches []*Branch, _,
	_ *LinkRendering, labelTest *klimt.Display, colors, endColors *color.Colors) Ftile {
	return d.createWithLinks(swimlane, branches, labelTest, colors, endColors)
}

// createNude is upstream's unused createNude.
func (d *FtileFactoryDelegatorSwitch) createNude(swimlane *Swimlane, branches []*Branch) Ftile {
	var ftiles []Ftile
	for _, branch := range branches {
		ftiles = append(ftiles, NewFtileMinWidthCentered(branch.Ftile(), 30))
	}
	return NewFtileSwitchNude(ftiles, swimlane)
}

// createWithDiamonds is upstream's unused createWithDiamonds.
func (d *FtileFactoryDelegatorSwitch) createWithDiamonds(swimlane *Swimlane, branches []*Branch,
	labelTest *klimt.Display, colors, endColors *color.Colors) Ftile {
	var ftiles []Ftile
	for _, branch := range branches {
		ftiles = append(ftiles, NewFtileMinWidthCentered(branch.Ftile(), 30))
	}
	diamond1 := d.diamond1(swimlane, branches[0], labelTest, colors)
	diamond2 := d.diamond2(swimlane, branches[0], endColors)
	return NewFtileSwitchWithDiamonds(ftiles, branches, swimlane, diamond1, diamond2, d.StringBounder())
}

func (d *FtileFactoryDelegatorSwitch) createWithLinks(swimlane *Swimlane, branches []*Branch,
	labelTest *klimt.Display, colors, endColors *color.Colors) Ftile {
	var ftiles []Ftile
	diamond1 := d.diamond1(swimlane, branches[0], labelTest, colors)
	diamond2 := d.diamond2(swimlane, branches[0], endColors)
	for _, branch := range branches {
		dimLabelIn := branch.TextBlockPositive().CalculateDimension(d.StringBounder())
		dimLabelOut := branch.TextBlockSpecial().CalculateDimension(d.StringBounder())
		ftiles = append(ftiles, NewFtileDecorateOutLabel(NewFtileDecorateInLabel(branch.Ftile(), dimLabelIn),
			dimLabelOut))
	}
	style := d.DefaultStyleDefinitionArrow().MergedStyle(d.SkinParam().CurrentStyleBuilder())
	arrowColor := RainbowFromStyle(style, d.SkinParam().IHtmlColorSet())
	if len(ftiles) == 1 {
		result := NewFtileSwitchWithOneLink(ftiles, branches, swimlane, diamond1, diamond2, d.StringBounder(),
			arrowColor)
		return result.AddLinks(d.StringBounder())
	}
	result := NewFtileSwitchWithManyLinks(ftiles, branches, swimlane, diamond1, diamond2, d.StringBounder(),
		arrowColor)
	return result.AddLinks(d.StringBounder())
}

func (d *FtileFactoryDelegatorSwitch) diamond1(swimlane *Swimlane, branch0 *Branch, test *klimt.Display,
	colors *color.Colors) Ftile {
	style := d.DefaultStyleDefinitionDiamond().MergedStyle(d.SkinParam().CurrentStyleBuilder())
	lineBreak := style.WrapWidth()
	fcDiamond := style.FontConfiguration(d.SkinParam().IHtmlColorSet(), nil)
	borderColor := style.Value(klimt.PNameLineColor).AsColor(d.SkinParam().IHtmlColorSet())
	backColor := branch0.Color()
	if backColor == nil {
		backColor = klimt.ColorsColorOf(colors, style, klimt.PNameBackGroundColor, d.SkinParam().IHtmlColorSet())
	}
	var tbTest klimt.TextBlock
	if klimt.DisplayIsNull(test) || test.IsWhite() {
		tbTest = klimt.TextBlockUtilsEmpty(0, 0)
	} else {
		tbTest = test.Create0(fcDiamond, branch0.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft),
			branch0.SkinParam(), lineBreak, klimt.CreoleModeFull, nil, nil)
	}
	return NewFtileDiamondInside(tbTest, branch0.SkinParam(), backColor, borderColor, swimlane)
}

func (d *FtileFactoryDelegatorSwitch) diamond2(swimlane *Swimlane, branch0 *Branch, endColors *color.Colors) Ftile {
	style := d.DefaultStyleDefinitionDiamond().MergedStyle(d.SkinParam().CurrentStyleBuilder())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(d.SkinParam().IHtmlColorSet())
	backColor := branch0.Color()
	if backColor == nil {
		backColor = klimt.ColorsColorOf(endColors, style, klimt.PNameBackGroundColor, d.SkinParam().IHtmlColorSet())
	}
	return NewFtileDiamondInside(klimt.TextBlockUtilsEmpty(0, 0), branch0.SkinParam(), backColor, borderColor,
		swimlane)
}
