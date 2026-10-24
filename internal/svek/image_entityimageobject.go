// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageObject.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageMap.java
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

const (
	objectMarginEmptyFieldsOrMethod = 13
	objectXMarginCircle             = 5
)

// EntityImageObject is the image of an object: its underlined name and its
// fields.
type EntityImageObject struct {
	abstractEntityImage
	memo        klimt.TextBlockMemoized
	name        klimt.TextBlock
	stereo      klimt.TextBlock
	fields      klimt.TextBlock
	url         *klimt.URL
	roundCorner float64
	showFields  bool
}

var (
	_ klimt.IEntityImage = (*EntityImageObject)(nil)
	_ klimt.Stencil      = (*EntityImageObject)(nil)
	_ klimt.WithPorts    = (*EntityImageObject)(nil)
)

func newEntityImageObject(entity *abel.Entity, portionShower PortionShower) *EntityImageObject {
	o := &EntityImageObject{abstractEntityImage: newAbstractEntityImage(entity)}
	o.roundCorner = o.style().Value(klimt.PNameRoundCorner).AsDouble()
	fcHeader := o.styleHeader().FontConfiguration(o.skinParam.IHtmlColorSet(), nil)
	tmp := o.underlinedName(entity).Create(fcHeader, geom.HorizontalAlignmentCenter, o.skinParam)
	o.name = klimt.TextBlockUtilsWithMargins(tmp, o.style().Padding())
	o.stereo = objectStereo(entity, portionShower, o.skinParam)
	o.showFields = portionShower.ShowPortion(uml.PortionField, entity)
	if entity.Bodier().FieldsToDisplay().Size() == 0 && o.showFields {
		o.fields = klimt.NewTextBlockLineBefore(o.style().Value(klimt.PNameLineThickness).AsDouble(),
			klimt.NewTextBlockEmpty(10, 16), 0, nil)
	} else {
		o.fields = entityBody(entity, o.skinParam, false, o.showFields, entity.Stereotype(), o.style(), nil)
	}
	o.url = entity.URL99()
	return o
}

// objectStereo is the stereotype of an object or a map, nil when it is
// not shown.
func objectStereo(entity *abel.Entity, portionShower PortionShower, skinParam klimt.ISkinParam) klimt.TextBlock {
	stereotype := entity.Stereotype()
	if stereotype == nil || !portionShower.ShowPortion(uml.PortionStereotype, entity) {
		return nil
	}
	if _, ok := stereotype.Label(klimt.GuillemetDoubleComparator()); !ok {
		return nil
	}
	return klimt.DisplayCreate(stereotype.Labels(skinParam.Guillemet())...).Create(
		klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamObjectStereotype, stereotype),
		geom.HorizontalAlignmentCenter, skinParam)
}

func (o *EntityImageObject) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameObject).WithTOBECHANGED(o.entity.Stereotype()).MergedStyle(o.skinParam.CurrentStyleBuilder())
}

func (o *EntityImageObject) styleHeader() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameObject, klimt.SNameHeader).WithTOBECHANGED(o.entity.Stereotype()).
		MergedStyle(o.skinParam.CurrentStyleBuilder())
}

func (o *EntityImageObject) underlinedName(entity *abel.Entity) *klimt.Display {
	if o.skinParam.StrictUmlStyle() {
		return entity.Display().UnderlinedName()
	}
	return entity.Display()
}

// CalculateDimension returns the size of the object.
func (o *EntityImageObject) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return o.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		dimTitle := nameAndStereotypeDimension(o.name, o.stereo, sb)
		dimFields := o.fields.CalculateDimension(sb)
		width := math.Max(dimFields.Width(), dimTitle.Width()+2*objectXMarginCircle)
		if minClassWidth := o.style().Value(klimt.PNameMinimumWidth).AsDouble(); width < minClassWidth {
			width = minClassWidth
		}
		height := o.methodOrFieldHeight(dimFields) + dimTitle.Height()
		return geom.NewXDimension2D(width, height)
	})
}

