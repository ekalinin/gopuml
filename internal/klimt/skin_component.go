// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/Component.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/AbstractComponent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/AbstractTextualComponent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowComponent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/Area.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/Context2D.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/SimpleContext2D.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ComponentType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LifeEventType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Component is an element of a sequence diagram drawn by the skin (an
// arrow, a participant head, a note...): it has a preferred size and draws
// itself in an area.
type Component interface {
	WithStyle
	PreferredDimension(stringBounder font.StringBounder) geom.XDimension2D
	PreferredWidth(stringBounder font.StringBounder) float64
	PreferredHeight(stringBounder font.StringBounder) float64
	DrawU(ug UGraphic, area *Area, context Context2D)
}

// ArrowComponent is the component of a message arrow.
type ArrowComponent interface {
	Component
	StartPoint(stringBounder font.StringBounder, dimensionToUse geom.XDimension2D) geom.XPoint2D
	EndPoint(stringBounder font.StringBounder, dimensionToUse geom.XDimension2D) geom.XPoint2D
	PaddingY() float64
	YPoint(stringBounder font.StringBounder) float64
	PosArrow(stringBounder font.StringBounder) float64
}

// AbstractComponentHooks are the methods of a concrete component that
// AbstractComponent calls (upstream's abstract and overridable methods).
// AbstractComponent provides DrawBackgroundInternalU, PaddingX and PaddingY;
// a component embedding it gets them by promotion unless it defines its
// own.
type AbstractComponentHooks interface {
	DrawInternalU(ug UGraphic, area *Area)
	DrawBackgroundInternalU(ug UGraphic, area *Area)
	PaddingX() float64
	PaddingY() float64
	PreferredWidth(stringBounder font.StringBounder) float64
	PreferredHeight(stringBounder font.StringBounder) float64
}

// AbstractComponent is upstream's AbstractComponent, the base of the
// components: a style and the skin parameters. A concrete component embeds
// it and calls InitAbstractComponent with itself, so that DrawU and
// PreferredDimension reach its methods.
type AbstractComponent struct {
	self      AbstractComponentHooks
	style     *Style
	skinParam ISkinParam
}

// InitAbstractComponent sets the concrete component, its style and the skin
// parameters.
func (c *AbstractComponent) InitAbstractComponent(self AbstractComponentHooks, style *Style, skinParam ISkinParam) {
	c.self = self
	c.style = style
	c.skinParam = skinParam
}

// UsedStyles panics, as upstream throws UnsupportedOperationException.
func (c *AbstractComponent) UsedStyles() []*Style {
	throw(unsupportedOperation, "")
	return nil
}

// StyleSignature panics, as upstream throws UnsupportedOperationException.
func (c *AbstractComponent) StyleSignature() StyleSignature {
	throw(unsupportedOperation, "")
	return nil
}

// Style returns the style of the component.
func (c *AbstractComponent) Style() *Style { return c.style }

// SkinParam returns the skin parameters.
func (c *AbstractComponent) SkinParam() ISkinParam { return c.skinParam }

// IHtmlColorSet returns the color set of the skin parameters.
func (c *AbstractComponent) IHtmlColorSet() *color.HColorSet { return c.skinParam.IHtmlColorSet() }

// SymbolContext returns the colors, stroke, shadow and corners of the
// style.
func (c *AbstractComponent) SymbolContext() *Fashion {
	return c.style.SymbolContext(c.IHtmlColorSet(), nil)
}

// SymbolContextColors returns SymbolContext with colors applied.
func (c *AbstractComponent) SymbolContextColors(colors *color.Colors) *Fashion {
	return c.style.SymbolContext(c.IHtmlColorSet(), colors)
}

func (c *AbstractComponent) colorOf(name PName) color.HColor {
	return c.style.Value(name).AsColor(c.IHtmlColorSet())
}

// ColorFont returns the FontColor of the style.
func (c *AbstractComponent) ColorFont() color.HColor { return c.colorOf(PNameFontColor) }

// ColorLine returns the LineColor of the style.
func (c *AbstractComponent) ColorLine() color.HColor { return c.colorOf(PNameLineColor) }

// ColorBackGround returns the BackGroundColor of the style.
func (c *AbstractComponent) ColorBackGround() color.HColor { return c.colorOf(PNameBackGroundColor) }

// RoundCorner returns the RoundCorner of the style.
func (c *AbstractComponent) RoundCorner() float64 {
	return float64(c.style.Value(PNameRoundCorner).AsInt())
}

