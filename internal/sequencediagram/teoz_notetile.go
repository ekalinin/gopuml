// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/NoteTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/NotesTile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/uml"
)

// NoteTile is a note on participants ("note left of", "note over"...).
type NoteTile struct {
	abstractTile
	livingSpace1    *LivingSpace
	livingSpace2    *LivingSpace
	skin            *rose.Rose
	skinParam       klimt.ISkinParam
	note            *Note
	yGauge          *YGauge
	cachedComponent klimt.Component
}

// Event implements Tile.
func (t *NoteTile) Event() Event { return t.note }

// ContactPointRelative implements Tile: the vertical center of the note.
func (t *NoteTile) ContactPointRelative() float64 {
	return t.component(t.StringBounder()).PreferredHeight(t.StringBounder()) / 2
}

func newNoteTile(stringBounder font.StringBounder, livingSpace1, livingSpace2 *LivingSpace, note *Note,
	skin *rose.Rose, skinParam klimt.ISkinParam, currentY *YGauge) *NoteTile {
	t := &NoteTile{livingSpace1: livingSpace1, livingSpace2: livingSpace2, note: note, skin: skin,
		skinParam: skinParam}
	t.initTile(t, stringBounder)
	// Same chaining as CommunicationTile: a note published with "& note"
	// shares the contact line of the previous tile
	contactRelative := t.ContactPointRelative()
	height := t.PreferredHeight()
	if note.IsParallel() {
		t.yGauge = YGaugeCreateParallel(currentY, contactRelative, height)
	} else {
		t.yGauge = YGaugeCreateWithContact(currentY, contactRelative, height)
	}
	return t
}

// YGauge implements Tile.
func (t *NoteTile) YGauge() *YGauge { return t.yGauge }

func (t *NoteTile) component(font.StringBounder) klimt.Component {
	if t.cachedComponent == nil {
		t.cachedComponent = t.skin.CreateComponentNoteAt(t.note.UsedStyles(), noteComponentType(t.note.NoteStyle()),
			t.note.SkinParamBackcolored(t.skinParam), t.note.Display(), t.note.Colors(), t.note.Position())
	}
	return t.cachedComponent
}

// noteComponentType is upstream's NoteTile.getNoteComponentType.
func noteComponentType(noteStyle NoteStyle) klimt.ComponentType {
	switch noteStyle {
	case NoteStyleHexagonal:
		return klimt.ComponentTypeNoteHexagonal
	case NoteStyleBox:
		return klimt.ComponentTypeNoteBox
	}
	return klimt.ComponentTypeNote
}

// DrawU draws the note.
func (t *NoteTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	comp := t.component(stringBounder)
	dim := comp.PreferredDimension(stringBounder)
	x := t.x(stringBounder).CurrentValue()
	area := klimt.AreaCreate(t.usedWidth(stringBounder), dim.Height())

	ug = ug.Apply(klimt.UTranslateDx(x))
	comp.DrawU(ug, area, context2D(ug))
}

func (t *NoteTile) usedWidth(stringBounder font.StringBounder) float64 {
	comp := t.component(stringBounder)
	dim := comp.PreferredDimension(stringBounder)
	width := dim.Width()
	if t.note.Position() == NoteOverSeveral {
		x1 := t.livingSpace1.PosB(stringBounder).CurrentValue()
		x2 := t.livingSpace2.PosD(stringBounder).CurrentValue()
		w := x2 - x1
		if width < w {
			return w
		}
	}
	return width
}

func (t *NoteTile) x(stringBounder font.StringBounder) Real {
	position := t.note.Position()
	width := t.usedWidth(stringBounder)
	switch position {
	case NoteLeft:
		return t.livingSpace1.PosC(stringBounder).AddFixed(-width)
	case NoteRight:
		level := t.livingSpace1.LevelAt(t, IgnoreFutureDeactivate)
		dx := float64(level * LiveDeltaSize)
		return t.livingSpace1.PosC(stringBounder).AddFixed(dx)
	case NoteOverSeveral:
		x1 := t.livingSpace1.PosC(stringBounder)
		x2 := t.livingSpace2.PosC(stringBounder)
		return RealUtilsMiddle(x1, x2).AddFixed(-width / 2)
	case NoteOver:
		return t.livingSpace1.PosC(stringBounder).AddFixed(-width / 2)
	}
	uml.Throw(uml.UnsupportedOperationException, position.String())
	return nil
}

// PreferredHeight implements Tile.
func (t *NoteTile) PreferredHeight() float64 {
	return t.component(t.StringBounder()).PreferredDimension(t.StringBounder()).Height()
}

