// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageClass.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// EntityImageClass is the image of a class-like entity: the header and the
// member compartments in a (possibly rounded) rectangle.
type EntityImageClass struct {
	abstractEntityImage
	memo        klimt.TextBlockMemoized
	body        klimt.TextBlock
	header      *EntityImageClassHeader
	url         *klimt.URL
	roundCorner float64
}

var (
	_ klimt.IEntityImage = (*EntityImageClass)(nil)
	_ klimt.Stencil      = (*EntityImageClass)(nil)
	_ klimt.WithPorts    = (*EntityImageClass)(nil)
)

func newEntityImageClass(entity *abel.Entity, portionShower PortionShower) *EntityImageClass {
	c := &EntityImageClass{abstractEntityImage: newAbstractEntityImage(entity)}
	c.roundCorner = c.style().Value(klimt.PNameRoundCorner).AsDouble()
	showMethods := portionShower.ShowPortion(uml.PortionMethod, entity)
	showFields := portionShower.ShowPortion(uml.PortionField, entity)
	c.body = entityBody(entity, c.skinParam, showMethods, showFields, entity.Stereotype(), c.style(), nil)
	c.header = newEntityImageClassHeader(entity, portionShower)
	c.url = entity.URL99()
	return c
}

// CalculateDimension returns the size of the class.
func (c *EntityImageClass) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return c.memo.Dimension(stringBounder, c.calculateDimensionSlow)
}

func (c *EntityImageClass) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	dimHeader := c.header.CalculateDimension(stringBounder)
	dimBody := geom.NewXDimension2D(0, 0)
	if c.body != nil {
		dimBody = c.body.CalculateDimension(stringBounder)
	}
	width := math.Max(dimBody.Width(), dimHeader.Width())
	if minClassWidth := c.style().Value(klimt.PNameMinimumWidth).AsDouble(); width < minClassWidth {
		width = minClassWidth
	}
	if paramSameClassWidth := c.skinParam.ParamSameClassWidth(); width < paramSameClassWidth {
		width = paramSameClassWidth
	}
	height := dimBody.Height() + dimHeader.Height()
	return geom.NewXDimension2D(math.Max(width, c.kalWidth()*1.3), height)
}

func (c *EntityImageClass) kalWidth() float64 {
	widthUp, widthDown := 0.0, 0.0
	for _, kal := range c.entity.Kals(klimt.DirectionUp) {
		widthUp += kal.(*Kal).Dimension().Width()
	}
	for _, kal := range c.entity.Kals(klimt.DirectionDown) {
		widthDown += kal.(*Kal).Dimension().Width()
	}
	return math.Max(widthUp, widthDown)
}

// InnerPosition returns the position of a member in the class.
func (c *EntityImageClass) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	result := c.body.InnerPosition(member, stringBounder)
	if result == nil {
		return nil
	}
	dimHeader := c.header.CalculateDimension(stringBounder)
	r := klimt.UTranslateDy(dimHeader.Height()).Apply(*result)
	return &r
}

// DrawU draws the class.
func (c *EntityImageClass) DrawU(ug klimt.UGraphic) {
	ug.Draw(klimt.NewUComment("class " + c.entity.Name()))
	if c.url != nil {
		ug.StartURL(c.url)
	}
	ug.StartGroup(entityGroup(c.entity))
	c.drawInternal(ug)
	ug.CloseGroup()
	if c.url != nil {
		ug.CloseURL()
	}
}

func (c *EntityImageClass) style() *klimt.Style {
	return classStyleSignature().WithTOBECHANGED(c.entity.Stereotype()).With(c.entity.Stereostyles()).
		MergedStyle(c.entity.CurrentStyleBuilder())
}

func (c *EntityImageClass) styleHeader() *klimt.Style {
	return classStyleSignature(klimt.SNameHeader).WithTOBECHANGED(c.entity.Stereotype()).
		With(c.entity.Stereostyles()).MergedStyle(c.entity.CurrentStyleBuilder())
}

