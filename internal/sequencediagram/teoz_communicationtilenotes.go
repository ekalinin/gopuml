// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteLeft.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteRight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteBottomTopAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteBottom.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteTop.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelfNoteLeft.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelfNoteRight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// The note decorators wrap the tile of a message with a note attached to
// it ("note left" after a message, and so on). Their gauge starts where
// the inner one does, propagates its contact line, and is anchored with
// AddAtLeast (see YGauge).

func noteDecoratorGauge(tile Tile, preferredHeight float64) *YGauge {
	innerGauge := tile.YGauge()
	return newYGaugeContact(innerGauge.Min(), innerGauge.Min().AddAtLeast(preferredHeight), innerGauge.Contact(),
		innerGauge.Origin())
}

// stringBounderOf is upstream's ((AbstractTile) tile).getStringBounder().
func stringBounderOf(tile Tile) font.StringBounder {
	return tile.(interface{ StringBounder() font.StringBounder }).StringBounder()
}

// CommunicationTileNoteLeft is a message with a note on its left.
type CommunicationTileNoteLeft struct {
	abstractTile
	tile          Tile
	message       AbstractMessage
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	livingSpace   *LivingSpace
	noteOnMessage *Note
	yGauge        *YGauge
}

// Event implements Tile.
func (t *CommunicationTileNoteLeft) Event() Event { return t.message }

// ContactPointRelative implements Tile.
func (t *CommunicationTileNoteLeft) ContactPointRelative() float64 {
	return t.tile.ContactPointRelative()
}

func newCommunicationTileNoteLeft(tile Tile, message AbstractMessage, skin *rose.Rose, skinParam klimt.ISkinParam,
	livingSpace *LivingSpace, noteOnMessage *Note, _ *YGauge) *CommunicationTileNoteLeft {
	t := &CommunicationTileNoteLeft{tile: tile, message: message, skin: skin, skinParam: skinParam,
		noteOnMessage: noteOnMessage, livingSpace: livingSpace}
	t.initTile(t, stringBounderOf(tile))
	t.yGauge = noteDecoratorGauge(tile, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *CommunicationTileNoteLeft) YGauge() *YGauge { return t.yGauge }

// OnGaugeResolved implements Tile.
func (t *CommunicationTileNoteLeft) OnGaugeResolved() { t.tile.OnGaugeResolved() }

func (t *CommunicationTileNoteLeft) component() klimt.Component {
	return t.skin.CreateComponentNote(t.noteOnMessage.UsedStyles(), noteComponentType(t.noteOnMessage.NoteStyle()),
		t.noteOnMessage.SkinParamBackcolored(t.skinParam), t.noteOnMessage.Display(), t.noteOnMessage.Colors())
}

func (t *CommunicationTileNoteLeft) notePosition(stringBounder font.StringBounder) Real {
	dim := t.component().PreferredDimension(stringBounder)
	return t.livingSpace.PosC(stringBounder).AddFixed(-dim.Width())
}

// DrawU draws the message, then the note on the same row.
func (t *CommunicationTileNoteLeft) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())
	t.tile.DrawU(ug)
	p := t.notePosition(stringBounder)
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	comp.DrawU(ug.Apply(klimt.UTranslateDx(p.CurrentValue())), area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *CommunicationTileNoteLeft) PreferredHeight() float64 {
	dim := t.component().PreferredDimension(t.StringBounder())
	return jmath.Max(t.tile.PreferredHeight(), dim.Height())
}

// AddConstraints implements Tile.
func (t *CommunicationTileNoteLeft) AddConstraints() { t.tile.AddConstraints() }

// MinX implements Tile.
func (t *CommunicationTileNoteLeft) MinX() Real { return t.notePosition(t.StringBounder()) }

// MaxX implements Tile.
func (t *CommunicationTileNoteLeft) MaxX() Real { return t.tile.MaxX() }

// CommunicationTileNoteRight is a message with a note on its right.
type CommunicationTileNoteRight struct {
	abstractTile
	tile          Tile
	message       AbstractMessage
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	livingSpace   *LivingSpace
	noteOnMessage *Note
	yGauge        *YGauge
}

