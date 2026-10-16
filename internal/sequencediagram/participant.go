// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/Participant.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/ParticipantType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/ParticipantEnglober.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/asciiverse"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// ParticipantType is the kind of a participant.
type ParticipantType int

// Participant types, in upstream's order.
const (
	TypeParticipant ParticipantType = iota
	TypeActor
	TypeBoundary
	TypeControl
	TypeEntity
	TypeQueue
	TypeDatabase
	TypeCollections
)

var participantTypeNames = [...]string{"PARTICIPANT", "ACTOR", "BOUNDARY", "CONTROL", "ENTITY", "QUEUE",
	"DATABASE", "COLLECTIONS"}

// String returns upstream's name, for example "ACTOR".
func (t ParticipantType) String() string { return participantTypeNames[t] }

// participantTypeOf is ParticipantType.valueOf of an upper-cased keyword.
func participantTypeOf(s string) ParticipantType {
	for i, n := range participantTypeNames {
		if n == s {
			return ParticipantType(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.sequencediagram.ParticipantType."+s)
	return 0
}

// StyleSignature returns the style signature of the type.
func (t ParticipantType) StyleSignature() klimt.StyleSignatureBasic {
	names := [...]klimt.SName{klimt.SNameParticipant, klimt.SNameActor, klimt.SNameBoundary, klimt.SNameControl,
		klimt.SNameEntity, klimt.SNameQueue, klimt.SNameDatabase, klimt.SNameCollections}
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram, names[t])
}

// Participant is a lifeline of a sequence diagram.
type Participant struct {
	code     string
	display  *klimt.Display
	typ      ParticipantType
	location *text.LineLocation
	uid      string
	order    int

	initialLife    int
	liveBackcolors []*klimt.Fashion

	stereotype            *klimt.Stereotype
	stereotypePositionTop bool
	// hiddenPortions is shared with the diagram, as upstream shares its set.
	hiddenPortions *uml.EntityPortions
	styleBuilder   *klimt.StyleBuilder

	colors *color.Colors
	url    *klimt.URL
}

func newParticipant(location *text.LineLocation, typ ParticipantType, code string, display *klimt.Display,
	hiddenPortions *uml.EntityPortions, order int, styleBuilder *klimt.StyleBuilder, uid string) *Participant {
	if code == "" {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if klimt.DisplayIsNull(display) || display.Size() == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &Participant{location: location, uid: uid, hiddenPortions: hiddenPortions, styleBuilder: styleBuilder,
		order: order, code: code, typ: typ, display: display, colors: color.EmptyColors()}
}

// Code returns the name of the participant in the source.
func (p *Participant) Code() string { return p.code }

// String is upstream's toString: the code.
func (p *Participant) String() string { return p.code }

// AsciiDimension returns the size of the head box in the text output: the
// code (not the display) with a border, 3 rows.
func (p *Participant) AsciiDimension() asciiverse.ADimension2D {
	return asciiverse.NewADimension2D(uml.JavaLength(p.code)+2, 3)
}

// AsciiDraw draws the head box with the code at the position of plan.
func (p *Participant) AsciiDraw(plan *asciiverse.InfinitePlan) {
	plan.DrawBox(uml.JavaLength(p.code)+2, 3)
	plan.Move(1, 1).DrawString(p.code)
}

// RawDisplay returns the text of the participant, without stereotype.
func (p *Participant) RawDisplay() *klimt.Display { return p.display }

// Display returns the text of the participant, underlined or not, with its
// stereotype unless "hide stereotype" is set.
func (p *Participant) Display(underlined bool) *klimt.Display {
	result := p.display
	if underlined {
		result = result.Underlined()
	}
	if p.stereotype != nil && !p.hiddenPortions.Contains(uml.PortionStereotype) {
		if p.stereotypePositionTop {
			result = result.AddFirst(p.stereotype)
		} else {
			result = result.Add(p.stereotype)
		}
	}
	return result
}

// Type returns the kind of the participant.
func (p *Participant) Type() ParticipantType { return p.typ }

// SetStereotype sets the stereotype; it panics when one is already set, as
// upstream.
func (p *Participant) SetStereotype(stereotype *klimt.Stereotype, positionTop bool) {
	if p.stereotype != nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	if stereotype == nil {
		uml.ThrowNPE("")
	}
	p.stereotype = stereotype
	p.stereotypePositionTop = positionTop
}

// Stereotype returns the stereotype, nil for none.
func (p *Participant) Stereotype() *klimt.Stereotype { return p.stereotype }

// StereotypePositionTop reports whether the stereotype is shown above the
// name.
func (p *Participant) StereotypePositionTop() bool { return p.stereotypePositionTop }

// InitialLife is the number of activations before the first message.
func (p *Participant) InitialLife() int { return p.initialLife }

// IncInitialLife adds an activation before the first message.
func (p *Participant) IncInitialLife(colors *klimt.Fashion) {
	p.initialLife++
	p.liveBackcolors = append(p.liveBackcolors, colors)
}

// LiveSpecificBackColors returns the colors of initial activation which.
func (p *Participant) LiveSpecificBackColors(which int) *klimt.Fashion {
	return p.liveBackcolors[which]
}

// Colors returns the specific colors of the participant.
func (p *Participant) Colors() *color.Colors { return p.colors }

// SetSpecificColorTOBEREMOVED adds a color, unless it is nil.
func (p *Participant) SetSpecificColorTOBEREMOVED(t color.ColorType, c color.HColor) {
	if c != nil {
		p.colors = p.colors.Add(t, c)
	}
}

// SetColors sets the specific colors.
func (p *Participant) SetColors(colors *color.Colors) { p.colors = colors }

// URL returns the link of the participant, nil for none.
func (p *Participant) URL() *klimt.URL { return p.url }

// SetURL sets the link.
func (p *Participant) SetURL(url *klimt.URL) { p.url = url }

// Order returns the "order" value.
func (p *Participant) Order() int { return p.order }

// Location returns the source line that declared the participant.
func (p *Participant) Location() *text.LineLocation { return p.location }

// UID returns the identifier used in the SVG output, for example "part1".
func (p *Participant) UID() string { return p.uid }

// StyleBuilder returns the style sheet at the time of declaration.
func (p *Participant) StyleBuilder() *klimt.StyleBuilder { return p.styleBuilder }

// StyleSignature returns the style signature of the participant.
func (p *Participant) StyleSignature() klimt.StyleSignature { return p.styleSignatureBasic() }

func (p *Participant) styleSignatureBasic() klimt.StyleSignatureBasic {
	return p.typ.StyleSignature().AddClickable(p.url)
}

// UsedStyles returns the style of the participant and the style of its
// stereotype.
func (p *Participant) UsedStyles() []*klimt.Style {
	tmp := p.StyleSignature().WithTOBECHANGED(p.stereotype).MergedStyle(p.styleBuilder)
	tmp = tmp.EventuallyOverrideColors(p.colors)
	stereo := p.styleSignatureBasic().ForStereotypeItself(p.stereotype).MergedStyle(p.styleBuilder)
	if tmp != nil {
		stereo = tmp.MergeWith(stereo, klimt.MergeOverwriteExistingValue)
	}
	return []*klimt.Style{tmp, stereo}
}

var (
	_ rose.Participant = (*Participant)(nil)
	_ klimt.WithStyle  = (*Participant)(nil)
)

// ParticipantEnglober is a "box" around participants.
type ParticipantEnglober struct {
	parent     *ParticipantEnglober
	title      *klimt.Display
	boxColor   color.HColor
	stereotype *klimt.Stereotype
}

// String is upstream's toString: the title.
func (e *ParticipantEnglober) String() string { return e.title.String() }

// Title returns the title of the box.
func (e *ParticipantEnglober) Title() *klimt.Display { return e.title }

// BoxColor returns the color of the box, nil for none.
func (e *ParticipantEnglober) BoxColor() color.HColor { return e.boxColor }

// Stereotype returns the stereotype of the box, nil for none.
func (e *ParticipantEnglober) Stereotype() *klimt.Stereotype { return e.stereotype }

// Parent returns the enclosing box, nil for none.
func (e *ParticipantEnglober) Parent() *ParticipantEnglober { return e.parent }

// Genealogy returns the boxes from the outermost to this one.
func (e *ParticipantEnglober) Genealogy() []*ParticipantEnglober {
	var result []*ParticipantEnglober
	for current := e; current != nil; current = current.parent {
		result = append([]*ParticipantEnglober{current}, result...)
	}
	return result
}
