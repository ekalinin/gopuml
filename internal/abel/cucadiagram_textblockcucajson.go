// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/TextBlockCucaJSon.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"math"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// TextBlockCucaJSon is the body of a JSON element: objects as tables of
// names and values, arrays as stacked values.
type TextBlockCucaJSon struct {
	klimt.TextBlockDefaults

	skinParam         klimt.ISkinParam
	fontConfiguration *klimt.FontConfiguration
	json              json.Value
	jsonTextBlock     klimt.TextBlock
	mainTotalWidth    float64
	wordWrap          klimt.LineBreakStrategy
}

var _ klimt.WithPorts = (*TextBlockCucaJSon)(nil)

// NewTextBlockCucaJSon returns the body of the JSON data.
func NewTextBlockCucaJSon(fontConfiguration *klimt.FontConfiguration, skinParam klimt.ISkinParam, value json.Value,
	wordWrap klimt.LineBreakStrategy) *TextBlockCucaJSon {
	return &TextBlockCucaJSon{skinParam: skinParam, json: value, fontConfiguration: fontConfiguration,
		wordWrap: wordWrap}
}

func (t *TextBlockCucaJSon) getJSONTextBlock() klimt.TextBlock {
	if t.jsonTextBlock == nil {
		t.jsonTextBlock = t.textBlockValue(t.json, 0)
	}
	switch b := t.jsonTextBlock.(type) {
	case *textBlockJSONObject:
		b.jsonTotalWidth = t.mainTotalWidth
	case *textBlockJSONArray:
		b.arrayTotalWidth = t.mainTotalWidth
	}
	return t.jsonTextBlock
}

func (t *TextBlockCucaJSon) textBlockValue(value json.Value, width2 float64) klimt.TextBlock {
	switch {
	case value.IsString():
		return t.textBlock(value.AsString())
	case value.IsNull() || value.IsTrue() || value.IsFalse() || value.IsNumber():
		return t.textBlock(value.String())
	case value.IsArray():
		return &textBlockJSONArray{owner: t, array: value.AsArray(), arrayTotalWidth: width2}
	case value.IsObject():
		return &textBlockJSONObject{owner: t, obj: value.AsObject(), jsonTotalWidth: width2}
	}
	return t.textBlock("JsonValue")
}

// Ports implements klimt.WithPorts: none.
func (t *TextBlockCucaJSon) Ports(font.StringBounder) *klimt.Ports { return klimt.NewPorts() }

// CalculateDimension returns the size of the data.
func (t *TextBlockCucaJSon) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.getJSONTextBlock().CalculateDimension(stringBounder)
}

// DrawU draws the data.
func (t *TextBlockCucaJSon) DrawU(ug klimt.UGraphic) { t.getJSONTextBlock().DrawU(ug) }

func (t *TextBlockCucaJSon) textBlock(key string) klimt.TextBlock {
	display := klimt.DisplayWithNewlines(t.skinParam.Pragma(), key)
	result := display.Create0(t.fontConfiguration, geom.HorizontalAlignmentLeft, t.skinParam, t.wordWrap,
		klimt.CreoleModeFull, nil, nil)
	return klimt.TextBlockUtilsWithMargin(result, 5, 2)
}

// SetTotalWidth sets the width of the element box, the length of the row
// lines.
func (t *TextBlockCucaJSon) SetTotalWidth(totalWidth float64) { t.mainTotalWidth = totalWidth }

// textBlockJSONObject is upstream's TextBlockJson: the members of an
// object, names on the left, values on the right.
type textBlockJSONObject struct {
	klimt.TextBlockDefaults
	memo           klimt.TextBlockMemoized
	owner          *TextBlockCucaJSon
	obj            *json.Object
	jsonTotalWidth float64
}

func (b *textBlockJSONObject) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(b.width1(sb)+b.width2(sb), b.height(sb))
	})
}

func (b *textBlockJSONObject) width1(stringBounder font.StringBounder) float64 {
	result := 0.0
	for name := range b.obj.All() {
		tb1 := b.owner.textBlock(name)
		result = math.Max(result, tb1.CalculateDimension(stringBounder).Width())
	}
	return result
}

func (b *textBlockJSONObject) width2(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, value := range b.obj.All() {
		tb2 := b.owner.textBlockValue(value, b.jsonTotalWidth)
		result = math.Max(result, tb2.CalculateDimension(stringBounder).Width())
	}
	return result
}

func (b *textBlockJSONObject) height(stringBounder font.StringBounder) float64 {
	result := 0.0
	for name, value := range b.obj.All() {
		tb1 := b.owner.textBlock(name)
		tb2 := b.owner.textBlockValue(value, b.jsonTotalWidth)
		dim1 := tb1.CalculateDimension(stringBounder)
		dim2 := tb2.CalculateDimension(stringBounder)
		result += math.Max(dim1.Height(), dim2.Height())
	}
	return result
}

func (b *textBlockJSONObject) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	width1 := b.width1(stringBounder)
	height := b.height(stringBounder)
	ug.Apply(klimt.UTranslateDx(width1)).Draw(klimt.ULineV(height))
	hline := klimt.ULineH(b.jsonTotalWidth)
	for name, value := range b.obj.All() {
		tb1 := b.owner.textBlock(name)
		tb2 := b.owner.textBlockValue(value, b.jsonTotalWidth-width1)
		dim1 := tb1.CalculateDimension(stringBounder)
		dim2 := tb2.CalculateDimension(stringBounder)
		ug.Draw(hline)
		tb1.DrawU(ug)
		tb2.DrawU(ug.Apply(klimt.UTranslateDx(width1)))
		ug = ug.Apply(klimt.UTranslateDy(math.Max(dim1.Height(), dim2.Height())))
	}
}

// textBlockJSONArray is upstream's TextBlockArray: the values of an array,
// stacked, separated by lines.
type textBlockJSONArray struct {
	klimt.TextBlockDefaults
	memo            klimt.TextBlockMemoized
	owner           *TextBlockCucaJSon
	array           *json.Array
	arrayTotalWidth float64
}

func (b *textBlockJSONArray) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		result := geom.NewXDimension2D(0, 0)
		for _, element := range b.array.All() {
			tb := b.owner.textBlockValue(element, b.arrayTotalWidth)
			result = result.MergeTB(tb.CalculateDimension(sb))
		}
		return result
	})
}

func (b *textBlockJSONArray) DrawU(ug klimt.UGraphic) {
	hline := klimt.ULineH(b.arrayTotalWidth)
	nb := 0
	for _, element := range b.array.All() {
		tb := b.owner.textBlockValue(element, b.arrayTotalWidth)
		if nb > 0 {
			ug.Draw(hline)
		}
		nb++
		tb.DrawU(ug)
		ug = ug.Apply(klimt.UTranslateDy(tb.CalculateDimension(ug.StringBounder()).Height()))
	}
}
