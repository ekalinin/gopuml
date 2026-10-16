// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ISkinParam.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ISkinSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/NoStyleAvailableException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// ISkinSimple is the part of the skin parameters that texts need.
type ISkinSimple interface {
	SpriteContainer
	// Value returns the value of a skinparam; ok is false when it is not
	// set (null upstream).
	Value(key string) (value string, ok bool)
	Values() map[string]string
	Padding() ClockwiseTopRightBottomLeft
	MonospacedFamily() string
	TabSize() int
	IHtmlColorSet() *color.HColorSet
	Dpi() int
	CopyAllFrom(other map[string]string)
	// Sheet returns the creole parser of a text.
	Sheet(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
		creoleMode CreoleMode) SheetBuilder
	// SheetStereo returns the creole parser of a text with a stereotype
	// font (upstream's four-argument sheet).
	SheetStereo(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
		creoleMode CreoleMode, stereo *FontConfiguration) SheetBuilder
	Pragma() *Pragma
	Options() *ConfigurationStore[OptionKey]
	// PathSystem returns the render context, which stands for upstream's
	// nio.PathSystem (the file access of includes and images).
	PathSystem() *core.RenderContext
}

// ISkinParamSwimlaneWidthSame is the swimlane width of "skinparam
// swimlaneWidth same".
const ISkinParamSwimlaneWidthSame = -1

// ISkinParam is the skin parameters of a diagram: skinparam values and the
// style sheet.
type ISkinParam interface {
	ISkinSimple
	HyperlinkColor() color.HColor
	// UseUnderlineForHyperlink returns the underline of links, nil for none.
	UseUnderlineForHyperlink() *UStroke
	BackgroundColor() color.HColor
	// HtmlColor returns the color of a parameter, nil when it is not set.
	HtmlColor(param ColorParam, stereotype *Stereotype, clickable bool) color.HColor
	Colors(param ColorParam, stereotype *Stereotype) (*color.Colors, error)
	FontHtmlColor(stereotype *Stereotype, param ...FontParam) color.HColor
	// Thickness returns the stroke of a line parameter, nil when it is not
	// set.
	Thickness(param LineParam, stereotype *Stereotype) *UStroke
	Font(stereotype *Stereotype, inPackageTitle bool, fontParam ...FontParam) font.UFont
	// HorizontalAlignment returns the alignment of a parameter;
	// arrowDirection may be ArrowDirectionNull and overrideDefault
	// HorizontalAlignmentNull.
	HorizontalAlignment(param AlignmentParam, arrowDirection ArrowDirection, isReverseDefine bool,
		overrideDefault geom.HorizontalAlignment) geom.HorizontalAlignment
	DefaultTextAlignment(defaultValue geom.HorizontalAlignment) geom.HorizontalAlignment
	StereotypeAlignment() geom.HorizontalAlignment
	CircledCharacterRadius() int
	CircledCharacter(stereotype *Stereotype) rune
	ClassAttributeIconSize() int
	DotSplines() DotSplines
	Shadowing(stereotype *Stereotype) bool
	ShadowingForNote(stereotype *Stereotype) bool
	PackageStyle() PackageStyle
	ComponentStyle() ComponentStyle
	StereotypePositionTop() bool
	UseSwimlanes(diagramType core.DiagramType) bool
	Nodesep() float64
	Ranksep() float64
	RoundCorner(param CornerParam, stereotype *Stereotype) float64
	DiagonalCorner(param CornerParam, stereotype *Stereotype) float64
	MaxMessageSize() LineBreakStrategy
	SwimlaneWrapTitleWidth() LineBreakStrategy
	StrictUmlStyle() bool
	ForceSequenceParticipantUnderlined() bool
	ConditionStyle() ConditionStyle
	ConditionEndStyle() ConditionEndStyle
	SameClassWidth() bool
	Rankdir() geom.Rankdir
	UseOctagonForActivity(stereotype *Stereotype) bool
	GroupInheritance() int
	Guillemet() *Guillemet
	Handwritten() bool
	SvgLinkTarget() string
	PreserveAspectRatio() string
	MaxAsciiMessageLength() int
	ColorArrowSeparationSpace() int
	SplitParam() *SplitParam
	SwimlaneWidth() int
	DiagramType() core.DiagramType
	// HoverPathColor returns the color of hovered links, nil for none.
	HoverPathColor() color.HColor
	TikzFontDistortion() *TikzFontDistortion
	PaddingTOBEREMOVED(param PaddingParam) ClockwiseTopRightBottomLeft
	UseRankSame() bool
	DisplayGenericWithOldFashion() bool
	ResponseMessageBelowArrow() bool
	SvgDimensionStyle() bool
	FixCircleLabelOverlapping() bool
	SetUseVizJs(useVizJs bool)
	IsUseVizJs() bool
	SequenceDiagramPadder() *Padder
	CurrentStyleBuilder() *StyleBuilder
	MuteStyle(modifiedStyles []*Style)
	AllSpriteNames() []string
	DefaultSkin() string
	SetDefaultSkin(newSkin string)
	ActorStyle() ActorStyle
	SetSvgSize(origin, sizeToUse string)
	LengthAdjust() LengthAdjust
	ParamSameClassWidth() float64
	Arrows() Arrows
}

// NoStyleAvailableError is upstream's NoStyleAvailableException: no style
// sheet could be loaded.
type NoStyleAvailableError struct{}

func (*NoStyleAvailableError) Error() string {
	return "net.sourceforge.plantuml.style.NoStyleAvailableException"
}
