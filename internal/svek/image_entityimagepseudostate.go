// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImagePseudoState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageDeepHistory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageTransitionLabel.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// pseudoStateSize is the size of upstream's EntityImagePseudoState.
const pseudoStateSize = 22

// EntityImagePseudoState is upstream's EntityImagePseudoState: the circle
// of a history state with "H" (EntityImageDeepHistory: "H*").
type EntityImagePseudoState struct {
	abstractEntityImage
	desc  klimt.TextBlock
	style *klimt.Style
}

var _ klimt.IEntityImage = (*EntityImagePseudoState)(nil)

func newEntityImagePseudoState(entity *abel.Entity, historyText string) *EntityImagePseudoState {
	e := &EntityImagePseudoState{abstractEntityImage: newAbstractEntityImage(entity)}
	e.style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, e.styleName(), klimt.SNameDiamond).
		WithTOBECHANGED(entity.Stereotype()).MergedStyle(e.skinParam.CurrentStyleBuilder())
	fontConfiguration := e.style.FontConfiguration(e.skinParam.IHtmlColorSet(), nil)
	e.desc = klimt.DisplayCreate(historyText).Create(fontConfiguration, geom.HorizontalAlignmentCenter, e.skinParam)
	return e
}

// CalculateDimension returns 22 x 22.
func (e *EntityImagePseudoState) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(pseudoStateSize, pseudoStateSize)
}

// DrawU draws the circle with its text.
func (e *EntityImagePseudoState) DrawU(ug klimt.UGraphic) {
	circle := klimt.UEllipseBuild(pseudoStateSize, pseudoStateSize)
	colorSet := e.skinParam.IHtmlColorSet()
	borderColor := e.style.Value(klimt.PNameLineColor).AsColor(colorSet)
	backgroundColor := e.style.Value(klimt.PNameBackGroundColor).AsColor(colorSet)
	circle.SetDeltaShadow(e.style.Shadowing())
	ug = ug.Apply(e.style.Stroke())
	ug = ug.Apply(klimt.Bg(backgroundColor)).Apply(borderColor)
	ug.Draw(circle)
	dimDesc := e.desc.CalculateDimension(ug.StringBounder())
	x := (pseudoStateSize - dimDesc.Width()) / 2
	y := (pseudoStateSize - dimDesc.Height()) / 2
	e.desc.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
}

// ShapeType returns CIRCLE.
func (e *EntityImagePseudoState) ShapeType() klimt.ShapeType { return klimt.ShapeTypeCircle }

// transitionLabelMargin is the margin of upstream's
// EntityImageTransitionLabel.
const transitionLabelMargin = 4

// EntityImageTransitionLabel is upstream's EntityImageTransitionLabel: the
// label of a transition drawn as a node (the "node" label style).
type EntityImageTransitionLabel struct {
	abstractEntityImage
	textBlock klimt.TextBlock
}

var _ klimt.IEntityImage = (*EntityImageTransitionLabel)(nil)

func newEntityImageTransitionLabel(entity *abel.Entity) *EntityImageTransitionLabel {
	e := &EntityImageTransitionLabel{abstractEntityImage: newAbstractEntityImage(entity)}
	fontConfiguration := klimt.FontConfigurationFromSkinParam(e.skinParam, klimt.FontParamArrow, entity.Stereotype())
	display := entity.Display()
	if klimt.DisplayIsNull(display) {
		display = klimt.DisplayCreate(entity.Name())
	}
	e.textBlock = display.Create(fontConfiguration, geom.HorizontalAlignmentCenter, e.skinParam)
	return e
}

// CalculateDimension returns the size of the text with the margins.
func (e *EntityImageTransitionLabel) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.textBlock.CalculateDimension(stringBounder).DeltaWH(transitionLabelMargin*2, transitionLabelMargin*2)
}

// DrawU draws the label in a rounded box.
func (e *EntityImageTransitionLabel) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimText := e.textBlock.CalculateDimension(stringBounder)
	ug.Draw(klimt.URectangleBuild(dimTotal.Width(), dimTotal.Height()).Rounded(5))
	xText := (dimTotal.Width() - dimText.Width()) / 2
	yText := (dimTotal.Height() - dimText.Height()) / 2
	e.textBlock.DrawU(ug.Apply(klimt.NewUTranslate(xText, yText)))
}

// ShapeType returns RECTANGLE.
func (e *EntityImageTransitionLabel) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }
