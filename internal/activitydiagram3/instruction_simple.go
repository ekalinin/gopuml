// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSpot.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionGoto.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionLabel.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionStart.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionStop.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionEnd.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionBreak.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// InstructionSimple is an action, ":text;".
type InstructionSimple struct {
	monoSwimable
	killed          bool
	label           *klimt.Display
	colors          *color.Colors
	inlinkRendering *LinkRendering
	boxStyle        BoxStyle
	url             *klimt.URL
	stereotype      *klimt.Stereotype
	styleBuilder    *klimt.StyleBuilder
}

// NewInstructionSimple returns an action; url and stereotype may be nil.
func NewInstructionSimple(label *klimt.Display, inlinkRendering *LinkRendering, swimlane *Swimlane,
	boxStyle BoxStyle, url *klimt.URL, colors *color.Colors, stereotype *klimt.Stereotype,
	styleBuilder *klimt.StyleBuilder) *InstructionSimple {
	if colors == nil {
		uml.ThrowNPE("")
	}
	return &InstructionSimple{monoSwimable: monoSwimable{swimlane: swimlane}, stereotype: stereotype, url: url,
		boxStyle: boxStyle, label: label, inlinkRendering: requireLink(inlinkRendering), colors: colors,
		styleBuilder: styleBuilder}
}

// ContainsBreak implements Instruction.
func (*InstructionSimple) ContainsBreak() bool { return false }

// Add implements Instruction: it panics.
func (*InstructionSimple) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: the action is killed.
func (i *InstructionSimple) Kill() bool {
	i.killed = true
	return true
}

// InLinkRendering implements Instruction.
func (i *InstructionSimple) InLinkRendering() *LinkRendering { return i.inlinkRendering }

// IsKilled reports whether the action is killed.
func (i *InstructionSimple) IsKilled() bool { return i.killed }

// Label returns the text of the action.
func (i *InstructionSimple) Label() *klimt.Display { return i.label }

// Colors returns the colors of the action.
func (i *InstructionSimple) Colors() *color.Colors { return i.colors }

// BoxStyle returns the shape of the action.
func (i *InstructionSimple) BoxStyle() BoxStyle { return i.boxStyle }

// URL returns the link of the action, nil for none.
func (i *InstructionSimple) URL() *klimt.URL { return i.url }

// Stereotype returns the stereotype of the action, nil for none.
func (i *InstructionSimple) Stereotype() *klimt.Stereotype { return i.stereotype }

// StyleBuilder returns the style builder of the action.
func (i *InstructionSimple) StyleBuilder() *klimt.StyleBuilder { return i.styleBuilder }

// InstructionSpot is a connector, "(A)".
type InstructionSpot struct {
	monoSwimable
	killed          bool
	inlinkRendering *LinkRendering
	spot            string
	color           color.HColor
}

// NewInstructionSpot returns a connector; c may be nil.
func NewInstructionSpot(spot string, c color.HColor, inlinkRendering *LinkRendering,
	swimlane *Swimlane) *InstructionSpot {
	return &InstructionSpot{monoSwimable: monoSwimable{swimlane: swimlane}, spot: spot,
		inlinkRendering: requireLink(inlinkRendering), color: c}
}

// ContainsBreak implements Instruction.
func (*InstructionSpot) ContainsBreak() bool { return false }

// Add implements Instruction: it panics.
func (*InstructionSpot) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: the connector is killed.
func (i *InstructionSpot) Kill() bool {
	i.killed = true
	return true
}

// InLinkRendering implements Instruction.
func (i *InstructionSpot) InLinkRendering() *LinkRendering { return i.inlinkRendering }

// IsKilled reports whether the connector is killed.
func (i *InstructionSpot) IsKilled() bool { return i.killed }

// Spot returns the letter of the connector.
func (i *InstructionSpot) Spot() string { return i.spot }

// Color returns the color of the connector, nil for none.
func (i *InstructionSpot) Color() color.HColor { return i.color }

// InstructionGoto is "goto name".
type InstructionGoto struct {
	monoSwimable
	name string
}

// NewInstructionGoto returns a jump to the label name.
func NewInstructionGoto(swimlane *Swimlane, name string) *InstructionGoto {
	return &InstructionGoto{monoSwimable: monoSwimable{swimlane: swimlane}, name: name}
}

