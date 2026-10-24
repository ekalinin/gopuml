// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/SingleLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TileText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockRaw.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TileImage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TileImageSvg.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// SingleLine is a line of text made of tiles.
type SingleLine struct {
	TextBlockDefaults
	memo                TextBlockMemoized
	blocs               []TextBlock
	horizontalAlignment geom.HorizontalAlignment
}

var _ Line = (*SingleLine)(nil)

// SingleLineRawText returns a line of plain text; an empty text is a space.
func SingleLineRawText(text string, fontConfiguration *FontConfiguration) *SingleLine {
	result := &SingleLine{horizontalAlignment: geom.HorizontalAlignmentLeft}
	if text == "" {
		text = " "
	}
	result.blocs = append(result.blocs, NewTileText(text, fontConfiguration, nil))
	return result
}

// CalculateDimension returns the sum of the tile widths and the highest
// tile, memoized.
func (s *SingleLine) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return s.memo.Dimension(stringBounder, s.calculateDimensionSlow)
}

func (s *SingleLine) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	width := 0.0
	height := 0.0
	for _, b := range s.blocs {
		size2D := b.CalculateDimension(stringBounder)
		width += size2D.Width()
		height = math.Max(height, size2D.Height())
	}
	return geom.NewXDimension2D(width, height)
}

func (s *SingleLine) maxDeltaY(ug UGraphic) float64 {
	result := 0.0
	d := s.CalculateDimension(ug.StringBounder())
	for _, b := range s.blocs {
		tile, ok := b.(*TileText)
		if !ok {
			continue
		}
		dimBloc := b.CalculateDimension(ug.StringBounder())
		deltaY := d.Height() - dimBloc.Height() + tile.FontSize2D()
		result = math.Max(result, deltaY)
	}
	return result
}

// DrawU draws the tiles left to right, texts on a common baseline and other
// tiles at the bottom.
func (s *SingleLine) DrawU(ug UGraphic) {
	deltaY := s.maxDeltaY(ug)
	stringBounder := ug.StringBounder()
	d := s.CalculateDimension(stringBounder)
	x := 0.0
	for _, b := range s.blocs {
		if _, ok := b.(*TileText); ok {
			b.DrawU(ug.Apply(NewUTranslate(x, deltaY)))
		} else {
			dy := d.Height() - b.CalculateDimension(stringBounder).Height()
			b.DrawU(ug.Apply(NewUTranslate(x, dy)))
		}
		x += b.CalculateDimension(stringBounder).Width()
	}
}

// HorizontalAlignment returns the alignment of the line.
func (s *SingleLine) HorizontalAlignment() geom.HorizontalAlignment { return s.horizontalAlignment }

// TileText is a piece of text in one font, possibly a link. Tabulations
// move to the next multiple of eight spaces.
type TileText struct {
	TextBlockDefaults
	memo              TextBlockMemoized
	text              string
	fontConfiguration *FontConfiguration
	url               *URL
}

// NewTileText returns a piece of text; url may be nil.
func NewTileText(text string, fontConfiguration *FontConfiguration, url *URL) *TileText {
	return &TileText{text: text, fontConfiguration: fontConfiguration, url: url}
}

// CalculateDimension returns the size of the text, at least 10 high plus
// the position shift, memoized.
func (t *TileText) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, t.calculateDimensionSlow)
}

func (t *TileText) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	rect := dim(stringBounder.CalculateDimension(t.fontConfiguration.Font(), t.text))
	spaceBottom := t.fontConfiguration.Space()
	if spaceBottom < 0 {
		spaceBottom = -spaceBottom
	}
	h := rect.Height()
	if h < 10 {
		h = 10
	}
	width := rect.Width()
	if strings.ContainsRune(t.text, '\t') || strings.ContainsRune(t.text, JawsBlockE1RealTabulation) {
		width = t.width(stringBounder)
	}
	return geom.NewXDimension2D(width, h+float64(spaceBottom))
}

// FontSize2D returns the size of the font.
func (t *TileText) FontSize2D() float64 { return t.fontConfiguration.Font().Size2D() }

func (t *TileText) tabSize(stringBounder font.StringBounder) float64 {
	return stringBounder.CalculateDimension(t.fontConfiguration.Font(), "        ").Width
}

const tileTextDelims = "\t" + string(JawsBlockE1RealTabulation)

func isTileTab(s string) bool { return s == "\t" || s == string(JawsBlockE1RealTabulation) }

