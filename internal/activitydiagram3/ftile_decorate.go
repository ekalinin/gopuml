// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDecorate.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDecorateIn.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDecorateOut.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDecorateInLabel.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileDecorateOutLabel.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileWithConnection.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileWithUrl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileDecorateWelding.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileWithSwimlanes.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileMargedVertically.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileMargedWest.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileMinWidthCentered.java
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

// FtileDecorate is a tile that wraps another and forwards to it. A
// decorator embeds it and calls initDecorate with itself, so that the size
// of the text block reaches its Geometry.
type FtileDecorate struct {
	self  Ftile
	ftile Ftile
}

func (d *FtileDecorate) initDecorate(self, ftile Ftile) {
	d.self = self
	d.ftile = ftile
}

// OutLinkRendering implements Ftile.
func (d *FtileDecorate) OutLinkRendering() *LinkRendering { return d.ftile.OutLinkRendering() }

// InLinkRendering implements Ftile.
func (d *FtileDecorate) InLinkRendering() *LinkRendering { return d.ftile.InLinkRendering() }

// DrawU implements klimt.TextBlock.
func (d *FtileDecorate) DrawU(ug klimt.UGraphic) { d.ftile.DrawU(ug) }

// Geometry implements Ftile.
func (d *FtileDecorate) Geometry(stringBounder font.StringBounder) FtileGeometry {
	return d.ftile.Geometry(stringBounder)
}

// CalculateDimension implements klimt.TextBlock: the size of the geometry.
func (d *FtileDecorate) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return d.self.Geometry(stringBounder).Dim()
}

// InnerConnections implements Ftile.
func (d *FtileDecorate) InnerConnections() []Connection { return d.ftile.InnerConnections() }

// Swimlanes implements Swimable.
func (d *FtileDecorate) Swimlanes() SwimlaneSet { return d.ftile.Swimlanes() }

// SwimlaneIn implements Swimable.
func (d *FtileDecorate) SwimlaneIn() *Swimlane { return d.ftile.SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (d *FtileDecorate) SwimlaneOut() *Swimlane { return d.ftile.SwimlaneOut() }

// SkinParam implements Ftile.
func (d *FtileDecorate) SkinParam() klimt.ISkinParam { return d.ftile.SkinParam() }

// FtileDelegated returns the wrapped tile.
func (d *FtileDecorate) FtileDelegated() Ftile { return d.ftile }

// WeldingPoints implements Ftile.
func (d *FtileDecorate) WeldingPoints() []WeldingPoint { return d.ftile.WeldingPoints() }

// TranslateFor implements Ftile.
func (d *FtileDecorate) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == d.ftile {
		t := klimt.UTranslateNone()
		return &t
	}
	return d.ftile.TranslateFor(child, stringBounder)
}

// MyChildren implements Ftile.
func (d *FtileDecorate) MyChildren() []Ftile {
	if d.self == d.ftile {
		uml.Throw(uml.IllegalStateException, "")
	}
	return []Ftile{d.ftile}
}

// ArrowHorizontalAlignment implements Ftile.
func (d *FtileDecorate) ArrowHorizontalAlignment() geom.HorizontalAlignment {
	return d.ftile.ArrowHorizontalAlignment()
}

// MinMax implements klimt.TextBlock: it panics, as upstream's default.
func (d *FtileDecorate) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return klimt.TextBlockDefaults{}.MinMax(stringBounder)
}

// InnerPosition implements klimt.TextBlock: it panics, as upstream's
// default.
func (d *FtileDecorate) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return klimt.TextBlockDefaults{}.InnerPosition(member, stringBounder)
}

// MagneticBorder implements klimt.TextBlock.
func (d *FtileDecorate) MagneticBorder() klimt.MagneticBorder { return klimt.MagneticBorderNone{} }

// Backcolor implements klimt.TextBlock: none.
func (d *FtileDecorate) Backcolor() color.HColor { return nil }

// FtileDecorateIn sets the arrow that enters a tile.
type FtileDecorateIn struct {
	FtileDecorate
	linkRendering *LinkRendering
}

// NewFtileDecorateIn returns ftile entered by linkRendering.
func NewFtileDecorateIn(ftile Ftile, linkRendering *LinkRendering) *FtileDecorateIn {
	d := &FtileDecorateIn{linkRendering: requireLink(linkRendering)}
	d.initDecorate(d, ftile)
	return d
}

// InLinkRendering implements Ftile.
func (d *FtileDecorateIn) InLinkRendering() *LinkRendering { return d.linkRendering }

// FtileDecorateOut sets the arrow that leaves a tile.
type FtileDecorateOut struct {
	FtileDecorate
	linkRendering *LinkRendering
}

// NewFtileDecorateOut returns ftile left by linkRendering.
func NewFtileDecorateOut(ftile Ftile, linkRendering *LinkRendering) *FtileDecorateOut {
	d := &FtileDecorateOut{linkRendering: requireLink(linkRendering)}
	d.initDecorate(d, ftile)
	return d
}

