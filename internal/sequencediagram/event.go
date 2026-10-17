// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Event.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/AbstractEvent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/EventWithDeactivate.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/EventWithNote.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LifeEvent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LifeEventType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Divider.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Delay.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/HSpace.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Newpage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Reference.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LinkAnchor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Event is an element of the vertical flow of a sequence diagram: a
// message, a note, a fragment, a divider and so on.
type Event interface {
	// DealWith reports whether the event involves the participant.
	DealWith(p *Participant) bool
	// URL returns the link of the event, nil for none.
	URL() *klimt.URL
	HasURL() bool
	IsParallel() bool
}

// EventWithDeactivate is an event that activations can attach to.
type EventWithDeactivate interface {
	Event
	SetPosYendLevel(y float64)
	PosYendLevel() float64
	AddLifeEvent(lifeEvent *LifeEvent) bool
}

// EventWithNote is an event that notes can attach to ("note left" after
// a message).
type EventWithNote interface {
	Event
	// AddNote attaches a note; the warning, when not nil, tells that the
	// position is ignored.
	AddNote(note *Note) *klimt.Warning
}

// abstractEvent provides upstream's AbstractEvent defaults.
type abstractEvent struct{}

func (abstractEvent) URL() *klimt.URL  { return nil }
func (abstractEvent) HasURL() bool     { return false }
func (abstractEvent) IsParallel() bool { return false }

// posYend holds the end level of upstream's EventWithDeactivate events.
type posYend struct{ posYendLevel float64 }

func (p *posYend) SetPosYendLevel(y float64) { p.posYendLevel = y }
func (p *posYend) PosYendLevel() float64     { return p.posYendLevel }

// LifeEventType is the kind of a lifeline event.
type LifeEventType int

// Lifeline event types, in upstream's order.
const (
	LifeEventActivate LifeEventType = iota
	LifeEventDeactivate
	LifeEventDestroy
	LifeEventCreate
)

var lifeEventTypeNames = [...]string{"ACTIVATE", "DEACTIVATE", "DESTROY", "CREATE"}

// String returns upstream's name, for example "ACTIVATE".
func (t LifeEventType) String() string { return lifeEventTypeNames[t] }

