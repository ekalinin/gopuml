// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/BoxedCharacter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/EntityImageDomain.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/EntityImageDesignedDomain.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/EntityImageMachine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/EntityImageRequirement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The Problem Frames elements of description diagrams (upstream's
// descdiagram package): domains, designed domains, machines and
// requirements.

// boxedCharacter is upstream's BoxedCharacter: a letter in a square, the
// type tag of a domain.
type boxedCharacter struct {
	klimt.TextBlockDefaults
	c                              rune
	font                           font.UFont
	innerCircle, circle, fontColor color.HColor
	radius                         float64
}

func (b *boxedCharacter) DrawU(ug klimt.UGraphic) {
	if b.circle != nil {
		ug = ug.Apply(b.circle)
	}
	if b.innerCircle == nil {
		ug = ug.Apply(klimt.Bg(color.None()))
	} else {
		ug = ug.Apply(klimt.Bg(b.innerCircle))
	}
	ug.Draw(klimt.URectangleBuild(2*b.radius, 2*b.radius))
	ug = ug.Apply(b.fontColor)
	ug = ug.Apply(klimt.NewUTranslate(b.radius, b.radius))
	ug.Draw(klimt.NewUCenteredCharacter(b.c, b.font))
}

func (b *boxedCharacter) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(2*b.radius, 2*b.radius)
}

// domainStereo is the stereotype block of the Problem Frames images, nil
// without a stereotype label.
func domainStereo(skinParam klimt.ISkinParam, stereotype *klimt.Stereotype, fontParam klimt.FontParam) klimt.TextBlock {
	if stereotype == nil {
		return nil
	}
	if _, ok := stereotype.Label(klimt.GuillemetDoubleComparator()); !ok {
		return nil
	}
	return klimt.DisplayCreate(stereotype.Labels(skinParam.Guillemet())...).Create(
		klimt.FontConfigurationFromSkinParam(skinParam, fontParam, stereotype), geom.HorizontalAlignmentCenter,
		skinParam)
}

// domainName is the display of a Problem Frames element with a margin of 2.
func domainName(entity *abel.Entity, skinParam klimt.ISkinParam, fc *klimt.FontConfiguration) klimt.TextBlock {
	return klimt.TextBlockUtilsWithMargin(entity.Display().Create(fc, geom.HorizontalAlignmentCenter, skinParam),
		2, 2)
}

// nameAndStereoDimension is getNameAndSteretypeDimension: the name above
// the stereotype.
func nameAndStereoDimension(name, stereo klimt.TextBlock, stringBounder font.StringBounder) geom.XDimension2D {
	nameDim := name.CalculateDimension(stringBounder)
	stereoDim := geom.NewXDimension2D(0, 0)
	if stereo != nil {
		stereoDim = stereo.CalculateDimension(stringBounder)
	}
	return geom.NewXDimension2D(max(nameDim.Width(), stereoDim.Width()), nameDim.Height()+stereoDim.Height())
}

func domainStroke(skinParam klimt.ISkinParam, param klimt.LineParam, stereotype *klimt.Stereotype,
	defaultStroke klimt.UStroke) klimt.UStroke {
	if stroke := skinParam.Thickness(param, stereotype); stroke != nil {
		return *stroke
	}
	return defaultStroke
}

// domainBackColor is the back color of the entity, or the color of the
// skin parameter.
func domainBackColor(entity *abel.Entity, skinParam klimt.ISkinParam, param klimt.ColorParam) color.HColor {
	if c := entity.Colors().Color(color.TypeBack); c != nil {
		return c
	}
	return klimt.SkinParamUtilsColor(skinParam, entity.Stereotype(), param)
}

// EntityImageDomain is upstream's EntityImageDomain: a box with the name
// and the type letter in the bottom right corner ("X" lexical, "C" causal,
// "B" biddable, "P" otherwise).
type EntityImageDomain struct {
	abstractEntityImage
	name, tag, stereoBlock klimt.TextBlock
	url                    *klimt.URL
}

var _ klimt.IEntityImage = (*EntityImageDomain)(nil)

