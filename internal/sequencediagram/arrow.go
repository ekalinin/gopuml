// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowConfiguration.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowBody.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowDecoration.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowDressing.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowHead.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowPart.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ComponentType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

// The arrow configuration classes belong to upstream's skin package, which
// lives in internal/klimt; they are here until the sequence drawing moves
// them there with the other skin components.

import (
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// ArrowBody is the line of an arrow.
type ArrowBody int

// Arrow bodies, in upstream's order.
const (
	ArrowBodyNormal ArrowBody = iota
	ArrowBodyDotted
	ArrowBodyDashed
	ArrowBodyHidden
	ArrowBodyBold
)

var arrowBodyNames = [...]string{"NORMAL", "DOTTED", "DASHED", "HIDDEN", "BOLD"}

func (b ArrowBody) String() string { return arrowBodyNames[b] }

// ArrowDecoration is what ends an arrow besides its head.
type ArrowDecoration int

// Arrow decorations, in upstream's order.
const (
	ArrowDecorationNone ArrowDecoration = iota
	ArrowDecorationCircle
)

var arrowDecorationNames = [...]string{"NONE", "CIRCLE"}

func (d ArrowDecoration) String() string { return arrowDecorationNames[d] }

// ArrowHead is the head of an arrow.
type ArrowHead int

// Arrow heads, in upstream's order.
const (
	ArrowHeadNormal ArrowHead = iota
	ArrowHeadCrossX
	ArrowHeadAsync
	ArrowHeadNone
)

var arrowHeadNames = [...]string{"NORMAL", "CROSSX", "ASYNC", "NONE"}

func (h ArrowHead) String() string { return arrowHeadNames[h] }

// ArrowPart is the half of a head that is drawn.
type ArrowPart int

// Arrow parts, in upstream's order.
const (
	ArrowPartFull ArrowPart = iota
	ArrowPartTopPart
	ArrowPartBottomPart
)

var arrowPartNames = [...]string{"FULL", "TOP_PART", "BOTTOM_PART"}

func (p ArrowPart) String() string { return arrowPartNames[p] }

// ArrowDressing is one end of an arrow: its head and the drawn part of it.
type ArrowDressing struct {
	Head ArrowHead
	Part ArrowPart
}

// newArrowDressing is upstream's ArrowDressing.create(): no head, full.
func newArrowDressing() ArrowDressing { return ArrowDressing{Head: ArrowHeadNone, Part: ArrowPartFull} }

// String is upstream's toString: the name of the head.
func (d ArrowDressing) String() string { return d.Head.String() }

// ArrowConfiguration describes a message arrow. It is a value: the With
// methods return modified copies.
type ArrowConfiguration struct {
	body                     ArrowBody
	dressing1, dressing2     ArrowDressing
	decoration1, decoration2 ArrowDecoration
	color                    color.HColor
	isSelf                   bool
	thickness                float64
	reverseDefine            bool
	inclination              int
}

// ArrowConfigurationNormal is a left to right arrow.
func ArrowConfigurationNormal() ArrowConfiguration {
	d2 := newArrowDressing()
	d2.Head = ArrowHeadNormal
	return ArrowConfiguration{body: ArrowBodyNormal, dressing1: newArrowDressing(), dressing2: d2, thickness: 1}
}

// ArrowConfigurationBoth is an arrow with heads at both ends.
func ArrowConfigurationBoth() ArrowConfiguration {
	c := ArrowConfigurationNormal()
	c.dressing1.Head = ArrowHeadNormal
	return c
}

// ArrowConfigurationSelf is a self arrow.
func ArrowConfigurationSelf(reverseDefine bool) ArrowConfiguration {
	c := ArrowConfigurationNormal()
	c.isSelf = true
	c.reverseDefine = reverseDefine
	return c
}

// Name is upstream's name() and toString(), for example
// "NORMAL(NONE NONE)(NORMAL NONE)false null".
func (c ArrowConfiguration) Name() string {
	col := "null"
	if c.color != nil {
		col = c.color.String()
	}
	return c.body.String() + "(" + c.dressing1.String() + " " + c.decoration1.String() + ")(" +
		c.dressing2.String() + " " + c.decoration2.String() + ")" + strconv.FormatBool(c.isSelf) + " " + col
}

func (c ArrowConfiguration) String() string { return c.Name() }

// Reverse swaps the ends.
func (c ArrowConfiguration) Reverse() ArrowConfiguration {
	c.dressing1, c.dressing2 = c.dressing2, c.dressing1
	c.decoration1, c.decoration2 = c.decoration2, c.decoration1
	return c
}

// Self returns the configuration as a self arrow.
func (c ArrowConfiguration) Self() ArrowConfiguration {
	c.isSelf = true
	return c
}

// WithBody returns the configuration with another body.
func (c ArrowConfiguration) WithBody(body ArrowBody) ArrowConfiguration {
	c.body = body
	return c
}

// WithHead sets the head of the ends that have one.
func (c ArrowConfiguration) WithHead(head ArrowHead) ArrowConfiguration {
	if c.dressing1.Head != ArrowHeadNone {
		c.dressing1.Head = head
	}
	if c.dressing2.Head != ArrowHeadNone {
		c.dressing2.Head = head
	}
	return c
}

// WithHead1 sets the head of the first end.
func (c ArrowConfiguration) WithHead1(head ArrowHead) ArrowConfiguration {
	c.dressing1.Head = head
	return c
}

// WithHead2 sets the head of the second end.
func (c ArrowConfiguration) WithHead2(head ArrowHead) ArrowConfiguration {
	c.dressing2.Head = head
	return c
}

// WithPart sets the drawn part of the head, on the second end when it has
// a head and on the first otherwise.
func (c ArrowConfiguration) WithPart(part ArrowPart) ArrowConfiguration {
	if c.dressing2.Head != ArrowHeadNone {
		c.dressing2.Part = part
	} else {
		c.dressing1.Part = part
	}
	return c
}

// WithDecoration1 sets the decoration of the first end.
func (c ArrowConfiguration) WithDecoration1(d ArrowDecoration) ArrowConfiguration {
	c.decoration1 = d
	return c
}

// WithDecoration2 sets the decoration of the second end.
func (c ArrowConfiguration) WithDecoration2(d ArrowDecoration) ArrowConfiguration {
	c.decoration2 = d
	return c
}

// WithColor sets the color, nil for none.
func (c ArrowConfiguration) WithColor(col color.HColor) ArrowConfiguration {
	c.color = col
	return c
}

// WithThickness sets the thickness.
func (c ArrowConfiguration) WithThickness(thickness float64) ArrowConfiguration {
	c.thickness = thickness
	return c
}

// ReverseDefine flips whether the arrow was written right to left.
func (c ArrowConfiguration) ReverseDefine() ArrowConfiguration {
	c.reverseDefine = !c.reverseDefine
	return c
}

// WithInclination sets the inclination of the arrow.
func (c ArrowConfiguration) WithInclination(inclination int) ArrowConfiguration {
	c.inclination = inclination
	return c
}

// Body returns the line of the arrow.
func (c ArrowConfiguration) Body() ArrowBody { return c.body }

// Decoration1 returns the decoration of the first end.
func (c ArrowConfiguration) Decoration1() ArrowDecoration { return c.decoration1 }

// Decoration2 returns the decoration of the second end.
func (c ArrowConfiguration) Decoration2() ArrowDecoration { return c.decoration2 }

// Dressing1 returns the first end.
func (c ArrowConfiguration) Dressing1() ArrowDressing { return c.dressing1 }

// Dressing2 returns the second end.
func (c ArrowConfiguration) Dressing2() ArrowDressing { return c.dressing2 }

// Color returns the color, nil for none.
func (c ArrowConfiguration) Color() color.HColor { return c.color }

// Thickness returns the thickness.
func (c ArrowConfiguration) Thickness() float64 { return c.thickness }

// IsSelf reports whether this is a self arrow.
func (c ArrowConfiguration) IsSelf() bool { return c.isSelf }

// IsReverseDefine reports whether the arrow was written right to left.
func (c ArrowConfiguration) IsReverseDefine() bool { return c.reverseDefine }

// Inclination returns the raw inclination.
func (c ArrowConfiguration) Inclination() int { return c.inclination }

// ArrowDirection is the direction of the arrow.
func (c ArrowConfiguration) ArrowDirection() klimt.ArrowDirection {
	switch {
	case c.isSelf:
		return klimt.ArrowDirectionSelf
	case c.dressing1.Head == ArrowHeadNone && c.dressing2.Head != ArrowHeadNone:
		return klimt.ArrowDirectionLeftToRightNormal
	case c.dressing1.Head != ArrowHeadNone && c.dressing2.Head == ArrowHeadNone:
		return klimt.ArrowDirectionRightToLeftReverse
	}
	return klimt.ArrowDirectionBothDirection
}

// IsSelfArrow reports whether the direction is SELF.
func (c ArrowConfiguration) IsSelfArrow() bool { return c.ArrowDirection() == klimt.ArrowDirectionSelf }

// IsDotted reports whether the line is dotted.
func (c ArrowConfiguration) IsDotted() bool { return c.body == ArrowBodyDotted }

// IsHidden reports whether the arrow is hidden.
func (c ArrowConfiguration) IsHidden() bool { return c.body == ArrowBodyHidden }

// Head returns the head of the second end when it has one, else the head
// of the first end.
func (c ArrowConfiguration) Head() ArrowHead {
	if c.dressing2.Head != ArrowHeadNone {
		return c.dressing2.Head
	}
	return c.dressing1.Head
}

// IsAsync1 reports whether the first end is asynchronous.
func (c ArrowConfiguration) IsAsync1() bool { return c.dressing1.Head == ArrowHeadAsync }

// IsAsync2 reports whether the second end is asynchronous.
func (c ArrowConfiguration) IsAsync2() bool { return c.dressing2.Head == ArrowHeadAsync }

// Part returns the drawn part of the head.
func (c ArrowConfiguration) Part() ArrowPart {
	if c.dressing2.Head != ArrowHeadNone {
		return c.dressing2.Part
	}
	return c.dressing1.Part
}

// Inclination1 is the inclination of the first end.
func (c ArrowConfiguration) Inclination1() int {
	if c.dressing2.Head == ArrowHeadNone || c.dressing2.Head == ArrowHeadCrossX {
		return c.inclination
	}
	return 0
}

// Inclination2 is the inclination of the second end.
func (c ArrowConfiguration) Inclination2() int {
	if c.dressing1.Head == ArrowHeadNone || c.dressing1.Head == ArrowHeadCrossX ||
		c.dressing1.Head == ArrowHeadNormal {
		return c.inclination
	}
	return 0
}

// ComponentType is a kind of sequence diagram drawing component, upstream's
// skin.ComponentType.
type ComponentType int

// Component types, in upstream's order.
const (
	ComponentArrow ComponentType = iota
	ComponentActorHead
	ComponentActorTail
	ComponentBoundaryHead
	ComponentBoundaryTail
	ComponentControlHead
	ComponentControlTail
	ComponentEntityHead
	ComponentEntityTail
	ComponentQueueHead
	ComponentQueueTail
	ComponentDatabaseHead
	ComponentDatabaseTail
	ComponentCollectionsHead
	ComponentCollectionsTail
	ComponentActivationBoxCloseClose
	ComponentActivationBoxCloseOpen
	ComponentActivationBoxOpenClose
	ComponentActivationBoxOpenOpen
	ComponentDelayText
	ComponentDestroy
	ComponentDelayLine
	ComponentParticipantLine
	ComponentGroupingElseLegacy
	ComponentGroupingElseTeoz
	ComponentGroupingHeaderLegacy
	ComponentGroupingHeaderTeoz
	ComponentGroupingSpace
	ComponentNewpage
	ComponentNote
	ComponentNoteHexagonal
	ComponentNoteBox
	ComponentDivider
	ComponentReference
	ComponentEnglober
	ComponentParticipantHead
	ComponentParticipantTail
)

var componentTypeNames = [...]string{"ARROW", "ACTOR_HEAD", "ACTOR_TAIL", "BOUNDARY_HEAD", "BOUNDARY_TAIL",
	"CONTROL_HEAD", "CONTROL_TAIL", "ENTITY_HEAD", "ENTITY_TAIL", "QUEUE_HEAD", "QUEUE_TAIL", "DATABASE_HEAD",
	"DATABASE_TAIL", "COLLECTIONS_HEAD", "COLLECTIONS_TAIL", "ACTIVATION_BOX_CLOSE_CLOSE",
	"ACTIVATION_BOX_CLOSE_OPEN", "ACTIVATION_BOX_OPEN_CLOSE", "ACTIVATION_BOX_OPEN_OPEN", "DELAY_TEXT", "DESTROY",
	"DELAY_LINE", "PARTICIPANT_LINE", "GROUPING_ELSE_LEGACY", "GROUPING_ELSE_TEOZ", "GROUPING_HEADER_LEGACY",
	"GROUPING_HEADER_TEOZ", "GROUPING_SPACE", "NEWPAGE", "NOTE", "NOTE_HEXAGONAL", "NOTE_BOX", "DIVIDER",
	"REFERENCE", "ENGLOBER", "PARTICIPANT_HEAD", "PARTICIPANT_TAIL"}

func (t ComponentType) String() string { return componentTypeNames[t] }

// IsArrow reports whether t is ComponentArrow.
func (t ComponentType) IsArrow() bool { return t == ComponentArrow }

// StyleSignature returns the style signature of the component. Types
// without one panic with UnsupportedOperationException, as upstream.
func (t ComponentType) StyleSignature() klimt.StyleSignatureBasic {
	seq := func(names ...klimt.SName) klimt.StyleSignatureBasic {
		return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
			klimt.SNameSequenceDiagram}, names...)...)
	}
	switch t {
	case ComponentParticipantHead, ComponentParticipantTail:
		return seq(klimt.SNameParticipant)
	case ComponentParticipantLine:
		return seq(klimt.SNameLifeLine)
	case ComponentActivationBoxCloseClose, ComponentActivationBoxCloseOpen, ComponentActivationBoxOpenClose,
		ComponentActivationBoxOpenOpen:
		return seq(klimt.SNameActivationBox)
	case ComponentDestroy:
		return LifeEventDestroy.StyleSignature()
	case ComponentDivider:
		return seq(klimt.SNameSeparator)
	case ComponentEnglober:
		return seq(klimt.SNameBox)
	case ComponentNewpage:
		return seq(klimt.SNameNewpage)
	case ComponentNote:
		return seq(klimt.SNameNote)
	case ComponentDelayText, ComponentDelayLine:
		return seq(klimt.SNameLifeLine, klimt.SNameDelay)
	}
	panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException", Message: t.String()})
}
