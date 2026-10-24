// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileBox.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileCircleStart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileCircleStop.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileCircleEndCross.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileCircleSpot.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileBlackBlock.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileThinSplit.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// isDebugOption is upstream's Boolean.parseBoolean of the DEBUG option.
func isDebugOption(skinParam klimt.ISkinParam) bool {
	v, _ := skinParam.Options().Value(klimt.OptionDebug)
	return strings.EqualFold(v, "true")
}

// FtileBox is an action box.
type FtileBox struct {
	AbstractFtile
	padding             klimt.ClockwiseTopRightBottomLeft
	tb                  klimt.TextBlock
	roundCorner         float64
	shadowing           float64
	horizontalAlignment geom.HorizontalAlignment
	minimumWidth        float64
	inRendering         *LinkRendering
	swimlane            *Swimlane
	boxStyle            BoxStyle
	borderColor         color.HColor
	backColor           color.HColor
	style               *klimt.Style
	print               string
}

// FtileBoxStyleSignature is the style signature of action boxes.
func FtileBoxStyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity)
}

// FtileBoxStyleSignatureArrow is the style signature of the arrows.
func FtileBoxStyleSignatureArrow() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameArrow)
}

// InLinkRendering implements Ftile.
func (f *FtileBox) InLinkRendering() *LinkRendering { return f.inRendering }

// Swimlanes implements Swimable.
func (f *FtileBox) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *FtileBox) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileBox) SwimlaneOut() *Swimlane { return f.swimlane }

// FtileBoxCreate returns the box of label; stereotype and styleBuilder may
// be nil.
func FtileBoxCreate(skinParam klimt.ISkinParam, label *klimt.Display, swimlane *Swimlane, boxStyle BoxStyle,
	stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *FtileBox {
	if stereotype == nil {
		stereotype = boxStyle.Stereotype()
	}
	if styleBuilder == nil {
		styleBuilder = skinParam.CurrentStyleBuilder()
	}
	style := FtileBoxStyleSignature().WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
	styleArrow := FtileBoxStyleSignatureArrow().MergedStyle(styleBuilder)
	return newFtileBox(skinParam, label, swimlane, boxStyle, style, styleArrow)
}

func newFtileBox(skinParam klimt.ISkinParam, label *klimt.Display, swimlane *Swimlane, boxStyle BoxStyle,
	style, styleArrow *klimt.Style) *FtileBox {
	f := &FtileBox{style: style, boxStyle: boxStyle, swimlane: swimlane}
	f.initFtile(f, skinParam)
	f.inRendering = LinkRenderingCreate(RainbowFromStyle(styleArrow, f.IHtmlColorSet()))
	var specBack *color.Colors
	if colors, ok := skinParam.(*klimt.SkinParamColors); ok {
		specBack = colors.ColorsOf()
	}
	style = style.EventuallyOverrideColors(specBack)
	f.borderColor = style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
	f.backColor = style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
	f.horizontalAlignment = style.HorizontalAlignment()
	f.padding = style.Padding()
	f.roundCorner = style.Value(klimt.PNameRoundCorner).AsDouble()
	f.shadowing = style.Shadowing()
	wrapWidth := style.WrapWidth()
	f.minimumWidth = style.Value(klimt.PNameMinimumWidth).AsDouble()
	sheet := skinParam.Sheet(fc, skinParam.DefaultTextAlignment(f.horizontalAlignment), klimt.CreoleModeFull).
		CreateSheet(label)
	f.tb = klimt.NewSheetBlock2(klimt.NewSheetBlock1Padding(sheet, wrapWidth, skinParam.Padding()),
		klimt.StencilFunc{
			Starting: func(font.StringBounder, float64) float64 { return -f.padding.Left() },
			Ending: func(stringBounder font.StringBounder, y float64) float64 {
				return f.CalculateDimension(stringBounder).Width() - f.padding.Right()
			},
		}, klimt.UStrokeWithThickness(1))
	f.print = label.String()
	return f
}

// String is upstream's toString: the label.
func (f *FtileBox) String() string { return f.print }

// DrawU implements klimt.TextBlock.
func (f *FtileBox) DrawU(ug klimt.UGraphic) {
	dimTotal := f.CalculateDimension(ug.StringBounder())
	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
	if isDebugOption(f.SkinParam()) {
		ug = ug.Apply(color.Blue).Apply(klimt.UStrokeWithThickness(2))
		f.boxStyle.DrawMeDebug(ug, widthTotal, heightTotal, f.shadowing, f.roundCorner)
	}
	thickness := f.style.Stroke()
	if f.borderColor == nil {
		ug = ug.Apply(color.None())
	} else {
		ug = ug.Apply(f.borderColor)
	}
	if f.backColor == nil {
		ug = ug.Apply(klimt.Bg(color.None()))
	} else {
		ug = ug.Apply(klimt.Bg(f.backColor))
	}
	ug = ug.Apply(thickness)
	f.boxStyle.DrawMe(ug, widthTotal, heightTotal, f.shadowing, f.roundCorner)

	switch f.horizontalAlignment {
	case geom.HorizontalAlignmentLeft:
		f.tb.DrawU(ug.Apply(klimt.NewUTranslate(f.padding.Left(), f.padding.Top())))
	case geom.HorizontalAlignmentRight:
		dimTb := f.tb.CalculateDimension(ug.StringBounder())
		f.tb.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width()-dimTb.Width()-f.padding.Right(),
			f.padding.Bottom())))
	case geom.HorizontalAlignmentCenter:
		dimTb := f.tb.CalculateDimension(ug.StringBounder())
		f.tb.DrawU(ug.Apply(klimt.NewUTranslate((dimTotal.Width()-dimTb.Width())/2, f.padding.Bottom())))
	}
}

