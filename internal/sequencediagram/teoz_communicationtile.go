// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
)

// LiveDeltaSize is the horizontal shift of an activation level.
const LiveDeltaSize = 5

// CommunicationTile is a message between two participants.
type CommunicationTile struct {
	abstractCommunicationTile
	livingSpace1 *LivingSpace
	livingSpace2 *LivingSpace
	livingSpaces *LivingSpaces
	message      *Message
	skin         *rose.Rose
	skinParam    klimt.ISkinParam
	yGauge       *YGauge

	// cachedComponent memoizes one arrow component per orientation
	// (index 1 for reverse): IsReverse reads the solver and may change
	// between the layout phases.
	cachedComponent [2]klimt.ArrowComponent
}

// Event implements Tile.
func (t *CommunicationTile) Event() Event { return t.message }

func newCommunicationTile(stringBounder font.StringBounder, livingSpaces *LivingSpaces, message *Message,
	skin *rose.Rose, skinParam klimt.ISkinParam, currentY *YGauge) *CommunicationTile {
	t := &CommunicationTile{livingSpace1: livingSpaces.Get(message.Participant1()),
		livingSpace2: livingSpaces.Get(message.Participant2()), livingSpaces: livingSpaces, message: message,
		skin: skin, skinParam: skinParam}
	t.initTile(t, stringBounder)
	t.skinParamCommunication = skinParam

	if message.IsCreate() {
		t.livingSpace2.GoCreate()
	}

	comp := t.component(t.StringBounder())
	contactRelative := comp.YPoint(t.StringBounder())
	// PreferredHeight, not the arrow height: a create message is as tall
	// as the created participant's head
	height := t.PreferredHeight()

	// The contact line (arrow line) is published in the gauge; parallel (&)
	// messages share it
	if message.IsParallel() {
		t.yGauge = YGaugeCreateParallel(currentY, contactRelative, height)
	} else {
		t.yGauge = YGaugeCreateWithContact(currentY, contactRelative, height)
	}
	return t
}

// YGauge implements Tile.
func (t *CommunicationTile) YGauge() *YGauge { return t.yGauge }

// IsReverse reports whether the message goes from right to left.
func (t *CommunicationTile) IsReverse(stringBounder font.StringBounder) bool {
	point1 := t.livingSpace1.PosC(stringBounder)
	point2 := t.livingSpace2.PosC(stringBounder)
	return point1.CurrentValue() > point2.CurrentValue()
}

func (t *CommunicationTile) isCreate() bool { return t.message.IsCreate() }

func (t *CommunicationTile) arrowThickness() float64 {
	result := t.skinParam.Thickness(klimt.LineParamSequenceArrow, nil)
	if result == nil {
		return 1
	}
	return result.Thickness()
}

func (t *CommunicationTile) component(stringBounder font.StringBounder) klimt.ArrowComponent {
	reverse := t.IsReverse(stringBounder)
	idx := 0
	if reverse {
		idx = 1
	}
	if t.cachedComponent[idx] == nil {
		arrowConfiguration := t.message.ArrowConfiguration()
		if reverse {
			arrowConfiguration = arrowConfiguration.Reverse()
		}
		arrowConfiguration = arrowConfiguration.WithThickness(t.arrowThickness())
		t.cachedComponent[idx] = t.skin.CreateComponentArrow(t.message.UsedStyles(), arrowConfiguration,
			t.skinParam, t.message.LabelNumbered())
	}
	return t.cachedComponent[idx]
}

func (t *CommunicationTile) componentMulticast(reverse bool) klimt.ArrowComponent {
	arrowConfiguration := t.message.ArrowConfiguration()
	if reverse {
		arrowConfiguration = arrowConfiguration.Reverse()
	}
	arrowConfiguration = arrowConfiguration.WithThickness(t.arrowThickness())
	return t.skin.CreateComponentArrow(t.message.UsedStyles(), arrowConfiguration, t.skinParam, klimt.DisplayNULL())
}

// ContactPointRelative implements Tile: the y of the arrow line.
func (t *CommunicationTile) ContactPointRelative() float64 {
	return t.component(t.StringBounder()).YPoint(t.StringBounder())
}

// OnGaugeResolved records the creation and the livebox steps.
func (t *CommunicationTile) OnGaugeResolved() {
	y := t.YGauge().Min().CurrentValue()
	if t.message.IsCreate() {
		t.livingSpace2.GoCreateAt(y)
	}
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	arrowY := comp.StartPoint(t.StringBounder(), dim).Y

	t.livingSpace1.AddStepForLivebox(t.Event(), y+arrowY)
	t.livingSpace2.AddStepForLivebox(t.Event(), y+arrowY)
}

