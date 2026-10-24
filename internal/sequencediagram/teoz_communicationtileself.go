// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelf.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// CommunicationTileSelf is a message from a participant to itself.
type CommunicationTileSelf struct {
	abstractCommunicationTile
	livingSpace1    *LivingSpace
	message         *Message
	skin            *rose.Rose
	skinParam       klimt.ISkinParam
	livingSpaces    *LivingSpaces
	yGauge          *YGauge
	cachedComponent klimt.ArrowComponent
}

// Event implements Tile.
func (t *CommunicationTileSelf) Event() Event { return t.message }

// ContactPointRelative implements Tile.
func (t *CommunicationTileSelf) ContactPointRelative() float64 {
	return t.component(t.StringBounder()).YPoint(t.StringBounder())
}

func newCommunicationTileSelf(stringBounder font.StringBounder, livingSpace1 *LivingSpace, message *Message,
	skin *rose.Rose, skinParam klimt.ISkinParam, livingSpaces *LivingSpaces, currentY *YGauge) *CommunicationTileSelf {
	t := &CommunicationTileSelf{livingSpace1: livingSpace1, livingSpaces: livingSpaces, message: message,
		skin: skin, skinParam: skinParam}
	t.initTile(t, stringBounder)
	t.skinParamCommunication = skinParam
	if message.IsParallel() {
		t.yGauge = YGaugeCreateParallel(currentY, t.ContactPointRelative(), t.PreferredHeight())
	} else {
		t.yGauge = YGaugeCreateWithContact(currentY, t.ContactPointRelative(), t.PreferredHeight())
	}
	return t
}

// YGauge implements Tile.
func (t *CommunicationTileSelf) YGauge() *YGauge { return t.yGauge }

func (t *CommunicationTileSelf) component(font.StringBounder) klimt.ArrowComponent {
	if t.cachedComponent == nil {
		arrowConfiguration := t.message.ArrowConfiguration().Self()
		t.cachedComponent = t.skin.CreateComponentArrow(t.message.UsedStyles(), arrowConfiguration, t.skinParam,
			t.message.LabelNumbered())
	}
	return t.cachedComponent
}

// OnGaugeResolved records the livebox step of an activation, deactivation
// or destruction.
func (t *CommunicationTileSelf) OnGaugeResolved() {
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	p1 := comp.StartPoint(t.StringBounder(), dim)
	p2 := comp.EndPoint(t.StringBounder(), dim)
	y := t.YGauge().Min().CurrentValue()

	switch {
	case t.message.IsActivate():
		t.livingSpace1.AddStepForLivebox(t.Event(), y+p2.Y)
	case t.message.IsDeactivate():
		t.livingSpace1.AddStepForLivebox(t.Event(), y+p1.Y)
	case t.message.IsDestroy():
		t.livingSpace1.AddStepForLivebox(t.Event(), y+p2.Y)
	}
}

// DrawU draws the loop.
func (t *CommunicationTileSelf) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	comp := t.component(stringBounder)
	dim := comp.PreferredDimension(stringBounder)
	x1 := t.MinX().CurrentValue()
	levelIgnore := t.livingSpace1.LevelAt(t, IgnoreFutureActivate)
	levelConsidere := t.livingSpace1.LevelAt(t, ConsiderFutureDeactivate)
	if !t.isReverseDefine() {
		x1 += float64(LiveDeltaSize * levelIgnore)
		if levelIgnore < levelConsidere {
			x1 += float64(LiveDeltaSize * (levelConsidere - levelIgnore))
		}
	}

	area := klimt.AreaCreate(dim.Width(), dim.Height())
	area.SetDeltaX1(float64((levelIgnore - levelConsidere) * LiveDeltaSize))
	area.SetLevel(levelIgnore)
	area.SetLiveDeltaSize(LiveDeltaSize)
	ug = ug.Apply(klimt.UTranslateDx(x1))
	ug.StartGroup(t.message.groupTypeMessage())
	comp.DrawU(ug, area, context2D(ug))
	ug.CloseGroup()
}

// PreferredHeight implements Tile.
func (t *CommunicationTileSelf) PreferredHeight() float64 {
	comp := t.component(t.StringBounder())
	return comp.PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints keeps the neighbour participant clear of the loop, and
// widens the margin of the participant for its box.
func (t *CommunicationTileSelf) AddConstraints() {
	if t.sequenceMessageSpan() {
		return
	}
	if t.isReverseDefine() {
		previous := t.previous()
		if previous != nil {
			t.livingSpace1.PosC(t.StringBounder()).EnsureBiggerThan(
				previous.PosC2(t.StringBounder()).AddFixed(t.compWidth()))
		}
		t.ensureOwnBoxClearsLoop()
	} else {
		next := t.next()
		if next != nil {
			next.PosC(t.StringBounder()).EnsureBiggerThan(t.MaxX())
		}
		t.ensureOwnBoxClearsLoop()
	}
}

// ensureOwnBoxClearsLoop grows the margin of the participant to cover the
// loop past its head, for the box ("box" ... "end box") drawn from the
// margins (upstream's issue #2791).
func (t *CommunicationTileSelf) ensureOwnBoxClearsLoop() {
	stringBounder := t.StringBounder()
	if t.isReverseDefine() {
		overflowLeft := t.livingSpace1.PosB(stringBounder).CurrentValue() - t.MinX().CurrentValue()
		if overflowLeft > 0 {
			t.livingSpace1.EnsureMarginBefore(overflowLeft)
		}
	} else {
		overflowRight := t.MaxX().CurrentValue() - t.livingSpace1.PosD(stringBounder).CurrentValue()
		if overflowRight > 0 {
			t.livingSpace1.EnsureMarginAfter(overflowRight)
		}
	}
}

func (t *CommunicationTileSelf) isReverseDefine() bool {
	return t.message.ArrowConfiguration().IsReverseDefine()
}

// LivingSpace1 returns the column of the participant.
func (t *CommunicationTileSelf) LivingSpace1() *LivingSpace { return t.livingSpace1 }

func (t *CommunicationTileSelf) previous() *LivingSpace {
	var previous *LivingSpace
	for _, current := range t.livingSpaces.Values() {
		if current == t.livingSpace1 {
			return previous
		}
		previous = current
	}
	return nil
}

func (t *CommunicationTileSelf) next() *LivingSpace {
	values := t.livingSpaces.Values()
	for i, current := range values {
		if current == t.livingSpace1 && i+1 < len(values) {
			return values[i+1]
		}
	}
	return nil
}

func (t *CommunicationTileSelf) point1(stringBounder font.StringBounder) Real {
	return t.livingSpace1.PosC(stringBounder)
}

func (t *CommunicationTileSelf) liveDeltaWidthAdjustment() float64 {
	if t.livingSpace1.LevelAt(t, IgnoreFutureActivate) > 0 {
		return LiveDeltaSize
	}
	return 0.0
}

// MinX implements Tile.
func (t *CommunicationTileSelf) MinX() Real {
	if t.isReverseDefine() {
		adjustment := t.liveDeltaWidthAdjustment()
		return t.livingSpace1.PosC(t.StringBounder()).AddFixed(-t.compWidth() - adjustment)
	}
	return t.point1(t.StringBounder())
}

// MaxX implements Tile.
func (t *CommunicationTileSelf) MaxX() Real {
	if t.isReverseDefine() {
		return t.livingSpace1.PosC2(t.StringBounder())
	}
	return t.livingSpace1.PosC2(t.StringBounder()).AddFixed(t.compWidth())
}

func (t *CommunicationTileSelf) compWidth() float64 {
	comp := t.component(t.StringBounder())
	return comp.PreferredDimension(t.StringBounder()).Width()
}

// DrawnMinX is MinX with the drawn width of the loop instead of its layout
// width.
func (t *CommunicationTileSelf) DrawnMinX() Real {
	if t.isReverseDefine() {
		adjustment := t.liveDeltaWidthAdjustment()
		return t.livingSpace1.PosC(t.StringBounder()).AddFixed(-t.drawnWidth() - adjustment)
	}
	return t.MinX()
}

// DrawnMaxX is MaxX with the drawn width of the loop.
func (t *CommunicationTileSelf) DrawnMaxX() Real {
	if t.isReverseDefine() {
		return t.MaxX()
	}
	return t.livingSpace1.PosC2(t.StringBounder()).AddFixed(t.drawnWidth())
}

func (t *CommunicationTileSelf) drawnWidth() float64 {
	comp := t.component(t.StringBounder())
	return comp.(*rose.ComponentRoseSelfArrow).DrawnWidth(t.StringBounder())
}
