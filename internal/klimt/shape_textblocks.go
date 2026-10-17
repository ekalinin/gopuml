// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockEmpty.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockMarged.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockMinWidth.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockHorizontal.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockVertical.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/SimpleTextBlockBordered.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockGeneric.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockSprited.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/TextBlockWithUrl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// TextBlockEmpty is an empty block of a fixed size.
type TextBlockEmpty struct {
	TextBlockDefaults
	width  float64
	height float64
}

// NewTextBlockEmpty returns an empty block of a size.
func NewTextBlockEmpty(width, height float64) *TextBlockEmpty {
	return &TextBlockEmpty{width: width, height: height}
}

// CalculateDimension returns the size.
func (t *TextBlockEmpty) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.width, t.height)
}

// DrawU draws nothing.
func (t *TextBlockEmpty) DrawU(UGraphic) {}

// textBlockMarged adds margins around a block.
type textBlockMarged struct {
	TextBlockDefaults
	textBlock                TextBlock
	top, right, bottom, left float64
}

var _ WithPorts = (*textBlockMarged)(nil)

func newTextBlockMarged(textBlock TextBlock, top, right, bottom, left float64) *textBlockMarged {
	return &textBlockMarged{textBlock: textBlock, top: top, right: right, bottom: bottom, left: left}
}

func newTextBlockMargedClockwise(textBlock TextBlock, margins ClockwiseTopRightBottomLeft) *textBlockMarged {
	return newTextBlockMarged(textBlock, margins.Top(), margins.Right(), margins.Bottom(), margins.Left())
}

func (t *textBlockMarged) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.textBlock.CalculateDimension(stringBounder)
	return d.DeltaWH(t.left+t.right, t.top+t.bottom)
}

func (t *textBlockMarged) DrawU(ug UGraphic) {
	d := t.CalculateDimension(ug.StringBounder())
	if d.Width() > 0 {
		ug.Draw(UEmptyCreate(d))
		translate := NewUTranslate(t.left, t.top)
		t.textBlock.DrawU(ug.Apply(translate))
	}
}

func (t *textBlockMarged) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	parent := t.textBlock.InnerPosition(member, stringBounder)
	if parent == nil {
		return nil
	}
	r := NewUTranslate(t.left, t.top).Apply(*parent)
	return &r
}

func (t *textBlockMarged) Ports(stringBounder font.StringBounder) *Ports {
	// Upstream casts without checking: a block without ports throws
	// ClassCastException.
	wp, ok := t.textBlock.(WithPorts)
	if !ok {
		throw("java.lang.ClassCastException", "")
	}
	return wp.Ports(stringBounder).TranslateY(t.top)
}

// textBlockMinWidth widens a block to a minimum width.
type textBlockMinWidth struct {
	TextBlockDefaults
	textBlock           TextBlock
	minWidth            float64
	horizontalAlignment geom.HorizontalAlignment
}

func (t *textBlockMinWidth) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.textBlock.CalculateDimension(stringBounder)
	return d.AtLeast(t.minWidth, 0)
}

func (t *textBlockMinWidth) DrawU(ug UGraphic) {
	switch t.horizontalAlignment {
	case geom.HorizontalAlignmentLeft:
		t.textBlock.DrawU(ug)
	case geom.HorizontalAlignmentCenter:
		dimText := t.textBlock.CalculateDimension(ug.StringBounder())
		dimFull := t.CalculateDimension(ug.StringBounder())
		diffx := dimFull.Width() - dimText.Width()
		t.textBlock.DrawU(ug.Apply(UTranslateDx(diffx / 2)))
	case geom.HorizontalAlignmentRight:
		dimText := t.textBlock.CalculateDimension(ug.StringBounder())
		dimFull := t.CalculateDimension(ug.StringBounder())
		diffx := dimFull.Width() - dimText.Width()
		t.textBlock.DrawU(ug.Apply(UTranslateDx(diffx)))
	default:
		throw(unsupportedOperation, "")
	}
}

// TextBlockHorizontal puts blocks side by side.
type TextBlockHorizontal struct {
	TextBlockDefaults
	memo      TextBlockMemoized
	blocks    []TextBlock
	alignment geom.VerticalAlignment
}

// NewTextBlockHorizontal returns the blocks side by side; it panics with
// fewer than two blocks, as upstream throws.
func NewTextBlockHorizontal(all []TextBlock, alignment geom.VerticalAlignment) *TextBlockHorizontal {
	if len(all) < 2 {
		throw(illegalArgument, "")
	}
	return &TextBlockHorizontal{blocks: append([]TextBlock(nil), all...), alignment: alignment}
}

