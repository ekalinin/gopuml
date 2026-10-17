// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Doll.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/Dolls.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Doll is a box ("box" ... "end box") around participants, drawn by
// Teoz.
type Doll struct {
	participants  []*Participant
	alls          *Dolls
	englober      *ParticipantEnglober
	styleBuilder  *klimt.StyleBuilder
	tileArguments *TileArguments
}

// newDollTeoz is upstream's Doll.createTeoz.
func newDollTeoz(englober *ParticipantEnglober, tileArguments *TileArguments, alls *Dolls) *Doll {
	if englober == nil || tileArguments == nil {
		uml.ThrowNPE("")
	}
	return &Doll{englober: englober, tileArguments: tileArguments,
		styleBuilder: tileArguments.SkinParam().CurrentStyleBuilder(), alls: alls}
}

// StyleSignature returns the style signature of boxes.
func (d *Doll) StyleSignature() klimt.StyleSignature {
	return klimt.ComponentTypeEnglober.StyleSignature()
}

// UsedStyles returns the style of the box, with its color.
func (d *Doll) UsedStyles() []*klimt.Style {
	tmp := d.StyleSignature().WithTOBECHANGED(d.englober.Stereotype()).MergedStyle(d.styleBuilder)
	if tmp != nil {
		tmp = tmp.EventuallyOverrideColor(klimt.PNameBackGroundColor, d.englober.BoxColor())
	}
	return []*klimt.Style{tmp}
}

// ParticipantEnglober returns the box definition.
func (d *Doll) ParticipantEnglober() *ParticipantEnglober { return d.englober }

func (d *Doll) component() klimt.Component {
	englober := d.ParticipantEnglober()
	s := d.tileArguments.SkinParam()
	if englober.BoxColor() != nil {
		s = klimt.NewSkinParamBackcolored(d.tileArguments.SkinParam(), englober.BoxColor(), nil, false)
	}
	return d.tileArguments.Skin().CreateComponent(d.UsedStyles(), klimt.ComponentTypeEnglober, nil, s,
		englober.Title())
}

// TitlePreferredHeight returns the height of the title of the box.
func (d *Doll) TitlePreferredHeight() float64 {
	comp := d.tileArguments.Skin().CreateComponent(d.UsedStyles(), klimt.ComponentTypeEnglober, nil,
		d.tileArguments.SkinParam(), d.ParticipantEnglober().Title())
	return comp.PreferredHeight(d.tileArguments.StringBounder())
}

func (d *Doll) first() *Participant { return d.participants[0] }

func (d *Doll) last() *Participant { return d.participants[len(d.participants)-1] }

func (d *Doll) posA(stringBounder font.StringBounder) Real {
	return d.firstLivingSpace().PosA(stringBounder)
}

func (d *Doll) posALive(stringBounder font.StringBounder) Real {
	return d.firstLivingSpace().PosALive(stringBounder)
}

func (d *Doll) posB(stringBounder font.StringBounder) Real {
	return d.firstLivingSpace().PosB(stringBounder)
}

func (d *Doll) posD(stringBounder font.StringBounder) Real {
	return d.lastLivingSpace().PosD(stringBounder)
}

func (d *Doll) posE(stringBounder font.StringBounder) Real {
	return d.lastLivingSpace().PosE(stringBounder)
}

func (d *Doll) posELive(stringBounder font.StringBounder) Real {
	return d.lastLivingSpace().PosELive(stringBounder)
}

func (d *Doll) posAA(stringBounder font.StringBounder) Real {
	previous := d.tileArguments.LivingSpaces().Previous(d.firstLivingSpace())
	if previous == nil {
		return d.tileArguments.XOrigin()
	}
	return previous.PosD(stringBounder)
}

func (d *Doll) firstLivingSpace() *LivingSpace { return d.tileArguments.LivingSpace(d.first()) }

func (d *Doll) lastLivingSpace() *LivingSpace { return d.tileArguments.LivingSpace(d.last()) }

// Contains reports whether p is in the box.
func (d *Doll) Contains(p *Participant) bool { return slices.Contains(d.participants, p) }

// AddParticipant adds a participant to the box.
func (d *Doll) AddParticipant(p *Participant) {
	if p == nil {
		uml.ThrowNPE("")
	}
	d.participants = append(d.participants, p)
}

func (d *Doll) titleWidth() float64 {
	return d.component().PreferredWidth(d.tileArguments.StringBounder())
}

// DrawMe draws the box, below the titles of the boxes containing it.
func (d *Doll) DrawMe(ug klimt.UGraphic, height float64, context klimt.Context2D, group *Doll) {
	stringBounder := ug.StringBounder()
	x1 := d.posA(stringBounder).CurrentValue() - 4
	x2 := d.posE(stringBounder).CurrentValue() + 4

	for current := group; current != nil; current = current.Parent() {
		titlePreferredHeight := current.TitlePreferredHeight()
		ug = ug.Apply(klimt.UTranslateDy(titlePreferredHeight))
		height -= titlePreferredHeight
	}

	dim := geom.NewXDimension2D(x2-x1, height)
	d.component().DrawU(ug.Apply(klimt.NewUTranslate(x1, 1)), klimt.NewArea(dim), context)
}

// Parent returns the box containing this one, nil for none.
func (d *Doll) Parent() *Doll {
	parent := d.ParticipantEnglober().Parent()
	if parent == nil {
		return nil
	}
	return d.alls.get(parent)
}

