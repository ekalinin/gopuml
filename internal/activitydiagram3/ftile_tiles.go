// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileEmpty.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileLabel.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileGoto.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileBreak.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileKilled.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileMarged.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileMargedRight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileHeightFixedCentered.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileHeightFixedMarged.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileAssemblySimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// singleSwimlane is upstream's emptySet or singleton of a lane.
func singleSwimlane(swimlane *Swimlane) SwimlaneSet {
	if swimlane == nil {
		return nil
	}
	return SwimlaneSet{swimlane}
}

// FtileEmpty is an empty tile of a given size.
type FtileEmpty struct {
	AbstractFtile
	width, height float64
	swimlane      *Swimlane
}

// NewFtileEmpty returns an empty tile of size width x height in swimlane
// (nil for none).
func NewFtileEmpty(skinParam klimt.ISkinParam, width, height float64, swimlane *Swimlane) *FtileEmpty {
	f := &FtileEmpty{}
	f.initFtileEmpty(f, skinParam, width, height, swimlane)
	return f
}

func (f *FtileEmpty) initFtileEmpty(self Ftile, skinParam klimt.ISkinParam, width, height float64,
	swimlane *Swimlane) {
	f.initFtile(self, skinParam)
	f.width = width
	f.height = height
	f.swimlane = swimlane
}

// MyChildren implements Ftile: none.
func (f *FtileEmpty) MyChildren() []Ftile { return nil }

// DrawU implements klimt.TextBlock: nothing.
func (f *FtileEmpty) DrawU(klimt.UGraphic) {}

func (f *FtileEmpty) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	return f.calculateDimensionEmpty()
}

func (f *FtileEmpty) calculateDimensionEmpty() FtileGeometry {
	return NewFtileGeometryOut(f.width, f.height, f.width/2, 0, f.height)
}

// SwimlaneIn implements Swimable.
func (f *FtileEmpty) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileEmpty) SwimlaneOut() *Swimlane { return f.swimlane }

// Swimlanes implements Swimable.
func (f *FtileEmpty) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// FtileLabel is the empty tile of "label name".
type FtileLabel struct {
	FtileEmpty
	name string
}

// NewFtileLabel returns the tile of a label.
func NewFtileLabel(skinParam klimt.ISkinParam, swimlane *Swimlane, name string) *FtileLabel {
	f := &FtileLabel{name: name}
	f.initFtileEmpty(f, skinParam, 0, 0, swimlane)
	return f
}

// Name returns the name of the label.
func (f *FtileLabel) Name() string { return f.name }

// FtileGoto is the empty tile of "goto name", without exit point.
type FtileGoto struct {
	FtileEmpty
	name string
}

// NewFtileGoto returns the tile of a jump.
func NewFtileGoto(skinParam klimt.ISkinParam, swimlane *Swimlane, name string) *FtileGoto {
	f := &FtileGoto{name: name}
	f.initFtileEmpty(f, skinParam, 0, 0, swimlane)
	return f
}

func (f *FtileGoto) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	return f.calculateDimensionEmpty().WithoutPointOut()
}

// Name returns the label of the jump.
func (f *FtileGoto) Name() string { return f.name }

// FtileBreak is the empty tile of "break", without exit point; it is its
// own welding point.
type FtileBreak struct {
	FtileEmpty
}

// NewFtileBreak returns the tile of a break.
func NewFtileBreak(skinParam klimt.ISkinParam, swimlane *Swimlane) *FtileBreak {
	f := &FtileBreak{}
	f.initFtileEmpty(f, skinParam, 0, 0, swimlane)
	return f
}

func (*FtileBreak) isWeldingPoint() {}

func (f *FtileBreak) calculateDimensionFtile(font.StringBounder) FtileGeometry {
	return f.calculateDimensionEmpty().WithoutPointOut()
}

// WeldingPoints implements Ftile: the break itself.
func (f *FtileBreak) WeldingPoints() []WeldingPoint { return []WeldingPoint{f} }

// FtileKilled is a tile whose flow stops (kill, detach): no exit point.
type FtileKilled struct {
	AbstractFtile
	tile Ftile
}

// NewFtileKilled returns tileToKill without exit point.
func NewFtileKilled(tileToKill Ftile) *FtileKilled {
	f := &FtileKilled{tile: tileToKill}
	f.initFtile(f, tileToKill.SkinParam())
	return f
}

// MyChildren implements Ftile: the children of the killed tile.
func (f *FtileKilled) MyChildren() []Ftile { return f.tile.MyChildren() }

// Swimlanes implements Swimable.
func (f *FtileKilled) Swimlanes() SwimlaneSet { return f.tile.Swimlanes() }