// CalculateDimension returns the merged size, memoized.
func (t *TextBlockHorizontal) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, t.calculateDimensionSlow)
}

func (t *TextBlockHorizontal) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.blocks[0].CalculateDimension(stringBounder)
	for i := 1; i < len(t.blocks); i++ {
		d = d.MergeLR(t.blocks[i].CalculateDimension(stringBounder))
	}
	return d
}

// DrawU draws the blocks left to right, aligned vertically.
func (t *TextBlockHorizontal) DrawU(ug UGraphic) {
	x := 0.0
	dimtotal := t.CalculateDimension(ug.StringBounder())
	for _, block := range t.blocks {
		dimb := block.CalculateDimension(ug.StringBounder())
		var dy float64
		switch t.alignment {
		case geom.VerticalAlignmentCenter:
			dy = (dimtotal.Height() - dimb.Height()) / 2
		case geom.VerticalAlignmentBottom:
			dy = dimtotal.Height() - dimb.Height()
		}
		block.DrawU(ug.Apply(NewUTranslate(x, dy)))
		x += dimb.Width()
	}
}

// TextBlockVertical stacks blocks.
type TextBlockVertical struct {
	TextBlockDefaults
	memo                TextBlockMemoized
	blocks              []TextBlock
	horizontalAlignment geom.HorizontalAlignment
}

var _ WithPorts = (*TextBlockVertical)(nil)

// NewTextBlockVertical returns the blocks stacked; it panics with fewer than
// two blocks, as upstream throws.
func NewTextBlockVertical(all []TextBlock, horizontalAlignment geom.HorizontalAlignment) *TextBlockVertical {
	if len(all) < 2 {
		throw(illegalArgument, "")
	}
	return &TextBlockVertical{blocks: append([]TextBlock(nil), all...), horizontalAlignment: horizontalAlignment}
}

// CalculateDimension returns the merged size, memoized.
func (t *TextBlockVertical) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, t.calculateDimensionSlow)
}

func (t *TextBlockVertical) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.blocks[0].CalculateDimension(stringBounder)
	for i := 1; i < len(t.blocks); i++ {
		d = d.MergeTB(t.blocks[i].CalculateDimension(stringBounder))
	}
	return d
}

// DrawU draws the blocks top to bottom, each over its background color.
func (t *TextBlockVertical) DrawU(ug UGraphic) {
	y := 0.0
	dimtotal := t.CalculateDimension(ug.StringBounder())
	for _, block := range t.blocks {
		dimb := block.CalculateDimension(ug.StringBounder())
		back := block.Backcolor()
		if back != nil && !back.IsTransparent() {
			ug.Apply(UTranslateDy(y)).Apply(back).Apply(Bg(back)).
				Draw(URectangleBuild(dimtotal.Width(), dimb.Height()))
		}
		switch t.horizontalAlignment {
		case geom.HorizontalAlignmentLeft:
			block.DrawU(ug.Apply(UTranslateDy(y)))
		case geom.HorizontalAlignmentCenter:
			dx := (dimtotal.Width() - dimb.Width()) / 2
			block.DrawU(ug.Apply(NewUTranslate(dx, y)))
		case geom.HorizontalAlignmentRight:
			dx := dimtotal.Width() - dimb.Width()
			block.DrawU(ug.Apply(NewUTranslate(dx, y)))
		default:
			throw(unsupportedOperation, "")
		}
		y += dimb.Height()
	}
}

// Ports returns the ports of the blocks, each moved down to its block.
func (t *TextBlockVertical) Ports(stringBounder font.StringBounder) *Ports {
	y := 0.0
	result := NewPorts()
	for _, block := range t.blocks {
		dimb := block.CalculateDimension(stringBounder)
		wp, ok := block.(WithPorts)
		if !ok {
			throw("java.lang.ClassCastException", "")
		}
		result.AddThis(wp.Ports(stringBounder).TranslateY(y))
		y += dimb.Height()
	}
	return result
}

// InnerPosition returns the position of member in the first block that has
// it.
func (t *TextBlockVertical) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	y := 0.0
	for _, block := range t.blocks {
		dimb := block.CalculateDimension(stringBounder)
		if result := block.InnerPosition(member, stringBounder); result != nil {
			r := UTranslateDy(y).Apply(*result)
			return &r
		}
		y += dimb.Height()
	}
	return nil
}

// simpleTextBlockBordered draws a rectangle around a block.
type simpleTextBlockBordered struct {
	TextBlockDefaults
	textBlock TextBlock
	color     color.HColor
}

