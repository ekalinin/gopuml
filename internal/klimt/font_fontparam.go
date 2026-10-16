// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/font/FontParam.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SkinParamUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/Rose.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// The defaults of upstream's FontParamConstant.
const (
	fontParamFamily = "SansSerif"
	fontParamColor  = "black"
)

// FontParam is a font skinparam of the legacy skin, with its default size,
// face, color and family.
type FontParam int

// fontParamInfo is the data of a FontParam.
type fontParamInfo struct {
	name          string
	defaultSize   int
	defaultFace   font.UFontFace
	defaultColor  string
	defaultFamily string
}

// Font parameters, in upstream's order.
const (
	FontParamTiming FontParam = iota
	FontParamActivity
	FontParamActivityDiamond
	FontParamArrow
	FontParamCircledCharacter
	FontParamObjectAttribute
	FontParamObject
	FontParamObjectStereotype
	FontParamClassAttribute
	FontParamClass
	FontParamClassStereotype
	FontParamComponent
	FontParamInterface
	FontParamInterfaceStereotype
	FontParamComponentStereotype
	FontParamNote
	FontParamPackage
	FontParamPackageStereotype
	FontParamActor
	FontParamArtifact
	FontParamCloud
	FontParamFolder
	FontParamFile
	FontParamFrame
	FontParamStorage
	FontParamBoundary
	FontParamControl
	FontParamEntity
	FontParamAgent
	FontParamRectangle
	FontParamLabel
	FontParamHexagon
	FontParamPerson
	FontParamArchimate
	FontParamCard
	FontParamNode
	FontParamDatabase
	FontParamQueue
	FontParamStack
	FontParamSequenceBox
	FontParamSequenceDivider
	FontParamSequenceReference
	FontParamSequenceDelay
	FontParamSequenceGroup
	FontParamSequenceGroupHeader
	FontParamParticipant
	FontParamParticipantStereotype
	FontParamState
	FontParamStateAttribute
	FontParamLegend
	FontParamTitle
	FontParamCaption
	FontParamSwimlaneTitle
	FontParamFooter
	FontParamHeader
	FontParamUsecase
	FontParamUsecaseStereotype
	FontParamArtifactStereotype
	FontParamCloudStereotype
	FontParamStorageStereotype
	FontParamBoundaryStereotype
	FontParamControlStereotype
	FontParamEntityStereotype
	FontParamAgentStereotype
	FontParamRectangleStereotype
	FontParamLabelStereotype
	FontParamPersonStereotype
	FontParamHexagonStereotype
	FontParamArchimateStereotype
	FontParamCardStereotype
	FontParamNodeStereotype
	FontParamFolderStereotype
	FontParamFileStereotype
	FontParamFrameStereotype
	FontParamDatabaseStereotype
	FontParamQueueStereotype
	FontParamStackStereotype
	FontParamActorStereotype
	FontParamSequenceStereotype
	FontParamPartition
	FontParamDesignedDomain
	FontParamDesignedDomainStereotype
	FontParamDomain
	FontParamDomainStereotype
	FontParamMachine
	FontParamMachineStereotype
	FontParamRequirement
	FontParamRequirementStereotype
	fontParamCount
)

