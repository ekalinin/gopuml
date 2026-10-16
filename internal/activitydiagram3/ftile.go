// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Ftile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/AbstractFtile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileGeometry.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileGeometryMerger.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Connection.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/AbstractConnection.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/ConnectionTranslatable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/WeldingPoint.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/MergeStrategy.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/ConnectionCross.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// Ftile is a tile of the activity layout: a text block with an entry point
// at the top and, unless the flow stops inside, an exit point at the bottom.
//
// Upstream's calculateDimension returns an FtileGeometry, a subclass of
// XDimension2D; here Geometry returns it and CalculateDimension (the text
// block method) returns its size.
type Ftile interface {
	klimt.TextBlock
	Swimable
	SkinParam() klimt.ISkinParam
	InLinkRendering() *LinkRendering
	OutLinkRendering() *LinkRendering
	// Geometry is upstream's calculateDimension.
	Geometry(stringBounder font.StringBounder) FtileGeometry
	// TranslateFor returns the position of a descendant tile; nil is
	// upstream's null.
	TranslateFor(child Ftile, stringBounder font.StringBounder) *klimt.UTranslate
	MyChildren() []Ftile
	InnerConnections() []Connection
	WeldingPoints() []WeldingPoint
	ArrowHorizontalAlignment() geom.HorizontalAlignment
}

// ftileDimensionHook is upstream's abstract calculateDimensionFtile of
// AbstractFtile, which concrete tiles define.
type ftileDimensionHook interface {
	calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry
}

// AbstractFtile holds the defaults of upstream's AbstractFtile. A tile
// embeds it and calls initFtile with itself, so that the cached geometry
// reaches its calculateDimensionFtile.
type AbstractFtile struct {
	self           Ftile
	skinParam      klimt.ISkinParam
	cachedGeometry *FtileGeometry
}

func (a *AbstractFtile) initFtile(self Ftile, skinParam klimt.ISkinParam) {
	a.self = self
	a.skinParam = skinParam
}

// SkinParam implements Ftile; it panics without skin parameters, as
// upstream's IllegalStateException.
func (a *AbstractFtile) SkinParam() klimt.ISkinParam {
	if a.skinParam == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	return a.skinParam
}

// IHtmlColorSet returns the color set of the skin parameters.
func (a *AbstractFtile) IHtmlColorSet() *color.HColorSet { return a.skinParam.IHtmlColorSet() }

// InLinkRendering implements Ftile: no arrow.
func (a *AbstractFtile) InLinkRendering() *LinkRendering { return LinkRenderingNone() }

// OutLinkRendering implements Ftile: no arrow.
func (a *AbstractFtile) OutLinkRendering() *LinkRendering { return LinkRenderingNone() }

// InnerConnections implements Ftile: none.
func (a *AbstractFtile) InnerConnections() []Connection { return nil }

