// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/atmp/PixelImage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/awt/PortableImageFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// pixelImage is upstream's net.atmp.PixelImage: an image kept at scale 1,
// scaled on demand.
type pixelImage struct {
	imageScale1 PortableImage
	scale       float64
	typ         AffineTransformType
	cache       PortableImage
}

var _ MutableImage = (*pixelImage)(nil)

// NewPixelImage returns the mutable image of pixels; type is the
// interpolation used to scale them.
func NewPixelImage(img PortableImage, typ AffineTransformType) MutableImage {
	return newPixelImage(img, typ)
}

func newPixelImage(img PortableImage, typ AffineTransformType) *pixelImage {
	return &pixelImage{imageScale1: img, scale: 1, typ: typ}
}

func (p *pixelImage) WithScale(scale float64) MutableImage {
	return &pixelImage{imageScale1: p.imageScale1, scale: p.scale * scale, typ: p.typ}
}

func (p *pixelImage) Image() PortableImage {
	if p.scale == 1 {
		return p.imageScale1
	}
	if p.cache == nil {
		p.cache = p.imageScale1.Scale(p.scale, p.typ)
	}
	return p.cache
}

const (
	pixelMaskA   = int32(-0x1000000) // 0xFF000000
	pixelMaskRGB = int32(0x00FFFFFF)
)

func (p *pixelImage) Monochrome() MutableImage {
	cp := p.deepCopy()
	for i := 0; i < p.imageScale1.Width(); i++ {
		for j := 0; j < p.imageScale1.Height(); j++ {
			c := p.imageScale1.RGB(i, j)
			rgb := c & pixelMaskRGB
			grayScale := int32(color.GrayScaleFromRGB(rgb))
			// Upstream writes "grayScale + grayScale << 8 + grayScale << 16",
			// which Java parses as ((g + g) << (8 + g)) << 16, with shift
			// distances taken modulo 32.
			gray := ((grayScale + grayScale) << uint((8+grayScale)&31)) << 16
			a := c & pixelMaskA
			cp.SetRGB(i, j, gray+a)
		}
	}
	return &pixelImage{imageScale1: cp, scale: p.scale, typ: p.typ}
}

func (p *pixelImage) MuteColor(newColor color.XColor) MutableImage {
	darkerRgb := p.darkerRgb()
	cp := p.deepCopy()
	for i := 0; i < p.imageScale1.Width(); i++ {
		for j := 0; j < p.imageScale1.Height(); j++ {
			c := p.imageScale1.RGB(i, j)
			rgb := c & pixelMaskRGB
			a := c & pixelMaskA
			if a != 0 && rgb == darkerRgb {
				cp.SetRGB(i, j, newColor.RGB()+a)
			}
		}
	}
	return &pixelImage{imageScale1: cp, scale: p.scale, typ: p.typ}
}

func (p *pixelImage) MuteTransparentColor(newColor color.XColor) MutableImage {
	cp := p.deepCopy()
	for i := 0; i < p.imageScale1.Width(); i++ {
		for j := 0; j < p.imageScale1.Height(); j++ {
			c := p.imageScale1.RGB(i, j)
			if c&pixelMaskA == 0 {
				cp.SetRGB(i, j, newColor.RGB())
			}
		}
	}
	return &pixelImage{imageScale1: cp, scale: p.scale, typ: p.typ}
}

func (p *pixelImage) darkerRgb() int32 {
	darkerRgb := int32(-1)
	for i := 0; i < p.imageScale1.Width(); i++ {
		for j := 0; j < p.imageScale1.Height(); j++ {
			c := p.imageScale1.RGB(i, j)
			rgb := c & pixelMaskRGB
			if c&pixelMaskA != pixelMaskA {
				continue
			}
			gray := color.GrayScaleFromRGB(rgb)
			if darkerRgb == -1 || gray < color.GrayScaleFromRGB(darkerRgb) {
				darkerRgb = rgb
			}
		}
	}
	return darkerRgb
}

func (p *pixelImage) deepCopy() PortableImage {
	result := NewPortableImageARGB(p.imageScale1.Width(), p.imageScale1.Height())
	for i := 0; i < p.imageScale1.Width(); i++ {
		for j := 0; j < p.imageScale1.Height(); j++ {
			result.SetRGB(i, j, p.imageScale1.RGB(i, j))
		}
	}
	return result
}

func (p *pixelImage) Scale() float64 { return p.scale }

// PortableImageARGB is an in-memory ARGB image, the Go stand-in for the
// BufferedImage that upstream's PortableImageFactory builds.
type PortableImageARGB struct {
	width, height int
	pix           []int32
}

var _ PortableImage = (*PortableImageARGB)(nil)

// NewPortableImageARGB returns a transparent image of a size.
func NewPortableImageARGB(width, height int) *PortableImageARGB {
	return &PortableImageARGB{width: width, height: height, pix: make([]int32, width*height)}
}

// Width returns the width.
func (p *PortableImageARGB) Width() int { return p.width }

// Height returns the height.
func (p *PortableImageARGB) Height() int { return p.height }

// RGB returns the ARGB value of a pixel.
func (p *PortableImageARGB) RGB(x, y int) int32 {
	p.check(x, y)
	return p.pix[y*p.width+x]
}

// SetRGB sets the ARGB value of a pixel.
func (p *PortableImageARGB) SetRGB(x, y int, rgb int32) {
	p.check(x, y)
	p.pix[y*p.width+x] = rgb
}

func (p *PortableImageARGB) check(x, y int) {
	if x < 0 || y < 0 || x >= p.width || y >= p.height {
		throw("java.lang.ArrayIndexOutOfBoundsException", "Coordinate out of bounds!")
	}
}

// ToPngDataURL returns the image as a PNG data URL.
func (p *PortableImageARGB) ToPngDataURL() string {
	img := image.NewNRGBA(image.Rect(0, 0, p.width, p.height))
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			v := uint32(p.pix[y*p.width+x])
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = byte(v>>16), byte(v>>8), byte(v), byte(v>>24)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		throw(illegalState, err.Error())
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// Transparency returns TRANSLUCENT (3), as for TYPE_INT_ARGB images.
func (p *PortableImageARGB) Transparency() int { return 3 }

// Subimage returns a copy of a part of the image (Java shares the pixels).
func (p *PortableImageARGB) Subimage(x, y, width, height int) PortableImage {
	result := NewPortableImageARGB(width, height)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			result.SetRGB(i, j, p.RGB(x+i, y+j))
		}
	}
	return result
}

// Scale returns the image scaled with an interpolation (see scaleARGB).
func (p *PortableImageARGB) Scale(scaleFactor float64, interpolationType AffineTransformType) PortableImage {
	return scaleARGB(p, scaleFactor, interpolationType)
}
