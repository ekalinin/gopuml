// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/MindMap.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// MindMap is one mindmap of a diagram: its right (regular) and left
// (reverse) branches, which share the root.
type MindMap struct {
	klimt.TextBlockDefaults
	memo      klimt.TextBlockMemoized
	regular   *branch
	reverse   *branch
	skinParam klimt.ISkinParam
	// multiplier is the depth of the first node after the root, by which
	// the depth of every later node is divided.
	multiplier int
}

// NewMindMap returns an empty mindmap.
func NewMindMap(skinParam klimt.ISkinParam) *MindMap {
	return &MindMap{regular: &branch{}, reverse: &branch{}, skinParam: skinParam}
}

var _ klimt.TextBlock = (*MindMap)(nil)

// computeFinger builds the fingers of the branches once: the left one when
// it has nodes, the right one when it has nodes or the left one has none.
// The root is drawn by the right finger only.
func (m *MindMap) computeFinger() {
	if !m.reverse.hasFinger() && !m.regular.hasFinger() {
		if m.reverse.hasChildren() {
			m.reverse.initFinger(m.skinParam, false)
		}
		if !m.reverse.hasFinger() || m.regular.hasChildren() {
			m.regular.initFinger(m.skinParam, true)
		}
		if m.reverse.hasFinger() && m.regular.hasFinger() {
			m.reverse.doNotDrawFirstPhalanx()
		}
	}
}

// CalculateDimension implements klimt.TextBlock: upstream's
// calculateDimensionSlow, memoized.
func (m *MindMap) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return m.memo.Dimension(stringBounder, m.calculateDimensionSlow)
}

func (m *MindMap) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	m.computeFinger()
	y1 := m.regular.halfThickness(stringBounder)
	y2 := m.reverse.halfThickness(stringBounder)
	y := max(y1, y2)
	width := m.reverse.x12(stringBounder) + m.regular.x12(stringBounder)
	height := y + max(m.reverse.halfThickness(stringBounder), m.regular.halfThickness(stringBounder))
	if m.skinParam.Rankdir() == geom.RankdirTopToBottom {
		return geom.NewXDimension2D(height, width)
	}
	return geom.NewXDimension2D(width, height)
}

// DrawU implements klimt.TextBlock: both branches from the root, which is
// placed after the left branch.
func (m *MindMap) DrawU(ug klimt.UGraphic) {
	if !m.reverse.hasRoot() && !m.regular.hasRoot() {
		return
	}
	m.computeFinger()
	stringBounder := ug.StringBounder()
	y1 := m.regular.halfThickness(stringBounder)
	y2 := m.reverse.halfThickness(stringBounder)
	y := max(y1, y2)
	x := m.reverse.x12(stringBounder)
	if m.skinParam.Rankdir() == geom.RankdirTopToBottom {
		ug = ug.Apply(klimt.NewUTranslate(y, x))
	} else {
		ug = ug.Apply(klimt.NewUTranslate(x, y))
	}
	m.regular.DrawU(ug)
	m.reverse.DrawU(ug)
}

// addIdeaInternal adds a node: the root when there is none yet, otherwise a
// node of the right branch (direction true) or of the left one. The
// NoStyleAvailableException branch of upstream is not ported: upstream
// throws it when no skin file can be loaded, and the skins are bundled.
func (m *MindMap) addIdeaInternal(stereotype *klimt.Stereotype, backColor color.HColor, level int,
	label *klimt.Display, shape IdeaShape, direction bool) command.Result {
	if !m.reverse.hasRoot() && !m.regular.hasRoot() {
		level = 0
	}
	if level == 0 {
		m.regular.initRoot(m.skinParam.CurrentStyleBuilder(), backColor, label, shape, stereotype)
		m.reverse.initRoot(m.skinParam.CurrentStyleBuilder(), backColor, label, shape, stereotype)
		return command.Ok()
	}
	if m.multiplier == 0 {
		m.multiplier = level
	}
	if level%m.multiplier != 0 {
		return command.Error("Bad indentation")
	}
	level /= m.multiplier
	if !direction {
		return m.reverse.add(m.skinParam.CurrentStyleBuilder(), backColor, level, label, shape, stereotype)
	}
	return m.regular.add(m.skinParam.CurrentStyleBuilder(), backColor, level, label, shape, stereotype)
}

// isFull reports whether a node at level starts a new mindmap.
func (m *MindMap) isFull(level int) bool { return level == 0 && m.regular.hasRoot() }
