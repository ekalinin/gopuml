// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/MindMap.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// MindMap is one mindmap of a diagram: its right (regular) and left
// (reverse) branches, which share the root.
type MindMap struct {
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
