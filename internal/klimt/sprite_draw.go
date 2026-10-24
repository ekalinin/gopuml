// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteMonochrome.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteColor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteImage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/StdlibSprite.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/StdlibSpriteSvg.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"image"
	"image/draw"
	"math"
	"sync"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/stdlib"
)

// SpriteMonochrome draws upstream's SpriteMonochrome: Gray holds the level
// of each pixel, row by row, from 0 to GrayLevel-1.
type SpriteMonochrome struct {
	Width, Height, GrayLevel int
	Gray                     []int32
}

// ToUImage is SpriteMonochrome.toUImage: each level is a color between the
// back color and the color, and the lowest levels fade out.
func (s *SpriteMonochrome) ToUImage(mapper *color.ColorMapper, backcolor, c color.HColor) *UImage {
	if backcolor == nil || backcolor.IsTransparent() {
		backcolor = color.White.WithDark(color.Black)
	}
	if c == nil || c.IsTransparent() {
		c = color.Black.WithDark(color.White)
	}
	im := NewPortableImageARGB(s.Width, s.Height)
	gradient := color.Gradient(backcolor, c, 0)
	maxCoef := 0.0
	for _, g := range s.Gray {
		maxCoef = math.Max(maxCoef, float64(g)/float64(s.GrayLevel-1))
	}
	for line := 0; line < s.Height; line++ {
		for col := 0; col < s.Width; col++ {
			coef := float64(s.Gray[line*s.Width+col]) / float64(s.GrayLevel-1)
			alpha := 255
			if !(coef > maxCoef/4) {
				// 0/0 is NaN, which Java's (int) makes 0.
				alpha = javaD2I(255 * (coef * 4 / maxCoef))
			}
			im.SetRGB(col, line, gradient.Color(mapper, coef, alpha).RGB())
		}
	}
	return NewUImage(NewPixelImage(im, AffineTransformBilinear))
}

// AsTextBlock draws the sprite in the forced color, or else the font color,
// over the back color of the graphic.
func (s *SpriteMonochrome) AsTextBlock(fontColor, forcedColor color.HColor, scale float64, _ color.HColor) TextBlock {
	c := fontColor
	if forcedColor != nil {
		c = forcedColor
	}
	return &spriteTextBlock{width: float64(s.Width) * scale, height: float64(s.Height) * scale,
		draw: func(ug UGraphic) {
			ug.Draw(s.ToUImage(ug.ColorMapper(), ug.Param().Backcolor(), c).Scale(scale))
		}}
}

// SpriteColor draws upstream's SpriteColor: Color holds the RGB color of
// each pixel, row by row, -1 for a pixel drawn from Gray (16 levels) between
// the back color and the font color.
type SpriteColor struct {
	Width, Height int
	Gray, Color   []int32
}

// ToUImage is SpriteColor.toUImage. Upstream's image is TYPE_INT_RGB, so
// every pixel is opaque.
func (s *SpriteColor) ToUImage(mapper *color.ColorMapper, backcolor, forecolor color.HColor) *UImage {
	im := NewPortableImageARGB(s.Width, s.Height)
	if backcolor == nil {
		backcolor = color.White
	}
	if forecolor == nil {
		forecolor = color.Black
	}
	gradient := color.Gradient(backcolor, forecolor, 0)
	for line := 0; line < s.Height; line++ {
		for col := 0; col < s.Width; col++ {
			i := line*s.Width + col
			rgb := s.Color[i]
			if rgb == -1 {
				coef := float64(s.Gray[i]) / (16 - 1)
				rgb = gradient.RGB(mapper, coef, 255)
			}
			im.SetRGB(col, line, rgb|-0x1000000)
		}
	}
	return NewUImage(NewPixelImage(im, AffineTransformBilinear))
}

// AsTextBlock draws the sprite with the font color for its gray pixels.
func (s *SpriteColor) AsTextBlock(fontColor, _ color.HColor, scale float64, _ color.HColor) TextBlock {
	return &spriteTextBlock{width: float64(s.Width) * scale, height: float64(s.Height) * scale,
		draw: func(ug UGraphic) {
			ug.Draw(s.ToUImage(ug.ColorMapper(), ug.Param().Backcolor(), fontColor).Scale(scale))
		}}
}

// SpriteImage draws upstream's SpriteImage: a raster image, recolored with
// the forced color, or else the font color, when there is one.
type SpriteImage struct {
	img *UImage
}

// NewSpriteImage returns the sprite of an image.
func NewSpriteImage(img PortableImage) *SpriteImage {
	return &SpriteImage{img: NewUImage(NewPixelImage(img, AffineTransformBilinear))}
}

