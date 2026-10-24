// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageAssociation.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageAssociationPoint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageLollipopInterface.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const associationSize = 12

// EntityImageAssociation is the diamond of an n-ary association
// ("<>").
type EntityImageAssociation struct {
	abstractEntityImage
}

var _ klimt.IEntityImage = (*EntityImageAssociation)(nil)

func newEntityImageAssociation(entity *abel.Entity) *EntityImageAssociation {
	return &EntityImageAssociation{abstractEntityImage: newAbstractEntityImage(entity)}
}

// CalculateDimension returns the size of the diamond.
func (a *EntityImageAssociation) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(associationSize*2, associationSize*2)
}

// DrawU draws the diamond.
func (a *EntityImageAssociation) DrawU(ug klimt.UGraphic) {
	diams := klimt.NewUPolygon()
	diams.AddPoint(associationSize, 0)
	diams.AddPoint(associationSize*2, associationSize)
	diams.AddPoint(associationSize, associationSize*2)
	diams.AddPoint(0, associationSize)
	diams.AddPoint(associationSize, 0)

	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, a.styleName(), klimt.SNameDiamond).
		MergedStyle(a.skinParam.CurrentStyleBuilder())
	set := a.skinParam.IHtmlColorSet()
	borderColor := style.Value(klimt.PNameLineColor).AsColor(set)
	backgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(set)
	diams.SetDeltaShadow(style.Shadowing())
	ug.Apply(borderColor).Apply(klimt.Bg(backgroundColor)).Apply(style.Stroke()).Draw(diams)
}

// ShapeType returns DIAMOND.
func (a *EntityImageAssociation) ShapeType() klimt.ShapeType { return klimt.ShapeTypeDiamond }

const associationPointSize = 4

// EntityImageAssociationPoint is the point where the links of an
// association class meet.
type EntityImageAssociationPoint struct {
	abstractEntityImage
}

var _ klimt.IEntityImage = (*EntityImageAssociationPoint)(nil)

func newEntityImageAssociationPoint(entity *abel.Entity) *EntityImageAssociationPoint {
	return &EntityImageAssociationPoint{abstractEntityImage: newAbstractEntityImage(entity)}
}

// CalculateDimension returns the size of the point.
func (a *EntityImageAssociationPoint) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(associationPointSize, associationPointSize)
}

// DrawU draws the point in the color of the arrows.
func (a *EntityImageAssociationPoint) DrawU(ug klimt.UGraphic) {
	circle := klimt.UEllipseBuild(associationPointSize, associationPointSize)
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameClassDiagram,
		klimt.SNameArrow).WithTOBECHANGED(a.stereo()).MergedStyle(a.skinParam.CurrentStyleBuilder())
	c := style.Value(klimt.PNameLineColor).AsColor(a.skinParam.IHtmlColorSet())
	ug.Apply(c).Apply(klimt.CopyForegroundColorToBackgroundColor{}).Draw(circle)
}

// ShapeType returns CIRCLE.
func (a *EntityImageAssociationPoint) ShapeType() klimt.ShapeType { return klimt.ShapeTypeCircle }

const lollipopSize = 10

// EntityImageLollipopInterface is the circle (or half circle) of a
// lollipop interface ("()-" and "(-").
type EntityImageLollipopInterface struct {
	abstractEntityImage
	desc  klimt.TextBlock
	url   *klimt.URL
	angle float64
}

var _ klimt.IEntityImage = (*EntityImageLollipopInterface)(nil)

func newEntityImageLollipopInterface(entity *abel.Entity) *EntityImageLollipopInterface {
	l := &EntityImageLollipopInterface{abstractEntityImage: newAbstractEntityImage(entity)}
	fc := klimt.FontConfigurationFromStyle(l.skinParam, l.style(), nil)
	l.desc = entity.Display().Create(fc, geom.HorizontalAlignmentCenter, l.skinParam)
	l.url = entity.URL99()
	return l
}

func (l *EntityImageLollipopInterface) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, l.styleName(), klimt.SNameCircle).
		WithTOBECHANGED(l.stereo()).MergedStyle(l.skinParam.CurrentStyleBuilder())
}

// CalculateDimension returns the size of the circle.
func (l *EntityImageLollipopInterface) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(lollipopSize, lollipopSize)
}

// DrawU draws the circle and the name below it.
func (l *EntityImageLollipopInterface) DrawU(ug klimt.UGraphic) {
	style := l.style()
	set := l.skinParam.IHtmlColorSet()
	backgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(set)
	borderColor := style.Value(klimt.PNameLineColor).AsColor(set)
	shadow := style.Shadowing()

	var circle *klimt.UEllipse
	if l.entity.LeafType() == abel.LeafTypeLollipopHalf {
		circle = klimt.NewUEllipse(lollipopSize, lollipopSize, l.angle-90, 180)
	} else {
		circle = klimt.UEllipseBuild(lollipopSize, lollipopSize)
		if l.skinParam.Shadowing(l.entity.Stereotype()) {
			circle.SetDeltaShadow(shadow)
		}
	}
	ug = ug.Apply(klimt.Bg(backgroundColor)).Apply(borderColor)
	if l.url != nil {
		ug.StartURL(l.url)
	}
	ug.StartGroup(entityGroup(l.entity))
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Draw(circle)
	ug.CloseGroup()

	dimDesc := l.desc.CalculateDimension(ug.StringBounder())
	widthDesc := dimDesc.Width()
	x := lollipopSize/2 - widthDesc/2
	y := float64(lollipopSize)
	l.desc.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
	if l.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns CIRCLE.
func (l *EntityImageLollipopInterface) ShapeType() klimt.ShapeType { return klimt.ShapeTypeCircle }

// AddImpact records the angle of the link of a half lollipop.
func (l *EntityImageLollipopInterface) AddImpact(angle float64) {
	l.angle = 180 - (angle * 180 / math.Pi)
}
