// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileSwitch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

const ftileSwitchXSeparation = 20.0

// FtileSwitch is upstream's first layout of a switch, which upstream no
// longer uses (FtileFactoryDelegatorSwitch builds FtileSwitchWithOneLink
// and FtileSwitchWithManyLinks).
type FtileSwitch struct {
	AbstractFtile
	tiles      []Ftile
	arrowColor *Rainbow
}

func newFtileSwitch(_ []float64, tiles []Ftile, arrowColor *Rainbow) *FtileSwitch {
	f := &FtileSwitch{tiles: slices.Clone(tiles), arrowColor: arrowColor}
	f.initFtile(f, tiles[0].SkinParam())
	return f
}

// Swimlanes implements Swimable.
func (f *FtileSwitch) Swimlanes() SwimlaneSet { return singleSwimlane(f.SwimlaneIn()) }

// SwimlaneIn implements Swimable.
func (f *FtileSwitch) SwimlaneIn() *Swimlane { return f.tiles[0].SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileSwitch) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

// FtileSwitchCreate is upstream's FtileSwitch.create.
func FtileSwitchCreate(swimlane *Swimlane, borderColor, backColor color.HColor, arrowColor *Rainbow,
	ftileFactory FtileFactory, _ klimt.ConditionStyle, thens []*Branch, fcArrow *klimt.FontConfiguration,
	_, afterEndwhile *LinkRendering, fcTest *klimt.FontConfiguration) Ftile {
	requireLink(afterEndwhile)
	var tiles []Ftile
	for _, branch := range thens {
		tiles = append(tiles, NewFtileMinWidthCentered(branch.Ftile(), 30))
	}
	var inlabelSizes []float64
	for _, branch := range thens {
		tb1 := branch.DisplayPositive().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
		tbTest := branch.LabelTest().Create(fcTest,
			ftileFactory.SkinParam().DefaultTextAlignment(geom.HorizontalAlignmentLeft), ftileFactory.SkinParam())
		diamondColor := branch.Color()
		if diamondColor == nil {
			diamondColor = backColor
		}
		diamond := NewFtileDiamondInside2(tbTest, branch.SkinParam(), diamondColor, borderColor, swimlane)
		if klimt.DisplayIsNull(branch.Inlabel()) {
			inlabelSizes = append(inlabelSizes, 0.0)
		} else {
			tbInlabel := branch.Inlabel().Create(fcArrow, geom.HorizontalAlignmentLeft, ftileFactory.SkinParam())
			inlabelSizes = append(inlabelSizes, tbInlabel.CalculateDimension(ftileFactory.StringBounder()).Width())
			diamond = diamond.WithWest(tbInlabel)
		}
		diamond.WithNorth(tb1)
	}
	return newFtileSwitch(inlabelSizes, tiles, arrowColor)
}

// MyChildren implements Ftile.
func (f *FtileSwitch) MyChildren() []Ftile { return slices.Clone(f.tiles) }

// TranslateFor implements Ftile.
func (f *FtileSwitch) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if slices.Contains(f.tiles, child) {
		t := f.translate1(child, stringBounder)
		return &t
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (f *FtileSwitch) translate1(tile Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	x1 := 0.0
	for _, candidate := range f.tiles {
		dim1 := candidate.Geometry(stringBounder)
		if candidate == tile {
			return klimt.NewUTranslate(x1, 25)
		}
		x1 += dim1.Width() + ftileSwitchXSeparation
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}

// DrawU implements klimt.TextBlock.
func (f *FtileSwitch) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	for _, tile := range f.tiles {
		ug.Apply(f.translate1(tile, stringBounder)).Draw(tile)
	}
}

func (f *FtileSwitch) calculateDimensionInternal(stringBounder font.StringBounder) FtileGeometry {
	result := geom.NewXDimension2D(0, 0)
	for _, couple := range f.tiles {
		result = result.MergeLR(couple.CalculateDimension(stringBounder))
	}
	result = result.DeltaWH(float64(ftileSwitchXSeparation*float64(len(f.tiles)-1)), 100)
	return NewFtileGeometryDim(result, result.Width()/2, 0)
}

func (f *FtileSwitch) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimTotal := f.calculateDimensionInternal(stringBounder).Dim()
	for _, tmp := range f.tiles {
		if tmp.Geometry(stringBounder).HasPointOut() {
			return NewFtileGeometryDimOut(dimTotal, dimTotal.Width()/2, 0, dimTotal.Height())
		}
	}
	return NewFtileGeometryDim(dimTotal, dimTotal.Width()/2, 0)
}
