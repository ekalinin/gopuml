// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionFork.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSplit.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ForkStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// ForkStyle is how the branches of a parallel block join.
type ForkStyle int

// Fork styles, in upstream's order.
const (
	ForkStyleFork ForkStyle = iota
	ForkStyleSplit
	ForkStyleMerge
)

var forkStyleNames = [...]string{"FORK", "SPLIT", "MERGE"}

// String returns upstream's name, for example "MERGE".
func (s ForkStyle) String() string { return forkStyleNames[s] }

// InstructionFork is "fork" ... "end fork" (or "end merge").
type InstructionFork struct {
	withNote
	forks           []*InstructionList
	parent          Instruction
	inlinkRendering *LinkRendering
	swimlaneIn      *Swimlane
	swimlaneOut     *Swimlane
	style           ForkStyle
	label           string
	hasLabel        bool
	finished        bool
	colors          *color.Colors
}

// NewInstructionFork returns a fork with one empty branch.
func NewInstructionFork(parent Instruction, inlinkRendering *LinkRendering, swimlane *Swimlane,
	colors *color.Colors) *InstructionFork {
	return &InstructionFork{
		parent:          parent,
		colors:          colors,
		inlinkRendering: requireLink(inlinkRendering),
		swimlaneIn:      swimlane,
		swimlaneOut:     swimlane,
		style:           ForkStyleFork,
		forks:           []*InstructionList{InstructionListEmpty()},
	}
}

// ContainsBreak implements Instruction.
func (f *InstructionFork) ContainsBreak() bool {
	for _, fork := range f.forks {
		if fork.ContainsBreak() {
			return true
		}
	}
	return false
}

func (f *InstructionFork) lastList() *InstructionList { return f.forks[len(f.forks)-1] }

// Add implements Instruction: the instruction goes to the last branch.
func (f *InstructionFork) Add(ins Instruction) command.Result { return f.lastList().Add(ins) }

// Parent returns the instruction that holds the fork.
func (f *InstructionFork) Parent() Instruction { return f.parent }

// ForkAgain starts a new branch, in swimlane.
func (f *InstructionFork) ForkAgain(swimlane *Swimlane) {
	f.swimlaneOut = swimlane
	f.forks = append(f.forks, InstructionListEmpty())
}

// Kill implements Instruction: the last instruction of the last branch is
// killed.
func (f *InstructionFork) Kill() bool { return f.lastList().Kill() }

// InLinkRendering implements Instruction.
func (f *InstructionFork) InLinkRendering() *LinkRendering { return f.inlinkRendering }

