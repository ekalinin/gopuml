// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Note.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Notes.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/NotePosition.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/NoteStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/NoteType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// NotePosition is the position of a note.
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

// notePositionOf is NotePosition.valueOf of an upper-cased keyword.
func notePositionOf(s string) NotePosition {
	for i, n := range notePositionNames {
		if n == s {
			return NotePosition(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.sequencediagram.NotePosition."+s)
	return 0
}

// NoteStyle is the shape of a note.
type NoteStyle int

// Note styles, in upstream's order.
const (
	NoteStyleNormal NoteStyle = iota
	NoteStyleHexagonal
	NoteStyleBox
)

var noteStyleNames = [...]string{"NORMAL", "HEXAGONAL", "BOX"}

// String returns upstream's name, for example "HEXAGONAL".
func (s NoteStyle) String() string { return noteStyleNames[s] }

// noteStyleOf is NoteStyle.getNoteStyle: "hnote" and "rnote" give their
// shapes, anything else the normal note.
func noteStyleOf(s string) NoteStyle {
	switch {
	case uml.EqualsIgnoreCase(s, "hnote"):
		return NoteStyleHexagonal
	case uml.EqualsIgnoreCase(s, "rnote"):
		return NoteStyleBox
	}
	return NoteStyleNormal
}

// ComponentType returns the drawing component of the shape.
func (s NoteStyle) ComponentType() ComponentType {
	switch s {
	case NoteStyleHexagonal:
		return ComponentNoteHexagonal
	case NoteStyleBox:
		return ComponentNoteBox
	}
	return ComponentNote
}

// DefaultStyleDefinition returns the style signature of the shape.
func (s NoteStyle) DefaultStyleDefinition() klimt.StyleSignatureBasic {
	switch s {
	case NoteStyleHexagonal:
		return seqSignature(klimt.SNameNote, klimt.SNameHnote)
	case NoteStyleBox:
		return seqSignature(klimt.SNameNote, klimt.SNameRnote)
	}
	return seqSignature(klimt.SNameNote)
}

// Note is a note of a sequence diagram.
type Note struct {
	abstractEvent
	p, p2        *Participant
	strings      *klimt.Display
	position     NotePosition
	styleBuilder *klimt.StyleBuilder
	noteStyle    NoteStyle
	colors       *color.Colors
	// shadowing is the shadowing that Colors.applyStereotypeForNote sets,
	// nil when unset.
	shadowing  *bool
	url        *klimt.URL
	stereotype *klimt.Stereotype
	parallel   bool
}

// newNote returns a note; p and p2 may be nil.
func newNote(p, p2 *Participant, position NotePosition, strings *klimt.Display,
	styleBuilder *klimt.StyleBuilder) *Note {
	return &Note{p: p, p2: p2, styleBuilder: styleBuilder, position: position, strings: strings,
		colors: color.EmptyColors()}
}

// TemporaryProtectedUntilTeozIsStandard turns top and bottom notes into
// left notes.
func (n *Note) TemporaryProtectedUntilTeozIsStandard() {
	if n.position == NoteBottom || n.position == NoteTop {
		n.position = NoteLeft
	}
}

// StyleSignature returns the style signature of the note.
func (n *Note) StyleSignature() klimt.StyleSignatureBasic {
	return n.noteStyle.DefaultStyleDefinition()
}

// UsedStyles returns the style of the note.
func (n *Note) UsedStyles() []*klimt.Style {
	signature := n.StyleSignature()
	if n.stereotype != nil {
		signature = signature.MergeWithStyles(n.stereotype.Styles(n.styleBuilder))
	}
	return []*klimt.Style{signature.MergedStyle(n.styleBuilder).EventuallyOverrideColors(n.colors)}
}

// SetStereotype sets the stereotype of the note.
func (n *Note) SetStereotype(stereotype *klimt.Stereotype) { n.stereotype = stereotype }

// Stereotype returns the stereotype of the note, nil for none.
func (n *Note) Stereotype() *klimt.Stereotype { return n.stereotype }

// WithPosition returns the note at another position.
func (n *Note) WithPosition(position NotePosition) *Note {
	if n.position == position {
		return n
	}
	result := newNote(n.p, n.p2, position, n.strings, n.styleBuilder)
	result.noteStyle = n.noteStyle
	result.url = n.url
	result.colors = n.colors
	result.shadowing = n.shadowing
	result.parallel = n.parallel
	return result
}

// Participant returns the first participant of the note, nil for none.
func (n *Note) Participant() *Participant { return n.p }

// Participant2 returns the second participant of a note over several, nil
// for none.
func (n *Note) Participant2() *Participant { return n.p2 }

// Display returns the text of the note.
func (n *Note) Display() *klimt.Display { return n.strings }

// Position returns the position of the note.
func (n *Note) Position() NotePosition { return n.position }

// Colors returns the specific colors of the note.
func (n *Note) Colors() *color.Colors { return n.colors }

// SetColors sets the specific colors of the note.
func (n *Note) SetColors(colors *color.Colors) { n.colors = colors }

// Shadowing returns the shadowing of the note's stereotype; ok is false when
// unset.
func (n *Note) Shadowing() (shadowing, ok bool) {
	if n.shadowing == nil {
		return false, false
	}
	return *n.shadowing, true
}

// DealWith implements Event.
func (n *Note) DealWith(someone *Participant) bool { return n.p == someone || n.p2 == someone }

// URL implements Event.
func (n *Note) URL() *klimt.URL { return n.url }

// HasURL implements Event.
func (n *Note) HasURL() bool { return n.url != nil }

// SetURL sets the link of the note.
func (n *Note) SetURL(url *klimt.URL) { n.url = url }

// NoteStyle returns the shape of the note.
func (n *Note) NoteStyle() NoteStyle { return n.noteStyle }

// SetNoteStyle sets the shape of the note.
func (n *Note) SetNoteStyle(style NoteStyle) { n.noteStyle = style }

// GoParallel marks the note as parallel to the previous event.
func (n *Note) GoParallel() { n.parallel = true }

// IsParallel implements Event.
func (n *Note) IsParallel() bool { return n.parallel }

// Notes is a group of notes side by side ("/ note").
type Notes struct {
	abstractEvent
	notes []*Note
}

// Add adds a note to the group.
func (n *Notes) Add(note *Note) { n.notes = append(n.notes, note) }

// DealWith implements Event.
func (n *Notes) DealWith(someone *Participant) bool {
	for _, note := range n.notes {
		if note.DealWith(someone) {
			return true
		}
	}
	return false
}

// List returns the notes of the group.
func (n *Notes) List() []*Note { return n.notes }