// lifeEventTypeOf is LifeEventType.valueOf of an upper-cased keyword.
func lifeEventTypeOf(s string) LifeEventType {
	for i, n := range lifeEventTypeNames {
		if n == s {
			return LifeEventType(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.sequencediagram.LifeEventType."+s)
	return 0
}

// StyleSignature returns the style signature of a destroy event; the other
// types have none and panic, as upstream.
func (t LifeEventType) StyleSignature() klimt.StyleSignatureBasic {
	if t == LifeEventDestroy {
		return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
			klimt.SNameLifeLine, klimt.SNameDestroy)
	}
	panic(&color.RuntimeError{Class: uml.UnsupportedOperationException})
}

// LifeEvent is an activation, deactivation, creation or destruction of a
// lifeline.
type LifeEvent struct {
	abstractEvent
	p            *Participant
	typ          LifeEventType
	backcolor    *klimt.Fashion
	styleBuilder *klimt.StyleBuilder
	message      AbstractMessage
}

// NewLifeEvent returns a lifeline event.
func NewLifeEvent(p *Participant, typ LifeEventType, backcolor *klimt.Fashion,
	styleBuilder *klimt.StyleBuilder) *LifeEvent {
	return &LifeEvent{p: p, typ: typ, backcolor: backcolor, styleBuilder: styleBuilder}
}

// String is upstream's toString.
func (e *LifeEvent) String() string { return "LifeEvent:" + e.p.String() + " " + e.typ.String() }

// Participant returns the participant of the event.
func (e *LifeEvent) Participant() *Participant { return e.p }

// Type returns the kind of the event.
func (e *LifeEvent) Type() LifeEventType { return e.typ }

// SpecificColors returns the colors of an activation.
func (e *LifeEvent) SpecificColors() *klimt.Fashion { return e.backcolor }

// DealWith implements Event.
func (e *LifeEvent) DealWith(someone *Participant) bool { return e.p == someone }

// IsActivate reports whether the event is an activation.
func (e *LifeEvent) IsActivate() bool { return e.typ == LifeEventActivate }

// IsDeactivateOrDestroy reports whether the event ends an activation.
func (e *LifeEvent) IsDeactivateOrDestroy() bool {
	return e.typ == LifeEventDeactivate || e.typ == LifeEventDestroy
}

// IsDeactivate reports whether the event is a deactivation.
func (e *LifeEvent) IsDeactivate() bool { return e.typ == LifeEventDeactivate }

// IsDestroy reports whether the event destroys p.
func (e *LifeEvent) IsDestroy(p *Participant) bool { return e.p == p && e.typ == LifeEventDestroy }

// SetMessage attaches the event to a message.
func (e *LifeEvent) SetMessage(message AbstractMessage) { e.message = message }

// Message returns the message the event is attached to, nil for none.
func (e *LifeEvent) Message() AbstractMessage { return e.message }

// StyleBuilder returns the style sheet at the time of the event.
func (e *LifeEvent) StyleBuilder() *klimt.StyleBuilder { return e.styleBuilder }

// UsedStyle returns the style of the event (destroy events only).
func (e *LifeEvent) UsedStyle() []*klimt.Style {
	return []*klimt.Style{e.typ.StyleSignature().MergedStyle(e.styleBuilder)}
}

// Divider is a "== text ==" separator.
type Divider struct {
	abstractEvent
	posYend
	text  *klimt.Display
	style *klimt.Style
}

func newDivider(text *klimt.Display, styleBuilder *klimt.StyleBuilder) *Divider {
	return &Divider{text: text, style: klimt.ComponentTypeDivider.StyleSignature().MergedStyle(styleBuilder)}
}

// Text returns the label of the divider.
func (d *Divider) Text() *klimt.Display { return d.text }

// UsedStyles returns the style of the divider.
func (d *Divider) UsedStyles() []*klimt.Style { return []*klimt.Style{d.style} }

// DealWith implements Event.
func (*Divider) DealWith(*Participant) bool { return false }

// AddLifeEvent implements EventWithDeactivate: a divider closes the
// message before it, so lifeline events after it stand alone.
func (*Divider) AddLifeEvent(*LifeEvent) bool { return true }

// Delay is a "..." or "... text ..." delay.
type Delay struct {
	abstractEvent
	text  *klimt.Display
	style *klimt.Style
}

func newDelay(text *klimt.Display, styleBuilder *klimt.StyleBuilder) *Delay {
	return &Delay{text: text, style: klimt.ComponentTypeDelayText.StyleSignature().MergedStyle(styleBuilder)}
}

// Text returns the label of the delay.
func (d *Delay) Text() *klimt.Display { return d.text }

// UsedStyles returns the style of the delay.
func (d *Delay) UsedStyles() []*klimt.Style { return []*klimt.Style{d.style} }

// DealWith implements Event.
func (*Delay) DealWith(*Participant) bool { return false }

// HSpace is a "|||" or "||n||" vertical space.
type HSpace struct {
	abstractEvent
	pixel int
}

// Pixel returns the height of the space.
func (h *HSpace) Pixel() int { return h.pixel }

// DealWith implements Event.
func (*HSpace) DealWith(*Participant) bool { return false }

// Newpage is a page break.
type Newpage struct {
	abstractEvent
	style *klimt.Style
}

func newNewpage(styleBuilder *klimt.StyleBuilder) *Newpage {
	return &Newpage{style: klimt.ComponentTypeNewpage.StyleSignature().MergedStyle(styleBuilder)}
}

// UsedStyles returns the style of the page break.
func (n *Newpage) UsedStyles() []*klimt.Style { return []*klimt.Style{n.style} }

// DealWith implements Event.
func (*Newpage) DealWith(*Participant) bool { return false }

// Reference is a "ref over" frame.
type Reference struct {
	abstractEvent
	participants     []*Participant
	url              *klimt.URL
	backColorGeneral color.HColor
	backColorElement color.HColor
	strings          *klimt.Display
	style            *klimt.Style
	styleHeader      *klimt.Style
	noteOnMessages   []*Note
}

func newReference(participants []*Participant, url *klimt.URL, strings *klimt.Display,
	backColorGeneral, backColorElement color.HColor, styleBuilder *klimt.StyleBuilder) *Reference {
	r := &Reference{url: url, strings: strings, backColorGeneral: backColorGeneral,
		backColorElement: backColorElement}
	for _, p := range participants {
		if !slices.Contains(r.participants, p) {
			r.participants = append(r.participants, p)
		}
	}
	seq := func(names ...klimt.SName) klimt.StyleSignatureBasic {
		return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
			klimt.SNameSequenceDiagram}, names...)...)
	}
	r.style = seq(klimt.SNameReference).MergedStyle(styleBuilder)
	flat := seq(klimt.SNameReferenceHeader).MergedStyle(styleBuilder)
	nested := seq(klimt.SNameReference, klimt.SNameHeader).MergedStyle(styleBuilder)
	if flat == nil {
		r.styleHeader = nested
	} else {
		r.styleHeader = flat.MergeNestedChildOver(nested, r.style, klimt.MergeOverwriteExistingValue)
	}
	return r
}