// TranslateFor implements Ftile: it panics, as upstream.
func (a *AbstractFtile) TranslateFor(Ftile, font.StringBounder) *klimt.UTranslate {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// WeldingPoints implements Ftile: none.
func (a *AbstractFtile) WeldingPoints() []WeldingPoint { return nil }

// MyChildren implements Ftile: it panics, as upstream.
func (a *AbstractFtile) MyChildren() []Ftile {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// ArrowHorizontalAlignment implements Ftile: the arrow message alignment.
func (a *AbstractFtile) ArrowHorizontalAlignment() geom.HorizontalAlignment {
	return a.skinParam.HorizontalAlignment(klimt.AlignmentParamArrowMessageAlignment, klimt.ArrowDirectionNull,
		false, klimt.HorizontalAlignmentNull)
}

// Geometry implements Ftile: the cached result of calculateDimensionFtile.
func (a *AbstractFtile) Geometry(stringBounder font.StringBounder) FtileGeometry {
	if a.cachedGeometry == nil {
		g := a.self.(ftileDimensionHook).calculateDimensionFtile(stringBounder)
		a.cachedGeometry = &g
	}
	return *a.cachedGeometry
}

func (a *AbstractFtile) invalidateGeometryCache() { a.cachedGeometry = nil }

// CalculateDimension implements klimt.TextBlock: the size of the geometry.
func (a *AbstractFtile) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.self.Geometry(stringBounder).Dim()
}

// MinMax implements klimt.TextBlock: it panics, as upstream.
func (a *AbstractFtile) MinMax(font.StringBounder) geom.MinMax {
	uml.Throw(uml.UnsupportedOperationException, "")
	return geom.MinMax{}
}

// InnerPosition implements klimt.TextBlock: it panics, as upstream's
// default.
func (a *AbstractFtile) InnerPosition(member string, _ font.StringBounder) *geom.XRectangle2D {
	return klimt.TextBlockDefaults{}.InnerPosition(member, nil)
}

// MagneticBorder implements klimt.TextBlock.
func (a *AbstractFtile) MagneticBorder() klimt.MagneticBorder { return klimt.MagneticBorderNone{} }

// Backcolor implements klimt.TextBlock: none.
func (a *AbstractFtile) Backcolor() color.HColor { return nil }

// minNormal is Java's Double.MIN_NORMAL, the "no exit point" marker of
// FtileGeometry.
const minNormal = 2.2250738585072014e-308

// FtileGeometry is the size of a tile with its entry point (left, inY) and
// exit point (left, outY); outY is minNormal when there is no exit point.
type FtileGeometry struct {
	width, height float64
	left          float64
	inY           float64
	outY          float64
}

// NewFtileGeometry returns a geometry without exit point.
func NewFtileGeometry(width, height, left, inY float64) FtileGeometry {
	return FtileGeometry{width: width, height: height, left: left, inY: inY, outY: minNormal}
}

// NewFtileGeometryOut returns a geometry with an exit point.
func NewFtileGeometryOut(width, height, left, inY, outY float64) FtileGeometry {
	return FtileGeometry{width: width, height: height, left: left, inY: inY, outY: outY}
}

// NewFtileGeometryDim returns a geometry of size dim without exit point.
func NewFtileGeometryDim(dim geom.XDimension2D, left, inY float64) FtileGeometry {
	return NewFtileGeometry(dim.Width(), dim.Height(), left, inY)
}

// NewFtileGeometryDimOut returns a geometry of size dim with an exit point.
func NewFtileGeometryDimOut(dim geom.XDimension2D, left, inY, outY float64) FtileGeometry {
	return NewFtileGeometryOut(dim.Width(), dim.Height(), left, inY, outY)
}

// Width returns the width.
func (g FtileGeometry) Width() float64 { return g.width }

// Height returns the height.
func (g FtileGeometry) Height() float64 { return g.height }

// Dim returns the size.
func (g FtileGeometry) Dim() geom.XDimension2D { return geom.NewXDimension2D(g.width, g.height) }

// String is upstream's toString.
func (g FtileGeometry) String() string {
	return "[" + jfmt.DoubleToString(g.width) + "x" + jfmt.DoubleToString(g.height) + " left=" +
		jfmt.DoubleToString(g.left) + "]"
}

// PointA returns the entry point.
func (g FtileGeometry) PointA() geom.XPoint2D { return geom.NewXPoint2D(g.left, g.inY) }

// PointIn returns the entry point.
func (g FtileGeometry) PointIn() geom.XPoint2D { return geom.NewXPoint2D(g.left, g.inY) }

func (g FtileGeometry) checkOut() {
	if g.outY == minNormal {
		uml.Throw(uml.UnsupportedOperationException, "")
	}
}

// PointB returns the middle of the right side.
func (g FtileGeometry) PointB() geom.XPoint2D {
	g.checkOut()
	return geom.NewXPoint2D(g.width, (g.inY+g.outY)/2)
}

// PointC returns the exit point.
func (g FtileGeometry) PointC() geom.XPoint2D {
	g.checkOut()
	return geom.NewXPoint2D(g.left, g.outY)
}

// PointD returns the middle of the left side.
func (g FtileGeometry) PointD() geom.XPoint2D {
	g.checkOut()
	return geom.NewXPoint2D(0, (g.inY+g.outY)/2)
}

// PointOut returns the exit point; it panics without one, as upstream.
func (g FtileGeometry) PointOut() geom.XPoint2D {
	g.checkOut()
	return geom.NewXPoint2D(g.left, g.outY)
}

// IncHeight returns the geometry with more height below.
func (g FtileGeometry) IncHeight(northHeight float64) FtileGeometry {
	return NewFtileGeometryOut(g.width, g.height+northHeight, g.left, g.inY, g.outY)
}

// AddTop returns the geometry with room added above.
func (g FtileGeometry) AddTop(northHeight float64) FtileGeometry {
	if g.HasPointOut() {
		return NewFtileGeometryOut(g.width, g.height+northHeight, g.left, g.inY+northHeight, g.outY+northHeight)
	}
	return NewFtileGeometryOut(g.width, g.height+northHeight, g.left, g.inY+northHeight, minNormal)
}

// AddBottom returns the geometry with room added below.
func (g FtileGeometry) AddBottom(southHeight float64) FtileGeometry {
	if g.HasPointOut() {
		return NewFtileGeometryOut(g.width, g.height+southHeight, g.left, g.inY, g.outY)
	}
	return NewFtileGeometryOut(g.width, g.height+southHeight, g.left, g.inY, minNormal)
}

// IncRight returns the geometry with room added on the right.
func (g FtileGeometry) IncRight(missing float64) FtileGeometry {
	return NewFtileGeometryOut(g.width+missing, g.height, g.left, g.inY, g.outY)
}

// IncLeft returns the geometry with room added on the left.
func (g FtileGeometry) IncLeft(missing float64) FtileGeometry {
	return NewFtileGeometryOut(g.width+missing, g.height, g.left+missing, g.inY, g.outY)
}

// IncVertically returns the geometry with room added above and below.
func (g FtileGeometry) IncVertically(missing1, missing2 float64) FtileGeometry {
	outY := g.outY
	if g.HasPointOut() {
		outY = g.outY + missing1
	}
	return NewFtileGeometryOut(g.width, g.height+missing1+missing2, g.left, g.inY+missing1, outY)
}

// IncInY returns the geometry with the entry point moved down.
func (g FtileGeometry) IncInY(missing float64) FtileGeometry {
	return NewFtileGeometryOut(g.width, g.height, g.left, g.inY+missing, g.outY)
}

// HasPointOut reports whether there is an exit point.
func (g FtileGeometry) HasPointOut() bool { return g.outY != minNormal }

// WithoutPointOut returns the geometry without exit point.
func (g FtileGeometry) WithoutPointOut() FtileGeometry {
	return NewFtileGeometry(g.width, g.height, g.left, g.inY)
}

// Translate returns the geometry with its points moved.
func (g FtileGeometry) Translate(translate klimt.UTranslate) FtileGeometry {
	dx := translate.Dx()
	dy := translate.Dy()
	if g.outY == minNormal {
		return NewFtileGeometry(g.width, g.height, g.left+dx, g.inY+dy)
	}
	return NewFtileGeometryOut(g.width, g.height, g.left+dx, g.inY+dy, g.outY+dy)
}

// InY returns the ordinate of the entry point.
func (g FtileGeometry) InY() float64 { return g.inY }

// Left returns the abscissa of the entry and exit points.
func (g FtileGeometry) Left() float64 { return g.left }

// Right returns the width right of the entry point.
func (g FtileGeometry) Right() float64 { return g.width - g.left }

// OutY returns the ordinate of the exit point (minNormal for none).
func (g FtileGeometry) OutY() float64 { return g.outY }

// AddDim returns the geometry enlarged, the exit point moved down.
func (g FtileGeometry) AddDim(deltaWidth, deltaHeight float64) FtileGeometry {
	return NewFtileGeometryOut(g.width+deltaWidth, g.height+deltaHeight, g.left, g.inY, g.outY+deltaHeight)
}

// AddMarginX returns the geometry with marginx on both sides.
func (g FtileGeometry) AddMarginX(marginx float64) FtileGeometry {
	return NewFtileGeometryOut(g.width+2*marginx, g.height, g.left+marginx, g.inY, g.outY)
}

// AddMarginX2 returns the geometry with margins on the left and right.
func (g FtileGeometry) AddMarginX2(margin1, margin2 float64) FtileGeometry {
	return NewFtileGeometryOut(g.width+margin1+margin2, g.height, g.left+margin1, g.inY, g.outY)
}

// FixedHeight returns the geometry with another height.
func (g FtileGeometry) FixedHeight(fixedHeight float64) FtileGeometry {
	return NewFtileGeometryOut(g.width, fixedHeight, g.left, g.inY, g.outY)
}

// AppendBottom returns the geometry of other placed below, upstream's
// FtileGeometryMerger.
func (g FtileGeometry) AppendBottom(other FtileGeometry) FtileGeometry {
	left := max(g.Left(), other.Left())
	dx1 := left - g.Left()
	dx2 := left - other.Left()
	width := max(g.Width()+dx1, other.Width()+dx2)
	height := g.Height() + other.Height()
	if other.HasPointOut() {
		return NewFtileGeometryOut(width, height, left, g.InY(), other.OutY()+g.Height())
	}
	return NewFtileGeometry(width, height, left, g.InY())
}

// EnsureHeight returns the geometry at least newHeight high.
func (g FtileGeometry) EnsureHeight(newHeight float64) FtileGeometry {
	if g.Height() > newHeight {
		return g
	}
	return g.FixedHeight(newHeight)
}

// Connection is an arrow between two tiles, drawn as a shape so that the
// lane interceptors can select it.
type Connection interface {
	klimt.UDrawable
	Ftile1() Ftile
	Ftile2() Ftile
}

// ConnectionTranslatable is a connection that can be drawn across lanes.
type ConnectionTranslatable interface {
	Connection
	DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate)
}

