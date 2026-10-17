// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowConfiguration.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowBody.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowDecoration.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowDressing.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowHead.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ArrowPart.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// ArrowBody is the line of a sequence arrow.
type ArrowBody int

// The arrow bodies, in upstream's order.
const (
	ArrowBodyNormal ArrowBody = iota
	ArrowBodyDotted
	ArrowBodyDashed
	ArrowBodyHidden
	ArrowBodyBold
)

// String returns upstream's constant name.
func (b ArrowBody) String() string {
	return enumName("ArrowBody", []string{"NORMAL", "DOTTED", "DASHED", "HIDDEN", "BOLD"}, int(b))
}

// ArrowDecoration is the decoration at an end of a sequence arrow.
type ArrowDecoration int

// The arrow decorations, in upstream's order.
const (
	ArrowDecorationNone ArrowDecoration = iota
	ArrowDecorationCircle
)

// String returns upstream's constant name.
func (d ArrowDecoration) String() string {
	return enumName("ArrowDecoration", []string{"NONE", "CIRCLE"}, int(d))
}

// ArrowHead is the head at an end of a sequence arrow.
type ArrowHead int

// The arrow heads, in upstream's order.
const (
	ArrowHeadNormal ArrowHead = iota
	ArrowHeadCrossx
	ArrowHeadAsync
	ArrowHeadNone
)

// String returns upstream's constant name.
func (h ArrowHead) String() string {
	return enumName("ArrowHead", []string{"NORMAL", "CROSSX", "ASYNC", "NONE"}, int(h))
}

// ArrowPart is the part of an arrow head that is drawn.
type ArrowPart int

// The arrow parts, in upstream's order.
const (
	ArrowPartFull ArrowPart = iota
	ArrowPartTopPart
	ArrowPartBottomPart
)

// String returns upstream's constant name.
func (p ArrowPart) String() string {
	return enumName("ArrowPart", []string{"FULL", "TOP_PART", "BOTTOM_PART"}, int(p))
}

// ArrowDressing is an end of a sequence arrow: its head and the part of
// the head that is drawn.
type ArrowDressing struct {
	head ArrowHead
	part ArrowPart
}

// ArrowDressingCreate returns the end without head.
func ArrowDressingCreate() ArrowDressing {
	return ArrowDressing{head: ArrowHeadNone, part: ArrowPartFull}
}

// Name returns the name of the head.
func (d ArrowDressing) Name() string { return d.String() }

// String returns the name of the head.
func (d ArrowDressing) String() string { return d.head.String() }

// WithHead returns the end with another head.
func (d ArrowDressing) WithHead(head ArrowHead) ArrowDressing {
	return ArrowDressing{head: head, part: d.part}
}

// WithPart returns the end with another part.
func (d ArrowDressing) WithPart(part ArrowPart) ArrowDressing {
	return ArrowDressing{head: d.head, part: part}
}

// Head returns the head.
func (d ArrowDressing) Head() ArrowHead { return d.head }

// Part returns the part.
func (d ArrowDressing) Part() ArrowPart { return d.part }

// ArrowConfiguration is the look of a sequence arrow: body, both ends,
// decorations, color, thickness and inclination. It is immutable.
type ArrowConfiguration struct {
	body          ArrowBody
	dressing1     ArrowDressing
	dressing2     ArrowDressing
	decoration1   ArrowDecoration
	decoration2   ArrowDecoration
	color         color.HColor
	isSelf        bool
	thickness     float64
	reverseDefine bool
	inclination   int
}

// String is upstream's toString: the name.
func (a *ArrowConfiguration) String() string { return a.Name() }

// Name describes the configuration.
func (a *ArrowConfiguration) Name() string {
	c := "null"
	if a.color != nil {
		c = a.color.String()
	}
	return a.body.String() + "(" + a.dressing1.Name() + " " + a.decoration1.String() + ")(" + a.dressing2.Name() +
		" " + a.decoration2.String() + ")" + strconv.FormatBool(a.isSelf) + " " + c
}

func (a *ArrowConfiguration) with(change func(c *ArrowConfiguration)) *ArrowConfiguration {
	c := *a
	change(&c)
	return &c
}

// ArrowConfigurationWithDirectionNormal returns the arrow from left to
// right.
func ArrowConfigurationWithDirectionNormal() *ArrowConfiguration {
	return &ArrowConfiguration{body: ArrowBodyNormal, dressing1: ArrowDressingCreate(),
		dressing2: ArrowDressingCreate().WithHead(ArrowHeadNormal), thickness: 1}
}

