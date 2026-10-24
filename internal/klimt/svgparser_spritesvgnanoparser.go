// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svg/parser/SpriteSvgNanoParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// SpriteSvgNanoParser is a sprite of a raster image, drawn in gray levels
// on a monochrome graphic and recolored with a forced color. Upstream does
// not use it.
type SpriteSvgNanoParser struct {
	img *UImage
}

var _ Sprite = (*SpriteSvgNanoParser)(nil)

// NewSpriteSvgNanoParser returns the sprite of an image. A nil image panics,
// as upstream's Objects.requireNonNull throws NullPointerException.
func NewSpriteSvgNanoParser(img PortableImage) *SpriteSvgNanoParser {
	if img == nil {
		throw(nullPointer, "")
	}
	return &SpriteSvgNanoParser{img: NewUImage(NewPixelImage(img, AffineTransformBilinear))}
}

// AsTextBlock returns the image at scale; backColor is ignored, as upstream.
func (s *SpriteSvgNanoParser) AsTextBlock(fontColor, forcedColor color.HColor, scale float64,
	backColor color.HColor) TextBlock {
	return &spriteSvgNanoParserBlock{img: s.img, forcedColor: forcedColor, scale: scale}
}

// spriteSvgNanoParserBlock is the memoized text block of upstream's
// asTextBlock.
type spriteSvgNanoParserBlock struct {
	TextBlockDefaults
	memo        TextBlockMemoized
	img         *UImage
	forcedColor color.HColor
	scale       float64
}

// DrawU draws the image.
func (b *spriteSvgNanoParserBlock) DrawU(ug UGraphic) {
	colorMapper := ug.ColorMapper()
	switch {
	case colorMapper == color.MonochromeMapper:
		ug.Draw(b.img.Monochrome().Scale(b.scale))
	case b.forcedColor == nil:
		ug.Draw(b.img.Scale(b.scale))
	default:
		ug.Draw(b.img.MuteColor(b.forcedColor.ToColor(colorMapper)).Scale(b.scale))
	}
}

// CalculateDimension returns the size of the image (each minus one) times
// the scale.
func (b *spriteSvgNanoParserBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(float64(b.img.Width())*b.scale, float64(b.img.Height())*b.scale)
	})
}
