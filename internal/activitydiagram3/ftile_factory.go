// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/FtileFactoryDelegator.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/VCompactFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorAddUrl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorAssembly.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/ConnectionVerticalDown.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileForkInner.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// FtileFactoryDelegator forwards every method to the next factory of the
// chain; the delegators embed it and override the methods of their
// feature.
type FtileFactoryDelegator struct {
	factory FtileFactory
}

func newFtileFactoryDelegator(factory FtileFactory) FtileFactoryDelegator {
	return FtileFactoryDelegator{factory: factory}
}

// DefaultStyleDefinitionActivity is the style signature of actions.
func (d *FtileFactoryDelegator) DefaultStyleDefinitionActivity() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity)
}

// DefaultStyleDefinitionDiamond is the style signature of diamonds.
func (d *FtileFactoryDelegator) DefaultStyleDefinitionDiamond() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity, klimt.SNameDiamond)
}

// DefaultStyleDefinitionArrow is the style signature of arrows.
func (d *FtileFactoryDelegator) DefaultStyleDefinitionArrow() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameArrow)
}

// InLinkRenderingColor is getInLinkRenderingColor: the colors of the arrow
// into tile, the arrow style for none.
func (d *FtileFactoryDelegator) InLinkRenderingColor(tile Ftile) *Rainbow {
	linkRendering := tile.InLinkRendering()
	if linkRendering == nil || linkRendering.Rainbow().Size() == 0 {
		style := d.DefaultStyleDefinitionArrow().MergedStyle(d.SkinParam().CurrentStyleBuilder())
		return RainbowFromStyle(style, d.SkinParam().IHtmlColorSet())
	}
	return linkRendering.Rainbow()
}

// TextBlock is getTextBlock: the label of an arrow, nil for the null
// display.
func (d *FtileFactoryDelegator) TextBlock(display *klimt.Display) klimt.TextBlock {
	// DUP3945
	if klimt.DisplayIsNull(display) {
		return nil
	}
	style := d.DefaultStyleDefinitionArrow().MergedStyle(d.SkinParam().CurrentStyleBuilder())
	fontConfiguration := style.FontConfiguration(d.SkinParam().IHtmlColorSet(), nil)
	return display.Create7(fontConfiguration, geom.HorizontalAlignmentLeft, d.SkinParam(),
		klimt.CreoleModeSimpleLine)
}

// InLinkRenderingDisplay is getInLinkRenderingDisplay: the label of the
// arrow into tile.
func (d *FtileFactoryDelegator) InLinkRenderingDisplay(tile Ftile) *klimt.Display {
	linkRendering := tile.InLinkRendering()
	if linkRendering == nil {
		return klimt.DisplayNULL()
	}
	return linkRendering.Display()
}

// Start implements FtileFactory.
func (d *FtileFactoryDelegator) Start(swimlane *Swimlane, colors *color.Colors) Ftile {
	return d.factory.Start(swimlane, colors)
}

// End implements FtileFactory.
func (d *FtileFactoryDelegator) End(swimlane *Swimlane, colors *color.Colors) Ftile {
	return d.factory.End(swimlane, colors)
}

// Stop implements FtileFactory.
func (d *FtileFactoryDelegator) Stop(swimlane *Swimlane, colors *color.Colors) Ftile {
	return d.factory.Stop(swimlane, colors)
}

// Spot implements FtileFactory.
func (d *FtileFactoryDelegator) Spot(swimlane *Swimlane, spot string, c color.HColor) Ftile {
	return d.factory.Spot(swimlane, spot, c)
}

// Activity implements FtileFactory.
func (d *FtileFactoryDelegator) Activity(label *klimt.Display, swimlane *Swimlane, style BoxStyle,
	colors *color.Colors, stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) Ftile {
	return d.factory.Activity(label, swimlane, style, colors, stereotype, styleBuilder)
}

// AddNote implements FtileFactory.
func (d *FtileFactoryDelegator) AddNote(ftile Ftile, swimlane *Swimlane, notes []*PositionedNote,
	verticalAlignment geom.VerticalAlignment) Ftile {
	return d.factory.AddNote(ftile, swimlane, notes, verticalAlignment)
}

// AddURL implements FtileFactory.
func (d *FtileFactoryDelegator) AddURL(ftile Ftile, url *klimt.URL) Ftile {
	return d.factory.AddURL(ftile, url)
}

