// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageChenEntity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageChenRelationship.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageChenAttribute.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageChenCircle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// chenAttributeMargin is the MARGIN of upstream's EntityImageChenAttribute:
// the top of its name.
const chenAttributeMargin = 6

// entityImageChen is what upstream repeats in each Chen image: the name, the
// URL, the styles and the colors of an element of a Chen diagram.
type entityImageChen struct {
	abstractEntityImage
	// name is the SName of the element in the chenEerDiagram styles.
	name  klimt.SName
	title klimt.TextBlock
	url   *klimt.URL
}

func newEntityImageChen(entity *abel.Entity, name klimt.SName, underline bool) entityImageChen {
	e := entityImageChen{abstractEntityImage: newAbstractEntityImage(entity), name: name}
	titleFontConfiguration := e.styleOf(klimt.SNameTitle).FontConfiguration(e.skinParam.IHtmlColorSet(),
		entity.Colors())
	if underline {
		titleFontConfiguration = titleFontConfiguration.Underline()
	}
	e.title = entity.Display().Create8(titleFontConfiguration, geom.HorizontalAlignmentCenter, e.skinParam,
		klimt.CreoleModeFull, e.style().WrapWidth())
	e.url = entity.URL99()
	return e
}

// styleOf is upstream's StyleQueries.CHENEER_* of the element, with names
// added, for the stereotype of the entity.
func (e *entityImageChen) styleOf(names ...klimt.SName) *klimt.Style {
	signature := klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameChenEerDiagram, e.name}, names...)...)
	return signature.WithTOBECHANGED(e.entity.Stereotype()).MergedStyle(e.skinParam.CurrentStyleBuilder())
}

// style is the style of the element.
func (e *entityImageChen) style() *klimt.Style { return e.styleOf() }

// hasStereotype reports whether the stereotype of the entity contains
// stereotype.
func (e *entityImageChen) hasStereotype(stereotype string) bool {
	st := e.entity.Stereotype()
	return st != nil && strings.Contains(st.String(), stereotype)
}

// applyColor applies the line and background colors of the entity, or of
// the style.
func (e *entityImageChen) applyColor(ug klimt.UGraphic) klimt.UGraphic {
	colors := e.entity.Colors()
	border := colors.Color(color.TypeLine)
	if border == nil {
		border = e.style().Value(klimt.PNameLineColor).AsColor(e.skinParam.IHtmlColorSet())
	}
	ug = ug.Apply(border)
	backcolor := colors.Color(color.TypeBack)
	if backcolor == nil {
		backcolor = e.style().Value(klimt.PNameBackGroundColor).AsColor(e.skinParam.IHtmlColorSet())
	}
	return ug.Apply(klimt.Bg(backcolor))
}

// startGroup opens the group of the element, with only its id, and its URL.
func (e *entityImageChen) startGroup(ug klimt.UGraphic) {
	ug.StartGroup(klimt.UGroupSingleton(klimt.UGroupID, e.entity.Quark().StringPoint()))
	if e.url != nil {
		ug.StartURL(e.url)
	}
}

// closeGroup closes the URL and the group of the element.
func (e *entityImageChen) closeGroup(ug klimt.UGraphic) {
	if e.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

// EntityImageChenEntity is upstream's EntityImageChenEntity: a rectangle
// around the name, doubled for a "<<weak>>" entity.
type EntityImageChenEntity struct {
	entityImageChen
	isWeak bool
}

var _ klimt.IEntityImage = (*EntityImageChenEntity)(nil)

func newEntityImageChenEntity(entity *abel.Entity) *EntityImageChenEntity {
	e := &EntityImageChenEntity{entityImageChen: newEntityImageChen(entity, klimt.SNameChenEntity, false)}
	e.isWeak = e.hasStereotype("<<weak>>")
	return e
}

// ShapeType returns RECTANGLE.
func (e *EntityImageChenEntity) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// CalculateDimension returns the size of the name with the margins.
func (e *EntityImageChenEntity) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.title.CalculateDimension(stringBounder).Delta(klimt.IEntityImageMargin*2 + 2*klimt.IEntityImageMarginLine)
}

