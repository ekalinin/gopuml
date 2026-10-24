// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ColorParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/LineParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/AlignmentParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/CornerParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/PaddingParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/SplitParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ComponentStyle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ActorStyle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ArrowDirection.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ColorParam is a color skinparam of the legacy (pre-style) skin.
type ColorParam int

// Color parameters, in upstream's order.
const (
	ColorParamBackground ColorParam = iota
	ColorParamHyperlink
	ColorParamActivityBackground
	ColorParamActivityBorder
	ColorParamClassBackground
	ColorParamClassBorder
	ColorParamArrowHead
	ColorParamStateBorder
	ColorParamNoteBackground
	ColorParamNoteBorder
	ColorParamDiagramBorder
	ColorParamActorBackground
	ColorParamActorBorder
	ColorParamSequenceGroupBodyBackground
	ColorParamSequenceReferenceHeaderBackground
	ColorParamSequenceReferenceBackground
	ColorParamSequenceLifeLineBorder
	ColorParamSequenceNewpageSeparator
	ColorParamSequenceBoxBorder
	ColorParamIconPrivate
	ColorParamIconPrivateBackground
	ColorParamIconPackage
	ColorParamIconPackageBackground
	ColorParamIconProtected
	ColorParamIconProtectedBackground
	ColorParamIconPublic
	ColorParamIconPublicBackground
	ColorParamIconIEMandatory
	ColorParamArrowLollipop
	ColorParamMachineBackground
	ColorParamMachineBorder
	ColorParamRequirementBackground
	ColorParamRequirementBorder
	ColorParamDesignedBackground
	ColorParamDesignedBorder
	ColorParamDomainBackground
	ColorParamDomainBorder
	ColorParamLexicalBackground
	ColorParamLexicalBorder
	ColorParamBiddableBackground
	ColorParamBiddableBorder
	colorParamCount
)

// colorParamInfo is the data of a ColorParam: its name, default value,
// whether it is a background and its color type (-1 for none).
type colorParamInfo struct {
	name         string
	defaultValue color.HColor
	isBackground bool
	colorType    color.ColorType
}

const noColorType color.ColorType = -1

var colorParams = [...]colorParamInfo{
	{"background", color.White, true, color.TypeBack},
	{"hyperlink", color.Blue, false, noColorType},
	{"activityBackground", color.MyYellow, true, color.TypeBack},
	{"activityBorder", color.MyRed, false, color.TypeLine},
	{"classBackground", color.MyYellow, true, color.TypeBack},
	{"classBorder", color.MyRed, false, color.TypeLine},
	{"arrowHead", color.MyRed, false, noColorType},
	{"stateBorder", color.MyRed, false, color.TypeLine},
	{"noteBackground", color.ColFBFB77, true, color.TypeBack},
	{"noteBorder", color.MyRed, false, color.TypeLine},
	{"diagramBorder", nil, false, color.TypeLine},
	{"actorBackground", color.MyYellow, true, color.TypeBack},
	{"actorBorder", color.MyRed, false, color.TypeLine},
	{"sequenceGroupBodyBackground", color.Red, true, color.TypeBack},
	{"sequenceReferenceHeaderBackground", color.ColEEEEEE, true, color.TypeBack},
	{"sequenceReferenceBackground", color.White, true, color.TypeBack},
	{"sequenceLifeLineBorder", color.MyRed, false, color.TypeLine},
	{"sequenceNewpageSeparator", color.Black, false, color.TypeLine},
	{"sequenceBoxBorder", color.MyRed, false, color.TypeLine},
	{"iconPrivate", color.ColC82930, false, noColorType},
	{"iconPrivateBackground", color.ColF24D5C, false, noColorType},
	{"iconPackage", color.Col1963A0, false, noColorType},
	{"iconPackageBackground", color.Col4177AF, false, noColorType},
	{"iconProtected", color.ColB38D22, false, noColorType},
	{"iconProtectedBackground", color.ColFFFF44, false, noColorType},
	{"iconPublic", color.Col038048, false, noColorType},
	{"iconPublicBackground", color.Col84BE84, false, noColorType},
	{"iconIEMandatory", color.Black, false, noColorType},
	{"arrowLollipop", color.White, false, noColorType},
	{"machineBackground", color.White, false, noColorType},
	{"machineBorder", color.Black, false, color.TypeLine},
	{"requirementBackground", color.White, false, noColorType},
	{"requirementBorder", color.Black, false, color.TypeLine},
	{"designedBackground", color.White, false, noColorType},
	{"designedBorder", color.Black, false, color.TypeLine},
	{"domainBackground", color.White, false, noColorType},
	{"domainBorder", color.Black, false, color.TypeLine},
	{"lexicalBackground", color.White, false, noColorType},
	{"lexicalBorder", color.Black, false, color.TypeLine},
	{"biddableBackground", color.White, false, noColorType},
	{"biddableBorder", color.Black, false, color.TypeLine},
}