// DrawU draws the object.
func (o *EntityImageObject) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := o.CalculateDimension(stringBounder)
	dimTitle := nameAndStereotypeDimension(o.name, o.stereo, stringBounder)

	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
	rect := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(o.roundCorner)

	backcolor := o.entity.Colors().Color(color.TypeBack)
	headerBackcolor := o.entity.Colors().Color(color.TypeHeader)
	style := o.style()
	set := o.skinParam.IHtmlColorSet()
	borderColor := style.Value(klimt.PNameLineColor).AsColor(set)
	if headerBackcolor == nil {
		if backcolor == nil {
			headerBackcolor = o.styleHeader().Value(klimt.PNameBackGroundColor).AsColor(set)
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
	if o.url != nil {
		ug.StartURL(o.url)
	}
	ug.StartGroup(entityGroup(o.entity))
	ug.Apply(stroke).Draw(rect)

	if headerBackcolor != nil && !backcolor.Equals(headerBackcolor) {
		rect2 := klimt.URectangleBuild(widthTotal, dimTitle.Height()).HalfRounded(o.roundCorner)
		ugHeader := ug.Apply(klimt.Bg(headerBackcolor))
		ugHeader.Apply(stroke).Draw(rect2)
	}
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(stringBounder))
	if o.stereo != nil {
		header.Add(o.stereo)
	}
	header.Add(o.name)
	header.DrawU(ug, dimTotal.Width(), dimTitle.Height())

	ug2 := klimt.UGraphicStencilCreate(ug, o, stroke)
	o.fields.DrawU(ug2.Apply(klimt.UTranslateDy(dimTitle.Height())))

	if o.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

func (o *EntityImageObject) methodOrFieldHeight(dim geom.XDimension2D) float64 {
	fieldsHeight := dim.Height()
	if fieldsHeight == 0 && o.showFields {
		return objectMarginEmptyFieldsOrMethod
	}
	return fieldsHeight
}

func nameAndStereotypeDimension(name, stereo klimt.TextBlock, stringBounder font.StringBounder) geom.XDimension2D {
	nameDim := name.CalculateDimension(stringBounder)
	stereoDim := geom.NewXDimension2D(0, 0)
	if stereo != nil {
		stereoDim = stereo.CalculateDimension(stringBounder)
	}
	return geom.NewXDimension2D(math.Max(nameDim.Width(), stereoDim.Width()), nameDim.Height()+stereoDim.Height())
}

// ShapeType returns RECTANGLE_HTML_FOR_PORTS when fields are used as
// ports, else RECTANGLE.
func (o *EntityImageObject) ShapeType() klimt.ShapeType {
	if len(o.entity.PortShortNames()) > 0 {
		return klimt.ShapeTypeRectangleHTMLForPorts
	}
	return klimt.ShapeTypeRectangle
}

// StartingX implements klimt.Stencil.
func (o *EntityImageObject) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX implements klimt.Stencil.
func (o *EntityImageObject) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return o.CalculateDimension(stringBounder).Width()
}

// Ports returns the ports of the fields, below the name.
func (o *EntityImageObject) Ports(stringBounder font.StringBounder) *klimt.Ports {
	dimHeader := nameAndStereotypeDimension(o.name, o.stereo, stringBounder)
	if wp, ok := o.fields.(klimt.WithPorts); ok {
		return wp.Ports(stringBounder).TranslateY(dimHeader.Height())
	}
	return klimt.NewPorts()
}

// InnerPosition returns the position of a field.
func (o *EntityImageObject) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	dimTitle := nameAndStereotypeDimension(o.name, o.stereo, stringBounder)
	r := o.fields.InnerPosition(member, stringBounder)
	if r == nil {
		// Upstream's UTranslate.apply(null) throws NullPointerException.
		throwNPE()
	}
	moved := klimt.UTranslateDy(dimTitle.Height()).Apply(*r)
	return &moved
}

// EntityImageMap is the image of a map: its name and its entries.
type EntityImageMap struct {
	abstractEntityImage
	memo        klimt.TextBlockMemoized
	name        klimt.TextBlock
	stereo      klimt.TextBlock
	entries     klimt.TextBlock
	url         *klimt.URL
	roundCorner float64
}

var (
	_ klimt.IEntityImage = (*EntityImageMap)(nil)
	_ klimt.Stencil      = (*EntityImageMap)(nil)
	_ klimt.WithPorts    = (*EntityImageMap)(nil)
)

