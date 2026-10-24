// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileSwitchNude.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// FtileSwitchNude places the cases of a switch side by side.
type FtileSwitchNude struct {
	FtileDimensionMemoize
	xSeparation float64
	tiles       []Ftile
	in          *Swimlane
}

// NewFtileSwitchNude returns the cases tiles side by side.
func NewFtileSwitchNude(tiles []Ftile, in *Swimlane) *FtileSwitchNude {
	f := &FtileSwitchNude{}
	f.initFtileSwitchNude(f, tiles, in)
	return f
}

func (f *FtileSwitchNude) initFtileSwitchNude(self Ftile, tiles []Ftile, in *Swimlane) {
	f.initFtile(self, tiles[0].SkinParam())
	f.xSeparation = 20
	f.tiles = tiles
	f.in = in
}

// MyChildren implements Ftile.
func (f *FtileSwitchNude) MyChildren() []Ftile { return f.tiles }

// Swimlanes implements Swimable.
func (f *FtileSwitchNude) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if f.SwimlaneIn() != nil {
		result = result.add(f.SwimlaneIn())
	}
	for _, tile := range f.tiles {
		result = result.addAll(tile.Swimlanes())
	}
	return result
}

// SwimlaneIn implements Swimable.
func (f *FtileSwitchNude) SwimlaneIn() *Swimlane { return f.in }

// SwimlaneOut implements Swimable.
func (f *FtileSwitchNude) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

// TranslateFor implements Ftile: only the cases have a position.
func (f *FtileSwitchNude) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if slices.Contains(f.tiles, child) {
		t := f.translateNude(child, stringBounder)
		return &t
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (f *FtileSwitchNude) translateNude(tile Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	x1 := 0.0
	for _, candidate := range f.tiles {
		dim1 := candidate.Geometry(stringBounder)
		if candidate == tile {
			return klimt.UTranslateDx(x1)
		}
		x1 += dim1.Width() + f.xSeparation
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}

// DrawU implements klimt.TextBlock.
func (f *FtileSwitchNude) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	for _, tile := range f.tiles {
		ug.Apply(f.translateNude(tile, stringBounder)).Draw(tile)
	}
}

func (f *FtileSwitchNude) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder)
	for _, tile := range f.tiles {
		if tile.Geometry(stringBounder).HasPointOut() {
			return dimTotal
		}
	}
	return dimTotal.WithoutPointOut()
}

func (f *FtileSwitchNude) calculateDimensionInternalSlow(stringBounder font.StringBounder) FtileGeometry {
	result := geom.NewXDimension2D(0, 0)
	for _, couple := range f.tiles {
		result = result.MergeLR(couple.CalculateDimension(stringBounder))
	}
	result = result.DeltaWH(float64(f.xSeparation*float64(len(f.tiles)-1)), 100)
	return NewFtileGeometry(result.Width(), result.Height(), result.Width()/2, 0)
}