// SwimlaneIn implements Swimable.
func (f *FtileKilled) SwimlaneIn() *Swimlane { return f.tile.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileKilled) SwimlaneOut() *Swimlane { return f.tile.SwimlaneOut() }

func (f *FtileKilled) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	g := f.tile.Geometry(stringBounder)
	return NewFtileGeometryDim(f.tile.CalculateDimension(stringBounder), g.Left(), g.InY())
}

// DrawU implements klimt.TextBlock.
func (f *FtileKilled) DrawU(ug klimt.UGraphic) { ug.Draw(f.tile) }

// ftileWrapper holds the forwarding methods of the tiles that wrap one
// tile (upstream repeats them in each class).
type ftileWrapper struct {
	AbstractFtile
	tile Ftile
}

// MyChildren implements Ftile: the wrapped tile.
func (f *ftileWrapper) MyChildren() []Ftile { return []Ftile{f.tile} }

// InLinkRendering implements Ftile.
func (f *ftileWrapper) InLinkRendering() *LinkRendering { return f.tile.InLinkRendering() }

// OutLinkRendering implements Ftile.
func (f *ftileWrapper) OutLinkRendering() *LinkRendering { return f.tile.OutLinkRendering() }

// Swimlanes implements Swimable.
func (f *ftileWrapper) Swimlanes() SwimlaneSet { return f.tile.Swimlanes() }

// SwimlaneIn implements Swimable.
func (f *ftileWrapper) SwimlaneIn() *Swimlane { return f.tile.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *ftileWrapper) SwimlaneOut() *Swimlane { return f.tile.SwimlaneOut() }

// FtileMarged adds margins on the left and right of a tile.
type FtileMarged struct {
	ftileWrapper
	margin1, margin2 float64
}

// NewFtileMarged returns tile with margins on the left and right.
func NewFtileMarged(tile Ftile, margin1, margin2 float64) *FtileMarged {
	f := &FtileMarged{ftileWrapper: ftileWrapper{tile: tile}, margin1: margin1, margin2: margin2}
	f.initFtile(f, tile.SkinParam())
	return f
}

func (f *FtileMarged) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	orig := f.tile.Geometry(stringBounder)
	return NewFtileGeometryOut(orig.Width()+f.margin1+f.margin2, orig.Height(), orig.Left()+f.margin1, orig.InY(),
		orig.OutY())
}

// TranslateFor implements Ftile.
func (f *FtileMarged) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == f.tile {
		t := f.translate()
		return &t
	}
	result := f.tile.TranslateFor(child, stringBounder).Compose(f.translate())
	return &result
}

func (f *FtileMarged) translate() klimt.UTranslate { return klimt.UTranslateDx(f.margin1) }

// DrawU implements klimt.TextBlock.
func (f *FtileMarged) DrawU(ug klimt.UGraphic) { ug.Apply(f.translate()).Draw(f.tile) }

// FtileMargedRight widens a tile on the right up to maxX.
type FtileMargedRight struct {
	ftileWrapper
	maxX float64
}

// NewFtileMargedRight returns tile widened to maxX.
func NewFtileMargedRight(tile Ftile, maxX float64) *FtileMargedRight {
	f := &FtileMargedRight{ftileWrapper: ftileWrapper{tile: tile}, maxX: maxX}
	f.initFtile(f, tile.SkinParam())
	return f
}

func (f *FtileMargedRight) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	orig := f.tile.Geometry(stringBounder)
	if orig.Width() > f.maxX {
		uml.Throw(uml.IllegalStateException, "")
	}
	return NewFtileGeometryOut(f.maxX, orig.Height(), orig.Left(), orig.InY(), orig.OutY())
}

// DrawU implements klimt.TextBlock.
func (f *FtileMargedRight) DrawU(ug klimt.UGraphic) { ug.Draw(f.tile) }

// TranslateFor implements Ftile.
func (f *FtileMargedRight) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == f.tile {
		t := klimt.UTranslateNone()
		return &t
	}
	return f.AbstractFtile.TranslateFor(child, stringBounder)
}

// FtileHeightFixedCentered centers a tile vertically in a fixed height.
type FtileHeightFixedCentered struct {
	ftileWrapper
	fixedHeight float64
}

// NewFtileHeightFixedCentered returns tile centered in fixedHeight.
func NewFtileHeightFixedCentered(tile Ftile, fixedHeight float64) *FtileHeightFixedCentered {
	f := &FtileHeightFixedCentered{ftileWrapper: ftileWrapper{tile: tile}, fixedHeight: fixedHeight}
	f.initFtile(f, tile.SkinParam())
	return f
}