// Add implements Instruction: it panics.
func (*InstructionGoto) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: a jump cannot be killed.
func (*InstructionGoto) Kill() bool { return false }

// InLinkRendering implements Instruction: no arrow.
func (*InstructionGoto) InLinkRendering() *LinkRendering { return LinkRenderingNone() }

// ContainsBreak implements Instruction.
func (*InstructionGoto) ContainsBreak() bool { return false }

// Name returns the label of the jump.
func (i *InstructionGoto) Name() string { return i.name }

// InstructionLabel is "label name".
type InstructionLabel struct {
	monoSwimable
	name string
}

// NewInstructionLabel returns the label name.
func NewInstructionLabel(swimlane *Swimlane, name string) *InstructionLabel {
	return &InstructionLabel{monoSwimable: monoSwimable{swimlane: swimlane}, name: name}
}

// Add implements Instruction: it panics.
func (*InstructionLabel) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: a label cannot be killed.
func (*InstructionLabel) Kill() bool { return false }

// InLinkRendering implements Instruction: no arrow.
func (*InstructionLabel) InLinkRendering() *LinkRendering { return LinkRenderingNone() }

// ContainsBreak implements Instruction.
func (*InstructionLabel) ContainsBreak() bool { return false }

// Name returns the name of the label.
func (i *InstructionLabel) Name() string { return i.name }

// terminal is the state of "start", "stop" and "end".
type terminal struct {
	monoSwimable
	inlinkRendering *LinkRendering
	colors          *color.Colors
}

func newTerminal(colors *color.Colors, swimlane *Swimlane, inlinkRendering *LinkRendering) terminal {
	return terminal{monoSwimable: monoSwimable{swimlane: swimlane}, colors: colors,
		inlinkRendering: requireLink(inlinkRendering)}
}

// ContainsBreak implements Instruction.
func (*terminal) ContainsBreak() bool { return false }

// Add implements Instruction: it panics.
func (*terminal) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: a terminal cannot be killed.
func (*terminal) Kill() bool { return false }

// InLinkRendering implements Instruction.
func (t *terminal) InLinkRendering() *LinkRendering { return t.inlinkRendering }

// Colors returns the colors of the terminal.
func (t *terminal) Colors() *color.Colors { return t.colors }

// InstructionStart is "start".
type InstructionStart struct{ terminal }

// NewInstructionStart returns a start node.
func NewInstructionStart(colors *color.Colors, swimlane *Swimlane, inlinkRendering *LinkRendering) *InstructionStart {
	return &InstructionStart{newTerminal(colors, swimlane, inlinkRendering)}
}

// InstructionStop is "stop".
type InstructionStop struct{ terminal }

// NewInstructionStop returns a stop node.
func NewInstructionStop(colors *color.Colors, swimlane *Swimlane, inlinkRendering *LinkRendering) *InstructionStop {
	return &InstructionStop{newTerminal(colors, swimlane, inlinkRendering)}
}

// InstructionEnd is "end".
type InstructionEnd struct{ terminal }

// NewInstructionEnd returns an end node.
func NewInstructionEnd(colors *color.Colors, swimlane *Swimlane, inlinkRendering *LinkRendering) *InstructionEnd {
	return &InstructionEnd{newTerminal(colors, swimlane, inlinkRendering)}
}

// InstructionBreak is "break".
type InstructionBreak struct {
	monoSwimable
	inlinkRendering *LinkRendering
}

// NewInstructionBreak returns a break.
func NewInstructionBreak(swimlane *Swimlane, inlinkRendering *LinkRendering) *InstructionBreak {
	return &InstructionBreak{monoSwimable: monoSwimable{swimlane: swimlane},
		inlinkRendering: requireLink(inlinkRendering)}
}

// Add implements Instruction: it panics.
func (*InstructionBreak) Add(Instruction) command.Result { return throwAddUnsupported() }

// Kill implements Instruction: a break cannot be killed.
func (*InstructionBreak) Kill() bool { return false }

// InLinkRendering implements Instruction.
func (i *InstructionBreak) InLinkRendering() *LinkRendering { return i.inlinkRendering }

// ContainsBreak implements Instruction: true.
func (*InstructionBreak) ContainsBreak() bool { return true }
