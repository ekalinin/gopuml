// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LiveBoxes.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LiveBoxesDrawer.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/Stairs.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/Step.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/MutingLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/uml"
)

// LiveBoxes computes and draws the activation boxes of a participant.
type LiveBoxes struct {
	skin      *rose.Rose
	skinParam klimt.ISkinParam
	delays    sortedFloatMap[float64]
	p         *Participant
	events    []Event
	// eventsStep maps an event to its y; upstream's HashMap is only read
	// by key and by containsValue.
	eventsStep map[Event]float64

	eventInfo      map[Event][2]int
	nextLife       []int
	levelCacheSize int
}

// NewLiveBoxes returns the boxes of p.
func NewLiveBoxes(p *Participant, events []Event, skin *rose.Rose, skinParam klimt.ISkinParam) *LiveBoxes {
	return &LiveBoxes{p: p, events: events, skin: skin, skinParam: skinParam, eventsStep: map[Event]float64{},
		levelCacheSize: -1}
}

// AddStep records the y of an event involving the participant. A
// deactivation at the y of another event is moved 5 pixels down.
func (l *LiveBoxes) AddStep(event Event, y float64) {
	if event.DealWith(l.p) {
		if le, ok := event.(*LifeEvent); ok && le.IsDeactivate() && l.eventsStepContainsValue(y) {
			y += 5.0
		}
		l.eventsStep[event] = y
	}
}

func (l *LiveBoxes) eventsStepContainsValue(y float64) bool {
	for _, v := range l.eventsStep {
		if v == y {
			return true
		}
	}
	return false
}

// Participant returns the participant.
func (l *LiveBoxes) Participant() *Participant { return l.p }

// LevelAt returns the activation level at an event.
func (l *LiveBoxes) LevelAt(event Event, mode EventsHistoryMode) int {
	return l.levelAtInternal(event, mode)
}

// ensureLevelCache fills, in one pass, the index of each event and the
// level right after it, and for each position the next life event the
// message lookahead can reach (-1 when another kind of event stops it).
func (l *LiveBoxes) ensureLevelCache() {
	if l.eventInfo != nil && l.levelCacheSize == len(l.events) {
		return
	}
	l.eventInfo = map[Event][2]int{}
	l.levelCacheSize = len(l.events)
	level := 0
	for i, current := range l.events {
		if le, ok := current.(*LifeEvent); ok {
			if le.Participant() == l.p && le.IsActivate() {
				level++
			}
			if le.Participant() == l.p && le.IsDeactivateOrDestroy() {
				level = max(0, level-1)
			}
		}
		if _, ok := l.eventInfo[current]; !ok {
			l.eventInfo[current] = [2]int{i, level}
		}
	}
	l.nextLife = make([]int, len(l.events)+1)
	l.nextLife[len(l.events)] = -1
	for i := len(l.events) - 1; i >= 0; i-- {
		switch l.events[i].(type) {
		case *LifeEvent:
			l.nextLife[i] = i
		case *Note, AbstractMessage:
			l.nextLife[i] = l.nextLife[i+1]
		default:
			l.nextLife[i] = -1
		}
	}
}