// ArrowConfigurationWithDirectionBoth returns the arrow with a head at
// both ends.
func ArrowConfigurationWithDirectionBoth() *ArrowConfiguration {
	return &ArrowConfiguration{body: ArrowBodyNormal, dressing1: ArrowDressingCreate().WithHead(ArrowHeadNormal),
		dressing2: ArrowDressingCreate().WithHead(ArrowHeadNormal), thickness: 1}
}

// ArrowConfigurationWithDirectionSelf returns the arrow to self.
func ArrowConfigurationWithDirectionSelf(reverseDefine bool) *ArrowConfiguration {
	return &ArrowConfiguration{body: ArrowBodyNormal, dressing1: ArrowDressingCreate(),
		dressing2: ArrowDressingCreate().WithHead(ArrowHeadNormal), isSelf: true, thickness: 1,
		reverseDefine: reverseDefine}
}

// ArrowConfigurationWithDirectionReverse returns the arrow from right to
// left.
func ArrowConfigurationWithDirectionReverse() *ArrowConfiguration {
	return ArrowConfigurationWithDirectionNormal().Reverse()
}

// Reverse returns the arrow with its ends swapped.
func (a *ArrowConfiguration) Reverse() *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) {
		c.dressing1, c.dressing2 = a.dressing2, a.dressing1
		c.decoration1, c.decoration2 = a.decoration2, a.decoration1
	})
}

// Self returns the arrow as an arrow to self.
func (a *ArrowConfiguration) Self() *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.isSelf = true })
}

// WithBody returns the arrow with another body.
func (a *ArrowConfiguration) WithBody(typ ArrowBody) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.body = typ })
}

// WithHead returns the arrow with another head on the ends that have one.
func (a *ArrowConfiguration) WithHead(head ArrowHead) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) {
		c.dressing1 = arrowAddHead(a.dressing1, head)
		c.dressing2 = arrowAddHead(a.dressing2, head)
	})
}

func arrowAddHead(dressing ArrowDressing, head ArrowHead) ArrowDressing {
	if dressing.Head() == ArrowHeadNone {
		return dressing
	}
	return dressing.WithHead(head)
}

// WithHead1 returns the arrow with another head at its start.
func (a *ArrowConfiguration) WithHead1(head ArrowHead) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.dressing1 = a.dressing1.WithHead(head) })
}

// WithHead2 returns the arrow with another head at its end.
func (a *ArrowConfiguration) WithHead2(head ArrowHead) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.dressing2 = a.dressing2.WithHead(head) })
}

// WithPart returns the arrow with another part of the head (the end head,
// else the start head).
func (a *ArrowConfiguration) WithPart(part ArrowPart) *ArrowConfiguration {
	if a.dressing2.Head() != ArrowHeadNone {
		return a.with(func(c *ArrowConfiguration) { c.dressing2 = a.dressing2.WithPart(part) })
	}
	return a.with(func(c *ArrowConfiguration) { c.dressing1 = a.dressing1.WithPart(part) })
}

// WithDecoration1 returns the arrow with another start decoration.
func (a *ArrowConfiguration) WithDecoration1(decoration1 ArrowDecoration) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.decoration1 = decoration1 })
}

// WithDecoration2 returns the arrow with another end decoration.
func (a *ArrowConfiguration) WithDecoration2(decoration2 ArrowDecoration) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.decoration2 = decoration2 })
}

// WithColor returns the arrow with another color (nil for none).
func (a *ArrowConfiguration) WithColor(c color.HColor) *ArrowConfiguration {
	return a.with(func(n *ArrowConfiguration) { n.color = c })
}

// Decoration1 returns the start decoration.
func (a *ArrowConfiguration) Decoration1() ArrowDecoration { return a.decoration1 }

// Decoration2 returns the end decoration.
func (a *ArrowConfiguration) Decoration2() ArrowDecoration { return a.decoration2 }

// ArrowDirection returns the direction given by the heads.
func (a *ArrowConfiguration) ArrowDirection() ArrowDirection {
	switch {
	case a.isSelf:
		return ArrowDirectionSelf
	case a.dressing1.Head() == ArrowHeadNone && a.dressing2.Head() != ArrowHeadNone:
		return ArrowDirectionLeftToRightNormal
	case a.dressing1.Head() != ArrowHeadNone && a.dressing2.Head() == ArrowHeadNone:
		return ArrowDirectionRightToLeftReverse
	}
	return ArrowDirectionBothDirection
}

