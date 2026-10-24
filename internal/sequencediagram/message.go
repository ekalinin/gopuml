// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/AbstractMessage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/Message.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/MessageExo.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/MessageExoType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/MessageNumber.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// AbstractMessage is a message between participants (Message) or from or
// to the border of the diagram (MessageExo).
type AbstractMessage interface {
	EventWithDeactivate
	EventWithNote
	// Base returns the shared message state.
	Base() *MessageBase
	Participant1() *Participant
	Participant2() *Participant
	IsSelfMessage() bool
	// CompatibleForCreate reports whether the message can create p (the
	// message after "create p" must go to p).
	CompatibleForCreate(p *Participant) bool
	overrideNotePosition(position NotePosition) NotePosition
}

// MessageBase is upstream's AbstractMessage: the state that Message and
// MessageExo share.
type MessageBase struct {
	posYend
	self AbstractMessage

	stereotype         *klimt.Stereotype
	location           *text.LineLocation
	label              *klimt.Display
	arrowConfiguration *klimt.ArrowConfiguration
	lifeEventsType     [4]bool
	url                *klimt.URL
	messageNumber      string
	hasMessageNumber   bool
	parallel           bool
	parallelBrother    AbstractMessage
	styleBuilder       *klimt.StyleBuilder
	noteOnMessages     []*Note

	firstIsActivate        bool
	noActivationAuthorized []*Participant

	posYstartLevel           float64
	anchor, anchor1, anchor2 string
	hasAnchor                [3]bool
}

func newMessageBase(self AbstractMessage, styleBuilder *klimt.StyleBuilder, label *klimt.Display,
	arrowConfiguration *klimt.ArrowConfiguration, messageNumber string, hasNumber bool,
	location *text.LineLocation) MessageBase {
	return MessageBase{self: self, styleBuilder: styleBuilder, location: location, label: label,
		arrowConfiguration: arrowConfiguration, messageNumber: messageNumber, hasMessageNumber: hasNumber}
}

// Base implements AbstractMessage.
func (m *MessageBase) Base() *MessageBase { return m }

// SetStereotype sets the stereotype of the message (upstream's oddly named
// getStereotype(Stereotype)).
func (m *MessageBase) SetStereotype(stereotype *klimt.Stereotype) { m.stereotype = stereotype }

// Stereotype returns the stereotype of the message, nil for none.
func (m *MessageBase) Stereotype() *klimt.Stereotype { return m.stereotype }

// StyleSignature returns the style signature of the message.
func (m *MessageBase) StyleSignature() klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
		klimt.SNameArrow).WithTOBECHANGED(m.stereotype)
}

// UsedStyles returns the style of the message, with its arrow color.
func (m *MessageBase) UsedStyles() []*klimt.Style {
	style := m.StyleSignature().MergedStyle(m.styleBuilder)
	if style != nil && m.arrowConfiguration.Color() != nil {
		style = style.EventuallyOverrideColor(klimt.PNameLineColor, m.arrowConfiguration.Color())
	}
	return []*klimt.Style{style}
}

// SetURL sets the link of the message.
func (m *MessageBase) SetURL(url *klimt.URL) { m.url = url }

// GoParallel marks the message as parallel to the previous one ("&").
func (m *MessageBase) GoParallel() { m.parallel = true }

// SetParallelBrother sets the message this one is parallel to.
func (m *MessageBase) SetParallelBrother(brother AbstractMessage) { m.parallelBrother = brother }

// ParallelBrother returns the message this one is parallel to, nil for
// none.
func (m *MessageBase) ParallelBrother() AbstractMessage { return m.parallelBrother }

// IsParallel implements Event.
func (m *MessageBase) IsParallel() bool { return m.parallel }

// IsParallelWith reports whether the message is message or is parallel to
// it, directly or through other parallel messages.
func (m *MessageBase) IsParallelWith(message AbstractMessage) bool {
	if m.self == message {
		return true
	}
	return m.parallelBrother != nil && (m.parallelBrother == message || m.parallelBrother.Base().IsParallelWith(message))
}

