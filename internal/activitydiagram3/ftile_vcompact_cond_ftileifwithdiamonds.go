// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileIfWithDiamonds.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const ftileIfWithDiamondsSuppWidth = 20

// ftileIfWithDiamondsVirtual holds the method of FtileIfWithDiamonds that
// FtileIfWithLinks overrides.
type ftileIfWithDiamondsVirtual interface {
	yDeltaForLabels(stringBounder font.StringBounder) float64
}

// FtileIfWithDiamonds is an "if" with its diamonds above and below the
// branches, and the notes of the test beside the first diamond.
type FtileIfWithDiamonds struct {
	FtileIfNude
	diamond1, diamond2 Ftile
	opaleLeft          klimt.TextBlock
	opaleRight         klimt.TextBlock
}

// NewFtileIfWithDiamonds returns the "if" of the diamonds and branches.
func NewFtileIfWithDiamonds(diamond1, tile1, tile2, diamond2 Ftile, in *Swimlane, stringBounder font.StringBounder,
	notes []*PositionedNote) *FtileIfWithDiamonds {
	f := &FtileIfWithDiamonds{}
	f.initFtileIfWithDiamonds(f, diamond1, tile1, tile2, diamond2, in, stringBounder, notes)
	return f
}

func (f *FtileIfWithDiamonds) initFtileIfWithDiamonds(self Ftile, diamond1, tile1, tile2, diamond2 Ftile,
	in *Swimlane, stringBounder font.StringBounder, notes []*PositionedNote) {
	f.initFtileIfNude(self, tile1, tile2, in)
	f.diamond1 = diamond1
	f.diamond2 = diamond2
	f.opaleLeft = klimt.TextBlockUtilsEmptyTextBlock()
	f.opaleRight = klimt.TextBlockUtilsEmptyTextBlock()
	for _, first := range notes {
		switch first.NotePosition() {
		case NoteLeft:
			if f.opaleLeft != klimt.TextBlockUtilsEmptyTextBlock() {
				continue
			}
			f.opaleLeft = FtileIfWithDiamondsCreateOpale(first, f.SkinParam())
			pos1 := f.translateDiamond1(stringBounder).Dx()
			dimOpale := f.opaleLeft.CalculateDimension(stringBounder)
			opaleWith := dimOpale.Width()
			if opaleWith > pos1 {
				f.xDeltaNote = opaleWith - pos1
			}
			f.yDeltaNote = max(f.yDeltaNote, dimOpale.Height())
		case NoteRight:
			if f.opaleRight != klimt.TextBlockUtilsEmptyTextBlock() {
				continue
			}
			f.opaleRight = FtileIfWithDiamondsCreateOpale(first, f.SkinParam())
			dimOpale := f.opaleRight.CalculateDimension(stringBounder)
			pos1 := f.translateDiamond1(stringBounder).Dx() + f.diamond1.CalculateDimension(stringBounder).Width() +
				dimOpale.Width()
			pos2 := f.self.(ftileDimensionMemoizeHook).calculateDimensionInternalSlow(stringBounder).Width()
			if pos1 > pos2 {
				f.suppWidthNode = pos1 - pos2
			}
			f.yDeltaNote = max(f.yDeltaNote, dimOpale.Height())
		}
		f.clearCacheDimensionInternal()
	}
}

// FtileIfWithDiamondsCreateOpale returns the note of a test.
func FtileIfWithDiamondsCreateOpale(first *PositionedNote, skinParam klimt.ISkinParam) *klimt.Opale {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameNote).MergedStyle(skinParam.CurrentStyleBuilder()).EventuallyOverrideColors(first.Colors())
	noteBackgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(skinParam.IHtmlColorSet())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	fc := style.FontConfiguration(skinParam.IHtmlColorSet(), nil)
	shadowing := style.Shadowing()
	wrapWidth := style.WrapWidth()
	stroke := style.Stroke()
	align := skinParam.HorizontalAlignment(klimt.AlignmentParamNoteTextAlignment, klimt.ArrowDirectionNull, false,
		klimt.HorizontalAlignmentNull)
	sheet := skinParam.Sheet(fc, align, klimt.CreoleModeFull).CreateSheet(first.Display())
	tmp1 := klimt.NewSheetBlock1Padding(sheet, wrapWidth, skinParam.Padding())
	tmp2 := klimt.NewSheetBlock2(tmp1, klimt.StencilFunc{
		Starting: func(stringBounder font.StringBounder, y float64) float64 {
			return tmp1.StartingX(stringBounder, y) - klimt.OpaleMarginX1
		},
		Ending: func(stringBounder font.StringBounder, y float64) float64 {
			return tmp1.EndingX(stringBounder, y) + klimt.OpaleMarginX2
		},
	}, stroke)
	return klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, tmp2, false, &stroke)
}

// MyChildren implements Ftile.
func (f *FtileIfWithDiamonds) MyChildren() []Ftile {
	return []Ftile{f.diamond1, f.diamond2, f.tile1, f.tile2}
}

func (f *FtileIfWithDiamonds) yDelta1a(font.StringBounder) int {
	if len(f.Swimlanes()) > 1 {
		return 20
	}
	return 10
}

