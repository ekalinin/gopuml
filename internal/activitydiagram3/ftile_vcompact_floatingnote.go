// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FloatingNote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FloatingNote is a note drawn on its own (upstream uses it in commented-out
// code of FtileGroup and in EBNF diagrams).
type FloatingNote struct {
	klimt.TextBlockDefaults
	opale *klimt.Opale
}

func newFloatingNote(note *klimt.Display, skinParam klimt.ISkinParam, sname klimt.SName, withLink bool) *FloatingNote {
	f := &FloatingNote{}
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, sname, klimt.SNameNote).
		MergedStyle(skinParam.CurrentStyleBuilder())
	wrapWidth := style.WrapWidth()
	fc := klimt.FontConfigurationFromStyle(skinParam, style, nil)
	noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(skinParam.IHtmlColorSet())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	stroke := style.Stroke()
	shadowing := style.Shadowing()
	sheet := skinParam.Sheet(fc, skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft), klimt.CreoleModeFull).
		CreateSheet(note)
	sheetBlock2 := klimt.NewSheetBlock2(klimt.NewSheetBlock1Padding(sheet, wrapWidth, skinParam.Padding()), f, stroke)
	f.opale = klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, sheetBlock2, withLink, &stroke)
	return f
}

// FloatingNoteCreate returns a note without pointer.
func FloatingNoteCreate(note *klimt.Display, skinParam klimt.ISkinParam, sname klimt.SName) *FloatingNote {
	return newFloatingNote(note, skinParam, sname, false)
}

// FloatingNoteCreateOpale returns a note with a pointer.
func FloatingNoteCreateOpale(note *klimt.Display, skinParam klimt.ISkinParam, sname klimt.SName) *FloatingNote {
	return newFloatingNote(note, skinParam, sname, true)
}

// SetOpale sets the side of the pointer and its two points.
func (f *FloatingNote) SetOpale(strategy klimt.Direction, pp1, pp2 geom.XPoint2D) {
	f.opale.SetOpale(strategy, pp1, pp2)
}

// DrawU implements klimt.TextBlock.
func (f *FloatingNote) DrawU(ug klimt.UGraphic) { f.opale.DrawU(ug) }

// CalculateDimension implements klimt.TextBlock.
func (f *FloatingNote) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return f.opale.CalculateDimension(stringBounder)
}

// StartingX implements klimt.Stencil.
func (f *FloatingNote) StartingX(font.StringBounder, float64) float64 {
	return -float64(f.opale.MarginX1())
}

// EndingX implements klimt.Stencil.
func (f *FloatingNote) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return f.opale.CalculateDimension(stringBounder).Width() - float64(f.opale.MarginX1())
}