func newEntityImageDomain(entity *abel.Entity, typeLetter rune) *EntityImageDomain {
	e := &EntityImageDomain{abstractEntityImage: newAbstractEntityImage(entity)}
	stereotype := entity.Stereotype()
	fc := klimt.FontConfigurationFromSkinParam(e.skinParam, klimt.FontParamDesignedDomain, stereotype)
	e.name = domainName(entity, e.skinParam, fc)
	e.stereoBlock = domainStereo(e.skinParam, stereotype, klimt.FontParamDesignedDomainStereotype)
	var innerCircle color.HColor
	if stereotype != nil {
		innerCircle = stereotype.HtmlColor()
	} else {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.stereo.Stereotype.getHtmlColor()" because ` +
			`"stereotype" is null`)
	}
	e.tag = &boxedCharacter{c: typeLetter, radius: 8, font: font.ByDefault(8), innerCircle: innerCircle,
		fontColor: fc.Color()}
	e.url = entity.URL99()
	return e
}

func (e *EntityImageDomain) titleDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return nameAndStereoDimension(e.name, e.stereoBlock, stringBounder)
}

func (e *EntityImageDomain) tagDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.tag.CalculateDimension(stringBounder)
}

// CalculateDimension returns the size of the name, with the tag below.
func (e *EntityImageDomain) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimTitle := e.titleDimension(stringBounder)
	return geom.NewXDimension2D(dimTitle.Width(), dimTitle.Height()+e.tagDimension(stringBounder).Height())
}

// DrawU draws the box, the name and the tag.
func (e *EntityImageDomain) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimTitle := e.titleDimension(stringBounder)
	dimTag := e.tagDimension(stringBounder)
	rect := klimt.URectangleBuild(dimTotal.Width(), dimTotal.Height())
	ug = ug.Apply(klimt.SkinParamUtilsColor(e.skinParam, e.stereo(), klimt.ColorParamDomainBorder))
	ug = ug.Apply(klimt.Bg(domainBackColor(e.entity, e.skinParam, klimt.ColorParamDomainBackground)))
	if e.url != nil {
		ug.StartURL(e.url)
	}
	stroke := domainStroke(e.skinParam, klimt.LineParamDomainBorder, e.stereo(), klimt.UStrokeWithThickness(1.5))
	ug.Apply(stroke).Draw(rect)
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(stringBounder))
	header.Add(e.name)
	header.DrawU(ug, dimTotal.Width(), dimTitle.Height())
	footer := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(stringBounder))
	footer.Add(e.tag)
	footer.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width()-dimTag.Width(), dimTitle.Height())),
		dimTag.Width(), dimTag.Height())
	if e.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns RECTANGLE.
func (e *EntityImageDomain) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// EntityImageDesignedDomain is upstream's EntityImageDesignedDomain: a box
// with a band on its left side.
type EntityImageDesignedDomain struct {
	abstractEntityImage
	name, stereoBlock klimt.TextBlock
	url               *klimt.URL
}

var _ klimt.IEntityImage = (*EntityImageDesignedDomain)(nil)

func newEntityImageDesignedDomain(entity *abel.Entity) *EntityImageDesignedDomain {
	e := &EntityImageDesignedDomain{abstractEntityImage: newAbstractEntityImage(entity)}
	stereotype := entity.Stereotype()
	e.name = domainName(entity, e.skinParam,
		klimt.FontConfigurationFromSkinParam(e.skinParam, klimt.FontParamDesignedDomain, stereotype))
	e.stereoBlock = domainStereo(e.skinParam, stereotype, klimt.FontParamDesignedDomainStereotype)
	e.url = entity.URL99()
	return e
}

// CalculateDimension returns the size of the name, with the band.
func (e *EntityImageDesignedDomain) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimTitle := nameAndStereoDimension(e.name, e.stereoBlock, stringBounder)
	return geom.NewXDimension2D(dimTitle.Width()+5, dimTitle.Height())
}

// DrawU draws the box, the band and the name.
func (e *EntityImageDesignedDomain) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimTitle := nameAndStereoDimension(e.name, e.stereoBlock, stringBounder)
	rect := klimt.URectangleBuild(dimTotal.Width(), dimTotal.Height())
	rect2 := klimt.URectangleBuild(4, dimTotal.Height())
	ug = ug.Apply(klimt.SkinParamUtilsColor(e.skinParam, e.stereo(), klimt.ColorParamDesignedBorder))
	ug = ug.Apply(klimt.Bg(domainBackColor(e.entity, e.skinParam, klimt.ColorParamDesignedBackground)))
	if e.url != nil {
		ug.StartURL(e.url)
	}
	stroke := domainStroke(e.skinParam, klimt.LineParamDesignedDomainBorder, e.stereo(),
		klimt.UStrokeWithThickness(1.5))
	ug.Apply(stroke).Draw(rect)
	ug.Apply(stroke).Draw(rect2)
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(stringBounder))
	header.Add(e.name)
	header.DrawU(ug.Apply(klimt.UTranslateDx(4)), dimTotal.Width(), dimTitle.Height())
	if e.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns RECTANGLE.
func (e *EntityImageDesignedDomain) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// EntityImageMachine is upstream's EntityImageMachine: a box with a band
// set in from its left side.
type EntityImageMachine struct {
	abstractEntityImage
	name, stereoBlock klimt.TextBlock
	url               *klimt.URL
}

var _ klimt.IEntityImage = (*EntityImageMachine)(nil)

func newEntityImageMachine(entity *abel.Entity) *EntityImageMachine {
	e := &EntityImageMachine{abstractEntityImage: newAbstractEntityImage(entity)}
	stereotype := entity.Stereotype()
	e.name = domainName(entity, e.skinParam,
		klimt.FontConfigurationFromSkinParam(e.skinParam, klimt.FontParamMachine, stereotype))
	e.stereoBlock = domainStereo(e.skinParam, stereotype, klimt.FontParamMachineStereotype)
	e.url = entity.URL99()
	return e
}

// CalculateDimension returns the size of the name, with the band.
func (e *EntityImageMachine) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimTitle := nameAndStereoDimension(e.name, e.stereoBlock, stringBounder)
	return geom.NewXDimension2D(dimTitle.Width()+10, dimTitle.Height())
}

// DrawU draws the box, the band and the name.
func (e *EntityImageMachine) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := e.CalculateDimension(stringBounder)
	dimTitle := nameAndStereoDimension(e.name, e.stereoBlock, stringBounder)
	rect := klimt.URectangleBuild(dimTotal.Width(), dimTotal.Height())
	rect2 := klimt.URectangleBuild(4, dimTotal.Height())
	ug = ug.Apply(klimt.SkinParamUtilsColor(e.skinParam, e.stereo(), klimt.ColorParamMachineBorder))
	ug = ug.Apply(klimt.Bg(domainBackColor(e.entity, e.skinParam, klimt.ColorParamMachineBackground)))
	if e.url != nil {
		ug.StartURL(e.url)
	}
	stroke := domainStroke(e.skinParam, klimt.LineParamMachineBorder, e.stereo(), klimt.UStrokeWithThickness(1.5))
	ug.Apply(stroke).Draw(rect)
	ug.Apply(klimt.UTranslateDx(4)).Apply(stroke).Draw(rect2)
	header := klimt.NewULayoutGroup(klimt.NewPlacementStrategyY1Y2(stringBounder))
	header.Add(e.name)
	header.DrawU(ug.Apply(klimt.UTranslateDx(5)), dimTotal.Width(), dimTitle.Height())
	if e.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns RECTANGLE.
func (e *EntityImageMachine) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// EntityImageRequirement is upstream's EntityImageRequirement: the text in
// a dashed ellipse.
type EntityImageRequirement struct {
	abstractEntityImage
	desc klimt.TextBlock
	url  *klimt.URL
}

var _ klimt.IEntityImage = (*EntityImageRequirement)(nil)

func newEntityImageRequirement(entity *abel.Entity) *EntityImageRequirement {
	e := &EntityImageRequirement{abstractEntityImage: newAbstractEntityImage(entity)}
	stereotype := entity.Stereotype()
	tmp := abel.BodyFactoryCreate2(e.skinParam.DefaultTextAlignment(geom.HorizontalAlignmentCenter),
		entity.Display(), e.skinParam, stereotype, entity, e.style())
	e.desc = tmp
	if stereotype != nil {
		if _, ok := stereotype.Label(klimt.GuillemetDoubleComparator()); ok {
			label, hasLabel := stereotype.Label(e.skinParam.Guillemet())
			stereo := klimt.DisplayWithNewlinesString(e.skinParam.Pragma(), label, hasLabel).Create(
				klimt.FontConfigurationFromSkinParam(e.skinParam, klimt.FontParamRequirementStereotype,
					stereotype), geom.HorizontalAlignmentCenter, e.skinParam)
			e.desc = klimt.TextBlockUtilsMergeTB(stereo, tmp, geom.HorizontalAlignmentCenter)
		}
	}
	e.url = entity.URL99()
	return e
}