// DiagonalCorner returns the DiagonalCorner of the style.
func (c *AbstractComponent) DiagonalCorner() float64 {
	return float64(c.style.Value(PNameDiagonalCorner).AsInt())
}

// Stroke returns the stroke of the style.
func (c *AbstractComponent) Stroke() UStroke { return c.style.Stroke() }

// Shadowing returns the shadowing of the style.
func (c *AbstractComponent) Shadowing() float64 { return c.style.Shadowing() }

// UFont returns the font of the style.
func (c *AbstractComponent) UFont() font.UFont { return c.style.UFont() }

// FontConfiguration returns the font configuration of the style.
func (c *AbstractComponent) FontConfiguration() *FontConfiguration {
	return c.style.FontConfiguration(c.IHtmlColorSet(), nil)
}

// HorizontalAlignment returns the horizontal alignment of the style.
func (c *AbstractComponent) HorizontalAlignment() geom.HorizontalAlignment {
	return c.style.HorizontalAlignment()
}

// DrawBackgroundInternalU draws nothing: the default background drawing.
func (c *AbstractComponent) DrawBackgroundInternalU(UGraphic, *Area) {}

// DrawU draws the component after its padding: its background in a
// background context, else its content.
func (c *AbstractComponent) DrawU(ug UGraphic, area *Area, context Context2D) {
	ug = ug.Apply(NewUTranslate(c.self.PaddingX(), c.self.PaddingY()))
	if context.IsBackground() {
		c.self.DrawBackgroundInternalU(ug, area)
	} else {
		c.self.DrawInternalU(ug, area)
	}
}

// PaddingX returns 0, the default horizontal padding.
func (c *AbstractComponent) PaddingX() float64 { return 0 }

// PaddingY returns 0, the default vertical padding.
func (c *AbstractComponent) PaddingY() float64 { return 0 }

// PreferredDimension returns the preferred width and height of the
// component.
func (c *AbstractComponent) PreferredDimension(stringBounder font.StringBounder) geom.XDimension2D {
	w := c.self.PreferredWidth(stringBounder)
	h := c.self.PreferredHeight(stringBounder)
	return geom.NewXDimension2D(w, h)
}

// AbstractTextualComponent is upstream's AbstractTextualComponent: a
// component with a text block (its label) and a padding.
type AbstractTextualComponent struct {
	AbstractComponent
	textBlock TextBlock
	padding   ClockwiseTopRightBottomLeft
}

// InitAbstractTextualComponentLabel initializes the component with a label
// whose lines are separated by "\n"; ok false (a null label) is the empty
// label.
func (c *AbstractTextualComponent) InitAbstractTextualComponentLabel(self AbstractComponentHooks, style *Style,
	maxMessageSize LineBreakStrategy, padding ClockwiseTopRightBottomLeft, skinParam ISkinParam, label string,
	ok bool) {
	if !ok {
		label = ""
	}
	c.InitAbstractTextualComponent(self, style, style, maxMessageSize, padding, skinParam,
		DisplayWithNewlines(skinParam.Pragma(), label), false)
}

// InitAbstractTextualComponent initializes the component with a display; the
// stereotype lines use the font of stereo, and an enhanced display has
// class-like separators ("--", "==", "..", "__").
func (c *AbstractTextualComponent) InitAbstractTextualComponent(self AbstractComponentHooks, style, stereo *Style,
	maxMessageSize LineBreakStrategy, padding ClockwiseTopRightBottomLeft, skinParam ISkinParam, display *Display,
	enhanced bool) {
	c.InitAbstractComponent(self, style, skinParam)
	c.padding = padding
	fc := c.FontConfiguration()
	horizontalAlignment := c.HorizontalAlignment()
	fontForStereotype := stereo.UFont()
	htmlColorForStereotype := stereo.Value(PNameFontColor).AsColor(c.IHtmlColorSet())
	display = display.WithoutStereotypeIfNeeded(style)
	switch {
	case display.Size() == 1 && javaLen(display.Get(0).String()) == 0:
		c.textBlock = NewTextBlockEmpty(0, 0)
	case enhanced:
		c.textBlock = BodyFactoryCreate3(display, skinParam, horizontalAlignment, fc, maxMessageSize, style)
	default:
		c.textBlock = display.Create0Margins(fc, horizontalAlignment, skinParam, maxMessageSize, CreoleModeFull,
			&fontForStereotype, htmlColorForStereotype, padding.Left(), padding.Right())
	}
}