// URL implements Event: the link of the message, else the first link of
// its notes.
func (m *MessageBase) URL() *klimt.URL {
	if m.url == nil {
		for _, n := range m.noteOnMessages {
			if n.URL() != nil {
				return n.URL()
			}
		}
	}
	return m.url
}

// HasURL implements Event.
func (m *MessageBase) HasURL() bool {
	for _, n := range m.noteOnMessages {
		if n.HasURL() {
			return true
		}
	}
	if m.label != nil && m.label.HasURL() {
		return true
	}
	return m.URL() != nil
}

// AddLifeEvent implements EventWithDeactivate. It returns false when the
// participant cannot be activated again by this message.
func (m *MessageBase) AddLifeEvent(lifeEvent *LifeEvent) bool {
	lifeEvent.SetMessage(m.self)
	if m.self.IsSelfMessage() && lifeEvent.Participant() != m.self.Participant1() {
		return true
	}
	m.lifeEventsType[lifeEvent.Type()] = true
	if m.lifeEventsTypeSize() == 1 && m.IsActivate() {
		m.firstIsActivate = true
	}
	if lifeEvent.Type() == LifeEventActivate && containsParticipant(m.noActivationAuthorized, lifeEvent.Participant()) {
		return false
	}
	if lifeEvent.Type() == LifeEventDeactivate || lifeEvent.Type() == LifeEventDestroy {
		if !containsParticipant(m.noActivationAuthorized, lifeEvent.Participant()) {
			m.noActivationAuthorized = append(m.noActivationAuthorized, lifeEvent.Participant())
		}
	}
	return true
}

func containsParticipant(list []*Participant, p *Participant) bool {
	for _, q := range list {
		if q == p {
			return true
		}
	}
	return false
}

func (m *MessageBase) lifeEventsTypeSize() int {
	n := 0
	for _, b := range m.lifeEventsType {
		if b {
			n++
		}
	}
	return n
}

// LifeEventTypes returns the kinds of the lifeline events of the message,
// in enum order.
func (m *MessageBase) LifeEventTypes() []LifeEventType {
	var out []LifeEventType
	for t, b := range m.lifeEventsType {
		if b {
			out = append(out, LifeEventType(t))
		}
	}
	return out
}

// IsCreate reports whether the message creates its target.
func (m *MessageBase) IsCreate() bool { return m.lifeEventsType[LifeEventCreate] }

// IsActivate reports whether the message activates a participant.
func (m *MessageBase) IsActivate() bool { return m.lifeEventsType[LifeEventActivate] }

// IsDeactivate reports whether the message deactivates a participant.
func (m *MessageBase) IsDeactivate() bool { return m.lifeEventsType[LifeEventDeactivate] }

// IsDestroy reports whether the message destroys a participant.
func (m *MessageBase) IsDestroy() bool { return m.lifeEventsType[LifeEventDestroy] }

// IsActivateAndDeactive reports whether the message first activated a
// participant and then deactivated or destroyed one.
func (m *MessageBase) IsActivateAndDeactive() bool {
	return m.firstIsActivate && (m.IsDeactivate() || m.IsDestroy())
}

// Label returns the label of the message.
func (m *MessageBase) Label() *klimt.Display { return m.label }

// LabelNumbered returns the label preceded by the message number, when
// there is one.
func (m *MessageBase) LabelNumbered() *klimt.Display {
	if !m.hasMessageNumber {
		return m.label
	}
	return klimt.DisplayEmpty().Add(MessageNumber(m.messageNumber)).AddAll(m.label)
}

// *klimt.ArrowConfiguration returns the arrow of the message.
func (m *MessageBase) ArrowConfiguration() *klimt.ArrowConfiguration { return m.arrowConfiguration }

// NoteOnMessages returns the notes attached to the message.
func (m *MessageBase) NoteOnMessages() []*Note { return m.noteOnMessages }