func (f *FtileBox) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimRaw := f.tb.CalculateDimension(stringBounder)
	dimRaw = dimRaw.DeltaWH(f.padding.Left()+f.padding.Right(), f.padding.Bottom()+f.padding.Top())
	dimRaw = dimRaw.AtLeast(f.minimumWidth, 0)
	return NewFtileGeometryOut(dimRaw.Width()+f.boxStyle.Shield(), dimRaw.Height(), dimRaw.Width()/2, 0,
		dimRaw.Height())
}

// MyChildren implements Ftile: none.
func (f *FtileBox) MyChildren() []Ftile { return nil }

// ftileLeaf holds the lane of the tiles without children.
type ftileLeaf struct {
	AbstractFtile
	swimlane *Swimlane
}

// MyChildren implements Ftile: none.
func (f *ftileLeaf) MyChildren() []Ftile { return nil }

// Swimlanes implements Swimable.
func (f *ftileLeaf) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *ftileLeaf) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *ftileLeaf) SwimlaneOut() *Swimlane { return f.swimlane }

// FtileCircleStart is the start node.
type FtileCircleStart struct {
	ftileLeaf
	circle *klimt.CircleStart
}

// NewFtileCircleStart returns a start node.
func NewFtileCircleStart(skinParam klimt.ISkinParam, swimlane *Swimlane, style *klimt.Style,
	colors *color.Colors) *FtileCircleStart {
	f := &FtileCircleStart{ftileLeaf: ftileLeaf{swimlane: swimlane},
		circle: klimt.NewCircleStart(skinParam, style, colors)}
	f.initFtile(f, skinParam)
	return f
}

// DrawU implements klimt.TextBlock.
func (f *FtileCircleStart) DrawU(ug klimt.UGraphic) { f.circle.DrawU(ug) }

func (f *FtileCircleStart) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	size := f.circle.CalculateDimension(stringBounder).Width()
	return NewFtileGeometryOut(size, size, size/2, 0, size)
}

// FtileCircleStop is the stop node: no exit point.
type FtileCircleStop struct {
	ftileLeaf
	circle *klimt.CircleEnd
}

// NewFtileCircleStop returns a stop node.
func NewFtileCircleStop(skinParam klimt.ISkinParam, swimlane *Swimlane, style *klimt.Style,
	colors *color.Colors) *FtileCircleStop {
	f := &FtileCircleStop{ftileLeaf: ftileLeaf{swimlane: swimlane},
		circle: klimt.NewCircleEnd(skinParam, style, colors)}
	f.initFtile(f, skinParam)
	return f
}

// DrawU implements klimt.TextBlock.
func (f *FtileCircleStop) DrawU(ug klimt.UGraphic) { f.circle.DrawU(ug) }

func (f *FtileCircleStop) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	const size = 22
	return NewFtileGeometry(size, size, size/2, 0)
}

// FtileCircleEndCross is the end node, a crossed circle: no exit point.
type FtileCircleEndCross struct {
	ftileLeaf
	lineColor color.HColor
	backColor color.HColor
	style     *klimt.Style
	shadowing float64
}

