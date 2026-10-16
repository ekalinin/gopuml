// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionGroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileGroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbols.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// GroupSymbol is the shape of a "partition", "package", "rectangle",
// "card" or "group" block: upstream's USymbols constants that the activity
// diagrams use. The layout step draws them with the decoration symbols.
type GroupSymbol int

// Group symbols.
const (
	SymbolPartition GroupSymbol = iota
	SymbolPackage
	SymbolRectangle
	SymbolCard
	SymbolGroup
)

var groupSymbolNames = [...]string{"PARTITION", "PACKAGE", "RECTANGLE", "CARD", "GROUP"}

var groupSymbolSNames = [...]klimt.SName{klimt.SNamePartition, klimt.SNamePackage, klimt.SNameRectangle,
	klimt.SNameCard, klimt.SNameGroup}

// String returns the name of upstream's USymbols constant, for example
// "PARTITION".
func (s GroupSymbol) String() string { return groupSymbolNames[s] }

// SNames returns the style names of the symbol.
func (s GroupSymbol) SNames() []klimt.SName { return []klimt.SName{groupSymbolSNames[s]} }

// groupStyleSignature is FtileGroup.getStyleSignature(USymbol).
func groupStyleSignature(symbol GroupSymbol) klimt.StyleSignatureBasic {
	names := []klimt.SName{klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram}
	names = append(names, symbol.SNames()...)
	return klimt.StyleSignatureBasicOf(append(names, klimt.SNameComposite)...)
}

// InstructionGroup is a "partition", "group" or other block, closed by "}".
type InstructionGroup struct {
	list          *InstructionList
	parent        Instruction
	backColor     color.HColor
	linkRendering *LinkRendering
	typ           GroupSymbol

	title *klimt.Display
	note  *PositionedNote
	style *klimt.Style
}

var _ InstructionCollection = (*InstructionGroup)(nil)

// NewInstructionGroup returns an empty block of shape typ; backColor may be
// nil.
func NewInstructionGroup(parent Instruction, title *klimt.Display, backColor color.HColor, swimlane *Swimlane,
	linkRendering *LinkRendering, typ GroupSymbol, style *klimt.Style) *InstructionGroup {
	return &InstructionGroup{list: NewInstructionList(swimlane), typ: typ, linkRendering: linkRendering,
		parent: parent, title: title, style: style, backColor: backColor}
}

// ContainsBreak implements Instruction.
func (g *InstructionGroup) ContainsBreak() bool { return g.list.ContainsBreak() }

// Add implements Instruction.
func (g *InstructionGroup) Add(ins Instruction) command.Result { return g.list.Add(ins) }

// Parent returns the instruction that holds the block.
func (g *InstructionGroup) Parent() Instruction { return g.parent }

// Kill implements Instruction: the last instruction is killed.
func (g *InstructionGroup) Kill() bool { return g.list.Kill() }

// InLinkRendering implements Instruction.
func (g *InstructionGroup) InLinkRendering() *LinkRendering { return g.linkRendering }

// AddNote implements Instruction: a note on an empty block replaces the
// note of the block, else it goes to the last instruction.
func (g *InstructionGroup) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if g.list.IsEmpty() {
		g.note = NewPositionedNote(note, position, typ, swimlaneNote, colors, stereotype)
		return true
	}
	return g.list.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable.
func (g *InstructionGroup) Swimlanes() SwimlaneSet { return g.list.Swimlanes() }

// SwimlaneIn implements Swimable.
func (g *InstructionGroup) SwimlaneIn() *Swimlane { return g.list.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (g *InstructionGroup) SwimlaneOut() *Swimlane { return g.list.SwimlaneOut() }

// Last implements InstructionCollection.
func (g *InstructionGroup) Last() Instruction { return g.list.Last() }

// List returns the instructions of the block.
func (g *InstructionGroup) List() *InstructionList { return g.list }

// BackColor returns the background of the block, nil for none.
func (g *InstructionGroup) BackColor() color.HColor { return g.backColor }

// Type returns the shape of the block.
func (g *InstructionGroup) Type() GroupSymbol { return g.typ }

// Title returns the title of the block.
func (g *InstructionGroup) Title() *klimt.Display { return g.title }

// Note returns the note of the block, nil for none.
func (g *InstructionGroup) Note() *PositionedNote { return g.note }

// Style returns the style of the block.
func (g *InstructionGroup) Style() *klimt.Style { return g.style }