// AddNote implements EventWithNote: left, right, top and bottom notes are
// attached; other positions are ignored with a warning.
func (m *MessageBase) AddNote(note *Note) *klimt.Warning {
	switch note.Position() {
	case NoteLeft, NoteRight, NoteBottom, NoteTop:
	default:
		return klimt.NewWarning("This position is ignored: " + note.Position().String())
	}
	note = note.WithPosition(m.self.overrideNotePosition(note.Position()))
	m.noteOnMessages = append(m.noteOnMessages, note)
	return nil
}

// MessageNumber returns the autonumber of the message; ok is false when
// there is none.
func (m *MessageBase) MessageNumber() (string, bool) { return m.messageNumber, m.hasMessageNumber }

// PosYstartLevel returns the start level of the message.
func (m *MessageBase) PosYstartLevel() float64 { return m.posYstartLevel }

// SetPosYstartLevel sets the start level of the message.
func (m *MessageBase) SetPosYstartLevel(y float64) { m.posYstartLevel = y }

// SetAnchor sets the "{anchor}" of the message; ok false sets none.
func (m *MessageBase) SetAnchor(anchor string, ok bool) {
	m.anchor, m.hasAnchor[0] = anchor, ok
	if ok && len(anchor) > 0 && anchor[0] == '{' {
		uml.Throw(uml.IllegalArgumentException, anchor)
	}
}

// SetPart1Anchor sets the anchor of the first participant.
func (m *MessageBase) SetPart1Anchor(anchor string, ok bool) { m.anchor1, m.hasAnchor[1] = anchor, ok }

// SetPart2Anchor sets the anchor of the second participant.
func (m *MessageBase) SetPart2Anchor(anchor string, ok bool) { m.anchor2, m.hasAnchor[2] = anchor, ok }

// Anchor returns the anchor of the message; ok is false when there is none.
func (m *MessageBase) Anchor() (string, bool) { return m.anchor, m.hasAnchor[0] }

// Part1Anchor returns the anchor of the first participant.
func (m *MessageBase) Part1Anchor() (string, bool) { return m.anchor1, m.hasAnchor[1] }

// Part2Anchor returns the anchor of the second participant.
func (m *MessageBase) Part2Anchor() (string, bool) { return m.anchor2, m.hasAnchor[2] }

// StyleBuilder returns the style sheet at the time of the message.
func (m *MessageBase) StyleBuilder() *klimt.StyleBuilder { return m.styleBuilder }

// LineLocation returns the source line of the message.
func (m *MessageBase) LineLocation() *text.LineLocation { return m.location }

// Message is a message between two participants.
type Message struct {
	MessageBase
	p1, p2    *Participant
	multicast []*Participant
}

// NewMessage returns a message from p1 to p2; hasNumber is false when the
// message is not numbered.
func NewMessage(styleBuilder *klimt.StyleBuilder, p1, p2 *Participant, label *klimt.Display,
	arrowConfiguration *klimt.ArrowConfiguration, messageNumber string, hasNumber bool,
	location *text.LineLocation) *Message {
	m := &Message{p1: p1, p2: p2}
	m.MessageBase = newMessageBase(m, styleBuilder, label, arrowConfiguration, messageNumber, hasNumber, location)
	return m
}

// Participant1 implements AbstractMessage.
func (m *Message) Participant1() *Participant { return m.p1 }

// Participant2 implements AbstractMessage.
func (m *Message) Participant2() *Participant { return m.p2 }

// DealWith implements Event.
func (m *Message) DealWith(someone *Participant) bool { return someone == m.p1 || someone == m.p2 }

// CompatibleForCreate implements AbstractMessage.
func (m *Message) CompatibleForCreate(p *Participant) bool { return m.p1 != p && m.p2 == p }

// IsSelfMessage implements AbstractMessage.
func (m *Message) IsSelfMessage() bool { return m.p1 == m.p2 }

func (m *Message) overrideNotePosition(position NotePosition) NotePosition { return position }

// SetMulticast sets the other targets of the message ("A -> B & C").
func (m *Message) SetMulticast(multicast []*Participant) { m.multicast = multicast }