// IsSelfArrow reports whether the arrow goes to self.
func (a *ArrowConfiguration) IsSelfArrow() bool { return a.ArrowDirection() == ArrowDirectionSelf }

// IsDotted reports whether the body is dotted.
func (a *ArrowConfiguration) IsDotted() bool { return a.body == ArrowBodyDotted }

// IsHidden reports whether the body is hidden.
func (a *ArrowConfiguration) IsHidden() bool { return a.body == ArrowBodyHidden }

// Head returns the end head, else the start head.
func (a *ArrowConfiguration) Head() ArrowHead {
	if a.dressing2.Head() != ArrowHeadNone {
		return a.dressing2.Head()
	}
	return a.dressing1.Head()
}

// IsAsync1 reports whether the start head is asynchronous.
func (a *ArrowConfiguration) IsAsync1() bool { return a.dressing1.Head() == ArrowHeadAsync }

// IsAsync2 reports whether the end head is asynchronous.
func (a *ArrowConfiguration) IsAsync2() bool { return a.dressing2.Head() == ArrowHeadAsync }

// Part returns the part of the end head, else of the start head.
func (a *ArrowConfiguration) Part() ArrowPart {
	if a.dressing2.Head() != ArrowHeadNone {
		return a.dressing2.Part()
	}
	return a.dressing1.Part()
}

// Color returns the color, nil for none.
func (a *ArrowConfiguration) Color() color.HColor { return a.color }

// Dressing1 returns the start.
func (a *ArrowConfiguration) Dressing1() ArrowDressing { return a.dressing1 }

// Dressing2 returns the end.
func (a *ArrowConfiguration) Dressing2() ArrowDressing { return a.dressing2 }

// ArrowConfigurationStroke applies a stroke to ug.
func ArrowConfigurationStroke(ug UGraphic, dashVisible, dashSpace, thickness float64) UGraphic {
	return ug.Apply(NewUStroke(dashVisible, dashSpace, thickness))
}

// ApplyStroke applies the stroke of the style; a dotted arrow whose style
// stroke has no dashes gets 2-2 dashes (return arrows).
func (a *ArrowConfiguration) ApplyStroke(ug UGraphic, style *Style) UGraphic {
	stroke := style.Stroke()
	// Exception for return arrows
	if a.IsDotted() && stroke.DashVisible() == 0 && stroke.DashSpace() == 0 {
		return ug.Apply(NewUStroke(2, 2, stroke.Thickness()))
	}
	return ug.Apply(stroke)
}

// ApplyStrokeLegacy is upstream's deprecated applyStroke(UGraphic): the
// thickness of the configuration, dashed when dotted.
func (a *ArrowConfiguration) ApplyStrokeLegacy(ug UGraphic) UGraphic {
	if a.IsDotted() {
		return ug.Apply(NewUStroke(2, 2, a.thickness))
	}
	return ug.Apply(UStrokeWithThickness(a.thickness))
}

// ApplyThicknessOnly applies the thickness of the configuration.
func (a *ArrowConfiguration) ApplyThicknessOnly(ug UGraphic) UGraphic {
	return ug.Apply(UStrokeWithThickness(a.thickness))
}

// WithThickness returns the arrow with another thickness.
func (a *ArrowConfiguration) WithThickness(thickness float64) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.thickness = thickness })
}

// ReverseDefine returns the arrow with the reverse definition flag toggled.
func (a *ArrowConfiguration) ReverseDefine() *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.reverseDefine = !a.reverseDefine })
}

// IsReverseDefine reports whether the arrow was defined right to left.
func (a *ArrowConfiguration) IsReverseDefine() bool { return a.reverseDefine }

// WithInclination returns the arrow with another inclination.
func (a *ArrowConfiguration) WithInclination(inclination int) *ArrowConfiguration {
	return a.with(func(c *ArrowConfiguration) { c.inclination = inclination })
}

// Inclination1 returns the inclination at the start: the inclination when
// the end has no head or a cross, else 0.
func (a *ArrowConfiguration) Inclination1() int {
	if a.dressing2.Head() == ArrowHeadNone || a.dressing2.Head() == ArrowHeadCrossx {
		return a.inclination
	}
	return 0
}

// Inclination2 returns the inclination at the end: the inclination when
// the start has no head, a cross or a normal head, else 0.
func (a *ArrowConfiguration) Inclination2() int {
	if a.dressing1.Head() == ArrowHeadNone || a.dressing1.Head() == ArrowHeadCrossx {
		return a.inclination
	}
	if a.dressing1.Head() == ArrowHeadNormal {
		return a.inclination
	}
	return 0
}
