// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileNoteAlone.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileNoteAlone is a note without tile, the notes of an empty list.
type FtileNoteAlone struct {
	AbstractFtile
	opale        *klimt.Opale
	withOutPoint bool
	swimlane     *Swimlane
}

// StyleSignature returns the style signature of notes.
func (f *FtileNoteAlone) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameNote)
}

// MyChildren implements Ftile: none.
func (f *FtileNoteAlone) MyChildren() []Ftile { return nil }

// Swimlanes implements Swimable.
func (f *FtileNoteAlone) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *FtileNoteAlone) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileNoteAlone) SwimlaneOut() *Swimlane { return f.swimlane }

// NewFtileNoteAlone returns the note; withOutPoint gives it an exit point.
// shadow is unused, as upstream.
func NewFtileNoteAlone(shadow bool, note *klimt.Display, skinParam klimt.ISkinParam, withOutPoint bool,
	swimlane *Swimlane) *FtileNoteAlone {
	f := &FtileNoteAlone{swimlane: swimlane, withOutPoint: withOutPoint}
	f.initFtile(f, skinParam)
	style := f.StyleSignature().MergedStyle(skinParam.CurrentStyleBuilder())
	noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
	shadowing := style.Shadowing()
	wrapWidth := style.WrapWidth()
	stroke := style.Stroke()
	fc := klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamNote, nil)
	sheet := skinParam.Sheet(fc, skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft), klimt.CreoleModeFull).
		CreateSheet(note)
	text := klimt.NewSheetBlock2(klimt.NewSheetBlock1Padding(sheet, wrapWidth, skinParam.Padding()), f, stroke)
	f.opale = klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, text, false, &stroke)
	return f
}

// DrawU implements klimt.TextBlock.
func (f *FtileNoteAlone) DrawU(ug klimt.UGraphic) { f.opale.DrawU(ug) }

func (f *FtileNoteAlone) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.opale.CalculateDimension(stringBounder)
	if f.withOutPoint {
		return NewFtileGeometryDimOut(dimTotal, dimTotal.Width()/2, 0, dimTotal.Height())
	}
	return NewFtileGeometryDim(dimTotal, dimTotal.Width()/2, 0)
}

// StartingX implements klimt.Stencil.
func (f *FtileNoteAlone) StartingX(font.StringBounder, float64) float64 {
	return -float64(f.opale.MarginX1())
}

// EndingX implements klimt.Stencil.
func (f *FtileNoteAlone) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return f.opale.CalculateDimension(stringBounder).Width() - float64(f.opale.MarginX1())
}