// TextBlock returns the text block of the label.
func (c *AbstractTextualComponent) TextBlock() TextBlock { return c.textBlock }

// PureTextWidth returns the width of the text block.
func (c *AbstractTextualComponent) PureTextWidth(stringBounder font.StringBounder) float64 {
	return c.TextBlock().CalculateDimension(stringBounder).Width()
}

// TextWidth returns the width of the text block with the horizontal
// padding.
func (c *AbstractTextualComponent) TextWidth(stringBounder font.StringBounder) float64 {
	return c.PureTextWidth(stringBounder) + c.OldPaddingX1() + c.OldPaddingX2()
}

// TextHeight returns the height of the text block with the vertical
// padding.
func (c *AbstractTextualComponent) TextHeight(stringBounder font.StringBounder) float64 {
	size := c.TextBlock().CalculateDimension(stringBounder)
	return size.Height() + c.padding.Top() + c.padding.Bottom()
}

// OldPaddingX1 returns the left padding.
func (c *AbstractTextualComponent) OldPaddingX1() float64 { return c.padding.Left() }

// OldPaddingX2 returns the right padding.
func (c *AbstractTextualComponent) OldPaddingX2() float64 { return c.padding.Right() }

// OldPaddingY returns the top padding.
func (c *AbstractTextualComponent) OldPaddingY() float64 { return c.padding.Top() }

// Padding returns the padding.
func (c *AbstractTextualComponent) Padding() ClockwiseTopRightBottomLeft { return c.padding }

// Area is the space given to a component: its dimension and the offsets of
// the sequence layout.
type Area struct {
	dimensionToUse geom.XDimension2D
	deltaX1        float64
	liveDeltaSize  float64
	level          int
	textDeltaX     float64
}

// NewArea returns the area of a dimension.
func NewArea(dimensionToUse geom.XDimension2D) *Area { return &Area{dimensionToUse: dimensionToUse} }

// AreaCreate returns the area of a width and a height.
func AreaCreate(width, height float64) *Area { return NewArea(geom.NewXDimension2D(width, height)) }

// String is upstream's toString.
func (a *Area) String() string { return a.dimensionToUse.String() + " (" + jstr(a.deltaX1) + ")" }

// DimensionToUse returns the dimension.
func (a *Area) DimensionToUse() geom.XDimension2D { return a.dimensionToUse }

// SetDeltaX1 sets the horizontal offset.
func (a *Area) SetDeltaX1(deltaX1 float64) { a.deltaX1 = deltaX1 }

// DeltaX1 returns the horizontal offset.
func (a *Area) DeltaX1() float64 { return a.deltaX1 }

// SetLevel sets the activation level.
func (a *Area) SetLevel(level int) { a.level = level }

// Level returns the activation level.
func (a *Area) Level() int { return a.level }

// SetLiveDeltaSize sets the size offset of the activation boxes.
func (a *Area) SetLiveDeltaSize(liveDeltaSize float64) { a.liveDeltaSize = liveDeltaSize }

// LiveDeltaSize returns the size offset of the activation boxes.
func (a *Area) LiveDeltaSize() float64 { return a.liveDeltaSize }

// TextDeltaX returns the horizontal offset of the text.
func (a *Area) TextDeltaX() float64 { return a.textDeltaX }

// SetTextDeltaX sets the horizontal offset of the text.
func (a *Area) SetTextDeltaX(textDeltaX float64) { a.textDeltaX = textDeltaX }

// Context2D tells a component whether it draws its background or its
// content.
type Context2D interface {
	IsBackground() bool
}

// SimpleContext2D is a fixed Context2D.
type SimpleContext2D struct{ isBackground bool }

// NewSimpleContext2D returns a context.
func NewSimpleContext2D(isBackground bool) SimpleContext2D {
	return SimpleContext2D{isBackground: isBackground}
}

// IsBackground reports whether the background is drawn.
func (c SimpleContext2D) IsBackground() bool { return c.isBackground }

// ComponentType is the kind of a sequence component.
type ComponentType int