// DrawU draws the rectangle and the name.
func (e *EntityImageChenEntity) DrawU(ug klimt.UGraphic) {
	e.startGroup(ug)
	dimTotal := e.CalculateDimension(ug.StringBounder())
	dimTitle := e.title.CalculateDimension(ug.StringBounder())
	stroke := e.style().StrokeOf(e.entity.Colors())
	ug2 := e.applyColor(ug).Apply(stroke)
	ug2.Draw(klimt.URectangleBuildDim(dimTotal))
	if e.isWeak {
		ug2.Apply(klimt.NewUTranslate(3, 3)).Draw(klimt.URectangleBuildDim(dimTotal.Delta(-6)))
	}
	xTitle := (dimTotal.Width() - dimTitle.Width()) / 2
	e.title.DrawU(ug2.Apply(klimt.NewUTranslate(xTitle, klimt.IEntityImageMargin+klimt.IEntityImageMarginLine)))
	e.closeGroup(ug)
}

// EntityImageChenRelationship is upstream's EntityImageChenRelationship: a
// diamond around the name, doubled for an "<<identifying>>" relationship.
type EntityImageChenRelationship struct {
	entityImageChen
	isIdentifying bool
}

var _ klimt.IEntityImage = (*EntityImageChenRelationship)(nil)

func newEntityImageChenRelationship(entity *abel.Entity) *EntityImageChenRelationship {
	e := &EntityImageChenRelationship{entityImageChen: newEntityImageChen(entity, klimt.SNameChenRelationship,
		false)}
	e.isIdentifying = e.hasStereotype("<<identifying>>")
	return e
}

// ShapeType returns DIAMOND.
func (e *EntityImageChenRelationship) ShapeType() klimt.ShapeType { return klimt.ShapeTypeDiamond }

// CalculateDimension fits a diamond of aspect ratio 2 around the name: its
// diagonal is the name's size projected on (1, 2) / sqrt(5), plus the
// margins.
func (e *EntityImageChenRelationship) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimTitle := e.title.CalculateDimension(stringBounder)
	diagonal := (dimTitle.Width()+2*dimTitle.Height())/math.Sqrt(5) + 2*klimt.IEntityImageMargin
	return geom.NewXDimension2D(diagonal*math.Sqrt(5), diagonal*math.Sqrt(5)/2)
}

// DrawU draws the diamond and the name.
func (e *EntityImageChenRelationship) DrawU(ug klimt.UGraphic) {
	e.startGroup(ug)
	dimTotal := e.CalculateDimension(ug.StringBounder())
	dimTitle := e.title.CalculateDimension(ug.StringBounder())
	stroke := e.style().StrokeOf(e.entity.Colors())
	ug2 := e.applyColor(ug).Apply(stroke)
	ug2.Draw(chenDiamond(dimTotal))
	if e.isIdentifying {
		ug2.Apply(klimt.NewUTranslate(10, 5)).Draw(chenDiamond(dimTotal.DeltaWH(-20, -10)))
	}
	xTitle := (dimTotal.Width() - dimTitle.Width()) / 2
	yTitle := (dimTotal.Height() - dimTitle.Height()) / 2
	e.title.DrawU(ug2.Apply(klimt.NewUTranslate(xTitle, yTitle)))
	e.closeGroup(ug)
}

// chenDiamond is the diamond of a relationship of size dimTotal.
func chenDiamond(dimTotal geom.XDimension2D) *klimt.UPolygon {
	width := dimTotal.Width()
	height := dimTotal.Height()
	return klimt.NewUPolygonPoints([]geom.XPoint2D{
		geom.NewXPoint2D(0, height/2),
		geom.NewXPoint2D(width/2, 0),
		geom.NewXPoint2D(width, height/2),
		geom.NewXPoint2D(width/2, height),
	})
}