// DecorateIn implements FtileFactory.
func (d *FtileFactoryDelegator) DecorateIn(ftile Ftile, linkRendering *LinkRendering) Ftile {
	return d.factory.DecorateIn(ftile, requireLink(linkRendering))
}

// DecorateOut implements FtileFactory.
func (d *FtileFactoryDelegator) DecorateOut(ftile Ftile, linkRendering *LinkRendering) Ftile {
	return d.factory.DecorateOut(ftile, requireLink(linkRendering))
}

// Assembly implements FtileFactory.
func (d *FtileFactoryDelegator) Assembly(tile1, tile2 Ftile) Ftile {
	return d.factory.Assembly(tile1, tile2)
}

// Repeat implements FtileFactory.
func (d *FtileFactoryDelegator) Repeat(stereotype, stereotype2 *klimt.Stereogroup, boxStyleIn BoxStyle,
	swimlane, swimlaneOut *Swimlane, startLabel *klimt.Display, repeat Ftile, test, yes, out *klimt.Display,
	backward Ftile, noOut bool, incoming1, incoming2 *LinkRendering, currentStyleBuilder *klimt.StyleBuilder) Ftile {
	return d.factory.Repeat(stereotype, stereotype2, boxStyleIn, swimlane, swimlaneOut, startLabel, repeat, test, yes,
		out, backward, noOut, incoming1, incoming2, currentStyleBuilder)
}

// CreateWhile implements FtileFactory.
func (d *FtileFactoryDelegator) CreateWhile(outColor *LinkRendering, swimlane *Swimlane, whileBlock Ftile,
	test, yes *klimt.Display, c color.HColor, specialOut Instruction, back Ftile, incoming1,
	incoming2 *LinkRendering, styleBuilder *klimt.StyleBuilder) Ftile {
	return d.factory.CreateWhile(outColor, swimlane, whileBlock, test, yes, c, specialOut, back, incoming1, incoming2,
		styleBuilder)
}

// CreateIf implements FtileFactory.
func (d *FtileFactoryDelegator) CreateIf(swimlane *Swimlane, thens []*Branch, elseBranch *Branch,
	afterEndwhile, topInlinkRendering *LinkRendering, url *klimt.URL, notes []*PositionedNote,
	stereotype *klimt.Stereotype, currentStyleBuilder *klimt.StyleBuilder) Ftile {
	return d.factory.CreateIf(swimlane, thens, elseBranch, afterEndwhile, topInlinkRendering, url, notes, stereotype,
		currentStyleBuilder)
}

// CreateSwitch implements FtileFactory.
func (d *FtileFactoryDelegator) CreateSwitch(swimlane *Swimlane, branches []*Branch, afterEndwhile,
	topInlinkRendering *LinkRendering, labelTest *klimt.Display, colors, endColors *color.Colors) Ftile {
	return d.factory.CreateSwitch(swimlane, branches, afterEndwhile, topInlinkRendering, labelTest, colors, endColors)
}

// CreateParallel implements FtileFactory.
func (d *FtileFactoryDelegator) CreateParallel(all []Ftile, style ForkStyle, label string, hasLabel bool,
	in, out *Swimlane, colors *color.Colors) Ftile {
	return d.factory.CreateParallel(all, style, label, hasLabel, in, out, colors)
}

// CreateGroup implements FtileFactory.
func (d *FtileFactoryDelegator) CreateGroup(list Ftile, name *klimt.Display, backColor color.HColor,
	note *PositionedNote, typ GroupSymbol, style *klimt.Style) Ftile {
	return d.factory.CreateGroup(list, name, backColor, note, typ, style)
}

// StringBounder implements FtileFactory.
func (d *FtileFactoryDelegator) StringBounder() font.StringBounder { return d.factory.StringBounder() }

// SkinParam implements FtileFactory.
func (d *FtileFactoryDelegator) SkinParam() klimt.ISkinParam { return d.factory.SkinParam() }

// Factory returns the next factory of the chain.
func (d *FtileFactoryDelegator) Factory() FtileFactory { return d.factory }

// VCompactFactory is the end of the factory chain: it builds the simple
// tiles and stacks the others.
type VCompactFactory struct {
	skinParam     klimt.ISkinParam
	stringBounder font.StringBounder
}

var _ FtileFactory = (*VCompactFactory)(nil)

// NewVCompactFactory returns the base factory.
func NewVCompactFactory(skinParam klimt.ISkinParam, stringBounder font.StringBounder) *VCompactFactory {
	return &VCompactFactory{skinParam: skinParam, stringBounder: stringBounder}
}

