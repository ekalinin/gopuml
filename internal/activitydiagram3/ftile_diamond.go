// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDiamondWIP.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDiamond.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDiamondInside.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDiamondInside2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDiamondSquare.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/Hexagon.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// hexagonHalfSize is Hexagon.hexagonHalfSize.
const hexagonHalfSize = 12.0

// HexagonAsPolygon returns the diamond of 24 x 24.
func HexagonAsPolygon(shadowing float64) *klimt.UPolygon {
	diams := klimt.NewUPolygon()
	diams.AddPoint(hexagonHalfSize, 0)
	diams.AddPoint(hexagonHalfSize*2, hexagonHalfSize)
	diams.AddPoint(hexagonHalfSize, hexagonHalfSize*2)
	diams.AddPoint(0, hexagonHalfSize)
	diams.AddPoint(hexagonHalfSize, 0)
	diams.SetDeltaShadow(shadowing)
	return diams
}

// HexagonAsPolygonSize returns the hexagon of size width x height.
func HexagonAsPolygonSize(shadowing, width, height float64) *klimt.UPolygon {
	diams := klimt.NewUPolygon()
	diams.AddPoint(hexagonHalfSize, 0)
	diams.AddPoint(width-hexagonHalfSize, 0)
	diams.AddPoint(width, height/2)
	diams.AddPoint(width-hexagonHalfSize, height)
	diams.AddPoint(hexagonHalfSize, height)
	diams.AddPoint(0, height/2)
	diams.AddPoint(hexagonHalfSize, 0)
	diams.SetDeltaShadow(shadowing)
	return diams
}

// HexagonAsStencil returns the stencil of a hexagon around tb.
func HexagonAsStencil(tb klimt.TextBlock) klimt.Stencil {
	deltaX := func(height, y float64) float64 {
		p := y / height * 2
		if p <= 1 {
			return hexagonHalfSize * p
		}
		return hexagonHalfSize * (2 - p)
	}
	return klimt.StencilFunc{
		Starting: func(stringBounder font.StringBounder, y float64) float64 {
			dim := tb.CalculateDimension(stringBounder)
			return -deltaX(dim.Height(), y)
		},
		Ending: func(stringBounder font.StringBounder, y float64) float64 {
			dim := tb.CalculateDimension(stringBounder)
			return dim.Width() + deltaX(dim.Height(), y)
		},
	}
}

// HexagonAsPolygonSquare returns the diamond of size width x height.
func HexagonAsPolygonSquare(shadowing, width, height float64) *klimt.UPolygon {
	diams := klimt.NewUPolygon()
	diams.AddPoint(width/2, 0)
	diams.AddPoint(width, height/2)
	diams.AddPoint(width/2, height)
	diams.AddPoint(0, height/2)
	diams.SetDeltaShadow(shadowing)
	return diams
}

// FtileDiamondWIP holds the state of the diamond tiles: colors, lane, the
// label inside and the four labels around.
type FtileDiamondWIP struct {
	AbstractFtile
	backColor   color.HColor
	borderColor color.HColor
	swimlane    *Swimlane
	label       klimt.TextBlock
	north       klimt.TextBlock
	south       klimt.TextBlock
	west        klimt.TextBlock
	east        klimt.TextBlock
	shadowing   float64
}

func (f *FtileDiamondWIP) initDiamond(self Ftile, label klimt.TextBlock, skinParam klimt.ISkinParam,
	backColor, borderColor color.HColor, swimlane *Swimlane, north, south, east, west klimt.TextBlock) {
	f.initFtile(self, skinParam)
	style := f.StyleSignature().MergedStyle(skinParam.CurrentStyleBuilder())
	f.borderColor = borderColor
	f.backColor = backColor
	f.shadowing = style.Shadowing()
	f.swimlane = swimlane
	f.label = label
	f.north = north
	f.west = west
	f.east = east
	f.south = south
}

// SwapEastWest swaps the east and west labels.
func (f *FtileDiamondWIP) SwapEastWest() { f.west, f.east = f.east, f.west }

// StyleSignature returns the style signature of diamonds.
func (f *FtileDiamondWIP) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity, klimt.SNameDiamond)
}

