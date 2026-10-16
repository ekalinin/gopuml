// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/GraphicStrings.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/BigFrame.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// GraphicStrings is a block of text lines on a background, possibly with an
// image: the text of error and information images.
type GraphicStrings struct {
	TextBlockDefaults
	memo              TextBlockMemoized
	margin            float64
	background        color.HColor
	strings           []string
	image             PortableImage
	imagePadding      float64
	position          geom.GraphicPosition
	fontConfiguration *FontConfiguration
	mode              CreoleMode
	isCrash           bool
}

var _ IEntityImage = (*GraphicStrings)(nil)

func graphicStringsTextColor() color.HColor { return color.Black.WithDark(color.White) }

// GraphicStringsCreateForError returns the black on light red (or green on
// black) bold text of an error image, without creole.
func GraphicStringsCreateForError(strs []string, useRed bool) IEntityImage {
	return newGraphicStrings(strs, GraphicStringsSansSerif14(graphicStringsForeColor(useRed)).Bold(),
		graphicStringsBackColor(useRed), nil, 0, CreoleModeNoCreole, true)
}

func graphicStringsForeColor(useRed bool) color.HColor {
	if useRed {
		return color.Black
	}
	return color.MyGreen
}

func graphicStringsBackColor(useRed bool) color.HColor {
	if useRed {
		return color.RedLight
	}
	return color.Black
}

// GraphicStringsCreateGreenOnBlackMonospaced returns green monospaced text
// on black.
func GraphicStringsCreateGreenOnBlackMonospaced(strs []string) TextBlock {
	return newGraphicStrings(strs, graphicStringsMonospaced14(color.Green), color.Black, nil, 0,
		CreoleModeSimpleLine, false)
}

// GraphicStringsCreateBlackOnWhite returns black text on white.
func GraphicStringsCreateBlackOnWhite(strs []string) TextBlock {
	return newGraphicStrings(strs, GraphicStringsSansSerif12(graphicStringsTextColor()), color.White, nil, 0,
		CreoleModeFull, false)
}

// GraphicStringsCreateBlackOnWhiteMonospaced returns black monospaced text on
// white.
func GraphicStringsCreateBlackOnWhiteMonospaced(strs []string) TextBlock {
	return newGraphicStrings(strs, graphicStringsMonospaced14(graphicStringsTextColor()), color.White, nil, 0,
		CreoleModeFull, false)
}

// GraphicStringsCreateBlackOnWhiteImage returns black text on white with an
// image at a position.
func GraphicStringsCreateBlackOnWhiteImage(strs []string, image PortableImage,
	position geom.GraphicPosition) TextBlock {
	return newGraphicStrings(strs, GraphicStringsSansSerif12(graphicStringsTextColor()), color.White, image,
		position, CreoleModeFullButUnderscore, false)
}

// GraphicStringsSansSerif12 returns SansSerif 12 in a color, with blue
// underlined hyperlinks.
func GraphicStringsSansSerif12(c color.HColor) *FontConfiguration {
	return NewFontConfiguration(font.SansSerif(12), c, color.Blue, strokePtr(UStrokeSimple()))
}

// GraphicStringsSansSerif14 returns SansSerif 14 in a color, with blue
// underlined hyperlinks.
func GraphicStringsSansSerif14(c color.HColor) *FontConfiguration {
	return NewFontConfiguration(font.SansSerif(14), c, color.Blue, strokePtr(UStrokeSimple()))
}

func graphicStringsMonospaced14(c color.HColor) *FontConfiguration {
	return NewFontConfiguration(font.Monospace(14), c, color.Blue, strokePtr(UStrokeSimple()))
}

func newGraphicStrings(strs []string, fontConfiguration *FontConfiguration, background color.HColor,
	image PortableImage, position geom.GraphicPosition, mode CreoleMode, isCrash bool) *GraphicStrings {
	return &GraphicStrings{margin: 5, imagePadding: 30, strings: strs, background: background, image: image,
		position: position, mode: mode, fontConfiguration: fontConfiguration, isCrash: isCrash}
}

func (g *GraphicStrings) textBlock() TextBlock {
	display := DisplayCreate(g.strings...)
	if g.mode == CreoleModeNoCreole {
		return NewTextBlockRaw(g.strings, g.fontConfiguration)
	}
	return display.Create7(g.fontConfiguration, geom.HorizontalAlignmentLeft, SpriteContainerEmpty{}, g.mode)
}

// DrawU draws the text and the image after the margin.
func (g *GraphicStrings) DrawU(ug UGraphic) {
	ug = ug.Apply(NewUTranslate(g.margin, g.margin))
	size := g.CalculateDimension(ug.StringBounder()).Delta(-2 * g.margin)
	g.textBlock().DrawU(ug.Apply(g.fontConfiguration.Color()))
	if g.image != nil {
		iw, ih := float64(g.image.Width()), float64(g.image.Height())
		img := NewUImage(newPixelImage(g.image, AffineTransformBilinear))
		switch g.position {
		case geom.GraphicPositionBottom:
			ug.Apply(NewUTranslate((size.Width()-iw)/2, size.Height()-ih)).Draw(img)
		case geom.GraphicPositionBackgroundCornerBottomRight:
			ug.Apply(NewUTranslate(size.Width()-iw, size.Height()-ih)).Draw(img)
		case geom.GraphicPositionBackgroundCornerTopRight:
			ug.Apply(NewUTranslate(size.Width()-iw-1, 1)).Draw(img)
		}
	}
}

