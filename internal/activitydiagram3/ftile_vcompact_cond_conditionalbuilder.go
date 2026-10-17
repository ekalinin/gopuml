// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/ConditionalBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ConditionalBuilder builds the tile of an "if" with a single test: the
// branches side by side, or one branch below the diamond when the other is
// empty or only stops.
type ConditionalBuilder struct {
	swimlane          *Swimlane
	borderColor       color.HColor
	backColor         color.HColor
	diamondLineBreak  klimt.LineBreakStrategy
	labelLineBreak    klimt.LineBreakStrategy
	arrowColor        *Rainbow
	ftileFactory      FtileFactory
	conditionStyle    klimt.ConditionStyle
	conditionEndStyle klimt.ConditionEndStyle
	branch1           *Branch
	branch2           *Branch
	skinParam         klimt.ISkinParam
	stringBounder     font.StringBounder
	fontArrow         *klimt.FontConfiguration
	styleDiamonFont   *klimt.FontConfiguration

	tile1        Ftile
	tile2        Ftile
	url          *klimt.URL
	notes        []*PositionedNote
	styleDiamond *klimt.Style
}

func conditionalBuilderStyleSignatureDiamond() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity, klimt.SNameDiamond)
}

// NewConditionalBuilder returns the builder of an "if"; backColor must not
// be nil.
func NewConditionalBuilder(swimlane *Swimlane, backColor color.HColor, ftileFactory FtileFactory,
	conditionStyle klimt.ConditionStyle, conditionEndStyle klimt.ConditionEndStyle, branch1, branch2 *Branch,
	skinParam klimt.ISkinParam, stringBounder font.StringBounder, url *klimt.URL, styleArrow,
	styleDiamond *klimt.Style, notes []*PositionedNote) *ConditionalBuilder {
	if backColor == nil {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &ConditionalBuilder{
		styleDiamond:      styleDiamond,
		styleDiamonFont:   styleDiamond.FontConfiguration(skinParam.IHtmlColorSet(), nil),
		fontArrow:         styleArrow.FontConfiguration(skinParam.IHtmlColorSet(), nil),
		diamondLineBreak:  styleDiamond.WrapWidth(),
		labelLineBreak:    styleArrow.WrapWidth(),
		backColor:         backColor,
		notes:             notes,
		borderColor:       styleDiamond.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet()),
		arrowColor:        RainbowFromStyle(styleArrow, skinParam.IHtmlColorSet()),
		ftileFactory:      ftileFactory,
		swimlane:          swimlane,
		conditionStyle:    conditionStyle,
		conditionEndStyle: conditionEndStyle,
		branch1:           branch1,
		branch2:           branch2,
		skinParam:         skinParam,
		stringBounder:     stringBounder,
		url:               url,
		tile1:             NewFtileMinWidthCentered(branch1.Ftile(), 30),
		tile2:             NewFtileMinWidthCentered(branch2.Ftile(), 30),
	}
}

// ConditionalBuilderCreate returns the tile of an "if" with a single test.
func ConditionalBuilderCreate(swimlane *Swimlane, backColor color.HColor, ftileFactory FtileFactory,
	conditionStyle klimt.ConditionStyle, conditionEndStyle klimt.ConditionEndStyle, branch1, branch2 *Branch,
	skinParam klimt.ISkinParam, stringBounder font.StringBounder, url *klimt.URL, styleArrow,
	styleDiamond *klimt.Style, notes []*PositionedNote) Ftile {
	builder := NewConditionalBuilder(swimlane, backColor, ftileFactory, conditionStyle, conditionEndStyle, branch1,
		branch2, skinParam, stringBounder, url, styleArrow, styleDiamond, notes)
	if conditionalBuilderIsEmptyOrOnlySingleStopOrSpot(branch2) &&
		!conditionalBuilderIsEmptyOrOnlySingleStopOrSpot(branch1) {
		return builder.createDown(builder.branch1, builder.branch2)
	}
	if branch1.IsEmpty() && branch2.IsOnlySingleStopOrSpot() {
		return builder.createDown(builder.branch1, builder.branch2)
	}
	if conditionalBuilderIsEmptyOrOnlySingleStopOrSpot(branch1) &&
		!conditionalBuilderIsEmptyOrOnlySingleStopOrSpot(branch2) {
		return builder.createDown(builder.branch2, builder.branch1)
	}
	if branch2.IsEmpty() && branch1.IsOnlySingleStopOrSpot() {
		return builder.createDown(builder.branch2, builder.branch1)
	}
	return builder.createWithLinks()
}

