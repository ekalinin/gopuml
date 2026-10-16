// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/HeaderLayout.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// headerLayout places the parts of a class header: the spot, the
// stereotype above the name, and the generic box at the top right.
type headerLayout struct {
	name, stereo, generic, circledCharacter klimt.TextBlock
}

func newHeaderLayout(circledCharacter, stereo, name, generic klimt.TextBlock) headerLayout {
	return headerLayout{circledCharacter: protectAgaintNull(circledCharacter), stereo: protectAgaintNull(stereo),
		name: protectAgaintNull(name), generic: protectAgaintNull(generic)}
}

func protectAgaintNull(block klimt.TextBlock) klimt.TextBlock {
	if block == nil {
		return klimt.NewTextBlockEmpty(0, 0)
	}
	return block
}

func (h headerLayout) dimension(stringBounder font.StringBounder) geom.XDimension2D {
	nameDim := h.name.CalculateDimension(stringBounder)
	genericDim := h.generic.CalculateDimension(stringBounder)
	stereoDim := h.stereo.CalculateDimension(stringBounder)
	circleDim := h.circledCharacter.CalculateDimension(stringBounder)

	width := circleDim.Width() + math.Max(stereoDim.Width(), nameDim.Width()) + genericDim.Width()
	height := math.Max(math.Max(circleDim.Height(), stereoDim.Height()+nameDim.Height()+10), genericDim.Height())
	return geom.NewXDimension2D(width, height)
}

func (h headerLayout) drawU(ug klimt.UGraphic, width, height float64) {
	stringBounder := ug.StringBounder()
	nameDim := h.name.CalculateDimension(stringBounder)
	genericDim := h.generic.CalculateDimension(stringBounder)
	stereoDim := h.stereo.CalculateDimension(stringBounder)
	circleDim := h.circledCharacter.CalculateDimension(stringBounder)

	widthStereoAndName := math.Max(stereoDim.Width(), nameDim.Width())
	suppWith := math.Max(0, width-circleDim.Width()-widthStereoAndName-genericDim.Width())

	h2 := math.Min(circleDim.Width()/4, float64(suppWith*0.1))
	h1 := (suppWith - h2) / 2

	xCircle := h1
	yCircle := (height - circleDim.Height()) / 2
	h.circledCharacter.DrawU(ug.Apply(klimt.NewUTranslate(xCircle, yCircle)))

	diffHeight := height - stereoDim.Height() - nameDim.Height()
	xStereo := circleDim.Width() + (widthStereoAndName-stereoDim.Width())/2 + h1 + h2
	yStereo := diffHeight / 2
	h.stereo.DrawU(ug.Apply(klimt.NewUTranslate(xStereo, yStereo)))

	xName := circleDim.Width() + (widthStereoAndName-nameDim.Width())/2 + h1 + h2
	yName := diffHeight/2 + stereoDim.Height()
	h.name.DrawU(ug.Apply(klimt.NewUTranslate(xName, yName)))

	if genericDim.Width() > 0 {
		const delta = 4
		xGeneric := width - genericDim.Width() + delta
		yGeneric := -delta
		h.generic.DrawU(ug.Apply(klimt.NewUTranslate(xGeneric, float64(yGeneric))))
	}
}