// DrawU draws the text pieces between the tabulations.
func (t *TileText) DrawU(ug UGraphic) {
	x := 0.0
	if t.url != nil {
		ug.StartURL(t.url)
	}
	ug = ug.Apply(t.fontConfiguration.Color())
	tokens := javaTokenizeWithDelims(t.text, tileTextDelims)
	if len(tokens) > 0 {
		tabSize := t.tabSize(ug.StringBounder())
		for _, s := range tokens {
			if isTileTab(s) {
				remainder := math.Mod(x, tabSize)
				x += tabSize - remainder
			} else {
				utext := UTextBuild(s, t.fontConfiguration)
				d := ug.StringBounder().CalculateDimension(t.fontConfiguration.Font(), s)
				space := t.fontConfiguration.Space()
				ypos := float64(space)
				ug.Apply(NewUTranslate(x, ypos)).Draw(utext)
				x += d.Width
			}
		}
	}
	if t.url != nil {
		ug.CloseURL()
	}
}

func (t *TileText) width(stringBounder font.StringBounder) float64 {
	tabSize := t.tabSize(stringBounder)
	x := 0.0
	for _, s := range javaTokenizeWithDelims(t.text, tileTextDelims) {
		if isTileTab(s) {
			remainder := math.Mod(x, tabSize)
			x += tabSize - remainder
		} else {
			x += stringBounder.CalculateDimension(t.fontConfiguration.Font(), s).Width
		}
	}
	return x
}

// TextBlockRaw is lines of plain text, without creole.
type TextBlockRaw struct {
	TextBlockDefaults
	memo              TextBlockMemoized
	lines2            []Line
	strings           []string
	fontConfiguration *FontConfiguration
}

// NewTextBlockRaw returns the lines in a font configuration.
func NewTextBlockRaw(strs []string, fontConfiguration *FontConfiguration) *TextBlockRaw {
	return &TextBlockRaw{strings: strs, fontConfiguration: fontConfiguration}
}

func (t *TextBlockRaw) lines(stringBounder font.StringBounder) []Line {
	if t.lines2 == nil {
		if stringBounder == nil {
			throw(illegalState, "")
		}
		t.lines2 = []Line{}
		for _, s := range t.strings {
			t.lines2 = append(t.lines2, SingleLineRawText(s, t.fontConfiguration))
		}
	}
	return t.lines2
}

// CalculateDimension returns the widest line and the sum of the heights,
// memoized.
func (t *TextBlockRaw) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, t.TextDimension)
}

// TextDimension computes the dimension without the cache.
func (t *TextBlockRaw) TextDimension(stringBounder font.StringBounder) geom.XDimension2D {
	width := 0.0
	height := 0.0
	for _, line := range t.lines(stringBounder) {
		size2D := line.CalculateDimension(stringBounder)
		height += size2D.Height()
		width = math.Max(width, size2D.Width())
	}
	return geom.NewXDimension2D(width, height)
}

// DrawU draws the lines top to bottom.
func (t *TextBlockRaw) DrawU(ug UGraphic) {
	y := 0.0
	for _, line := range t.lines(ug.StringBounder()) {
		line.DrawU(ug.Apply(UTranslateDy(y)))
		y += line.CalculateDimension(ug.StringBounder()).Height()
	}
}

// TileImage is a raster image in a text, with vertical space around it.
type TileImage struct {
	TextBlockDefaults
	image  PortableImage
	vspace int
}

// NewTileImage returns the tile of an image; valign is ignored, as
// upstream does.
func NewTileImage(image PortableImage, valign geom.ImgValign, vspace int) *TileImage {
	if image == nil {
		throw(nullPointer, "")
	}
	return &TileImage{image: image, vspace: vspace}
}

// CalculateDimension returns the image size plus the vertical space.
func (t *TileImage) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(float64(t.image.Width()), float64(t.image.Height()+2*t.vspace))
}

// DrawU draws the image. Upstream wraps it in a net.atmp.PixelImage, a
// MutableImage that the raster step ports.
func (t *TileImage) DrawU(ug UGraphic) {
	ug.Apply(UTranslateDy(float64(t.vspace))).Draw(NewUImage(newPixelImage(t.image, AffineTransformBilinear)))
}

// TileImageSvg is an SVG image in a text.
type TileImageSvg struct {
	TextBlockDefaults
	svg   *UImageSvg
	scale float64
}

// NewTileImageSvg returns the tile of an SVG text at a scale.
func NewTileImageSvg(svg string, scale float64) *TileImageSvg {
	return &TileImageSvg{svg: NewUImageSvg(svg, scale), scale: scale}
}

// CalculateDimension returns the scaled size of the SVG.
func (t *TileImageSvg) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.svg.Width(), t.svg.Height())
}

// DrawU draws the SVG.
func (t *TileImageSvg) DrawU(ug UGraphic) { ug.Draw(t.svg) }
