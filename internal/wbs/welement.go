// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WElement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/mindmap"
)

// wElement is a node of a WBS diagram, with its children on the left and on
// the right of the vertical line below it.
type wElement struct {
	backColor     color.HColor
	label         *klimt.Display
	level         int
	stereotype    *klimt.Stereotype
	parent        *wElement
	styleBuilder  *klimt.StyleBuilder
	childrenLeft  []*wElement
	childrenRight []*wElement
	shape         mindmap.IdeaShape
	// position and dimension are recorded while the node is drawn; position
	// is nil until then.
	position  *klimt.UTranslate
	dimension geom.XDimension2D
}

// wbsNode is upstream's StyleQueries.WBSDIAG_NODE with the SNames that
// follow node.
func wbsNode(names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement,
		klimt.SNameWbsDiagram, klimt.SNameNode}, names...)...)
}

// wbsArrow is upstream's StyleQueries.WBSDIAG_ARROW.
func wbsArrow() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameWbsDiagram, klimt.SNameArrow)
}

// defaultStyleDefinitionNode is the style query of the node at level, from
// its shape and its children, as mindmap's Idea does.
func (e *wElement) defaultStyleDefinitionNode(level int) klimt.StyleSignatureBasic {
	var query klimt.StyleSignatureBasic
	switch {
	case level == 0 && e.shape == mindmap.IdeaShapeNone:
		query = wbsNode(klimt.SNameRootNode, klimt.SNameBoxless)
	case level == 0:
		query = wbsNode(klimt.SNameRootNode)
	case e.shape == mindmap.IdeaShapeNone && e.isLeaf():
		query = wbsNode(klimt.SNameLeafNode, klimt.SNameBoxless)
	case e.isLeaf():
		query = wbsNode(klimt.SNameLeafNode)
	case e.shape == mindmap.IdeaShapeNone:
		query = wbsNode(klimt.SNameBoxless)
	default:
		query = wbsNode()
	}
	return query.AddStereotypeOf(e.stereotype).AddLevel(level)
}

// withBackColor returns the skin parameters with the background color of
// the node, when it has one.
func (e *wElement) withBackColor(skinParam klimt.ISkinParam) klimt.ISkinParam {
	if e.backColor == nil {
		return skinParam
	}
	return klimt.NewSkinParamColors(skinParam, color.EmptyColors().Add(color.TypeBack, e.backColor))
}

// style resolves the style of the node, cascading down the starred styles
// of its ancestors: a nearer ancestor's declaration beats a farther one's,
// each ancestor ranking one below the previous one.
func (e *wElement) style() *klimt.Style {
	ancestorRank := 0
	result := e.styleBuilder.MergedStyleSpecial(e.defaultStyleDefinitionNode(e.level), ancestorRank)
	for up := e.parent; up != nil; up = up.parent {
		ss := up.defaultStyleDefinitionNode(e.level).AddStar()
		ancestorRank--
		styleParent := e.styleBuilder.MergedStyleSpecial(ss, ancestorRank)
		result = result.MergeWith(styleParent, klimt.MergeOverwriteExistingValue)
	}
	return result
}

// newRootElement returns the root.
func newRootElement(backColor color.HColor, label *klimt.Display, stereotype *klimt.Stereotype,
	styleBuilder *klimt.StyleBuilder, shape mindmap.IdeaShape) *wElement {
	return newWElement(backColor, 0, label, stereotype, nil, shape, styleBuilder)
}

func newWElement(backColor color.HColor, level int, label *klimt.Display, stereotype *klimt.Stereotype,
	parent *wElement, shape mindmap.IdeaShape, styleBuilder *klimt.StyleBuilder) *wElement {
	return &wElement{label: label, backColor: backColor, level: level, parent: parent, shape: shape,
		styleBuilder: styleBuilder, stereotype: stereotype}
}

func (e *wElement) isLeaf() bool { return len(e.childrenLeft) == 0 && len(e.childrenRight) == 0 }

// createElement adds a child to the node. A child on the left side at level
// one is put first in the right children and also added to the left
// children, as upstream does.
func (e *wElement) createElement(backColor color.HColor, newLevel int, newLabel *klimt.Display,
	stereotype *klimt.Stereotype, direction klimt.Direction, shape mindmap.IdeaShape,
	styleBuilder *klimt.StyleBuilder) *wElement {
	result := newWElement(backColor, newLevel, newLabel, stereotype, e, shape, styleBuilder)
	if direction == klimt.DirectionLeft && newLevel == 1 {
		e.childrenRight = append([]*wElement{result}, e.childrenRight...)
	}
	if direction == klimt.DirectionLeft {
		e.childrenLeft = append(e.childrenLeft, result)
	} else {
		e.childrenRight = append(e.childrenRight, result)
	}
	return result
}

// String is upstream's toString: the label.
func (e *wElement) String() string { return e.label.String() }

// children returns the children on the left side, or on the right side for
// any other direction.
func (e *wElement) children(direction klimt.Direction) []*wElement {
	if direction == klimt.DirectionLeft {
		return e.childrenLeft
	}
	return e.childrenRight
}

// setGeometry records the position and the dimension of the drawn node.
func (e *wElement) setGeometry(position klimt.UTranslate, dimension geom.XDimension2D) {
	e.position = &position
	e.dimension = dimension
}