// AddInternalConstraints widens the first and last columns for the title.
func (d *Doll) AddInternalConstraints(stringBounder font.StringBounder) {
	titleWidth := d.titleWidth()
	x1 := d.posB(stringBounder).CurrentValue()
	x2 := d.posD(stringBounder).CurrentValue()
	actualWidth := x2 - x1
	marginX := (titleWidth + 10 - actualWidth) / 2
	if marginX > 0 {
		d.firstLivingSpace().EnsureMarginBefore(marginX)
		d.lastLivingSpace().EnsureMarginAfter(marginX)
	}
}

// AddNeighborConstraintBefore keeps the frame of the box clear of the
// previous participant (upstream's issue #2791).
func (d *Doll) AddNeighborConstraintBefore(stringBounder font.StringBounder) {
	padding := d.padding()
	d.posALive(stringBounder).EnsureBiggerThan(d.posAA(stringBounder).AddFixed(10 + padding.Left()))
}

// AddConstraintAfter keeps the next participant clear of the frame of the
// box.
func (d *Doll) AddConstraintAfter(stringBounder font.StringBounder) {
	next := d.tileArguments.LivingSpaces().Next(d.lastLivingSpace())
	if next == nil {
		return
	}
	padding := d.padding()
	next.PosALive(stringBounder).EnsureBiggerThan(
		d.posELive(stringBounder).AddFixed(20 + padding.Left() + padding.Right()))
}

func (d *Doll) padding() klimt.ClockwiseTopRightBottomLeft {
	return d.tileArguments.SkinParam().PaddingTOBEREMOVED(klimt.PaddingParamBox)
}

// MinX returns the left of the box.
func (d *Doll) MinX(stringBounder font.StringBounder) Real { return d.posA(stringBounder) }

// MaxX returns the right of the box.
func (d *Doll) MaxX(stringBounder font.StringBounder) Real { return d.posE(stringBounder).AddFixed(10) }

// Dolls are the boxes of a diagram, in order of first participant.
type Dolls struct {
	order []*ParticipantEnglober
	alls  map[*ParticipantEnglober]*Doll
}

// NewDolls returns the boxes of the participants of the arguments.
func NewDolls(tileArguments *TileArguments) *Dolls {
	d := &Dolls{alls: map[*ParticipantEnglober]*Doll{}}
	for _, p := range tileArguments.LivingSpaces().Participants() {
		englober := tileArguments.LivingSpaces().Get(p).Englober()
		if englober != nil {
			for _, pe := range englober.Genealogy() {
				d.addParticipant(p, pe, tileArguments)
			}
		}
	}
	return d
}

func (d *Dolls) get(englober *ParticipantEnglober) *Doll { return d.alls[englober] }

func (d *Dolls) values() []*Doll {
	result := make([]*Doll, len(d.order))
	for i, e := range d.order {
		result[i] = d.alls[e]
	}
	return result
}

func (d *Dolls) addParticipant(p *Participant, englober *ParticipantEnglober, tileArguments *TileArguments) {
	already := d.alls[englober]
	if already == nil {
		already = newDollTeoz(englober, tileArguments, d)
		d.alls[englober] = already
		d.order = append(d.order, englober)
	}
	already.AddParticipant(p)
}

// Size returns the number of boxes.
func (d *Dolls) Size() int { return len(d.order) }

// OffsetForEnglobers returns the height of the deepest stack of box
// titles.
func (d *Dolls) OffsetForEnglobers(font.StringBounder) float64 {
	result := 0.0
	for _, doll := range d.values() {
		height := 0.0
		for current := doll; current != nil; current = current.Parent() {
			height += current.TitlePreferredHeight()
		}
		if height > result {
			result = height
		}
	}
	return result
}

// AddConstraints registers the constraints of every box: the title widths,
// then the clearance of the previous and next participants.
func (d *Dolls) AddConstraints(stringBounder font.StringBounder) {
	for _, doll := range d.values() {
		doll.AddInternalConstraints(stringBounder)
	}
	for _, doll := range d.values() {
		doll.AddNeighborConstraintBefore(stringBounder)
	}
	for _, doll := range d.values() {
		doll.AddConstraintAfter(stringBounder)
	}
}

// DrawEnglobers draws the boxes.
func (d *Dolls) DrawEnglobers(ug klimt.UGraphic, height float64, context klimt.Context2D) {
	for _, doll := range d.values() {
		doll.DrawMe(ug, height, context, doll.Parent())
	}
}

// MinX returns the left of the leftmost box; it panics without box.
func (d *Dolls) MinX(stringBounder font.StringBounder) Real {
	if d.Size() == 0 {
		uml.Throw(uml.IllegalStateException, "")
	}
	var result []Real
	for _, doll := range d.values() {
		result = append(result, doll.MinX(stringBounder))
	}
	return RealUtilsMin(result...)
}

// MaxX returns the right of the rightmost box; it panics without box.
func (d *Dolls) MaxX(stringBounder font.StringBounder) Real {
	if d.Size() == 0 {
		uml.Throw(uml.IllegalStateException, "")
	}
	var result []Real
	for _, doll := range d.values() {
		result = append(result, doll.MaxX(stringBounder))
	}
	return RealUtilsMax(result...)
}