// Style returns the style of diamonds.
func (f *FtileDiamondWIP) Style() *klimt.Style {
	return f.StyleSignature().MergedStyle(f.SkinParam().CurrentStyleBuilder())
}

// MyChildren implements Ftile: none.
func (f *FtileDiamondWIP) MyChildren() []Ftile { return nil }

// Swimlanes implements Swimable.
func (f *FtileDiamondWIP) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *FtileDiamondWIP) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileDiamondWIP) SwimlaneOut() *Swimlane { return f.swimlane }

func emptyTextBlock00() klimt.TextBlock { return klimt.TextBlockUtilsEmpty(0, 0) }

// FtileDiamond is the small diamond of a test, with labels around it.
type FtileDiamond struct {
	FtileDiamondWIP
}

// NewFtileDiamond returns a diamond without labels.
func NewFtileDiamond(skinParam klimt.ISkinParam, backColor, borderColor color.HColor, swimlane *Swimlane) *FtileDiamond {
	return newFtileDiamond(skinParam, backColor, borderColor, swimlane, emptyTextBlock00(), emptyTextBlock00(),
		emptyTextBlock00(), emptyTextBlock00())
}

func newFtileDiamond(skinParam klimt.ISkinParam, backColor, borderColor color.HColor, swimlane *Swimlane,
	north, south, east, west klimt.TextBlock) *FtileDiamond {
	f := &FtileDiamond{}
	f.initDiamond(f, nil, skinParam, backColor, borderColor, swimlane, north, south, east, west)
	return f
}

// WithNorth returns the diamond with a label above.
func (f *FtileDiamond) WithNorth(north klimt.TextBlock) *FtileDiamond {
	return newFtileDiamond(f.SkinParam(), f.backColor, f.borderColor, f.swimlane, north, f.south, f.east, f.west)
}

// WithWest returns the diamond with a label on the left; nil changes
// nothing.
func (f *FtileDiamond) WithWest(west1 klimt.TextBlock) *FtileDiamond {
	if west1 == nil {
		return f
	}
	return newFtileDiamond(f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south, f.east, west1)
}

// WithEast returns the diamond with a label on the right; nil changes
// nothing.
func (f *FtileDiamond) WithEast(east1 klimt.TextBlock) *FtileDiamond {
	if east1 == nil {
		return f
	}
	return newFtileDiamond(f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south, east1, f.west)
}

// WithSouth returns the diamond with a label below.
func (f *FtileDiamond) WithSouth(south klimt.TextBlock) *FtileDiamond {
	return newFtileDiamond(f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, south, f.east, f.west)
}

// DrawU implements klimt.TextBlock.
func (f *FtileDiamond) DrawU(ug klimt.UGraphic) {
	suppY1 := f.north.CalculateDimension(ug.StringBounder()).Height()
	ug = ug.Apply(klimt.UTranslateDy(suppY1))
	ug.Apply(f.borderColor).Apply(f.Style().Stroke()).Apply(klimt.Bg(f.backColor)).Draw(HexagonAsPolygon(f.shadowing))
	f.north.DrawU(ug.Apply(klimt.NewUTranslate(hexagonHalfSize*1.5, -suppY1)))
	f.south.DrawU(ug.Apply(klimt.NewUTranslate(hexagonHalfSize*1.5, 2*hexagonHalfSize)))
	dimWeat1 := f.west.CalculateDimension(ug.StringBounder())
	f.west.DrawU(ug.Apply(klimt.NewUTranslate(-dimWeat1.Width(), -dimWeat1.Height()+hexagonHalfSize)))
	dimEast1 := f.east.CalculateDimension(ug.StringBounder())
	f.east.DrawU(ug.Apply(klimt.NewUTranslate(hexagonHalfSize*2, -dimEast1.Height()+hexagonHalfSize)))
}

func (f *FtileDiamond) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	suppY1 := f.north.CalculateDimension(stringBounder).Height()
	dim := geom.NewXDimension2D(hexagonHalfSize*2, hexagonHalfSize*2+suppY1)
	return NewFtileGeometryDimOut(dim, dim.Width()/2, suppY1, dim.Height())
}

