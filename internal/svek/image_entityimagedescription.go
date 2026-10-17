// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageDescription.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/decoration/symbol"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// EntityImageDescription is the image of a description element (component,
// actor, use case, interface circle, node...) or of an empty package drawn
// with a symbol: the symbol drawn small around the name.
type EntityImageDescription struct {
	abstractEntityImage
	memo klimt.TextBlockMemoized

	shapeType klimt.ShapeType
	url       *klimt.URL
	asSmall   klimt.TextBlock
	name      klimt.TextBlock
	desc      klimt.TextBlock
	stereo    klimt.TextBlock

	hideText                  bool
	links                     []*abel.Link
	fixCircleLabelOverlapping bool
	bibliotekon               *Bibliotekon
	ctx                       *klimt.Fashion
	symbol                    uSymbol
}

var _ klimt.IEntityImage = (*EntityImageDescription)(nil)

func newEntityImageDescription(entity *abel.Entity, portionShower PortionShower, links []*abel.Link,
	bibliotekon *Bibliotekon) *EntityImageDescription {
	e := &EntityImageDescription{abstractEntityImage: newAbstractEntityImage(entity), bibliotekon: bibliotekon,
		links: links}
	skinParam := e.skinParam
	e.fixCircleLabelOverlapping = skinParam.FixCircleLabelOverlapping()
	e.symbol = e.uSymbol(entity)
	switch {
	case e.symbol == symbol.Folder || e.symbol == symbol.Package:
		e.shapeType = klimt.ShapeTypeFolder
	case e.symbol == symbol.Hexagon:
		e.shapeType = klimt.ShapeTypeHexagon
	case e.symbol == symbol.Usecase || e.symbol == symbol.UsecaseBusiness:
		e.shapeType = klimt.ShapeTypeOval
	case e.symbol == symbol.Interface:
		if skinParam.FixCircleLabelOverlapping() {
			e.shapeType = klimt.ShapeTypeRectangleWithCircleInside
		} else {
			e.shapeType = klimt.ShapeTypeRectangle
		}
	default:
		e.shapeType = klimt.ShapeTypeRectangle
	}
	e.hideText = e.symbol == symbol.Interface
	e.url = entity.URL99()

	colors := entity.Colors()
	styleName := e.styleName()
	var signatureTitle klimt.StyleSignatureBasic
	if e.symbol == symbol.ActorStickmanBusiness {
		signatureTitle = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, styleName, klimt.SNameActor,
			klimt.SNameBusiness, klimt.SNameTitle)
	} else {
		signatureTitle = klimt.StyleSignatureBasicOf3(klimt.SNameRoot, klimt.SNameElement, styleName,
			e.symbol.SNames(), klimt.SNameTitle)
	}
	stereotype := entity.Stereotype()
	styleTitle := signatureTitle.WithTOBECHANGED(stereotype).MergedStyle(entity.CurrentStyleBuilder()).
		EventuallyOverrideColors(colors)
	styleStereo := klimt.StyleSignatureBasicOf3(klimt.SNameRoot, klimt.SNameElement, styleName, e.symbol.SNames(),
		klimt.SNameStereotype).ForStereotypeItself(stereotype).MergedStyle(entity.CurrentStyleBuilder())
	style := klimt.StyleSignatureBasicOf3(klimt.SNameRoot, klimt.SNameElement, styleName, e.symbol.SNames()).
		WithTOBECHANGED(stereotype).MergedStyle(entity.CurrentStyleBuilder()).EventuallyOverrideColors(colors)

	set := skinParam.IHtmlColorSet()
	forecolor := styleTitle.Value(klimt.PNameLineColor).AsColor(set)
	backcolor := colors.Color(color.TypeBack)
	if backcolor == nil {
		backcolor = styleTitle.Value(klimt.PNameBackGroundColor).AsColor(set)
	}
	roundCorner := styleTitle.Value(klimt.PNameRoundCorner).AsDouble()
	diagonalCorner := styleTitle.Value(klimt.PNameDiagonalCorner).AsDouble()
	deltaShadow := styleTitle.Shadowing()
	stroke := styleTitle.StrokeOf(colors)
	fcTitle := styleTitle.FontConfiguration(set, nil)
	fc := style.FontConfiguration(set, nil)
	fcStereo := styleStereo.FontConfiguration(set, nil)
	defaultAlign := styleTitle.HorizontalAlignment()

	e.ctx = klimt.NewFashion(backcolor, forecolor).WithStroke(stroke).WithShadow(deltaShadow).
		WithCorner(roundCorner, diagonalCorner)

	codeDisplay := klimt.DisplayWithNewlines(skinParam.Pragma(), entity.Name())
	switch {
	case entity.Display().EqualsLike(codeDisplay) && e.symbol.SNames()[0] == klimt.SNamePackage,
		entity.Display().IsWhite():
		e.desc = klimt.TextBlockUtilsEmpty(style.Value(klimt.PNameMinimumWidth).AsDouble(), 0)
	case entity.Display().EqualsLike(codeDisplay):
		e.desc = klimt.BodyFactoryCreate3(entity.Display(), skinParam, defaultAlign, fcTitle, style.WrapWidth(),
			styleTitle)
	default:
		e.desc = klimt.BodyFactoryCreate3(entity.Display(), skinParam, defaultAlign, fc, style.WrapWidth(), style)
	}

	stereotypeLabels := portionShower.VisibleStereotypeLabels(entity)
	switch {
	case stereotype != nil && stereotype.Sprite(skinParam) != nil:
		e.stereo = stereotype.Sprite(skinParam)
	case stereotype == nil || len(stereotypeLabels) == 0:
		e.stereo = klimt.TextBlockUtilsEmpty(0, 0)
	default:
		if _, ok := stereotype.Label(klimt.GuillemetDoubleComparator()); !ok {
			e.stereo = klimt.TextBlockUtilsEmpty(0, 0)
		} else {
			e.stereo = klimt.TextBlockUtilsWithMargin(klimt.DisplayCreate(stereotypeLabels...).Create(fcStereo,
				geom.HorizontalAlignmentCenter, skinParam), 1, 0)
		}
	}

	e.name = abel.BodyFactoryCreate2(skinParam.DefaultTextAlignment(geom.HorizontalAlignmentCenter), codeDisplay,
		skinParam, stereotype, entity, styleTitle)

	stereotypeAlignment := styleStereo.HorizontalAlignment()
	if e.hideText {
		e.asSmall = e.symbol.AsSmall(klimt.TextBlockUtilsEmpty(0, 0), klimt.TextBlockUtilsEmpty(0, 0),
			klimt.TextBlockUtilsEmpty(0, 0), e.ctx, stereotypeAlignment)
	} else {
		e.asSmall = e.symbol.AsSmall(e.name, e.desc, e.stereo, e.ctx, stereotypeAlignment)
	}
	return e
}

