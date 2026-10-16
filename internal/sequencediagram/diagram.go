// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/SequenceDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// SequenceDiagram is a parsed sequence diagram.
type SequenceDiagram struct {
	uml.Base

	hideUnlinkedData bool
	participantsList []*Participant
	events           []Event
	titles           []uml.DisplayPositioned
	englobers        map[*Participant]*ParticipantEnglober

	cpt int

	lastEventWithDeactivate EventWithDeactivate
	lastDelay               *Delay
	pendingCreate           *LifeEvent
	activationState         []AbstractMessage
	openGroupings           []*GroupingStart

	ignoreNewpage bool
	countNewpage  int

	autoNumber   AutoNumber
	showFootbox  bool
	englober     *ParticipantEnglober
	autoactivate bool

	hiddenPortions uml.EntityPortions
	linkAnchors    []*LinkAnchor
}

// NewSequenceDiagram returns an empty sequence diagram for src; md5map
// holds the base64 images of the source.
func NewSequenceDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *SequenceDiagram {
	return &SequenceDiagram{
		Base:        uml.NewBase(ctx, src, core.Sequence, md5map),
		englobers:   map[*Participant]*ParticipantEnglober{},
		cpt:         1,
		showFootbox: true,
	}
}

var _ core.TypedDiagram = (*SequenceDiagram)(nil)

// Description implements core.Diagram: "(N participants)".
func (d *SequenceDiagram) Description() string {
	return "(" + strconv.Itoa(len(d.participantsList)) + " participants)"
}

// IsHideUnlinkedData reports whether "hide unlinked" is set.
func (d *SequenceDiagram) IsHideUnlinkedData() bool { return d.hideUnlinkedData }

// SetHideUnlinkedData sets "hide unlinked".
func (d *SequenceDiagram) SetHideUnlinkedData(hide bool) { d.hideUnlinkedData = hide }

// TitleOfPage returns the title of page index: the diagram title for the
// first page, the "newpage" title for the others.
func (d *SequenceDiagram) TitleOfPage(index int) uml.DisplayPositioned {
	if index > 0 {
		return d.titles[index-1]
	}
	return d.Title()
}

// PageTitles returns the "newpage" titles, one for each page after the
// first.
func (d *SequenceDiagram) PageTitles() []uml.DisplayPositioned { return d.titles }

// Counter returns the next participant uid number.
func (d *SequenceDiagram) Counter() int { return d.cpt }

// OrCreateParticipant returns the participant code, creating it with the
// text of its code when it does not exist.
func (d *SequenceDiagram) OrCreateParticipant(location *text.LineLocation, code string) *Participant {
	return d.OrCreateParticipantDisplay(location, code, klimt.DisplayWithNewlines(d.Pragma(), code))
}

// OrCreateParticipantDisplay returns the participant code, creating it with
// display when it does not exist.
func (d *SequenceDiagram) OrCreateParticipantDisplay(location *text.LineLocation, code string,
	display *klimt.Display) *Participant {
	result := d.participant(code)
	if result == nil {
		result = newParticipant(location, TypeParticipant, code, display, &d.hiddenPortions, 0,
			d.CurrentStyleBuilder(), d.createUID())
		d.addWithOrder(result)
		d.englobers[result] = d.englober
	}
	return result
}

func (d *SequenceDiagram) participant(code string) *Participant {
	for _, p := range d.participantsList {
		if p.Code() == code {
			return p
		}
	}
	return nil
}

// LastEventWithDeactivate returns the last event activations can attach
// to, nil for none.
func (d *SequenceDiagram) LastEventWithDeactivate() EventWithDeactivate {
	for i := len(d.events) - 1; i >= 0; i-- {
		if e, ok := d.events[i].(EventWithDeactivate); ok {
			return e
		}
	}
	return nil
}

// LastEventWithNote returns the last event notes can attach to, nil for
// none.
func (d *SequenceDiagram) LastEventWithNote() EventWithNote {
	for i := len(d.events) - 1; i >= 0; i-- {
		if e, ok := d.events[i].(EventWithNote); ok {
			return e
		}
	}
	return nil
}