// StringBounder implements FtileFactory.
func (f *VCompactFactory) StringBounder() font.StringBounder { return f.stringBounder }

// CurrentStyleBuilder returns the style builder of the skin parameters.
func (f *VCompactFactory) CurrentStyleBuilder() *klimt.StyleBuilder {
	return f.skinParam.CurrentStyleBuilder()
}

func vcompactCircleSignature(name klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameCircle, name)
}

// Start implements FtileFactory.
func (f *VCompactFactory) Start(swimlane *Swimlane, colors *color.Colors) Ftile {
	style := vcompactCircleSignature(klimt.SNameStart).MergedStyle(f.skinParam.CurrentStyleBuilder())
	return NewFtileCircleStart(f.SkinParam(), swimlane, style, colors)
}

// Stop implements FtileFactory.
func (f *VCompactFactory) Stop(swimlane *Swimlane, colors *color.Colors) Ftile {
	style := vcompactCircleSignature(klimt.SNameStop).MergedStyle(f.skinParam.CurrentStyleBuilder())
	return NewFtileCircleStop(f.SkinParam(), swimlane, style, colors)
}

// Spot implements FtileFactory.
func (f *VCompactFactory) Spot(swimlane *Swimlane, spot string, c color.HColor) Ftile {
	uFont := f.skinParam.Font(nil, false, klimt.FontParamActivity)
	style := vcompactCircleSignature(klimt.SNameSpot).MergedStyle(f.skinParam.CurrentStyleBuilder())
	return NewFtileCircleSpot(f.SkinParam(), swimlane, spot, uFont, c, style)
}

// End implements FtileFactory.
func (f *VCompactFactory) End(swimlane *Swimlane, colors *color.Colors) Ftile {
	style := vcompactCircleSignature(klimt.SNameEnd).MergedStyle(f.skinParam.CurrentStyleBuilder())
	return NewFtileCircleEndCross(f.SkinParam(), swimlane, style, colors)
}

// Activity implements FtileFactory.
func (f *VCompactFactory) Activity(label *klimt.Display, swimlane *Swimlane, boxStyle BoxStyle,
	colors *color.Colors, stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) Ftile {
	if stereotype != nil && stereotype.IsIcon() {
		return FtileBoxEmojiCreate(klimt.ColorsMute(colors, f.skinParam), label, swimlane, stereotype, styleBuilder)
	}
	return FtileBoxCreate(klimt.ColorsMute(colors, f.skinParam), label, swimlane, boxStyle, stereotype, styleBuilder)
}

// AddNote implements FtileFactory: the tile unchanged.
func (f *VCompactFactory) AddNote(ftile Ftile, _ *Swimlane, _ []*PositionedNote, _ geom.VerticalAlignment) Ftile {
	return ftile
}

// AddURL implements FtileFactory: the tile unchanged.
func (f *VCompactFactory) AddURL(ftile Ftile, _ *klimt.URL) Ftile { return ftile }

// Assembly implements FtileFactory.
func (f *VCompactFactory) Assembly(tile1, tile2 Ftile) Ftile {
	return NewFtileAssemblySimple(tile1, tile2)
}

// Repeat implements FtileFactory: the inner tile.
func (f *VCompactFactory) Repeat(_, _ *klimt.Stereogroup, _ BoxStyle, _, _ *Swimlane, _ *klimt.Display,
	repeat Ftile, _, _, _ *klimt.Display, _ Ftile, _ bool, _, _ *LinkRendering, _ *klimt.StyleBuilder) Ftile {
	return repeat
}

// CreateWhile implements FtileFactory: the inner tile.
func (f *VCompactFactory) CreateWhile(_ *LinkRendering, _ *Swimlane, whileBlock Ftile, _, _ *klimt.Display,
	_ color.HColor, _ Instruction, _ Ftile, _, _ *LinkRendering, _ *klimt.StyleBuilder) Ftile {
	return whileBlock
}

// CreateIf implements FtileFactory: the branches side by side.
func (f *VCompactFactory) CreateIf(_ *Swimlane, thens []*Branch, elseBranch *Branch, _, _ *LinkRendering,
	_ *klimt.URL, _ []*PositionedNote, _ *klimt.Stereotype, _ *klimt.StyleBuilder) Ftile {
	var ftiles []Ftile
	for _, branch := range thens {
		ftiles = append(ftiles, branch.Ftile())
	}
	ftiles = append(ftiles, elseBranch.Ftile())
	return NewFtileForkInner(ftiles)
}

