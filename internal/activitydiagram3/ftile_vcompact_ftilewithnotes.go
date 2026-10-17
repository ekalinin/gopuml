// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileWithNotes.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileWithNotes is a tile with notes stacked on its left and right,
// without pointers.
type FtileWithNotes struct {
	AbstractFtile
	tile              Ftile
	left              klimt.TextBlock
	right             klimt.TextBlock
	verticalAlignment geom.VerticalAlignment
}

// StyleSignature returns the style signature of notes.
func (f *FtileWithNotes) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameNote)
}

// Swimlanes implements Swimable.
func (f *FtileWithNotes) Swimlanes() SwimlaneSet { return f.tile.Swimlanes() }

// SwimlaneIn implements Swimable.
func (f *FtileWithNotes) SwimlaneIn() *Swimlane { return f.tile.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileWithNotes) SwimlaneOut() *Swimlane { return f.tile.SwimlaneOut() }

// NewFtileWithNotes returns tile with notes.
func NewFtileWithNotes(tile Ftile, notes []*PositionedNote, verticalAlignment geom.VerticalAlignment) *FtileWithNotes {
	f := &FtileWithNotes{verticalAlignment: verticalAlignment, tile: tile}
	f.initFtile(f, tile.SkinParam())
	for _, note := range notes {
		style := f.StyleSignature().MergedStyle(f.SkinParam().CurrentStyleBuilder()).
			EventuallyOverrideColors(note.Colors())
		noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
		borderColor := style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
		fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
		shadowing := style.Shadowing()
		wrapWidth := style.WrapWidth()
		stroke := style.Stroke()
		sheet := f.SkinParam().Sheet(fc, f.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft),
			klimt.CreoleModeFull).CreateSheet(note.Display())
		sheet1 := klimt.NewSheetBlock1Padding(sheet, wrapWidth, f.SkinParam().Padding())
		// -6 and 15 come from Opale, as upstream
		sheet2 := klimt.NewSheetBlock2(sheet1, klimt.StencilFunc{
			Starting: func(font.StringBounder, float64) float64 { return -6 },
			Ending: func(stringBounder font.StringBounder, y float64) float64 {
				return sheet1.EndingX(stringBounder, y) + 15
			},
		}, klimt.UStrokeSimple())
		opale := klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, sheet2, false, &stroke)
		opaleMarged := klimt.TextBlockUtilsWithMargin(opale, 10, 10)
		if note.NotePosition() == NoteLeft {
			if f.left == nil {
				f.left = opaleMarged
			} else {
				f.left = klimt.TextBlockUtilsMergeTB(f.left, opaleMarged, geom.HorizontalAlignmentCenter)
			}
		} else {
			if f.right == nil {
				f.right = opaleMarged
			} else {
				f.right = klimt.TextBlockUtilsMergeTB(f.right, opaleMarged, geom.HorizontalAlignmentCenter)
			}
		}
	}
	if f.left == nil {
		f.left = klimt.TextBlockUtilsEmpty(0, 0)
	}
	if f.right == nil {
		f.right = klimt.TextBlockUtilsEmpty(0, 0)
	}
	return f
}

func (f *FtileWithNotes) translate(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimTile := f.tile.CalculateDimension(stringBounder)
	xDelta := f.left.CalculateDimension(stringBounder).Width()
	yDelta := 0.0
	if f.verticalAlignment != geom.VerticalAlignmentTop {
		yDelta = (dimTotal.Height() - dimTile.Height()) / 2
	}
	return klimt.NewUTranslate(xDelta, yDelta)
}

func (f *FtileWithNotes) translateForLeft(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimLeft := f.left.CalculateDimension(stringBounder)
	yDelta := 0.0
	if f.verticalAlignment != geom.VerticalAlignmentTop {
		yDelta = (dimTotal.Height() - dimLeft.Height()) / 2
	}
	return klimt.NewUTranslate(0, yDelta)
}

func (f *FtileWithNotes) translateForRight(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimRight := f.right.CalculateDimension(stringBounder)
	xDelta := dimTotal.Width() - dimRight.Width()
	yDelta := 0.0
	if f.verticalAlignment != geom.VerticalAlignmentTop {
		yDelta = (dimTotal.Height() - dimRight.Height()) / 2
	}
	return klimt.NewUTranslate(xDelta, yDelta)
}

// DrawU implements klimt.TextBlock.
func (f *FtileWithNotes) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	f.left.DrawU(ug.Apply(f.translateForLeft(stringBounder)))
	f.right.DrawU(ug.Apply(f.translateForRight(stringBounder)))
	ug.Apply(f.translate(stringBounder)).Draw(f.tile)
}

func (f *FtileWithNotes) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	orig := f.tile.Geometry(stringBounder)
	translate := f.translate(stringBounder)
	if orig.HasPointOut() {
		return NewFtileGeometryDimOut(dimTotal, orig.Left()+translate.Dx(), orig.InY()+translate.Dy(),
			orig.OutY()+translate.Dy())
	}
	return NewFtileGeometryDim(dimTotal, orig.Left()+translate.Dx(), orig.InY()+translate.Dy())
}

func (f *FtileWithNotes) calculateDimensionInternal(stringBounder font.StringBounder) geom.XDimension2D {
	dimTile := f.tile.CalculateDimension(stringBounder)
	dimLeft := f.left.CalculateDimension(stringBounder)
	dimRight := f.right.CalculateDimension(stringBounder)
	height := max(dimLeft.Height(), dimRight.Height(), dimTile.Height())
	return geom.NewXDimension2D(dimTile.Width()+dimLeft.Width()+dimRight.Width(), height)
}

// InLinkRendering implements Ftile.
func (f *FtileWithNotes) InLinkRendering() *LinkRendering { return f.tile.InLinkRendering() }
