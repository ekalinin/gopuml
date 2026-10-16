// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationExoTile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
)

// CommunicationExoTile is a message from or to the border of the diagram.
type CommunicationExoTile struct {
	abstractTile
	livingSpace     *LivingSpace
	message         *MessageExo
	skin            *rose.Rose
	skinParam       klimt.ISkinParam
	tileArguments   *TileArguments
	yGauge          *YGauge
	cachedComponent klimt.ArrowComponent
}

// Event implements Tile.
func (t *CommunicationExoTile) Event() Event { return t.message }

func newCommunicationExoTile(livingSpace *LivingSpace, message *MessageExo, skin *rose.Rose,
	skinParam klimt.ISkinParam, tileArguments *TileArguments, currentY *YGauge) *CommunicationExoTile {
	t := &CommunicationExoTile{tileArguments: tileArguments, livingSpace: livingSpace, message: message,
		skin: skin, skinParam: skinParam}
	t.initTile(t, tileArguments.StringBounder())
	if message.IsParallel() {
		t.yGauge = YGaugeCreateParallel(currentY, t.ContactPointRelative(), t.PreferredHeight())
	} else {
		t.yGauge = YGaugeCreateWithContact(currentY, t.ContactPointRelative(), t.PreferredHeight())
	}
	return t
}

// YGauge implements Tile.
func (t *CommunicationExoTile) YGauge() *YGauge { return t.yGauge }

// ContactPointRelative implements Tile.
func (t *CommunicationExoTile) ContactPointRelative() float64 {
	return t.component(t.StringBounder()).YPoint(t.StringBounder())
}

func (t *CommunicationExoTile) component(font.StringBounder) klimt.ArrowComponent {
	if t.cachedComponent == nil {
		arrowConfiguration := t.message.ArrowConfiguration()
		if t.message.Type().Direction() == -1 {
			arrowConfiguration = arrowConfiguration.Reverse()
		}
		t.cachedComponent = t.skin.CreateComponentArrow(t.message.UsedStyles(), arrowConfiguration, t.skinParam,
			t.message.LabelNumbered())
	}
	return t.cachedComponent
}

// DrawU draws the arrow between the border and the lifeline.
func (t *CommunicationExoTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	comp := t.component(stringBounder)
	dim := comp.PreferredDimension(stringBounder)
	x1 := t.point1Value(stringBounder)
	x2 := t.point2Value(stringBounder)

	textDeltaX := 0.0
	if t.isFromLeftBorderMessage() {
		textDeltaX = t.point1(stringBounder).CurrentValue() - x1
	} else if t.isFromRightBorderMessage() {
		textDeltaX = t.point2(stringBounder).CurrentValue() - x2
	}

	level := t.livingSpace.LevelAt(t, IgnoreFutureDeactivate)
	if level > 0 {
		if t.message.Type().IsRightBorder() {
			x1 += float64(LiveDeltaSize * level)
		} else {
			x2 += float64(LiveDeltaSize * (level - 2))
		}
	}

	arrowConfiguration := t.message.ArrowConfiguration()
	typ := t.message.Type()
	circle := float64(rose.DiamCircle)/2 + 2
	if arrowConfiguration.Decoration1() == klimt.ArrowDecorationCircle && typ == ExoFromLeft {
		x1 += circle
	}
	if arrowConfiguration.Decoration2() == klimt.ArrowDecorationCircle && typ == ExoToLeft {
		x1 += circle
	}
	if arrowConfiguration.Decoration2() == klimt.ArrowDecorationCircle && typ == ExoToRight {
		x2 -= circle
	}
	if arrowConfiguration.Decoration1() == klimt.ArrowDecorationCircle && typ == ExoFromRight {
		x2 -= circle
	}

	area := klimt.AreaCreate(x2-x1, dim.Height())
	area.SetTextDeltaX(textDeltaX)
	ug = ug.Apply(klimt.UTranslateDx(x1))
	comp.DrawU(ug, area, context2D(ug))
}

func (t *CommunicationExoTile) isShortArrow() bool { return t.message.IsShortArrow() }

// PreferredHeight implements Tile.
func (t *CommunicationExoTile) PreferredHeight() float64 {
	comp := t.component(t.StringBounder())
	return comp.PreferredDimension(t.StringBounder()).Height()
}

func (t *CommunicationExoTile) preferredWidth(stringBounder font.StringBounder) float64 {
	comp := t.component(stringBounder)
	return comp.PreferredDimension(stringBounder).Width()
}

// AddConstraints keeps the lifeline far enough from the left border for
// the arrow.
func (t *CommunicationExoTile) AddConstraints() {
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	width := dim.Width()

	if !t.message.Type().IsRightBorder() {
		t.livingSpace.PosC(t.StringBounder()).EnsureBiggerThan(t.tileArguments.XOrigin().AddFixed(width))
	}
}

// OnGaugeResolved records the livebox step.
func (t *CommunicationExoTile) OnGaugeResolved() {
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	arrowY := comp.StartPoint(t.StringBounder(), dim).Y
	t.livingSpace.AddStepForLivebox(t.Event(), t.YGauge().Min().CurrentValue()+arrowY)
}

func (t *CommunicationExoTile) point1(stringBounder font.StringBounder) Real {
	if t.message.Type().IsRightBorder() {
		return t.livingSpace.PosC(stringBounder)
	}
	return t.livingSpace.PosC(stringBounder).AddFixed(-t.preferredWidth(stringBounder))
}

func (t *CommunicationExoTile) point2(stringBounder font.StringBounder) Real {
	if t.message.Type().IsLeftBorder() {
		return t.livingSpace.PosC(stringBounder)
	}
	return t.point1(stringBounder).AddFixed(t.preferredWidth(stringBounder))
}

func (t *CommunicationExoTile) point1Value(stringBounder font.StringBounder) float64 {
	if t.isFromLeftBorderMessage() {
		return t.tileArguments.Border1()
	}
	return t.point1(stringBounder).CurrentValue()
}

func (t *CommunicationExoTile) point2Value(stringBounder font.StringBounder) float64 {
	if t.isFromRightBorderMessage() {
		return t.tileArguments.Border2()
	}
	return t.point2(stringBounder).CurrentValue()
}

// MinX implements Tile.
func (t *CommunicationExoTile) MinX() Real { return t.point1(t.StringBounder()) }

// MaxX implements Tile.
func (t *CommunicationExoTile) MaxX() Real { return t.point2(t.StringBounder()) }

// MiddleX implements Tile: a message from the left border is centered on
// its arrow, not on the border.
func (t *CommunicationExoTile) MiddleX() float64 {
	if !t.isFromLeftBorderMessage() {
		return t.abstractTile.MiddleX()
	}
	minv := t.point1Value(t.StringBounder())
	maxv := minv + t.preferredWidth(t.StringBounder())
	return (minv + maxv) / 2
}

func (t *CommunicationExoTile) isFromRightBorderMessage() bool {
	return t.message.Type().IsRightBorder() && !t.isShortArrow()
}

func (t *CommunicationExoTile) isFromLeftBorderMessage() bool {
	return t.message.Type().IsLeftBorder() && !t.isShortArrow()
}