// CreateSwitch implements FtileFactory: the branches side by side.
func (f *VCompactFactory) CreateSwitch(_ *Swimlane, branches []*Branch, _, _ *LinkRendering, _ *klimt.Display,
	_, _ *color.Colors) Ftile {
	var ftiles []Ftile
	for _, branch := range branches {
		ftiles = append(ftiles, branch.Ftile())
	}
	return NewFtileForkInner(ftiles)
}

// CreateParallel implements FtileFactory: the branches side by side.
func (f *VCompactFactory) CreateParallel(all []Ftile, _ ForkStyle, _ string, _ bool, _, _ *Swimlane,
	_ *color.Colors) Ftile {
	return NewFtileForkInner(all)
}

// CreateGroup implements FtileFactory: the inner tile.
func (f *VCompactFactory) CreateGroup(list Ftile, _ *klimt.Display, _ color.HColor, _ *PositionedNote,
	_ GroupSymbol, _ *klimt.Style) Ftile {
	return list
}

// DecorateIn implements FtileFactory.
func (f *VCompactFactory) DecorateIn(ftile Ftile, linkRendering *LinkRendering) Ftile {
	return NewFtileDecorateIn(ftile, linkRendering)
}

// DecorateOut implements FtileFactory.
func (f *VCompactFactory) DecorateOut(ftile Ftile, linkRendering *LinkRendering) Ftile {
	return NewFtileDecorateOut(ftile, linkRendering)
}

// SkinParam implements FtileFactory.
func (f *VCompactFactory) SkinParam() klimt.ISkinParam { return f.skinParam }

// FtileFactoryDelegatorAddUrl links action boxes to their URL.
type FtileFactoryDelegatorAddUrl struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorAddUrl returns the delegator over factory.
func NewFtileFactoryDelegatorAddUrl(factory FtileFactory) *FtileFactoryDelegatorAddUrl {
	return &FtileFactoryDelegatorAddUrl{newFtileFactoryDelegator(factory)}
}

// AddURL implements FtileFactory.
func (d *FtileFactoryDelegatorAddUrl) AddURL(ftile Ftile, url *klimt.URL) Ftile {
	switch ftile.(type) {
	case *FtileBox, *FtileBoxEmoji:
		return NewFtileWithUrl(ftile, url)
	}
	return ftile
}

// FtileFactoryDelegatorAssembly stacks tiles with an arrow between them.
type FtileFactoryDelegatorAssembly struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorAssembly returns the delegator over factory.
func NewFtileFactoryDelegatorAssembly(factory FtileFactory) *FtileFactoryDelegatorAssembly {
	return &FtileFactoryDelegatorAssembly{newFtileFactoryDelegator(factory)}
}

// Assembly implements FtileFactory.
func (d *FtileFactoryDelegatorAssembly) Assembly(tile1, tile2 Ftile) Ftile {
	height := 35.0
	textBlock := d.TextBlock(d.InLinkRenderingDisplay(tile2))
	stringBounder := d.StringBounder()
	if textBlock != nil {
		height += textBlock.CalculateDimension(stringBounder).Height()
	}
	tile1andSpace := FtileUtilsAddBottom(tile1, height)
	result := d.FtileFactoryDelegator.Assembly(tile1andSpace, tile2)
	g := tile1.Geometry(stringBounder)
	if !g.HasPointOut() {
		return result
	}
	translate1 := result.TranslateFor(tile1andSpace, stringBounder)
	p1 := g.Translate(*translate1).PointOut()
	translate2 := result.TranslateFor(tile2, stringBounder)
	p2 := tile2.Geometry(stringBounder).Translate(*translate2).PointIn()

	c := d.InLinkRenderingColor(tile2)
	connection := NewConnectionVerticalDown(tile1, tile2, p1, p2, c, textBlock)
	result = FtileUtilsAddConnection(result, connection)
	if textBlock != nil {
		dim := result.Geometry(stringBounder)
		width := dim.Width()
		maxX := connection.MaxX(stringBounder)
		if width < maxX {
			result = NewFtileMargedRight(result, maxX)
		}
	}
	return result
}

// ConnectionVerticalDown is the straight arrow between two stacked tiles.
type ConnectionVerticalDown struct {
	AbstractConnection
	p1, p2    geom.XPoint2D
	color     *Rainbow
	textBlock klimt.TextBlock
}

