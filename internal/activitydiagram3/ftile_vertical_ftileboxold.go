// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileBoxOld.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileBoxOld is the box of a mindmap or WBS node. Unlike FtileBox, its text
// takes the padding of its style, without the padding of the skin
// parameters, and is at least MinimumWidth wide.
type FtileBoxOld struct {
	AbstractFtile
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

// InLinkRendering implements Ftile.
func (f *FtileBoxOld) InLinkRendering() *LinkRendering { return f.inRendering }

// Swimlanes implements Swimable.
func (f *FtileBoxOld) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *FtileBoxOld) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileBoxOld) SwimlaneOut() *Swimlane { return f.swimlane }

// SetMinimumWidth is upstream's setMinimumWidth: the box becomes at least
// minimumWidth wide, and its geometry is computed again.
func (f *FtileBoxOld) SetMinimumWidth(minimumWidth float64) {
	f.minimumWidth = max(f.minimumWidth, minimumWidth)
	f.invalidateGeometryCache()
}

// FtileBoxOldCreateWbs is upstream's createWbs: the plain box of a WBS node,
// drawn with style. It is the box itself, so that WBS diagrams can widen it.
func FtileBoxOldCreateWbs(style *klimt.Style, skinParam klimt.ISkinParam, label *klimt.Display) *FtileBoxOld {
	styleArrow := style
	return newFtileBoxOld(skinParam, label, nil, BoxStylePlain, style, styleArrow)
}

// FtileBoxOldCreateMindMap is upstream's createMindMap: the plain box of a
// mindmap node, drawn with style.
func FtileBoxOldCreateMindMap(style *klimt.Style, skinParam klimt.ISkinParam, label *klimt.Display) klimt.TextBlock {
	return newFtileBoxOld(skinParam, label, nil, BoxStylePlain, style, style)
}

func newFtileBoxOld(skinParam klimt.ISkinParam, label *klimt.Display, swimlane *Swimlane, boxStyle BoxStyle,
	style, styleArrow *klimt.Style) *FtileBoxOld {
	f := &FtileBoxOld{}
	f.initFtile(f, skinParam)
	var specBack *color.Colors
	if colors, ok := skinParam.(*klimt.SkinParamColors); ok {
		specBack = colors.ColorsOf()
	}
	style = style.EventuallyOverrideColors(specBack)
	f.style = style
	f.boxStyle = boxStyle
	f.swimlane = swimlane
	f.inRendering = LinkRenderingCreate(RainbowFromStyle(styleArrow, f.IHtmlColorSet()))
	f.borderColor = style.Value(klimt.PNameLineColor).AsColor(f.IHtmlColorSet())
	f.backColor = style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
	f.horizontalAlignment = style.HorizontalAlignment()
	f.roundCorner = style.Value(klimt.PNameRoundCorner).AsDouble()
	f.shadowing = style.Shadowing()
	wrapWidth := style.WrapWidth()
	f.minimumWidth = style.Value(klimt.PNameMinimumWidth).AsDouble()
	sheet := skinParam.Sheet(fc, f.horizontalAlignment, klimt.CreoleModeFull).CreateSheet(label)
	f.tb = klimt.NewSheetBlock2(klimt.NewSheetBlock1Style(sheet, wrapWidth, style),
		klimt.StencilFunc{
			Starting: func(font.StringBounder, float64) float64 { return 0 },
			Ending: func(stringBounder font.StringBounder, y float64) float64 {
				return f.CalculateDimension(stringBounder).Width()
			},
		}, klimt.UStrokeWithThickness(1))
	f.print = label.String()
	return f
}

// String is upstream's toString: the label.
func (f *FtileBoxOld) String() string { return f.print }

// DrawU implements klimt.TextBlock.
func (f *FtileBoxOld) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := f.CalculateDimension(stringBounder)
	widthTotal := dimTotal.Width()
	heightTotal := dimTotal.Height()
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
		f.tb.DrawU(ug)
	case geom.HorizontalAlignmentRight:
		f.tb.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width()-f.tbWidth(stringBounder), 0)))
	case geom.HorizontalAlignmentCenter:
		f.tb.DrawU(ug.Apply(klimt.NewUTranslate((dimTotal.Width()-f.tbWidth(stringBounder))/2, 0)))
	}
}

func (f *FtileBoxOld) tbWidth(stringBounder font.StringBounder) float64 {
	return max(f.minimumWidth, f.tb.CalculateDimension(stringBounder).Width())
}

func (f *FtileBoxOld) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimRaw := f.tb.CalculateDimension(stringBounder)
	dimRaw = dimRaw.AtLeast(f.minimumWidth, 0)
	return NewFtileGeometryOut(dimRaw.Width()+f.boxStyle.Shield(), dimRaw.Height(), dimRaw.Width()/2, 0,
		dimRaw.Height())
}

// MyChildren implements Ftile: none.
func (f *FtileBoxOld) MyChildren() []Ftile { return nil }
