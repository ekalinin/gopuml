// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/InstructionList.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// InstructionList is a sequence of instructions.
type InstructionList struct {
	withNote
	all              []Instruction
	defaultSwimlane  *Swimlane
	outlinkRendering *LinkRendering
}

var _ InstructionCollection = (*InstructionList)(nil)

// InstructionListEmpty returns an empty list without default lane.
func InstructionListEmpty() *InstructionList { return NewInstructionList(nil) }

// NewInstructionList returns an empty list whose lane is defaultSwimlane.
func NewInstructionList(defaultSwimlane *Swimlane) *InstructionList {
	return &InstructionList{defaultSwimlane: defaultSwimlane}
}

// ContainsBreak implements Instruction.
func (l *InstructionList) ContainsBreak() bool {
	for _, ins := range l.all {
		if ins.ContainsBreak() {
			return true
		}
	}
	return false
}

// IsEmpty reports whether the list has no instruction.
func (l *InstructionList) IsEmpty() bool { return len(l.all) == 0 }

// All returns the instructions; the slice must not be modified.
func (l *InstructionList) All() []Instruction { return l.all }

// IsOnlySingleStopOrSpot reports whether the list is a single killed
// action, a connector, or a stop or end without notes.
func (l *InstructionList) IsOnlySingleStopOrSpot() bool {
	if len(l.all) != 1 {
		return false
	}
	switch last := l.Last().(type) {
	case *InstructionSimple:
		return last.IsKilled()
	case *InstructionSpot:
		return true
	case *InstructionStop:
		return !last.HasNotes()
	case *InstructionEnd:
		return !last.HasNotes()
	}
	return false
}

// Add implements Instruction.
func (l *InstructionList) Add(ins Instruction) command.Result {
	l.all = append(l.all, ins)
	return command.Ok()
}

// Kill implements Instruction: the last instruction is killed.
func (l *InstructionList) Kill() bool {
	if len(l.all) == 0 {
		return false
	}
	return l.Last().Kill()
}

// InLinkRendering implements Instruction: the arrow of the first
// instruction. An empty list panics, as upstream's iterator.
func (l *InstructionList) InLinkRendering() *LinkRendering {
	if len(l.all) == 0 {
		uml.Throw(uml.NoSuchElementException, "")
	}
	return l.all[0].InLinkRendering()
}

// Last implements InstructionCollection.
func (l *InstructionList) Last() Instruction {
	if len(l.all) == 0 {
		return nil
	}
	return l.all[len(l.all)-1]
}

// AddNote implements Instruction: the note goes to the last instruction,
// or to the list when it is empty.
func (l *InstructionList) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if l.Last() == nil {
		return l.withNote.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	return l.Last().AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable: the lanes of the instructions.
func (l *InstructionList) Swimlanes() SwimlaneSet { return swimlanes2(l.all) }

// SwimlaneIn implements Swimable: the default lane.
func (l *InstructionList) SwimlaneIn() *Swimlane { return l.defaultSwimlane }

// SwimlaneOut implements Swimable.
func (l *InstructionList) SwimlaneOut() *Swimlane {
	swimlanes := l.Swimlanes()
	switch len(swimlanes) {
	case 0:
		return nil
	case 1:
		return swimlanes[0]
	}
	return l.Last().SwimlaneOut()
}

// swimlanes2 is InstructionList.getSwimlanes2: the lanes of instructions.
func swimlanes2[I Instruction](list []I) SwimlaneSet {
	var result SwimlaneSet
	for _, ins := range list {
		result = result.addAll(ins.Swimlanes())
	}
	return result
}

// SetOutRendering sets the arrow that leaves the list; nil for none.
func (l *InstructionList) SetOutRendering(outlinkRendering *LinkRendering) {
	l.outlinkRendering = outlinkRendering
}

// OutRendering returns the arrow that leaves the list, nil for none.
func (l *InstructionList) OutRendering() *LinkRendering { return l.outlinkRendering }

// DefaultSwimlane returns the lane of the list.
func (l *InstructionList) DefaultSwimlane() *Swimlane { return l.defaultSwimlane }