// CreateNewParticipant declares a participant; it panics when code exists.
// A null display gives the text of the code.
func (d *SequenceDiagram) CreateNewParticipant(location *text.LineLocation, typ ParticipantType, code string,
	display *klimt.Display, order int) *Participant {
	if d.participant(code) != nil {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if klimt.DisplayIsNull(display) {
		display = klimt.DisplayWithNewlines(d.Pragma(), code)
	}
	result := newParticipant(location, typ, code, display, &d.hiddenPortions, order, d.CurrentStyleBuilder(),
		d.createUID())
	d.addWithOrder(result)
	d.englobers[result] = d.englober
	return result
}

func (d *SequenceDiagram) createUID() string {
	uid := "part" + strconv.Itoa(d.cpt)
	d.cpt++
	return uid
}

func (d *SequenceDiagram) addWithOrder(result *Participant) {
	for i, p := range d.participantsList {
		if result.Order() < p.Order() {
			d.participantsList = append(d.participantsList[:i], append([]*Participant{result}, d.participantsList[i:]...)...)
			return
		}
	}
	d.participantsList = append(d.participantsList, result)
}

// Participants returns the participants in display order.
func (d *SequenceDiagram) Participants() []*Participant { return d.participantsList }

// ParticipantsContainsKey reports whether participant code exists.
func (d *SequenceDiagram) ParticipantsContainsKey(code string) bool {
	return d.participant(code) != nil
}

// AddMessage adds a message.
func (d *SequenceDiagram) AddMessage(m AbstractMessage) command.Result {
	if m.IsParallel() {
		if last := d.lastAbstractMessage(); last != nil {
			m.Base().SetParallelBrother(last)
		} else {
			m.Base().SetParallelBrother(nil)
		}
	}
	d.lastEventWithDeactivate = m
	d.lastDelay = nil
	d.events = append(d.events, m)
	if d.pendingCreate != nil {
		if !m.CompatibleForCreate(d.pendingCreate.Participant()) {
			return command.Error("After create command, you have to send a message to \"" +
				d.pendingCreate.Participant().String() + "\"")
		}
		m.Base().AddLifeEvent(d.pendingCreate)
		d.pendingCreate = nil
	}
	return command.Ok()
}

func (d *SequenceDiagram) lastAbstractMessage() AbstractMessage {
	for i := len(d.events) - 1; i >= 0; i-- {
		if m, ok := d.events[i].(AbstractMessage); ok {
			return m
		}
	}
	return nil
}

// AddNote adds a note; with tryMerge ("/ note"), a note following another
// note is put beside it.
func (d *SequenceDiagram) AddNote(n *Note, tryMerge bool) {
	if tryMerge && len(d.events) > 0 {
		switch last := d.events[len(d.events)-1].(type) {
		case *Note:
			d.events[len(d.events)-1] = &Notes{notes: []*Note{last, n}}
			return
		case *Notes:
			last.Add(n)
			return
		}
	}
	d.events = append(d.events, n)
}

// Newpage adds a page break with the title of the next page, unless
// "ignore newpage" is set.
func (d *SequenceDiagram) Newpage(title uml.DisplayPositioned) {
	if d.ignoreNewpage {
		return
	}
	d.titles = append(d.titles, title)
	d.events = append(d.events, newNewpage(d.CurrentStyleBuilder()))
	d.countNewpage++
}

// IgnoreNewpage sets "ignore newpage".
func (d *SequenceDiagram) IgnoreNewpage() { d.ignoreNewpage = true }

// SetAutonewpage is "autonewpage n", which upstream ignores.
func (d *SequenceDiagram) SetAutonewpage(int) {}

// Divider adds a divider.
func (d *SequenceDiagram) Divider(strings *klimt.Display) {
	divider := newDivider(strings, d.CurrentStyleBuilder())
	d.events = append(d.events, divider)
	d.lastEventWithDeactivate = divider
}

// HSpace adds a vertical space of pixel pixels.
func (d *SequenceDiagram) HSpace(pixel int) { d.events = append(d.events, &HSpace{pixel: pixel}) }

// Delay adds a delay.
func (d *SequenceDiagram) Delay(strings *klimt.Display) {
	delay := newDelay(strings, d.CurrentStyleBuilder())
	d.events = append(d.events, delay)
	d.lastDelay = delay
}

// Events returns the events in order.
func (d *SequenceDiagram) Events() []Event { return d.events }

// ActivatingMessage returns the message of the last activation not yet
// deactivated, for "return" and "deactivate"; nil for none.
func (d *SequenceDiagram) ActivatingMessage() AbstractMessage {
	if len(d.activationState) == 0 {
		return nil
	}
	return d.activationState[len(d.activationState)-1]
}

// Activate adds a lifeline event. It returns an error message, or "".
func (d *SequenceDiagram) Activate(p *Participant, typ LifeEventType, backcolor, linecolor color.HColor) string {
	if d.lastDelay != nil {
		return "You cannot Activate/Deactivate just after a ..."
	}
	lifeEvent := NewLifeEvent(p, typ, klimt.NewFashion(backcolor, linecolor), d.CurrentStyleBuilder())
	d.events = append(d.events, lifeEvent)
	if typ == LifeEventCreate {
		d.pendingCreate = lifeEvent
		return ""
	}
	if d.lastEventWithDeactivate == nil {
		if typ == LifeEventActivate {
			p.IncInitialLife(klimt.NewFashion(backcolor, linecolor))
			return ""
		}
		if p.InitialLife() == 0 {
			return "You cannot deactivate here"
		}
		return ""
	}
	if !d.lastEventWithDeactivate.DealWith(p) {
		return ""
	}
	lastMessage, isMessage := d.lastEventWithDeactivate.(AbstractMessage)
	if typ == LifeEventActivate && isMessage {
		d.activationState = append(d.activationState, lastMessage)
	} else if typ == LifeEventDeactivate && len(d.activationState) > 0 {
		d.activationState = d.activationState[:len(d.activationState)-1]
	}
	ok := d.lastEventWithDeactivate.AddLifeEvent(lifeEvent)
	if isMessage {
		lifeEvent.SetMessage(lastMessage)
	}
	if ok {
		return ""
	}
	return "Activate/Deactivate already done on " + p.Code()
}

// Grouping adds a fragment line. It returns false for an "else" or "end"
// outside any fragment.
func (d *SequenceDiagram) Grouping(title, comment string, hasComment bool, typ GroupingType, backColorGeneral,
	backColorElement color.HColor, parallel bool) bool {
	if !typ.IsStart() && len(d.openGroupings) == 0 {
		return false
	}
	if backColorGeneral == nil {
		backColorGeneral = d.Skin().HtmlColor(klimt.ColorParamSequenceGroupBodyBackground, nil, false)
	}
	var top *GroupingStart
	if len(d.openGroupings) > 0 {
		top = d.openGroupings[0]
	}
	var g Grouping
	switch typ {
	case GroupingStartPartition, GroupingStartType:
		g = newGroupingStart(typ, title, comment, hasComment, backColorGeneral, backColorElement, top,
			d.CurrentStyleBuilder())
	default:
		g = newGroupingLeaf(title, comment, hasComment, typ, backColorGeneral, backColorElement, top,
			d.CurrentStyleBuilder())
	}
	d.events = append(d.events, g)
	if typ.IsStart() {
		start := g.(*GroupingStart)
		if parallel {
			start.GoParallel()
		}
		d.openGroupings = append([]*GroupingStart{start}, d.openGroupings...)
	} else if typ == GroupingEnd {
		d.openGroupings = d.openGroupings[1:]
		d.lastEventWithDeactivate = g.(*GroupingLeaf)
	}
	return true
}

// AutonumberGo starts numbering.
func (d *SequenceDiagram) AutonumberGo(startingNumber *DottedNumber, increment int, format *decimalFormat) {
	d.autoNumber.Go(startingNumber, increment, format)
}

// AutonumberStop suspends numbering.
func (d *SequenceDiagram) AutonumberStop() { d.autoNumber.Stop() }

// AutoNumber returns the numbering state.
func (d *SequenceDiagram) AutoNumber() *AutoNumber { return &d.autoNumber }

// NextMessageNumber returns the number of the next message; ok is false
// when numbering is off.
func (d *SequenceDiagram) NextMessageNumber() (string, bool) { return d.autoNumber.NextMessageNumber() }

// IsShowFootbox reports whether the participants are repeated at the
// bottom.
func (d *SequenceDiagram) IsShowFootbox() bool {
	if d.Skin().StrictUmlStyle() {
		return false
	}
	footbox, ok := d.Skin().Value("footbox")
	if !ok {
		return d.showFootbox
	}
	return !uml.EqualsIgnoreCase(footbox, "hide")
}

// SetShowFootbox sets "hide footbox" or "show footbox".
func (d *SequenceDiagram) SetShowFootbox(footbox bool) { d.showFootbox = footbox }

// BoxStart opens a box, inside the current one if any.
func (d *SequenceDiagram) BoxStart(comment *klimt.Display, boxColor color.HColor, stereotype *klimt.Stereotype) {
	d.englober = &ParticipantEnglober{parent: d.englober, title: comment, boxColor: boxColor, stereotype: stereotype}
}

// EndBox closes the current box.
func (d *SequenceDiagram) EndBox() {
	if d.englober == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	d.englober = d.englober.parent
}

// IsBoxPending reports whether a box is open.
func (d *SequenceDiagram) IsBoxPending() bool { return d.englober != nil }

// Pages returns the number of images: one more than the page breaks.
func (d *SequenceDiagram) Pages() int { return d.countNewpage + 1 }

// RemoveHiddenParticipants removes the participants that no event involves
// ("hide unlinked").
func (d *SequenceDiagram) RemoveHiddenParticipants() {
	for _, p := range append([]*Participant(nil), d.participantsList...) {
		if d.isAlone(p) {
			d.remove(p)
		}
	}
}

func (d *SequenceDiagram) remove(p *Participant) {
	for i, q := range d.participantsList {
		if q == p {
			d.participantsList = append(d.participantsList[:i], d.participantsList[i+1:]...)
			delete(d.englobers, p)
			return
		}
	}
	uml.Throw(uml.IllegalArgumentException, "")
}

func (d *SequenceDiagram) isAlone(p *Participant) bool {
	for _, ev := range d.events {
		if ev.DealWith(p) {
			return false
		}
	}
	return true
}

// PutParticipantInLast moves an existing participant after the others of
// the same order, in the current box (a participant declared again).
func (d *SequenceDiagram) PutParticipantInLast(code string) {
	p := d.participant(code)
	if p == nil {
		uml.ThrowNPE(code)
	}
	for i, q := range d.participantsList {
		if q == p {
			d.participantsList = append(d.participantsList[:i], d.participantsList[i+1:]...)
			break
		}
	}
	d.addWithOrder(p)
	d.englobers[p] = d.englober
}

// Englober returns the box of a participant, nil for none.
func (d *SequenceDiagram) Englober(p *Participant) *ParticipantEnglober { return d.englobers[p] }

// SetAutoactivate sets "autoactivate on|off".
func (d *SequenceDiagram) SetAutoactivate(autoactivate bool) { d.autoactivate = autoactivate }

// IsAutoactivate reports whether "autoactivate on" is set.
func (d *SequenceDiagram) IsAutoactivate() bool { return d.autoactivate }

// HasURL reports whether the diagram has a link.
func (d *SequenceDiagram) HasURL() bool {
	for _, p := range d.participantsList {
		if p.URL() != nil {
			return true
		}
	}
	for _, ev := range d.events {
		if ev.HasURL() {
			return true
		}
	}
	return !d.Legend().IsNull() && d.Legend().HasURL()
}

// AddReference adds a "ref over" frame.
func (d *SequenceDiagram) AddReference(ref *Reference) { d.events = append(d.events, ref) }

// IsIncomplete implements command.Diagram: a diagram without participant
// is not a sequence diagram, so the other parsers get a chance.
func (d *SequenceDiagram) IsIncomplete() bool { return len(d.participantsList) == 0 }

// CheckFinalError implements command.Diagram; "hide unlinked" removes the
// unused participants here.
func (d *SequenceDiagram) CheckFinalError() string {
	if d.hideUnlinkedData {
		d.RemoveHiddenParticipants()
	}
	return ""
}

// HiddenPortions returns the portions hidden by "hide stereotype" and the
// like.
func (d *SequenceDiagram) HiddenPortions() uml.EntityPortions { return d.hiddenPortions }

// HideOrShow hides or shows portions.
func (d *SequenceDiagram) HideOrShow(portions uml.EntityPortions, show bool) {
	if show {
		d.hiddenPortions = d.hiddenPortions.Without(portions)
	} else {
		d.hiddenPortions = d.hiddenPortions.With(portions)
	}
}

// HideShowByGender implements uml.GenderHider: "hide stereotype" and the
// like hide portions of every participant.
func (d *SequenceDiagram) HideShowByGender(arg *regex.Result) command.Result {
	portion := uml.GenderPortion(arg.Str("PORTION", 0))
	d.HideOrShow(portion.AsSet(), uml.EqualsIgnoreCase(arg.Str("COMMAND", 0), "show"))
	return command.Ok()
}

// ManageVariable replaces "%autonumber%" by the last message number.
func (d *SequenceDiagram) ManageVariable(labels *klimt.Display) *klimt.Display {
	return labels.Replace("%autonumber%", d.autoNumber.CurrentMessageNumber(false))
}

// LinkAnchor adds a link between two anchors.
func (d *SequenceDiagram) LinkAnchor(anchor1, anchor2, message string, hasMessage bool) command.Result {
	d.linkAnchors = append(d.linkAnchors, &LinkAnchor{Anchor1: anchor1, Anchor2: anchor2, Message: message,
		HasMessage: hasMessage})
	return command.Ok()
}

// LinkAnchors returns the links between anchors.
func (d *SequenceDiagram) LinkAnchors() []*LinkAnchor { return d.linkAnchors }

// DefaultMargins are the margins of sequence diagrams: 5 on each side for
// Teoz, the only engine.
func (d *SequenceDiagram) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return klimt.ClockwiseSame(5)
}
