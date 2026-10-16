// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionIf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/Branch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// Branch is a branch of an "if" or a "switch": its test, its labels and
// its instructions. The ftile of the branch is added with the layout.
type Branch struct {
	list      *InstructionList
	labelTest *klimt.Display

	labelPositive *LinkRendering

	inlinkRendering *LinkRendering
	inlabel         *LinkRendering
	special         *LinkRendering

	color         color.HColor
	specialColors *color.Colors
}

// NewBranch returns an empty branch in swimlane. labelPositive, labelTest
// and inlabel must not be nil; c and stereotype may be nil.
func NewBranch(swimlane *Swimlane, labelPositive *LinkRendering, labelTest *klimt.Display, c color.HColor,
	inlabel *LinkRendering, stereotype *klimt.Stereotype) *Branch {
	return &Branch{
		inlabel:         requireLink(inlabel),
		labelTest:       requireDisplay(labelTest),
		labelPositive:   requireLink(labelPositive),
		color:           c,
		list:            NewInstructionList(swimlane),
		inlinkRendering: LinkRenderingNone(),
	}
}

// ArrowStyleSignature is getDefaultStyleDefinitionArrow: the style
// signature of the arrows of the branch.
func (b *Branch) ArrowStyleSignature() klimt.StyleSignatureBasic { return arrowStyleSignature() }

// ContainsBreak reports whether a "break" is in the branch.
func (b *Branch) ContainsBreak() bool { return b.list.ContainsBreak() }

// Add appends an instruction to the branch.
func (b *Branch) Add(ins Instruction) command.Result {
	b.list.Add(ins)
	return command.Ok()
}

// Kill kills the last instruction of the branch.
func (b *Branch) Kill() bool { return b.list.Kill() }

// AddNote attaches a note to the last instruction of the branch.
func (b *Branch) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	return b.list.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// SetInlinkRendering sets the arrow that leaves the branch.
func (b *Branch) SetInlinkRendering(inlinkRendering *LinkRendering) {
	b.inlinkRendering = requireLink(inlinkRendering)
}

// InlinkRendering returns the arrow that leaves the branch.
func (b *Branch) InlinkRendering() *LinkRendering { return b.inlinkRendering }

// Swimlanes returns the lanes of the instructions of the branch.
func (b *Branch) Swimlanes() SwimlaneSet { return b.list.Swimlanes() }

// LabelTest returns the test of the branch.
func (b *Branch) LabelTest() *klimt.Display { return b.labelTest }

// LabelPositive returns the arrow that enters the branch.
func (b *Branch) LabelPositive() *LinkRendering { return b.labelPositive }

// Out is getOut: the colors of the special arrow, else of the arrow that
// leaves the branch.
func (b *Branch) Out() *Rainbow {
	if b.special != nil {
		return b.special.Rainbow()
	}
	if b.inlinkRendering == nil {
		return nil
	}
	return b.inlinkRendering.Rainbow()
}

// Inlabel returns the label of the arrow that comes from the previous
// branch ("(label) elseif").
func (b *Branch) Inlabel() *klimt.Display { return b.inlabel.Display() }

// InlabelRendering returns the arrow that comes from the previous branch.
func (b *Branch) InlabelRendering() *LinkRendering { return b.inlabel }

// InRainbow returns the colors of the arrow that comes from the previous
// branch, defaultColor when it has none.
func (b *Branch) InRainbow(defaultColor *Rainbow) *Rainbow { return b.inlabel.RainbowOr(defaultColor) }

// LabelPositiveRainbow returns the colors of the arrow that enters the
// branch, defaultColor when it has none.
func (b *Branch) LabelPositiveRainbow(defaultColor *Rainbow) *Rainbow {
	return b.labelPositive.RainbowOr(defaultColor)
}

// Color is getColor: the background of the special colors, else the
// color of the branch.
func (b *Branch) Color() color.HColor {
	if b.specialColors != nil && b.specialColors.Color(color.TypeBack) != nil {
		return b.specialColors.Color(color.TypeBack)
	}
	return b.color
}

// OwnColor returns the color given to the branch, nil for none.
func (b *Branch) OwnColor() color.HColor { return b.color }

// IsEmpty reports whether the branch has no instruction.
func (b *Branch) IsEmpty() bool { return b.list.IsEmpty() }