func (f *FtileIfWithDiamonds) yDelta1b(stringBounder font.StringBounder) int {
	if len(f.Swimlanes()) > 1 {
		return 10
	}
	if f.HasTwoBranches(stringBounder) {
		return 6
	}
	return 0
}

func (f *FtileIfWithDiamonds) yDeltaForLabels(font.StringBounder) float64 { return 0 }

func (f *FtileIfWithDiamonds) widthInner(stringBounder font.StringBounder) float64 {
	dim1 := f.diamond1.Geometry(stringBounder)
	return max(f.FtileIfNude.widthInner(stringBounder), dim1.Width()+ftileIfWithDiamondsSuppWidth)
}

func (f *FtileIfWithDiamonds) calculateDimensionInternalSlow(stringBounder font.StringBounder) FtileGeometry {
	dim1 := f.diamond1.Geometry(stringBounder)
	dim2 := f.diamond2.Geometry(stringBounder)
	dimNude := f.FtileIfNude.calculateDimensionInternalSlow(stringBounder)
	all := dim1.AppendBottom(dimNude).AppendBottom(dim2)
	deltaHeight := float64(f.yDelta1a(stringBounder)+f.yDelta1b(stringBounder)) +
		f.self.(ftileIfWithDiamondsVirtual).yDeltaForLabels(stringBounder)
	return all.AddDim(0, deltaHeight).IncInY(f.yDeltaNote)
}

// DrawU implements klimt.TextBlock.
func (f *FtileIfWithDiamonds) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	if !klimt.TextBlockUtilsIsEmpty(f.opaleLeft, stringBounder) {
		xOpale := f.translateDiamond1(stringBounder).Dx() - f.opaleLeft.CalculateDimension(stringBounder).Width()
		f.opaleLeft.DrawU(ug.Apply(klimt.UTranslateDx(xOpale)))
	}
	if !klimt.TextBlockUtilsIsEmpty(f.opaleRight, stringBounder) {
		xOpale := f.translateDiamond1(stringBounder).Dx() + f.diamond1.CalculateDimension(stringBounder).Width()
		f.opaleRight.DrawU(ug.Apply(klimt.UTranslateDx(xOpale)))
	}
	ug.Apply(f.translateDiamond1(stringBounder)).Draw(f.diamond1)
	f.FtileIfNude.DrawU(ug)
	ug.Apply(f.translateDiamond2(stringBounder)).Draw(f.diamond2)
}

func (f *FtileIfWithDiamonds) translateBranch1(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	return f.FtileIfNude.translateBranch1(stringBounder).Compose(
		klimt.UTranslateDy(dimDiamond1.Height() + float64(f.yDelta1a(stringBounder))))
}

func (f *FtileIfWithDiamonds) translateBranch2(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	return f.FtileIfNude.translateBranch2(stringBounder).Compose(
		klimt.UTranslateDy(dimDiamond1.Height() + float64(f.yDelta1a(stringBounder))))
}

func (f *FtileIfWithDiamonds) translateDiamond1(stringBounder font.StringBounder) klimt.UTranslate {
	y1 := f.yDeltaNote
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	x1 := dimTotal.Left() - dimDiamond1.Left()
	return klimt.NewUTranslate(x1, y1)
}

func (f *FtileIfWithDiamonds) translateDiamond2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimDiamond2 := f.diamond2.Geometry(stringBounder)
	x2 := dimTotal.Left() - dimDiamond2.Width()/2
	y2 := dimTotal.Height() - dimDiamond2.Height()
	return klimt.NewUTranslate(x2, y2)
}

// ComputeMarginNeedForBranchLabe1 returns the room the label of the first
// branch needs on the left.
func (f *FtileIfWithDiamonds) ComputeMarginNeedForBranchLabe1(stringBounder font.StringBounder,
	label1 geom.XDimension2D) float64 {
	widthLabelBranch1 := label1.Width()
	dxDiamond := f.translateDiamond1(stringBounder).Dx()
	if diff := widthLabelBranch1 - dxDiamond; diff > 0 {
		return diff
	}
	return 0
}

// ComputeMarginNeedForBranchLabe2 returns the room the label of the second
// branch needs on the right.
func (f *FtileIfWithDiamonds) ComputeMarginNeedForBranchLabe2(stringBounder font.StringBounder,
	label2 geom.XDimension2D) float64 {
	widthLabelBranch2 := label2.Width()
	theoricalEndNeeded := f.translateDiamond1(stringBounder).Dx() +
		f.diamond1.CalculateDimension(stringBounder).Width() + widthLabelBranch2
	if diff := theoricalEndNeeded - f.self.CalculateDimension(stringBounder).Width(); diff > 0 {
		return diff
	}
	return 0
}

// ComputeVerticalMarginNeedForBranchs returns the room the labels of the
// branches need above.
func (f *FtileIfWithDiamonds) ComputeVerticalMarginNeedForBranchs(stringBounder font.StringBounder, label1,
	label2 geom.XDimension2D) float64 {
	heightLabels := max(label1.Height(), label2.Height())
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	dyDiamond := dimDiamond1.Height()
	if diff := heightLabels - dyDiamond; diff > 0 {
		return diff
	}
	return 0
}
