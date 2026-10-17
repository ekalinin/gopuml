// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/Instruction.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionCollection.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/AbstractInstruction.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/WithNote.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/MonoSwimable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/PositionedNote.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/NotePosition.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/NoteType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Instruction is a node of the instruction tree that the commands build.
// Upstream's createGtile (the other layout, which upstream never selects)
// is not ported.
type Instruction interface {
	Swimable
	// CreateFtile builds the tile of the instruction.
	CreateFtile(factory FtileFactory) Ftile
	// Add appends an instruction to the current part of a composite
	// instruction; simple instructions panic, as upstream.
	Add(other Instruction) command.Result
	// Kill marks the last instruction as killed ("kill", "detach"); it
	// returns false when that is not possible.
	Kill() bool
	// InLinkRendering returns the arrow that leads to the instruction.
	InLinkRendering() *LinkRendering
	// AddNote attaches a note; it returns false when that is not possible.
	AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
		swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool
	// ContainsBreak reports whether a "break" is inside.
	ContainsBreak() bool
}

// InstructionCollection is an instruction that holds other instructions.
type InstructionCollection interface {
	Instruction
	// Last returns the last instruction added, nil for none.
	Last() Instruction
}

// NotePosition is the side of a note, upstream's
// sequencediagram.NotePosition.
type NotePosition int

// Note positions, in upstream's order.
const (
	NoteLeft NotePosition = iota
	NoteRight
	NoteOver
	NoteOverSeveral
	NoteBottom
	NoteTop
)

var notePositionNames = [...]string{"LEFT", "RIGHT", "OVER", "OVER_SEVERAL", "BOTTOM", "TOP"}

// String returns upstream's name, for example "LEFT".
func (p NotePosition) String() string { return notePositionNames[p] }

// notePositionDefaultLeft is NotePosition.defaultLeft: LEFT for ok false,
// else the position named by s.
func notePositionDefaultLeft(s string, ok bool) NotePosition {
	if !ok {
		return NoteLeft
	}
	name := uml.UpperCase(s)
	for i, n := range notePositionNames {
		if n == name {
			return NotePosition(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.sequencediagram.NotePosition."+name)
	return NoteLeft
}

// NoteType tells a note attached to an instruction from a floating note,
// upstream's sequencediagram.NoteType.
type NoteType int

// Note types, in upstream's order.
const (
	NoteTypeNote NoteType = iota
	NoteTypeFloatingNote
)

var noteTypeNames = [...]string{"NOTE", "FLOATING_NOTE"}

// String returns upstream's name, for example "FLOATING_NOTE".
func (t NoteType) String() string { return noteTypeNames[t] }

// noteTypeDefault is NoteType.defaultType: NOTE for ok false, else the type
// named by s.
func noteTypeDefault(s string, ok bool) NoteType {
	if !ok {
		return NoteTypeNote
	}
	name := strings.ReplaceAll(uml.UpperCase(s), " ", "_")
	for i, n := range noteTypeNames {
		if n == name {
			return NoteType(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.sequencediagram.NoteType."+name)
	return NoteTypeNote
}

// PositionedNote is a note with its position, type and lane.
type PositionedNote struct {
	display      *klimt.Display
	notePosition NotePosition
	typ          NoteType
	colors       *color.Colors
	swimlaneNote *Swimlane
	stereotype   *klimt.Stereotype
}

// NewPositionedNote returns a note; colors and stereotype may be nil.
func NewPositionedNote(display *klimt.Display, position NotePosition, typ NoteType, swimlaneNote *Swimlane,
	colors *color.Colors, stereotype *klimt.Stereotype) *PositionedNote {
	return &PositionedNote{display: display, notePosition: position, typ: typ, colors: colors,
		swimlaneNote: swimlaneNote, stereotype: stereotype}
}

// String is upstream's toString.
func (n *PositionedNote) String() string {
	return "type=" + n.typ.String() + " notePosition=" + n.notePosition.String() + " " + n.display.String()
}

// Display returns the text of the note.
func (n *PositionedNote) Display() *klimt.Display { return n.display }

// NotePosition returns the side of the note.
func (n *PositionedNote) NotePosition() NotePosition { return n.notePosition }

// Type returns the type of the note.
func (n *PositionedNote) Type() NoteType { return n.typ }

// Colors returns the colors of the note, nil for none.
func (n *PositionedNote) Colors() *color.Colors { return n.colors }

// SwimlaneNote returns the lane of the note.
func (n *PositionedNote) SwimlaneNote() *Swimlane { return n.swimlaneNote }

// Stereotype returns the stereotype of the note, nil for none.
func (n *PositionedNote) Stereotype() *klimt.Stereotype { return n.stereotype }

// withNote holds the notes of an instruction, upstream's WithNote.
type withNote struct {
	notes []*PositionedNote
}

// AddNote implements Instruction: the note is kept.
func (w *withNote) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	w.notes = append(w.notes, NewPositionedNote(note, position, typ, swimlaneNote, colors, stereotype))
	return true
}

// PositionedNotes returns the notes; the slice must not be modified.
func (w *withNote) PositionedNotes() []*PositionedNote { return w.notes }

// HasNotes reports whether the instruction has notes.
func (w *withNote) HasNotes() bool { return len(w.notes) > 0 }

// monoSwimable is an instruction in a single lane, upstream's
// MonoSwimable.
type monoSwimable struct {
	withNote
	swimlane *Swimlane
}

// Swimlanes implements Swimable: the lane, none for a nil lane.
func (m *monoSwimable) Swimlanes() SwimlaneSet {
	if m.swimlane == nil {
		return nil
	}
	return SwimlaneSet{m.swimlane}
}

// SwimlaneIn implements Swimable.
func (m *monoSwimable) SwimlaneIn() *Swimlane { return m.swimlane }

// SwimlaneOut implements Swimable.
func (m *monoSwimable) SwimlaneOut() *Swimlane { return m.swimlane }

// throwAddUnsupported is the Add of simple instructions.
func throwAddUnsupported() command.Result {
	uml.Throw(uml.UnsupportedOperationException, "")
	return command.Ok()
}

// requireLink panics for a nil link, as upstream's requireNonNull.
func requireLink(l *LinkRendering) *LinkRendering {
	if l == nil {
		uml.ThrowNPE("")
	}
	return l
}

// requireDisplay panics for a nil display, as upstream's requireNonNull
// (the null Display is not nil).
func requireDisplay(d *klimt.Display) *klimt.Display {
	if d == nil {
		uml.ThrowNPE("")
	}
	return d
}