// Last returns the last instruction of the branch, nil for none.
func (b *Branch) Last() Instruction { return b.list.Last() }

// List returns the instructions of the branch.
func (b *Branch) List() *InstructionList { return b.list }

// IsOnlySingleStopOrSpot reports whether the branch only stops or jumps.
func (b *Branch) IsOnlySingleStopOrSpot() bool { return b.list.IsOnlySingleStopOrSpot() }

// SetSpecial sets the arrow that leaves the branch to the merge point and
// the colors of the merge point; colors may be nil.
func (b *Branch) SetSpecial(link *LinkRendering, colors *color.Colors) {
	b.special = link
	b.specialColors = colors
}

// Special returns the arrow set by SetSpecial, nil for none.
func (b *Branch) Special() *LinkRendering { return b.special }

// SpecialColors returns the colors set by SetSpecial, nil for none.
func (b *Branch) SpecialColors() *color.Colors { return b.specialColors }

// SpecialDisplay returns the label of the special arrow, nil when it has
// none.
func (b *Branch) SpecialDisplay() *klimt.Display {
	if b.special != nil && !klimt.DisplayIsNull(b.special.Display()) {
		return b.special.Display()
	}
	return nil
}

// InstructionIf is "if" ... "endif".
type InstructionIf struct {
	withNote
	thens       []*Branch
	elseBranch  *Branch
	endifCalled bool
	skinParam   klimt.ISkinParam
	url         *klimt.URL

	parent Instruction

	current            *Branch
	topInlinkRendering *LinkRendering
	outColor           *LinkRendering
	stereotype         *klimt.Stereotype

	swimlane            *Swimlane
	currentStyleBuilder *klimt.StyleBuilder
}

var _ InstructionCollection = (*InstructionIf)(nil)

// NewInstructionIf returns an "if" whose first branch has the test
// labelTest and the arrow whenThen; c, url and stereotype may be nil.
func NewInstructionIf(swimlane *Swimlane, parent Instruction, labelTest *klimt.Display, whenThen *LinkRendering,
	inlinkRendering *LinkRendering, c color.HColor, skinParam klimt.ISkinParam, url *klimt.URL,
	stereotype *klimt.Stereotype) *InstructionIf {
	i := &InstructionIf{
		url:                 url,
		stereotype:          stereotype,
		parent:              parent,
		skinParam:           skinParam,
		topInlinkRendering:  requireLink(inlinkRendering),
		swimlane:            swimlane,
		currentStyleBuilder: skinParam.CurrentStyleBuilder(),
		outColor:            LinkRenderingNone(),
	}
	i.thens = append(i.thens, NewBranch(swimlane, whenThen, labelTest, c, LinkRenderingNone(), stereotype))
	i.current = i.thens[0]
	return i
}

// ContainsBreak implements Instruction.
func (i *InstructionIf) ContainsBreak() bool {
	for _, branch := range i.thens {
		if branch.ContainsBreak() {
			return true
		}
	}
	if i.elseBranch != nil {
		return i.elseBranch.ContainsBreak()
	}
	return false
}

// Add implements Instruction: the instruction goes to the current branch.
func (i *InstructionIf) Add(ins Instruction) command.Result { return i.current.Add(ins) }

// Parent returns the instruction that holds the "if".
func (i *InstructionIf) Parent() Instruction { return i.parent }

// SwitchToElse2 starts the "else" branch; it returns false when there is
// already one.
func (i *InstructionIf) SwitchToElse2(whenElse, nextLinkRenderer *LinkRendering) bool {
	i.current.SetSpecial(nextLinkRenderer, nil)
	if i.elseBranch != nil {
		return false
	}
	i.current.SetInlinkRendering(nextLinkRenderer)
	i.elseBranch = NewBranch(i.swimlane, whenElse, klimt.DisplayNULL(), nil, LinkRenderingNone(), i.stereotype)
	i.current = i.elseBranch
	return true
}

// ElseIf starts a new branch; it returns false after the "else".
func (i *InstructionIf) ElseIf(inlabel *LinkRendering, test *klimt.Display, whenThen, nextLinkRenderer *LinkRendering,
	c color.HColor) bool {
	if i.elseBranch != nil {
		return false
	}
	i.current.SetSpecial(nextLinkRenderer, nil)
	i.current = NewBranch(i.swimlane, whenThen, test, c, inlabel, i.stereotype)
	i.thens = append(i.thens, i.current)
	return true
}