// Event implements Tile.
func (t *CommunicationTileNoteRight) Event() Event { return t.message }

func (t *CommunicationTileNoteRight) isCreate() bool { return t.message.Base().IsCreate() }

// ContactPointRelative implements Tile.
func (t *CommunicationTileNoteRight) ContactPointRelative() float64 {
	return t.tile.ContactPointRelative()
}

func newCommunicationTileNoteRight(tile Tile, message AbstractMessage, skin *rose.Rose, skinParam klimt.ISkinParam,
	livingSpace *LivingSpace, noteOnMessage *Note, _ *YGauge) *CommunicationTileNoteRight {
	t := &CommunicationTileNoteRight{tile: tile, message: message, skin: skin, skinParam: skinParam,
		noteOnMessage: noteOnMessage, livingSpace: livingSpace}
	t.initTile(t, stringBounderOf(tile))
	t.yGauge = noteDecoratorGauge(tile, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *CommunicationTileNoteRight) YGauge() *YGauge { return t.yGauge }

// OnGaugeResolved implements Tile.
func (t *CommunicationTileNoteRight) OnGaugeResolved() { t.tile.OnGaugeResolved() }

func (t *CommunicationTileNoteRight) component() klimt.Component {
	return t.skin.CreateComponentNote(t.noteOnMessage.UsedStyles(), noteComponentType(t.noteOnMessage.NoteStyle()),
		t.noteOnMessage.SkinParamBackcolored(t.skinParam), t.noteOnMessage.Display(), t.noteOnMessage.Colors())
}

func (t *CommunicationTileNoteRight) notePosition(stringBounder font.StringBounder) Real {
	if t.isCreate() {
		return t.livingSpace.PosD(stringBounder)
	}
	level := t.livingSpace.LevelAt(t, IgnoreFutureDeactivate)
	return t.livingSpace.PosC(stringBounder).AddFixed(float64(level * LiveDeltaSize))
}

// DrawU draws the message, then the note on the same row.
func (t *CommunicationTileNoteRight) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())
	t.tile.DrawU(ug)
	p := t.notePosition(stringBounder)
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	comp.DrawU(ug.Apply(klimt.UTranslateDx(p.CurrentValue())), area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *CommunicationTileNoteRight) PreferredHeight() float64 {
	dim := t.component().PreferredDimension(t.StringBounder())
	return jmath.Max(t.tile.PreferredHeight(), dim.Height())
}

// AddConstraints implements Tile.
func (t *CommunicationTileNoteRight) AddConstraints() { t.tile.AddConstraints() }

// MinX implements Tile.
func (t *CommunicationTileNoteRight) MinX() Real { return t.tile.MinX() }

// MaxX implements Tile.
func (t *CommunicationTileNoteRight) MaxX() Real {
	dim := t.component().PreferredDimension(t.StringBounder())
	return t.notePosition(t.StringBounder()).AddFixed(dim.Width())
}

// communicationTileNoteBottomTop is upstream's
// CommunicationTileNoteBottomTopAbstract: a message with a note below or
// above it, linked by a dotted line.
type communicationTileNoteBottomTop struct {
	abstractTile
	tile          Tile
	message       AbstractMessage
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	noteOnMessage *Note
	yGauge        *YGauge
}

// noteBottomTopSpaceY is upstream's spacey.
const noteBottomTopSpaceY = 10

func (t *communicationTileNoteBottomTop) init(self Tile, tile Tile, message AbstractMessage, skin *rose.Rose,
	skinParam klimt.ISkinParam, noteOnMessage *Note) {
	t.tile = tile
	t.message = message
	t.skin = skin
	t.skinParam = skinParam
	t.noteOnMessage = noteOnMessage
	t.initTile(self, stringBounderOf(tile))
	t.yGauge = noteDecoratorGauge(tile, t.PreferredHeight())
}

// Event implements Tile.
func (t *communicationTileNoteBottomTop) Event() Event { return t.message }

// ContactPointRelative implements Tile.
func (t *communicationTileNoteBottomTop) ContactPointRelative() float64 {
	return t.tile.ContactPointRelative()
}