func (e *EntityImageRequirement) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameComponentDiagram,
		klimt.SNameRequirement).WithTOBECHANGED(e.entity.Stereotype()).MergedStyle(e.skinParam.CurrentStyleBuilder())
}

// CalculateDimension returns the size of the ellipse.
func (e *EntityImageRequirement) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return klimt.NewTextBlockInContainingEllipse(e.desc, stringBounder).CalculateDimension(stringBounder)
}

// DrawU draws the text in its ellipse, the separators of the text stopping
// at the ellipse.
func (e *EntityImageRequirement) DrawU(ug klimt.UGraphic) {
	ellipse := klimt.NewTextBlockInContainingEllipse(e.desc, ug.StringBounder())
	if e.url != nil {
		ug.StartURL(e.url)
	}
	ug = ug.Apply(domainStroke(e.skinParam, klimt.LineParamRequirementBorder, e.stereo(),
		klimt.NewUStroke(7, 7, 1.5)))
	linecolor := e.entity.Colors().Color(color.TypeLine)
	if linecolor == nil {
		linecolor = klimt.SkinParamUtilsColor(e.skinParam, e.stereo(), klimt.ColorParamRequirementBorder)
	}
	ug = ug.Apply(linecolor)
	ug = ug.Apply(klimt.Bg(domainBackColor(e.entity, e.skinParam, klimt.ColorParamRequirementBackground)))
	ellipse.DrawU(newRequirementUGraphicEllipse(ug, 0, 0, ellipse.UEllipse()))
	if e.url != nil {
		ug.CloseURL()
	}
}

// ShapeType returns OVAL.
func (e *EntityImageRequirement) ShapeType() klimt.ShapeType { return klimt.ShapeTypeOval }

// requirementUGraphicEllipse is EntityImageRequirement.MyUGraphicEllipse:
// the separators of the text stop at the ellipse.
type requirementUGraphicEllipse struct {
	klimt.AbstractUGraphicHorizontalLine
	startingX          float64
	yTheoricalPosition float64
	ellipse            *klimt.UEllipse
}

func newRequirementUGraphicEllipse(ug klimt.UGraphic, startingX, yTheoricalPosition float64,
	ellipse *klimt.UEllipse) *requirementUGraphicEllipse {
	g := &requirementUGraphicEllipse{AbstractUGraphicHorizontalLine: klimt.NewAbstractUGraphicHorizontalLine(ug),
		startingX: startingX, yTheoricalPosition: yTheoricalPosition, ellipse: ellipse}
	g.SetSelf(g)
	return g
}

// CopyHorizontalLine returns the graphic over ug.
func (g *requirementUGraphicEllipse) CopyHorizontalLine(ug klimt.UGraphic) klimt.HorizontalLineGraphic {
	return newRequirementUGraphicEllipse(ug, g.startingX, g.yTheoricalPosition, g.ellipse)
}

func (g *requirementUGraphicEllipse) normalized(y float64) float64 {
	if y < g.yTheoricalPosition {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	y = y - g.yTheoricalPosition
	if y > g.ellipse.Height() {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return y
}

// requirementEllipseStencil is the anonymous stencil of upstream's
// getStencil2.
type requirementEllipseStencil struct {
	g  *requirementUGraphicEllipse
	dy float64
}

func (s requirementEllipseStencil) StartingX(_ font.StringBounder, y float64) float64 {
	return s.g.startingX + s.g.ellipse.StartingX(s.g.normalized(y+s.dy))
}

func (s requirementEllipseStencil) EndingX(_ font.StringBounder, y float64) float64 {
	return s.g.startingX + s.g.ellipse.EndingX(s.g.normalized(y+s.dy))
}

// DrawHline draws the separator inside the ellipse with a 1.5 stroke.
func (g *requirementUGraphicEllipse) DrawHline(ug klimt.UGraphic, line *klimt.UHorizontalLine,
	translate klimt.UTranslate) {
	line.DrawLineInternal(ug.Apply(translate), requirementEllipseStencil{g: g, dy: translate.Dy()}, 0,
		klimt.UStrokeWithThickness(1.5))
}
