// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"image"
	"image/color"
)

// Image is the drawing surface: a BufferedImage of TYPE_INT_RGB (opaque) or
// TYPE_INT_ARGB (non-premultiplied), as upstream's EmptyImageBuilder
// creates it.
type Image struct {
	width, height int
	// Pix holds 0xAARRGGBB values row by row; the alpha of an opaque image
	// is always 0xFF.
	pix      []uint32
	hasAlpha bool
}

// NewImage returns a transparent black ARGB image, or a black opaque one
// when hasAlpha is false.
func NewImage(width, height int, hasAlpha bool) *Image {
	im := &Image{width: width, height: height, pix: make([]uint32, width*height), hasAlpha: hasAlpha}
	if !hasAlpha {
		for i := range im.pix {
			im.pix[i] = 0xFF000000
		}
	}
	return im
}

// Width returns the width in pixels.
func (im *Image) Width() int { return im.width }

// Height returns the height in pixels.
func (im *Image) Height() int { return im.height }

// HasAlpha reports whether the image is TYPE_INT_ARGB.
func (im *Image) HasAlpha() bool { return im.hasAlpha }

// ARGB returns the pixel at (x, y) as 0xAARRGGBB.
func (im *Image) ARGB(x, y int) uint32 { return im.pix[y*im.width+x] }

// Pix returns the pixels, 0xAARRGGBB row by row. The slice is the image's
// storage.
func (im *Image) Pix() []uint32 { return im.pix }

// ToNRGBA returns a copy of the image as an image.NRGBA.
func (im *Image) ToNRGBA() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, im.width, im.height))
	for i, v := range im.pix {
		out.Pix[4*i] = uint8(v >> 16)
		out.Pix[4*i+1] = uint8(v >> 8)
		out.Pix[4*i+2] = uint8(v)
		out.Pix[4*i+3] = uint8(v >> 24)
	}
	return out
}

// ColorModel, Bounds and At make the image an image.Image.
func (im *Image) ColorModel() color.Model { return color.NRGBAModel }

// Bounds implements image.Image.
func (im *Image) Bounds() image.Rectangle { return image.Rect(0, 0, im.width, im.height) }

// At implements image.Image.
func (im *Image) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= im.width || y >= im.height {
		return color.NRGBA{}
	}
	v := im.pix[y*im.width+x]
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: uint8(v >> 24)}
}

// mul8 and div8 are the 8-bit multiplication and division tables of
// Java2D's software loops (AlphaMath.c): mul8[a][b] is a * b / 255 and
// div8[a][b] is b * 255 / a, both rounded, the division clamped to 255.
var mul8, div8 [256][256]uint8

func init() {
	for i := 1; i < 256; i++ {
		inc := uint32(i<<16 + i<<8 + i)
		val := inc + 1<<23
		for j := 1; j < 256; j++ {
			mul8[i][j] = uint8(val >> 24)
			val += inc
		}
	}
	for i := 1; i < 256; i++ {
		inc := (uint32(0xFF)<<24 + uint32(i/2)) / uint32(i)
		val := uint32(1 << 23)
		for j := range i {
			div8[i][j] = uint8(val >> 24)
			val += inc
		}
		for j := i; j < 256; j++ {
			div8[i][j] = 255
		}
	}
}

// blend composites a non-premultiplied ARGB color into pixel i with the
// SrcOver rule and a coverage of pathA, with the arithmetic of Java2D's
// mask fill loops for INT_RGB and INT_ARGB surfaces.
func (im *Image) blend(i int, src uint32, pathA uint8) {
	srcA := src >> 24
	if srcA == 0 || pathA == 0 {
		return
	}
	srcR, srcG, srcB := (src>>16)&0xFF, (src>>8)&0xFF, src&0xFF
	if srcA != 0xFF {
		srcR, srcG, srcB = uint32(mul8[srcA][srcR]), uint32(mul8[srcA][srcG]), uint32(mul8[srcA][srcB])
	}
	resA, resR, resG, resB := srcA, srcR, srcG, srcB
	if pathA != 0xFF {
		resA = uint32(mul8[pathA][srcA])
		resR, resG, resB = uint32(mul8[pathA][srcR]), uint32(mul8[pathA][srcG]), uint32(mul8[pathA][srcB])
	}
	if resA != 0xFF {
		dstF := 0xFF - resA
		dst := im.pix[i]
		dstA := uint32(0xFF)
		if im.hasAlpha {
			dstA = dst >> 24
		}
		dstA = uint32(mul8[dstF][dstA])
		resA += dstA
		if dstA != 0 {
			dR, dG, dB := (dst>>16)&0xFF, (dst>>8)&0xFF, dst&0xFF
			if dstA != 0xFF {
				dR, dG, dB = uint32(mul8[dstA][dR]), uint32(mul8[dstA][dG]), uint32(mul8[dstA][dB])
			}
			resR += dR
			resG += dG
			resB += dB
		}
		if im.hasAlpha && resA != 0 && resA < 0xFF {
			resR, resG, resB = uint32(div8[resA][resR]), uint32(div8[resA][resG]), uint32(div8[resA][resB])
		}
	}
	if !im.hasAlpha {
		resA = 0xFF
	}
	im.pix[i] = resA<<24 | min(resR, 0xFF)<<16 | min(resG, 0xFF)<<8 | min(resB, 0xFF)
}
