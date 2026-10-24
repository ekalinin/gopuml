// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/DividerTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/DelayTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/HSpaceTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LifeEventTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/NewpageTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/ReferenceTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/ElseTile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// DividerTile is a "== text ==" divider across the diagram.
type DividerTile struct {
	abstractTile
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	divider       *Divider
	xorigin       Real
	tileArguments *TileArguments
	yGauge        *YGauge
}

// Event implements Tile.
func (t *DividerTile) Event() Event { return t.divider }

func newDividerTile(divider *Divider, tileArguments *TileArguments, currentY *YGauge) *DividerTile {
	t := &DividerTile{tileArguments: tileArguments, divider: divider, skin: tileArguments.Skin(),
		skinParam: tileArguments.SkinParam(), xorigin: tileArguments.XOrigin()}
	t.initTile(t, tileArguments.StringBounder())
	t.yGauge = YGaugeCreate(currentY.Max(), t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *DividerTile) YGauge() *YGauge { return t.yGauge }

func (t *DividerTile) component() klimt.Component {
	return t.skin.CreateComponent(t.divider.UsedStyles(), klimt.ComponentTypeDivider, nil, t.skinParam,
		t.divider.Text())
}

// DrawU draws the divider across the diagram.
func (t *DividerTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(t.tileArguments.Border2()-t.tileArguments.Border1()-t.xorigin.CurrentValue(),
		dim.Height())
	ug = ug.Apply(klimt.UTranslateDx(t.tileArguments.Border1()))
	comp.DrawU(ug, area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *DividerTile) PreferredHeight() float64 {
	return t.component().PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints implements Tile.
func (t *DividerTile) AddConstraints() {}

// MinX implements Tile.
func (t *DividerTile) MinX() Real { return t.xorigin }

// MaxX implements Tile.
func (t *DividerTile) MaxX() Real {
	return t.xorigin.AddFixed(t.component().PreferredDimension(t.StringBounder()).Width())
}

// DelayTile is a "..." delay, centered between the first and last
// participants.
type DelayTile struct {
	abstractTile
	delay         *Delay
	tileArguments *TileArguments
	middle        Real
	yGauge        *YGauge
}

// Event implements Tile.
func (t *DelayTile) Event() Event { return t.delay }

func newDelayTile(delay *Delay, tileArguments *TileArguments, currentY *YGauge) *DelayTile {
	t := &DelayTile{delay: delay, tileArguments: tileArguments}
	t.initTile(t, tileArguments.StringBounder())
	t.yGauge = YGaugeCreate(currentY.Max(), t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *DelayTile) YGauge() *YGauge { return t.yGauge }

func (t *DelayTile) init(stringBounder font.StringBounder) {
	if t.middle != nil {
		return
	}
	first := t.tileArguments.FirstLivingSpace().PosC(stringBounder)
	last := t.tileArguments.LastLivingSpace().PosC(stringBounder)
	t.middle = RealUtilsMiddle(first, last)
}

func (t *DelayTile) component() klimt.Component {
	return t.tileArguments.Skin().CreateComponent(t.delay.UsedStyles(), klimt.ComponentTypeDelayText, nil,
		t.tileArguments.SkinParam(), t.delay.Text())
}

func (t *DelayTile) preferredWidth(stringBounder font.StringBounder) float64 {
	return t.component().PreferredDimension(stringBounder).Width()
}

// DrawU draws the text of the delay and records the delay on the
// lifelines.
func (t *DelayTile) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	t.init(stringBounder)
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(t.preferredWidth(stringBounder), dim.Height())
	// Self-translate prologue: absolute gauge position
	ypos := t.YGauge().Min().CurrentValue()
	ug = ug.Apply(klimt.UTranslateDy(ypos))
	t.tileArguments.LivingSpaces().DelayOn(ypos, dim.Height())

	ug = ug.Apply(klimt.UTranslateDx(t.MinX().CurrentValue()))
	comp.DrawU(ug, area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *DelayTile) PreferredHeight() float64 {
	return t.component().PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints implements Tile.
func (t *DelayTile) AddConstraints() {}

// MinX implements Tile.
func (t *DelayTile) MinX() Real {
	t.init(t.StringBounder())
	return t.middle.AddFixed(-t.preferredWidth(t.StringBounder()) / 2)
}

// MaxX implements Tile.
func (t *DelayTile) MaxX() Real {
	t.init(t.StringBounder())
	return t.middle.AddFixed(t.preferredWidth(t.StringBounder()) / 2)
}

// HSpaceTile is a "|||" vertical space.
type HSpaceTile struct {
	abstractTile
	hspace  *HSpace
	xorigin Real
	yGauge  *YGauge
}

// Event implements Tile.
func (t *HSpaceTile) Event() Event { return t.hspace }

func newHSpaceTile(hspace *HSpace, tileArguments *TileArguments, currentY *YGauge) *HSpaceTile {
	t := &HSpaceTile{hspace: hspace, xorigin: tileArguments.XOrigin()}
	t.initTile(t, tileArguments.StringBounder())
	// A pure spacer: no contact line of its own, but it propagates the one
	// of the previous tile, between parallel messages
	t.yGauge = YGaugeCreatePropagating(currentY, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *HSpaceTile) YGauge() *YGauge { return t.yGauge }

// DrawU draws nothing.
func (t *HSpaceTile) DrawU(klimt.UGraphic) {}

// PreferredHeight implements Tile.
func (t *HSpaceTile) PreferredHeight() float64 { return float64(t.hspace.Pixel()) }

// AddConstraints implements Tile.
func (t *HSpaceTile) AddConstraints() {}

// MinX implements Tile.
func (t *HSpaceTile) MinX() Real { return t.xorigin }

// MaxX implements Tile.
func (t *HSpaceTile) MaxX() Real { return t.xorigin.AddFixed(10) }

// LifeEventTile is an activation, deactivation, creation or destruction.
type LifeEventTile struct {
	abstractTile
	lifeEvent     *LifeEvent
	tileArguments *TileArguments
	livingSpace   *LivingSpace
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	yGauge        *YGauge
}

// OnGaugeResolved records the livebox step and the destruction.
func (t *LifeEventTile) OnGaugeResolved() {
	y := t.YGauge().Min().CurrentValue()
	t.livingSpace.AddStepForLivebox(t.Event(), y)
	if t.lifeEvent.Type() == LifeEventDestroy {
		t.livingSpace.GoDestroy(y)
	}
}

// Event implements Tile.
func (t *LifeEventTile) Event() Event { return t.lifeEvent }

// ContactPointRelative implements Tile.
func (t *LifeEventTile) ContactPointRelative() float64 { return 0 }

func newLifeEventTile(lifeEvent *LifeEvent, tileArguments *TileArguments, livingSpace *LivingSpace,
	skin *rose.Rose, skinParam klimt.ISkinParam, currentY *YGauge) *LifeEventTile {
	t := &LifeEventTile{lifeEvent: lifeEvent, tileArguments: tileArguments, livingSpace: livingSpace, skin: skin,
		skinParam: skinParam}
	t.initTile(t, tileArguments.StringBounder())
	// The contact line is propagated through life events
	t.yGauge = YGaugeCreatePropagating(currentY, t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *LifeEventTile) YGauge() *YGauge { return t.yGauge }

func (t *LifeEventTile) usedStyle() []*klimt.Style {
	return []*klimt.Style{klimt.ComponentTypeDestroy.StyleSignature().MergedStyle(t.skinParam.CurrentStyleBuilder())}
}

// DrawU draws the cross of a destruction without message.
func (t *LifeEventTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	if t.isDestroyWithoutMessage() {
		cross := t.skin.CreateComponent(t.usedStyle(), klimt.ComponentTypeDestroy, nil, t.skinParam, nil)
		dimCross := cross.PreferredDimension(ug.StringBounder())
		x := t.livingSpace.PosC(ug.StringBounder()).CurrentValue()
		cross.DrawU(ug.Apply(klimt.UTranslateDx(x-dimCross.Width()/2)), nil, context2D(ug))
	}
}

func (t *LifeEventTile) isDestroyWithoutMessage() bool {
	return t.lifeEvent.Message() == nil && t.lifeEvent.Type() == LifeEventDestroy
}

// PreferredHeight implements Tile: the height of the cross of a
// destruction without message, else 0.
func (t *LifeEventTile) PreferredHeight() float64 {
	if t.isDestroyWithoutMessage() {
		cross := t.skin.CreateComponent(t.usedStyle(), klimt.ComponentTypeDestroy, nil, t.skinParam, nil)
		return cross.PreferredDimension(t.StringBounder()).Height()
	}
	return 0
}

// AddConstraints implements Tile.
func (t *LifeEventTile) AddConstraints() {}

// MinX implements Tile.
func (t *LifeEventTile) MinX() Real {
	levelAt := t.livingSpace.LevelAt(t, IgnoreFutureActivate)
	adjustment := 0.0
	if levelAt > 0 {
		adjustment = LiveDeltaSize
	}
	return t.livingSpace.PosC(t.StringBounder()).AddFixed(-adjustment)
}

// MaxX implements Tile.
func (t *LifeEventTile) MaxX() Real {
	levelAt := t.livingSpace.LevelAt(t, IgnoreFutureActivate)
	adjustment := 0.0
	if levelAt > 0 {
		adjustment = float64(levelAt * LiveDeltaSize)
	}
	return t.livingSpace.PosC(t.StringBounder()).AddFixed(adjustment)
}

// newpageMarginY is the space before and after the dashed separator of a
// page break.
const newpageMarginY = 10

// NewpageTile is a page break.
type NewpageTile struct {
	abstractTile
	newpage       *Newpage
	tileArguments *TileArguments
	yGauge        *YGauge
}

// ContactPointRelative implements Tile.
func (t *NewpageTile) ContactPointRelative() float64 { return 0 }

func newNewpageTile(newpage *Newpage, tileArguments *TileArguments, currentY *YGauge) *NewpageTile {
	t := &NewpageTile{newpage: newpage, tileArguments: tileArguments}
	t.initTile(t, tileArguments.StringBounder())
	t.yGauge = YGaugeCreate(currentY.Max(), t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *NewpageTile) YGauge() *YGauge { return t.yGauge }

func (t *NewpageTile) component() klimt.Component {
	return t.tileArguments.Skin().CreateComponentNewPage(t.newpage.UsedStyles(), t.tileArguments.SkinParam())
}

// DrawU draws the dashed separator, in the foreground only.
func (t *NewpageTile) DrawU(ug klimt.UGraphic) {
	// The dashed separator is displayed at the bottom of the page ending
	// here and at the top of the page starting here
	if context2D(ug).IsBackground() {
		return
	}
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))

	comp := t.component()
	area := klimt.AreaCreate(t.tileArguments.Border2()-t.tileArguments.Border1()-
		t.tileArguments.XOrigin().CurrentValue(), comp.PreferredHeight(t.StringBounder()))
	ug = ug.Apply(klimt.NewUTranslate(t.tileArguments.Border1(), newpageMarginY))
	comp.DrawU(ug, area, context2D(ug))
}

// PreferredHeight implements Tile.
func (t *NewpageTile) PreferredHeight() float64 {
	return t.component().PreferredHeight(t.StringBounder()) + 2*newpageMarginY
}

// AddConstraints implements Tile.
func (t *NewpageTile) AddConstraints() {}

// MinX implements Tile.
func (t *NewpageTile) MinX() Real { return t.tileArguments.XOrigin() }

// MaxX implements Tile.
func (t *NewpageTile) MaxX() Real { return t.tileArguments.XOrigin() }

// Event implements Tile.
func (t *NewpageTile) Event() Event { return t.newpage }

// ReferenceTile is a "ref over" frame.
type ReferenceTile struct {
	abstractTile
	reference     *Reference
	tileArguments *TileArguments
	first         Real
	last          Real
	yGauge        *YGauge

	noteLeft  klimt.Component
	noteRight klimt.Component
}

// Event implements Tile.
func (t *ReferenceTile) Event() Event { return t.reference }

func newReferenceTile(reference *Reference, tileArguments *TileArguments, currentY *YGauge) *ReferenceTile {
	t := &ReferenceTile{reference: reference, tileArguments: tileArguments}
	t.initTile(t, tileArguments.StringBounder())
	t.yGauge = YGaugeCreate(currentY.Max(), t.PreferredHeight())

	for _, noteOnMessage := range reference.NoteOnMessages() {
		skinParam2 := noteOnMessage.SkinParamBackcolored(tileArguments.SkinParam())
		note := tileArguments.Skin().CreateComponentNote(noteOnMessage.UsedStyles(),
			noteOnMessage.NoteStyle().ComponentType(), skinParam2, noteOnMessage.Display(), noteOnMessage.Colors())
		if noteOnMessage.Position() == NoteRight {
			t.noteRight = note
		} else {
			t.noteLeft = note
		}
	}
	return t
}

// YGauge implements Tile.
func (t *ReferenceTile) YGauge() *YGauge { return t.yGauge }

func (t *ReferenceTile) init(stringBounder font.StringBounder) {
	if t.first != nil {
		return
	}
	for _, p := range t.reference.Participants() {
		livingSpace := t.tileArguments.LivingSpace(p)
		pos := livingSpace.PosC(stringBounder)
		if t.first == nil || pos.CurrentValue() < t.first.CurrentValue() {
			t.first = livingSpace.PosB(stringBounder)
		}
		if t.last == nil || pos.CurrentValue() > t.last.CurrentValue() {
			t.last = livingSpace.PosD(stringBounder)
		}
	}
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	if len(t.reference.Participants()) == 1 {
		t.last = t.last.AddAtLeast(0)
	}
	t.last.EnsureBiggerThan(t.first.AddFixed(dim.Width()))
}

func (t *ReferenceTile) component() klimt.Component {
	strings := klimt.DisplayEmpty()
	strings = strings.Add(klimt.Text("ref"))
	strings = strings.AddAll(t.reference.Strings())
	return t.tileArguments.Skin().CreateComponent(t.reference.UsedStyles(), klimt.ComponentTypeReference, nil,
		t.tileArguments.SkinParam(), strings)
}

// DrawU draws the frame and its notes.
func (t *ReferenceTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	t.init(stringBounder)
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	area := klimt.AreaCreate(t.last.CurrentValue()-t.first.CurrentValue(), dim.Height())

	comp.DrawU(ug.Apply(klimt.UTranslateDx(t.first.CurrentValue())), area, context2D(ug))

	if t.noteLeft != nil {
		wn := t.noteLeft.PreferredWidth(stringBounder)
		hn := t.noteLeft.PreferredHeight(stringBounder)
		t.noteLeft.DrawU(ug.Apply(klimt.UTranslateDx(t.first.CurrentValue()-wn)),
			klimt.NewArea(geom.NewXDimension2D(wn, hn)), context2D(ug))
	}

	if t.noteRight != nil {
		wn := t.noteRight.PreferredWidth(stringBounder)
		hn := t.noteRight.PreferredHeight(stringBounder)
		t.noteRight.DrawU(ug.Apply(klimt.UTranslateDx(t.last.CurrentValue())),
			klimt.NewArea(geom.NewXDimension2D(wn, hn)), context2D(ug))
	}
}

// PreferredHeight implements Tile.
func (t *ReferenceTile) PreferredHeight() float64 {
	return t.component().PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints implements Tile.
func (t *ReferenceTile) AddConstraints() {}

// MinX implements Tile.
func (t *ReferenceTile) MinX() Real {
	t.init(t.StringBounder())
	if t.noteLeft != nil {
		wn := t.noteLeft.PreferredWidth(t.StringBounder())
		return t.first.AddFixed(-wn)
	}
	return t.first
}

// MaxX implements Tile.
func (t *ReferenceTile) MaxX() Real {
	t.init(t.StringBounder())
	if t.noteRight != nil {
		wn := t.noteRight.PreferredWidth(t.StringBounder())
		return t.last.AddFixed(wn)
	}
	return t.last
}

// ElseTile is an "else" of a fragment: it draws its own divider.
type ElseTile struct {
	abstractTile
	skin      *rose.Rose
	skinParam klimt.ISkinParam
	anElse    *GroupingLeaf
	parent    Tile
	yGauge    *YGauge
}

// Event implements Tile.
func (t *ElseTile) Event() Event { return t.anElse }

// ContactPointRelative implements Tile.
func (t *ElseTile) ContactPointRelative() float64 { return 0 }

func newElseTile(anElse *GroupingLeaf, skin *rose.Rose, skinParam klimt.ISkinParam, parent Tile,
	currentY *YGauge) *ElseTile {
	t := &ElseTile{anElse: anElse, skin: skin, skinParam: skinParam, parent: parent}
	t.initTile(t, stringBounderOf(parent))
	t.yGauge = YGaugeCreate(currentY.Max(), t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *ElseTile) YGauge() *YGauge { return t.yGauge }

// BackColorGeneral returns the background of the section, nil for none.
func (t *ElseTile) BackColorGeneral() color.HColor { return t.anElse.BackColorGeneral() }

func (t *ElseTile) component() klimt.Component {
	tmp := klimt.NewSkinParamBackcolored(t.skinParam, t.anElse.BackColorElement(), t.anElse.BackColorGeneral(),
		false)
	display := displayCreateOpt(optString(t.anElse.Comment()))
	return t.skin.CreateComponent(t.anElse.UsedStyles(), klimt.ComponentTypeGroupingElseTeoz, nil, tmp, display)
}

// DrawU draws the dashed divider and the label of the section across the
// frame of the fragment.
func (t *ElseTile) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))

	stringBounder := ug.StringBounder()
	comp := t.component()
	dim := comp.PreferredDimension(stringBounder)
	// The frame extents of the group, not its bounds: those also reserve
	// the notes of its end
	parent := t.parent.(groupingTileLike).grouping()
	minv := parent.frameMinX()
	maxv := parent.frameMaxX()
	context := context2D(ug)
	height := dim.Height()
	area := klimt.AreaCreate(maxv.CurrentValue()-minv.CurrentValue(), height)
	ug = ug.Apply(klimt.NewUTranslate(minv.CurrentValue(), 0))
	comp.DrawU(ug, area, context)
}

// PreferredHeight implements Tile.
func (t *ElseTile) PreferredHeight() float64 {
	return t.component().PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints implements Tile.
func (t *ElseTile) AddConstraints() {}

// MinX implements Tile.
func (t *ElseTile) MinX() Real { return t.parent.MinX() }

// MaxX implements Tile.
func (t *ElseTile) MaxX() Real {
	return t.MinX().AddFixed(t.component().PreferredDimension(t.StringBounder()).Width())
}