const ftileCircleEndCrossSize = 20

// NewFtileCircleEndCross returns an end node.
func NewFtileCircleEndCross(skinParam klimt.ISkinParam, swimlane *Swimlane, style *klimt.Style,
	colors *color.Colors) *FtileCircleEndCross {
	f := &FtileCircleEndCross{ftileLeaf: ftileLeaf{swimlane: swimlane}, style: style, shadowing: style.Shadowing(),
		backColor: klimt.ColorsColorOf(colors, style, klimt.PNameBackGroundColor, skinParam.IHtmlColorSet()),
		lineColor: klimt.ColorsColorOf(colors, style, klimt.PNameLineColor, skinParam.IHtmlColorSet())}
	f.initFtile(f, skinParam)
	return f
}

// DrawU implements klimt.TextBlock.
func (f *FtileCircleEndCross) DrawU(ug klimt.UGraphic) {
	const size = ftileCircleEndCrossSize
	xTheoricalPosition := float64(jmathRound(0))
	yTheoricalPosition := float64(jmathRound(0))
	circle := klimt.UEllipseBuild(size, size)
	circle.SetDeltaShadow(f.shadowing)
	ug = ug.Apply(f.lineColor).Apply(klimt.Bg(color.Transparent()))
	ug.Apply(f.style.Stroke()).Apply(klimt.NewUTranslate(xTheoricalPosition, yTheoricalPosition)).Draw(circle)

	ug = ug.Apply(klimt.Bg(f.backColor))
	thickness := 2.5
	size2 := (size - thickness) / math.Sqrt(2)
	delta := (size - size2) / 2
	ug = ug.Apply(klimt.UStrokeWithThickness(thickness))
	ug.Apply(klimt.NewUTranslate(delta, delta)).Draw(klimt.NewULine(size2, size2))
	ug.Apply(klimt.NewUTranslate(delta, size-delta)).Draw(klimt.NewULine(size2, -size2))
}

// jmathRound is Java's Math.round of a double.
func jmathRound(v float64) int64 { return int64(math.Floor(v + 0.5)) }

func (f *FtileCircleEndCross) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	const size = ftileCircleEndCrossSize
	return NewFtileGeometry(size, size, size/2, 0)
}

// FtileCircleSpot is a connector, a circle with a letter.
type FtileCircleSpot struct {
	ftileLeaf
	spot      string
	fc        *klimt.FontConfiguration
	backColor color.HColor
	style     *klimt.Style
}

const ftileCircleSpotSize = 20

// NewFtileCircleSpot returns a connector; backColor may be nil.
func NewFtileCircleSpot(skinParam klimt.ISkinParam, swimlane *Swimlane, spot string, _ font.UFont,
	backColor color.HColor, style *klimt.Style) *FtileCircleSpot {
	f := &FtileCircleSpot{ftileLeaf: ftileLeaf{swimlane: swimlane}, style: style, spot: spot, backColor: backColor}
	f.initFtile(f, skinParam)
	f.fc = style.FontConfiguration(f.IHtmlColorSet(), nil)
	return f
}

// DrawU implements klimt.TextBlock.
func (f *FtileCircleSpot) DrawU(ug klimt.UGraphic) {
	const size = ftileCircleSpotSize
	circle := klimt.UEllipseBuild(size, size)
	backColor := f.backColor
	if backColor == nil {
		backColor = f.style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	}
	borderColor := f.style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
	circle.SetDeltaShadow(f.style.Shadowing())
	ug.Apply(borderColor).Apply(klimt.Bg(backColor)).Apply(f.style.Stroke()).Draw(circle)
	units := utf16.Encode([]rune(f.spot))
	if len(units) == 0 {
		uml.ThrowIndexOutOfBounds(0, 0)
	}
	ug.Apply(f.fc.Color()).Apply(klimt.NewUTranslate(size/2, size/2)).
		Draw(klimt.NewUCenteredCharacter(rune(units[0]), f.fc.Font()))
}

func (f *FtileCircleSpot) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	const size = ftileCircleSpotSize
	return NewFtileGeometryOut(size, size, size/2, 0, size)
}

// FtileBlackBlock is the bar of a fork.
type FtileBlackBlock struct {
	AbstractFtile
	width, height float64
	colors        *color.Colors
	label         klimt.TextBlock
	swimlane      *Swimlane
}

const ftileBlackBlockLabelMargin = 5

