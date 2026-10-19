// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/InnerStateAutonom.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// InnerStateAutonom is upstream's InnerStateAutonom: a composite state laid
// out on its own, drawn as a rounded box with its name, its description
// and the image of its inside.
type InnerStateAutonom struct {
	klimt.TextBlockDefaults
	entity                                          *abel.Entity
	im                                              klimt.IEntityImage
	name, attribute                                 klimt.TextBlock
	borderColor                                     color.HColor
	url                                             *klimt.URL
	withSymbol                                      bool
	stroke                                          klimt.UStroke
	rounded, shadowing                              float64
	styleDescription                                *klimt.Style
	northBackcolor, centerBackColor, southBackcolor color.HColor
}

var _ klimt.IEntityImage = (*InnerStateAutonom)(nil)

func newInnerStateAutonom(im klimt.IEntityImage, group *abel.Entity) *InnerStateAutonom {
	if im == nil {
		throwNPE()
	}
	s := &InnerStateAutonom{im: im, entity: group}
	skinParam := group.SkinParam()
	styleBuilder := skinParam.CurrentStyleBuilder()
	stereotype := group.Stereotype()
	styleName := styleStateName(stereotype, styleBuilder)
	style := styleStateOf(stereotype, styleBuilder)
	s.styleDescription = styleStateDescription(stereotype, styleBuilder)
	s.rounded = style.Value(klimt.PNameRoundCorner).AsDouble()
	s.shadowing = style.Shadowing()
	colorSet := skinParam.IHtmlColorSet()
	titleFontConfiguration := styleName.FontConfiguration(colorSet, nil)
	s.name = group.Display().Create(titleFontConfiguration, styleName.HorizontalAlignment(), skinParam)
	s.attribute = stateDescription(group, skinParam)
	s.withSymbol = stereotype != nil && stereotype.IsWithOOSymbol()
	s.url = group.URL99()
	if stroke := klimt.ColorsSpecificLineStroke(group.Colors()); stroke != nil {
		s.stroke = *stroke
	} else {
		s.stroke = style.Stroke()
	}
	s.borderColor = group.Colors().Color(color.TypeLine)
	if s.borderColor == nil {
		s.borderColor = style.Value(klimt.PNameLineColor).AsColor(colorSet)
	}
	s.northBackcolor = group.Colors().Color(color.TypeBack)
	s.centerBackColor = s.northBackcolor
	s.southBackcolor = s.northBackcolor
	if s.northBackcolor == nil {
		backColor := func(part klimt.SName) color.HColor {
			return stateStyle().AddSName(part).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder).
				Value(klimt.PNameBackGroundColor).AsColor(colorSet)
		}
		s.northBackcolor = backColor(klimt.SNameName)
		s.centerBackColor = backColor(klimt.SNameDescription)
		s.southBackcolor = backColor(klimt.SNameBody)
	}
	return s
}

// DrawU draws the box, then the name, the description and the inside.
func (s *InnerStateAutonom) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	text := s.name.CalculateDimension(stringBounder)
	attr := s.attribute.CalculateDimension(stringBounder)
	total := s.CalculateDimension(stringBounder)
	marginForFields := 0.0
	if attr.Height() > 0 {
		marginForFields = klimt.IEntityImageMargin
	}
	nameHeight := klimt.IEntityImageMargin + text.Height() + klimt.IEntityImageMarginLine
	descriptionHeight := attr.Height() + marginForFields
	r := newRoundedContainer(s.borderColor, total, nameHeight, descriptionHeight, s.northBackcolor,
		s.centerBackColor, s.southBackcolor, s.stroke, s.rounded, s.shadowing)
	ug.StartGroup(entityGroup(s.entity))
	if s.url != nil {
		ug.StartURL(s.url)
	}
	r.DrawU(ug)
	s.name.DrawU(ug.Apply(klimt.NewUTranslate((total.Width()-text.Width())/2, klimt.IEntityImageMargin)))
	drawAligned(s.styleDescription.HorizontalAlignment(), ug, s.attribute, klimt.IEntityImageMargin,
		klimt.IEntityImageMargin+text.Height()+klimt.IEntityImageMargin, total.Width())
	s.im.DrawU(ug.Apply(klimt.NewUTranslate(klimt.IEntityImageMargin, s.spaceYforURL(stringBounder))))
	if s.withSymbol {
		drawStateSymbol(ug.Apply(s.borderColor), total.Width(), total.Height())
	}
	if s.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

// spaceYforURL is the top of the inside, below the name and description.
func (s *InnerStateAutonom) spaceYforURL(stringBounder font.StringBounder) float64 {
	text := s.name.CalculateDimension(stringBounder)
	attr := s.attribute.CalculateDimension(stringBounder)
	marginForFields := 0.0
	if attr.Height() > 0 {
		marginForFields = klimt.IEntityImageMargin
	}
	titreHeight := klimt.IEntityImageMargin + text.Height() + klimt.IEntityImageMarginLine
	suppY := titreHeight + marginForFields + attr.Height()
	return suppY + klimt.IEntityImageMarginLine
}

// CalculateDimension returns the size of the name, the description and
// the inside, with the margins.
func (s *InnerStateAutonom) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	img := s.im.CalculateDimension(stringBounder)
	text := s.name.CalculateDimension(stringBounder)
	attr := s.attribute.CalculateDimension(stringBounder)
	dim := text.MergeTB3(attr, img)
	marginForFields := 0.0
	if attr.Height() > 0 {
		marginForFields = klimt.IEntityImageMargin
	}
	return dim.Delta(klimt.IEntityImageMargin*2 + 2*klimt.IEntityImageMarginLine + marginForFields)
}

// ShapeType returns ROUND_RECTANGLE.
func (s *InnerStateAutonom) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRoundRectangle }

// Shield returns no margin.
func (s *InnerStateAutonom) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// IsHidden reports whether the inside is hidden.
func (s *InnerStateAutonom) IsHidden() bool { return s.im.IsHidden() }

// OverscanX returns 0.
func (s *InnerStateAutonom) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (s *InnerStateAutonom) IsCrash() bool { return false }