// AsTextBlock draws the image, in gray for the monochrome color mapper.
func (s *SpriteImage) AsTextBlock(fontColor, forcedColor color.HColor, scale float64, _ color.HColor) TextBlock {
	used := fontColor
	if forcedColor != nil {
		used = forcedColor
	}
	return &spriteTextBlock{width: float64(s.img.Width()) * scale, height: float64(s.img.Height()) * scale,
		draw: func(ug UGraphic) {
			mapper := ug.ColorMapper()
			switch {
			case mapper == color.MonochromeMapper:
				ug.Draw(s.img.Monochrome().Scale(scale))
			case used == nil:
				ug.Draw(s.img.Scale(scale))
			default:
				ug.Draw(s.img.MuteColor(used.ToColor(mapper)).Scale(scale))
			}
		}}
}

// SpriteSvg draws upstream's SpriteSvg, the sprite of an SVG file: the SVG
// itself, which the PNG output does not draw.
type SpriteSvg struct {
	svg string
}

// NewSpriteSvg returns the sprite of an SVG text.
func NewSpriteSvg(svg string) *SpriteSvg { return &SpriteSvg{svg: svg} }

// AsTextBlock draws the SVG image at the scale.
func (s *SpriteSvg) AsTextBlock(_, _ color.HColor, scale float64, _ color.HColor) TextBlock {
	img := NewUImageSvg(s.svg, scale)
	return &spriteTextBlock{width: img.Width() * scale, height: img.Height() * scale,
		draw: func(ug UGraphic) { ug.Draw(img) }}
}

// StdlibSprite draws upstream's StdlibSprite, a monochrome sprite of the
// standard library with 16 levels, two pixels per byte (the high nibble
// first), decoded on first use.
type StdlibSprite struct {
	width, height int
	data          []byte
	once          sync.Once
	sprite        *SpriteMonochrome
}

// NewStdlibSprite returns the sprite of packed standard library data.
func NewStdlibSprite(width, height int, data []byte) *StdlibSprite {
	return &StdlibSprite{width: width, height: height, data: data}
}

// AsTextBlock decodes the sprite once and draws it as a SpriteMonochrome.
func (s *StdlibSprite) AsTextBlock(fontColor, forcedColor color.HColor, scale float64, backColor color.HColor) TextBlock {
	s.once.Do(func() {
		sprite := &SpriteMonochrome{Width: s.width, Height: s.height, GrayLevel: 16,
			Gray: make([]int32, s.width*s.height)}
		set := func(x, y int, level int32) {
			if y < s.height {
				sprite.Gray[y*s.width+x] = level
			}
		}
		pos := 0
		for j := 0; j < (s.height+1)/2; j++ {
			for i := 0; i < s.width; i++ {
				b := s.data[pos]
				pos++
				set(i, j*2, int32(b>>4))
				set(i, j*2+1, int32(b&0x0F))
			}
		}
		s.sprite, s.data = sprite, nil // free memory, as upstream does
	})
	return s.sprite.AsTextBlock(fontColor, forcedColor, scale, backColor)
}

// StdlibSpriteSvg draws upstream's StdlibSpriteSvg, an SVG sprite of the
// standard library, read and parsed with the default parser on first use.
// A library without SVG sprites throws UncheckedIOException, and a name
// the library does not have draws nothing.
type StdlibSpriteSvg struct {
	lib    *stdlib.Library
	name   string
	once   sync.Once
	sprite Sprite
	err    error
}

// NewStdlibSpriteSvg returns the SVG sprite of a library name.
func NewStdlibSpriteSvg(lib *stdlib.Library, name string) *StdlibSpriteSvg {
	return &StdlibSpriteSvg{lib: lib, name: name}
}

// AsTextBlock draws the parsed sprite.
func (s *StdlibSpriteSvg) AsTextBlock(fontColor, forcedColor color.HColor, scale float64, backColor color.HColor) TextBlock {
	s.once.Do(func() {
		svg, ok, err := s.lib.SvgSprite(s.name)
		switch {
		case err != nil:
			s.err = err
		case ok:
			s.sprite = SvgSpriteParserFactoryCreateDefault(svg)
		}
	})
	if s.err != nil {
		throw("java.io.UncheckedIOException", stdlib.IOExceptionString(s.err))
	}
	if s.sprite == nil {
		return TextBlockUtilsEmptyTextBlock()
	}
	return s.sprite.AsTextBlock(fontColor, forcedColor, scale, backColor)
}

// spriteTextBlock is the text block of a sprite: a fixed size and a
// drawing function.
type spriteTextBlock struct {
	TextBlockDefaults
	width, height float64
	draw          func(ug UGraphic)
}

func (b *spriteTextBlock) DrawU(ug UGraphic) { b.draw(ug) }

func (b *spriteTextBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// PortableImageFromImage converts a decoded image to a PortableImage with
// non-premultiplied ARGB pixels, as SImageIO.read gives.
func PortableImageFromImage(img image.Image) PortableImage {
	b := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, b.Min, draw.Src)
	result := NewPortableImageARGB(b.Dx(), b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			o := nrgba.PixOffset(x, y)
			p := nrgba.Pix[o : o+4]
			result.SetRGB(x, y, int32(uint32(p[3])<<24|uint32(p[0])<<16|uint32(p[1])<<8|uint32(p[2])))
		}
	}
	return result
}