// The component types, in upstream's order.
const (
	ComponentTypeArrow ComponentType = iota
	ComponentTypeActorHead
	ComponentTypeActorTail
	ComponentTypeBoundaryHead
	ComponentTypeBoundaryTail
	ComponentTypeControlHead
	ComponentTypeControlTail
	ComponentTypeEntityHead
	ComponentTypeEntityTail
	ComponentTypeQueueHead
	ComponentTypeQueueTail
	ComponentTypeDatabaseHead
	ComponentTypeDatabaseTail
	ComponentTypeCollectionsHead
	ComponentTypeCollectionsTail
	ComponentTypeActivationBoxCloseClose
	ComponentTypeActivationBoxCloseOpen
	ComponentTypeActivationBoxOpenClose
	ComponentTypeActivationBoxOpenOpen
	ComponentTypeDelayText
	ComponentTypeDestroy
	ComponentTypeDelayLine
	ComponentTypeParticipantLine
	ComponentTypeGroupingElseLegacy
	ComponentTypeGroupingElseTeoz
	ComponentTypeGroupingHeaderLegacy
	ComponentTypeGroupingHeaderTeoz
	ComponentTypeGroupingSpace
	ComponentTypeNewpage
	ComponentTypeNote
	ComponentTypeNoteHexagonal
	ComponentTypeNoteBox
	ComponentTypeDivider
	ComponentTypeReference
	ComponentTypeEnglober
	ComponentTypeParticipantHead
	ComponentTypeParticipantTail
)

var componentTypeNames = []string{"ARROW", "ACTOR_HEAD", "ACTOR_TAIL", "BOUNDARY_HEAD", "BOUNDARY_TAIL",
	"CONTROL_HEAD", "CONTROL_TAIL", "ENTITY_HEAD", "ENTITY_TAIL", "QUEUE_HEAD", "QUEUE_TAIL", "DATABASE_HEAD",
	"DATABASE_TAIL", "COLLECTIONS_HEAD", "COLLECTIONS_TAIL", "ACTIVATION_BOX_CLOSE_CLOSE",
	"ACTIVATION_BOX_CLOSE_OPEN", "ACTIVATION_BOX_OPEN_CLOSE", "ACTIVATION_BOX_OPEN_OPEN", "DELAY_TEXT", "DESTROY",
	"DELAY_LINE", "PARTICIPANT_LINE", "GROUPING_ELSE_LEGACY", "GROUPING_ELSE_TEOZ", "GROUPING_HEADER_LEGACY",
	"GROUPING_HEADER_TEOZ", "GROUPING_SPACE", "NEWPAGE", "NOTE", "NOTE_HEXAGONAL", "NOTE_BOX", "DIVIDER",
	"REFERENCE", "ENGLOBER", "PARTICIPANT_HEAD", "PARTICIPANT_TAIL"}

// String returns upstream's constant name.
func (t ComponentType) String() string { return enumName("ComponentType", componentTypeNames, int(t)) }

// ComponentTypeValues returns the component types in upstream's order.
func ComponentTypeValues() []ComponentType {
	result := make([]ComponentType, len(componentTypeNames))
	for i := range result {
		result[i] = ComponentType(i)
	}
	return result
}

// IsArrow reports whether the type is ARROW.
func (t ComponentType) IsArrow() bool { return t == ComponentTypeArrow }

// StyleSignature returns the style signature of the type; the types without
// one panic, as upstream throws UnsupportedOperationException.
func (t ComponentType) StyleSignature() StyleSignature {
	seq := func(names ...SName) StyleSignature {
		return StyleSignatureBasicOf(append([]SName{SNameRoot, SNameElement, SNameSequenceDiagram}, names...)...)
	}
	switch t {
	case ComponentTypeParticipantHead, ComponentTypeParticipantTail:
		return seq(SNameParticipant)
	case ComponentTypeParticipantLine:
		return seq(SNameLifeLine)
	case ComponentTypeActivationBoxCloseClose, ComponentTypeActivationBoxCloseOpen,
		ComponentTypeActivationBoxOpenClose, ComponentTypeActivationBoxOpenOpen:
		return seq(SNameActivationBox)
	case ComponentTypeDestroy:
		// sequencediagram.LifeEventType.DESTROY.getStyleSignature()
		return seq(SNameLifeLine, SNameDestroy)
	case ComponentTypeDivider:
		return seq(SNameSeparator)
	case ComponentTypeEnglober:
		return seq(SNameBox)
	case ComponentTypeNewpage:
		return seq(SNameNewpage)
	case ComponentTypeNote:
		return seq(SNameNote)
	case ComponentTypeDelayText, ComponentTypeDelayLine:
		return seq(SNameLifeLine, SNameDelay)
	}
	throw(unsupportedOperation, t.String())
	return nil
}
