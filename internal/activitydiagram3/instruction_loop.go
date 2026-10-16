// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionWhile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionRepeat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// InstructionWhile is "while" ... "endwhile".
type InstructionWhile struct {
	withNote
	repeatList       *InstructionList
	parent           Instruction
	nextLinkRenderer *LinkRendering
	color            color.HColor
	killed           bool

	test *klimt.Display
	yes  *klimt.Display

	testCalled bool

	outColor *LinkRendering
	swimlane *Swimlane

	specialOut Instruction

	boxStyle    BoxStyle
	hasBoxStyle bool

	backward   *klimt.Display
	stereotype *klimt.Stereotype

	incoming1           *LinkRendering
	incoming2           *LinkRendering
	backwardCalled      bool
	currentStyleBuilder *klimt.StyleBuilder
}

var _ InstructionCollection = (*InstructionWhile)(nil)

// NewInstructionWhile returns a loop of condition test whose arrow into the
// loop is labeled yes; c may be nil.
func NewInstructionWhile(swimlane *Swimlane, parent Instruction, test *klimt.Display,
	nextLinkRenderer *LinkRendering, yes *klimt.Display, c color.HColor,
	currentStyleBuilder *klimt.StyleBuilder) *InstructionWhile {
	return &InstructionWhile{
		repeatList:          InstructionListEmpty(),
		parent:              parent,
		test:                requireDisplay(test),
		nextLinkRenderer:    requireLink(nextLinkRenderer),
		yes:                 requireDisplay(yes),
		swimlane:            swimlane,
		currentStyleBuilder: currentStyleBuilder,
		color:               c,
		outColor:            LinkRenderingNone(),
		backward:            klimt.DisplayNULL(),
		incoming1:           LinkRenderingNone(),
		incoming2:           LinkRenderingNone(),
	}
}

// OverwriteYes replaces the label of the arrow into the loop.
func (w *InstructionWhile) OverwriteYes(yes *klimt.Display) { w.yes = yes }

// Add implements Instruction.
func (w *InstructionWhile) Add(ins Instruction) command.Result { return w.repeatList.Add(ins) }

// Parent returns the instruction that holds the loop.
func (w *InstructionWhile) Parent() Instruction { return w.parent }

// Kill implements Instruction: after "endwhile" the loop itself is killed.
func (w *InstructionWhile) Kill() bool {
	if w.testCalled {
		w.killed = true
		return true
	}
	return w.repeatList.Kill()
}

// InLinkRendering implements Instruction.
func (w *InstructionWhile) InLinkRendering() *LinkRendering { return w.nextLinkRenderer }

// OutDisplay sets the label of the arrow that leaves the loop.
func (w *InstructionWhile) OutDisplay(out *klimt.Display) {
	w.outColor = w.outColor.WithDisplay(requireDisplay(out))
}

// OutColor sets the colors of the arrow that leaves the loop.
func (w *InstructionWhile) OutColor(rainbow *Rainbow) { w.outColor = w.outColor.WithRainbow(rainbow) }