// YGauge implements Tile.
func (t *communicationTileNoteBottomTop) YGauge() *YGauge { return t.yGauge }

// OnGaugeResolved implements Tile.
func (t *communicationTileNoteBottomTop) OnGaugeResolved() { t.tile.OnGaugeResolved() }

func (t *communicationTileNoteBottomTop) component() klimt.Component {
	return t.skin.CreateComponentNote(t.noteOnMessage.UsedStyles(), klimt.ComponentTypeNote,
		t.noteOnMessage.SkinParamBackcolored(t.skinParam), t.noteOnMessage.Display(), t.noteOnMessage.Colors())
}

func (t *communicationTileNoteBottomTop) notePosition() Real { return t.tile.MinX() }

func (t *communicationTileNoteBottomTop) drawLine(ug klimt.UGraphic, x1, y1, x2, y2 float64) {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram).
		MergedStyle(t.skinParam.CurrentStyleBuilder())
	c := style.Value(klimt.PNameLineColor).AsColor(t.skinParam.IHtmlColorSet())
	dx := x2 - x1
	dy := y2 - y1
	ug.Apply(klimt.NewUTranslate(x1, y1)).Apply(c).Apply(klimt.NewUStroke(2, 2, 1)).Draw(klimt.NewULine(dx, dy))
}

// PreferredHeight implements Tile: the message and the note stacked.
func (t *communicationTileNoteBottomTop) PreferredHeight() float64 {
	dim := t.component().PreferredDimension(t.StringBounder())
	return t.tile.PreferredHeight() + dim.Height() + noteBottomTopSpaceY
}

// AddConstraints implements Tile.
func (t *communicationTileNoteBottomTop) AddConstraints() { t.tile.AddConstraints() }

// MinX implements Tile.
func (t *communicationTileNoteBottomTop) MinX() Real { return t.tile.MinX() }

// MaxX implements Tile.
func (t *communicationTileNoteBottomTop) MaxX() Real {
	maxMessage := t.tile.MaxX()
	dim := t.component().PreferredDimension(t.StringBounder())
	maxNote := t.tile.MinX().AddFixed(dim.Width())
	return RealUtilsMax(maxMessage, maxNote)
}

func (t *communicationTileNoteBottomTop) middleMsg() float64 {
	return (t.tile.MinX().CurrentValue() + t.tile.MaxX().CurrentValue()) / 2
}

// CommunicationTileNoteBottom is a message with a note below it.
type CommunicationTileNoteBottom struct {
	communicationTileNoteBottomTop
}

func newCommunicationTileNoteBottom(tile Tile, message AbstractMessage, skin *rose.Rose, skinParam klimt.ISkinParam,
	noteOnMessage *Note, _ *YGauge) *CommunicationTileNoteBottom {
	t := &CommunicationTileNoteBottom{}
	t.init(t, tile, message, skin, skinParam, noteOnMessage)
	return t
}

// DrawU draws the message, the note below and the line between them.
func (t *CommunicationTileNoteBottom) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())
	t.tile.DrawU(ug)

	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))

	middleMsg := t.middleMsg()
	xNote := t.notePosition().CurrentValue()
	yNote := t.tile.PreferredHeight()

	comp.DrawU(ug.Apply(klimt.NewUTranslate(xNote, yNote+noteBottomTopSpaceY)), area, context2D(ug))

	t.drawLine(ug, middleMsg, t.tile.ContactPointRelative(), xNote+dim.Width()/2,
		yNote+noteBottomTopSpaceY+rose.PaddingY)
}

// CommunicationTileNoteTop is a message with a note above it.
type CommunicationTileNoteTop struct {
	communicationTileNoteBottomTop
}

func newCommunicationTileNoteTop(tile Tile, message AbstractMessage, skin *rose.Rose, skinParam klimt.ISkinParam,
	noteOnMessage *Note, _ *YGauge) *CommunicationTileNoteTop {
	t := &CommunicationTileNoteTop{}
	t.init(t, tile, message, skin, skinParam, noteOnMessage)
	return t
}

