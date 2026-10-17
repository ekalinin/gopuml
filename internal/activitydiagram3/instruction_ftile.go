// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionList.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionIf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSwitch.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionWhile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionRepeat.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionFork.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionSplit.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/InstructionGroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/Branch.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/WithNote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The createFtile methods of the instructions: each builds the tile of the
// instruction with the factory chain.

// eventuallyAddNote adds the notes to ftile (which may be nil), unless
// there are none.
func (w *withNote) eventuallyAddNote(factory FtileFactory, ftile Ftile, swimlane *Swimlane,
	verticalAlignment geom.VerticalAlignment) Ftile {
	if len(w.notes) == 0 {
		return ftile
	}
	return factory.AddNote(ftile, swimlane, w.notes, verticalAlignment)
}

// CreateFtile implements Instruction.
func (l *InstructionList) CreateFtile(factory FtileFactory) Ftile {
	if len(l.all) == 0 {
		var result Ftile = NewFtileEmpty(factory.SkinParam(), 0, 0, l.defaultSwimlane)
		// Not a typo: in that case, upstream decorates the entry link
		if l.outlinkRendering != nil {
			result = factory.DecorateIn(result, l.outlinkRendering)
		}
		return result
	}
	var breaks []WeldingPoint
	result := l.eventuallyAddNote(factory, nil, l.SwimlaneIn(), geom.VerticalAlignmentCenter)
	for _, ins := range l.all {
		cur := ins.CreateFtile(factory)
		breaks = append(breaks, cur.WeldingPoints()...)
		if !ins.InLinkRendering().IsNone() {
			cur = factory.DecorateIn(cur, ins.InLinkRendering())
		}
		if result == nil {
			result = cur
		} else {
			result = factory.Assembly(result, cur)
		}
	}
	if l.outlinkRendering != nil {
		result = factory.DecorateOut(result, l.outlinkRendering)
	}
	if len(breaks) > 0 {
		result = NewFtileDecorateWelding(result, breaks)
	}
	return result
}

// CreateFtile implements Instruction.
func (i *InstructionSimple) CreateFtile(factory FtileFactory) Ftile {
	result := factory.Activity(i.label, i.SwimlaneIn(), i.boxStyle, i.colors, i.stereotype, i.styleBuilder)
	if i.url != nil {
		result = factory.AddURL(result, i.url)
	}
	result = i.eventuallyAddNote(factory, result, result.SwimlaneIn(), geom.VerticalAlignmentCenter)
	if i.killed {
		return NewFtileKilled(result)
	}
	return result
}

// CreateFtile implements Instruction.
func (i *InstructionSpot) CreateFtile(factory FtileFactory) Ftile {
	result := factory.Spot(i.SwimlaneIn(), i.spot, i.color)
	result = i.eventuallyAddNote(factory, result, result.SwimlaneIn(), geom.VerticalAlignmentCenter)
	if i.killed {
		return NewFtileKilled(result)
	}
	return result
}

// CreateFtile implements Instruction.
func (i *InstructionGoto) CreateFtile(factory FtileFactory) Ftile {
	return NewFtileGoto(factory.SkinParam(), i.SwimlaneIn(), i.name)
}

// CreateFtile implements Instruction.
func (i *InstructionLabel) CreateFtile(factory FtileFactory) Ftile {
	return NewFtileLabel(factory.SkinParam(), i.SwimlaneIn(), i.name)
}

// CreateFtile implements Instruction.
func (i *InstructionStart) CreateFtile(factory FtileFactory) Ftile {
	result := factory.Start(i.SwimlaneIn(), i.colors)
	return i.eventuallyAddNote(factory, result, result.SwimlaneIn(), geom.VerticalAlignmentCenter)
}

// CreateFtile implements Instruction.
func (i *InstructionStop) CreateFtile(factory FtileFactory) Ftile {
	result := factory.Stop(i.SwimlaneIn(), i.colors)
	return i.eventuallyAddNote(factory, result, result.SwimlaneIn(), geom.VerticalAlignmentCenter)
}

// CreateFtile implements Instruction.
func (i *InstructionEnd) CreateFtile(factory FtileFactory) Ftile {
	result := factory.End(i.SwimlaneIn(), i.colors)
	return i.eventuallyAddNote(factory, result, result.SwimlaneIn(), geom.VerticalAlignmentCenter)
}