// String returns upstream's constant name, for example "noteBackground".
func (p ColorParam) String() string { return colorParams[p].name }

// IsBackground reports whether the parameter is a background color.
func (p ColorParam) IsBackground() bool { return colorParams[p].isBackground }

// DefaultValue returns the default color, nil for none.
func (p ColorParam) DefaultValue() color.HColor { return colorParams[p].defaultValue }

// ColorType returns the color type; ok is false when there is none (null
// upstream).
func (p ColorParam) ColorType() (color.ColorType, bool) {
	t := colorParams[p].colorType
	return t, t != noColorType
}

// LineParam is a line thickness skinparam of the legacy skin.
type LineParam int

// Line parameters, in upstream's order.
const (
	LineParamSequenceActorBorder LineParam = iota
	LineParamSequenceDividerBorder
	LineParamSequenceLifeLineBorder
	LineParamSequenceParticipantBorder
	LineParamNoteBorder
	LineParamSequenceGroupBorder
	LineParamSequenceReferenceBorder
	LineParamLegendBorder
	LineParamSequenceArrow
	LineParamArrow
	LineParamClassBorder
	LineParamObjectBorder
	LineParamUsecaseBorder
	LineParamPartitionBorder
	LineParamPackageBorder
	LineParamSwimlaneBorder
	LineParamActivityBorder
	LineParamTitleBorder
	LineParamDiagramBorder
	LineParamRectangleBorder
	LineParamHexagonBorder
	LineParamPersonBorder
	LineParamArchimateBorder
	LineParamComponentBorder
	LineParamCardBorder
	LineParamQueueBorder
	LineParamAgentBorder
	LineParamDomainBorder
	LineParamDesignedDomainBorder
	LineParamMachineBorder
	LineParamRequirementBorder
	lineParamCount
)

var lineParamNames = []string{"sequenceActorBorder", "sequenceDividerBorder", "sequenceLifeLineBorder",
	"sequenceParticipantBorder", "noteBorder", "sequenceGroupBorder", "sequenceReferenceBorder", "legendBorder",
	"sequenceArrow", "arrow", "classBorder", "objectBorder", "usecaseBorder", "partitionBorder", "packageBorder",
	"swimlaneBorder", "activityBorder", "titleBorder", "diagramBorder", "rectangleBorder", "hexagonBorder",
	"personBorder", "archimateBorder", "componentBorder", "cardBorder", "queueBorder", "agentBorder",
	"domainBorder", "designedDomainBorder", "machineBorder", "requirementBorder"}

// String returns upstream's constant name.
func (p LineParam) String() string { return enumName("LineParam", lineParamNames, int(p)) }

// AlignmentParam is an alignment skinparam of the legacy skin.
type AlignmentParam int

// Alignment parameters, in upstream's order.
const (
	AlignmentParamArrowMessageAlignment AlignmentParam = iota
	AlignmentParamStateMessageAlignment
	AlignmentParamSequenceMessageAlignment
	AlignmentParamSequenceMessageTextAlignment
	AlignmentParamSequenceReferenceAlignment
	AlignmentParamPackageTitleAlignment
	AlignmentParamNoteTextAlignment
	alignmentParamCount
)

var alignmentParams = [...]struct {
	name         string
	defaultValue geom.HorizontalAlignment
}{
	{"arrowMessageAlignment", geom.HorizontalAlignmentLeft},
	{"stateMessageAlignment", geom.HorizontalAlignmentCenter},
	{"sequenceMessageAlignment", geom.HorizontalAlignmentLeft},
	{"sequenceMessageTextAlignment", geom.HorizontalAlignmentLeft},
	{"sequenceReferenceAlignment", geom.HorizontalAlignmentCenter},
	{"packageTitleAlignment", geom.HorizontalAlignmentCenter},
	{"noteTextAlignment", geom.HorizontalAlignmentLeft},
}

// String returns upstream's constant name.
func (p AlignmentParam) String() string { return alignmentParams[p].name }

// DefaultValue returns the default alignment.
func (p AlignmentParam) DefaultValue() geom.HorizontalAlignment {
	return alignmentParams[p].defaultValue
}

// CornerParam is a corner skinparam of the legacy skin.
type CornerParam int

// Corner parameters, in upstream's order.
const (
	CornerParamDefault CornerParam = iota
	CornerParamDiagramBorder
	CornerParamTitleBorder
	CornerParamRectangle
	CornerParamPerson
	CornerParamHexagon
	CornerParamArchimate
	CornerParamComponent
	CornerParamCard
	CornerParamAgent
)

var cornerParamNames = []string{"DEFAULT", "diagramBorder", "titleBorder", "rectangle", "person", "hexagon",
	"archimate", "component", "card", "agent"}

// String returns upstream's constant name.
func (p CornerParam) String() string { return enumName("CornerParam", cornerParamNames, int(p)) }