// WithWestAndEast returns the diamond with labels on both sides.
func (f *FtileDiamond) WithWestAndEast(tb1, tb2 klimt.TextBlock) Ftile {
	return f.WithWest(tb1).WithEast(tb2)
}

// WestEastLabelHeight returns the height of the higher side label.
func (f *FtileDiamond) WestEastLabelHeight(stringBounder font.StringBounder) float64 {
	dimEast := f.east.CalculateDimension(stringBounder)
	dimWest := f.west.CalculateDimension(stringBounder)
	return max(dimEast.Height(), dimWest.Height())
}

// EastLabelWidth returns the width of the east label.
func (f *FtileDiamond) EastLabelWidth(stringBounder font.StringBounder) float64 {
	return f.east.CalculateDimension(stringBounder).Width()
}

// SouthLabelHeight returns the height of the south label.
func (f *FtileDiamond) SouthLabelHeight(stringBounder font.StringBounder) float64 {
	return f.south.CalculateDimension(stringBounder).Height()
}

// diamondInsideAlone is calculateDimensionAlone of FtileDiamondInside and
// FtileDiamondInside2: the hexagon around the label.
func diamondInsideAlone(label klimt.TextBlock, stringBounder font.StringBounder) FtileGeometry {
	dimLabel := label.CalculateDimension(stringBounder)
	var dim geom.XDimension2D
	if dimLabel.Width() == 0 || dimLabel.Height() == 0 {
		dim = geom.NewXDimension2D(hexagonHalfSize*2, hexagonHalfSize*2)
	} else {
		dim = dimLabel.AtLeast(hexagonHalfSize*2, hexagonHalfSize*2).DeltaWH(hexagonHalfSize*2, 0)
	}
	return NewFtileGeometryDimOut(dim, dim.Width()/2, 0, dim.Height())
}

// drawDiamondInside is the drawing of FtileDiamondInside and
// FtileDiamondInside2.
func (f *FtileDiamondWIP) drawDiamondInside(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimLabel := f.label.CalculateDimension(stringBounder)
	dimTotal := diamondInsideAlone(f.label, stringBounder)
	ug = ug.Apply(f.borderColor).Apply(f.Style().Stroke()).Apply(klimt.Bg(f.backColor))
	ug.Draw(HexagonAsPolygonSize(f.shadowing, dimTotal.Width(), dimTotal.Height()))
	f.north.DrawU(ug.Apply(klimt.NewUTranslate(4+dimTotal.Width()/2, dimTotal.Height())))
	f.south.DrawU(ug.Apply(klimt.NewUTranslate(4+dimTotal.Width()/2, dimTotal.Height())))
	lx := (dimTotal.Width() - dimLabel.Width()) / 2
	ly := (dimTotal.Height() - dimLabel.Height()) / 2
	f.label.DrawU(ug.Apply(klimt.NewUTranslate(lx, ly)))
	dimWest := f.west.CalculateDimension(stringBounder)
	f.west.DrawU(ug.Apply(klimt.NewUTranslate(-dimWest.Width(), -dimWest.Height()+dimTotal.Height()/2)))
	dimEast := f.east.CalculateDimension(stringBounder)
	f.east.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width(), -dimEast.Height()+dimTotal.Height()/2)))
}

// FtileDiamondInside is a hexagon with the test inside.
type FtileDiamondInside struct {
	FtileDiamondWIP
}

// NewFtileDiamondInside returns the hexagon of label, without labels
// around.
func NewFtileDiamondInside(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane) *FtileDiamondInside {
	return newFtileDiamondInside(label, skinParam, backColor, borderColor, swimlane, emptyTextBlock00(),
		emptyTextBlock00(), emptyTextBlock00(), emptyTextBlock00())
}

func newFtileDiamondInside(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane, north, south, west, east klimt.TextBlock) *FtileDiamondInside {
	f := &FtileDiamondInside{}
	f.initDiamond(f, label, skinParam, backColor, borderColor, swimlane, north, south, east, west)
	return f
}

// WithNorth returns the hexagon with a label above.
func (f *FtileDiamondInside) WithNorth(north klimt.TextBlock) *FtileDiamondInside {
	return newFtileDiamondInside(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, north, f.south,
		f.west, f.east)
}

