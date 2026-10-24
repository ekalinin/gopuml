// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageJson.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageEmptyPackage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// EntityImageJson is the image of a JSON element: its name and its data.
type EntityImageJson struct {
	abstractEntityImage
	memo        klimt.TextBlockMemoized
	name        klimt.TextBlock
	stereo      klimt.TextBlock
	entries     klimt.TextBlock
	url         *klimt.URL
	roundCorner float64
}

var (
	_ klimt.IEntityImage = (*EntityImageJson)(nil)
	_ klimt.Stencil      = (*EntityImageJson)(nil)
	_ klimt.WithPorts    = (*EntityImageJson)(nil)
)

func newEntityImageJson(entity *abel.Entity, portionShower PortionShower) *EntityImageJson {
	j := &EntityImageJson{abstractEntityImage: newAbstractEntityImage(entity)}
	j.roundCorner = j.style().Value(klimt.PNameRoundCorner).AsDouble()
	fcHeader := j.styleHeader().FontConfiguration(j.skinParam.IHtmlColorSet(), nil)
	j.name = klimt.TextBlockUtilsWithMargin(entity.Display().Create(fcHeader, geom.HorizontalAlignmentCenter,
		j.skinParam), 2, 2)
	j.stereo = objectStereo(entity, portionShower, j.skinParam)
	fontConfiguration := j.style().FontConfiguration(j.skinParam.IHtmlColorSet(), nil)
	j.entries = entityBody(entity, j.skinParam, false, false, entity.Stereotype(), j.style(), fontConfiguration)
	j.url = entity.URL99()
	return j
}

func (j *EntityImageJson) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameJson).WithTOBECHANGED(j.entity.Stereotype()).MergedStyle(j.skinParam.CurrentStyleBuilder())
}

func (j *EntityImageJson) styleHeader() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameJson, klimt.SNameHeader).WithTOBECHANGED(j.entity.Stereotype()).
		MergedStyle(j.skinParam.CurrentStyleBuilder())
}

// Ports returns the ports of the data, below the name.
func (j *EntityImageJson) Ports(stringBounder font.StringBounder) *klimt.Ports {
	dimTitle := nameAndStereotypeDimension(j.name, j.stereo, stringBounder)
	return j.entries.(klimt.WithPorts).Ports(stringBounder).TranslateY(dimTitle.Height())
}

// CalculateDimension returns the size of the element.
func (j *EntityImageJson) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return j.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		dimTitle := nameAndStereotypeDimension(j.name, j.stereo, sb)
		dimFields := j.entries.CalculateDimension(sb)
		width := math.Max(dimFields.Width(), dimTitle.Width()+2*objectXMarginCircle)
		if minimumWidth := j.style().Value(klimt.PNameMinimumWidth).AsDouble(); width < minimumWidth {
			width = minimumWidth
		}
		fieldsHeight := dimFields.Height()
		if fieldsHeight == 0 && j.entity.LeafType() != abel.LeafTypeMap {
			fieldsHeight = objectMarginEmptyFieldsOrMethod
		}
		return geom.NewXDimension2D(width, fieldsHeight+dimTitle.Height())
	})
}

// DrawU draws the element.
func (j *EntityImageJson) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := j.CalculateDimension(stringBounder)
	dimTitle := nameAndStereotypeDimension(j.name, j.stereo, stringBounder)

	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
	rect := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(j.roundCorner)
	headerBackcolor := j.entity.Colors().Color(color.TypeHeader)
	backcolor := j.entity.Colors().Color(color.TypeBack)
	style := j.style()
	set := j.skinParam.IHtmlColorSet()
	borderColor := style.Value(klimt.PNameLineColor).AsColor(set)
	if headerBackcolor == nil {
		if backcolor == nil {
			headerBackcolor = j.styleHeader().Value(klimt.PNameBackGroundColor).AsColor(set)
		} else {
			headerBackcolor = backcolor
		}
	}
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(set)
	}
	rect.SetDeltaShadow(style.Shadowing())
	stroke := style.Stroke()

	ug = ug.Apply(borderColor).Apply(klimt.Bg(backcolor))
	if j.url != nil {
		ug.StartURL(j.url)
	}
	ug.StartGroup(entityGroup(j.entity))
	ug.Apply(stroke).Draw(rect)
	if headerBackcolor != nil && !backcolor.Equals(headerBackcolor) {
		rect2 := klimt.URectangleBuild(widthTotal, dimTitle.Height()).HalfRounded(j.roundCorner)
		ugHeader := ug.Apply(klimt.Bg(headerBackcolor))
		ugHeader.Apply(stroke).Draw(rect2)
	}
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(ug.StringBounder()))
	if j.stereo != nil {
		header.Add(j.stereo)
	}
	header.Add(j.name)
	header.DrawU(ug, dimTotal.Width(), dimTitle.Height())

	ug2 := klimt.UGraphicStencilCreate(ug, j, stroke)
	j.entries.(*abel.TextBlockCucaJSon).SetTotalWidth(dimTotal.Width())
	j.entries.DrawU(ug2.Apply(klimt.UTranslateDy(dimTitle.Height())))

	ug.CloseGroup()
	if j.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns RECTANGLE_HTML_FOR_PORTS.