// NewFtileBlackBlock returns an empty bar; colors may be nil.
func NewFtileBlackBlock(skinParam klimt.ISkinParam, swimlane *Swimlane, colors *color.Colors) *FtileBlackBlock {
	f := &FtileBlackBlock{colors: colors, swimlane: swimlane, label: klimt.TextBlockUtilsEmpty(0, 0)}
	f.initFtile(f, skinParam)
	return f
}

// SetBlackBlockDimension sets the size of the bar.
func (f *FtileBlackBlock) SetBlackBlockDimension(width, height float64) {
	f.height = height
	f.width = width
}

// SetLabel sets the label drawn right of the bar.
func (f *FtileBlackBlock) SetLabel(label klimt.TextBlock) {
	if label == nil {
		uml.ThrowNPE("")
	}
	f.label = label
}

func (f *FtileBlackBlock) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	supp := f.label.CalculateDimension(stringBounder).Width()
	if supp > 0 {
		supp += ftileBlackBlockLabelMargin
	}
	return NewFtileGeometryOut(f.width+supp, f.height, f.width/2, 0, f.height)
}

// DrawU implements klimt.TextBlock.
func (f *FtileBlackBlock) DrawU(ug klimt.UGraphic) {
	rect := klimt.URectangleBuild(f.width, f.height).Rounded(5).IgnoreForCompressionOnX()
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivityBar).MergedStyle(f.SkinParam().CurrentStyleBuilder())
	rect.SetDeltaShadow(style.Shadowing())
	var colorBar color.HColor
	if f.colors == nil {
		colorBar = style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	} else {
		colorBar = klimt.ColorsColorOf(f.colors, style, klimt.PNameBackGroundColor, f.IHtmlColorSet())
	}
	ug.Apply(colorBar).Apply(klimt.Bg(colorBar)).Draw(rect)
	dimLabel := f.label.CalculateDimension(ug.StringBounder())
	f.label.DrawU(ug.Apply(klimt.NewUTranslate(f.width+ftileBlackBlockLabelMargin, -dimLabel.Height()/2)))
}

// Swimlanes implements Swimable: the lane, possibly nil.
func (f *FtileBlackBlock) Swimlanes() SwimlaneSet { return SwimlaneSet{f.swimlane} }

// SwimlaneIn implements Swimable.
func (f *FtileBlackBlock) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileBlackBlock) SwimlaneOut() *Swimlane { return f.swimlane }

// MyChildren implements Ftile: none.
func (f *FtileBlackBlock) MyChildren() []Ftile { return nil }

// FtileThinSplit is the thin bar of a split.
type FtileThinSplit struct {
	AbstractFtile
	width, first, last float64
	colorBar           color.HColor
	swimlane           *Swimlane
}

const ftileThinSplitHeight = 1.5

// NewFtileThinSplit returns an empty split bar; colorBar may be nil.
func NewFtileThinSplit(skinParam klimt.ISkinParam, colorBar color.HColor, swimlane *Swimlane) *FtileThinSplit {
	f := &FtileThinSplit{colorBar: colorBar, swimlane: swimlane}
	f.initFtile(f, skinParam)
	return f
}

// MyChildren implements Ftile: none.
func (f *FtileThinSplit) MyChildren() []Ftile { return nil }

// SetGeom sets the extent of the bar and the width of the tile.
func (f *FtileThinSplit) SetGeom(first, last, width float64) {
	f.width = width
	f.first = first
	f.last = last
}

func (f *FtileThinSplit) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	return NewFtileGeometryOut(f.width, ftileThinSplitHeight, f.width/2, 0, ftileThinSplitHeight)
}

// DrawU implements klimt.TextBlock.
func (f *FtileThinSplit) DrawU(ug klimt.UGraphic) {
	rect := klimt.ULineH(f.last - f.first)
	ug = ug.Apply(klimt.UTranslateDx(f.first))
	if f.colorBar == nil {
		ug = ug.Apply(color.None())
	} else {
		ug = ug.Apply(f.colorBar)
	}
	ug.Apply(klimt.UStrokeWithThickness(1.5)).Draw(rect)
}

// Swimlanes implements Swimable: the lane, possibly nil.
func (f *FtileThinSplit) Swimlanes() SwimlaneSet { return SwimlaneSet{f.swimlane} }

// SwimlaneIn implements Swimable.
func (f *FtileThinSplit) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileThinSplit) SwimlaneOut() *Swimlane { return f.swimlane }
