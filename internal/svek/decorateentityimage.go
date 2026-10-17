// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/DecorateEntityImage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/EntityImageLegend.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// decorateEntityImage is upstream's DecorateEntityImage: a block with a
// text above and below it.
type decorateEntityImage struct {
	klimt.TextBlockDefaults
	original    klimt.TextBlock
	group1      *klimt.UGroup
	horizontal1 geom.HorizontalAlignment
	text1       klimt.TextBlock
	group2      *klimt.UGroup
	horizontal2 geom.HorizontalAlignment
	text2       klimt.TextBlock
}

// decorateEntityImageAdd is DecorateEntityImage.add: text above the block
// for the TOP alignment, below it otherwise.
func decorateEntityImageAdd(group *klimt.UGroup, original, text klimt.TextBlock, horizontal geom.HorizontalAlignment,
	verticalAlignment geom.VerticalAlignment) klimt.TextBlock {
	if original == nil {
		throwNPE()
	}
	if verticalAlignment == geom.VerticalAlignmentTop {
		return &decorateEntityImage{original: original, group1: group, text1: text, horizontal1: horizontal}
	}
	return &decorateEntityImage{original: original, group2: group, text2: text, horizontal2: horizontal}
}

func textDim(text klimt.TextBlock, stringBounder font.StringBounder) geom.XDimension2D {
	if text == nil {
		return geom.NewXDimension2D(0, 0)
	}
	return text.CalculateDimension(stringBounder)
}

func textX(dimText, dimTotal geom.XDimension2D, h geom.HorizontalAlignment) float64 {
	switch h {
	case geom.HorizontalAlignmentCenter:
		return (dimTotal.Width() - dimText.Width()) / 2
	case geom.HorizontalAlignmentLeft:
		return 0
	case geom.HorizontalAlignmentRight:
		return dimTotal.Width() - dimText.Width()
	}
	throwIllegalState("")
	return 0
}

// DrawU draws the texts and the block.
func (d *decorateEntityImage) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimOriginal := d.original.CalculateDimension(stringBounder)
	dimText1 := textDim(d.text1, stringBounder)
	dimText2 := textDim(d.text2, stringBounder)
	dimTotal := d.CalculateDimension(stringBounder)
	yImage := dimText1.Height()
	yText2 := yImage + dimOriginal.Height()
	xImage := (dimTotal.Width() - dimOriginal.Width()) / 2
	if d.text1 != nil {
		xText1 := textX(dimText1, dimTotal, d.horizontal1)
		if d.group1 != nil {
			ug.StartGroup(d.group1)
		}
		d.text1.DrawU(ug.Apply(klimt.UTranslateDx(xText1)))
		if d.group1 != nil {
			ug.CloseGroup()
		}
	}
	d.original.DrawU(ug.Apply(klimt.NewUTranslate(xImage, yImage)))
	if d.text2 != nil {
		xText2 := textX(dimText2, dimTotal, d.horizontal2)
		if d.group2 != nil {
			ug.StartGroup(d.group2)
		}
		d.text2.DrawU(ug.Apply(klimt.NewUTranslate(xText2, yText2)))
		if d.group2 != nil {
			ug.CloseGroup()
		}
	}
}

// Backcolor returns the background of the block.
func (d *decorateEntityImage) Backcolor() color.HColor { return d.original.Backcolor() }

// CalculateDimension returns the size of the block with its texts.
func (d *decorateEntityImage) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimOriginal := d.original.CalculateDimension(stringBounder)
	dimText := textDim(d.text1, stringBounder).MergeTB(textDim(d.text2, stringBounder))
	return dimOriginal.MergeTB(dimText)
}

// MinMax returns the bounds of the size.
func (d *decorateEntityImage) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return geom.MinMaxFromDim(d.CalculateDimension(stringBounder))
}

// entityImageLegend is upstream's EntityImageLegend.create: a legend in the
// legend style of the diagram type.
func entityImageLegend(note *klimt.Display, skinParam klimt.ISkinParam) klimt.TextBlock {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameRoot, klimt.SNameDocument,
		uml.DiagramTypeStyleName(skinParam.DiagramType()), klimt.SNameLegend).
		MergedStyle(skinParam.CurrentStyleBuilder())
	return style.CreateTextBlockBordered(note, skinParam.IHtmlColorSet(), skinParam, "_legend", style.WrapWidth())
}
