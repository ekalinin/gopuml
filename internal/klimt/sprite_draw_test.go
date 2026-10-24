// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// drawnSprite is a sprite of spriteDrawings of SpriteGolden.java.
type drawnSprite struct {
	name   string
	sprite Sprite
}

// drawnSprites builds the sprites of the harness, in its order.
func drawnSprites(t *testing.T) []drawnSprite {
	mono := func(width, height, levels int, gray func(i, j int) int32) *SpriteMonochrome {
		s := &SpriteMonochrome{Width: width, Height: height, GrayLevel: levels, Gray: make([]int32, width*height)}
		for j := 0; j < height; j++ {
			for i := 0; i < width; i++ {
				s.Gray[j*width+i] = gray(i, j)
			}
		}
		return s
	}
	colorSprite := func(width, height int, set func(s *SpriteColor)) *SpriteColor {
		s := &SpriteColor{Width: width, Height: height, Gray: make([]int32, width*height),
			Color: make([]int32, width*height)}
		set(s)
		return s
	}
	setGray := func(s *SpriteColor, x, y int, level int32) {
		s.Gray[y*s.Width+x], s.Color[y*s.Width+x] = level, -1
	}
	setColor := func(s *SpriteColor, x, y int, rgb int32) {
		s.Gray[y*s.Width+x], s.Color[y*s.Width+x] = -1, rgb
	}
	return []drawnSprite{
		{"mono2", mono(3, 1, 2, func(i, _ int) int32 { return map[bool]int32{true: 1}[i == 1] })},
		{"mono4", mono(4, 4, 4, func(i, j int) int32 { return int32((i + j) % 4) })},
		{"mono8", mono(4, 2, 8, func(i, j int) int32 {
			if j == 0 {
				return int32(i * 2)
			}
			return int32(7 - i)
		})},
		{"mono16", mono(4, 4, 16, func(i, j int) int32 { return int32(i*4 + j) })},
		{"mono-zero", mono(2, 2, 16, func(int, int) int32 { return 0 })},
		{"mono-uniform", mono(2, 2, 16, func(int, int) int32 { return 15 })},
		{"color", colorSprite(3, 2, func(s *SpriteColor) {
			setColor(s, 0, 0, 0xFF0000)
			setColor(s, 1, 0, 0x00FF00)
			setGray(s, 2, 0, 8)
			setGray(s, 0, 1, 15)
			setColor(s, 1, 1, 0x123456)
		})},
		{"color-uniform", colorSprite(2, 2, func(s *SpriteColor) {
			for i := range s.Color {
				s.Gray[i], s.Color[i] = -1, 0x336699
			}
		})},
		{"image", NewSpriteImage(spriteDecodePng(t, spriteFourPng))},
		{"image-uniform", NewSpriteImage(spriteDecodePng(t, spriteUniformPng))},
		{"svgfile", NewSpriteSvg(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="6">` +
			`<rect width="10" height="6" fill="#00F"/></svg>`)},
		{"stdlib", NewStdlibSprite(3, 3, []byte{0x0F, 0xF0, 0x88, 0x1E, 0x2D, 0x3C})},
	}
}

// TestGoldenSpriteDraw compares the drawings of the sprite classes with
// spritedraw.golden, written by SpriteGolden.java.
func TestGoldenSpriteDraw(t *testing.T) {
	type config struct {
		name               string
		font, forced, back color.HColor
		mapper             *color.ColorMapper
		scale              float64
	}
	configs := []config{
		{"plain", nil, nil, nil, color.IdentityMapper, 1},
		{"font", spriteColor("#FF0000"), nil, nil, color.IdentityMapper, 1},
		{"forced", spriteColor("#FF0000"), spriteColor("#0000FF"), nil, color.IdentityMapper, 1},
		{"back", spriteColor("#00FF00"), nil, spriteColor("#FFEEDD"), color.IdentityMapper, 1},
		{"mono", spriteColor("#FF8000"), nil, nil, color.MonochromeMapper, 1},
		{"s2", spriteColor("#FF0000"), nil, nil, color.IdentityMapper, 2},
	}
	var out []string
	for _, s := range drawnSprites(t) {
		uniform := strings.HasSuffix(s.name, "-uniform") || s.name == "svgfile"
		for _, cfg := range configs {
			spriteDrawing(&out, s.name+"\t"+cfg.name, false, cfg.mapper, func(rec *spriteRec) {
				var ug UGraphic = rec
				if cfg.back != nil {
					ug = ug.Apply(Bg(cfg.back))
				}
				tb := s.sprite.AsTextBlock(cfg.font, cfg.forced, cfg.scale, nil)
				d := tb.CalculateDimension(fakeBounder{})
				*rec.log = append(*rec.log, "dim "+d2s(d.Width())+","+d2s(d.Height()))
				// Only uniform sprites are drawn scaled, as in the harness.
				if cfg.scale == 1 || uniform {
					tb.DrawU(ug)
				}
			})
		}
	}
	compareGolden(t, "spritedraw.golden", out)
}