func newEntityImageMap(entity *abel.Entity, portionShower PortionShower) *EntityImageMap {
	m := &EntityImageMap{abstractEntityImage: newAbstractEntityImage(entity)}
	m.roundCorner = m.style().Value(klimt.PNameRoundCorner).AsDouble()
	fcHeader := m.styleHeader().FontConfiguration(m.skinParam.IHtmlColorSet(), nil)
	m.name = klimt.TextBlockUtilsWithMargin(entity.Display().Create(fcHeader, geom.HorizontalAlignmentCenter,
		m.skinParam), 2, 2)
	m.stereo = objectStereo(entity, portionShower, m.skinParam)
	fontConfiguration := m.style().FontConfiguration(m.skinParam.IHtmlColorSet(), nil)
	m.entries = entityBody(entity, m.skinParam, false, false, entity.Stereotype(), m.style(), fontConfiguration)
	m.url = entity.URL99()
	return m
}

func (m *EntityImageMap) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameMap).WithTOBECHANGED(m.entity.Stereotype()).MergedStyle(m.skinParam.CurrentStyleBuilder())
}

func (m *EntityImageMap) styleHeader() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameObjectDiagram,
		klimt.SNameMap, klimt.SNameHeader).WithTOBECHANGED(m.entity.Stereotype()).
		MergedStyle(m.skinParam.CurrentStyleBuilder())
}

// Ports returns the ports of the entries, below the name.
func (m *EntityImageMap) Ports(stringBounder font.StringBounder) *klimt.Ports {
	dimTitle := nameAndStereotypeDimension(m.name, m.stereo, stringBounder)
	return m.entries.(klimt.WithPorts).Ports(stringBounder).TranslateY(dimTitle.Height())
}

// CalculateDimension returns the size of the map.
func (m *EntityImageMap) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return m.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		dimTitle := nameAndStereotypeDimension(m.name, m.stereo, sb)
		dimFields := m.entries.CalculateDimension(sb)
		width := math.Max(dimFields.Width(), dimTitle.Width()+2*objectXMarginCircle)
		if minClassWidth := m.style().Value(klimt.PNameMinimumWidth).AsDouble(); width < minClassWidth {
			width = minClassWidth
		}
		height := m.methodOrFieldHeight(dimFields) + dimTitle.Height()
		return geom.NewXDimension2D(width, height)
	})
}

// DrawU draws the map.
func (m *EntityImageMap) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := m.CalculateDimension(stringBounder)
	dimTitle := nameAndStereotypeDimension(m.name, m.stereo, stringBounder)

	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
	rect := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(m.roundCorner)

	borderColor := m.entity.Colors().Color(color.TypeLine)
	backcolor := m.entity.Colors().Color(color.TypeBack)
	headerBackcolor := m.entity.Colors().Color(color.TypeHeader)
	style := m.style()
	set := m.skinParam.IHtmlColorSet()
	if borderColor == nil {
		borderColor = style.Value(klimt.PNameLineColor).AsColor(set)
	}
	if headerBackcolor == nil {
		if backcolor == nil {
			headerBackcolor = m.styleHeader().Value(klimt.PNameBackGroundColor).AsColor(set)
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
	if m.url != nil {
		ug.StartURL(m.url)
	}
	ug.StartGroup(entityGroup(m.entity))
	ug.Apply(stroke).Draw(rect)

	if headerBackcolor != nil && !backcolor.Equals(headerBackcolor) {
		rect2 := klimt.URectangleBuild(widthTotal, dimTitle.Height()).HalfRounded(m.roundCorner)
		ugHeader := ug.Apply(klimt.Bg(headerBackcolor))
		ugHeader.Apply(stroke).Draw(rect2)
	}
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(ug.StringBounder()))
	if m.stereo != nil {
		header.Add(m.stereo)
	}
	header.Add(m.name)
	header.DrawU(ug, dimTotal.Width(), dimTitle.Height())

	ug2 := klimt.UGraphicStencilCreate(ug, m, stroke)
	m.entries.(*abel.TextBlockMap).SetTotalWidth(dimTotal.Width())
	m.entries.DrawU(ug2.Apply(klimt.UTranslateDy(dimTitle.Height())))

	if m.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

func (m *EntityImageMap) methodOrFieldHeight(dim geom.XDimension2D) float64 {
	fieldsHeight := dim.Height()
	if fieldsHeight == 0 && m.entity.LeafType() != abel.LeafTypeMap {
		return objectMarginEmptyFieldsOrMethod
	}
	return fieldsHeight
}

// ShapeType returns RECTANGLE_HTML_FOR_PORTS.
func (m *EntityImageMap) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangleHTMLForPorts }

// StartingX implements klimt.Stencil.
func (m *EntityImageMap) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX implements klimt.Stencil.
func (m *EntityImageMap) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return m.CalculateDimension(stringBounder).Width()
}