// Endif closes the "if", adding an empty "else" branch when there is none;
// colors may be nil.
func (i *InstructionIf) Endif(nextLinkRenderer *LinkRendering, colors *color.Colors) {
	i.endifCalled = true
	if i.elseBranch == nil {
		i.elseBranch = NewBranch(i.swimlane, LinkRenderingNone(), klimt.DisplayNULL(), nil, LinkRenderingNone(),
			i.stereotype)
	}
	i.elseBranch.SetSpecial(nextLinkRenderer, colors)
	i.current.SetInlinkRendering(nextLinkRenderer)
}

// Kill implements Instruction. After "endif", it only looks at the first
// branch and the "else" branch, as upstream (its loop returns on the first
// branch).
func (i *InstructionIf) Kill() bool {
	if i.endifCalled {
		for _, branch := range i.thens {
			if branch.Last() != nil && !branch.Last().Kill() {
				return false
			}
			if i.elseBranch != nil && i.elseBranch.Last() != nil && !i.elseBranch.Last().Kill() {
				return false
			}
			return true
		}
	}
	return i.current.Kill()
}

// InLinkRendering implements Instruction.
func (i *InstructionIf) InLinkRendering() *LinkRendering { return i.topInlinkRendering }

// AddNote implements Instruction: before "endif", the note goes to the
// current branch unless it is empty.
func (i *InstructionIf) AddNote(note *klimt.Display, position NotePosition, typ NoteType, colors *color.Colors,
	swimlaneNote *Swimlane, stereotype *klimt.Stereotype) bool {
	if i.endifCalled || i.current.IsEmpty() {
		return i.withNote.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
	}
	return i.current.AddNote(note, position, typ, colors, swimlaneNote, stereotype)
}

// Swimlanes implements Swimable.
func (i *InstructionIf) Swimlanes() SwimlaneSet {
	var result SwimlaneSet
	if i.swimlane != nil {
		result = result.add(i.swimlane)
	}
	for _, branch := range i.thens {
		result = result.addAll(branch.Swimlanes())
	}
	if i.elseBranch != nil {
		result = result.addAll(i.elseBranch.Swimlanes())
	}
	return result
}

// SwimlaneIn implements Swimable.
func (i *InstructionIf) SwimlaneIn() *Swimlane { return i.swimlane }

// SwimlaneOut implements Swimable.
func (i *InstructionIf) SwimlaneOut() *Swimlane { return i.swimlane }

// Last implements InstructionCollection: the last instruction of the
// current part.
func (i *InstructionIf) Last() Instruction {
	if i.elseBranch == nil {
		return i.thens[len(i.thens)-1].Last()
	}
	return i.elseBranch.Last()
}

// OutColor sets the arrow that leaves the "if".
func (i *InstructionIf) OutColor(outColor *LinkRendering) { i.outColor = outColor }

// Thens returns the branches before the "else".
func (i *InstructionIf) Thens() []*Branch { return i.thens }

// ElseBranch returns the "else" branch, nil before "else" or "endif".
func (i *InstructionIf) ElseBranch() *Branch { return i.elseBranch }

// Current returns the branch that receives the instructions.
func (i *InstructionIf) Current() *Branch { return i.current }

// IsEndifCalled reports whether "endif" was read.
func (i *InstructionIf) IsEndifCalled() bool { return i.endifCalled }

// URL returns the link of the diamond, nil for none.
func (i *InstructionIf) URL() *klimt.URL { return i.url }

// Stereotype returns the stereotype of the "if", nil for none.
func (i *InstructionIf) Stereotype() *klimt.Stereotype { return i.stereotype }

// OutRendering returns the arrow that leaves the "if".
func (i *InstructionIf) OutRendering() *LinkRendering { return i.outColor }

// Swimlane returns the lane of the "if".
func (i *InstructionIf) Swimlane() *Swimlane { return i.swimlane }

// SkinParam returns the skin parameters of the diagram.
func (i *InstructionIf) SkinParam() klimt.ISkinParam { return i.skinParam }

// CurrentStyleBuilder returns the style builder of the diagram.
func (i *InstructionIf) CurrentStyleBuilder() *klimt.StyleBuilder { return i.currentStyleBuilder }