// Participants returns the participants the frame spans, without
// duplicates.
func (r *Reference) Participants() []*Participant { return r.participants }

// Strings returns the text of the frame.
func (r *Reference) Strings() *klimt.Display { return r.strings }

// DealWith implements Event.
func (r *Reference) DealWith(someone *Participant) bool {
	return slices.Contains(r.participants, someone)
}

// URL implements Event.
func (r *Reference) URL() *klimt.URL { return r.url }

// HasURL implements Event.
func (r *Reference) HasURL() bool { return r.url != nil }

// String is upstream's toString: the participant codes joined by "-".
func (r *Reference) String() string {
	codes := make([]string, len(r.participants))
	for i, p := range r.participants {
		codes[i] = p.Code()
	}
	return strings.Join(codes, "-")
}

// BackColorGeneral returns the background color, nil for none.
func (r *Reference) BackColorGeneral() color.HColor { return r.backColorGeneral }

// BackColorElement returns the header color, nil for none.
func (r *Reference) BackColorElement() color.HColor { return r.backColorElement }

// UsedStyles returns the styles of the frame and of its header.
func (r *Reference) UsedStyles() []*klimt.Style {
	header := r.styleHeader
	if header != nil {
		header = header.EventuallyOverrideColor(klimt.PNameBackGroundColor, r.backColorElement)
	}
	return []*klimt.Style{r.style, header}
}

// AddNote implements EventWithNote.
func (r *Reference) AddNote(note *Note) *klimt.Warning {
	if note.Position() != NoteLeft && note.Position() != NoteRight {
		return klimt.NewWarning("This position is ignored: " + note.Position().String())
	}
	r.noteOnMessages = append(r.noteOnMessages, note)
	return nil
}

// NoteOnMessages returns the notes attached to the frame.
func (r *Reference) NoteOnMessages() []*Note { return r.noteOnMessages }

// LinkAnchor is a "{a} <-> {b} : text" link between two anchored messages.
type LinkAnchor struct {
	Anchor1, Anchor2 string
	// Message is the label; HasMessage is false when there is none.
	Message    string
	HasMessage bool
}

// String is upstream's toString.
func (l *LinkAnchor) String() string {
	msg := l.Message
	if !l.HasMessage {
		msg = "null"
	}
	return l.Anchor1 + "<->" + l.Anchor2 + " " + msg
}
