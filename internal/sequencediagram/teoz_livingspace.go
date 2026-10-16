// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/LivingSpace.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/LivingSpaces.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"slices"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// sortedFloatMap is a Java TreeMap keyed by doubles: entries are kept in
// key order and putting an existing key replaces its value.
type sortedFloatMap[V any] struct {
	keys   []float64
	values []V
}

func (m *sortedFloatMap[V]) put(key float64, value V) {
	i, found := slices.BinarySearchFunc(m.keys, key, compareJavaDouble)
	if found {
		m.values[i] = value
		return
	}
	m.keys = slices.Insert(m.keys, i, key)
	m.values = slices.Insert(m.values, i, value)
}

func (m *sortedFloatMap[V]) size() int { return len(m.keys) }

// compareJavaDouble is Double.compareTo for the keys of a TreeMap.
func compareJavaDouble(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// LivingSpace is the column of a participant: its horizontal positions on
// the X line (posB the left of the head, posC the lifeline, posD the right
// of the head), its lifeline and its activation boxes.
type LivingSpace struct {
	p          *Participant
	skinParam  klimt.ISkinParam
	headType   klimt.ComponentType
	tailType   klimt.ComponentType
	mutingLine *MutingLine
	rose       *rose.Rose
	liveboxes  *LiveBoxes

	posB Real
	posC Real
	posD Real

	create bool
	// aliveChanges records, by y, true for a creation and false for a
	// destruction.
	aliveChanges sortedFloatMap[bool]

	englober *ParticipantEnglober

	marginBefore float64
	marginAfter  float64

	// ascii holds the positions on the text X line (see teoz_ascii.go).
	ascii livingSpaceASCII
}

// LevelAt returns the activation level of the participant at the event of
// a tile.
func (s *LivingSpace) LevelAt(tile Tile, mode EventsHistoryMode) int {
	return s.liveboxes.LevelAt(tile.Event(), mode)
}

// AddStepForLivebox records the y of an event for the activation boxes.
func (s *LivingSpace) AddStepForLivebox(event Event, y float64) { s.liveboxes.AddStep(event, y) }

// String is upstream's toString.
func (s *LivingSpace) String() string {
	cur := func(pos Real) string {
		if pos == nil {
			return "null"
		}
		return jfmt.DoubleToString(pos.CurrentValue())
	}
	return s.p.Code() + " B=" + jfmt.DoubleToString(s.posB.CurrentValue()) + "/C=" + cur(s.posC) + "/D=" + cur(s.posD)
}

// NewLivingSpace returns the column of p starting at position.
func NewLivingSpace(p *Participant, englober *ParticipantEnglober, skin *rose.Rose, skinParam klimt.ISkinParam,
	position Real, events []Event) *LivingSpace {
	s := &LivingSpace{p: p, skinParam: skinParam, englober: englober, posB: position, rose: rose.NewRose()}
	switch p.Type() {
	case TypeParticipant:
		s.headType, s.tailType = klimt.ComponentTypeParticipantHead, klimt.ComponentTypeParticipantTail
	case TypeActor:
		s.headType, s.tailType = klimt.ComponentTypeActorHead, klimt.ComponentTypeActorTail
	case TypeBoundary:
		s.headType, s.tailType = klimt.ComponentTypeBoundaryHead, klimt.ComponentTypeBoundaryTail
	case TypeControl:
		s.headType, s.tailType = klimt.ComponentTypeControlHead, klimt.ComponentTypeControlTail
	case TypeEntity:
		s.headType, s.tailType = klimt.ComponentTypeEntityHead, klimt.ComponentTypeEntityTail
	case TypeQueue:
		s.headType, s.tailType = klimt.ComponentTypeQueueHead, klimt.ComponentTypeQueueTail
	case TypeDatabase:
		s.headType, s.tailType = klimt.ComponentTypeDatabaseHead, klimt.ComponentTypeDatabaseTail
	case TypeCollections:
		s.headType, s.tailType = klimt.ComponentTypeCollectionsHead, klimt.ComponentTypeCollectionsTail
	default:
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.mutingLine = NewMutingLine(skin, skinParam, events, p)
	s.liveboxes = NewLiveBoxes(p, events, skin, skinParam)
	return s
}

// DrawLineAndLiveboxes draws the lifeline over its alive segments, then
// the activation boxes.
func (s *LivingSpace) DrawLineAndLiveboxes(ug klimt.UGraphic, height float64, context klimt.Context2D) {
	// The line is drawn only over the alive segments: a participant may be
	// created and destroyed several times
	alive := !s.create
	aliveSince := 0.0
	for i, key := range s.aliveChanges.keys {
		value := s.aliveChanges.values[i]
		if !alive && value {
			aliveSince = key
		} else if alive && !value {
			s.mutingLine.DrawLine(ug, context, aliveSince, key)
		}
		alive = value
	}
	if alive {
		s.mutingLine.DrawLine(ug, context, aliveSince, height)
	}
	s.liveboxes.DrawBoxes(ug, context, s.firstCreateY(), height)
}

func (s *LivingSpace) firstCreateY() float64 {
	for i, key := range s.aliveChanges.keys {
		if s.aliveChanges.values[i] {
			return key
		}
	}
	return 0
}

// DrawHead draws the head of the participant.
func (s *LivingSpace) DrawHead(ug klimt.UGraphic, context klimt.Context2D, verticalAlignment geom.VerticalAlignment,
	horizontalAlignment geom.HorizontalAlignment) {
	s.drawHeadOrTail(ug, context, verticalAlignment, horizontalAlignment, s.headType)
}

// DrawTail draws the tail (footbox) of the participant.
func (s *LivingSpace) DrawTail(ug klimt.UGraphic, context klimt.Context2D, verticalAlignment geom.VerticalAlignment,
	horizontalAlignment geom.HorizontalAlignment) {
	s.drawHeadOrTail(ug, context, verticalAlignment, horizontalAlignment, s.tailType)
}

func (s *LivingSpace) drawHeadOrTail(ug klimt.UGraphic, context klimt.Context2D,
	verticalAlignment geom.VerticalAlignment, horizontalAlignment geom.HorizontalAlignment, typ klimt.ComponentType) {
	if s.create && verticalAlignment == geom.VerticalAlignmentBottom {
		return
	}
	comp := s.rose.CreateComponentParticipant(s.p, typ, nil, s.skinParam,
		s.p.Display(s.skinParam.ForceSequenceParticipantUnderlined()))
	dim := comp.PreferredDimension(ug.StringBounder())
	if horizontalAlignment == geom.HorizontalAlignmentRight {
		ug = ug.Apply(klimt.UTranslateDx(-dim.Width()))
	}
	if verticalAlignment == geom.VerticalAlignmentCenter {
		ug = ug.Apply(klimt.UTranslateDy(-dim.Height() / 2))
	}
	area := klimt.NewArea(dim)
	url := s.p.URL()
	if url != nil {
		ug.StartURL(url)
	}
	comp.DrawU(ug, area, context)
	if url != nil {
		ug.CloseURL()
	}
}

// HeadPreferredDimension returns the size of the head.
func (s *LivingSpace) HeadPreferredDimension(stringBounder font.StringBounder) geom.XDimension2D {
	comp := s.rose.CreateComponentParticipant(s.p, s.headType, nil, s.skinParam,
		s.p.Display(s.skinParam.ForceSequenceParticipantUnderlined()))
	return comp.PreferredDimension(stringBounder)
}

func (s *LivingSpace) preferredWidth(stringBounder font.StringBounder) float64 {
	return s.HeadPreferredDimension(stringBounder).Width()
}

// PosC returns the position of the lifeline.
func (s *LivingSpace) PosC(stringBounder font.StringBounder) Real {
	if s.posC == nil {
		s.posC = s.posB.AddFixed(s.preferredWidth(stringBounder) / 2)
	}
	return s.posC
}

// PosC2 returns the right of the widest activation box.
func (s *LivingSpace) PosC2(stringBounder font.StringBounder) Real {
	delta := s.liveboxes.MaxPosition(stringBounder)
	return s.PosC(stringBounder).AddFixed(delta)
}

// PosD returns the right of the head.
func (s *LivingSpace) PosD(stringBounder font.StringBounder) Real {
	if s.posD == nil {
		s.posD = s.posB.AddFixed(s.preferredWidth(stringBounder))
	}
	return s.posD
}

// PosB returns the left of the head.
func (s *LivingSpace) PosB(font.StringBounder) Real { return s.posB }

// Participant returns the participant.
func (s *LivingSpace) Participant() *Participant { return s.p }

// GoCreateAt records the creation of the participant at y.
func (s *LivingSpace) GoCreateAt(y float64) {
	s.aliveChanges.put(y, true)
	s.create = true
}

// GoCreate records that the participant is created by a message.
func (s *LivingSpace) GoCreate() { s.create = true }

// GoDestroy records the destruction of the participant at y.
func (s *LivingSpace) GoDestroy(y float64) { s.aliveChanges.put(y, false) }

// DelayOn records a delay on the lifeline.
func (s *LivingSpace) DelayOn(y, height float64) {
	s.mutingLine.DelayOn(y, height)
	s.liveboxes.DelayOn(y, height)
}

// Englober returns the box of the participant, nil for none.
func (s *LivingSpace) Englober() *ParticipantEnglober { return s.englober }

// EnsureMarginBefore widens the space before the head.
func (s *LivingSpace) EnsureMarginBefore(margin float64) {
	if margin < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.marginBefore = jmath.Max(s.marginBefore, margin)
}

// EnsureMarginAfter widens the space after the head.
func (s *LivingSpace) EnsureMarginAfter(margin float64) {
	if margin < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.marginAfter = jmath.Max(s.marginAfter, margin)
}

// PosA returns the left of the column, with the margin known now.
func (s *LivingSpace) PosA(stringBounder font.StringBounder) Real {
	return s.PosB(stringBounder).AddFixed(-s.marginBefore)
}

// PosE returns the right of the column, with the margin known now.
func (s *LivingSpace) PosE(stringBounder font.StringBounder) Real {
	return s.PosD(stringBounder).AddFixed(s.marginAfter)
}

// PosALive is PosA with the margin read when the value is (the box
// neighbour checks of upstream's issue #2791).
func (s *LivingSpace) PosALive(stringBounder font.StringBounder) Real {
	return RealUtilsWithLiveOffset(s.PosB(stringBounder), func() float64 { return -s.marginBefore })
}

// PosELive is PosE with the margin read when the value is.
func (s *LivingSpace) PosELive(stringBounder font.StringBounder) Real {
	return RealUtilsWithLiveOffset(s.PosD(stringBounder), func() float64 { return s.marginAfter })
}

// LivingSpaces are the columns of the participants, in display order.
type LivingSpaces struct {
	order []*Participant
	all   map[*Participant]*LivingSpace
}

// NewLivingSpaces returns no column.
func NewLivingSpaces() *LivingSpaces { return &LivingSpaces{all: map[*Participant]*LivingSpace{}} }

// Values returns the columns in order.
func (l *LivingSpaces) Values() []*LivingSpace {
	result := make([]*LivingSpace, len(l.order))
	for i, p := range l.order {
		result[i] = l.all[p]
	}
	return result
}

// AddConstraints keeps 10 pixels between neighbouring columns.
func (l *LivingSpaces) AddConstraints(stringBounder font.StringBounder) {
	var previous *LivingSpace
	for _, current := range l.Values() {
		if previous != nil {
			point1 := previous.PosE(stringBounder)
			point2 := current.PosA(stringBounder)
			point2.EnsureBiggerThan(point1.AddFixed(10))
		}
		previous = current
	}
}

// Previous returns the column before element, nil for none.
func (l *LivingSpaces) Previous(element *LivingSpace) *LivingSpace {
	var previous *LivingSpace
	for _, current := range l.Values() {
		if current == element {
			return previous
		}
		previous = current
	}
	return nil
}

// Next returns the column after element, nil for none.
func (l *LivingSpaces) Next(element *LivingSpace) *LivingSpace {
	values := l.Values()
	for i, current := range values {
		if current == element && i+1 < len(values) {
			return values[i+1]
		}
	}
	return nil
}

// Participants returns the participants in order.
func (l *LivingSpaces) Participants() []*Participant { return l.order }

// Put adds or replaces the column of a participant.
func (l *LivingSpaces) Put(participant *Participant, livingSpace *LivingSpace) {
	if _, ok := l.all[participant]; !ok {
		l.order = append(l.order, participant)
	}
	l.all[participant] = livingSpace
}

// Get returns the column of a participant, nil for none.
func (l *LivingSpaces) Get(participant *Participant) *LivingSpace { return l.all[participant] }

// DrawHeads draws the heads (BOTTOM, at the top of the diagram) or the
// footboxes (TOP, with the tail components).
func (l *LivingSpaces) DrawHeads(ug klimt.UGraphic, context klimt.Context2D, verticalAlignment geom.VerticalAlignment) {
	stringBounder := ug.StringBounder()
	headHeight := l.HeadHeight(stringBounder)
	for _, livingSpace := range l.Values() {
		x := livingSpace.PosB(stringBounder).CurrentValue()
		y := 0.0
		if verticalAlignment == geom.VerticalAlignmentBottom {
			dimHead := livingSpace.HeadPreferredDimension(stringBounder)
			y = headHeight - dimHead.Height()
		}
		// TOP means that we are drawing the footboxes at the bottom of the
		// diagram (the real heads are aligned on BOTTOM): the tail component
		// is used there
		if verticalAlignment == geom.VerticalAlignmentTop {
			livingSpace.DrawTail(ug.Apply(klimt.NewUTranslate(x, y)), context, verticalAlignment,
				geom.HorizontalAlignmentLeft)
		} else {
			livingSpace.DrawHead(ug.Apply(klimt.NewUTranslate(x, y)), context, verticalAlignment,
				geom.HorizontalAlignmentLeft)
		}
	}
}

// HeadHeight returns the height of the tallest head.
func (l *LivingSpaces) HeadHeight(stringBounder font.StringBounder) float64 {
	headHeight := 0.0
	for _, livingSpace := range l.Values() {
		headDim := livingSpace.HeadPreferredDimension(stringBounder)
		headHeight = jmath.Max(headHeight, headDim.Height())
	}
	return headHeight
}

// DrawLifeLines draws the lifelines and activation boxes.
func (l *LivingSpaces) DrawLifeLines(ug klimt.UGraphic, height float64, context klimt.Context2D) {
	for _, livingSpace := range l.Values() {
		x := livingSpace.PosC(ug.StringBounder()).CurrentValue()
		livingSpace.DrawLineAndLiveboxes(ug.Apply(klimt.UTranslateDx(x)), height, context)
	}
}

// DelayOn records a delay on every lifeline.
func (l *LivingSpaces) DelayOn(y, height float64) {
	for _, livingSpace := range l.Values() {
		livingSpace.DelayOn(y, height)
	}
}

// Size returns the number of columns.
func (l *LivingSpaces) Size() int { return len(l.order) }