// DrawU draws the arrow, and the head of a created participant.
func (t *CommunicationTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: every tile draws itself at its own absolute
	// gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	_, has1 := t.message.Part1Anchor()
	_, has2 := t.message.Part2Anchor()
	if has1 || has2 {
		return
	}

	stringBounder := ug.StringBounder()
	comp := t.component(stringBounder)
	dim := comp.PreferredDimension(stringBounder)
	x1 := t.point1(stringBounder).CurrentValue()
	x2 := t.point2(stringBounder).CurrentValue()
	t.drawMulticast(ug.Apply(klimt.UTranslateDy(comp.PosArrow(stringBounder))))

	var area *klimt.Area
	if t.IsReverse(stringBounder) {
		level1 := t.livingSpace1.LevelAt(t, IgnoreFutureDeactivate)
		level2 := t.livingSpace2.LevelAt(t, IgnoreFutureDeactivate)
		if level1 == 1 {
			x1 -= LiveDeltaSize
		} else if level1 > 2 {
			x1 += float64(LiveDeltaSize * (level1 - 2))
		}
		x2 += float64(LiveDeltaSize * level2)
		area = klimt.AreaCreate(x1-x2, dim.Height())
		ug = ug.Apply(klimt.UTranslateDx(x2))
		if t.isCreate() {
			t.livingSpace2.DrawHead(ug, context2D(ug), geom.VerticalAlignmentTop, geom.HorizontalAlignmentRight)
		}
	} else {
		level1 := t.livingSpace1.LevelAt(t, IgnoreFutureDeactivate)
		level2 := t.livingSpace2.LevelAt(t, IgnoreFutureDeactivate)
		if level2 > 0 {
			level2 = level2 - 2
		}
		x1 += float64(LiveDeltaSize * level1)
		x2 += float64(LiveDeltaSize * level2)
		area = klimt.AreaCreate(x2-x1, dim.Height())
		ug = ug.Apply(klimt.UTranslateDx(x1))
		if t.isCreate() {
			t.livingSpace2.DrawHead(ug.Apply(klimt.UTranslateDx(area.DimensionToUse().Width())), context2D(ug),
				geom.VerticalAlignmentTop, geom.HorizontalAlignmentLeft)
		}
	}
	comp.DrawU(ug, area, context2D(ug))
}

func (t *CommunicationTile) drawMulticast(ug klimt.UGraphic) {
	if len(t.message.Multicast()) == 0 {
		return
	}
	stringBounder := ug.StringBounder()
	x1 := t.point1(stringBounder).CurrentValue()
	dy := 2.0
	for _, participant := range t.message.Multicast() {
		x2 := t.livingSpaces.Get(participant).PosC(stringBounder).CurrentValue()
		reverse := x2 < x1
		comp := t.componentMulticast(reverse)
		dim := comp.PreferredDimension(stringBounder)
		area := klimt.AreaCreate(math.Abs(x2-x1), dim.Height())
		ug2 := ug.Apply(klimt.UTranslateDx(jmath.Min(x1, x2))).Apply(klimt.UTranslateDy(dy))
		dy += 2
		comp.DrawU(ug2, area, context2D(ug2))
	}
}

// PreferredHeight implements Tile.
func (t *CommunicationTile) PreferredHeight() float64 {
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	height := dim.Height()
	if t.isCreate() {
		height = jmath.Max(height, t.livingSpace2.HeadPreferredDimension(t.StringBounder()).Height())
	}
	return height
}

// AddConstraints keeps the participants far enough apart for the arrow.
func (t *CommunicationTile) AddConstraints() {
	if t.sequenceMessageSpan() {
		return
	}
	comp := t.component(t.StringBounder())
	dim := comp.PreferredDimension(t.StringBounder())
	width := dim.Width()

	point1 := t.point1(t.StringBounder())
	point2 := t.point2(t.StringBounder())
	if t.IsReverse(t.StringBounder()) {
		level1 := t.livingSpace1.LevelAt(t, IgnoreFutureDeactivate)
		level2 := t.livingSpace2.LevelAt(t, IgnoreFutureDeactivate)
		if level1 > 0 {
			point1 = point1.AddFixed(-LiveDeltaSize)
		}
		point2 = point2.AddFixed(float64(level2 * LiveDeltaSize))
		point1.EnsureBiggerThan(point2.AddFixed(width))
	} else {
		level2 := t.livingSpace2.LevelAt(t, IgnoreFutureDeactivate)
		if level2 > 0 {
			point2 = point2.AddFixed(-LiveDeltaSize)
		}
		point2.EnsureBiggerThan(point1.AddFixed(width))
	}
}

func (t *CommunicationTile) point1(stringBounder font.StringBounder) Real {
	return t.livingSpace1.PosC(stringBounder)
}

func (t *CommunicationTile) point2(stringBounder font.StringBounder) Real {
	if t.message.IsCreate() {
		if t.IsReverse(stringBounder) {
			return t.livingSpace2.PosD(stringBounder)
		}
		return t.livingSpace2.PosB(stringBounder)
	}
	return t.livingSpace2.PosC(stringBounder)
}

// MinX implements Tile.
func (t *CommunicationTile) MinX() Real {
	stringBounder := t.StringBounder()
	if t.IsReverse(stringBounder) {
		return t.point2(stringBounder)
	}
	return t.point1(stringBounder)
}

// MaxX implements Tile.
func (t *CommunicationTile) MaxX() Real {
	stringBounder := t.StringBounder()
	width := t.component(stringBounder).PreferredDimension(stringBounder).Width()
	var leftPoint, rightPoint Real
	if t.IsReverse(stringBounder) {
		leftPoint = t.point2(stringBounder)
	} else {
		leftPoint = t.point1(stringBounder)
	}
	if t.IsReverse(stringBounder) {
		rightPoint = t.point1(stringBounder)
	} else {
		rightPoint = t.point2(stringBounder)
	}
	return RealUtilsMax(rightPoint, leftPoint.AddFixed(width))
}
