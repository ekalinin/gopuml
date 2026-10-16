// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorIf.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// FtileFactoryDelegatorIf builds the tiles of "if": a conditional for one
// test, a long horizontal (or vertical) chain of diamonds for "elseif".
type FtileFactoryDelegatorIf struct {
	FtileFactoryDelegator
	pragma *klimt.Pragma
}

// NewFtileFactoryDelegatorIf returns the delegator over factory.
func NewFtileFactoryDelegatorIf(factory FtileFactory, pragma *klimt.Pragma) *FtileFactoryDelegatorIf {
	return &FtileFactoryDelegatorIf{FtileFactoryDelegator: newFtileFactoryDelegator(factory), pragma: pragma}
}

// CreateIf implements FtileFactory.
func (d *FtileFactoryDelegatorIf) CreateIf(swimlane *Swimlane, thens []*Branch, elseBranch *Branch,
	afterEndwhile, topInlinkRendering *LinkRendering, url *klimt.URL, notes []*PositionedNote,
	stereotype *klimt.Stereotype, currentStyleBuilder *klimt.StyleBuilder) Ftile {
	conditionStyle := d.SkinParam().ConditionStyle()
	conditionEndStyle := d.SkinParam().ConditionEndStyle()
	branch0 := thens[0]

	styleArrow := d.DefaultStyleDefinitionArrow().MergedStyle(currentStyleBuilder)
	styleDiamond := d.DefaultStyleDefinitionDiamond().WithTOBECHANGED(stereotype).MergedStyle(currentStyleBuilder)
	backColor := branch0.Color()
	if backColor == nil {
		backColor = styleDiamond.Value(klimt.PNameBackGroundColor).AsColor(d.SkinParam().IHtmlColorSet())
	}
	if len(thens) > 1 {
		if d.pragma.IsTrue(klimt.PragmaUseVerticalIf) {
			return FtileIfLongVerticalCreate(swimlane, backColor, d.Factory(), conditionStyle, thens, elseBranch,
				topInlinkRendering, afterEndwhile, styleArrow, styleDiamond)
		}
		return FtileIfLongHorizontalCreate(swimlane, backColor, d.Factory(), conditionStyle, thens, elseBranch,
			topInlinkRendering, afterEndwhile, styleArrow, styleDiamond)
	}
	return ConditionalBuilderCreate(swimlane, backColor, d.Factory(), conditionStyle, conditionEndStyle, thens[0],
		elseBranch, d.SkinParam(), d.StringBounder(), url, styleArrow, styleDiamond, notes)
}