// AbstractConnection holds the two tiles of a connection; either may be
// nil.
type AbstractConnection struct {
	ftile1, ftile2 Ftile
}

func newAbstractConnection(ftile1, ftile2 Ftile) AbstractConnection {
	return AbstractConnection{ftile1: ftile1, ftile2: ftile2}
}

// Ftile1 implements Connection.
func (c *AbstractConnection) Ftile1() Ftile { return c.ftile1 }

// Ftile2 implements Connection.
func (c *AbstractConnection) Ftile2() Ftile { return c.ftile2 }

// ArrowHorizontalAlignment returns the alignment of the first tile, else
// of the second, else left.
func (c *AbstractConnection) ArrowHorizontalAlignment() geom.HorizontalAlignment {
	if c.ftile1 != nil {
		return c.ftile1.ArrowHorizontalAlignment()
	}
	if c.ftile2 != nil {
		return c.ftile2.ArrowHorizontalAlignment()
	}
	return geom.HorizontalAlignmentLeft
}

// ConnectionCross draws a connection between tiles of different lanes,
// with the translations of the lanes.
type ConnectionCross struct {
	AbstractConnection
	connection Connection
}

// NewConnectionCross returns the cross-lane drawing of connection.
func NewConnectionCross(connection Connection) *ConnectionCross {
	return &ConnectionCross{AbstractConnection: newAbstractConnection(connection.Ftile1(), connection.Ftile2()),
		connection: connection}
}

