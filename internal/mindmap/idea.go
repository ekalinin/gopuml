// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/Idea.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/IdeaShape.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// IdeaShape is how a node is drawn: in a box, as text only, or not at all
// (PSEUDO, which only WBS diagrams use).
type IdeaShape int

// The shapes of upstream's IdeaShape.
const (
	IdeaShapeBox IdeaShape = iota
	IdeaShapeNone
	IdeaShapePseudo
)

// IdeaShapeFromDesc is upstream's IdeaShape.fromDesc: "_" is a node without
// a box, anything else (also "" for no match) a box.
func IdeaShapeFromDesc(s string) IdeaShape {
	if s == "_" {
		return IdeaShapeNone
	}
	return IdeaShapeBox
}

// idea is a node of a mindmap.
type idea struct {
	label        *klimt.Display
	level        int
	parent       *idea
	children     []*idea
	shape        IdeaShape
	backColor    color.HColor
	styleBuilder *klimt.StyleBuilder
	stereotype   *klimt.Stereotype
}

// mindmapNode is upstream's StyleQueries.MINDMAPDIAG_NODE with the SNames
// that follow node.
func mindmapNode(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameMindmapDiagram, klimt.SNameNode}, names...)...)
}

// defaultStyleDefinitionNode is the style query of the node at level, from
// its shape and its children, as WBS's WElement does.
func (i *idea) defaultStyleDefinitionNode(level int) klimt.StyleSignatureBasic {
	var query klimt.StyleSignatureBasic
	switch {
	case level == 0 && i.shape == IdeaShapeNone:
		query = mindmapNode(klimt.SNameRootNode, klimt.SNameBoxless)
	case level == 0:
		query = mindmapNode(klimt.SNameRootNode)
	case i.shape == IdeaShapeNone && len(i.children) == 0:
		query = mindmapNode(klimt.SNameLeafNode, klimt.SNameBoxless)
	case i.shape == IdeaShapeNone:
		query = mindmapNode(klimt.SNameBoxless)
	case len(i.children) == 0:
		query = mindmapNode(klimt.SNameLeafNode)
	default:
		query = mindmapNode()
	}
	return query.AddStereotypeOf(i.stereotype).AddLevel(level)
}

// style resolves the style of the node, cascading down the starred styles
// of its ancestors: a nearer ancestor's declaration beats a farther one's,
// each ancestor ranking one below the previous one.
func (i *idea) style() *klimt.Style {
	ancestorRank := 0
	result := i.styleBuilder.MergedStyleSpecial(i.defaultStyleDefinitionNode(i.level), ancestorRank)
	for up := i.parent; up != nil; up = up.parent {
		ss := up.defaultStyleDefinitionNode(i.level).AddStar()
		ancestorRank--
		styleParent := i.styleBuilder.MergedStyleSpecial(ss, ancestorRank)
		result = result.MergeWith(styleParent, klimt.MergeOverwriteExistingValue)
	}
	return result
}

// styleArrow is the style of the link to the node.
func (i *idea) styleArrow() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameMindmapDiagram,
		klimt.SNameArrow).AddStereotypeOf(i.stereotype).AddLevel(i.level).MergedStyle(i.styleBuilder)
}

// createIdeaSimple returns a root.
func createIdeaSimple(styleBuilder *klimt.StyleBuilder, backColor color.HColor, label *klimt.Display,
	shape IdeaShape, stereotype *klimt.Stereotype) *idea {
	return newIdea(styleBuilder, backColor, 0, nil, label, shape, stereotype)
}

// createIdea adds a child to the node.
func (i *idea) createIdea(styleBuilder *klimt.StyleBuilder, backColor color.HColor, newLevel int,
	newDisplay *klimt.Display, newShape IdeaShape, stereotype *klimt.Stereotype) *idea {
	result := newIdea(styleBuilder, backColor, newLevel, i, newDisplay, newShape, stereotype)
	i.children = append(i.children, result)
	return result
}

func newIdea(styleBuilder *klimt.StyleBuilder, backColor color.HColor, level int, parent *idea,
	label *klimt.Display, shape IdeaShape, stereotype *klimt.Stereotype) *idea {
	return &idea{backColor: backColor, styleBuilder: styleBuilder, label: label, level: level, parent: parent,
		shape: shape, stereotype: stereotype}
}

func (i *idea) hasChildren() bool { return len(i.children) > 0 }
