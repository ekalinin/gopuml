// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageActivity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The corner and margin of upstream's EntityImageActivity.
const (
	activityCorner = 25
	activityMargin = 10
)

// EntityImageActivity is upstream's EntityImageActivity: an activity of a
// legacy activity diagram, drawn as a rounded box, or as the octagon of its
// node with "skinparam activityShape octagon".
type EntityImageActivity struct {
	abstractEntityImage
	shadowing   float64
	desc        klimt.TextBlock
	url         *klimt.URL
	bibliotekon *Bibliotekon
}

var _ klimt.IEntityImage = (*EntityImageActivity)(nil)

func newEntityImageActivity(entity *abel.Entity, bibliotekon *Bibliotekon) *EntityImageActivity {
	e := &EntityImageActivity{abstractEntityImage: newAbstractEntityImage(entity), bibliotekon: bibliotekon}
	style := e.style()
	fontConfiguration := style.FontConfiguration(e.skinParam.IHtmlColorSet(), nil)
	e.shadowing = style.Shadowing()
	e.desc = entity.Display().Create(fontConfiguration, style.HorizontalAlignment(), e.skinParam)
	e.url = entity.URL99()
	return e
}

func (e *EntityImageActivity) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity).WithTOBECHANGED(e.stereo()).MergedStyle(e.skinParam.CurrentStyleBuilder())
}

// CalculateDimension returns the size of the text with the margins.
func (e *EntityImageActivity) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.desc.CalculateDimension(stringBounder).Delta(activityMargin * 2)
}

// DrawU draws the activity.
func (e *EntityImageActivity) DrawU(ug klimt.UGraphic) {
	if e.url != nil {
		ug.StartURL(e.url)
	}
	switch e.ShapeType() {
	case klimt.ShapeTypeRoundRectangle:
		e.drawNormal(ug)
	case klimt.ShapeTypeOctagon:
		e.drawOctagon(ug)
	default:
		throwUnsupported("")
	}
	if e.url != nil {
		ug.CloseURL()
	}
}

// drawOctagon draws the octagon of the node, or the rounded box when the
// node has none.
func (e *EntityImageActivity) drawOctagon(ug klimt.UGraphic) {
	node := e.bibliotekon.node(e.entity)
	octagon := node.polygon
	if octagon == nil {
		e.drawNormal(ug)
		return
	}
	octagon.SetDeltaShadow(e.shadowing)
	ug = e.applyColors(ug)
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Draw(octagon)
	e.desc.DrawU(ug.Apply(klimt.NewUTranslate(activityMargin, activityMargin)))
}

func (e *EntityImageActivity) drawNormal(ug klimt.UGraphic) {
	dimTotal := e.CalculateDimension(ug.StringBounder())
	rect := klimt.URectangleBuild(dimTotal.Width(), dimTotal.Height()).Rounded(activityCorner)
	rect.SetDeltaShadow(e.shadowing)
	ug = e.applyColors(ug)
	ug.Apply(e.style().Stroke()).Draw(rect)
	e.desc.DrawU(ug.Apply(klimt.NewUTranslate(activityMargin, activityMargin)))
}

func (e *EntityImageActivity) applyColors(ug klimt.UGraphic) klimt.UGraphic {
	style := e.style()
	borderColor := style.Value(klimt.PNameLineColor).AsColor(e.skinParam.IHtmlColorSet())
	backcolor := e.entity.Colors().Color(color.TypeBack)
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(e.skinParam.IHtmlColorSet())
	}
	return ug.Apply(borderColor).Apply(klimt.Bg(backcolor))
}

// ShapeType returns OCTAGON with "skinparam activityShape octagon",
// ROUND_RECTANGLE otherwise.
func (e *EntityImageActivity) ShapeType() klimt.ShapeType {
	if e.skinParam.UseOctagonForActivity(e.stereo()) {
		return klimt.ShapeTypeOctagon
	}
	return klimt.ShapeTypeRoundRectangle
}