// CreateFtile implements Instruction.
func (i *InstructionBreak) CreateFtile(factory FtileFactory) Ftile {
	return NewFtileBreak(factory.SkinParam(), i.SwimlaneIn())
}

// CreateFtile implements Instruction.
func (i *InstructionIf) CreateFtile(factory FtileFactory) Ftile {
	for _, branch := range i.thens {
		branch.UpdateFtile(factory)
	}
	if i.elseBranch == nil {
		i.elseBranch = NewBranch(i.swimlane, LinkRenderingNone(), klimt.DisplayNULL(), nil, LinkRenderingNone(),
			i.stereotype)
	}
	i.elseBranch.UpdateFtile(factory)
	result := factory.CreateIf(i.swimlane, i.thens, i.elseBranch, i.outColor, i.topInlinkRendering, i.url,
		i.PositionedNotes(), i.stereotype, i.currentStyleBuilder)
	var weldingPoints []WeldingPoint
	for _, branch := range i.thens {
		weldingPoints = append(weldingPoints, branch.WeldingPoints()...)
	}
	weldingPoints = append(weldingPoints, i.elseBranch.WeldingPoints()...)
	if len(weldingPoints) > 0 {
		result = NewFtileDecorateWelding(result, weldingPoints)
	}
	return result
}

// CreateFtile implements Instruction.
func (s *InstructionSwitch) CreateFtile(factory FtileFactory) Ftile {
	for _, branch := range s.switches {
		branch.UpdateFtile(factory)
	}
	result := factory.CreateSwitch(s.swimlane, s.switches, s.afterEndwhile, s.topInlinkRendering, s.labelTest,
		s.colors, s.colors.MergeWith(s.endColors))
	return s.eventuallyAddNote(factory, result, s.SwimlaneIn(), geom.VerticalAlignmentTop)
}

// CreateFtile implements Instruction.
func (w *InstructionWhile) CreateFtile(factory FtileFactory) Ftile {
	var back Ftile
	if !klimt.DisplayIsNull(w.backward) {
		back = factory.Activity(w.backward, w.swimlane, w.boxStyle, color.EmptyColors(), w.stereotype, nil)
	}
	tmp := w.repeatList.CreateFtile(factory)
	tmp = factory.CreateWhile(w.outColor, w.swimlane, tmp, w.test, w.yes, w.color, w.specialOut, back, w.incoming1,
		w.incoming2, w.currentStyleBuilder)
	if len(w.PositionedNotes()) > 0 {
		tmp = FtileWithNoteOpaleCreate(tmp, w.PositionedNotes(), false, geom.VerticalAlignmentCenter)
	}
	if w.killed || w.specialOut != nil {
		return NewFtileKilled(tmp)
	}
	return tmp
}

// CreateFtile implements Instruction. As upstream, a loop without "repeat
// while" takes the arrow of the next instruction as the arrow back.
func (r *InstructionRepeat) CreateFtile(factory FtileFactory) Ftile {
	back := r.ftileBackward(factory)
	tmp := r.repeatList.CreateFtile(factory)
	tmp = factory.DecorateOut(tmp, r.endRepeatLinkRendering)
	if !r.testCalled && r.incoming1.IsNone() {
		r.incoming1 = r.swimlanes.NextLinkRenderer()
	}
	tmp = factory.Repeat(r.stereogroupLoop, r.stereotype2, r.boxStyleIn, r.swimlane, r.swimlaneOut, r.startLabel,
		tmp, r.test, r.yes, r.out, back, r.IsLastOfTheParent(), r.incoming1, r.incoming2, r.currentStyleBuilder)
	tmp = FtileUtilsWithSwimlaneIn(tmp, r.swimlane)
	if r.killed {
		return NewFtileKilled(tmp)
	}
	return tmp
}

func (r *InstructionRepeat) ftileBackward(factory FtileFactory) Ftile {
	if klimt.DisplayIsNull(r.backward) {
		return nil
	}
	result := factory.Activity(r.backward, r.swimlaneBackward, r.boxStyle, color.EmptyColors(), r.stereotypeBack,
		nil)
	if len(r.backwardNotes) > 0 {
		result = factory.AddNote(result, r.swimlaneBackward, r.backwardNotes, geom.VerticalAlignmentCenter)
	}
	return result
}