// OutLinkRendering implements Ftile.
func (d *FtileDecorateOut) OutLinkRendering() *LinkRendering { return d.linkRendering }

// FtileDecorateInLabel adds room above a tile for the label of its
// incoming arrow.
type FtileDecorateInLabel struct {
	FtileDecorate
	xl, yl float64
}

// NewFtileDecorateInLabel returns ftile with room of size dim above.
func NewFtileDecorateInLabel(ftile Ftile, dim geom.XDimension2D) *FtileDecorateInLabel {
	d := &FtileDecorateInLabel{xl: dim.Width(), yl: dim.Height()}
	d.initDecorate(d, ftile)
	return d
}

// Geometry implements Ftile.
func (d *FtileDecorateInLabel) Geometry(stringBounder font.StringBounder) FtileGeometry {
	result := d.FtileDecorate.Geometry(stringBounder)
	result = result.AddTop(d.yl)
	if missing := d.xl - result.Right(); missing > 0 {
		result = result.IncRight(missing)
	}
	return result
}

// DrawU implements klimt.TextBlock.
func (d *FtileDecorateInLabel) DrawU(ug klimt.UGraphic) {
	d.FtileDecorate.DrawU(ug.Apply(klimt.UTranslateDy(d.yl)))
}

// FtileDecorateOutLabel adds room below a tile for the label of its
// outgoing arrow.
type FtileDecorateOutLabel struct {
	FtileDecorate
	xl, yl float64
}

// NewFtileDecorateOutLabel returns ftile with room of size dim below.
func NewFtileDecorateOutLabel(ftile Ftile, dim geom.XDimension2D) *FtileDecorateOutLabel {
	d := &FtileDecorateOutLabel{xl: dim.Width(), yl: dim.Height()}
	d.initDecorate(d, ftile)
	return d
}

// Geometry implements Ftile.
func (d *FtileDecorateOutLabel) Geometry(stringBounder font.StringBounder) FtileGeometry {
	result := d.FtileDecorate.Geometry(stringBounder)
	result = result.AddBottom(d.yl)
	if missing := d.xl - result.Right(); missing > 0 {
		result = result.IncRight(missing)
	}
	return result
}

// FtileWithConnection draws connections after a tile.
type FtileWithConnection struct {
	FtileDecorate
	connections []Connection
}

func newFtileWithConnection(ftile Ftile, connections []Connection) *FtileWithConnection {
	if len(connections) == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	d := &FtileWithConnection{connections: slices.Clone(connections)}
	d.initDecorate(d, ftile)
	return d
}

// DrawU implements klimt.TextBlock: the tile, then its connections as
// shapes.
func (d *FtileWithConnection) DrawU(ug klimt.UGraphic) {
	d.ftile.DrawU(ug)
	for _, c := range d.connections {
		ug.Draw(c)
	}
}

// InnerConnections implements Ftile: only the added connections, as
// upstream (it drops those of the wrapped tile).
func (d *FtileWithConnection) InnerConnections() []Connection { return d.connections }

// FtileWithUrl draws a tile inside a link.
type FtileWithUrl struct {
	FtileDecorate
	url *klimt.URL
}

// NewFtileWithUrl returns ftile linked to url.
func NewFtileWithUrl(ftile Ftile, url *klimt.URL) *FtileWithUrl {
	if url == nil {
		uml.ThrowNPE("")
	}
	d := &FtileWithUrl{url: url}
	d.initDecorate(d, ftile)
	return d
}

// DrawU implements klimt.TextBlock.
func (d *FtileWithUrl) DrawU(ug klimt.UGraphic) {
	ug.StartURL(d.url)
	d.ftile.DrawU(ug)
	ug.CloseURL()
}

// FtileDecorateWelding sets the welding points of a tile.
type FtileDecorateWelding struct {
	FtileDecorate
	breaks []WeldingPoint
}

// NewFtileDecorateWelding returns ftile with the welding points breaks.
func NewFtileDecorateWelding(ftile Ftile, breaks []WeldingPoint) *FtileDecorateWelding {
	d := &FtileDecorateWelding{breaks: slices.Clone(breaks)}
	d.initDecorate(d, ftile)
	return d
}

// WeldingPoints implements Ftile.
func (d *FtileDecorateWelding) WeldingPoints() []WeldingPoint { return d.breaks }

// FtileWithSwimlanes overrides the incoming and outgoing lanes of a tile.
type FtileWithSwimlanes struct {
	FtileDecorate
	in, out *Swimlane
}

func newFtileWithSwimlanes(ftile Ftile, in, out *Swimlane) *FtileWithSwimlanes {
	d := &FtileWithSwimlanes{in: in, out: out}
	d.initDecorate(d, ftile)
	return d
}

// SwimlaneIn implements Swimable.
func (d *FtileWithSwimlanes) SwimlaneIn() *Swimlane { return d.in }