func (l *LiveBoxes) levelAtInternal(event Event, mode EventsHistoryMode) int {
	l.ensureLevelCache()
	info, ok := l.eventInfo[event]
	if !ok {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	index := info[0]
	level := info[1]
	if msg, ok := event.(AbstractMessage); ok {
		seenActivate := false
		seenDeactivate := false
		for k := l.nextLife[index+1]; k != -1; k = l.nextLife[k+1] {
			le := l.events[k].(*LifeEvent)
			sameMessage := le.Message() == msg || (le.Message() != nil && le.Message().Base().IsParallelWith(msg))
			if !sameMessage {
				continue
			}
			if mode != IgnoreFutureActivate && le.IsActivate() && msg.DealWith(l.p) && le.Participant() == l.p {
				seenActivate = true
				if seenDeactivate {
					break
				}
				level++
			}
			if mode == ConsiderFutureDeactivate && le.IsDeactivateOrDestroy() && msg.DealWith(l.p) &&
				le.Participant() == l.p {
				seenDeactivate = true
				if seenActivate {
					break
				}
				level = max(0, level-1)
			}
		}
	}
	if _, ok := event.(*Note); ok {
		// An activate or deactivate attached to a previous message applies
		// retroactively at the message position, that is above this note:
		// the note placement must take it into account
		for k := index + 1; k < len(l.events); k++ {
			next := l.events[k]
			if _, isNote := next.(*Note); isNote {
				continue
			}
			le, isLife := next.(*LifeEvent)
			if !isLife {
				break
			}
			if le.Participant() != l.p || le.Message() == nil {
				continue
			}
			if mode != IgnoreFutureActivate && le.IsActivate() {
				level++
			}
			if mode == ConsiderFutureDeactivate && le.IsDeactivateOrDestroy() {
				level = max(0, level-1)
			}
		}
	}
	if level < 0 {
		return 0
	}
	return level
}

func (l *LiveBoxes) isNextEventADestroy(event Event) bool {
	l.ensureLevelCache()
	info, ok := l.eventInfo[event]
	if _, isMessage := event.(*Message); !ok || !isMessage {
		return false
	}
	// Several life events may follow the message (for example a
	// deactivate then a destroy): the destroy is not always the
	// first one
	for k := info[0] + 1; k < len(l.events); k++ {
		next := l.events[k]
		if _, isNote := next.(*Note); isNote {
			continue
		}
		le, isLife := next.(*LifeEvent)
		if !isLife {
			return false
		}
		if le.IsDestroy(l.p) {
			return true
		}
	}
	return false
}

func (l *LiveBoxes) activateColor(event Event) *klimt.Fashion {
	if le, ok := event.(*LifeEvent); ok && le.IsActivate() {
		return le.SpecificColors()
	}
	l.ensureLevelCache()
	info, ok := l.eventInfo[event]
	_, isMessage := event.(*Message)
	_, isExo := event.(*MessageExo)
	if !ok || (!isMessage && !isExo) {
		return nil
	}
	for k := info[0] + 1; k < len(l.events); k++ {
		next := l.events[k]
		if _, isNote := next.(*Note); isNote {
			continue
		}
		le, isLife := next.(*LifeEvent)
		if !isLife || le.Message() != event.(AbstractMessage) {
			return nil
		}
		if le.IsActivate() && le.Participant() == l.p {
			return le.SpecificColors()
		}
	}
	return nil
}

// Stairs returns the steps of the activation levels between createY and
// totalHeight.
func (l *LiveBoxes) Stairs(createY, totalHeight float64) *Stairs {
	stair := &Stairs{}
	indent := 0
	var lastMessage AbstractMessage
	var position *float64
	seenActivate := false
	seenDeactivate := false
	for _, event := range l.events {
		if _, ok := event.(*Note); ok {
			// This would be a participant positioned Note, not a Message based Note
			lastMessage = nil
			seenActivate = false
			seenDeactivate = false
		}

		var potentialPosition *float64
		if v, ok := l.eventsStep[event]; ok {
			potentialPosition = &v
		}

		if le, ok := event.(*LifeEvent); position == nil || lastMessage == nil {
			position = potentialPosition
		} else if ok {
			if le.DealWith(l.p) {
				if le.Message() == nil || !le.Message().Base().IsParallelWith(lastMessage) {
					position = potentialPosition
				} else if (le.IsActivate() && seenDeactivate) || (le.IsDeactivate() && seenActivate) {
					position = potentialPosition
				}
				seenActivate = seenActivate || le.IsActivate()
				seenDeactivate = seenDeactivate || le.IsDeactivate()
			} else {
				continue
			}
		} else {
			position = potentialPosition
		}

		if msg, ok := event.(AbstractMessage); ok {
			if msg.Base().IsParallelWith(lastMessage) {
				continue
			}
			seenActivate = false
			seenDeactivate = false
			lastMessage = msg
		}

		if position != nil {
			indent = l.LevelAt(event, ConsiderFutureDeactivate)
			activateColor := l.activateColor(event)
			var styleBuilder *klimt.StyleBuilder
			if msg, ok := event.(AbstractMessage); ok {
				styleBuilder = msg.Base().StyleBuilder()
			}
			if le, ok := event.(*LifeEvent); ok {
				styleBuilder = le.StyleBuilder()
			}
			step := NewStep(jmath.Max(createY, *position), l.isNextEventADestroy(event), indent, activateColor,
				styleBuilder)
			stair.AddStep(step)
		}
	}
	stair.AddStep(NewStep(totalHeight, false, indent, nil, nil))
	return stair
}

// MaxValue returns the highest activation level.
func (l *LiveBoxes) MaxValue() int {
	maxLevel := 0
	level := 0
	for _, current := range l.events {
		if le, ok := current.(*LifeEvent); ok {
			if le.Participant() == l.p && le.IsActivate() {
				level++
			}
			if level > maxLevel {
				maxLevel = level
			}
			if le.Participant() == l.p && le.IsDeactivateOrDestroy() {
				level--
			}
		}
	}
	return maxLevel
}

// MaxPosition returns how far right of the lifeline the activation boxes
// reach.
func (l *LiveBoxes) MaxPosition(stringBounder font.StringBounder) float64 {
	maxLevel := l.MaxValue()
	drawer := NewLiveBoxesDrawer(klimt.NewSimpleContext2D(true), l.skin, l.skinParam, &l.delays, l.p.Stereotype())
	return drawer.Width(stringBounder) / 2.0 * float64(maxLevel)
}

// DrawBoxes draws the activation boxes and the destructions.
func (l *LiveBoxes) DrawBoxes(ug klimt.UGraphic, context klimt.Context2D, createY, endY float64) {
	stairs := l.Stairs(createY, endY)
	maxIndent := stairs.MaxIndent()
	for i := 1; i <= maxIndent; i++ {
		l.drawOneLevel(ug, i, stairs, context)
	}
	// Destroys at indent 0 are always handled here, regardless of the
	// participant's global max indent; drawn last so that the cross is
	// never painted over by an activation box
	l.drawDestroys(ug, stairs, context)
}

func (l *LiveBoxes) drawDestroys(ug klimt.UGraphic, stairs *Stairs, context klimt.Context2D) {
	drawer := NewLiveBoxesDrawer(context, l.skin, l.skinParam, &l.delays, l.p.Stereotype())
	for _, yposition := range stairs.values {
		if yposition.Indent() == 0 {
			drawer.DrawDestroyIfNeeded(ug, yposition)
		}
	}
}

func (l *LiveBoxes) drawOneLevel(ug klimt.UGraphic, levelToDraw int, stairs *Stairs, context klimt.Context2D) {
	drawer := NewLiveBoxesDrawer(context, l.skin, l.skinParam, &l.delays, l.p.Stereotype())
	ug = ug.Apply(klimt.UTranslateDx(float64(levelToDraw-1) * drawer.Width(ug.StringBounder()) / 2.0))

	pending := true
	for i, yposition := range stairs.values {
		indent := yposition.Indent()
		hasNext := i+1 < len(stairs.values)
		if pending && indent == levelToDraw {
			drawer.AddStart(yposition.Value(), yposition.Colors(), yposition.StyleBuilder())
			pending = false
		} else if !pending && (!hasNext || indent < levelToDraw) {
			drawer.DoDrawing(ug, yposition.Value())
			// Indent-0 destroys are handled by drawDestroys
			if indent > 0 {
				drawer.DrawDestroyIfNeeded(ug, yposition)
			}
			pending = true
		}
	}
}

// DelayOn records a delay on the lifeline.
func (l *LiveBoxes) DelayOn(y, height float64) { l.delays.put(y, height) }

// LiveBoxesDrawer draws the activation boxes of one level.
type LiveBoxesDrawer struct {
	y1            float64
	symbolContext *klimt.Fashion
	styleBuilder  *klimt.StyleBuilder

	cross        klimt.Component
	context      klimt.Context2D
	skin         *rose.Rose
	skinParam    klimt.ISkinParam
	compForWidth klimt.Component
	delays       []Segment
	stereotype   *klimt.Stereotype
}

// NewLiveBoxesDrawer returns the drawer; delays maps the y of each delay to
// its height.
func NewLiveBoxesDrawer(context klimt.Context2D, skin *rose.Rose, skinParam klimt.ISkinParam,
	delays *sortedFloatMap[float64], stereotype *klimt.Stereotype) *LiveBoxesDrawer {
	d := &LiveBoxesDrawer{stereotype: stereotype, context: context, skin: skin, skinParam: skinParam}
	d.cross = skin.CreateComponent(
		[]*klimt.Style{klimt.ComponentTypeDestroy.StyleSignature().MergedStyle(skinParam.CurrentStyleBuilder())},
		klimt.ComponentTypeDestroy, nil, skinParam, nil)
	d.compForWidth = skin.CreateComponent(
		[]*klimt.Style{klimt.ComponentTypeActivationBoxCloseClose.StyleSignature().
			MergedStyle(skinParam.CurrentStyleBuilder())},
		klimt.ComponentTypeActivationBoxCloseClose, nil, skinParam, nil)
	for i, key := range delays.keys {
		d.delays = append(d.delays, NewSegment(key, key+delays.values[i]))
	}
	return d
}

// Width returns the width of an activation box.
func (d *LiveBoxesDrawer) Width(stringBounder font.StringBounder) float64 {
	return d.compForWidth.PreferredWidth(stringBounder)
}

// AddStart opens a box at y1.
func (d *LiveBoxesDrawer) AddStart(y1 float64, symbolContext *klimt.Fashion, styleBuilder *klimt.StyleBuilder) {
	d.y1 = y1
	d.symbolContext = symbolContext
	d.styleBuilder = styleBuilder
}

// DoDrawing draws the box opened by AddStart down to yposition, cut by the
// delays.
func (d *LiveBoxesDrawer) DoDrawing(ug klimt.UGraphic, yposition float64) {
	full := NewSegment(d.y1, yposition)
	segments := full.CutSegmentIfNeed(d.delays)
	typ := klimt.ComponentTypeActivationBoxCloseClose
	if len(segments) > 1 {
		typ = klimt.ComponentTypeActivationBoxCloseOpen
	}
	for i, seg := range segments {
		if i+1 == len(segments) && typ != klimt.ComponentTypeActivationBoxCloseClose {
			typ = klimt.ComponentTypeActivationBoxOpenClose
		}
		d.drawInternal(ug, seg.Pos1(), seg.Pos2(), typ)
		typ = klimt.ComponentTypeActivationBoxOpenOpen
	}
	d.y1 = math.MaxFloat64
}

// DrawDestroyIfNeeded draws the destruction cross of a step.
func (d *LiveBoxesDrawer) DrawDestroyIfNeeded(ug klimt.UGraphic, step *Step) {
	if step.IsDestroy() {
		dimCross := d.cross.PreferredDimension(ug.StringBounder())
		d.cross.DrawU(ug.Apply(klimt.NewUTranslate(-dimCross.Width()/2, step.Value()-dimCross.Height()/2)), nil,
			d.context)
	}
}

func (d *LiveBoxesDrawer) drawInternal(ug klimt.UGraphic, ya, yb float64, typ klimt.ComponentType) {
	width := d.Width(ug.StringBounder())
	area := klimt.AreaCreate(width, yb-ya)
	var back color.HColor
	if d.symbolContext != nil {
		back = d.symbolContext.BackColor()
	}
	skinParam2 := klimt.NewSkinParamBackcolored(d.skinParam, back, nil, false)

	currentStyleBuilder := d.styleBuilder
	if currentStyleBuilder == nil {
		currentStyleBuilder = d.skinParam.CurrentStyleBuilder()
	}

	style := typ.StyleSignature().WithTOBECHANGED(d.stereotype).MergedStyle(currentStyleBuilder)
	if style == nil {
		if d.symbolContext != nil {
			skinParam2.ForceColor(klimt.ColorParamSequenceLifeLineBorder, d.symbolContext.ForeColor())
		}
	} else {
		style = style.EventuallyOverrideFashion(d.symbolContext)
	}
	comp := d.skin.CreateComponent([]*klimt.Style{style}, typ, nil, skinParam2, nil)
	comp.DrawU(ug.Apply(klimt.NewUTranslate(-width/2, ya)), area, d.context)
}

// Stairs are the successive activation levels of a participant.
type Stairs struct {
	values []*Step
}

// AddStep adds a step; a step above the last one is ignored.
func (s *Stairs) AddStep(step *Step) {
	if step.Indent() < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if len(s.values) > 0 {
		lastY := s.values[len(s.values)-1].Value()
		if step.Value() < lastY {
			return
		}
	}
	s.values = append(s.values, step)
}

// MaxIndent returns the highest level, math.MinInt32 for no step.
func (s *Stairs) MaxIndent() int {
	maxIndent := math.MinInt32
	for _, step := range s.values {
		if v := step.Indent(); v > maxIndent {
			maxIndent = v
		}
	}
	return maxIndent
}

// Steps returns the steps.
func (s *Stairs) Steps() []*Step { return s.values }

// Step is an activation level from a y on.
type Step struct {
	value        float64
	destroy      bool
	indent       int
	color        *klimt.Fashion
	styleBuilder *klimt.StyleBuilder
}

// NewStep returns a step; a negative indent panics, as upstream.
func NewStep(value float64, destroy bool, indent int, color *klimt.Fashion, styleBuilder *klimt.StyleBuilder) *Step {
	if indent < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &Step{value: value, destroy: destroy, indent: indent, color: color, styleBuilder: styleBuilder}
}

// Value returns the y of the step.
func (s *Step) Value() float64 { return s.value }

// IsDestroy reports whether the participant is destroyed at the step.
func (s *Step) IsDestroy() bool { return s.destroy }

// Indent returns the level.
func (s *Step) Indent() int { return s.indent }

// Colors returns the colors of the activation, nil for none.
func (s *Step) Colors() *klimt.Fashion { return s.color }

// StyleBuilder returns the style sheet of the step, nil for none.
func (s *Step) StyleBuilder() *klimt.StyleBuilder { return s.styleBuilder }

// MutingLine draws the lifeline of a participant, with its delays.
type MutingLine struct {
	skin         *rose.Rose
	skinParam    klimt.ISkinParam
	delays       sortedFloatMap[float64]
	styleBuilder *klimt.StyleBuilder
	participant  *Participant
}

// NewMutingLine returns the lifeline of participant.
func NewMutingLine(skin *rose.Rose, skinParam klimt.ISkinParam, _ []Event, participant *Participant) *MutingLine {
	return &MutingLine{participant: participant, skin: skin, skinParam: skinParam,
		styleBuilder: skinParam.CurrentStyleBuilder()}
}

// DrawLine draws the lifeline from createY to endY, dashed over the delays
// inside it.
func (m *MutingLine) DrawLine(ug klimt.UGraphic, context klimt.Context2D, createY, endY float64) {
	defaultLineType := klimt.ComponentTypeParticipantLine
	if m.delays.size() > 0 {
		y := createY
		for i, key := range m.delays.keys {
			value := m.delays.values[i]
			// Only the delays fully inside [createY, endY] are relevant
			if key >= createY && key+value <= endY {
				m.drawInternal(ug, context, y, key, defaultLineType)
				m.drawInternal(ug, context, key, key+value, klimt.ComponentTypeDelayLine)
				y = key + value
			}
		}
		m.drawInternal(ug, context, y, endY, defaultLineType)
	} else {
		m.drawInternal(ug, context, createY, endY, defaultLineType)
	}
}

func (m *MutingLine) drawInternal(ug klimt.UGraphic, context klimt.Context2D, y1, y2 float64,
	defaultLineType klimt.ComponentType) {
	if y2 == y1 {
		return
	}
	if y2 < y1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	// The style is merged with the participant stereotype, so that
	// stereotyped rules apply to the life line
	style := defaultLineType.StyleSignature().WithTOBECHANGED(m.participant.Stereotype()).MergedStyle(m.styleBuilder)
	comp := m.skin.CreateComponent([]*klimt.Style{style}, defaultLineType, nil, m.skinParam,
		m.participant.Display(m.skinParam.ForceSequenceParticipantUnderlined()))
	dim := comp.PreferredDimension(ug.StringBounder())
	area := klimt.AreaCreate(dim.Width(), y2-y1)
	comp.DrawU(ug.Apply(klimt.UTranslateDy(y1)), area, context)
}

// DelayOn records a delay.
func (m *MutingLine) DelayOn(y, height float64) { m.delays.put(y, height) }
