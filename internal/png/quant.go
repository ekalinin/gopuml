// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/png/quant/Quantify555.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/png/quant/Cube555.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/png/quant/QuantifyPacked28.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/png/quant/QuantUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package png

// The color reductions upstream applies before writing a PNG, which change
// the pixels written.

// The RGB555 cubes, plus one slot for the transparent color.
const (
	cubeCountRGB555  = 32 * 32 * 32
	transparentCube  = cubeCountRGB555
	totalCubeSlots   = cubeCountRGB555 + 1
	minCubeThreshold = 32
)

// cube555 counts the sub-colors (the 3 low bits of each channel) of the
// pixels of one RGB555 cube.
type cube555 struct {
	counts [512]int32
	rgb555 int
}

// best returns the most frequent sub-color, the first one on ties.
func (c *cube555) best() int {
	best, besti := int32(0), 0
	for i, n := range c.counts {
		if n > best {
			best = n
			besti = i
		}
	}
	return besti
}

func isTransparent(argb uint32) bool { return (argb>>24)&0xFF < 127 }

// cubeAndSub returns the cube and sub-color of a pixel.
func cubeAndSub(argb uint32) (int, int) {
	if isTransparent(argb) {
		return transparentCube, 0
	}
	r8 := int(argb>>16) & 0xFF
	g8 := int(argb>>8) & 0xFF
	b8 := int(argb) & 0xFF
	cube := (r8>>3)<<10 | (g8>>3)<<5 | b8>>3
	sub := (r8&7)<<6 | (g8&7)<<3 | b8&7
	return cube, sub
}

// representativeARGB rebuilds an opaque color from a cube and a sub-color.
func representativeARGB(cube, sub int) uint32 {
	if cube == transparentCube {
		return 0
	}
	r8 := (cube>>10&0x1F)<<3 | sub>>6&7
	g8 := (cube>>5&0x1F)<<3 | sub>>3&7
	b8 := (cube&0x1F)<<3 | sub&7
	return 0xFF<<24 | uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
}

// quantify555 is upstream's Quantify555.packMeIfPossible: when the pixels
// fall in 32 to 256 RGB555 cubes (pixels with alpha below 127 count as one
// transparent cube), it returns a palette with the most frequent color of
// each cube, in cube order, and the palette index of each pixel; ok is
// false otherwise.
func quantify555(pix []uint32) (palette []uint32, indices []uint8, ok bool) {
	cubes := make([]*cube555, totalCubeSlots)
	nbCubes := 0
	for _, argb := range pix {
		ci, sub := cubeAndSub(argb)
		c := cubes[ci]
		if c == nil {
			if nbCubes > 255 {
				return nil, nil, false
			}
			nbCubes++
			c = &cube555{rgb555: ci}
			cubes[ci] = c
		}
		c.counts[sub]++
	}
	if nbCubes < minCubeThreshold {
		return nil, nil, false
	}
	cubeToPal := make([]int, totalCubeSlots)
	for _, c := range cubes {
		if c == nil {
			continue
		}
		cubeToPal[c.rgb555] = len(palette)
		palette = append(palette, representativeARGB(c.rgb555, c.best()))
	}
	indices = make([]uint8, len(pix))
	for i, argb := range pix {
		ci, _ := cubeAndSub(argb)
		indices[i] = uint8(cubeToPal[ci])
	}
	return palette, indices, true
}

// compressPackedARGB replaces the lowest bit of each channel by its highest
// bit, which upstream does to compress better (QuantUtils).
func compressPackedARGB(argb uint32) uint32 {
	return argb&0xFEFEFEFE | (argb>>7)&0x01010101
}