// AddNote implements Instruction: the note goes to the loop when it is
// empty, else to its last instruction.
func (w *InstructionWhile) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if w.repeatList.IsEmpty() {
		return w.withNote.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	return w.repeatList.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable.
func (w *InstructionWhile) Swimlanes() SwimlaneSet { return w.repeatList.Swimlanes() }

// SwimlaneIn implements Swimable: the lane of the parent.
func (w *InstructionWhile) SwimlaneIn() *Swimlane { return w.parent.SwimlaneIn() }

// SwimlaneOut implements Swimable: the lane of the parent.
func (w *InstructionWhile) SwimlaneOut() *Swimlane { return w.parent.SwimlaneOut() }

// Last implements InstructionCollection.
func (w *InstructionWhile) Last() Instruction { return w.repeatList.Last() }

// SetSpecial sets the "stop" or "end" that follows the loop.
func (w *InstructionWhile) SetSpecial(special Instruction) { w.specialOut = special }

// ContainsBreak implements Instruction.
func (w *InstructionWhile) ContainsBreak() bool { return w.repeatList.ContainsBreak() }

// SetBackward sets the "backward" action of the loop.
func (w *InstructionWhile) SetBackward(label *klimt.Display, boxStyle BoxStyle, incoming1, incoming2 *LinkRendering,
	stereotype *klimt.Stereotype) {
	w.backward = label
	w.boxStyle = boxStyle
	w.hasBoxStyle = true
	w.incoming1 = incoming1
	w.incoming2 = incoming2
	w.backwardCalled = true
	w.stereotype = stereotype
}

// Incoming sets the arrow back to the condition, unless "backward" set it,
// and marks the end of the loop.
func (w *InstructionWhile) Incoming(incoming *LinkRendering) {
	if !w.backwardCalled {
		w.incoming1 = incoming
		w.incoming2 = incoming
	}
	w.testCalled = true
}

// RepeatList returns the instructions of the loop.
func (w *InstructionWhile) RepeatList() *InstructionList { return w.repeatList }

// Test returns the condition.
func (w *InstructionWhile) Test() *klimt.Display { return w.test }

// Yes returns the label of the arrow into the loop.
func (w *InstructionWhile) Yes() *klimt.Display { return w.yes }

// Color returns the color of the diamond, nil for none.
func (w *InstructionWhile) Color() color.HColor { return w.color }

// IsKilled reports whether the loop is killed.
func (w *InstructionWhile) IsKilled() bool { return w.killed }

// IsTestCalled reports whether "endwhile" was read.
func (w *InstructionWhile) IsTestCalled() bool { return w.testCalled }

// OutRendering returns the arrow that leaves the loop.
func (w *InstructionWhile) OutRendering() *LinkRendering { return w.outColor }

// Swimlane returns the lane of the loop.
func (w *InstructionWhile) Swimlane() *Swimlane { return w.swimlane }

// Special returns the "stop" or "end" that follows the loop, nil for none.
func (w *InstructionWhile) Special() Instruction { return w.specialOut }

// BoxStyle returns the shape of the "backward" action; ok is false before
// "backward" (null upstream).
func (w *InstructionWhile) BoxStyle() (BoxStyle, bool) { return w.boxStyle, w.hasBoxStyle }

// Backward returns the "backward" action, the null display for none.
func (w *InstructionWhile) Backward() *klimt.Display { return w.backward }

// Stereotype returns the stereotype of the "backward" action, nil for none.
func (w *InstructionWhile) Stereotype() *klimt.Stereotype { return w.stereotype }

// Incoming1 returns the arrow into the "backward" action.
func (w *InstructionWhile) Incoming1() *LinkRendering { return w.incoming1 }

// Incoming2 returns the arrow out of the "backward" action.
func (w *InstructionWhile) Incoming2() *LinkRendering { return w.incoming2 }

// IsBackwardCalled reports whether "backward" was read.
func (w *InstructionWhile) IsBackwardCalled() bool { return w.backwardCalled }

// CurrentStyleBuilder returns the style builder of the diagram.
func (w *InstructionWhile) CurrentStyleBuilder() *klimt.StyleBuilder { return w.currentStyleBuilder }

// InstructionRepeat is "repeat" ... "repeat while".
type InstructionRepeat struct {
	repeatList       *InstructionList
	parent           Instruction
	nextLinkRenderer *LinkRendering
	swimlane         *Swimlane
	swimlanes        *Swimlanes
	swimlaneOut      *Swimlane
	swimlaneBackward *Swimlane
	boxStyle         BoxStyle
	hasBoxStyle      bool
	killed           bool
	boxStyleIn       BoxStyle

	backward *klimt.Display

	stereogroupLoop        *klimt.Stereogroup
	stereotypeBack         *klimt.Stereotype
	incoming1              *LinkRendering
	incoming2              *LinkRendering
	backwardNotes          []*PositionedNote
	test                   *klimt.Display
	yes                    *klimt.Display
	out                    *klimt.Display
	startLabel             *klimt.Display
	testCalled             bool
	endRepeatLinkRendering *LinkRendering

	currentStyleBuilder *klimt.StyleBuilder

	stereotype2 *klimt.Stereogroup
}

// ContainsBreak implements Instruction.
func (r *InstructionRepeat) ContainsBreak() bool { return r.repeatList.ContainsBreak() }

// NewInstructionRepeat returns a loop in the current lane of swimlanes,
// whose start action is startLabel (the null display for none). colors is
// unused, as upstream.
func NewInstructionRepeat(swimlanes *Swimlanes, parent Instruction, nextLinkRenderer *LinkRendering,
	startLabel *klimt.Display, boxStyleIn BoxStyle, colors *color.Colors,
	stereogroup *klimt.Stereogroup) *InstructionRepeat {
	swimlane := swimlanes.CurrentSwimlane()
	return &InstructionRepeat{
		currentStyleBuilder:    swimlanes.CurrentStyleBuilder(),
		swimlanes:              swimlanes,
		swimlane:               swimlane,
		repeatList:             NewInstructionList(swimlane),
		boxStyleIn:             boxStyleIn,
		startLabel:             startLabel,
		parent:                 parent,
		nextLinkRenderer:       requireLink(nextLinkRenderer),
		stereogroupLoop:        stereogroup,
		backward:               klimt.DisplayNULL(),
		incoming1:              LinkRenderingNone(),
		incoming2:              LinkRenderingNone(),
		test:                   klimt.DisplayNULL(),
		yes:                    klimt.DisplayNULL(),
		out:                    klimt.DisplayNULL(),
		endRepeatLinkRendering: LinkRenderingNone(),
		stereotype2:            klimt.StereogroupNone(),
	}
}

// IsLastOfTheParent reports whether the loop is the last instruction of
// its parent list.
func (r *InstructionRepeat) IsLastOfTheParent() bool {
	if list, ok := r.parent.(*InstructionList); ok {
		return list.Last() == Instruction(r)
	}
	return false
}

// SetBackward sets the "backward" action of the loop.
func (r *InstructionRepeat) SetBackward(label *klimt.Display, swimlaneBackward *Swimlane, boxStyle BoxStyle,
	incoming1, incoming2 *LinkRendering, stereotype *klimt.Stereotype) {
	r.backward = label
	r.swimlaneBackward = swimlaneBackward
	r.boxStyle = boxStyle
	r.hasBoxStyle = true
	r.incoming1 = incoming1
	r.incoming2 = incoming2
	r.stereotypeBack = stereotype
}

// HasBackward reports whether a "backward" action was set; it compares
// with the null display by identity, as upstream.
func (r *InstructionRepeat) HasBackward() bool { return r.backward != klimt.DisplayNULL() }

// Add implements Instruction.
func (r *InstructionRepeat) Add(ins Instruction) command.Result { return r.repeatList.Add(ins) }

// Parent returns the instruction that holds the loop.
func (r *InstructionRepeat) Parent() Instruction { return r.parent }

// SetTest closes the loop with its condition and labels; back, when not
// none, is the arrow back to the start.
func (r *InstructionRepeat) SetTest(test, yes, out *klimt.Display, endRepeatLinkRendering, back *LinkRendering,
	swimlaneOut *Swimlane, stereotype2 *klimt.Stereogroup) {
	r.stereotype2 = stereotype2
	r.swimlaneOut = swimlaneOut
	r.test = requireDisplay(test)
	r.yes = requireDisplay(yes)
	r.out = requireDisplay(out)
	r.endRepeatLinkRendering = endRepeatLinkRendering
	if !back.IsNone() {
		r.incoming1 = back
	}
	r.testCalled = true
}

// Kill implements Instruction: after "repeat while" the loop itself is
// killed.
func (r *InstructionRepeat) Kill() bool {
	if r.testCalled {
		r.killed = true
		return true
	}
	return r.repeatList.Kill()
}

// InLinkRendering implements Instruction.
func (r *InstructionRepeat) InLinkRendering() *LinkRendering { return r.nextLinkRenderer }

// AddNote implements Instruction: after "backward" the note goes to the
// backward action, else to the instructions of the loop.
func (r *InstructionRepeat) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if klimt.DisplayIsNull(r.backward) {
		return r.repeatList.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	r.backwardNotes = append(r.backwardNotes, NewPositionedNote(note, position, typ, swimlaneNote, colors, stereotype))
	return true
}

// Swimlanes implements Swimable.
func (r *InstructionRepeat) Swimlanes() SwimlaneSet { return r.repeatList.Swimlanes() }

// SwimlaneIn implements Swimable: the outgoing lane of the parent.
func (r *InstructionRepeat) SwimlaneIn() *Swimlane { return r.parent.SwimlaneOut() }

// SwimlaneOut implements Swimable: the outgoing lane of the parent.
func (r *InstructionRepeat) SwimlaneOut() *Swimlane { return r.parent.SwimlaneOut() }

// RepeatList returns the instructions of the loop.
func (r *InstructionRepeat) RepeatList() *InstructionList { return r.repeatList }

// Swimlane returns the lane of the loop.
func (r *InstructionRepeat) Swimlane() *Swimlane { return r.swimlane }

// SwimlaneOutOfTest returns the lane of "repeat while", nil before it.
func (r *InstructionRepeat) SwimlaneOutOfTest() *Swimlane { return r.swimlaneOut }

// SwimlaneBackward returns the lane of the "backward" action.
func (r *InstructionRepeat) SwimlaneBackward() *Swimlane { return r.swimlaneBackward }

// BoxStyle returns the shape of the "backward" action; ok is false before
// "backward" (null upstream).
func (r *InstructionRepeat) BoxStyle() (BoxStyle, bool) { return r.boxStyle, r.hasBoxStyle }

// IsKilled reports whether the loop is killed.
func (r *InstructionRepeat) IsKilled() bool { return r.killed }

// BoxStyleIn returns the shape of the start action.
func (r *InstructionRepeat) BoxStyleIn() BoxStyle { return r.boxStyleIn }

// Backward returns the "backward" action, the null display for none.
func (r *InstructionRepeat) Backward() *klimt.Display { return r.backward }

// StereogroupLoop returns the stereotypes of "repeat".
func (r *InstructionRepeat) StereogroupLoop() *klimt.Stereogroup { return r.stereogroupLoop }

// StereotypeBack returns the stereotype of the "backward" action.
func (r *InstructionRepeat) StereotypeBack() *klimt.Stereotype { return r.stereotypeBack }

// Incoming1 returns the arrow into the "backward" action.
func (r *InstructionRepeat) Incoming1() *LinkRendering { return r.incoming1 }

// Incoming2 returns the arrow out of the "backward" action.
func (r *InstructionRepeat) Incoming2() *LinkRendering { return r.incoming2 }

// BackwardNotes returns the notes of the "backward" action.
func (r *InstructionRepeat) BackwardNotes() []*PositionedNote { return r.backwardNotes }

// Test returns the condition.
func (r *InstructionRepeat) Test() *klimt.Display { return r.test }

// Yes returns the label of the arrow back to the start.
func (r *InstructionRepeat) Yes() *klimt.Display { return r.yes }

// Out returns the label of the arrow that leaves the loop.
func (r *InstructionRepeat) Out() *klimt.Display { return r.out }

// StartLabel returns the start action, the null display for none.
func (r *InstructionRepeat) StartLabel() *klimt.Display { return r.startLabel }

// IsTestCalled reports whether "repeat while" was read.
func (r *InstructionRepeat) IsTestCalled() bool { return r.testCalled }

// EndRepeatLinkRendering returns the arrow into the condition.
func (r *InstructionRepeat) EndRepeatLinkRendering() *LinkRendering { return r.endRepeatLinkRendering }

// Stereotype2 returns the stereotypes of "repeat while".
func (r *InstructionRepeat) Stereotype2() *klimt.Stereogroup { return r.stereotype2 }

// DiagramSwimlanes returns the lanes of the diagram, which the layout reads.
func (r *InstructionRepeat) DiagramSwimlanes() *Swimlanes { return r.swimlanes }

// CurrentStyleBuilder returns the style builder of the diagram.
func (r *InstructionRepeat) CurrentStyleBuilder() *klimt.StyleBuilder { return r.currentStyleBuilder }
