// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/InstructionSwitch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// InstructionSwitch is "switch" ... "endswitch".
type InstructionSwitch struct {
	withNote
	switches []*Branch

	parent Instruction

	current            *Branch
	topInlinkRendering *LinkRendering
	afterEndwhile      *LinkRendering
	labelTest          *klimt.Display

	swimlane *Swimlane

	colors    *color.Colors
	endColors *color.Colors
}

var _ InstructionCollection = (*InstructionSwitch)(nil)

// NewInstructionSwitch returns a switch without case.
func NewInstructionSwitch(swimlane *Swimlane, parent Instruction, labelTest *klimt.Display,
	inlinkRendering *LinkRendering, colors *color.Colors) *InstructionSwitch {
	return &InstructionSwitch{
		topInlinkRendering: requireLink(inlinkRendering),
		parent:             parent,
		labelTest:          labelTest,
		swimlane:           swimlane,
		colors:             colors,
		afterEndwhile:      LinkRenderingNone(),
	}
}

// ContainsBreak implements Instruction.
func (s *InstructionSwitch) ContainsBreak() bool {
	for _, branch := range s.switches {
		if branch.ContainsBreak() {
			return true
		}
	}
	return false
}

// Add implements Instruction: the instruction goes to the current case.
func (s *InstructionSwitch) Add(ins Instruction) command.Result {
	if s.current == nil {
		return command.Error("No 'case' in this switch")
	}
	return s.current.Add(ins)
}

// Kill implements Instruction: the last instruction of the current case is
// killed. Before the first case it panics, as upstream.
func (s *InstructionSwitch) Kill() bool {
	if s.current == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.activitydiagram3.Branch.kill()" because "this.current" is null`)
	}
	return s.current.Kill()
}

// InLinkRendering implements Instruction.
func (s *InstructionSwitch) InLinkRendering() *LinkRendering { return s.topInlinkRendering }

// Swimlanes implements Swimable.
func (s *InstructionSwitch) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if s.swimlane != nil {
		result = result.add(s.swimlane)
	}
	for _, branch := range s.switches {
		result = result.addAll(branch.Swimlanes())
	}
	return result
}

// SwimlaneIn implements Swimable.
func (s *InstructionSwitch) SwimlaneIn() *Swimlane { return s.swimlane }

// SwimlaneOut implements Swimable.
func (s *InstructionSwitch) SwimlaneOut() *Swimlane { return s.swimlane }

// Last implements InstructionCollection: the last instruction of the last
// case. Before the first case it panics, as upstream.
func (s *InstructionSwitch) Last() Instruction {
	if len(s.switches) == 0 {
		uml.ThrowIndexOutOfBounds(-1, 0)
	}
	return s.switches[len(s.switches)-1].Last()
}

// SwitchCase starts a new case.
func (s *InstructionSwitch) SwitchCase(labelCase *klimt.Display, nextLinkRenderer *LinkRendering) bool {
	if s.current != nil {
		s.current.SetSpecial(nextLinkRenderer, nil)
	}
	s.current = NewBranch(s.swimlane, LinkRenderingNone().WithDisplay(labelCase), labelCase, nil,
		LinkRenderingNone().WithDisplay(labelCase), nil)
	s.switches = append(s.switches, s.current)
	return true
}

// Parent returns the instruction that holds the switch.
func (s *InstructionSwitch) Parent() Instruction { return s.parent }

// EndSwitch closes the switch; endColors are the colors of the merge point.
func (s *InstructionSwitch) EndSwitch(nextLinkRenderer *LinkRendering, endColors *color.Colors) {
	s.endColors = endColors
	if s.current != nil {
		s.current.SetSpecial(nextLinkRenderer, endColors)
	}
}

// AddNote implements Instruction: the note goes to the current case unless
// it is empty.
func (s *InstructionSwitch) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if s.current == nil || s.current.IsEmpty() {
		return s.withNote.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	return s.current.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Switches returns the cases.
func (s *InstructionSwitch) Switches() []*Branch { return s.switches }

// Current returns the case that receives the instructions, nil before the
// first case.
func (s *InstructionSwitch) Current() *Branch { return s.current }

// AfterEndwhile returns the arrow that leaves the switch.
func (s *InstructionSwitch) AfterEndwhile() *LinkRendering { return s.afterEndwhile }

// LabelTest returns the test of the switch.
func (s *InstructionSwitch) LabelTest() *klimt.Display { return s.labelTest }

// Swimlane returns the lane of the switch.
func (s *InstructionSwitch) Swimlane() *Swimlane { return s.swimlane }

// Colors returns the colors of the switch diamond.
func (s *InstructionSwitch) Colors() *color.Colors { return s.colors }

// EndColors returns the colors of the merge point, nil before "endswitch".
func (s *InstructionSwitch) EndColors() *color.Colors { return s.endColors }
