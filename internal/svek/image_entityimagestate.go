// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageStateCommon.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageState2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageStateEmptyDescription.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/HorizontalAlignment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// styleStateName is EntityImageStateCommon.getStyleStateName.
func styleStateName(stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *klimt.Style {
	return stateStyle().AddSName(klimt.SNameName).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
}

// styleStateDescription is EntityImageStateCommon.getStyleStateDescription.
func styleStateDescription(stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *klimt.Style {
	return stateStyle().AddSName(klimt.SNameDescription).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
}

// styleStateOf is EntityImageStateCommon.getStyleState.
func styleStateOf(stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *klimt.Style {
	return stateStyle().WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
}

// styleStateBody is EntityImageStateCommon.getStyleStateBody.
func styleStateBody(stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *klimt.Style {
	return stateStyle().AddSName(klimt.SNameBody).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
}

// drawAligned is upstream's HorizontalAlignment.draw: tb drawn at
// ypadding, aligned in width with xpadding.
func drawAligned(alignment geom.HorizontalAlignment, ug klimt.UGraphic, tb klimt.TextBlock,
	xpadding, ypadding, width float64) {
	switch alignment {
	case geom.HorizontalAlignmentLeft:
		tb.DrawU(ug.Apply(klimt.NewUTranslate(xpadding, ypadding)))
	case geom.HorizontalAlignmentRight:
		dimTb := tb.CalculateDimension(ug.StringBounder())
		tb.DrawU(ug.Apply(klimt.NewUTranslate(width-dimTb.Width()-xpadding, ypadding)))
	case geom.HorizontalAlignmentCenter:
		dimTb := tb.CalculateDimension(ug.StringBounder())
		tb.DrawU(ug.Apply(klimt.NewUTranslate((width-dimTb.Width())/2, ypadding)))
	}
}

// entityImageStateCommon is upstream's EntityImageStateCommon: the name
// and the colors of a state.
type entityImageStateCommon struct {
	abstractEntityImage
	name klimt.TextBlock
	url  *klimt.URL
}

func newEntityImageStateCommon(entity *abel.Entity) entityImageStateCommon {
	e := entityImageStateCommon{abstractEntityImage: newAbstractEntityImage(entity)}
	styleBuilder := e.skinParam.CurrentStyleBuilder()
	nameFc := styleStateName(entity.Stereotype(), styleBuilder).FontConfiguration(e.skinParam.IHtmlColorSet(),
		entity.Colors())
	horizontalAlignment := styleStateOf(entity.Stereotype(), styleBuilder).HorizontalAlignment()
	e.name = entity.Display().Create8(nameFc, horizontalAlignment, e.skinParam, klimt.CreoleModeFull,
		e.styleState().WrapWidth())
	e.url = entity.URL99()
	return e
}

// styleState is the instance getStyleState.
func (e *entityImageStateCommon) styleState() *klimt.Style {
	return styleStateOf(e.entity.Stereotype(), e.skinParam.CurrentStyleBuilder())
}

// styleStateNameDescription is the instance getStyleStateDescription,
// which upstream builds with the name, not the description, style name.
func (e *entityImageStateCommon) styleStateNameDescription() *klimt.Style {
	return stateStyle().AddSName(klimt.SNameName).WithTOBECHANGED(e.entity.Stereotype()).
		MergedStyle(e.skinParam.CurrentStyleBuilder())
}

// ShapeType returns ROUND_RECTANGLE.
func (e *entityImageStateCommon) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRoundRectangle }

// shape is the rounded box of a state of size dimTotal.
func (e *entityImageStateCommon) shape(dimTotal geom.XDimension2D) *klimt.URectangle {
	corner := e.styleState().Value(klimt.PNameRoundCorner).AsDouble()
	deltaShadow := e.styleState().Shadowing()
	rect := klimt.URectangleBuildDim(dimTotal).Rounded(corner)
	rect.SetDeltaShadow(deltaShadow)
	return rect
}

// applyColor applies the border color, and the background color of the
// entity or of style.
func (e *entityImageStateCommon) applyColor(ug klimt.UGraphic, style *klimt.Style) klimt.UGraphic {
	ug = ug.Apply(e.borderColor())
	backcolor := e.entity.Colors().Color(color.TypeBack)
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(e.skinParam.IHtmlColorSet())
	}
	return ug.Apply(klimt.Bg(backcolor))
}

// borderColor is the line color of the entity, or of the state style.
func (e *entityImageStateCommon) borderColor() color.HColor {
	if border := e.entity.Colors().Color(color.TypeLine); border != nil {
		return border
	}
	return e.styleState().Value(klimt.PNameLineColor).AsColor(e.skinParam.IHtmlColorSet())
}

// The size of the oo symbol of upstream's EntityImageState.
const (
	stateMinWidth     = 50
	stateMinHeight    = 50
	stateSmallRadius  = 3
	stateSmallLine    = 3
	stateSmallMarginX = 7
	stateSmallMarginY = 4
)

// EntityImageState is upstream's EntityImageState: a rounded box with the
// name, a line, and the description lines.
type EntityImageState struct {
	entityImageStateCommon
	fields                          klimt.TextBlock
	withSymbol                      bool
	baseStyle, nameStyle, descStyle *klimt.Style
}

var _ klimt.IEntityImage = (*EntityImageState)(nil)

func newEntityImageState(entity *abel.Entity) *EntityImageState {
	e := &EntityImageState{entityImageStateCommon: newEntityImageStateCommon(entity)}
	stereotype := entity.Stereotype()
	e.withSymbol = stereotype != nil && stereotype.IsWithOOSymbol()
	list := klimt.DisplayCreateSeq(entity.Bodier().RawBody())
	styleBuilder := e.skinParam.CurrentStyleBuilder()
	e.baseStyle = stateStyle().WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
	e.nameStyle = stateStyle().AddSName(klimt.SNameName).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
	e.descStyle = stateStyle().AddSName(klimt.SNameDescription).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
	fieldsFontConfiguration := e.descStyle.FontConfiguration(e.skinParam.IHtmlColorSet(), nil)
	e.fields = list.Create8(fieldsFontConfiguration, e.descStyle.HorizontalAlignment(), e.skinParam,
		klimt.CreoleModeFull, e.styleState().WrapWidth())
	return e
}

// CalculateDimension returns the size of the name above the description,
// with the margins, at least 50 x 50.
func (e *EntityImageState) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dim := e.name.CalculateDimension(stringBounder).MergeTB(e.fields.CalculateDimension(stringBounder))
	heightSymbol := 0.0
	if e.withSymbol {
		heightSymbol += 2*stateSmallRadius + stateSmallMarginY
	}
	result := dim.Delta(klimt.IEntityImageMargin*2 + 2*klimt.IEntityImageMarginLine + heightSymbol)
	return result.AtLeast(stateMinWidth, stateMinHeight)
}

// DrawU draws the state.
func (e *EntityImageState) DrawU(ug klimt.UGraphic) {
	ug.StartGroup(entityGroup(e.entity))
	if e.url != nil {
		ug.StartURL(e.url)
	}
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimName := e.name.CalculateDimension(stringBounder)
	stroke := e.baseStyle.StrokeOf(e.entity.Colors())
	ug = e.applyColor(ug, e.styleState())
	ug = ug.Apply(stroke)
	colorSet := e.skinParam.IHtmlColorSet()
	yLine := klimt.IEntityImageMargin + dimName.Height() + klimt.IEntityImageMarginLine
	northBackcolor := e.nameStyle.Value(klimt.PNameBackGroundColor).AsColor(colorSet)
	southBackcolor := e.descStyle.Value(klimt.PNameBackGroundColor).AsColor(colorSet)
	if northBackcolor.Equals(southBackcolor) {
		ug.Draw(e.shape(dimTotal))
	} else {
		corner := e.styleState().Value(klimt.PNameRoundCorner).AsDouble()
		shadowing := e.styleState().Shadowing()
		newRoundedContainer(e.borderColor(), dimTotal, yLine, 0, northBackcolor, southBackcolor, southBackcolor,
			stroke, corner, shadowing).DrawU(ug)
	}
	ug.Apply(klimt.UTranslateDy(yLine)).Draw(klimt.ULineH(dimTotal.Width()))
	if e.withSymbol {
		drawStateSymbol(ug, dimTotal.Width(), dimTotal.Height())
	}
	xDesc := (dimTotal.Width() - dimName.Width()) / 2
	e.name.DrawU(ug.Apply(klimt.NewUTranslate(xDesc, klimt.IEntityImageMargin)))
	drawAligned(e.descStyle.HorizontalAlignment(), ug, e.fields, klimt.IEntityImageMargin,
		yLine+klimt.IEntityImageMarginLine, dimTotal.Width())
	if e.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

// drawStateSymbol is EntityImageState.drawSymbol: the two linked circles
// of a state with an oo stereotype, in the bottom right corner of
// xSymbol, ySymbol.
func drawStateSymbol(ug klimt.UGraphic, xSymbol, ySymbol float64) {
	xSymbol -= 4*stateSmallRadius + stateSmallLine + stateSmallMarginX
	ySymbol -= 2*stateSmallRadius + stateSmallMarginY
	small := klimt.UEllipseBuild(2*stateSmallRadius, 2*stateSmallRadius)
	ug.Apply(klimt.NewUTranslate(xSymbol, ySymbol)).Draw(small)
	ug.Apply(klimt.NewUTranslate(xSymbol+stateSmallLine+2*stateSmallRadius, ySymbol)).Draw(small)
	ug.Apply(klimt.NewUTranslate(xSymbol+2*stateSmallRadius, ySymbol+stateSmallLine)).
		Draw(klimt.ULineH(stateSmallLine))
}

// EntityImageState2 is upstream's EntityImageState2: a "<<sdlreceive>>"
// state, drawn as a small frame.
type EntityImageState2 struct {
	abstractEntityImage
	url     *klimt.URL
	asSmall klimt.TextBlock
}

var _ klimt.IEntityImage = (*EntityImageState2)(nil)

func newEntityImageState2(entity *abel.Entity) *EntityImageState2 {
	e := &EntityImageState2{abstractEntityImage: newAbstractEntityImage(entity)}
	stereotype := entity.Stereotype()
	sym := usymbolOf(abel.USymbolFrame)
	style := e.style()
	ctx := style.SymbolContext(e.skinParam.IHtmlColorSet(), nil)
	e.url = entity.URL99()
	stereo := klimt.NewTextBlockEmpty(0, 0)
	desc := abel.BodyFactoryCreate2(e.skinParam.DefaultTextAlignment(geom.HorizontalAlignmentCenter),
		entity.Display(), e.skinParam, stereotype, entity, style)
	e.asSmall = sym.AsSmall(nil, desc, stereo, ctx, e.skinParam.StereotypeAlignment())
	return e
}

func (e *EntityImageState2) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, e.styleName(), klimt.SNameState).
		MergedStyle(e.skinParam.CurrentStyleBuilder())
}