func conditionalBuilderIsEmptyOrOnlySingleStopOrSpot(branch *Branch) bool {
	return branch.IsEmpty() || branch.IsOnlySingleStopOrSpot()
}

func (b *ConditionalBuilder) createDown(branch1, branch2 *Branch) Ftile {
	tile1 := NewFtileMinWidthCentered(branch1.Ftile(), 30)
	tile2 := NewFtileMinWidthCentered(branch2.Ftile(), 30)
	tb1 := b.labelPositive(branch1)
	tb2 := b.labelPositive(branch2)
	diamond1 := b.shape1(false, tb1, tb2)
	diamond2 := b.shape2(branch1, branch2, true)
	if branch2.IsOnlySingleStopOrSpot() {
		return FtileIfDownCreate(diamond1, diamond2, b.swimlane, FtileUtilsAddHorizontalMargin(tile1, 10),
			b.arrowColor, b.conditionEndStyle, b.ftileFactory, branch2.Ftile(), branch2.Out(), b.notes)
	}
	if branch1.IsOnlySingleStopOrSpot() {
		return FtileIfDownCreate(diamond1, diamond2, b.swimlane, FtileUtilsAddHorizontalMargin(tile2, 10),
			b.arrowColor, b.conditionEndStyle, b.ftileFactory, branch1.Ftile(), branch1.Out(), b.notes)
	}
	if branch1.IsEmpty() {
		return FtileIfDownCreate(diamond1, diamond2, b.swimlane, FtileUtilsAddHorizontalMargin(tile2, 10),
			b.arrowColor, b.conditionEndStyle, b.ftileFactory, nil, nil, b.notes)
	}
	return FtileIfDownCreate(diamond1, diamond2, b.swimlane, FtileUtilsAddHorizontalMargin(tile1, 10),
		b.arrowColor, b.conditionEndStyle, b.ftileFactory, nil, branch2.Out(), b.notes)
}

func (b *ConditionalBuilder) createNude() Ftile { return NewFtileIfNude(b.tile1, b.tile2, b.swimlane) }

func (b *ConditionalBuilder) createWithDiamonds() Ftile {
	diamond1 := b.diamond1(true)
	diamond2 := b.shape2(b.branch1, b.branch2, false)
	ftile := NewFtileIfWithDiamonds(diamond1, b.tile1, b.tile2, diamond2, b.swimlane, b.stringBounder, b.notes)
	label1 := b.labelPositive(b.branch1).CalculateDimension(b.stringBounder)
	label2 := b.labelPositive(b.branch2).CalculateDimension(b.stringBounder)
	diff1 := ftile.ComputeMarginNeedForBranchLabe1(b.stringBounder, label1)
	diff2 := ftile.ComputeMarginNeedForBranchLabe2(b.stringBounder, label2)
	result := FtileUtilsAddHorizontalMargin2(ftile, diff1, diff2)
	suppHeight := ftile.ComputeVerticalMarginNeedForBranchs(b.stringBounder, label1, label2)
	return FtileUtilsAddVerticalMargin(result, suppHeight, 0)
}

func (b *ConditionalBuilder) createWithLinks() Ftile {
	diamond1 := b.diamond1(true)
	if b.url != nil {
		diamond1 = NewFtileWithUrl(diamond1, b.url)
	}
	diamond2 := b.shape2(b.branch1, b.branch2, false)
	tmp1 := FtileUtilsAddHorizontalMargin(b.tile1, 10)
	tmp2 := FtileUtilsAddHorizontalMargin(b.tile2, 10)
	ftile := NewFtileIfWithLinks(diamond1, tmp1, tmp2, diamond2, b.swimlane, b.arrowColor, b.conditionEndStyle,
		b.stringBounder, b.notes)
	label1 := b.labelPositive(b.branch1).CalculateDimension(b.stringBounder)
	label2 := b.labelPositive(b.branch2).CalculateDimension(b.stringBounder)
	diff1 := ftile.ComputeMarginNeedForBranchLabe1(b.stringBounder, label1)
	diff2 := ftile.ComputeMarginNeedForBranchLabe2(b.stringBounder, label2)
	suppHeight := ftile.ComputeVerticalMarginNeedForBranchs(b.stringBounder, label1, label2)
	result := ftile.AddLinks(b.branch1, b.branch2, b.stringBounder)
	result = FtileUtilsAddHorizontalMargin2(result, diff1, diff2)
	return FtileUtilsAddVerticalMargin(result, suppHeight, 0)
}

func (b *ConditionalBuilder) diamond1(eastWest bool) Ftile {
	return b.shape1(eastWest, b.labelPositive(b.branch1), b.labelPositive(b.branch2))
}

func (b *ConditionalBuilder) shape1(eastWest bool, tb1, tb2 klimt.TextBlock) Ftile {
	labelTest := b.branch1.LabelTest()
	horizontalAlignment := b.styleDiamond.HorizontalAlignment()
	sheet := b.skinParam.Sheet(b.styleDiamonFont, horizontalAlignment, klimt.CreoleModeFull).CreateSheet(labelTest)
	sheetBlock1 := klimt.NewSheetBlock1Padding(sheet, b.diamondLineBreak, b.skinParam.Padding())
	thickness := conditionalBuilderStyleSignatureDiamond().MergedStyle(b.skinParam.CurrentStyleBuilder()).Stroke()
	tbTest := klimt.NewSheetBlock2(sheetBlock1, HexagonAsStencil(sheetBlock1), thickness)

	switch b.conditionStyle {
	case klimt.ConditionStyleInsideHexagon:
		d := NewFtileDiamondInside(tbTest, b.tile1.SkinParam(), b.backColor, b.borderColor, b.swimlane)
		if eastWest {
			return d.WithWestAndEast(tb1, tb2)
		}
		return d.WithSouth(tb1).WithEast(tb2)
	case klimt.ConditionStyleEmptyDiamond:
		d := NewFtileDiamond(b.tile1.SkinParam(), b.backColor, b.borderColor, b.swimlane).WithNorth(tbTest)
		if eastWest {
			return d.WithWestAndEast(tb1, tb2)
		}
		return d.WithSouth(tb1).WithEast(tb2)
	case klimt.ConditionStyleInsideDiamond:
		d := NewFtileDiamondSquare(tbTest, b.tile1.SkinParam(), b.backColor, b.borderColor, b.swimlane)
		if eastWest {
			return d.WithWestAndEast(tb1, tb2)
		}
		return d.WithSouth(tb1).WithEast(tb2)
	}
	uml.Throw(uml.IllegalStateException, "")
	return nil
}

func (b *ConditionalBuilder) labelPositive(branch *Branch) klimt.TextBlock {
	return branch.DisplayPositive().Create0(b.fontArrow, geom.HorizontalAlignmentLeft, b.ftileFactory.SkinParam(),
		b.labelLineBreak, klimt.CreoleModeSimpleLine, nil, nil)
}

func (b *ConditionalBuilder) shape2(branch1, branch2 *Branch, useNorth bool) Ftile {
	if b.conditionEndStyle == klimt.ConditionEndStyleHline {
		return NewFtileEmpty(b.tile1.SkinParam(), 0, hexagonHalfSize, b.swimlane)
	}
	// else use default ConditionEndStyle.DIAMOND
	if b.HasTwoBranches() {
		var tbout1, tbout2 klimt.TextBlock
		if out1 := branch1.Ftile().OutLinkRendering().Display(); out1 != nil {
			tbout1 = out1.Create7(b.fontArrow, geom.HorizontalAlignmentLeft, b.ftileFactory.SkinParam(),
				klimt.CreoleModeSimpleLine)
		}
		if out2 := branch2.Ftile().OutLinkRendering().Display(); out2 != nil {
			tbout2 = out2.Create7(b.fontArrow, geom.HorizontalAlignmentLeft, b.ftileFactory.SkinParam(),
				klimt.CreoleModeSimpleLine)
		}
		c := branch2.Color()
		if c == nil {
			c = b.backColor
		}
		tmp := NewFtileDiamond(b.tile1.SkinParam(), c, b.borderColor, b.swimlane)
		if useNorth {
			tmp = tmp.WithNorth(tbout1)
		} else {
			tmp = tmp.WithWest(tbout1)
		}
		return tmp.WithEast(tbout2)
	}
	return NewFtileEmpty(b.tile1.SkinParam(), 0, hexagonHalfSize/2, b.swimlane)
}

// HasTwoBranches reports whether both branches have an exit point.
func (b *ConditionalBuilder) HasTwoBranches() bool {
	return b.tile1.Geometry(b.stringBounder).HasPointOut() && b.tile2.Geometry(b.stringBounder).HasPointOut()
}