// CreateFtile implements Instruction.
func (f *InstructionFork) CreateFtile(factory FtileFactory) Ftile {
	all := make([]Ftile, 0, len(f.forks))
	for _, list := range f.forks {
		all = append(all, list.CreateFtile(factory))
	}
	result := factory.CreateParallel(all, f.style, f.label, f.hasLabel, f.swimlaneIn, f.swimlaneOut, f.colors)
	if len(f.PositionedNotes()) > 0 {
		result = FtileWithNoteOpaleCreate(result, f.PositionedNotes(), false, geom.VerticalAlignmentCenter)
	}
	return result
}

// CreateFtile implements Instruction.
func (s *InstructionSplit) CreateFtile(factory FtileFactory) Ftile {
	all := make([]Ftile, 0, len(s.splits))
	for _, list := range s.splits {
		all = append(all, list.CreateFtile(factory))
	}
	return factory.CreateParallel(all, ForkStyleSplit, "", false, s.swimlaneIn, s.swimlaneOut, s.colors)
}

// CreateFtile implements Instruction.
func (g *InstructionGroup) CreateFtile(factory FtileFactory) Ftile {
	tmp := g.list.CreateFtile(factory)
	if g.note != nil {
		tmp = NewFtileWithNotes(tmp, []*PositionedNote{g.note}, geom.VerticalAlignmentCenter)
	}
	return factory.CreateGroup(tmp, g.title, g.backColor, nil, g.typ, g.style)
}

// UpdateFtile builds the tile of the branch.
func (b *Branch) UpdateFtile(factory FtileFactory) {
	b.ftile = factory.DecorateOut(b.list.CreateFtile(factory), b.inlinkRendering)
}

// Ftile returns the tile of the branch, nil before UpdateFtile.
func (b *Branch) Ftile() Ftile { return b.ftile }

// WeldingPoints returns the welding points of the tile of the branch.
func (b *Branch) WeldingPoints() []WeldingPoint { return b.ftile.WeldingPoints() }

// InColor is getInColor: the colors of the arrow into the branch.
func (b *Branch) InColor(arrowColor *Rainbow) *Rainbow {
	if b.IsEmpty() {
		return b.ftile.OutLinkRendering().RainbowOr(arrowColor)
	}
	if b.labelPositive.Rainbow().Size() > 0 {
		return b.labelPositive.Rainbow()
	}
	linkIn := b.ftile.InLinkRendering()
	c := linkIn.RainbowOr(arrowColor)
	if c.Size() == 0 {
		return arrowColor
	}
	return c
}

// SkinParam returns the skin parameters of the tile of the branch.
func (b *Branch) SkinParam() klimt.ISkinParam { return b.ftile.SkinParam() }

// DisplayPositive is getDisplayPositive: the label of the arrow into the
// tile, else the label of the branch.
func (b *Branch) DisplayPositive() *klimt.Display {
	if b.ftile != nil {
		in := b.ftile.InLinkRendering()
		if in != nil && !klimt.DisplayIsNull(in.Display()) {
			return in.Display()
		}
	}
	return b.labelPositive.Display()
}

// branchTextBlock is Branch.getTextBlock: the label of an arrow, the empty
// text block for a nil display.
func (b *Branch) branchTextBlock(display *klimt.Display) klimt.TextBlock {
	if display == nil {
		return klimt.TextBlockUtilsEmptyTextBlock()
	}
	style := b.ArrowStyleSignature().MergedStyle(b.SkinParam().CurrentStyleBuilder())
	lineBreak := style.WrapWidth()
	fcArrow := style.FontConfiguration(b.SkinParam().IHtmlColorSet(), nil)
	return display.Create0(fcArrow, geom.HorizontalAlignmentLeft, b.SkinParam(), lineBreak, klimt.CreoleModeSimpleLine,
		nil, nil)
}

// TextBlockPositive returns the label of the arrow into the branch.
func (b *Branch) TextBlockPositive() klimt.TextBlock { return b.branchTextBlock(b.DisplayPositive()) }

// TextBlockSpecial returns the label of the arrow out of the branch.
func (b *Branch) TextBlockSpecial() klimt.TextBlock { return b.branchTextBlock(b.SpecialDisplay()) }