// Multicast returns the other targets of the message.
func (m *Message) Multicast() []*Participant { return m.multicast }

// MessageExoType is the direction of a message from or to the border.
type MessageExoType int

// Border message types, in upstream's order.
const (
	ExoFromLeft MessageExoType = iota
	ExoToLeft
	ExoFromRight
	ExoToRight
)

var messageExoTypeNames = [...]string{"FROM_LEFT", "TO_LEFT", "FROM_RIGHT", "TO_RIGHT"}

// String returns upstream's name, for example "FROM_LEFT".
func (t MessageExoType) String() string { return messageExoTypeNames[t] }

// Direction is 1 for messages going right and -1 for messages going left.
func (t MessageExoType) Direction() int {
	if t == ExoFromLeft || t == ExoToRight {
		return 1
	}
	return -1
}

// IsLeftBorder reports whether the message starts or ends on the left
// border.
func (t MessageExoType) IsLeftBorder() bool { return t == ExoFromLeft || t == ExoToLeft }

// IsRightBorder reports whether the message starts or ends on the right
// border.
func (t MessageExoType) IsRightBorder() bool { return t == ExoFromRight || t == ExoToRight }

// Reverse returns the opposite direction on the same border.
func (t MessageExoType) Reverse() MessageExoType {
	switch t {
	case ExoFromLeft:
		return ExoToLeft
	case ExoToRight:
		return ExoFromRight
	case ExoFromRight:
		return ExoToRight
	}
	return ExoFromLeft
}

// MessageExo is a message from or to the border of the diagram ("[->",
// "->]", "?->").
type MessageExo struct {
	MessageBase
	typ         MessageExoType
	participant *Participant
	shortArrow  bool
}

// NewMessageExo returns a border message.
func NewMessageExo(styleBuilder *klimt.StyleBuilder, p *Participant, typ MessageExoType, label *klimt.Display,
	arrowConfiguration *klimt.ArrowConfiguration, messageNumber string, hasNumber bool, shortArrow bool,
	location *text.LineLocation) *MessageExo {
	m := &MessageExo{typ: typ, participant: p, shortArrow: shortArrow}
	m.MessageBase = newMessageBase(m, styleBuilder, label, arrowConfiguration, messageNumber, hasNumber, location)
	return m
}

// IsShortArrow reports whether the message is a short arrow ("?").
func (m *MessageExo) IsShortArrow() bool { return m.shortArrow }

func (m *MessageExo) overrideNotePosition(NotePosition) NotePosition {
	if m.typ == ExoFromLeft || m.typ == ExoToLeft {
		return NoteRight
	}
	return NoteLeft
}

// Participant1 implements AbstractMessage.
func (m *MessageExo) Participant1() *Participant { return m.participant }

// Participant2 implements AbstractMessage.
func (m *MessageExo) Participant2() *Participant { return m.participant }

// Participant returns the participant of the message.
func (m *MessageExo) Participant() *Participant { return m.participant }

// Type returns the direction of the message.
func (m *MessageExo) Type() MessageExoType { return m.typ }

// DealWith implements Event.
func (m *MessageExo) DealWith(someone *Participant) bool { return m.participant == someone }

// CompatibleForCreate implements AbstractMessage.
func (m *MessageExo) CompatibleForCreate(p *Participant) bool { return p == m.participant }

// IsSelfMessage implements AbstractMessage.
func (*MessageExo) IsSelfMessage() bool { return false }

// MessageNumber is the autonumber of a message in its label, a creole
// line of its own.
type MessageNumber string

// String implements klimt.CharSequence.
func (n MessageNumber) String() string { return string(n) }

// MessageNumberMarker implements klimt.MessageNumber.
func (MessageNumber) MessageNumberMarker() {}

var boldTags = regexp.MustCompile(`</?b>`)

// NumberRaw returns the number without bold tags.
func (n MessageNumber) NumberRaw() string { return boldTags.ReplaceAllString(string(n), "") }