// WithWest returns the hexagon with a label on the left.
func (f *FtileDiamondInside) WithWest(west klimt.TextBlock) *FtileDiamondInside {
	return newFtileDiamondInside(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south,
		west, f.east)
}

// WithEast returns the hexagon with a label on the right.
func (f *FtileDiamondInside) WithEast(east klimt.TextBlock) *FtileDiamondInside {
	return newFtileDiamondInside(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south,
		f.west, east)
}

// WithWestAndEast returns the hexagon with labels on both sides.
func (f *FtileDiamondInside) WithWestAndEast(tb1, tb2 klimt.TextBlock) Ftile {
	return f.WithWest(tb1).WithEast(tb2)
}

// WithSouth returns the hexagon with a label below.
func (f *FtileDiamondInside) WithSouth(south klimt.TextBlock) *FtileDiamondInside {
	return newFtileDiamondInside(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, south,
		f.west, f.east)
}

// DrawU implements klimt.TextBlock.
func (f *FtileDiamondInside) DrawU(ug klimt.UGraphic) { f.drawDiamondInside(ug) }

func (f *FtileDiamondInside) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimDiamonAlone := diamondInsideAlone(f.label, stringBounder)
	f.west.CalculateDimension(stringBounder)
	f.east.CalculateDimension(stringBounder)
	northHeight := f.north.CalculateDimension(stringBounder).Height()
	return dimDiamonAlone.IncHeight(northHeight)
}

// EastLabelWidth returns the width of the east label.
func (f *FtileDiamondInside) EastLabelWidth(stringBounder font.StringBounder) float64 {
	return f.east.CalculateDimension(stringBounder).Width()
}

// SouthLabelHeight returns the height of the south label.
func (f *FtileDiamondInside) SouthLabelHeight(stringBounder font.StringBounder) float64 {
	return f.south.CalculateDimension(stringBounder).Height()
}

// FtileDiamondInside2 is a hexagon with the test inside, whose label above
// widens it.
type FtileDiamondInside2 struct {
	FtileDiamondWIP
}

// NewFtileDiamondInside2 returns the hexagon of label, without labels
// around.
func NewFtileDiamondInside2(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane) *FtileDiamondInside2 {
	return newFtileDiamondInside2(label, skinParam, backColor, borderColor, swimlane, emptyTextBlock00(),
		emptyTextBlock00(), emptyTextBlock00(), emptyTextBlock00())
}

func newFtileDiamondInside2(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane, north, south, west, east klimt.TextBlock) *FtileDiamondInside2 {
	f := &FtileDiamondInside2{}
	f.initDiamond(f, label, skinParam, backColor, borderColor, swimlane, north, south, east, west)
	return f
}

// WithNorth returns the hexagon with a label above.
func (f *FtileDiamondInside2) WithNorth(north klimt.TextBlock) *FtileDiamondInside2 {
	return newFtileDiamondInside2(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, north, f.south,
		f.west, f.east)
}

// WithWest returns the hexagon with a label on the left.
func (f *FtileDiamondInside2) WithWest(west klimt.TextBlock) *FtileDiamondInside2 {
	return newFtileDiamondInside2(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south,
		west, f.east)
}

// WithEast returns the hexagon with a label on the right.
func (f *FtileDiamondInside2) WithEast(east klimt.TextBlock) *FtileDiamondInside2 {
	return newFtileDiamondInside2(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.south,
		f.west, east)
}

// WithSouth returns the hexagon with a label below.
func (f *FtileDiamondInside2) WithSouth(south klimt.TextBlock) *FtileDiamondInside2 {
	return newFtileDiamondInside2(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, south,
		f.west, f.east)
}

// DrawU implements klimt.TextBlock.
func (f *FtileDiamondInside2) DrawU(ug klimt.UGraphic) { f.drawDiamondInside(ug) }

func (f *FtileDiamondInside2) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	diamond := diamondInsideAlone(f.label, stringBounder)
	north := f.north.CalculateDimension(stringBounder)
	height := diamond.Height() + north.Height()
	left := diamond.Width() / 2
	width := diamond.Width()
	if north.Width() > left {
		width = left + north.Width()
	}
	return NewFtileGeometryOut(width, height, left, 0, diamond.Height())
}

