// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileWithNoteOpale.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

const ftileWithNoteOpaleSuppSpace = 20.0

// FtileWithNoteOpale is a tile with a note on its left or right, linked to
// it by the pointer of the note.
type FtileWithNoteOpale struct {
	AbstractFtile
	tile              Ftile
	opale             *klimt.Opale
	verticalAlignment geom.VerticalAlignment
	notePosition      NotePosition
	swimlaneNote      *Swimlane
}

// StyleSignature returns the style signature of notes.
func (f *FtileWithNoteOpale) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameNote)
}

// Swimlanes implements Swimable: the lanes of the tile and the lane of the
// note.
func (f *FtileWithNoteOpale) Swimlanes() SwimlaneSet {
	if f.swimlaneNote != nil {
		return SwimlaneSet(nil).addAll(f.tile.Swimlanes()).add(f.swimlaneNote)
	}
	return f.tile.Swimlanes()
}

// SwimlaneIn implements Swimable.
func (f *FtileWithNoteOpale) SwimlaneIn() *Swimlane { return f.tile.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileWithNoteOpale) SwimlaneOut() *Swimlane { return f.tile.SwimlaneOut() }

// MyChildren implements Ftile.
func (f *FtileWithNoteOpale) MyChildren() []Ftile { return []Ftile{f.tile} }

// FtileWithNoteOpaleCreate returns tile with its notes: several notes give
// FtileWithNotes, a single one FtileWithNoteOpale.
func FtileWithNoteOpaleCreate(tile Ftile, notes []*PositionedNote, withLink bool,
	verticalAlignment geom.VerticalAlignment) Ftile {
	if len(notes) > 1 {
		return NewFtileWithNotes(tile, notes, verticalAlignment)
	}
	if len(notes) == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return newFtileWithNoteOpale(tile, notes[0], withLink, verticalAlignment)
}

func newFtileWithNoteOpale(tile Ftile, note *PositionedNote, withLink bool,
	verticalAlignment geom.VerticalAlignment) *FtileWithNoteOpale {
	f := &FtileWithNoteOpale{verticalAlignment: verticalAlignment, swimlaneNote: note.SwimlaneNote(), tile: tile,
		notePosition: note.NotePosition()}
	skinParam := tile.SkinParam()
	if note.Colors() != nil {
		skinParam = klimt.ColorsMute(note.Colors(), tile.SkinParam())
	}
	f.initFtile(f, skinParam)
	if note.Type() == NoteTypeFloatingNote {
		withLink = false
	}
	stereotype := note.Stereotype()
	style := f.StyleSignature().WithTOBECHANGED(stereotype).MergedStyle(f.SkinParam().CurrentStyleBuilder()).
		EventuallyOverrideColors(note.Colors())
	noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
	fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
	shadowing := style.Shadowing()
	wrapWidth := style.WrapWidth()
	stroke := style.Stroke()
	align := f.SkinParam().HorizontalAlignment(klimt.AlignmentParamNoteTextAlignment, klimt.ArrowDirectionNull, false,
		klimt.HorizontalAlignmentNull)
	sheet := f.SkinParam().Sheet(fc, align, klimt.CreoleModeFull).CreateSheet(note.Display())
	text := klimt.NewSheetBlock2(klimt.NewSheetBlock1Padding(sheet, wrapWidth, f.SkinParam().Padding()), f, stroke)
	f.opale = klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, text, withLink, &stroke)
	return f
}

func (f *FtileWithNoteOpale) translate(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimNote := f.opale.CalculateDimension(stringBounder)
	dimTile := f.tile.CalculateDimension(stringBounder)
	yForFtile := (dimTotal.Height() - dimTile.Height()) / 2
	marge := 0.0
	if f.notePosition == NoteLeft {
		marge = dimNote.Width() + ftileWithNoteOpaleSuppSpace
	}
	return klimt.NewUTranslate(marge, yForFtile)
}