// SwimlaneOut implements Swimable.
func (d *FtileWithSwimlanes) SwimlaneOut() *Swimlane { return d.out }

// FtileMargedVertically adds margins above and below a tile.
type FtileMargedVertically struct {
	FtileDecorate
	margin1, margin2 float64
	cached           *FtileGeometry
}

// NewFtileMargedVertically returns ftile with margins above and below.
func NewFtileMargedVertically(ftile Ftile, margin1, margin2 float64) *FtileMargedVertically {
	d := &FtileMargedVertically{margin1: margin1, margin2: margin2}
	d.initDecorate(d, ftile)
	return d
}

// DrawU implements klimt.TextBlock.
func (d *FtileMargedVertically) DrawU(ug klimt.UGraphic) {
	if d.margin1 > 0 {
		ug = ug.Apply(klimt.UTranslateDy(d.margin1))
	}
	ug.Draw(d.ftile)
}

// Geometry implements Ftile.
func (d *FtileMargedVertically) Geometry(stringBounder font.StringBounder) FtileGeometry {
	if d.cached == nil {
		g := d.FtileDecorate.Geometry(stringBounder).IncVertically(d.margin1, d.margin2)
		d.cached = &g
	}
	return *d.cached
}

// FtileMargedWest adds a margin on the left of a tile.
type FtileMargedWest struct {
	FtileDecorate
	margin float64
}

// NewFtileMargedWest returns ftile with a margin on the left.
func NewFtileMargedWest(ftile Ftile, margin float64) *FtileMargedWest {
	d := &FtileMargedWest{margin: margin}
	d.initDecorate(d, ftile)
	return d
}

// DrawU implements klimt.TextBlock.
func (d *FtileMargedWest) DrawU(ug klimt.UGraphic) {
	d.FtileDecorate.DrawU(ug.Apply(klimt.UTranslateDx(d.margin)))
}

// Geometry implements Ftile.
func (d *FtileMargedWest) Geometry(stringBounder font.StringBounder) FtileGeometry {
	return d.FtileDecorate.Geometry(stringBounder).IncLeft(d.margin)
}

// FtileMinWidthCentered centers a tile in a minimum width.
type FtileMinWidthCentered struct {
	FtileDecorate
	minWidth                   float64
	calculateDimensionInternal *FtileGeometry
}

// NewFtileMinWidthCentered returns ftile centered in minWidth.
func NewFtileMinWidthCentered(ftile Ftile, minWidth float64) *FtileMinWidthCentered {
	d := &FtileMinWidthCentered{minWidth: minWidth}
	d.initDecorate(d, ftile)
	return d
}

// DrawU implements klimt.TextBlock.
func (d *FtileMinWidthCentered) DrawU(ug klimt.UGraphic) {
	change := d.uTranslateInternal(ug.StringBounder())
	d.FtileDecorate.DrawU(ug.Apply(change))
}

// Geometry implements Ftile.
func (d *FtileMinWidthCentered) Geometry(stringBounder font.StringBounder) FtileGeometry {
	if d.calculateDimensionInternal == nil {
		g := d.calculateDimensionSlow(stringBounder)
		d.calculateDimensionInternal = &g
	}
	return *d.calculateDimensionInternal
}

func (d *FtileMinWidthCentered) calculateDimensionSlow(stringBounder font.StringBounder) FtileGeometry {
	g := d.FtileDecorate.Geometry(stringBounder)
	left := d.point2(g.Left(), stringBounder)
	if !g.HasPointOut() {
		return NewFtileGeometryDim(d.dimensionInternal(stringBounder), left, g.InY())
	}
	return NewFtileGeometryDimOut(d.dimensionInternal(stringBounder), left, g.InY(), g.OutY())
}

func (d *FtileMinWidthCentered) dimensionInternal(stringBounder font.StringBounder) geom.XDimension2D {
	dim := d.ftile.CalculateDimension(stringBounder)
	if dim.Width() < d.minWidth {
		return geom.NewXDimension2D(d.minWidth, dim.Height())
	}
	return dim
}

func (d *FtileMinWidthCentered) uTranslateInternal(stringBounder font.StringBounder) klimt.UTranslate {
	dimTile := d.ftile.CalculateDimension(stringBounder)
	dimTotal := d.dimensionInternal(stringBounder)
	return klimt.UTranslateDx((dimTotal.Width() - dimTile.Width()) / 2)
}

// TranslateFor implements Ftile: nil (upstream's null) for a tile that is
// not the wrapped one.
func (d *FtileMinWidthCentered) TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	if child == d.ftile {
		t := d.uTranslateInternal(stringBounder)
		return &t
	}
	return nil
}

func (d *FtileMinWidthCentered) point2(x float64, stringBounder font.StringBounder) float64 {
	dim := d.ftile.CalculateDimension(stringBounder)
	if dim.Width() < d.minWidth {
		diff := d.minWidth - dim.Width()
		return x + diff/2
	}
	return x
}