var fontParams = [...]fontParamInfo{
	{"TIMING", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ACTIVITY", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ACTIVITY_DIAMOND", 11, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ARROW", 13, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CIRCLED_CHARACTER", 17, font.BoldFace(), fontParamColor, "Monospaced"},
	{"OBJECT_ATTRIBUTE", 10, font.NormalFace(), fontParamColor, fontParamFamily},
	{"OBJECT", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"OBJECT_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"CLASS_ATTRIBUTE", 10, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CLASS", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CLASS_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"COMPONENT", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"INTERFACE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"INTERFACE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"COMPONENT_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"NOTE", 13, font.NormalFace(), fontParamColor, fontParamFamily},
	{"PACKAGE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"PACKAGE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"ACTOR", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ARTIFACT", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CLOUD", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"FOLDER", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"FILE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"FRAME", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"STORAGE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"BOUNDARY", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CONTROL", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ENTITY", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"AGENT", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"RECTANGLE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"LABEL", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"HEXAGON", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"PERSON", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"ARCHIMATE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CARD", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"NODE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"DATABASE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"QUEUE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"STACK", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_BOX", 13, font.BoldFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_DIVIDER", 13, font.BoldFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_REFERENCE", 13, font.NormalFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_DELAY", 11, font.NormalFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_GROUP", 11, font.BoldFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_GROUP_HEADER", 13, font.BoldFace(), fontParamColor, fontParamFamily},
	{"PARTICIPANT", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"PARTICIPANT_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"STATE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"STATE_ATTRIBUTE", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"LEGEND", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"TITLE", 18, font.NormalFace(), fontParamColor, fontParamFamily},
	{"CAPTION", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"SWIMLANE_TITLE", 18, font.NormalFace(), fontParamColor, fontParamFamily},
	{"FOOTER", 10, font.NormalFace(), "#888888", fontParamFamily},
	{"HEADER", 10, font.NormalFace(), "#888888", fontParamFamily},
	{"USECASE", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"USECASE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"ARTIFACT_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"CLOUD_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"STORAGE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"BOUNDARY_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"CONTROL_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"ENTITY_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"AGENT_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"RECTANGLE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"LABEL_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"PERSON_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"HEXAGON_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"ARCHIMATE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"CARD_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"NODE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"FOLDER_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"FILE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"FRAME_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"DATABASE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"QUEUE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"STACK_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"ACTOR_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"SEQUENCE_STEREOTYPE", 14, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"PARTITION", 14, font.NormalFace(), fontParamColor, fontParamFamily},
	{"DESIGNED_DOMAIN", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"DESIGNED_DOMAIN_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"DOMAIN", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"DOMAIN_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"MACHINE", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"MACHINE_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
	{"REQUIREMENT", 12, font.NormalFace(), fontParamColor, fontParamFamily},
	{"REQUIREMENT_STEREOTYPE", 12, font.ItalicFace(), fontParamColor, fontParamFamily},
}

// String returns upstream's constant name, for example "CLASS_ATTRIBUTE".
func (p FontParam) String() string { return fontParams[p].name }

// DefaultSize returns the default size: 11 for activity arrows and class
// attributes.
func (p FontParam) DefaultSize(skinParam ISkinParam) int {
	if p == FontParamArrow && skinParam.DiagramType() == core.Activity {
		return 11
	}
	if p == FontParamClassAttribute {
		return 11
	}
	return fontParams[p].defaultSize
}

// DefaultFontFace returns the default face: bold for package titles.
func (p FontParam) DefaultFontFace(skinParam ISkinParam, inPackageTitle bool) font.UFontFace {
	if p == FontParamState {
		return fontParams[p].defaultFace
	}
	if inPackageTitle || p == FontParamPackage {
		return font.BoldFace()
	}
	return fontParams[p].defaultFace
}

// DefaultColor returns the default color.
func (p FontParam) DefaultColor() string { return fontParams[p].defaultColor }

// DefaultFamily returns the default family.
func (p FontParam) DefaultFamily() string { return fontParams[p].defaultFamily }

// FontConfiguration returns the font configuration of the parameter.
func (p FontParam) FontConfiguration(skinParam ISkinParam) *FontConfiguration {
	return FontConfigurationFromSkinParam(skinParam, p, nil)
}

// StyleDefinition returns the style signature of the parameter.
func (p FontParam) StyleDefinition(diagramType SName) StyleSignatureBasic {
	switch p {
	case FontParamFooter:
		return StyleSignatureBasicOf(SNameRoot, SNameDocument, SNameFooter)
	case FontParamHeader:
		return StyleSignatureBasicOf(SNameRoot, SNameDocument, SNameHeader)
	case FontParamTitle:
		return StyleSignatureBasicOf(SNameRoot, SNameDocument, SNameTitle)
	case FontParamClassAttribute:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameClassDiagram, SNameClass)
	case FontParamRectangle, FontParamNode:
		return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameComponentDiagram, SNameComponent)
	}
	return StyleSignatureBasicOf(SNameRoot, SNameElement, diagramType, SNameComponent)
}

// SkinParamUtilsFont is upstream's SkinParamUtils.getFont.
func SkinParamUtilsFont(skinParam ISkinParam, fontParam FontParam, stereo *Stereotype) font.UFont {
	return skinParam.Font(stereo, false, fontParam)
}

// SkinParamUtilsFontColor is upstream's SkinParamUtils.getFontColor.
func SkinParamUtilsFontColor(skinParam ISkinParam, fontParam FontParam, stereo *Stereotype) color.HColor {
	return skinParam.FontHtmlColor(stereo, fontParam)
}

// SkinParamUtilsColor is upstream's SkinParamUtils.getColor, through
// Rose.getHtmlColor: the first color parameter that is set, or the default
// of the first.
func SkinParamUtilsColor(skinParam ISkinParam, stereo *Stereotype, colorParams ...ColorParam) color.HColor {
	for _, param := range colorParams {
		if result := skinParam.HtmlColor(param, stereo, false); result != nil {
			return result
		}
	}
	if len(colorParams) == 0 {
		throw("java.lang.ArrayIndexOutOfBoundsException", "Index 0 out of bounds for length 0")
	}
	return colorParams[0].DefaultValue()
}
