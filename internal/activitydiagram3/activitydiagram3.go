// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ActivityDiagram3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"errors"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// ErrNotDrawable is returned by Export: the ftile layout and the drawing of
// activity diagrams are not ported yet.
var ErrNotDrawable = errors.New("drawing of activity diagrams is not implemented yet")

// SwimlaneStrategy tells whether swimlanes may still be declared.
type SwimlaneStrategy int

// Swimlane strategies; SwimlaneStrategyNull is upstream's null (no
// instruction and no swimlane yet).
const (
	SwimlaneStrategyNull SwimlaneStrategy = iota
	SwimlaneForbidden
	SwimlaneAllowed
)

var swimlaneStrategyNames = [...]string{"null", "SWIMLANE_FORBIDDEN", "SWIMLANE_ALLOWED"}

// String returns upstream's name, "null" for none.
func (s SwimlaneStrategy) String() string { return swimlaneStrategyNames[s] }

// ActivityDiagram3 is a parsed activity diagram (new syntax).
type ActivityDiagram3 struct {
	uml.Base

	swimlaneStrategy SwimlaneStrategy
	swimlanes        *Swimlanes
	hasURL           bool
}

// NewActivityDiagram3 returns an empty activity diagram for src; md5map
// holds the base64 images of the source.
func NewActivityDiagram3(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *ActivityDiagram3 {
	d := &ActivityDiagram3{Base: uml.NewBase(ctx, src, core.Activity, md5map)}
	d.swimlanes = NewSwimlanes(d.Skin(), d.Pragma())
	return d
}

var _ core.TypedDiagram = (*ActivityDiagram3)(nil)

// Description implements core.Diagram: "activity3".
func (d *ActivityDiagram3) Description() string { return "activity3" }

// Export draws page index of the diagram. Drawing is not ported yet: it
// returns ErrNotDrawable.
func (d *ActivityDiagram3) Export(index int) error { return ErrNotDrawable }

// Swimlanes returns the lanes and the instruction tree.
func (d *ActivityDiagram3) Swimlanes() *Swimlanes { return d.swimlanes }

// SwimlaneStrategy returns whether swimlanes may still be declared.
func (d *ActivityDiagram3) SwimlaneStrategy() SwimlaneStrategy { return d.swimlaneStrategy }

// HasURL reports whether an action or a note has a link.
func (d *ActivityDiagram3) HasURL() bool { return d.hasURL }

func (d *ActivityDiagram3) manageSwimlaneStrategy() {
	if d.swimlaneStrategy == SwimlaneStrategyNull {
		d.swimlaneStrategy = SwimlaneForbidden
	}
}

// Swimlane makes the lane name current; it fails once an instruction was
// added before any lane.
func (d *ActivityDiagram3) Swimlane(name string, c color.HColor, label *klimt.Display) command.Result {
	if d.swimlaneStrategy == SwimlaneStrategyNull {
		d.swimlaneStrategy = SwimlaneAllowed
	}
	if d.swimlaneStrategy == SwimlaneForbidden {
		return command.Error("This swimlane must be defined at the start of the diagram.")
	}
	d.swimlanes.Swimlane(name, c, label)
	return command.Ok()
}

func (d *ActivityDiagram3) setCurrent(ins Instruction) { d.swimlanes.SetCurrent(ins) }

func (d *ActivityDiagram3) current() Instruction { return d.swimlanes.Current() }

func (d *ActivityDiagram3) nextLinkRenderer() *LinkRendering { return d.swimlanes.NextLinkRenderer() }

// AddActivity adds an action; url may be nil. An unknown color in the
// stereotypes gives upstream's result for NoSuchColorException.
func (d *ActivityDiagram3) AddActivity(activity *klimt.Display, boxStyle BoxStyle, url *klimt.URL,
	stereogroup *klimt.Stereogroup) command.Result {
	result, err := d.addActivity(activity, boxStyle, url, stereogroup)
	if err != nil {
		return colorResult(err)
	}
	return result
}

// addActivity is AddActivity; it returns the error of an unknown color in
// the stereotypes (upstream's NoSuchColorException).
func (d *ActivityDiagram3) addActivity(activity *klimt.Display, boxStyle BoxStyle, url *klimt.URL,
	stereogroup *klimt.Stereogroup) (command.Result, error) {
	colors, err := stereogroup.InnerColors(d.Skin().IHtmlColorSet())
	if err != nil {
		return command.Result{}, err
	}
	d.manageSwimlaneStrategy()
	ins := NewInstructionSimple(activity, d.nextLinkRenderer(), d.swimlanes.CurrentSwimlane(), boxStyle, url, colors,
		stereogroup.BuildStereotype(), d.CurrentStyleBuilder())
	if added := d.current().Add(ins); !added.IsOK() {
		return added, nil
	}
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.manageHasURL(activity)
	if url != nil {
		d.hasURL = true
	}
	return command.Ok(), nil
}

// AddSpot adds a connector; c may be nil.
func (d *ActivityDiagram3) AddSpot(spot string, c color.HColor) {
	ins := NewInstructionSpot(spot, c, d.nextLinkRenderer(), d.swimlanes.CurrentSwimlane())
	d.current().Add(ins)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.manageSwimlaneStrategy()
}

// AddGoto adds "goto name".
func (d *ActivityDiagram3) AddGoto(name string) command.Result {
	ins := NewInstructionGoto(d.swimlanes.CurrentSwimlane(), name)
	d.current().Add(ins)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	return command.Ok()
}

// AddLabel adds "label name".
func (d *ActivityDiagram3) AddLabel(name string) command.Result {
	ins := NewInstructionLabel(d.swimlanes.CurrentSwimlane(), name)
	d.current().Add(ins)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	return command.Ok()
}

// Start adds a start node.
func (d *ActivityDiagram3) Start(colors *color.Colors) {
	d.manageSwimlaneStrategy()
	d.current().Add(NewInstructionStart(colors, d.swimlanes.CurrentSwimlane(), d.nextLinkRenderer()))
	d.setNextLinkRendererInternal(LinkRenderingNone())
}

// Stop adds a stop node.
func (d *ActivityDiagram3) Stop(colors *color.Colors) {
	d.manageSwimlaneStrategy()
	ins := NewInstructionStop(colors, d.swimlanes.CurrentSwimlane(), d.nextLinkRenderer())
	if d.manageSpecialStopEndAfterEndWhile(ins) {
		return
	}
	d.current().Add(ins)
}

// End adds an end node.
func (d *ActivityDiagram3) End(colors *color.Colors) {
	d.manageSwimlaneStrategy()
	ins := NewInstructionEnd(colors, d.swimlanes.CurrentSwimlane(), d.nextLinkRenderer())
	if d.manageSpecialStopEndAfterEndWhile(ins) {
		return
	}
	d.current().Add(ins)
}

// manageSpecialStopEndAfterEndWhile attaches a stop or end that follows a
// loop without break to the loop.
func (d *ActivityDiagram3) manageSpecialStopEndAfterEndWhile(special Instruction) bool {
	current, ok := d.current().(*InstructionList)
	if !ok {
		return false
	}
	instructionWhile, ok := current.Last().(*InstructionWhile)
	if !ok {
		return false
	}
	if instructionWhile.ContainsBreak() {
		return false
	}
	instructionWhile.SetSpecial(special)
	return true
}

// BreakInstruction adds "break".
func (d *ActivityDiagram3) BreakInstruction() {
	d.manageSwimlaneStrategy()
	d.current().Add(NewInstructionBreak(d.swimlanes.CurrentSwimlane(), d.nextLinkRenderer()))
}

// Fork starts a fork.
func (d *ActivityDiagram3) Fork(colors *color.Colors) {
	d.manageSwimlaneStrategy()
	instructionFork := NewInstructionFork(d.current(), d.nextLinkRenderer(), d.swimlanes.CurrentSwimlane(), colors)
	d.current().Add(instructionFork)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.setCurrent(instructionFork)
}

// ForkAgain starts a new branch of the current fork.
func (d *ActivityDiagram3) ForkAgain() command.Result {
	if currentFork, ok := d.current().(*InstructionFork); ok {
		currentFork.ManageOutRendering(d.nextLinkRenderer(), false)
		d.setNextLinkRendererInternal(LinkRenderingNone())
		currentFork.ForkAgain(d.swimlanes.CurrentSwimlane())
		return command.Ok()
	}
	return command.Error("Cannot find fork")
}

// EndFork closes the current fork; label is the "{label}" of the bar, ok
// false for none.
func (d *ActivityDiagram3) EndFork(forkStyle ForkStyle, label string, ok bool) command.Result {
	if currentFork, isFork := d.current().(*InstructionFork); isFork {
		currentFork.SetStyle(forkStyle, label, ok, d.swimlanes.CurrentSwimlane())
		currentFork.ManageOutRendering(d.nextLinkRenderer(), true)
		d.setNextLinkRendererInternal(LinkRenderingNone())
		d.setCurrent(currentFork.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find fork")
}

// Split starts a split.
func (d *ActivityDiagram3) Split() {
	instructionSplit := NewInstructionSplit(d.current(), d.nextLinkRenderer(), d.swimlanes.CurrentSwimlane(), nil)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.current().Add(instructionSplit)
	d.setCurrent(instructionSplit)
}

// SplitAgain starts a new branch of the current split.
func (d *ActivityDiagram3) SplitAgain() command.Result {
	if split, ok := d.current().(*InstructionSplit); ok {
		split.SplitAgain(d.nextLinkRenderer())
		d.setNextLinkRendererInternal(LinkRenderingNone())
		return command.Ok()
	}
	return command.Error("Cannot find split")
}

// EndSplit closes the current split.
func (d *ActivityDiagram3) EndSplit() command.Result {
	if split, ok := d.current().(*InstructionSplit); ok {
		split.EndSplit(d.nextLinkRenderer(), d.swimlanes.CurrentSwimlane())
		d.setNextLinkRendererInternal(LinkRenderingNone())
		d.setCurrent(split.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find split")
}

// StartSwitch starts a switch.
func (d *ActivityDiagram3) StartSwitch(test *klimt.Display, colors *color.Colors) {
	d.manageSwimlaneStrategy()
	instructionSwitch := NewInstructionSwitch(d.swimlanes.CurrentSwimlane(), d.current(), test, d.nextLinkRenderer(),
		colors)
	d.current().Add(instructionSwitch)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.setCurrent(instructionSwitch)
}

// SwitchCase starts a new case of the current switch.
func (d *ActivityDiagram3) SwitchCase(labelCase *klimt.Display) command.Result {
	if s, ok := d.current().(*InstructionSwitch); ok {
		if !s.SwitchCase(labelCase, d.nextLinkRenderer()) {
			return command.Error("You cannot put an elseIf here")
		}
		d.setNextLinkRendererInternal(LinkRenderingNone())
		return command.Ok()
	}
	return command.Error("Cannot find switch")
}

// EndSwitch closes the current switch.
func (d *ActivityDiagram3) EndSwitch(colors *color.Colors) command.Result {
	if s, ok := d.current().(*InstructionSwitch); ok {
		s.EndSwitch(d.nextLinkRenderer(), colors)
		d.setNextLinkRendererInternal(LinkRenderingNone())
		d.setCurrent(s.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find switch")
}

// StartIf starts an "if"; c, url and stereotype may be nil.
func (d *ActivityDiagram3) StartIf(test, whenThen *klimt.Display, c color.HColor, url *klimt.URL,
	stereotype *klimt.Stereotype) {
	d.manageSwimlaneStrategy()
	instructionIf := NewInstructionIf(d.swimlanes.CurrentSwimlane(), d.current(), test,
		LinkRenderingNone().WithDisplay(whenThen), d.nextLinkRenderer(), c, d.Skin(), url, stereotype)
	d.current().Add(instructionIf)
	d.setNextLinkRendererInternal(LinkRenderingNone())
	d.setCurrent(instructionIf)
}

// ElseIf starts a new branch of the current "if"; c may be nil.
func (d *ActivityDiagram3) ElseIf(inlabel *LinkRendering, test *klimt.Display, whenThen *LinkRendering,
	c color.HColor) command.Result {
	if i, ok := d.current().(*InstructionIf); ok {
		if !i.ElseIf(inlabel, test, whenThen, d.nextLinkRenderer(), c) {
			return command.Error("You cannot put an elseIf here")
		}
		d.setNextLinkRendererInternal(LinkRenderingNone())
		return command.Ok()
	}
	return command.Error("Cannot find if")
}

// Else2 starts the "else" branch of the current "if".
func (d *ActivityDiagram3) Else2(whenElse *LinkRendering) command.Result {
	if i, ok := d.current().(*InstructionIf); ok {
		if !i.SwitchToElse2(whenElse, d.nextLinkRenderer()) {
			return command.Error("Cannot find if")
		}
		d.setNextLinkRendererInternal(LinkRenderingNone())
		return command.Ok()
	}
	return command.Error("Cannot find if")
}

// Endif closes the current "if".
func (d *ActivityDiagram3) Endif(colors *color.Colors) command.Result {
	if i, ok := d.current().(*InstructionIf); ok {
		i.Endif(d.nextLinkRenderer(), colors)
		d.setNextLinkRendererInternal(LinkRenderingNone())
		d.setCurrent(i.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find if")
}

// StartRepeat starts a "repeat" loop.
func (d *ActivityDiagram3) StartRepeat(label *klimt.Display, boxStyleIn BoxStyle, colors *color.Colors,
	stereogroup *klimt.Stereogroup) {
	d.manageSwimlaneStrategy()
	instructionRepeat := NewInstructionRepeat(d.swimlanes, d.current(), d.nextLinkRenderer(), label, boxStyleIn,
		colors, stereogroup)
	d.current().Add(instructionRepeat)
	d.setCurrent(instructionRepeat)
	d.setNextLinkRendererInternal(LinkRenderingNone())
}

// RepeatWhile closes the current "repeat" loop.
func (d *ActivityDiagram3) RepeatWhile(label, yes, out, linkLabel *klimt.Display, linkColor *Rainbow,
	stereotype *klimt.Stereogroup) command.Result {
	d.manageSwimlaneStrategy()
	if instructionRepeat, ok := d.current().(*InstructionRepeat); ok {
		back := LinkRenderingCreate(linkColor).WithDisplay(linkLabel)
		instructionRepeat.SetTest(label, yes, out, d.nextLinkRenderer(), back, d.swimlanes.CurrentSwimlane(),
			stereotype)
		d.setCurrent(instructionRepeat.Parent())
		d.setNextLinkRendererInternal(LinkRenderingNone())
		return command.Ok()
	}
	return command.Error("Cannot find repeat")
}

// Backward sets the "backward" action of the current loop.
func (d *ActivityDiagram3) Backward(label *klimt.Display, boxStyle BoxStyle, incoming1, incoming2 *LinkRendering,
	stereotype *klimt.Stereotype) command.Result {
	d.manageSwimlaneStrategy()
	switch current := d.current().(type) {
	case *InstructionRepeat:
		current.SetBackward(label, d.swimlanes.CurrentSwimlane(), boxStyle, incoming1, incoming2, stereotype)
		return command.Ok()
	case *InstructionWhile:
		current.SetBackward(label, boxStyle, incoming1, incoming2, stereotype)
		return command.Ok()
	}
	return command.Error("Cannot find repeat")
}

// DoWhile starts a "while" loop; c may be nil.
func (d *ActivityDiagram3) DoWhile(test, yes *klimt.Display, c color.HColor) {
	d.manageSwimlaneStrategy()
	instructionWhile := NewInstructionWhile(d.swimlanes.CurrentSwimlane(), d.current(), test, d.nextLinkRenderer(),
		yes, c, d.Skin().CurrentStyleBuilder())
	d.current().Add(instructionWhile)
	d.setCurrent(instructionWhile)
}

// Endwhile closes the current "while" loop.
func (d *ActivityDiagram3) Endwhile(out *klimt.Display) command.Result {
	if w, ok := d.current().(*InstructionWhile); ok {
		w.Incoming(d.nextLinkRenderer())
		w.OutDisplay(out)
		d.setNextLinkRendererInternal(LinkRenderingNone())
		d.setCurrent(w.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find while")
}

// Kill kills the last instruction ("kill", "detach").
func (d *ActivityDiagram3) Kill() command.Result {
	if !d.current().Kill() {
		return command.Error("kill cannot be used here")
	}
	return command.Ok()
}

// StartGroup starts a "partition" or other block; backColor may be nil.
func (d *ActivityDiagram3) StartGroup(name *klimt.Display, backColor color.HColor, typ GroupSymbol,
	style *klimt.Style) {
	d.manageSwimlaneStrategy()
	instructionGroup := NewInstructionGroup(d.current(), name, backColor, d.swimlanes.CurrentSwimlane(),
		d.nextLinkRenderer(), typ, style)
	d.current().Add(instructionGroup)
	d.setCurrent(instructionGroup)
}

// CloseGroup closes the current block.
func (d *ActivityDiagram3) CloseGroup() command.Result {
	if g, ok := d.current().(*InstructionGroup); ok {
		d.setCurrent(g.Parent())
		return command.Ok()
	}
	return command.Error("Cannot find group")
}

func (d *ActivityDiagram3) setNextLinkRendererInternal(link *LinkRendering) {
	d.swimlanes.SetNextLinkRenderer(requireLink(link))
}

func (d *ActivityDiagram3) setNextLink(linkRenderer *LinkRendering) {
	requireLink(linkRenderer)
	if current, ok := d.current().(InstructionCollection); ok {
		switch last := current.Last().(type) {
		case *InstructionWhile:
			last.OutColor(linkRenderer.Rainbow())
		case *InstructionIf:
			last.OutColor(linkRenderer)
		}
	}
	d.setNextLinkRendererInternal(linkRenderer)
}

// SetLabelNextArrow sets the label of the next arrow; just after "while" it
// replaces the label of the arrow into the loop.
func (d *ActivityDiagram3) SetLabelNextArrow(label *klimt.Display) {
	if w, ok := d.current().(*InstructionWhile); ok && w.Last() == nil {
		w.OverwriteYes(label)
		return
	}
	d.setNextLinkRendererInternal(d.nextLinkRenderer().WithDisplay(label))
}

// SetColorNextArrow sets the colors of the next arrow; nil does nothing.
func (d *ActivityDiagram3) SetColorNextArrow(c *Rainbow) {
	if c == nil {
		return
	}
	d.setNextLink(LinkRenderingCreate(c))
}

// AddNote attaches a note to the last instruction.
func (d *ActivityDiagram3) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	stereotype *klimt.Stereotype) command.Result {
	if !d.current().AddNote(note, position, typ, colors, d.swimlanes.CurrentSwimlane(), stereotype) {
		return command.Error("Cannot add note here")
	}
	d.manageHasURL(note)
	return command.Ok()
}

func (d *ActivityDiagram3) manageHasURL(display *klimt.Display) {
	if display.HasURL() {
		d.hasURL = true
	}
}
