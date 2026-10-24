// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/AbstractParallelFtilesBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// abstractParallelFtilesBuilderBarHeight is the height of the bars.
const abstractParallelFtilesBuilderBarHeight = 6.0

// AbstractParallelFtilesBuilder holds the branches of a parallel block,
// each with its margins and the room for the labels of its arrows; the
// fork, merge and split builders add the bars and the arrows around them.
// Upstream's unused Rose instance is not ported.
type AbstractParallelFtilesBuilder struct {
	skinParam     klimt.ISkinParam
	stringBounder font.StringBounder
	list99        []Ftile
	colors        *color.Colors
}

// parallelFtilesBuilder is the part of the builders that
// FtileFactoryDelegatorCreateParallel uses.
type parallelFtilesBuilder interface {
	List99() []Ftile
	Build(inner Ftile) Ftile
}

func newAbstractParallelFtilesBuilder(skinParam klimt.ISkinParam, stringBounder font.StringBounder, all []Ftile,
	colors *color.Colors) AbstractParallelFtilesBuilder {
	b := AbstractParallelFtilesBuilder{skinParam: skinParam, stringBounder: stringBounder}
	b.list99 = append(b.list99, b.decorateAllTiles(all)...)
	b.colors = colors
	return b
}

// StyleSignature is the style signature of the bars and merge diamond.
func (b *AbstractParallelFtilesBuilder) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity)
}

// StyleSignatureArrow is the style signature of the arrows.
func (b *AbstractParallelFtilesBuilder) StyleSignatureArrow() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameArrow)
}

// List99 returns the decorated branches.
func (b *AbstractParallelFtilesBuilder) List99() []Ftile { return b.list99 }

func (b *AbstractParallelFtilesBuilder) decorateAllTiles(all []Ftile) []Ftile {
	maxHeight := b.computeMaxHeight(all)
	ymargin1 := b.suppSpace1(all, b.StringBounder())
	ymargin2 := b.suppSpace2(all, b.StringBounder())
	var result []Ftile
	for _, ftile := range all {
		result = append(result, b.computeNewFtile(ftile, maxHeight, ymargin1, ymargin2))
	}
	return result
}

func (b *AbstractParallelFtilesBuilder) suppSpace1(all []Ftile, stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, child := range all {
		text := b.TextBlock(child.InLinkRendering().Display())
		if text == nil {
			continue
		}
		result = math.Max(result, text.CalculateDimension(stringBounder).Height())
	}
	return result
}

func (b *AbstractParallelFtilesBuilder) suppSpace2(all []Ftile, stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, child := range all {
		text := b.TextBlock(child.OutLinkRendering().Display())
		if text == nil {
			continue
		}
		result = math.Max(result, text.CalculateDimension(stringBounder).Height())
	}
	return result
}

func (b *AbstractParallelFtilesBuilder) computeNewFtile(ftile Ftile, maxHeight, ymargin1, ymargin2 float64) Ftile {
	const spaceArroundBlackBar = 20
	const xMargin = 14
	tmp := FtileUtilsAddHorizontalMargin2(ftile, xMargin, xMargin+b.suppForIncomingArrow(ftile))
	tmp = NewFtileHeightFixedCentered(tmp, maxHeight+2*spaceArroundBlackBar)
	tmp = NewFtileHeightFixedMarged(ymargin1, tmp, ymargin2)
	return tmp
}

func (b *AbstractParallelFtilesBuilder) suppForIncomingArrow(ftile Ftile) float64 {
	x1 := b.xSuppForDisplay(ftile, ftile.InLinkRendering().Display())
	x2 := b.xSuppForDisplay(ftile, ftile.OutLinkRendering().Display())
	return math.Max(x1, x2)
}

func (b *AbstractParallelFtilesBuilder) xSuppForDisplay(ftile Ftile, label *klimt.Display) float64 {
	text := b.TextBlock(label)
	if text == nil {
		return 0
	}
	textWidth := text.CalculateDimension(b.StringBounder()).Width()
	ftileDim := ftile.Geometry(b.StringBounder())
	pos2 := ftileDim.Left() + textWidth
	if pos2 > ftileDim.Width() {
		return pos2 - ftileDim.Width()
	}
	return 0
}

func (b *AbstractParallelFtilesBuilder) computeMaxHeight(all []Ftile) float64 {
	height := 0.0
	for _, tmp := range all {
		height = math.Max(height, tmp.CalculateDimension(b.StringBounder()).Height())
	}
	return height
}

// StringBounder returns the string bounder of the builder.
func (b *AbstractParallelFtilesBuilder) StringBounder() font.StringBounder { return b.stringBounder }

// SkinParam returns the skin parameters of the builder.
func (b *AbstractParallelFtilesBuilder) SkinParam() klimt.ISkinParam { return b.skinParam }

// TextBlock is getTextBlock(Display): the label of an arrow, nil for the
// null display.
func (b *AbstractParallelFtilesBuilder) TextBlock(display *klimt.Display) klimt.TextBlock {
	// DUP3945
	if klimt.DisplayIsNull(display) {
		return nil
	}
	style := b.StyleSignatureArrow().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	fontConfiguration := style.FontConfiguration(b.SkinParam().IHtmlColorSet(), nil)
	return display.Create7(fontConfiguration, geom.HorizontalAlignmentLeft, b.SkinParam(),
		klimt.CreoleModeSimpleLine)
}

// TextBlockOfLinkRendering is getTextBlock(LinkRendering): the label of
// linkRendering.
func (b *AbstractParallelFtilesBuilder) TextBlockOfLinkRendering(linkRendering *LinkRendering) klimt.TextBlock {
	// DUP1433
	return b.TextBlock(linkRendering.Display())
}

func (b *AbstractParallelFtilesBuilder) heightOfMiddle(middle Ftile) float64 {
	return middle.CalculateDimension(b.StringBounder()).Height()
}

func (b *AbstractParallelFtilesBuilder) swimlaneOutForStep2() *Swimlane {
	return b.list99[len(b.list99)-1].SwimlaneOut()
}