func (f *FtileHeightFixedCentered) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	return f.tile.Geometry(stringBounder).Translate(f.translate(stringBounder)).FixedHeight(f.fixedHeight)
}

func (f *FtileHeightFixedCentered) translate(stringBounder font.StringBounder) klimt.UTranslate {
	dim := f.tile.CalculateDimension(stringBounder)
	if dim.Height() > f.fixedHeight {
		uml.Throw(uml.IllegalStateException, "")
	}
	return klimt.UTranslateDy((f.fixedHeight - dim.Height()) / 2)
}

// DrawU implements klimt.TextBlock.
func (f *FtileHeightFixedCentered) DrawU(ug klimt.UGraphic) {
	ug.Apply(f.translate(ug.StringBounder())).Draw(f.tile)
}

// FtileHeightFixedMarged adds margins above and below a tile.
type FtileHeightFixedMarged struct {
	ftileWrapper
	ymargin1, ymargin2 float64
}

// NewFtileHeightFixedMarged returns tile with margins above and below.
func NewFtileHeightFixedMarged(ymargin1 float64, tile Ftile, ymargin2 float64) *FtileHeightFixedMarged {
	f := &FtileHeightFixedMarged{ftileWrapper: ftileWrapper{tile: tile}, ymargin1: ymargin1, ymargin2: ymargin2}
	f.initFtile(f, tile.SkinParam())
	return f
}

func (f *FtileHeightFixedMarged) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dim := f.tile.Geometry(stringBounder)
	return dim.Translate(klimt.UTranslateDy(f.ymargin1)).FixedHeight(f.ymargin1 + dim.Height() + f.ymargin2)
}

// DrawU implements klimt.TextBlock.
func (f *FtileHeightFixedMarged) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.UTranslateDy(f.ymargin1)).Draw(f.tile)
}

// FtileAssemblySimple places a tile below another.
type FtileAssemblySimple struct {
	tile1, tile2       Ftile
	cachedTranslation  map[Ftile]klimt.UTranslate
	calculateDimension *FtileGeometry
}

var _ Ftile = (*FtileAssemblySimple)(nil)

// NewFtileAssemblySimple returns tile2 placed below tile1.
func NewFtileAssemblySimple(tile1, tile2 Ftile) *FtileAssemblySimple {
	return &FtileAssemblySimple{tile1: tile1, tile2: tile2, cachedTranslation: map[Ftile]klimt.UTranslate{}}
}

// SwimlaneIn implements Swimable.
func (f *FtileAssemblySimple) SwimlaneIn() *Swimlane { return f.tile1.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileAssemblySimple) SwimlaneOut() *Swimlane { return f.tile2.SwimlaneOut() }

// TranslateFor implements Ftile.
func (f *FtileAssemblySimple) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if result, ok := f.cachedTranslation[child]; ok {
		return &result
	}
	result := f.translateForSlow(child, stringBounder)
	f.cachedTranslation[child] = result
	return &result
}

func (f *FtileAssemblySimple) translateForSlow(child Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	if child == f.tile1 {
		return f.translated1(stringBounder)
	}
	if child == f.tile2 {
		return f.translated2(stringBounder)
	}
	if tmp := f.tile1.TranslateFor(child, stringBounder); tmp != nil {
		return tmp.Compose(f.translated1(stringBounder))
	}
	if tmp := f.tile2.TranslateFor(child, stringBounder); tmp != nil {
		return tmp.Compose(f.translated2(stringBounder))
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return klimt.UTranslate{}
}

// DrawU implements klimt.TextBlock.
func (f *FtileAssemblySimple) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug.Apply(*f.TranslateFor(f.tile1, stringBounder)).Draw(f.tile1)
	ug.Apply(*f.TranslateFor(f.tile2, stringBounder)).Draw(f.tile2)
}

// InLinkRendering implements Ftile.
func (f *FtileAssemblySimple) InLinkRendering() *LinkRendering { return f.tile1.InLinkRendering() }

// OutLinkRendering implements Ftile: none.
func (f *FtileAssemblySimple) OutLinkRendering() *LinkRendering { return LinkRenderingNone() }

// Geometry implements Ftile.
func (f *FtileAssemblySimple) Geometry(stringBounder font.StringBounder) FtileGeometry {
	if f.calculateDimension == nil {
		g := f.tile1.Geometry(stringBounder).AppendBottom(f.tile2.Geometry(stringBounder))
		f.calculateDimension = &g
	}
	return *f.calculateDimension
}

// CalculateDimension implements klimt.TextBlock.
func (f *FtileAssemblySimple) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return f.Geometry(stringBounder).Dim()
}

func (f *FtileAssemblySimple) translated1(stringBounder font.StringBounder) klimt.UTranslate {
	left := f.Geometry(stringBounder).Left()
	return klimt.UTranslateDx(left - f.tile1.Geometry(stringBounder).Left())
}