// NewConnectionVerticalDown returns the arrow from p1 down to p2;
// textBlock may be nil.
func NewConnectionVerticalDown(ftile1, ftile2 Ftile, p1, p2 geom.XPoint2D, c *Rainbow,
	textBlock klimt.TextBlock) *ConnectionVerticalDown {
	if c.Size() == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &ConnectionVerticalDown{AbstractConnection: newAbstractConnection(ftile1, ftile2), p1: p1, p2: p2,
		color: c, textBlock: textBlock}
}

// DrawU implements klimt.UDrawable.
func (c *ConnectionVerticalDown) DrawU(ug klimt.UGraphic) { ug.Draw(c.simpleSnake()) }

// MaxX returns the right of the arrow and its label.
func (c *ConnectionVerticalDown) MaxX(stringBounder font.StringBounder) float64 {
	return c.simpleSnake().MaxX(stringBounder)
}

func (c *ConnectionVerticalDown) simpleSnake() *Snake {
	skinParam := c.Ftile1().SkinParam()
	snake := SnakeCreateEnd(skinParam, c.color, skinParam.Arrows().AsToDown()).WithLabel(c.textBlock,
		c.ArrowHorizontalAlignment())
	snake.AddPointPt(c.p1)
	snake.AddPointPt(c.p2)
	return snake
}

// DrawTranslate implements ConnectionTranslatable: the arrow between
// lanes, with a horizontal segment in the middle.
func (c *ConnectionVerticalDown) DrawTranslate(ug klimt.UGraphic, translate1, translate2 klimt.UTranslate) {
	skinParam := c.Ftile1().SkinParam()
	snake := SnakeCreateEnd(skinParam, c.color, skinParam.Arrows().AsToDown()).WithLabel(c.textBlock,
		c.ArrowHorizontalAlignment())
	mp1a := translate1.Translated(c.p1)
	mp2b := translate2.Translated(c.p2)
	middle := (mp1a.Y + mp2b.Y) / 2.0
	snake.AddPointPt(mp1a)
	snake.AddPoint(mp1a.X, middle)
	snake.AddPoint(mp2b.X, middle)
	snake.AddPointPt(mp2b)
	ug.Draw(snake)
}

// FtileForkInner places tiles side by side.
type FtileForkInner struct {
	AbstractFtile
	forks []Ftile
}

// NewFtileForkInner returns the tiles side by side.
func NewFtileForkInner(forks []Ftile) *FtileForkInner {
	f := &FtileForkInner{forks: append([]Ftile(nil), forks...)}
	f.initFtile(f, forks[0].SkinParam())
	return f
}

// MyChildren implements Ftile.
func (f *FtileForkInner) MyChildren() []Ftile { return f.forks }

// SwimlaneIn implements Swimable.
func (f *FtileForkInner) SwimlaneIn() *Swimlane { return f.forks[0].SwimlaneIn() }

// SwimlaneOut implements Swimable.
func (f *FtileForkInner) SwimlaneOut() *Swimlane { return f.SwimlaneIn() }

// Swimlanes implements Swimable.
func (f *FtileForkInner) Swimlanes() SwimlaneSet { return mergeSwimlanes(f.forks) }

// mergeSwimlanes is FtileForkInner.mergeSwimlanes: the lanes of tiles.
func mergeSwimlanes(tiles []Ftile) SwimlaneSet {
	var result SwimlaneSet
	for _, tile := range tiles {
		result = result.addAll(tile.Swimlanes())
	}
	return result
}

// DrawU implements klimt.TextBlock.
func (f *FtileForkInner) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	xpos := 0.0
	for _, ftile := range f.forks {
		ug.Apply(klimt.UTranslateDx(xpos)).Draw(ftile)
		xpos += ftile.CalculateDimension(stringBounder).Width()
	}
}

func (f *FtileForkInner) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	height := 0.0
	width := 0.0
	for _, ftile := range f.forks {
		dim := ftile.CalculateDimension(stringBounder)
		width += dim.Width()
		if dim.Height() > height {
			height = dim.Height()
		}
	}
	dimTotal := geom.NewXDimension2D(width, height)
	return NewFtileGeometryDimOut(dimTotal, dimTotal.Width()/2, 0, dimTotal.Height())
}

// TranslateFor implements Ftile.
func (f *FtileForkInner) TranslateFor(searched Ftile, stringBounder font.StringBounder) *klimt.UTranslate {
	xpos := 0.0
	for _, ftile := range f.forks {
		if ftile == searched {
			t := klimt.UTranslateDx(xpos)
			return &t
		}
		xpos += ftile.CalculateDimension(stringBounder).Width()
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return nil
}
