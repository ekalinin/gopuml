// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/Grouping.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/GroupingLeaf.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/GroupingStart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/GroupingType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// GroupingType is the role of a fragment line.
type GroupingType int

// Grouping types, in upstream's order.
const (
	GroupingStartType GroupingType = iota
	GroupingStartPartition
	GroupingElse
	GroupingEnd
)

var groupingTypeNames = [...]string{"START", "START_PARTITION", "ELSE", "END"}

// String returns upstream's name, for example "START".
func (t GroupingType) String() string { return groupingTypeNames[t] }

// groupingTypeOf is GroupingType.getType of a lower-case keyword.
func groupingTypeOf(s string) GroupingType {
	switch uml.LowerCase(s) {
	case "opt", "alt", "loop", "par", "par2", "break", "group", "critical":
		return GroupingStartType
	case "partition":
		return GroupingStartPartition
	case "also", "else":
		return GroupingElse
	case "end":
		return GroupingEnd
	}
	uml.Throw(uml.IllegalArgumentException, s)
	return 0
}

// IsStart reports whether the type opens a fragment.
func (t GroupingType) IsStart() bool { return t == GroupingStartType || t == GroupingStartPartition }

// Grouping is a fragment line: GroupingStart opens a fragment, GroupingLeaf
// is an "else" or the "end".
type Grouping interface {
	Event
	base() *groupingBase
	Level() int
	BackColorGeneral() color.HColor
}

// groupingBase is upstream's Grouping.
type groupingBase struct {
	abstractEvent
	title            string
	typ              GroupingType
	comment          string
	hasComment       bool
	backColorElement color.HColor
	style            *klimt.Style
	styleHeader      *klimt.Style
}

func newGroupingBase(title, comment string, hasComment bool, typ GroupingType, backColorElement color.HColor,
	styleBuilder *klimt.StyleBuilder) groupingBase {
	g := groupingBase{title: title, comment: comment, hasComment: hasComment, typ: typ,
		backColorElement: backColorElement}
	g.style = g.StyleSignature().MergedStyle(styleBuilder)
	g.styleHeader = g.computeStyleHeader(styleBuilder)
	return g
}

func (g *groupingBase) base() *groupingBase { return g }

func seqSignature(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameSequenceDiagram}, names...)...)
}

// StyleSignature returns the style signature of the fragment.
func (g *groupingBase) StyleSignature() klimt.StyleSignatureBasic {
	if g.typ == GroupingStartPartition {
		return seqSignature(klimt.SNamePartition)
	}
	return seqSignature(klimt.SNameGroup)
}

func (g *groupingBase) computeStyleHeader(styleBuilder *klimt.StyleBuilder) *klimt.Style {
	var flat *klimt.Style
	if g.typ == GroupingStartPartition {
		flat = seqSignature(klimt.SNamePartition, klimt.SNameHeader).MergedStyle(styleBuilder)
		return flat
	}
	flat = seqSignature(klimt.SNameGroupHeader).MergedStyle(styleBuilder)
	nested := seqSignature(klimt.SNameGroup, klimt.SNameHeader).MergedStyle(styleBuilder)
	if flat == nil {
		return nested
	}
	return flat.MergeNestedChildOver(nested, g.style, klimt.MergeOverwriteExistingValue)
}

// Title returns the keyword of the fragment ("alt", "loop", ...), or the
// first part of a "group" label.
func (g *groupingBase) Title() string { return g.title }

// Type returns the role of the line.
func (g *groupingBase) Type() GroupingType { return g.typ }

// Comment returns the label; ok is false when there is none.
func (g *groupingBase) Comment() (string, bool) { return g.comment, g.hasComment }

// BackColorElement returns the header color, nil for none.
func (g *groupingBase) BackColorElement() color.HColor { return g.backColorElement }

func (g *groupingBase) usedStyles(backColorGeneral color.HColor) []*klimt.Style {
	var s, h *klimt.Style
	if g.style != nil {
		s = g.style.EventuallyOverrideColor(klimt.PNameBackGroundColor, backColorGeneral)
	}
	if g.styleHeader != nil {
		h = g.styleHeader.EventuallyOverrideColor(klimt.PNameBackGroundColor, g.backColorElement)
	}
	return []*klimt.Style{s, h}
}

// String is upstream's toString without the identity hash.
func (g *groupingBase) String() string { return g.typ.String() + " " + g.title }

// GroupingStart opens a fragment.
type GroupingStart struct {
	groupingBase
	children         []*GroupingLeaf
	backColorGeneral color.HColor
	parent           *GroupingStart
	parallel         bool
}