// CalculateDimension returns the size of the text and image plus the
// margins, memoized.
func (g *GraphicStrings) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return g.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return g.calculateDimensionInternal22(sb).Delta(2 * g.margin)
	})
}

func (g *GraphicStrings) calculateDimensionInternal22(stringBounder font.StringBounder) geom.XDimension2D {
	d := g.textBlock().CalculateDimension(stringBounder)
	if g.image != nil {
		iw, ih := float64(g.image.Width()), float64(g.image.Height())
		switch g.position {
		case geom.GraphicPositionBottom:
			d = geom.NewXDimension2D(d.Width(), d.Height()+ih)
		case geom.GraphicPositionBackgroundCornerBottomRight, geom.GraphicPositionBackgroundCornerTopRight:
			d = geom.NewXDimension2D(d.Width()+g.imagePadding+iw, d.Height())
		}
	}
	return d
}

// ShapeType returns the rectangle.
func (g *GraphicStrings) ShapeType() ShapeType { return ShapeTypeRectangle }

// Backcolor returns the background color.
func (g *GraphicStrings) Backcolor() color.HColor { return g.background }

// Shield returns no margins.
func (g *GraphicStrings) Shield(font.StringBounder) Margins { return MarginsNone() }

// IsHidden returns false.
func (g *GraphicStrings) IsHidden() bool { return false }

// OverscanX returns 0.
func (g *GraphicStrings) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash reports whether the text is a crash report.
func (g *GraphicStrings) IsCrash() bool { return g.isCrash }

// BigFrame is the "mainframe" border around a diagram: a rectangle with a
// title tab.
type BigFrame struct {
	TextBlockDefaults
	memo          TextBlockMemoized
	title         TextBlock
	original      TextBlock
	padding       ClockwiseTopRightBottomLeft
	symbolContext *Fashion
}

// NewBigFrame returns the frame of original with a title.
func NewBigFrame(title, original TextBlock, padding ClockwiseTopRightBottomLeft, symbolContext *Fashion) *BigFrame {
	return &BigFrame{title: title, original: original, padding: padding, symbolContext: symbolContext}
}

func (b *BigFrame) ypos(dimTitle geom.XDimension2D) float64 {
	if dimTitle.Width() == 0 {
		return 12
	}
	return dimTitle.Height() + 3
}

func (b *BigFrame) effectivePadding(stringBounder font.StringBounder) ClockwiseTopRightBottomLeft {
	dimTitle := b.title.CalculateDimension(stringBounder)
	return b.padding.IncTop(dimTitle.Height() + 10)
}

func (b *BigFrame) computeWidth(stringBounder font.StringBounder) float64 {
	dimTitle := b.title.CalculateDimension(stringBounder)
	effectivePadding := b.effectivePadding(stringBounder)
	originalMinMax := TextBlockUtilsMinMax(b.original, stringBounder, false)
	ww := originalMinMax.Width()
	if originalMinMax.MinX() >= 0 {
		ww = originalMinMax.MaxX()
	}
	return effectivePadding.Left() + math.Max(ww+12, dimTitle.Width()+10) + effectivePadding.Right()
}

func (b *BigFrame) computeHeight(stringBounder font.StringBounder) float64 {
	dimTitle := b.title.CalculateDimension(stringBounder)
	effectivePadding := b.effectivePadding(stringBounder)
	originalMinMax := TextBlockUtilsMinMax(b.original, stringBounder, false)
	hh := originalMinMax.Height()
	if originalMinMax.MinY() >= 0 {
		hh = originalMinMax.MaxY()
	}
	return effectivePadding.Top() + dimTitle.Height() + hh + effectivePadding.Bottom()
}

// DrawU draws the frame and its title.
func (b *BigFrame) DrawU(ug UGraphic) {
	stringBounder := ug.StringBounder()
	d := b.CalculateDimension(stringBounder)
	ug = b.symbolContext.Apply(ug)
	dimTitle := b.title.CalculateDimension(stringBounder)
	widthFull := d.Width()
	rectangle := URectangleBuild(widthFull, d.Height()).Rounded(b.symbolContext.RoundCorner()).
		IgnoreForCompressionOnX().IgnoreForCompressionOnY()
	rectangle.SetDeltaShadow(b.symbolContext.DeltaShadow())
	ug.Draw(rectangle)
	var textWidth float64
	var cornersize int
	if dimTitle.Width() == 0 {
		textWidth = widthFull / 3
		cornersize = 7
	} else {
		textWidth = dimTitle.Width() + 10
		cornersize = 10
	}
	textHeight := b.ypos(dimTitle)
	line := UPathNone()
	line.SetIgnoreForCompressionOnX()
	line.MoveTo(textWidth, 0)
	line.LineTo(textWidth, textHeight-float64(cornersize))
	line.LineTo(textWidth-float64(cornersize), textHeight)
	line.LineTo(0, textHeight)
	ug.Apply(Bg(color.None())).Draw(line)
	widthTitle := b.title.CalculateDimension(stringBounder).Width()
	// Temporary hack...
	if widthFull-widthTitle < 25 {
		b.title.DrawU(ug.Apply(NewUTranslate(3, 1)))
	} else {
		ug.Apply(NewUTranslate(3, 1)).Draw(NewSpecialText(b.title))
	}
}

// CalculateDimension returns the size of the frame, memoized.
func (b *BigFrame) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(b.computeWidth(sb), b.computeHeight(sb))
	})
}