func (f *FtileAssemblySimple) translated2(stringBounder font.StringBounder) klimt.UTranslate {
	dim1 := f.tile1.CalculateDimension(stringBounder)
	left := f.Geometry(stringBounder).Left()
	return klimt.NewUTranslate(left-f.tile2.Geometry(stringBounder).Left(), dim1.Height())
}

// InnerConnections implements Ftile: none.
func (f *FtileAssemblySimple) InnerConnections() []Connection { return nil }

// Swimlanes implements Swimable.
func (f *FtileAssemblySimple) Swimlanes() SwimlaneSet {
	return SwimlaneSet(nil).addAll(f.tile1.Swimlanes()).addAll(f.tile2.Swimlanes())
}

// SkinParam implements Ftile.
func (f *FtileAssemblySimple) SkinParam() klimt.ISkinParam { return f.tile1.SkinParam() }

// WeldingPoints implements Ftile.
func (f *FtileAssemblySimple) WeldingPoints() []WeldingPoint {
	return append(append([]WeldingPoint(nil), f.tile1.WeldingPoints()...), f.tile2.WeldingPoints()...)
}

// MyChildren implements Ftile.
func (f *FtileAssemblySimple) MyChildren() []Ftile { return []Ftile{f.tile1, f.tile2} }

// ArrowHorizontalAlignment implements Ftile.
func (f *FtileAssemblySimple) ArrowHorizontalAlignment() geom.HorizontalAlignment {
	return f.tile1.ArrowHorizontalAlignment()
}

// MinMax implements klimt.TextBlock: it panics, as upstream's default.
func (f *FtileAssemblySimple) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return klimt.TextBlockDefaults{}.MinMax(stringBounder)
}

// InnerPosition implements klimt.TextBlock: it panics, as upstream's
// default.
func (f *FtileAssemblySimple) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return klimt.TextBlockDefaults{}.InnerPosition(member, stringBounder)
}

// MagneticBorder implements klimt.TextBlock.
func (f *FtileAssemblySimple) MagneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderNone{}
}

// Backcolor implements klimt.TextBlock: none.
func (f *FtileAssemblySimple) Backcolor() color.HColor { return nil }

// FtileUtilsAddConnection returns ftile drawing connection after it.
func FtileUtilsAddConnection(ftile Ftile, connection Connection) Ftile {
	if connection == nil {
		uml.ThrowNPE("")
	}
	return newFtileWithConnection(ftile, []Connection{connection})
}

// FtileUtilsAddConnections returns ftile drawing connections after it;
// ftile itself for none.
func FtileUtilsAddConnections(ftile Ftile, connections []Connection) Ftile {
	if len(connections) == 0 {
		return ftile
	}
	return newFtileWithConnection(ftile, connections)
}

// FtileUtilsWithSwimlaneOut returns ftile leaving in out.
func FtileUtilsWithSwimlaneOut(ftile Ftile, out *Swimlane) Ftile {
	return newFtileWithSwimlanes(ftile, ftile.SwimlaneIn(), out)
}

// FtileUtilsWithSwimlaneIn returns ftile entered in in.
func FtileUtilsWithSwimlaneIn(ftile Ftile, in *Swimlane) Ftile {
	return newFtileWithSwimlanes(ftile, in, ftile.SwimlaneOut())
}

// FtileUtilsAddBottom returns ftile with a margin below.
func FtileUtilsAddBottom(ftile Ftile, marginBottom float64) Ftile {
	return NewFtileMargedVertically(ftile, 0, marginBottom)
}

// FtileUtilsAddVerticalMargin returns ftile with margins above and below;
// ftile itself for none.
func FtileUtilsAddVerticalMargin(ftile Ftile, marginTop, marginBottom float64) Ftile {
	if marginTop == 0 && marginBottom == 0 {
		return ftile
	}
	return NewFtileMargedVertically(ftile, marginTop, marginBottom)
}

// FtileUtilsAddHorizontalMargin2 returns ftile with margins on the left and
// right; ftile itself for none.
func FtileUtilsAddHorizontalMargin2(ftile Ftile, margin1, margin2 float64) Ftile {
	if margin1 == 0 && margin2 == 0 {
		return ftile
	}
	return NewFtileMarged(ftile, margin1, margin2)
}

// FtileUtilsAddHorizontalMargin returns ftile with margin on both sides;
// ftile itself for none.
func FtileUtilsAddHorizontalMargin(ftile Ftile, margin float64) Ftile {
	if margin == 0 {
		return ftile
	}
	return NewFtileMarged(ftile, margin, margin)
}
