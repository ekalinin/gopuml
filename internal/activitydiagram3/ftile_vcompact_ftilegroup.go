// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileGroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/decoration/symbol"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

const ftileGroupDiffYY2 = 20.0

// ftileGroupUSymbol maps the shape of a block to its decoration symbol.
func ftileGroupUSymbol(typ GroupSymbol) symbol.USymbol {
	switch typ {
	case SymbolPackage:
		return symbol.Package
	case SymbolRectangle:
		return symbol.Rectangle
	case SymbolCard:
		return symbol.Card
	case SymbolGroup:
		return symbol.Group
	}
	return symbol.Partition
}

// FtileGroup is a "partition", "package", "rectangle", "card" or "group"
// block drawn around a tile, with its title.
type FtileGroup struct {
	AbstractFtile
	inner                Ftile
	name                 klimt.TextBlock
	headerNote           klimt.TextBlock
	borderColor          color.HColor
	backColor            color.HColor
	shadowing            float64
	stroke               klimt.UStroke
	typ                  GroupSymbol
	roundCorner          float64
	cachedInnerDimension *FtileGeometry
}

// StyleSignature returns the style signature of the block.
func (f *FtileGroup) StyleSignature() klimt.StyleSignatureBasic { return groupStyleSignature(f.typ) }

// NewFtileGroup returns the block of shape typ around inner; title and
// backColor may be nil.
func NewFtileGroup(inner Ftile, title *klimt.Display, backColor color.HColor, skinParam klimt.ISkinParam,
	typ GroupSymbol, style *klimt.Style) *FtileGroup {
	f := &FtileGroup{typ: typ, inner: FtileUtilsAddHorizontalMargin(inner, 10)}
	f.initFtile(f, inner.SkinParam())
	fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
	f.shadowing = style.Shadowing()
	if backColor == nil {
		f.backColor = style.Value(klimt.PNameBackGroundColor).AsColor(f.IHtmlColorSet())
	} else {
		f.backColor = backColor
	}
	f.borderColor = style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	f.roundCorner = style.Value(klimt.PNameRoundCorner).AsDouble()
	if title == nil {
		f.name = klimt.TextBlockUtilsEmpty(0, 0)
	} else {
		f.name = title.Create(fc, geom.HorizontalAlignmentLeft, skinParam)
	}
	f.headerNote = klimt.TextBlockUtilsEmpty(0, 0)
	f.stroke = style.Stroke()
	return f
}

// MyChildren implements Ftile: the children of the margined inner tile.
func (f *FtileGroup) MyChildren() []Ftile { return f.inner.MyChildren() }

// InLinkRendering implements Ftile.
func (f *FtileGroup) InLinkRendering() *LinkRendering { return f.inner.InLinkRendering() }

// Swimlanes implements Swimable.
func (f *FtileGroup) Swimlanes() SwimlaneSet { return f.inner.Swimlanes() }

// SwimlaneIn implements Swimable.
func (f *FtileGroup) SwimlaneIn() *Swimlane { return f.inner.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileGroup) SwimlaneOut() *Swimlane { return f.inner.SwimlaneOut() }

func (f *FtileGroup) diffHeightTitle(stringBounder font.StringBounder) float64 {
	dimTitle := f.name.CalculateDimension(stringBounder)
	return max(25, dimTitle.Height()+20)
}

func (f *FtileGroup) translate(stringBounder font.StringBounder) klimt.UTranslate {
	suppWidth := f.SuppWidth(stringBounder)
	return klimt.NewUTranslate(suppWidth/2, f.diffHeightTitle(stringBounder)+f.headerNoteHeight(stringBounder))
}

func (f *FtileGroup) innerMinMax(stringBounder font.StringBounder) geom.MinMax {
	limitFinder := klimt.LimitFinderCreate(stringBounder, false)
	interceptor := NewUGraphicForSnake(limitFinder)
	interceptor2 := klimt.NewUGraphicDispatchDrawable(interceptor)
	f.inner.DrawU(interceptor2)
	interceptor2.FlushUg()
	return limitFinder.MinMax()
}

// SuppWidth returns the width the title adds to the inner tile.
func (f *FtileGroup) SuppWidth(stringBounder font.StringBounder) float64 {
	orig := f.innerDimension(stringBounder)
	dimTitle := f.name.CalculateDimension(stringBounder)
	dimHeaderNote := f.headerNote.CalculateDimension(stringBounder)
	return max(orig.Width(), dimTitle.Width()+20, dimHeaderNote.Width()+20) - orig.Width()
}

func (f *FtileGroup) innerDimension(stringBounder font.StringBounder) FtileGeometry {
	if f.cachedInnerDimension == nil {
		g := f.innerDimensionSlow(stringBounder)
		f.cachedInnerDimension = &g
	}
	return *f.cachedInnerDimension
}

func (f *FtileGroup) innerDimensionSlow(stringBounder font.StringBounder) FtileGeometry {
	orig := f.inner.Geometry(stringBounder)
	minMax := f.innerMinMax(stringBounder)
	missingWidth := minMax.MaxX() - orig.Width()
	if missingWidth > 0 {
		return orig.AddDim(missingWidth+5, 0)
	}
	return orig
}

func (f *FtileGroup) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	orig := f.innerDimension(stringBounder)
	suppWidth := f.SuppWidth(stringBounder)
	width := orig.Width() + suppWidth
	height := orig.Height() + f.diffHeightTitle(stringBounder) + ftileGroupDiffYY2 + f.headerNoteHeight(stringBounder)
	titleAndHeaderNoteHeight := f.diffHeightTitle(stringBounder) + f.headerNoteHeight(stringBounder)
	if orig.HasPointOut() {
		return NewFtileGeometryOut(width, height, orig.Left()+suppWidth/2, orig.InY()+titleAndHeaderNoteHeight,
			orig.OutY()+titleAndHeaderNoteHeight)
	}
	return NewFtileGeometry(width, height, orig.Left()+suppWidth/2, orig.InY()+titleAndHeaderNoteHeight)
}

func (f *FtileGroup) headerNoteHeight(stringBounder font.StringBounder) float64 {
	return f.headerNote.CalculateDimension(stringBounder).Height()
}

// DrawU implements klimt.TextBlock.
func (f *FtileGroup) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimTotal := f.CalculateDimension(stringBounder)
	symbolContext := klimt.NewFashion(f.backColor, f.borderColor).WithShadow(f.shadowing).WithStroke(f.stroke).
		WithCorner(f.roundCorner, 0)
	align := f.inner.SkinParam().HorizontalAlignment(klimt.AlignmentParamPackageTitleAlignment,
		klimt.ArrowDirectionNull, false, klimt.HorizontalAlignmentNull)
	ftileGroupUSymbol(f.typ).AsBig(f.name, align, klimt.TextBlockUtilsEmpty(0, 0), dimTotal.Width(),
		dimTotal.Height(), symbolContext, f.SkinParam().StereotypeAlignment()).DrawU(ug)
	dimHeaderNote := f.headerNote.CalculateDimension(stringBounder)
	f.headerNote.DrawU(ug.Apply(klimt.NewUTranslate(dimTotal.Width()-dimHeaderNote.Width()-10,
		f.diffHeightTitle(ug.StringBounder())-10)))
	ug.Apply(f.translate(stringBounder)).Draw(f.inner)
}