func (e *EntityImageDescription) uSymbol(entity *abel.Entity) uSymbol {
	var result uSymbol
	if entity.USymbol() == abel.USymbolNull {
		result = symbol.ComponentStyleToUSymbol(e.skinParam.ComponentStyle())
	} else {
		result = usymbolOf(entity.USymbol())
	}
	if result == nil {
		throwNPE()
	}
	return result
}

// CalculateDimension returns the size of the symbol.
func (e *EntityImageDescription) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.memo.Dimension(stringBounder, e.asSmall.CalculateDimension)
}

// Shield returns the room of the texts around an interface circle.
func (e *EntityImageDescription) Shield(stringBounder font.StringBounder) klimt.Margins {
	if !e.hideText {
		return klimt.MarginsNone()
	}
	if e.isThereADoubleLink() {
		return klimt.MarginsNone()
	}
	if !e.fixCircleLabelOverlapping && e.hasSomeHorizontalLinkVisible() {
		return klimt.MarginsNone()
	}
	if e.hasSomeHorizontalLinkDoubleDecorated() {
		return klimt.MarginsNone()
	}
	dimStereo := e.stereo.CalculateDimension(stringBounder)
	dimDesc := e.desc.CalculateDimension(stringBounder)
	dimSmall := e.asSmall.CalculateDimension(stringBounder)
	x := math.Max(dimStereo.Width(), dimDesc.Width())
	suppX := x - dimSmall.Width()
	if suppX < 1 {
		suppX = 1
	}
	y := math.Max(math.Max(1, dimDesc.Height()), dimStereo.Height())
	return klimt.NewMargins(suppX/2, suppX/2, y, y)
}