// RoundKey returns the skinparam of the round corner.
func (p CornerParam) RoundKey() string {
	if p == CornerParamDefault {
		return "roundcorner"
	}
	return p.String() + "roundcorner"
}

// DiagonalKey returns the skinparam of the diagonal corner.
func (p CornerParam) DiagonalKey() string {
	if p == CornerParamDefault {
		return "diagonalcorner"
	}
	return p.String() + "diagonalcorner"
}

// PaddingParam is a padding skinparam of the legacy skin.
type PaddingParam int

// Padding parameters, in upstream's order.
const (
	PaddingParamParticipant PaddingParam = iota
	PaddingParamBox
)

// String returns upstream's constant name.
func (p PaddingParam) String() string {
	return enumName("PaddingParam", []string{"PARTICIPANT", "BOX"}, int(p))
}

// SkinName returns the skinparam, for example "participantPadding".
func (p PaddingParam) SkinName() string { return javaLower(p.String()) + "Padding" }

// SplitParam is the page border, outside color and margin of split
// (multi-page) images.
type SplitParam struct {
	borderColor    color.HColor
	externalColor  color.HColor
	externalMargin int
}

// NewSplitParam returns the parameters; a border color without margin gets
// a margin of 1.
func NewSplitParam(borderColor, externalColor color.HColor, externalMargin int) *SplitParam {
	if borderColor != nil && externalMargin == 0 {
		externalMargin = 1
	}
	return &SplitParam{borderColor: borderColor, externalColor: externalColor, externalMargin: externalMargin}
}

// IsSet reports whether there is a margin.
func (s *SplitParam) IsSet() bool { return s.externalMargin > 0 }

// ExternalMargin returns the margin.
func (s *SplitParam) ExternalMargin() int { return s.externalMargin }

// BorderColor returns the border color; ok is false when there is none. A
// color that is not simple panics, as upstream's cast throws.
func (s *SplitParam) BorderColor() (color.XColor, bool) { return simpleAwtColor(s.borderColor) }

// ExternalColor returns the outside color; ok is false when there is none.
func (s *SplitParam) ExternalColor() (color.XColor, bool) { return simpleAwtColor(s.externalColor) }

func simpleAwtColor(c color.HColor) (color.XColor, bool) {
	if c == nil {
		return color.XColor{}, false
	}
	simple, ok := c.(*color.HColorSimple)
	if !ok {
		throw("java.lang.ClassCastException", "")
	}
	return simple.AwtColor(), true
}

// ComponentStyle is the shape of components. Upstream's toUSymbol belongs
// to the decoration.symbol package, which imports this one.
type ComponentStyle int

// Component styles, in upstream's order.
const (
	ComponentStyleUML1 ComponentStyle = iota
	ComponentStyleUML2
	ComponentStyleRectangle
)

// String returns upstream's constant name.
func (c ComponentStyle) String() string {
	return enumName("ComponentStyle", []string{"UML1", "UML2", "RECTANGLE"}, int(c))
}

// ActorStyle is the shape of actors. Its getTextBlock is TextBlock in
// skin_actors.go; upstream's toUSymbol belongs to the symbols step.
type ActorStyle int

// Actor styles, in upstream's order.
const (
	ActorStyleStickman ActorStyle = iota
	ActorStyleStickmanBusiness
	ActorStyleAwesome
	ActorStyleHollow
)

// String returns upstream's constant name.
func (a ActorStyle) String() string {
	return enumName("ActorStyle", []string{"STICKMAN", "STICKMAN_BUSINESS", "AWESOME", "HOLLOW"}, int(a))
}

// ArrowDirection is the direction of a sequence message.
type ArrowDirection int

// Arrow directions, in upstream's order.
const (
	ArrowDirectionLeftToRightNormal ArrowDirection = iota
	ArrowDirectionRightToLeftReverse
	ArrowDirectionSelf
	ArrowDirectionBothDirection
)

// ArrowDirectionNull stands for upstream's null direction, which many
// callers of ISkinParam.HorizontalAlignment pass.
const ArrowDirectionNull ArrowDirection = -1

// String returns upstream's constant name.
func (d ArrowDirection) String() string {
	return enumName("ArrowDirection", []string{"LEFT_TO_RIGHT_NORMAL", "RIGHT_TO_LEFT_REVERSE", "SELF",
		"BOTH_DIRECTION"}, int(d))
}

// Reverse returns the opposite direction; self and both panic, as upstream
// throws UnsupportedOperationException.
func (d ArrowDirection) Reverse() ArrowDirection {
	switch d {
	case ArrowDirectionLeftToRightNormal:
		return ArrowDirectionRightToLeftReverse
	case ArrowDirectionRightToLeftReverse:
		return ArrowDirectionLeftToRightNormal
	}
	throw(unsupportedOperation, "")
	return d
}
