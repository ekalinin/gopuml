// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageClassHeader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// EntityImageClassHeader is the header of a class: its spot, stereotype,
// name and generic parameters.
type EntityImageClassHeader struct {
	abstractEntityImage
	memo         klimt.TextBlockMemoized
	headerLayout headerLayout
}

func classStyleSignature(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameClassDiagram, klimt.SNameClass}, names...)...)
}

func newEntityImageClassHeader(entity *abel.Entity, portionShower PortionShower) *EntityImageClassHeader {
	h := &EntityImageClassHeader{abstractEntityImage: newAbstractEntityImage(entity)}
	skinParam := h.skinParam
	italic := entity.LeafType() == abel.LeafTypeAbstractClass || entity.LeafType() == abel.LeafTypeInterface

	stereotype := entity.Stereotype()
	displayGenericWithOldFashion := skinParam.DisplayGenericWithOldFashion()
	generic := entity.Generic()
	if displayGenericWithOldFashion {
		generic = abel.OptString{}
	}

	styleHeader := classStyleSignature(klimt.SNameHeader).WithTOBECHANGED(stereotype).
		With(entity.Stereostyles()).MergedStyle(skinParam.CurrentStyleBuilder())

	fontConfigurationName := klimt.FontConfigurationFromStyle(skinParam, styleHeader, entity.Colors())
	if italic {
		fontConfigurationName = fontConfigurationName.Italic()
	}
	display := entity.Display()
	if g := entity.Generic(); displayGenericWithOldFashion && g.OK {
		display = display.AddGeneric(g.S)
	}
	name := display.Create8(fontConfigurationName, geom.HorizontalAlignmentCenter, skinParam,
		klimt.CreoleModeFullButUnderscore, styleHeader.WrapWidth())
	if modifier, ok := entity.VisibilityModifier(); !ok {
		name = klimt.TextBlockUtilsWithMargin4(name, 3, 3, 0, 0)
	} else {
		r := rose.NewRose()
		back := roseColorOpt(r, skinParam, modifier)
		fore := r.HtmlColor(skinParam, modifier.Foreground())
		uBlock := klimt.TextBlockUtilsWithMargin4(
			modifier.UBlock(skinParam.ClassAttributeIconSize(), fore, back, false), 0, 0, 4, 0)
		name = klimt.TextBlockUtilsMergeLR(uBlock, name, geom.VerticalAlignmentCenter)
		name = klimt.TextBlockUtilsWithMargin4(name, 3, 3, 0, 0)
	}

	var stereo klimt.TextBlock
	stereotypeLabels := portionShower.VisibleStereotypeLabels(entity)
	if stereotype != nil {
		if _, ok := stereotype.Label(klimt.GuillemetDoubleComparator()); ok && len(stereotypeLabels) > 0 {
			stereo = klimt.TextBlockUtilsWithMargin(klimt.DisplayCreate(stereotypeLabels...).Create(
				klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamClassStereotype, stereotype),
				geom.HorizontalAlignmentCenter, skinParam), 1, 0)
		}
	}

	var genericBlock klimt.TextBlock
	if generic.OK {
		styleGeneric := classStyleSignature(klimt.SNameGeneric).WithTOBECHANGED(stereotype).
			With(entity.Stereostyles()).MergedStyle(skinParam.CurrentStyleBuilder())
		genericBlock = klimt.DisplayWithNewlines(skinParam.Pragma(), generic.S).Create(
			klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamClassStereotype, stereotype),
			geom.HorizontalAlignmentCenter, skinParam)
		genericBlock = klimt.TextBlockUtilsWithMargin(genericBlock, 1, 1)
		classBackground := styleGeneric.Value(klimt.PNameBackGroundColor).AsColor(skinParam.IHtmlColorSet())
		classBorder := styleGeneric.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
		genericBlock = klimt.NewTextBlockGeneric(genericBlock, classBackground, classBorder)
		genericBlock = klimt.TextBlockUtilsWithMargin(genericBlock, 1, 1)
	}

	var circledCharacter klimt.TextBlock
	if portionShower.ShowPortion(uml.PortionCircledCharacter, entity) {
		circledCharacter = klimt.TextBlockUtilsWithMargin4(h.circledCharacter(entity, skinParam), 4, 0, 5, 5)
	}
	h.headerLayout = newHeaderLayout(circledCharacter, stereo, name, genericBlock)
	return h
}