func (e *EntityImageDescription) hasSomeHorizontalLinkVisible() bool {
	for _, link := range e.links {
		if link.Length() == 1 && link.Contains(e.entity) && !link.IsInvis() {
			return true
		}
	}
	return false
}

func (e *EntityImageDescription) isThereADoubleLink() bool {
	others := map[*abel.Entity]bool{}
	for _, link := range e.links {
		if link.Contains(e.entity) {
			other := link.Other(e.entity)
			if others[other] {
				return true
			}
			others[other] = true
		}
	}
	return false
}

func (e *EntityImageDescription) hasSomeHorizontalLinkDoubleDecorated() bool {
	for _, link := range e.links {
		if link.Length() == 1 && link.Contains(e.entity) && link.Type().IsDoubleDecorated() {
			return true
		}
	}
	return false
}

// DrawU draws the symbol, and the texts around an interface circle.
func (e *EntityImageDescription) DrawU(ug klimt.UGraphic) {
	ug.Draw(klimt.NewUComment("entity " + e.entity.Name()))
	ug.StartGroup(entityGroup(e.entity))
	if e.url != nil {
		ug.StartURL(e.url)
	}
	if e.shapeType == klimt.ShapeTypeHexagon {
		e.drawHexagon(e.ctx.Apply(ug))
	}
	e.asSmall.DrawU(ug)
	if e.hideText {
		const space = 8
		dimSmall := e.asSmall.CalculateDimension(ug.StringBounder())
		dimDesc := e.desc.CalculateDimension(ug.StringBounder())
		posx1 := (dimSmall.Width() - dimDesc.Width()) / 2
		ugDesc := ug.Apply(klimt.NewUTranslate(posx1, space+dimSmall.Height()))
		ugDesc = klimt.UGraphicStencilCreateDim(ugDesc, dimDesc)
		e.desc.DrawU(ugDesc)

		dimStereo := e.stereo.CalculateDimension(ug.StringBounder())
		posx2 := (dimSmall.Width() - dimStereo.Width()) / 2
		e.stereo.DrawU(ug.Apply(klimt.NewUTranslate(posx2, -space-dimStereo.Height())))
	}
	if e.url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

func (e *EntityImageDescription) drawHexagon(ug klimt.UGraphic) {
	if e.bibliotekon == nil {
		throwIllegalState("")
	}
	node := e.bibliotekon.node(e.entity)
	if hexagon := node.polygon; hexagon != nil {
		hexagon.SetDeltaShadow(e.ctx.DeltaShadow())
		ug.Draw(hexagon)
	}
}

// ShapeType returns the shape of the symbol in the graph.
func (e *EntityImageDescription) ShapeType() klimt.ShapeType { return e.shapeType }

// OverscanX returns how far the texts of an interface circle go out of it.
func (e *EntityImageDescription) OverscanX(stringBounder font.StringBounder) float64 {
	if e.hideText {
		dimSmall := e.asSmall.CalculateDimension(stringBounder)
		dimDesc := e.desc.CalculateDimension(stringBounder)
		dimStereo := e.stereo.CalculateDimension(stringBounder)
		posx1 := (dimSmall.Width() - dimDesc.Width()) / 2
		posx2 := (dimSmall.Width() - dimStereo.Width()) / 2
		return math.Max(math.Max(-posx1, -posx2), 0)
	}
	return 0
}

// MagneticBorder returns the border of the folder symbol, none for the
// others.
func (e *EntityImageDescription) MagneticBorder() klimt.MagneticBorder {
	if e.shapeType == klimt.ShapeTypeFolder {
		return e.asSmall.MagneticBorder()
	}
	return klimt.MagneticBorderNone{}
}