// EntityImageChenAttribute is upstream's EntityImageChenAttribute: an
// ellipse around the name, which is underlined for a "<<key>>" attribute,
// doubled for a "<<multi>>" attribute and dashed for a "<<derived>>" one.
type EntityImageChenAttribute struct {
	entityImageChen
	isMulti, isDerived bool
}

var _ klimt.IEntityImage = (*EntityImageChenAttribute)(nil)

func newEntityImageChenAttribute(entity *abel.Entity) *EntityImageChenAttribute {
	return newEntityImageChenOval(entity, klimt.SNameChenAttribute)
}

// newEntityImageChenOval builds an attribute or a circle, which upstream
// writes as two classes with the same fields.
func newEntityImageChenOval(entity *abel.Entity, name klimt.SName) *EntityImageChenAttribute {
	isKey := entity.Stereotype() != nil && strings.Contains(entity.Stereotype().String(), "<<key>>")
	e := &EntityImageChenAttribute{entityImageChen: newEntityImageChen(entity, name, isKey)}
	e.isMulti = e.hasStereotype("<<multi>>")
	e.isDerived = e.hasStereotype("<<derived>>")
	return e
}

// ShapeType returns OVAL.
func (e *EntityImageChenAttribute) ShapeType() klimt.ShapeType { return klimt.ShapeTypeOval }

// CalculateDimension returns the size of the smallest ellipse around the
// name, upstream's TextBlockInEllipse.
func (e *EntityImageChenAttribute) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return klimt.NewTextBlockInContainingEllipse(e.title, stringBounder).CalculateDimension(stringBounder)
}

// DrawU draws the ellipse and the name.
func (e *EntityImageChenAttribute) DrawU(ug klimt.UGraphic) {
	e.drawOval(ug, e.CalculateDimension(ug.StringBounder()), func(geom.XDimension2D, geom.XDimension2D) float64 {
		return chenAttributeMargin
	})
}

// drawOval draws the ellipse of size dimTotal and the name at the y that
// yTitle returns for the sizes of the ellipse and of the name.
func (e *EntityImageChenAttribute) drawOval(ug klimt.UGraphic, dimTotal geom.XDimension2D,
	yTitle func(dimTotal, dimTitle geom.XDimension2D) float64) {
	e.startGroup(ug)
	dimTitle := e.title.CalculateDimension(ug.StringBounder())
	stroke := e.style().StrokeOf(e.entity.Colors())
	if e.isDerived {
		stroke = klimt.NewUStroke(10, 10, stroke.Thickness())
	}
	ug2 := e.applyColor(ug).Apply(stroke)
	ug2.Draw(klimt.UEllipseBuild(dimTotal.Width(), dimTotal.Height()))
	if e.isMulti {
		inner := dimTotal.Delta(-6)
		ug2.Apply(klimt.NewUTranslate(3, 3)).Draw(klimt.UEllipseBuild(inner.Width(), inner.Height()))
	}
	xTitle := (dimTotal.Width() - dimTitle.Width()) / 2
	e.title.DrawU(ug2.Apply(klimt.NewUTranslate(xTitle, yTitle(dimTotal, dimTitle))))
	e.closeGroup(ug)
}

// EntityImageChenCircle is upstream's EntityImageChenCircle: the circle of
// a group of subclasses, 25 pixels wide, with its symbol in the middle.
type EntityImageChenCircle struct {
	EntityImageChenAttribute
}

var _ klimt.IEntityImage = (*EntityImageChenCircle)(nil)

func newEntityImageChenCircle(entity *abel.Entity) *EntityImageChenCircle {
	return &EntityImageChenCircle{EntityImageChenAttribute: *newEntityImageChenOval(entity, klimt.SNameCircle)}
}

// CalculateDimension returns 25 x 25.
func (e *EntityImageChenCircle) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(25, 25)
}

// DrawU draws the circle and the symbol.
func (e *EntityImageChenCircle) DrawU(ug klimt.UGraphic) {
	e.drawOval(ug, e.CalculateDimension(ug.StringBounder()), func(dimTotal, dimTitle geom.XDimension2D) float64 {
		return (dimTotal.Height() - dimTitle.Height()) / 2
	})
}
