// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/TextBlockMap.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"math"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// mapRow is an entry of TextBlockMap's blocksMap: the key and value blocks
// of a row.
type mapRow struct {
	key, value klimt.TextBlock
}

// TextBlockMap is the body of a map: a table of two columns, the keys and
// their values; a linked entry ("key *--> target") has a dot instead of a
// value.
type TextBlockMap struct {
	klimt.TextBlockDefaults
	memo klimt.TextBlockMemoized

	skinParam           klimt.ISkinParam
	fontConfiguration   *klimt.FontConfiguration
	blocksMap           []mapRow
	keys                []string
	totalWidth          float64
	horizontalAlignment geom.HorizontalAlignment
}

var _ klimt.WithPorts = (*TextBlockMap)(nil)

// NewTextBlockMap returns the table of the entries, keys in order and the
// values by key.
func NewTextBlockMap(fontConfiguration *klimt.FontConfiguration, skinParam klimt.ISkinParam, keys []string,
	values map[string]string, wordWrap klimt.LineBreakStrategy,
	horizontalAlignment geom.HorizontalAlignment) *TextBlockMap {
	t := &TextBlockMap{skinParam: skinParam, horizontalAlignment: horizontalAlignment,
		fontConfiguration: fontConfiguration}
	for _, key := range keys {
		value := values[key]
		if klimt.VisibilityIsVisibilityCharacter(key) {
			key = string(utf16.Decode(utf16.Encode([]rune(key))[1:]))
		}
		t.keys = append(t.keys, key)
		block1 := t.textBlock(key, wordWrap)
		block2 := t.textBlock(value, wordWrap)
		t.blocksMap = append(t.blocksMap, mapRow{key: block1, value: block2})
	}
	return t
}

// Ports implements klimt.WithPorts: a port per key.
func (t *TextBlockMap) Ports(stringBounder font.StringBounder) *klimt.Ports {
	ports := klimt.NewPorts()
	y := 0.0
	for i, row := range t.blocksMap {
		height := heightOfRow(stringBounder, row.key, row.value)
		ports.Add(t.keys[i], 100, y, height)
		y += height
	}
	return ports
}

// CalculateDimension returns the size of the table, memoized.
func (t *TextBlockMap) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(t.widthColA(sb)+t.widthColB(sb), t.totalHeight(sb))
	})
}

func (t *TextBlockMap) widthColA(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, row := range t.blocksMap {
		width = math.Max(width, row.key.CalculateDimension(stringBounder).Width())
	}
	return width
}

func (t *TextBlockMap) widthColB(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, row := range t.blocksMap {
		width = math.Max(width, row.value.CalculateDimension(stringBounder).Width())
	}
	return width
}

// DrawU draws the rows, each after a horizontal line as wide as the map
// box.
func (t *TextBlockMap) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	fullDim := t.CalculateDimension(stringBounder)
	trueWidth := math.Max(fullDim.Width(), t.totalWidth)
	widthColA := t.widthColA(stringBounder)
	y := 0.0
	for _, row := range t.blocksMap {
		ugline := ug.Apply(klimt.UTranslateDy(y))
		ugline.Draw(klimt.ULineH(trueWidth))
		height := heightOfRow(stringBounder, row.key, row.value)
		keyWidth := row.key.CalculateDimension(stringBounder).Width()
		if _, ok := row.value.(*mapPoint); ok {
			posColA := t.horizontalAlignment.Position(keyWidth, trueWidth)
			row.key.DrawU(ugline.Apply(klimt.UTranslateDx(posColA)))
		} else {
			posColA := t.horizontalAlignment.Position(keyWidth, widthColA)
			row.key.DrawU(ugline.Apply(klimt.UTranslateDx(posColA)))
			row.value.DrawU(ugline.Apply(klimt.UTranslateDx(widthColA)))
			ugline.Apply(klimt.UTranslateDx(widthColA)).Draw(klimt.ULineV(height))
		}
		y += height
	}
}

func (t *TextBlockMap) totalHeight(stringBounder font.StringBounder) float64 {
	height := 0.0
	for _, row := range t.blocksMap {
		height += heightOfRow(stringBounder, row.key, row.value)
	}
	return height
}

func heightOfRow(stringBounder font.StringBounder, key, value klimt.TextBlock) float64 {
	return math.Max(key.CalculateDimension(stringBounder).Height(), value.CalculateDimension(stringBounder).Height())
}

func (t *TextBlockMap) textBlock(key string, wordWrap klimt.LineBreakStrategy) klimt.TextBlock {
	if key == "\x00" {
		return &mapPoint{color: t.fontConfiguration.Color()}
	}
	display := klimt.DisplayWithNewlines(t.skinParam.Pragma(), key)
	result := display.Create0(t.fontConfiguration, geom.HorizontalAlignmentLeft, t.skinParam, wordWrap,
		klimt.CreoleModeFull, nil, nil)
	return klimt.TextBlockUtilsWithMargin(result, 5, 2)
}

// SetTotalWidth sets the width of the map box, the length of the row lines.
func (t *TextBlockMap) SetTotalWidth(totalWidth float64) { t.totalWidth = totalWidth }

// mapPoint is upstream's TextBlockMap.Point: the dot of a linked entry.
type mapPoint struct {
	klimt.TextBlockDefaults
	color color.HColor
}

func (p *mapPoint) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(p.diameter(), p.diameter())
}

func (p *mapPoint) DrawU(ug klimt.UGraphic) {
	point := klimt.UEllipseBuild(p.diameter(), p.diameter())
	ug = ug.Apply(p.color).Apply(klimt.Bg(p.color))
	ug.Draw(point)
}

func (p *mapPoint) diameter() float64 { return 7 }