// DrawU draws the note, the message below and the line between them.
func (t *CommunicationTileNoteTop) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())

	t.tile.DrawU(ug.Apply(klimt.UTranslateDy(dim.Height() + noteBottomTopSpaceY)))

	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))

	middleMsg := t.middleMsg()
	xNote := t.notePosition().CurrentValue()

	comp.DrawU(ug.Apply(klimt.UTranslateDx(xNote)), area, context2D(ug))

	t.drawLine(ug, middleMsg, t.tile.ContactPointRelative()+dim.Height()+noteBottomTopSpaceY, xNote+dim.Width()/2,
		dim.Height()-2*rose.PaddingY)
}

// CommunicationTileSelfNoteLeft is a self message with a note on its
// left.
type CommunicationTileSelfNoteLeft struct {
	abstractTile
	tile          *CommunicationTileSelf
	message       AbstractMessage
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	noteOnMessage *Note
	yGauge        *YGauge
}

// Event implements Tile.
func (t *CommunicationTileSelfNoteLeft) Event() Event { return t.message }

// ContactPointRelative implements Tile.
func (t *CommunicationTileSelfNoteLeft) ContactPointRelative() float64 {
	return t.tile.ContactPointRelative()
}

func newCommunicationTileSelfNoteLeft(tile *CommunicationTileSelf, message AbstractMessage, skin *rose.Rose,
	skinParam klimt.ISkinParam, noteOnMessage *Note, _ *YGauge) *CommunicationTileSelfNoteLeft {
	t := &CommunicationTileSelfNoteLeft{tile: tile, message: message, skin: skin, skinParam: skinParam,
		noteOnMessage: noteOnMessage}
	t.initTile(t, tile.StringBounder())
	t.yGauge = noteDecoratorGauge(tile, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *CommunicationTileSelfNoteLeft) YGauge() *YGauge { return t.yGauge }

// OnGaugeResolved implements Tile.
func (t *CommunicationTileSelfNoteLeft) OnGaugeResolved() { t.tile.OnGaugeResolved() }

func (t *CommunicationTileSelfNoteLeft) component() klimt.Component {
	return t.skin.CreateComponentNote(t.noteOnMessage.UsedStyles(), klimt.ComponentTypeNote,
		t.noteOnMessage.SkinParamBackcolored(t.skinParam), t.noteOnMessage.Display(), t.noteOnMessage.Colors())
}

func (t *CommunicationTileSelfNoteLeft) notePosition(stringBounder font.StringBounder) Real {
	dim := t.component().PreferredDimension(stringBounder)
	return t.tile.MinX().AddFixed(-dim.Width())
}

// DrawU draws the loop, then the note on the same row.
func (t *CommunicationTileSelfNoteLeft) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())
	t.tile.DrawU(ug)
	p := t.notePosition(stringBounder)
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	comp.DrawU(ug.Apply(klimt.UTranslateDx(p.CurrentValue())), area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *CommunicationTileSelfNoteLeft) PreferredHeight() float64 {
	dim := t.component().PreferredDimension(t.StringBounder())
	return jmath.Max(t.tile.PreferredHeight(), dim.Height())
}

// AddConstraints implements Tile.
func (t *CommunicationTileSelfNoteLeft) AddConstraints() {
	t.tile.AddConstraints()
	t.ensureOwnBoxClearsNote()
}

// ensureOwnBoxClearsNote widens the margin of the participant to cover the
// note, as CommunicationTileSelf.ensureOwnBoxClearsLoop does for the loop.
func (t *CommunicationTileSelfNoteLeft) ensureOwnBoxClearsNote() {
	stringBounder := t.StringBounder()
	livingSpace1 := t.tile.LivingSpace1()
	overflowLeft := livingSpace1.PosB(stringBounder).CurrentValue() - t.MinX().CurrentValue()
	if overflowLeft > 0 {
		livingSpace1.EnsureMarginBefore(overflowLeft)
	}
}

// MinX implements Tile.
func (t *CommunicationTileSelfNoteLeft) MinX() Real { return t.notePosition(t.StringBounder()) }

// MaxX implements Tile.
func (t *CommunicationTileSelfNoteLeft) MaxX() Real { return t.tile.MaxX() }

