// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/Branch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// branch is one side of a mindmap: the root and the nodes of that side.
type branch struct {
	root   *idea
	last   *idea
	finger *fingerImpl
}

func (b *branch) initRoot(styleBuilder *klimt.StyleBuilder, backColor color.HColor, label *klimt.Display,
	shape IdeaShape, stereotype *klimt.Stereotype) {
	b.root = createIdeaSimple(styleBuilder, backColor, label, shape, stereotype)
	b.last = b.root
}

func (b *branch) initFinger(skinParam klimt.ISkinParam, direction bool) {
	b.finger = buildFinger(b.root, skinParam, direction)
}

func (b *branch) parentOfLast(nb int) *idea {
	result := b.last
	for range nb {
		result = result.parent
	}
	return result
}

// add adds a node at level after the last one: as its child one level
// below, or as the child of one of its ancestors.
func (b *branch) add(styleBuilder *klimt.StyleBuilder, backColor color.HColor, level int, label *klimt.Display,
	shape IdeaShape, stereotype *klimt.Stereotype) command.Result {
	if b.last == nil {
		return command.Error("Check your indentation ?")
	}
	if level == b.last.level+1 {
		b.last = b.last.createIdea(styleBuilder, backColor, level, label, shape, stereotype)
		return command.Ok()
	}
	if level <= b.last.level {
		diff := b.last.level - level + 1
		b.last = b.parentOfLast(diff).createIdea(styleBuilder, backColor, level, label, shape, stereotype)
		return command.Ok()
	}
	return command.Error("error42L")
}

func (b *branch) hasChildren() bool {
	if b.root == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.mindmap.Idea.hasChildren()" because "this.root" is null`)
	}
	return b.root.hasChildren()
}

func (b *branch) hasRoot() bool { return b.root != nil }

func (b *branch) hasFinger() bool { return b.finger != nil }

// DrawU draws the finger of the branch.
func (b *branch) DrawU(ug klimt.UGraphic) {
	if b.finger != nil {
		b.finger.DrawU(ug)
	}
}

func (b *branch) halfThickness(stringBounder font.StringBounder) float64 {
	if b.finger == nil {
		return 0
	}
	return b.finger.fullThickness(stringBounder) / 2
}

func (b *branch) doNotDrawFirstPhalanx() { b.finger.doNotDrawFirstPhalanx() }

func (b *branch) x12(stringBounder font.StringBounder) float64 {
	if b.finger == nil {
		return 0
	}
	return b.finger.fullElongation(stringBounder) + b.finger.x12()
}