func newSimpleTextBlockBordered(textBlock TextBlock, c color.HColor) *simpleTextBlockBordered {
	return &simpleTextBlockBordered{textBlock: textBlock, color: c}
}

func (t *simpleTextBlockBordered) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.textBlock.CalculateDimension(stringBounder).DeltaWH(1, 1)
}

func (t *simpleTextBlockBordered) DrawU(ug UGraphic) {
	d := t.textBlock.CalculateDimension(ug.StringBounder())
	t.textBlock.DrawU(ug.Apply(NewUTranslate(1, 1)))
	ug.Apply(t.color).Draw(URectangleBuild(d.Width(), d.Height()))
}

// TextBlockGeneric draws a dashed rectangle over a block.
type TextBlockGeneric struct {
	TextBlockDefaults
	textBlock  TextBlock
	background color.HColor
	border     color.HColor
}

// NewTextBlockGeneric returns the block with a dashed border.
func NewTextBlockGeneric(textBlock TextBlock, background, border color.HColor) *TextBlockGeneric {
	return &TextBlockGeneric{textBlock: textBlock, background: background, border: border}
}

// CalculateDimension returns the size of the block.
func (t *TextBlockGeneric) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.textBlock.CalculateDimension(stringBounder)
}

// DrawU draws the border, then the block.
func (t *TextBlockGeneric) DrawU(ug UGraphic) {
	ug = ug.Apply(Bg(t.background))
	ug = ug.Apply(t.border)
	d := t.CalculateDimension(ug.StringBounder())
	ug.Apply(NewUStroke(2, 2, 1)).Draw(URectangleBuild(d.Width(), d.Height()))
	t.textBlock.DrawU(ug)
}

// TextBlockSprited puts a sprite (or circled character) left of a block.
type TextBlockSprited struct {
	TextBlockDefaults
	parent TextBlock
	sprite TextBlock
}

// NewTextBlockSprited returns the sprite and the block.
func NewTextBlockSprited(sprite, parent TextBlock) *TextBlockSprited {
	return &TextBlockSprited{sprite: sprite, parent: parent}
}

// CalculateDimension returns the size of both, with a margin of 6 after the
// sprite.
func (t *TextBlockSprited) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	widthCircledCharacter := t.circledCharacterWithAndMargin(stringBounder)
	heightCircledCharacter := t.sprite.CalculateDimension(stringBounder).Height()
	d := t.parent.CalculateDimension(stringBounder)
	return geom.NewXDimension2D(d.Width()+widthCircledCharacter, math.Max(heightCircledCharacter, d.Height()))
}

func (t *TextBlockSprited) circledCharacterWithAndMargin(stringBounder font.StringBounder) float64 {
	return t.sprite.CalculateDimension(stringBounder).Width() + 6.0
}

// DrawU draws the sprite, then the block.
func (t *TextBlockSprited) DrawU(ug UGraphic) {
	stringBounder := ug.StringBounder()
	t.sprite.DrawU(ug)
	widthCircledCharacter := t.circledCharacterWithAndMargin(stringBounder)
	t.parent.DrawU(ug.Apply(UTranslateDx(widthCircledCharacter)))
}

// TextBlockWithURL draws a block inside a link.
type TextBlockWithURL struct {
	block TextBlock
	url   *URL
}

// TextBlockWithURLOf returns block inside the link url; block itself when
// url is nil.
func TextBlockWithURLOf(block TextBlock, url *URL) TextBlock {
	if url == nil {
		return block
	}
	return &TextBlockWithURL{block: block, url: url}
}

// DrawU draws the block between StartURL and CloseURL.
func (t *TextBlockWithURL) DrawU(ug UGraphic) {
	ug.StartURL(t.url)
	t.block.DrawU(ug)
	ug.CloseURL()
}

// CalculateDimension returns the size of the block.
func (t *TextBlockWithURL) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.block.CalculateDimension(stringBounder)
}

// MinMax returns the bounds of the block.
func (t *TextBlockWithURL) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return t.block.MinMax(stringBounder)
}

// InnerPosition returns the position of member in the block.
func (t *TextBlockWithURL) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return t.block.InnerPosition(member, stringBounder)
}

// MagneticBorder returns the border of the block.
func (t *TextBlockWithURL) MagneticBorder() MagneticBorder { return t.block.MagneticBorder() }

// Backcolor returns the background color of the block.
func (t *TextBlockWithURL) Backcolor() color.HColor { return t.block.Backcolor() }