// DrawnMinX is anchored on the drawn extent of the loop.
func (t *CommunicationTileSelfNoteLeft) DrawnMinX() Real {
	width := t.component().PreferredDimension(t.StringBounder()).Width()
	return t.tile.DrawnMinX().AddFixed(-width)
}

// DrawnMaxX implements Tile.
func (t *CommunicationTileSelfNoteLeft) DrawnMaxX() Real { return t.tile.DrawnMaxX() }

// CommunicationTileSelfNoteRight is a self message with a note on its
// right.
type CommunicationTileSelfNoteRight struct {
	abstractTile
	tile          *CommunicationTileSelf
	message       *Message
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	noteOnMessage *Note
	yGauge        *YGauge
}

// Event implements Tile.
func (t *CommunicationTileSelfNoteRight) Event() Event { return t.message }

// ContactPointRelative implements Tile.
func (t *CommunicationTileSelfNoteRight) ContactPointRelative() float64 {
	return t.tile.ContactPointRelative()
}

func newCommunicationTileSelfNoteRight(tile *CommunicationTileSelf, message *Message, skin *rose.Rose,
	skinParam klimt.ISkinParam, noteOnMessage *Note, _ *YGauge) *CommunicationTileSelfNoteRight {
	t := &CommunicationTileSelfNoteRight{tile: tile, message: message, skin: skin, skinParam: skinParam,
		noteOnMessage: noteOnMessage}
	t.initTile(t, tile.StringBounder())
	t.yGauge = noteDecoratorGauge(tile, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *CommunicationTileSelfNoteRight) YGauge() *YGauge { return t.yGauge }

// OnGaugeResolved implements Tile.
func (t *CommunicationTileSelfNoteRight) OnGaugeResolved() { t.tile.OnGaugeResolved() }

func (t *CommunicationTileSelfNoteRight) component() klimt.Component {
	return t.skin.CreateComponentNote(t.noteOnMessage.UsedStyles(), klimt.ComponentTypeNote,
		t.noteOnMessage.SkinParamBackcolored(t.skinParam), t.noteOnMessage.Display(), t.noteOnMessage.Colors())
}

func (t *CommunicationTileSelfNoteRight) notePosition() Real { return t.tile.MaxX() }

// DrawU draws the loop, then the note on the same row.
func (t *CommunicationTileSelfNoteRight) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(dim.Width(), dim.Height())
	t.tile.DrawU(ug)
	p := t.notePosition()
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	comp.DrawU(ug.Apply(klimt.UTranslateDx(p.CurrentValue())), area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *CommunicationTileSelfNoteRight) PreferredHeight() float64 {
	dim := t.component().PreferredDimension(t.StringBounder())
	return jmath.Max(t.tile.PreferredHeight(), dim.Height())
}

// AddConstraints implements Tile.
func (t *CommunicationTileSelfNoteRight) AddConstraints() {
	t.tile.AddConstraints()
	t.ensureOwnBoxClearsNote()
}

func (t *CommunicationTileSelfNoteRight) ensureOwnBoxClearsNote() {
	stringBounder := t.StringBounder()
	livingSpace1 := t.tile.LivingSpace1()
	overflowRight := t.MaxX().CurrentValue() - livingSpace1.PosD(stringBounder).CurrentValue()
	if overflowRight > 0 {
		livingSpace1.EnsureMarginAfter(overflowRight)
	}
}

// MinX implements Tile.
func (t *CommunicationTileSelfNoteRight) MinX() Real { return t.tile.MinX() }

// MaxX implements Tile.
func (t *CommunicationTileSelfNoteRight) MaxX() Real {
	dim := t.component().PreferredDimension(t.StringBounder())
	return t.notePosition().AddFixed(dim.Width())
}

// DrawnMinX implements Tile.
func (t *CommunicationTileSelfNoteRight) DrawnMinX() Real { return t.tile.DrawnMinX() }

// DrawnMaxX is anchored on the drawn extent of the loop.
func (t *CommunicationTileSelfNoteRight) DrawnMaxX() Real {
	width := t.component().PreferredDimension(t.StringBounder()).Width()
	return t.tile.DrawnMaxX().AddFixed(width)
}