// AddNote implements Instruction: after "end fork" the note goes to the
// fork, else to the last branch.
func (f *InstructionFork) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if f.finished {
		return f.withNote.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	return f.lastList().AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable: the lanes of the branches, the incoming
// and the outgoing lanes (which may be nil).
func (f *InstructionFork) Swimlanes() SwimlaneSet {
	result := SwimlaneSet(nil).addAll(swimlanes2(f.forks))
	result = result.add(f.swimlaneIn)
	return result.add(f.swimlaneOut)
}

// SwimlaneIn implements Swimable.
func (f *InstructionFork) SwimlaneIn() *Swimlane { return f.swimlaneIn }

// SwimlaneOut implements Swimable.
func (f *InstructionFork) SwimlaneOut() *Swimlane { return f.swimlaneOut }

// ManageOutRendering sets the arrow that leaves the last branch; endFork
// marks the end of the fork.
func (f *InstructionFork) ManageOutRendering(nextLinkRenderer *LinkRendering, endFork bool) {
	if endFork {
		f.finished = true
	}
	if nextLinkRenderer == nil {
		return
	}
	f.lastList().SetOutRendering(nextLinkRenderer)
}

// SetStyle sets how the branches join, the label of the bar ("{label}",
// ok false for none) and the outgoing lane.
func (f *InstructionFork) SetStyle(style ForkStyle, label string, ok bool, swimlane *Swimlane) {
	f.style = style
	f.label = label
	f.hasLabel = ok
	f.swimlaneOut = swimlane
}

// Forks returns the branches.
func (f *InstructionFork) Forks() []*InstructionList { return f.forks }

// Style returns how the branches join.
func (f *InstructionFork) Style() ForkStyle { return f.style }

// Label returns the label of the bar; ok is false for none.
func (f *InstructionFork) Label() (string, bool) { return f.label, f.hasLabel }

// IsFinished reports whether "end fork" was read.
func (f *InstructionFork) IsFinished() bool { return f.finished }

// Colors returns the colors of the bars.
func (f *InstructionFork) Colors() *color.Colors { return f.colors }

// InstructionSplit is "split" ... "end split".
type InstructionSplit struct {
	splits          []*InstructionList
	parent          Instruction
	inlinkRendering *LinkRendering
	swimlaneIn      *Swimlane
	colors          *color.Colors
	swimlaneOut     *Swimlane
}

// NewInstructionSplit returns a split with one empty branch; colors may be
// nil.
func NewInstructionSplit(parent Instruction, inlinkRendering *LinkRendering, swimlane *Swimlane,
	colors *color.Colors) *InstructionSplit {
	return &InstructionSplit{
		parent:          parent,
		swimlaneIn:      swimlane,
		colors:          colors,
		splits:          []*InstructionList{NewInstructionList(swimlane)},
		inlinkRendering: requireLink(inlinkRendering),
	}
}

// ContainsBreak implements Instruction.
func (s *InstructionSplit) ContainsBreak() bool {
	for _, split := range s.splits {
		if split.ContainsBreak() {
			return true
		}
	}
	return false
}

func (s *InstructionSplit) last() *InstructionList { return s.splits[len(s.splits)-1] }

// Add implements Instruction: the instruction goes to the last branch.
func (s *InstructionSplit) Add(ins Instruction) command.Result { return s.last().Add(ins) }

// Parent returns the instruction that holds the split.
func (s *InstructionSplit) Parent() Instruction { return s.parent }

// SplitAgain starts a new branch; inlinkRendering, when not nil, leaves the
// previous one.
func (s *InstructionSplit) SplitAgain(inlinkRendering *LinkRendering) {
	if inlinkRendering != nil {
		s.last().SetOutRendering(inlinkRendering)
	}
	s.splits = append(s.splits, NewInstructionList(s.swimlaneIn))
}

// EndSplit closes the split; inlinkRendering, when not nil, leaves the last
// branch.
func (s *InstructionSplit) EndSplit(inlinkRendering *LinkRendering, endSwimlane *Swimlane) {
	if inlinkRendering != nil {
		s.last().SetOutRendering(inlinkRendering)
	}
	s.swimlaneOut = endSwimlane
}

// Kill implements Instruction: the last instruction of the last branch is
// killed.
func (s *InstructionSplit) Kill() bool { return s.last().Kill() }

// InLinkRendering implements Instruction.
func (s *InstructionSplit) InLinkRendering() *LinkRendering { return s.inlinkRendering }

// AddNote implements Instruction: the note goes to the last branch.
func (s *InstructionSplit) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	return s.last().AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable.
func (s *InstructionSplit) Swimlanes() SwimlaneSet { return swimlanes2(s.splits) }

// SwimlaneIn implements Swimable: the outgoing lane of the parent.
func (s *InstructionSplit) SwimlaneIn() *Swimlane { return s.parent.SwimlaneOut() }

// SwimlaneOut implements Swimable: the lane of "end split".
func (s *InstructionSplit) SwimlaneOut() *Swimlane { return s.swimlaneOut }

// Splits returns the branches.
func (s *InstructionSplit) Splits() []*InstructionList { return s.splits }

// SplitSwimlaneIn returns the lane of "split".
func (s *InstructionSplit) SplitSwimlaneIn() *Swimlane { return s.swimlaneIn }

// Colors returns the colors of the split, nil for none.
func (s *InstructionSplit) Colors() *color.Colors { return s.colors }
