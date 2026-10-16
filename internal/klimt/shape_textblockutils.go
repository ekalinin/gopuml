// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockMemoized.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"reflect"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// TextBlockMemoized caches the dimension of a text block whose computation
// is expensive. The cache is invalidated when the StringBounder type
// changes, as upstream keys it by class. Text blocks belong to one render,
// so the cache is not synchronized.
type TextBlockMemoized struct {
	cached     *geom.XDimension2D
	lastCaller reflect.Type
}

// Dimension returns the cached dimension, computing it with slow when the
// cache is empty or the bounder type changed.
func (m *TextBlockMemoized) Dimension(stringBounder font.StringBounder,
	slow func(font.StringBounder) geom.XDimension2D) geom.XDimension2D {
	currentCaller := reflect.TypeOf(stringBounder)
	if m.cached == nil || m.lastCaller != currentCaller {
		d := slow(stringBounder)
		m.cached = &d
		m.lastCaller = currentCaller
	}
	return *m.cached
}

// InvalidateDimensionCache empties the cache.
func (m *TextBlockMemoized) InvalidateDimensionCache() {
	m.cached = nil
	m.lastCaller = nil
}

// textBlockEmptyFixed is the anonymous empty block of upstream's
// TextBlockUtils.empty.
type textBlockEmptyFixed struct {
	TextBlockDefaults
	width, height float64
}

func (t *textBlockEmptyFixed) DrawU(UGraphic) {}

func (t *textBlockEmptyFixed) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.width, t.height)
}

func (t *textBlockEmptyFixed) InnerPosition(string, font.StringBounder) *geom.XRectangle2D {
	return nil
}

// emptyTextBlock is upstream's TextBlockUtils.EMPTY_TEXT_BLOCK: a single
// read-only value, compared by identity.
var emptyTextBlock TextBlock = &textBlockEmptyFixed{}

// TextBlockUtilsEmptyTextBlock returns upstream's EMPTY_TEXT_BLOCK, the 0x0
// block that mergeLR and mergeTB skip.
func TextBlockUtilsEmptyTextBlock() TextBlock { return emptyTextBlock }

// TextBlockUtilsBordered returns textBlock in a box.
func TextBlockUtilsBordered(textBlock TextBlock, stroke *UStroke, borderColor, backgroundColor color.HColor,
	cornersize float64, margins ClockwiseTopRightBottomLeft, id string) TextBlock {
	return newTextBlockBordered(textBlock, stroke, borderColor, backgroundColor, cornersize, margins, id)
}

// TextBlockUtilsWithMargin returns textBlock with marginX on the left and
// right and marginY at the top and bottom; textBlock itself for no margin.
func TextBlockUtilsWithMargin(textBlock TextBlock, marginX, marginY float64) TextBlock {
	if marginX == 0 && marginY == 0 {
		return textBlock
	}
	return newTextBlockMarged(textBlock, marginY, marginX, marginY, marginX)
}

// TextBlockUtilsWithMargins returns textBlock with margins.
func TextBlockUtilsWithMargins(textBlock TextBlock, margins ClockwiseTopRightBottomLeft) TextBlock {
	return newTextBlockMargedClockwise(textBlock, margins)
}

// TextBlockUtilsWithMargin4 returns textBlock with four margins.
func TextBlockUtilsWithMargin4(textBlock TextBlock, marginX1, marginX2, marginY1, marginY2 float64) TextBlock {
	return newTextBlockMarged(textBlock, marginY1, marginX2, marginY2, marginX1)
}

// TextBlockUtilsWithMinWidth returns textBlock at least minWidth wide.
func TextBlockUtilsWithMinWidth(textBlock TextBlock, minWidth float64,
	horizontalAlignment geom.HorizontalAlignment) TextBlock {
	return &textBlockMinWidth{textBlock: textBlock, minWidth: minWidth, horizontalAlignment: horizontalAlignment}
}

// TextBlockUtilsEmpty returns an empty block of a size.
func TextBlockUtilsEmpty(width, height float64) TextBlock {
	return &textBlockEmptyFixed{width: width, height: height}
}

// TextBlockUtilsAsPositionable returns the block placed at pt.
func TextBlockUtilsAsPositionable(textBlock TextBlock, stringBounder font.StringBounder,
	pt geom.XPoint2D) geom.Positionable {
	return geom.PositionableImplCreate(pt, textBlock.CalculateDimension(stringBounder))
}

// TextBlockUtilsAsPositionableDim returns the dimension placed at pt.
func TextBlockUtilsAsPositionableDim(d geom.XDimension2D, pt geom.XPoint2D) geom.Positionable {
	return geom.PositionableImplCreate(pt, d)
}

// TextBlockUtilsMergeLR puts b2 to the right of b1, skipping
// EMPTY_TEXT_BLOCK.
func TextBlockUtilsMergeLR(b1, b2 TextBlock, verticalAlignment geom.VerticalAlignment) TextBlock {
	if b1 == emptyTextBlock {
		return b2
	}
	if b2 == emptyTextBlock {
		return b1
	}
	return &TextBlockHorizontal{blocks: []TextBlock{b1, b2}, alignment: verticalAlignment}
}

// TextBlockUtilsMergeTB puts b2 below b1, skipping EMPTY_TEXT_BLOCK.
func TextBlockUtilsMergeTB(b1, b2 TextBlock, horizontalAlignment geom.HorizontalAlignment) TextBlock {
	if b1 == emptyTextBlock {
		return b2
	}
	if b2 == emptyTextBlock {
		return b1
	}
	return &TextBlockVertical{blocks: []TextBlock{b1, b2}, horizontalAlignment: horizontalAlignment}
}

// TextBlockUtilsMinMax returns the bounds of what tb draws, measured with a
// LimitFinder.
func TextBlockUtilsMinMax(tb UDrawable, stringBounder font.StringBounder, initToZero bool) geom.MinMax {
	limitFinder := LimitFinderCreate(stringBounder, initToZero)
	tb.DrawU(limitFinder)
	return limitFinder.MinMax()
}

// TextBlockUtilsIsEmpty reports whether text is nil, EMPTY_TEXT_BLOCK or of
// size 0x0.
func TextBlockUtilsIsEmpty(text TextBlock, dummyStringBounder font.StringBounder) bool {
	if text == nil || text == emptyTextBlock {
		return true
	}
	d := text.CalculateDimension(dummyStringBounder)
	return d.Height() == 0 && d.Width() == 0
}

// textBlockBackcolored is the anonymous block of upstream's
// TextBlockUtils.addBackcolor.
type textBlockBackcolored struct {
	text      TextBlock
	backColor color.HColor
}

func (t *textBlockBackcolored) DrawU(ug UGraphic) { t.text.DrawU(ug) }

func (t *textBlockBackcolored) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return t.text.MinMax(stringBounder)
}

func (t *textBlockBackcolored) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return t.text.InnerPosition(member, stringBounder)
}

func (t *textBlockBackcolored) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.text.CalculateDimension(stringBounder)
}

func (t *textBlockBackcolored) Backcolor() color.HColor { return t.backColor }

func (t *textBlockBackcolored) MagneticBorder() MagneticBorder { return t.text.MagneticBorder() }

// TextBlockUtilsAddBackcolor returns text with a background color.
func TextBlockUtilsAddBackcolor(text TextBlock, backColor color.HColor) TextBlock {
	return &textBlockBackcolored{text: text, backColor: backColor}
}