// AddConstraints widens the margins of the participants for the note
// (their box is drawn from the margins).
func (t *NoteTile) AddConstraints() { t.ensureOwnBoxClearsNote() }

func (t *NoteTile) ensureOwnBoxClearsNote() {
	stringBounder := t.StringBounder()

	overflowLeft := t.livingSpace1.PosB(stringBounder).CurrentValue() - t.MinX().CurrentValue()
	if overflowLeft > 0 {
		t.livingSpace1.EnsureMarginBefore(overflowLeft)
	}

	rightAnchor := t.livingSpace1
	if t.note.Position() == NoteOverSeveral {
		rightAnchor = t.livingSpace2
	}
	overflowRight := t.MaxX().CurrentValue() - rightAnchor.PosD(stringBounder).CurrentValue()
	if overflowRight > 0 {
		rightAnchor.EnsureMarginAfter(overflowRight)
	}
}

// leftAnchor returns the participant the left edge of the note is measured
// from.
func (t *NoteTile) leftAnchor() *LivingSpace { return t.livingSpace1 }

// leftOverhang returns how far left of the lifeline of its left anchor the
// note reaches (0 over several participants).
func (t *NoteTile) leftOverhang() float64 {
	switch t.note.Position() {
	case NoteLeft:
		return t.usedWidth(t.StringBounder())
	case NoteOver:
		return t.usedWidth(t.StringBounder()) / 2
	}
	return 0
}

// stableMaxX returns the right edges of the note as Reals safe to
// constrain (no cached maximum).
func (t *NoteTile) stableMaxX() []Real {
	result := []Real{t.x(t.StringBounder()).AddFixed(t.usedWidth(t.StringBounder()))}
	if t.note.Position() == NoteOverSeveral {
		result = append(result, t.livingSpace2.PosD(t.StringBounder()))
	}
	return result
}

// stableMinX returns the left edges of the note as Reals safe to
// constrain.
func (t *NoteTile) stableMinX() []Real {
	result := []Real{t.x(t.StringBounder())}
	if t.note.Position() == NoteOverSeveral {
		result = append(result, t.livingSpace1.PosB(t.StringBounder()))
	}
	return result
}

// MinX implements Tile.
func (t *NoteTile) MinX() Real {
	result := t.x(t.StringBounder())
	if t.note.Position() == NoteOverSeveral {
		x1 := t.livingSpace1.PosB(t.StringBounder())
		return RealUtilsMin(result, x1)
	}
	return result
}

// MaxX implements Tile.
func (t *NoteTile) MaxX() Real {
	result := t.x(t.StringBounder()).AddFixed(t.usedWidth(t.StringBounder()))
	if t.note.Position() == NoteOverSeveral {
		x2 := t.livingSpace2.PosD(t.StringBounder())
		return RealUtilsMax(result, x2)
	}
	return result
}

// NotesTile is a group of notes side by side ("/ note").
type NotesTile struct {
	abstractTile
	livingSpaces *LivingSpaces
	skin         *rose.Rose
	skinParam    klimt.ISkinParam
	notes        *Notes
	yGauge       *YGauge
}

// Event implements Tile.
func (t *NotesTile) Event() Event { return t.notes }

// ContactPointRelative implements Tile: the vertical center of the notes.
func (t *NotesTile) ContactPointRelative() float64 { return t.PreferredHeight() / 2 }

func newNotesTile(stringBounder font.StringBounder, livingSpaces *LivingSpaces, notes *Notes, skin *rose.Rose,
	skinParam klimt.ISkinParam, currentY *YGauge) *NotesTile {
	t := &NotesTile{livingSpaces: livingSpaces, notes: notes, skin: skin, skinParam: skinParam}
	t.initTile(t, stringBounder)
	// The tile publishes a contact line, so that a following parallel (&)
	// message aligns on its center; a group of notes is never parallel
	t.yGauge = YGaugeCreateWithContact(currentY, t.ContactPointRelative(), t.PreferredHeight())
	return t
}

// YGauge implements Tile.
func (t *NotesTile) YGauge() *YGauge { return t.yGauge }

func (t *NotesTile) component(note *Note) klimt.Component {
	return t.skin.CreateComponentNoteAt(note.UsedStyles(), noteComponentType(note.NoteStyle()),
		note.SkinParamBackcolored(t.skinParam), note.Display(), note.Colors(), note.Position())
}