// TranslateFor implements Ftile: none for the tile, as upstream (although
// the tile is drawn moved).
func (f *FtileWithNoteOpale) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == f.tile {
		t := klimt.UTranslateNone()
		return &t
	}
	return f.AbstractFtile.TranslateFor(child, stringBounder)
}

func (f *FtileWithNoteOpale) translateForOpale(ug klimt.UGraphic) klimt.UTranslate {
	stringBounder := ug.StringBounder()
	dimTotal := f.CalculateDimension(stringBounder)
	dimNote := f.opale.CalculateDimension(stringBounder)
	yForNote := 0.0
	if f.verticalAlignment == geom.VerticalAlignmentCenter {
		yForNote = (dimTotal.Height() - dimNote.Height()) / 2
	}
	if f.notePosition == NoteLeft {
		return klimt.UTranslateDy(yForNote)
	}
	dx := dimTotal.Width() - dimNote.Width()
	return klimt.NewUTranslate(dx, yForNote)
}

// DrawU implements klimt.TextBlock: in a lane drawing, the note is drawn
// only in its lane.
func (f *FtileWithNoteOpale) DrawU(ug klimt.UGraphic) {
	oneSwimlane, isOneSwimlane := ug.(*UGraphicInterceptorOneSwimlane)
	var intoSw *Swimlane
	if isOneSwimlane {
		intoSw = oneSwimlane.Swimlane()
	}
	stringBounder := ug.StringBounder()
	dimNote := f.opale.CalculateDimension(stringBounder)
	if f.notePosition == NoteLeft {
		pp1 := geom.NewXPoint2D(dimNote.Width(), dimNote.Height()/2)
		pp2 := geom.NewXPoint2D(dimNote.Width()+ftileWithNoteOpaleSuppSpace, dimNote.Height()/2)
		f.opale.SetOpale(klimt.DirectionRight, pp1, pp2)
	} else {
		pp1 := geom.NewXPoint2D(0, dimNote.Height()/2)
		pp2 := geom.NewXPoint2D(-ftileWithNoteOpaleSuppSpace, dimNote.Height()/2)
		f.opale.SetOpale(klimt.DirectionLeft, pp1, pp2)
	}
	if !isOneSwimlane || f.swimlaneNote == nil || intoSw == f.swimlaneNote {
		f.opale.DrawU(ug.Apply(f.translateForOpale(ug)))
	}
	ug.Apply(f.translate(stringBounder)).Draw(f.tile)
}

func (f *FtileWithNoteOpale) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	orig := f.tile.Geometry(stringBounder)
	translate := f.translate(stringBounder)
	if orig.HasPointOut() {
		return NewFtileGeometryDimOut(dimTotal, orig.Left()+translate.Dx(), orig.InY()+translate.Dy(),
			orig.OutY()+translate.Dy())
	}
	return NewFtileGeometryDim(dimTotal, orig.Left()+translate.Dx(), orig.InY()+translate.Dy())
}

func (f *FtileWithNoteOpale) calculateDimensionInternal(stringBounder font.StringBounder) geom.XDimension2D {
	dimNote := f.opale.CalculateDimension(stringBounder)
	dimTile := f.tile.CalculateDimension(stringBounder)
	height := max(dimNote.Height(), dimTile.Height())
	return geom.NewXDimension2D(dimTile.Width()+1*dimNote.Width()+ftileWithNoteOpaleSuppSpace, height)
}

// StartingX implements klimt.Stencil.
func (f *FtileWithNoteOpale) StartingX(font.StringBounder, float64) float64 {
	return -float64(f.opale.MarginX1())
}

// EndingX implements klimt.Stencil.
func (f *FtileWithNoteOpale) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return f.opale.CalculateDimension(stringBounder).Width() - float64(f.opale.MarginX1())
}

// InLinkRendering implements Ftile.
func (f *FtileWithNoteOpale) InLinkRendering() *LinkRendering { return f.tile.InLinkRendering() }