// DrawU implements klimt.UDrawable.
func (c *ConnectionCross) DrawU(ug klimt.UGraphic) {
	conn, ok := c.connection.(ConnectionTranslatable)
	if !ok {
		return
	}
	swimlane1 := c.Ftile1().SwimlaneOut()
	swimlane2 := c.Ftile2().SwimlaneIn()
	if swimlane1 == nil || swimlane2 == nil {
		return
	}
	conn.DrawTranslate(ug, swimlane1.Translate(), swimlane2.Translate())
}

// WeldingPoint marks a point where a "break" leaves a loop.
type WeldingPoint interface {
	isWeldingPoint()
}

// MergeStrategy tells how snakes merge.
type MergeStrategy int

// Merge strategies, in upstream's order.
const (
	MergeFull MergeStrategy = iota
	MergeLimited
	MergeNone
)

// Max returns the stronger of the strategies.
func (m MergeStrategy) Max(other MergeStrategy) MergeStrategy { return max(m, other) }

// FtileFactory builds the tiles of the instructions; the factories are
// chained (FtileFactoryDelegator), each adding one feature.
type FtileFactory interface {
	StringBounder() font.StringBounder
	SkinParam() klimt.ISkinParam
	Start(swimlane *Swimlane, colors *color.Colors) Ftile
	Stop(swimlane *Swimlane, colors *color.Colors) Ftile
	End(swimlane *Swimlane, colors *color.Colors) Ftile
	Spot(swimlane *Swimlane, spot string, c color.HColor) Ftile
	Activity(label *klimt.Display, swimlane *Swimlane, style BoxStyle, colors *color.Colors,
		stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) Ftile
	AddNote(ftile Ftile, swimlane *Swimlane, notes []*PositionedNote, verticalAlignment geom.VerticalAlignment) Ftile
	AddURL(ftile Ftile, url *klimt.URL) Ftile
	DecorateIn(ftile Ftile, linkRendering *LinkRendering) Ftile
	DecorateOut(ftile Ftile, linkRendering *LinkRendering) Ftile
	Assembly(tile1, tile2 Ftile) Ftile
	Repeat(stereotype, stereotype2 *klimt.Stereogroup, boxStyleIn BoxStyle, swimlane, swimlaneOut *Swimlane,
		startLabel *klimt.Display, repeat Ftile, test, yes, out *klimt.Display, backward Ftile, noOut bool,
		incoming1, incoming2 *LinkRendering, currentStyleBuilder *klimt.StyleBuilder) Ftile
	CreateWhile(outColor *LinkRendering, swimlane *Swimlane, whileBlock Ftile, test, yes *klimt.Display,
		c color.HColor, specialOut Instruction, backward Ftile, incoming1, incoming2 *LinkRendering,
		currentStyleBuilder *klimt.StyleBuilder) Ftile
	CreateIf(swimlane *Swimlane, thens []*Branch, elseBranch *Branch, outColor, topInlinkRendering *LinkRendering,
		url *klimt.URL, notes []*PositionedNote, stereotype *klimt.Stereotype,
		currentStyleBuilder *klimt.StyleBuilder) Ftile
	CreateSwitch(swimlane *Swimlane, branches []*Branch, afterEndwhile, topInlinkRendering *LinkRendering,
		labelTest *klimt.Display, colors, endColors *color.Colors) Ftile
	CreateParallel(all []Ftile, style ForkStyle, label string, hasLabel bool, in, out *Swimlane,
		colors *color.Colors) Ftile
	CreateGroup(list Ftile, name *klimt.Display, backColor color.HColor, note *PositionedNote, typ GroupSymbol,
		style *klimt.Style) Ftile
}