func (c *EntityImageClass) drawInternal(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := c.CalculateDimension(stringBounder)
	dimHeader := c.header.CalculateDimension(stringBounder)

	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
	rect := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(c.roundCorner).WithCommentAndCodeLine("", "")

	style := c.style()
	set := c.skinParam.IHtmlColorSet()
	borderColor := c.entity.Colors().Color(color.TypeLine)
	headerBackcolor := c.entity.Colors().Color(color.TypeHeader)
	backcolor := c.entity.Colors().Color(color.TypeBack)

	shadow := style.Shadowing()
	if borderColor == nil {
		borderColor = style.Value(klimt.PNameLineColor).AsColor(set)
	}
	if headerBackcolor == nil {
		if backcolor == nil {
			headerBackcolor = c.styleHeader().Value(klimt.PNameBackGroundColor).AsColor(set)
		} else {
			headerBackcolor = backcolor
		}
	}
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(set)
	}
	rect.SetDeltaShadow(shadow)

	ug = ug.Apply(borderColor)
	ug = ug.Apply(klimt.Bg(backcolor))

	stroke := style.StrokeOf(c.entity.Colors())

	ugHeader := ug
	switch {
	case c.roundCorner == 0 && headerBackcolor != nil && !backcolor.Equals(headerBackcolor):
		ug.Apply(stroke).Draw(rect)
		rect2 := klimt.URectangleBuild(widthTotal, dimHeader.Height())
		rect2.SetDeltaShadow(0)
		ugHeader = ugHeader.Apply(klimt.Bg(headerBackcolor))
		ugHeader.Apply(stroke).Draw(rect2)
	case c.roundCorner != 0 && headerBackcolor != nil && !backcolor.Equals(headerBackcolor):
		ug.Apply(stroke).Draw(rect)
		rect2 := klimt.URectangleBuild(widthTotal, dimHeader.Height()).Rounded(c.roundCorner)
		rect3 := klimt.URectangleBuild(widthTotal, c.roundCorner/2)
		rect2.SetDeltaShadow(0)
		rect3.SetDeltaShadow(0)
		ugHeader = ugHeader.Apply(klimt.Bg(headerBackcolor)).Apply(headerBackcolor)
		ugHeader.Apply(stroke).Draw(rect2)
		ugHeader.Apply(stroke).Apply(klimt.UTranslateDy(dimHeader.Height() - rect3.Height())).Draw(rect3)
		rect.SetDeltaShadow(0)
		ug.Apply(stroke).Apply(klimt.Bg(color.None())).Draw(rect)
	default:
		ug.Apply(stroke).Draw(rect)
	}
	c.header.drawUSized(ugHeader, dimTotal.Width(), dimHeader.Height())

	if c.body != nil {
		ug2 := klimt.UGraphicStencilCreate(ug, c, stroke)
		translate := klimt.UTranslateDy(dimHeader.Height())
		c.body.DrawU(ug2.Apply(translate))
	}
}

// Ports returns the ports of the members, below the header.
func (c *EntityImageClass) Ports(stringBounder font.StringBounder) *klimt.Ports {
	dimHeader := c.header.CalculateDimension(stringBounder)
	if wp, ok := c.body.(klimt.WithPorts); ok {
		return wp.Ports(stringBounder).TranslateY(dimHeader.Height())
	}
	return klimt.NewPorts()
}

// ShapeType returns RECTANGLE_HTML_FOR_PORTS when members are used as
// ports, else RECTANGLE.
func (c *EntityImageClass) ShapeType() klimt.ShapeType {
	if len(c.entity.PortShortNames()) > 0 {
		return klimt.ShapeTypeRectangleHTMLForPorts
	}
	return klimt.ShapeTypeRectangle
}

// Shield returns the margins of the entity.
func (c *EntityImageClass) Shield(font.StringBounder) klimt.Margins { return c.entity.Margins() }

// StartingX implements klimt.Stencil.
func (c *EntityImageClass) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX implements klimt.Stencil.
func (c *EntityImageClass) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return c.CalculateDimension(stringBounder).Width()
}