// DrawU draws the notes.
func (t *NotesTile) DrawU(ug klimt.UGraphic) {
	// Self-translate prologue: absolute gauge position
	ug = ug.Apply(klimt.UTranslateDy(t.YGauge().Min().CurrentValue()))
	stringBounder := ug.StringBounder()
	for _, note := range t.notes.List() {
		comp := t.component(note)
		dim := comp.PreferredDimension(stringBounder)
		x := t.x(stringBounder, note).CurrentValue()
		area := klimt.AreaCreate(t.usedWidth(stringBounder, note), dim.Height())
		ug2 := ug.Apply(klimt.UTranslateDx(x))
		comp.DrawU(ug2, area, context2D(ug2))
	}
}

func (t *NotesTile) usedWidth(stringBounder font.StringBounder, note *Note) float64 {
	return t.component(note).PreferredDimension(stringBounder).Width()
}

func (t *NotesTile) xcenter(stringBounder font.StringBounder, note *Note) Real {
	return t.livingSpaces.Get(note.Participant()).PosC(stringBounder)
}

func (t *NotesTile) x(stringBounder font.StringBounder, note *Note) Real {
	livingSpace1 := t.livingSpaces.Get(note.Participant())
	position := note.Position()
	width := t.usedWidth(stringBounder, note)
	switch position {
	case NoteLeft:
		// Several LEFT notes on the same participant are stacked leftward
		return livingSpace1.PosC(stringBounder).AddFixed(-width - t.stackingOffset(stringBounder, note))
	case NoteRight:
		level := livingSpace1.LevelAt(t, IgnoreFutureDeactivate)
		dx := float64(level * LiveDeltaSize)
		return livingSpace1.PosC(stringBounder).AddFixed(dx + t.stackingOffset(stringBounder, note))
	case NoteOverSeveral:
		livingSpace2 := t.livingSpaces.Get(note.Participant2())
		x1 := livingSpace1.PosC(stringBounder)
		x2 := livingSpace2.PosC(stringBounder)
		return RealUtilsMiddle(x1, x2).AddFixed(-width / 2)
	case NoteOver:
		return livingSpace1.PosC(stringBounder).AddFixed(-width / 2)
	}
	uml.Throw(uml.UnsupportedOperationException, position.String())
	return nil
}

// stackingOffset returns the width of the previous notes of the group on
// the same participant and side.
func (t *NotesTile) stackingOffset(stringBounder font.StringBounder, note *Note) float64 {
	result := 0.0
	for _, other := range t.notes.List() {
		if other == note {
			return result
		}
		if other.Position() == note.Position() && other.Participant() == note.Participant() {
			result += t.usedWidth(stringBounder, other)
		}
	}
	return result
}

// PreferredHeight implements Tile: the height of the tallest note.
func (t *NotesTile) PreferredHeight() float64 {
	result := 0.0
	for _, note := range t.notes.List() {
		dim := t.component(note).PreferredDimension(t.StringBounder())
		result = jmath.Max(result, dim.Height())
	}
	return result
}

// AddConstraints keeps the notes of different participants apart.
func (t *NotesTile) AddConstraints() {
	all := t.notes.List()
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			center1 := t.xcenter(t.StringBounder(), all[i]).CurrentValue()
			center2 := t.xcenter(t.StringBounder(), all[j]).CurrentValue()
			// Notes anchored to the same point are rigid: a constraint between
			// them would be unsatisfiable
			if center2 == center1 {
				continue
			}
			if center2 > center1 {
				point1b := t.x2(t.StringBounder(), all[i])
				point2 := t.x(t.StringBounder(), all[j])
				point2.EnsureBiggerThan(point1b)
			} else {
				point1 := t.x(t.StringBounder(), all[i])
				point2b := t.x2(t.StringBounder(), all[j])
				point1.EnsureBiggerThan(point2b)
			}
		}
	}
}

// MinX implements Tile.
func (t *NotesTile) MinX() Real {
	var reals []Real
	for _, note := range t.notes.List() {
		reals = append(reals, t.x(t.StringBounder(), note))
	}
	return RealUtilsMin(reals...)
}

func (t *NotesTile) x2(stringBounder font.StringBounder, note *Note) Real {
	return t.x(stringBounder, note).AddFixed(t.usedWidth(stringBounder, note))
}

// MaxX implements Tile.
func (t *NotesTile) MaxX() Real {
	var reals []Real
	for _, note := range t.notes.List() {
		reals = append(reals, t.x2(t.StringBounder(), note))
	}
	return RealUtilsMax(reals...)
}