func newGroupingStart(typ GroupingType, title, comment string, hasComment bool, backColorGeneral,
	backColorElement color.HColor, parent *GroupingStart, styleBuilder *klimt.StyleBuilder) *GroupingStart {
	return &GroupingStart{groupingBase: newGroupingBase(title, comment, hasComment, typ, backColorElement, styleBuilder),
		backColorGeneral: backColorGeneral, parent: parent}
}

// UsedStyles returns the styles of the fragment and of its header.
func (g *GroupingStart) UsedStyles() []*klimt.Style {
	result := g.usedStyles(g.BackColorGeneral())
	if result[0] != nil {
		result[0] = result[0].EventuallyOverrideColor(klimt.PNameBackGroundColor, g.backColorGeneral)
	}
	return result
}

// Children returns the "else" and "end" lines of the fragment.
func (g *GroupingStart) Children() []*GroupingLeaf { return g.children }

func (g *GroupingStart) addChildren(leaf *GroupingLeaf) { g.children = append(g.children, leaf) }

// Level returns the nesting level, 0 for an outermost fragment.
func (g *GroupingStart) Level() int {
	if g.parent == nil {
		return 0
	}
	return g.parent.Level() + 1
}

// Parent returns the enclosing fragment, nil for none.
func (g *GroupingStart) Parent() *GroupingStart { return g.parent }

// BackColorGeneral returns the background color, nil for none.
func (g *GroupingStart) BackColorGeneral() color.HColor { return g.backColorGeneral }

// DealWith implements Event.
func (*GroupingStart) DealWith(*Participant) bool { return false }

// IsParallel implements Event.
func (g *GroupingStart) IsParallel() bool { return g.parallel || g.IsPar2GroupStart() }

// GoParallel marks the fragment as parallel to the previous event.
func (g *GroupingStart) GoParallel() { g.parallel = true }

// IsPar2GroupStart reports whether the fragment is a "par2" fragment.
func (g *GroupingStart) IsPar2GroupStart() bool { return g.title == "par2" }

// GroupingLeaf is an "else" or "end" line of a fragment.
type GroupingLeaf struct {
	groupingBase
	posYend
	start            *GroupingStart
	backColorGeneral color.HColor
	noteOnMessages   []*Note
}

func newGroupingLeaf(title, comment string, hasComment bool, typ GroupingType, backColorGeneral,
	backColorElement color.HColor, start *GroupingStart, styleBuilder *klimt.StyleBuilder) *GroupingLeaf {
	if start == nil {
		uml.ThrowNPE("")
	}
	g := &GroupingLeaf{groupingBase: newGroupingBase(title, comment, hasComment, typ, backColorElement, styleBuilder),
		backColorGeneral: backColorGeneral, start: start}
	start.addChildren(g)
	return g
}

// UsedStyles returns the styles of the line and of its header.
func (g *GroupingLeaf) UsedStyles() []*klimt.Style { return g.usedStyles(g.BackColorGeneral()) }

// JustAfter returns the next line of the fragment, nil after the last.
func (g *GroupingLeaf) JustAfter() Grouping {
	idx := -1
	for i, c := range g.start.children {
		if c == g {
			idx = i
		}
	}
	if idx == -1 {
		uml.Throw(uml.IllegalStateException, "")
	}
	if idx+1 >= len(g.start.children) {
		return nil
	}
	return g.start.children[idx+1]
}

// GroupingStart returns the fragment of the line.
func (g *GroupingLeaf) GroupingStart() *GroupingStart { return g.start }

// Level returns the nesting level of the fragment.
func (g *GroupingLeaf) Level() int { return g.start.Level() }

// BackColorGeneral returns the color of the line, else the color of the
// fragment.
func (g *GroupingLeaf) BackColorGeneral() color.HColor {
	if g.backColorGeneral == nil {
		return g.start.BackColorGeneral()
	}
	return g.backColorGeneral
}

// DealWith implements Event.
func (*GroupingLeaf) DealWith(*Participant) bool { return false }

// IsParallel implements Event: only the lines of a "par2" fragment are.
func (g *GroupingLeaf) IsParallel() bool { return g.start.IsPar2GroupStart() }

// AddLifeEvent implements EventWithDeactivate.
func (*GroupingLeaf) AddLifeEvent(*LifeEvent) bool { return true }

// AddNote implements EventWithNote: the note is kept even when its position
// is ignored.
func (g *GroupingLeaf) AddNote(note *Note) *klimt.Warning {
	g.noteOnMessages = append(g.noteOnMessages, note)
	if note.Position() != NoteLeft && note.Position() != NoteRight {
		return klimt.NewWarning("This position is ignored: " + note.Position().String())
	}
	return nil
}

// NoteOnMessages returns the notes attached to the line.
func (g *GroupingLeaf) NoteOnMessages() []*Note { return g.noteOnMessages }