// FtileDiamondSquare is a diamond with the test inside.
type FtileDiamondSquare struct {
	FtileDiamondWIP
}

// NewFtileDiamondSquare returns the diamond of label, without labels
// around.
func NewFtileDiamondSquare(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane) *FtileDiamondSquare {
	return newFtileDiamondSquare(label, skinParam, backColor, borderColor, swimlane, emptyTextBlock00(),
		emptyTextBlock00(), emptyTextBlock00(), emptyTextBlock00())
}

func newFtileDiamondSquare(label klimt.TextBlock, skinParam klimt.ISkinParam, backColor, borderColor color.HColor,
	swimlane *Swimlane, north, west, east, south klimt.TextBlock) *FtileDiamondSquare {
	f := &FtileDiamondSquare{}
	f.initDiamond(f, label, skinParam, backColor, borderColor, swimlane, north, south, east, west)
	return f
}

// WithNorth returns the diamond with a label above.
func (f *FtileDiamondSquare) WithNorth(north klimt.TextBlock) *FtileDiamondSquare {
	return newFtileDiamondSquare(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, north, f.west,
		f.east, f.south)
}

// WithWest returns the diamond with a label on the left.
func (f *FtileDiamondSquare) WithWest(west klimt.TextBlock) *FtileDiamondSquare {
	return newFtileDiamondSquare(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, west,
		f.east, f.south)
}

// WithEast returns the diamond with a label on the right.
func (f *FtileDiamondSquare) WithEast(east klimt.TextBlock) *FtileDiamondSquare {
	return newFtileDiamondSquare(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.west,
		east, f.south)
}

// WithSouth returns the diamond with a label below.
func (f *FtileDiamondSquare) WithSouth(south klimt.TextBlock) *FtileDiamondSquare {
	return newFtileDiamondSquare(f.label, f.SkinParam(), f.backColor, f.borderColor, f.swimlane, f.north, f.west,
		f.east, south)
}

// WithWestAndEast returns the diamond with labels on both sides.
func (f *FtileDiamondSquare) WithWestAndEast(tb1, tb2 klimt.TextBlock) Ftile {
	return f.WithWest(tb1).WithEast(tb2)
}

// DrawU implements klimt.TextBlock.
func (f *FtileDiamondSquare) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimLabel := f.label.CalculateDimension(stringBounder)
	dimTotal := f.calculateDimensionInternal(stringBounder)
	ug = ug.Apply(f.borderColor).Apply(f.Style().Stroke()).Apply(klimt.Bg(f.backColor))
	ug.Draw(HexagonAsPolygonSquare(f.shadowing, dimTotal.Width(), dimTotal.Height()))
	f.north.DrawU(ug.Apply(klimt.NewUTranslate(4+dimTotal.Width()/2, dimTotal.Height())))
	f.south.DrawU(ug.Apply(klimt.NewUTranslate(4+dimTotal.Width()/2, dimTotal.Height())))
	lx := (dimTotal.Width() - dimLabel.Width()) / 2
	ly := (dimTotal.Height() - dimLabel.Height()) / 2
	f.label.DrawU(ug.Apply(klimt.NewUTranslate(lx, ly)))
	dimWest := f.west.CalculateDimension(stringBounder)
	f.west.DrawU(ug.Apply(klimt.NewUTranslate(-dimWest.Width(), -dimWest.Height()+hexagonHalfSize)))
	dimEast := f.east.CalculateDimension(stringBounder)
	f.east.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width(), -dimEast.Height()+hexagonHalfSize+5)))
}

func (f *FtileDiamondSquare) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dim := f.calculateDimensionInternal(stringBounder)
	return NewFtileGeometryDimOut(dim, dim.Width()/2, 0, dim.Height())
}

func (f *FtileDiamondSquare) calculateDimensionInternal(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := f.label.CalculateDimension(stringBounder)
	if dimLabel.Width() == 0 || dimLabel.Height() == 0 {
		return geom.NewXDimension2D(hexagonHalfSize*2, hexagonHalfSize*2)
	}
	return dimLabel.DeltaWH(hexagonHalfSize*2, hexagonHalfSize*2)
}