func (h *EntityImageClassHeader) circledCharacter(entity *abel.Entity, skinParam klimt.ISkinParam) klimt.TextBlock {
	stereotype := entity.Stereotype()
	if stereotype != nil {
		if sprite := stereotype.Sprite(skinParam); sprite != nil {
			return sprite
		}
	}
	f := klimt.SkinParamUtilsFont(h.skinParam, klimt.FontParamCircledCharacter, nil)
	leafType := entity.LeafType()
	style := spotStyleSignature(leafType).MergedStyle(skinParam.CurrentStyleBuilder())
	spotBorder := style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	spotBackColor := style.Value(klimt.PNameBackGroundColor).AsColor(skinParam.IHtmlColorSet())
	fontColor := style.Value(klimt.PNameFontColor).AsColor(skinParam.IHtmlColorSet())
	radius := float64(h.skinParam.CircledCharacterRadius())
	if stereotype != nil && stereotype.Character() != 0 {
		back := stereotype.HtmlColor()
		if back == nil {
			back = spotBackColor
		}
		return klimt.NewCircledCharacter(stereotype.Character(), radius, f, back, spotBorder, fontColor)
	}
	var circledChar rune
	if stereotype != nil {
		circledChar = h.skinParam.CircledCharacter(stereotype)
	}
	if circledChar == 0 {
		circledChar = circledCharOf(leafType)
	}
	return klimt.NewCircledCharacter(circledChar, radius, f, spotBackColor, spotBorder, fontColor)
}

func spotStyleSignature(leafType abel.LeafType) klimt.StyleSignatureBasic {
	var spot klimt.SName
	switch leafType {
	case abel.LeafTypeAnnotation:
		spot = klimt.SNameSpotAnnotation
	case abel.LeafTypeAbstractClass:
		spot = klimt.SNameSpotAbstractClass
	case abel.LeafTypeClass:
		spot = klimt.SNameSpotClass
	case abel.LeafTypeInterface:
		spot = klimt.SNameSpotInterface
	case abel.LeafTypeEnum:
		spot = klimt.SNameSpotEnum
	case abel.LeafTypeEntity:
		spot = klimt.SNameSpotEntity
	case abel.LeafTypeProtocol:
		spot = klimt.SNameSpotProtocol
	case abel.LeafTypeStruct:
		spot = klimt.SNameSpotStruct
	case abel.LeafTypeException:
		spot = klimt.SNameSpotException
	case abel.LeafTypeMetaclass:
		spot = klimt.SNameSpotMetaClass
	case abel.LeafTypeStereotype:
		spot = klimt.SNameSpotStereotype
	case abel.LeafTypeDataclass:
		spot = klimt.SNameSpotDataClass
	case abel.LeafTypeRecord:
		spot = klimt.SNameSpotRecord
	default:
		throwIllegalState("")
	}
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSpot, spot)
}

func circledCharOf(leafType abel.LeafType) rune {
	switch leafType {
	case abel.LeafTypeAnnotation:
		return '@'
	case abel.LeafTypeAbstractClass:
		return 'A'
	case abel.LeafTypeClass:
		return 'C'
	case abel.LeafTypeInterface:
		return 'I'
	case abel.LeafTypeEnum, abel.LeafTypeEntity:
		return 'E'
	case abel.LeafTypeProtocol:
		return 'P'
	case abel.LeafTypeStruct, abel.LeafTypeStereotype:
		return 'S'
	case abel.LeafTypeException:
		return 'X'
	case abel.LeafTypeMetaclass:
		return 'M'
	case abel.LeafTypeDataclass:
		return 'D'
	case abel.LeafTypeRecord:
		return 'R'
	}
	return '?'
}

// CalculateDimension returns the size of the header.
func (h *EntityImageClassHeader) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return h.memo.Dimension(stringBounder, h.headerLayout.dimension)
}

// DrawU panics, as upstream: the header is drawn with a width and a
// height.
func (h *EntityImageClassHeader) DrawU(klimt.UGraphic) { throwUnsupported("") }

func (h *EntityImageClassHeader) drawUSized(ug klimt.UGraphic, width, height float64) {
	h.headerLayout.drawU(ug, width, height)
}

// ShapeType returns RECTANGLE.
func (h *EntityImageClassHeader) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }
