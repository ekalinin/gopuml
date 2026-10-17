// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileSwitchWithDiamonds.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

const ftileSwitchWithDiamondsSupp15 = 15.0

// ftileSwitchWithDiamondsMode tells whether the switch diamond is wider
// than the inner cases.
type ftileSwitchWithDiamondsMode int

const (
	ftileSwitchWithDiamondsBigDiamond ftileSwitchWithDiamondsMode = iota
	ftileSwitchWithDiamondsSmallDiamond
)

// ftileSwitchWithDiamondsYdeltaHook is upstream's overridable
// getYdelta1a.
type ftileSwitchWithDiamondsYdeltaHook interface {
	ydelta1a(stringBounder font.StringBounder) float64
}

// FtileSwitchWithDiamonds is a switch with its test diamond above the
// cases and its merge diamond below.
type FtileSwitchWithDiamonds struct {
	FtileSwitchNude
	diamond1 Ftile
	diamond2 Ftile
	branches []*Branch
	mode     ftileSwitchWithDiamondsMode
	w13      float64
	w9       float64
}

// NewFtileSwitchWithDiamonds returns the switch of tiles between two
// diamonds.
func NewFtileSwitchWithDiamonds(tiles []Ftile, branches []*Branch, in *Swimlane, diamond1, diamond2 Ftile,
	stringBounder font.StringBounder) *FtileSwitchWithDiamonds {
	f := &FtileSwitchWithDiamonds{}
	f.initFtileSwitchWithDiamonds(f, tiles, branches, in, diamond1, diamond2, stringBounder)
	return f
}

func (f *FtileSwitchWithDiamonds) initFtileSwitchWithDiamonds(self Ftile, tiles []Ftile, branches []*Branch,
	in *Swimlane, diamond1, diamond2 Ftile, stringBounder font.StringBounder) {
	f.initFtileSwitchNude(self, tiles, in)
	f.branches = branches
	f.diamond1 = diamond1
	f.diamond2 = diamond2
	f.w13 = diamond1.CalculateDimension(stringBounder).Width() - tiles[0].Geometry(stringBounder).Right() -
		tiles[len(tiles)-1].Geometry(stringBounder).Left()
	f.w9 = f.computeW9(stringBounder)
	if f.w13 > f.w9 {
		f.mode = ftileSwitchWithDiamondsBigDiamond
	} else {
		f.mode = ftileSwitchWithDiamondsSmallDiamond
	}
}

func (f *FtileSwitchWithDiamonds) computeW9(stringBounder font.StringBounder) float64 {
	result := 0.0
	for i := 1; i < len(f.tiles)-1; i++ {
		result += f.tiles[i].CalculateDimension(stringBounder).Width()
	}
	return result
}

// MyChildren implements Ftile: the cases and the two diamonds.
func (f *FtileSwitchWithDiamonds) MyChildren() []Ftile {
	result := append([]Ftile(nil), f.FtileSwitchNude.MyChildren()...)
	return append(result, f.diamond1, f.diamond2)
}

func (f *FtileSwitchWithDiamonds) ydelta1a(font.StringBounder) float64 { return 20 }

func (f *FtileSwitchWithDiamonds) ydelta1aVirtual(stringBounder font.StringBounder) float64 {
	return f.self.(ftileSwitchWithDiamondsYdeltaHook).ydelta1a(stringBounder)
}

func (f *FtileSwitchWithDiamonds) ydelta1b(font.StringBounder) float64 { return 10 }

func (f *FtileSwitchWithDiamonds) calculateDimensionInternalSlow(stringBounder font.StringBounder) FtileGeometry {
	dim1 := f.diamond1.Geometry(stringBounder)
	dim2 := f.diamond2.Geometry(stringBounder)
	nudeHeight := f.FtileSwitchNude.calculateDimensionInternalSlow(stringBounder).Height()
	if f.mode == ftileSwitchWithDiamondsBigDiamond {
		height := dim1.Height() + nudeHeight + dim2.Height() + f.ydelta1aVirtual(stringBounder) +
			f.ydelta1b(stringBounder)
		tile0 := f.tiles[0].Geometry(stringBounder)
		width := tile0.Width() + ftileSwitchWithDiamondsSupp15 + f.w13 + ftileSwitchWithDiamondsSupp15 +
			f.tiles[len(f.tiles)-1].Geometry(stringBounder).Width()
		return NewFtileGeometryOut(width, height, tile0.Left()+ftileSwitchWithDiamondsSupp15+dim1.Left(), 0, height)
	}
	dimNude := f.FtileSwitchNude.calculateDimensionInternalSlow(stringBounder)
	all := dim1.AppendBottom(dimNude).AppendBottom(dim2)
	return all.AddDim(0, f.ydelta1aVirtual(stringBounder)+f.ydelta1b(stringBounder))
}

// DrawU implements klimt.TextBlock.
func (f *FtileSwitchWithDiamonds) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug.Apply(f.translateDiamond1(stringBounder)).Draw(f.diamond1)
	if f.mode == ftileSwitchWithDiamondsBigDiamond {
		for _, tile := range f.tiles {
			tile.DrawU(ug.Apply(f.translateOf(tile, stringBounder)))
		}
	} else {
		f.FtileSwitchNude.DrawU(ug.Apply(f.translateMain(stringBounder)))
	}
	if f.Geometry(stringBounder).HasPointOut() {
		ug.Apply(f.translateDiamond2(stringBounder)).Draw(f.diamond2)
	}
}

func (f *FtileSwitchWithDiamonds) translateOf(tile Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	main := f.translateMain(stringBounder)
	if f.mode == ftileSwitchWithDiamondsBigDiamond {
		dx := 0.0
		suppx := (f.w13 - f.w9) / float64(len(f.tiles)-1)
		for i := 0; i < len(f.tiles)-1; i++ {
			if tile == f.tiles[i] {
				return main.Compose(klimt.UTranslateDx(dx))
			}
			dx += f.tiles[i].CalculateDimension(stringBounder).Width() + suppx
		}
		if tile == f.tiles[len(f.tiles)-1] {
			dx9 := f.tiles[0].CalculateDimension(stringBounder).Width() + f.w13 + ftileSwitchWithDiamondsSupp15 +
				ftileSwitchWithDiamondsSupp15
			return main.Compose(klimt.UTranslateDx(dx9))
		}
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return f.translateNude(tile, stringBounder).Compose(main)
}

func (f *FtileSwitchWithDiamonds) translateMain(stringBounder font.StringBounder) klimt.UTranslate {
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	dy1 := dimDiamond1.Height() + f.ydelta1aVirtual(stringBounder)
	return klimt.UTranslateDy(dy1)
}

func (f *FtileSwitchWithDiamonds) translateDiamond1(stringBounder font.StringBounder) klimt.UTranslate {
	y1 := 0.0
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimDiamond1 := f.diamond1.Geometry(stringBounder)
	x1 := dimTotal.Left() - dimDiamond1.Left()
	return klimt.NewUTranslate(x1, y1)
}

func (f *FtileSwitchWithDiamonds) translateDiamond2(stringBounder font.StringBounder) klimt.UTranslate {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	dimDiamond2 := f.diamond2.Geometry(stringBounder)
	y2 := dimTotal.Height() - dimDiamond2.Height()
	x2 := dimTotal.Left() - dimDiamond2.Width()/2
	return klimt.NewUTranslate(x2, y2)
}
