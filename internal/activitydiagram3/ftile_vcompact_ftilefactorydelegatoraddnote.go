// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorAddNote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// FtileFactoryDelegatorAddNote attaches notes to tiles.
type FtileFactoryDelegatorAddNote struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorAddNote returns the delegator over factory.
func NewFtileFactoryDelegatorAddNote(factory FtileFactory) *FtileFactoryDelegatorAddNote {
	return &FtileFactoryDelegatorAddNote{newFtileFactoryDelegator(factory)}
}

// AddNote implements FtileFactory: a nil ftile (notes on an empty list)
// gives the first note alone.
func (d *FtileFactoryDelegatorAddNote) AddNote(ftile Ftile, swimlane *Swimlane, notes []*PositionedNote,
	verticalAlignment geom.VerticalAlignment) Ftile {
	if len(notes) == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	skinParam := d.SkinParam()
	if ftile == nil {
		note := notes[0]
		if note.Colors() != nil {
			skinParam = klimt.ColorsMute(note.Colors(), skinParam)
		}
		return NewFtileNoteAlone(skinParam.Shadowing(nil), note.Display(), skinParam, note.Type() == NoteTypeNote,
			swimlane)
	}
	return FtileWithNoteOpaleCreate(ftile, notes, true, verticalAlignment)
}