func (j *EntityImageJson) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangleHTMLForPorts }

// StartingX implements klimt.Stencil.
func (j *EntityImageJson) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX implements klimt.Stencil.
func (j *EntityImageJson) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return j.CalculateDimension(stringBounder).Width()
}

const emptyPackageMargin = 10

// EntityImageEmptyPackage is the image of a package without content,
// drawn as a node of the graph.
type EntityImageEmptyPackage struct {
	abstractEntityImage
	memo           klimt.TextBlockMemoized
	desc           klimt.TextBlock
	stereotype     *klimt.Stereotype
	stereoBlock    klimt.TextBlock
	url            *klimt.URL
	shadowing      float64
	borderColor    color.HColor
	stroke         klimt.UStroke
	roundCorner    float64
	diagonalCorner float64
	back           color.HColor
}

var _ klimt.IEntityImage = (*EntityImageEmptyPackage)(nil)

func newEntityImageEmptyPackage(entity *abel.Entity, portionShower PortionShower) *EntityImageEmptyPackage {
	p := &EntityImageEmptyPackage{abstractEntityImage: newAbstractEntityImage(entity)}
	colors := entity.Colors()
	specificBackColor := colors.Color(color.TypeBack)
	p.stereotype = entity.Stereotype()
	p.url = entity.URL99()

	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, p.styleName(), klimt.SNamePackage,
		klimt.SNameTitle).WithTOBECHANGED(p.stereotype).MergedStyle(p.skinParam.CurrentStyleBuilder())
	style = style.EventuallyOverrideColors(colors)
	set := p.skinParam.IHtmlColorSet()
	p.borderColor = style.Value(klimt.PNameLineColor).AsColor(set)
	p.shadowing = style.Shadowing()
	p.stroke = style.StrokeOf(colors)
	p.roundCorner = style.Value(klimt.PNameRoundCorner).AsDouble()
	p.diagonalCorner = style.Value(klimt.PNameDiagonalCorner).AsDouble()
	if specificBackColor == nil {
		p.back = style.Value(klimt.PNameBackGroundColor).AsColor(set)
	} else {
		p.back = specificBackColor
	}
	titleFontConfiguration := style.FontConfiguration(set, nil)
	titleHorizontalAlignment := style.HorizontalAlignment()
	p.desc = entity.Display().Create(titleFontConfiguration, titleHorizontalAlignment, p.skinParam)

	if legend := entity.Legend(); legend != nil {
		p.stereoBlock = entityImageLegend(legend.Display(), p.skinParam)
	} else if p.stereotype == nil || !portionShower.ShowPortion(uml.PortionStereotype, entity) {
		p.stereoBlock = klimt.TextBlockUtilsEmpty(0, 0)
	} else if _, ok := p.stereotype.Label(klimt.GuillemetDoubleComparator()); !ok {
		p.stereoBlock = klimt.TextBlockUtilsEmpty(0, 0)
	} else {
		p.stereoBlock = klimt.TextBlockUtilsWithMargin(klimt.DisplayCreate(
			p.stereotype.Labels(p.skinParam.Guillemet())...).Create(
			klimt.FontConfigurationFromSkinParam(p.skinParam, klimt.FontParamPackageStereotype, p.stereotype),
			titleHorizontalAlignment, p.skinParam), 1, 0)
	}
	return p
}

// CalculateDimension returns the size of the package.
func (p *EntityImageEmptyPackage) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return p.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		dimDesc := p.desc.CalculateDimension(sb)
		dim := klimt.TextBlockUtilsMergeTB(p.desc, p.stereoBlock, geom.HorizontalAlignmentLeft).CalculateDimension(sb)
		dim = dim.AtLeast(0, 2*dimDesc.Height())
		return dim.Delta(emptyPackageMargin * 2)
	})
}

// DrawU draws the package with the package style of the diagram.
func (p *EntityImageEmptyPackage) DrawU(ug klimt.UGraphic) {
	if p.url != nil {
		ug.StartURL(p.url)
	}
	stringBounder := ug.StringBounder()
	dimTotal := p.CalculateDimension(stringBounder)
	rectangleArea := geom.NewRectangleArea(0, 0, dimTotal.Width(), dimTotal.Height())
	decoration := newClusterDecoration(p.skinParam.PackageStyle(), nil, p.desc, p.stereoBlock, rectangleArea,
		p.stroke)
	horizontalAlignment := p.skinParam.HorizontalAlignment(klimt.AlignmentParamPackageTitleAlignment,
		klimt.ArrowDirectionNull, false, klimt.HorizontalAlignmentNull)
	stereotypeAlignment := p.skinParam.StereotypeAlignment()
	decoration.drawU(ug, p.back, p.borderColor, p.shadowing, p.roundCorner, horizontalAlignment, stereotypeAlignment,
		p.diagonalCorner)
	if p.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns RECTANGLE.
func (p *EntityImageEmptyPackage) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }
