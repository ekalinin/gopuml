// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseGroupingHeader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseGroupingElse.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseGroupingSpace.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The folded corner and comment margin of a grouping header.
const (
	groupingHeaderCornersize    = 10
	groupingHeaderCommentMargin = 0
)

// ComponentRoseGroupingHeader is the header of a grouping (alt, loop,
// group...): the frame, the tab with the keyword and the comment in
// brackets.
type ComponentRoseGroupingHeader struct {
	klimt.AbstractTextualComponent
	commentTextBlock    klimt.TextBlock
	background          color.HColor
	symbolContext       *klimt.Fashion
	symbolContextCorner *klimt.Fashion
	roundCorner         float64
}

// NewComponentRoseGroupingHeader returns the header of a grouping: the
// first line of strings is the keyword, the second (if any and not nil)
// the comment. teoz makes the background transparent.
func NewComponentRoseGroupingHeader(teoz bool, style, styleHeader *klimt.Style, strings *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseGroupingHeader {
	c := &ComponentRoseGroupingHeader{}
	label, ok := charSequenceLabel(strings.Get(0))
	c.InitAbstractTextualComponentLabel(c, styleHeader, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(1, 30, 1, 15), skinParam, label, ok)

	c.roundCorner = float64(style.Value(klimt.PNameRoundCorner).AsInt())
	if teoz {
		c.background = color.Transparent()
	} else {
		c.background = style.Value(klimt.PNameBackGroundColor).AsColor(c.IHtmlColorSet())
	}
	c.symbolContext = style.SymbolContext(c.IHtmlColorSet(), nil)
	c.symbolContextCorner = styleHeader.SymbolContext(c.IHtmlColorSet(), nil)

	smallFont2 := style.FontConfiguration(c.IHtmlColorSet(), nil)

	if strings.Size() == 1 || strings.Get(1) == nil {
		c.commentTextBlock = nil
	} else {
		display := klimt.DisplayWithNewlines(skinParam.Pragma(), "["+strings.Get(1).String()+"]")
		c.commentTextBlock = display.Create(smallFont2, geom.HorizontalAlignmentLeft, skinParam)
	}
	if c.background == nil {
		throw(nullPointer, "")
	}
	return c
}

// charSequenceLabel returns the text of a line, ok false for a nil line
// (null upstream).
func charSequenceLabel(s klimt.CharSequence) (string, bool) {
	if s == nil {
		return "", false
	}
	return s.String(), true
}

func (c *ComponentRoseGroupingHeader) suppHeightForComment(stringBounder font.StringBounder) float64 {
	if c.commentTextBlock == nil {
		return 0
	}
	height := c.commentTextBlock.CalculateDimension(stringBounder).Height()
	if height > 15 {
		return height - 15
	}
	return 0
}

// PreferredWidth returns the width of the keyword and of the comment.
func (c *ComponentRoseGroupingHeader) PreferredWidth(stringBounder font.StringBounder) float64 {
	var sup float64
	if c.commentTextBlock == nil {
		sup = groupingHeaderCommentMargin * 2
	} else {
		size := c.commentTextBlock.CalculateDimension(stringBounder)
		sup = c.OldPaddingX1() + groupingHeaderCommentMargin + size.Width()
	}
	return c.TextWidth(stringBounder) + sup
}

// PreferredHeight returns the height of the keyword and of a multi-line
// comment.
func (c *ComponentRoseGroupingHeader) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 2*c.PaddingY() + c.suppHeightForComment(stringBounder)
}

// DrawBackgroundInternalU draws the frame filled with the background.
func (c *ComponentRoseGroupingHeader) DrawBackgroundInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = c.symbolContext.ApplyStroke(ug).Apply(c.symbolContext.ForeColor())
	rect := klimt.URectangleBuild(dimensionToUse.Width(), dimensionToUse.Height()).Rounded(c.roundCorner)
	rect.SetDeltaShadow(c.symbolContext.DeltaShadow())
	ug.Apply(klimt.Bg(c.background)).Draw(rect)
}

// DrawInternalU draws the tab, the frame, the keyword and the comment.
func (c *ComponentRoseGroupingHeader) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	stringBounder := ug.StringBounder()
	textWidth := c.TextWidth(stringBounder)
	textHeight := c.TextHeight(stringBounder)

	c.symbolContextCorner.Apply(ug).Draw(c.corner(textWidth, textHeight))

	ug = c.symbolContext.ApplyStroke(ug).Apply(c.symbolContext.ForeColor())
	rect := klimt.URectangleBuild(dimensionToUse.Width(), dimensionToUse.Height()).Rounded(c.roundCorner)
	ug.Draw(rect)

	ug = ug.Apply(klimt.UStrokeSimple())

	c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(c.OldPaddingX1(), c.OldPaddingY())))

	if c.commentTextBlock != nil {
		x1 := c.OldPaddingX1() + textWidth
		y2 := c.OldPaddingY() + 1

		c.commentTextBlock.DrawU(ug.Apply(klimt.NewUTranslate(x1+groupingHeaderCommentMargin, y2)))
	}
}

func (c *ComponentRoseGroupingHeader) corner(width, height float64) *klimt.UPath {
	polygon := klimt.UPathNone()
	if c.roundCorner == 0 {
		polygon.MoveTo(0, 0)
		polygon.LineTo(width, 0)

		polygon.LineTo(width, height-groupingHeaderCornersize)
		polygon.LineTo(width-groupingHeaderCornersize, height)

		polygon.LineTo(0, height)
		polygon.LineTo(0, 0)
	} else {
		polygon.MoveTo(c.roundCorner/2, 0)
		polygon.LineTo(width, 0)

		polygon.LineTo(width, height-groupingHeaderCornersize)
		polygon.LineTo(width-groupingHeaderCornersize, height)

		polygon.LineTo(0, height)
		polygon.LineTo(0, c.roundCorner/2)

		polygon.ArcTo(c.roundCorner/2, c.roundCorner/2, 0, 0, 1, c.roundCorner/2, 0)
	}
	return polygon
}

// ComponentRoseGroupingElse is the separator of an "else" inside a
// grouping: a dashed line and the comment in brackets.
type ComponentRoseGroupingElse struct {
	klimt.AbstractTextualComponent
	groupBorder     color.HColor
	backgroundColor color.HColor
	roundCorner     float64
	teoz            bool
}

// NewComponentRoseGroupingElse returns the separator of an else with its
// comment; comment may be nil (no comment). teoz makes the background
// transparent.
func NewComponentRoseGroupingElse(teoz bool, style *klimt.Style, comment klimt.CharSequence,
	skinParam klimt.ISkinParam) *ComponentRoseGroupingElse {
	c := &ComponentRoseGroupingElse{}
	label, ok := "", false
	if comment != nil {
		label, ok = "["+comment.String()+"]", true
	}
	c.InitAbstractTextualComponentLabel(c, style, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(1, 5, 1, 5), skinParam, label, ok)

	c.teoz = teoz
	c.roundCorner = c.RoundCorner()
	c.groupBorder = c.ColorLine()
	if teoz {
		c.backgroundColor = color.Transparent()
	} else {
		c.backgroundColor = c.ColorBackGround()
	}
	return c
}

// DrawBackgroundInternalU fills the area below the separator, with rounded
// bottom corners when the style asks for round corners.
func (c *ComponentRoseGroupingElse) DrawBackgroundInternalU(ug klimt.UGraphic, area *klimt.Area) {
	if c.backgroundColor.IsTransparent() {
		return
	}

	dimensionToUse := area.DimensionToUse()
	ug = ug.Apply(color.None()).Apply(klimt.Bg(c.backgroundColor))
	width := dimensionToUse.Width()
	height := dimensionToUse.Height()
	var rect klimt.UShape
	if c.roundCorner == 0 {
		rect = klimt.URectangleBuild(width, height)
	} else {
		path := klimt.UPathNone()
		path.MoveTo(0, 0)
		path.LineTo(width, 0)

		path.LineTo(width, height-c.roundCorner/2)
		path.ArcTo(c.roundCorner/2, c.roundCorner/2, 0, 0, 1, width-c.roundCorner/2, height)

		path.LineTo(c.roundCorner/2, height)
		path.ArcTo(c.roundCorner/2, c.roundCorner/2, 0, 0, 1, 0, height-c.roundCorner/2)

		path.LineTo(0, 0)
		rect = path
	}
	ug.Draw(rect)
}

// DrawInternalU draws the dashed line and the comment.
func (c *ComponentRoseGroupingElse) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = klimt.ArrowConfigurationStroke(ug, 2, 2, 1).Apply(c.groupBorder)
	ug.Apply(klimt.UTranslateDy(1)).Draw(klimt.ULineH(dimensionToUse.Width()))
	ug = ug.Apply(klimt.UStrokeSimple())
	if c.teoz {
		c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(c.OldPaddingX1(), c.OldPaddingY()+2)))
	} else {
		c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(c.OldPaddingX1(), c.OldPaddingY())))
	}
}

// PreferredHeight returns the height of the comment, 4 more with teoz.
func (c *ComponentRoseGroupingElse) PreferredHeight(stringBounder font.StringBounder) float64 {
	if c.teoz {
		return c.TextHeight(stringBounder) + 4
	}
	return c.TextHeight(stringBounder)
}

// PreferredWidth returns the width of the comment.
func (c *ComponentRoseGroupingElse) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder)
}

// ComponentRoseGroupingSpace is an empty vertical space of a grouping.
type ComponentRoseGroupingSpace struct {
	space float64
}

var _ klimt.Component = (*ComponentRoseGroupingSpace)(nil)

// NewComponentRoseGroupingSpace returns a space of a height.
func NewComponentRoseGroupingSpace(space float64) *ComponentRoseGroupingSpace {
	return &ComponentRoseGroupingSpace{space: space}
}

// UsedStyles panics, as upstream throws UnsupportedOperationException.
func (c *ComponentRoseGroupingSpace) UsedStyles() []*klimt.Style {
	throw(unsupportedOperation, "")
	return nil
}

// StyleSignature panics, as upstream throws UnsupportedOperationException.
func (c *ComponentRoseGroupingSpace) StyleSignature() klimt.StyleSignature {
	throw(unsupportedOperation, "")
	return nil
}

// PreferredWidth returns 0.
func (c *ComponentRoseGroupingSpace) PreferredWidth(font.StringBounder) float64 { return 0 }

// PreferredHeight returns the space.
func (c *ComponentRoseGroupingSpace) PreferredHeight(font.StringBounder) float64 { return c.space }

// DrawU draws nothing.
func (c *ComponentRoseGroupingSpace) DrawU(klimt.UGraphic, *klimt.Area, klimt.Context2D) {}

// PreferredDimension returns the preferred width and height.
func (c *ComponentRoseGroupingSpace) PreferredDimension(stringBounder font.StringBounder) geom.XDimension2D {
	w := c.PreferredWidth(stringBounder)
	h := c.PreferredHeight(stringBounder)
	return geom.NewXDimension2D(w, h)
}
