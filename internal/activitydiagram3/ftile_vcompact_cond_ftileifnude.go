// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileIfNude.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ftileIfNudeVirtual holds the methods of FtileIfNude that its subclasses
// override (FtileIfWithDiamonds, FtileIfWithLinks).
type ftileIfNudeVirtual interface {
	translateBranch1(stringBounder font.StringBounder) klimt.UTranslate
	translateBranch2(stringBounder font.StringBounder) klimt.UTranslate
	widthInner(stringBounder font.StringBounder) float64
}

// FtileIfNude places the two branches of an "if" side by side.
type FtileIfNude struct {
	FtileDimensionMemoize
	tile1, tile2  Ftile
	in            *Swimlane
	xDeltaNote    float64
	yDeltaNote    float64
	suppWidthNode float64
}

// NewFtileIfNude returns the two branches side by side.
func NewFtileIfNude(tile1, tile2 Ftile, in *Swimlane) *FtileIfNude {
	f := &FtileIfNude{}
	f.initFtileIfNude(f, tile1, tile2, in)
	return f
}

func (f *FtileIfNude) initFtileIfNude(self Ftile, tile1, tile2 Ftile, in *Swimlane) {
	f.initFtile(self, tile1.SkinParam())
	f.tile1 = tile1
	f.tile2 = tile2
	f.in = in
}

func (f *FtileIfNude) virtual() ftileIfNudeVirtual { return f.self.(ftileIfNudeVirtual) }

// MyChildren implements Ftile.
func (f *FtileIfNude) MyChildren() []Ftile { return []Ftile{f.tile1, f.tile2} }

// HasTwoBranches reports whether both branches have an exit point.
func (f *FtileIfNude) HasTwoBranches(stringBounder font.StringBounder) bool {
	return f.tile1.Geometry(stringBounder).HasPointOut() && f.tile2.Geometry(stringBounder).HasPointOut()
}

// Swimlanes implements Swimable.
func (f *FtileIfNude) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if f.SwimlaneIn() != nil {
		result = result.add(f.SwimlaneIn())
	}
	result = result.addAll(f.tile1.Swimlanes())
	return result.addAll(f.tile2.Swimlanes())
}

// SwimlaneIn implements Swimable.
func (f *FtileIfNude) SwimlaneIn() *Swimlane { return f.in }

// SwimlaneOut implements Swimable.
func (f *FtileIfNude) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

func (f *FtileIfNude) translateBranch1(font.StringBounder) klimt.UTranslate {
	return klimt.NewUTranslate(f.xDeltaNote, f.yDeltaNote)
}

func (f *FtileIfNude) translateBranch2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dim2 := f.tile2.CalculateDimension(stringBounder)
	x2 := dimTotal.Width() - dim2.Width() - f.suppWidthNode
	return klimt.NewUTranslate(x2, f.yDeltaNote)
}

// TranslateFor implements Ftile.
func (f *FtileIfNude) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	var t klimt.UTranslate
	switch child {
	case f.tile1:
		t = f.virtual().translateBranch1(stringBounder)
	case f.tile2:
		t = f.virtual().translateBranch2(stringBounder)
	default:
		uml.Throw(uml.UnsupportedOperationException, "")
	}
	return &t
}

// DrawU implements klimt.TextBlock.
func (f *FtileIfNude) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug.Apply(f.virtual().translateBranch1(stringBounder)).Draw(f.tile1)
	ug.Apply(f.virtual().translateBranch2(stringBounder)).Draw(f.tile2)
}

func (f *FtileIfNude) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	if f.tile1.Geometry(stringBounder).HasPointOut() || f.tile2.Geometry(stringBounder).HasPointOut() {
		return dimTotal
	}
	return dimTotal.WithoutPointOut()
}

func (f *FtileIfNude) calculateDimensionInternalSlow(stringBounder font.StringBounder) FtileGeometry {
	dim1 := f.tile1.Geometry(stringBounder)
	dim2 := f.tile2.Geometry(stringBounder)
	innerMargin := f.virtual().widthInner(stringBounder)
	width := f.xDeltaNote + dim1.Left() + innerMargin + (dim2.Width() - dim2.Left()) + f.suppWidthNode
	dim12 := dim1.Dim().MergeLR(dim2.Dim())
	height := f.yDeltaNote + dim12.Height()
	return NewFtileGeometryOut(width, height, f.xDeltaNote+dim1.Left()+innerMargin/2, f.yDeltaNote, height)
}

func (f *FtileIfNude) widthInner(stringBounder font.StringBounder) float64 {
	dim1 := f.tile1.Geometry(stringBounder)
	dim2 := f.tile2.Geometry(stringBounder)
	return (dim1.Width() - dim1.Left()) + dim2.Left()
}