// ShapeType returns RECTANGLE.
func (e *EntityImageState2) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// CalculateDimension returns the size of the frame.
func (e *EntityImageState2) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.asSmall.CalculateDimension(stringBounder)
}

// DrawU draws the frame.
func (e *EntityImageState2) DrawU(ug klimt.UGraphic) {
	if e.url != nil {
		ug.StartURL(e.url)
	}
	e.asSmall.DrawU(ug)
	if e.url != nil {
		ug.CloseURL()
	}
}

// The minimum size of upstream's EntityImageStateEmptyDescription.
const (
	stateEmptyMinWidth  = 50
	stateEmptyMinHeight = 40
)

// EntityImageStateEmptyDescription is upstream's
// EntityImageStateEmptyDescription: a rounded box with the name only, for
// "hide empty description".
type EntityImageStateEmptyDescription struct {
	entityImageStateCommon
}

var _ klimt.IEntityImage = (*EntityImageStateEmptyDescription)(nil)

func newEntityImageStateEmptyDescription(entity *abel.Entity) *EntityImageStateEmptyDescription {
	return &EntityImageStateEmptyDescription{entityImageStateCommon: newEntityImageStateCommon(entity)}
}

// CalculateDimension returns the size of the name with the margins, at
// least 50 x 40.
func (e *EntityImageStateEmptyDescription) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.name.CalculateDimension(stringBounder).Delta(klimt.IEntityImageMargin*2).
		AtLeast(stateEmptyMinWidth, stateEmptyMinHeight)
}

// DrawU draws the state.
func (e *EntityImageStateEmptyDescription) DrawU(ug klimt.UGraphic) {
	if e.url != nil {
		ug.StartURL(e.url)
	}
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimHeader := e.name.CalculateDimension(stringBounder)
	stroke := e.styleState().StrokeOf(e.entity.Colors())
	ug = e.applyColor(ug, e.styleStateNameDescription())
	ug = ug.Apply(stroke)
	ug.Draw(e.shape(dimTotal))
	xDesc := (dimTotal.Width() - dimHeader.Width()) / 2
	yDesc := (dimTotal.Height() - dimHeader.Height()) / 2
	e.name.DrawU(ug.Apply(klimt.NewUTranslate(xDesc, yDesc)))
	if e.url != nil {
		ug.CloseURL()
	}
}
