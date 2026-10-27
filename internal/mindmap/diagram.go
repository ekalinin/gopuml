// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/MindMapDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// MindMapDiagram is upstream's MindMapDiagram: a parsed mindmap diagram, with
// its mindmaps drawn one below the other.
type MindMapDiagram struct {
	uml.Base

	mindmaps []*MindMap
	// defaultDirection is the side of the next "*" and "#" nodes: true for
	// the right side.
	defaultDirection bool
	// first is the markers of the first org-mode node, which getSmartLevel
	// compares the others with; hasFirst is false until then.
	first    string
	hasFirst bool
}

// NewMindMapDiagram returns an empty mindmap diagram for src; md5map holds
// the base64 images of the source. Its rank direction is left to right, as
// upstream's constructor sets it.
func NewMindMapDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *MindMapDiagram {
	d := &MindMapDiagram{Base: uml.NewBase(ctx, src, core.MindMap, md5map), defaultDirection: true}
	d.Skin().SetRankdir(geom.RankdirLeftToRight)
	d.mindmaps = append(d.mindmaps, NewMindMap(d.SkinParam()))
	return d
}

var _ core.TypedDiagram = (*MindMapDiagram)(nil)

// Description implements core.Diagram: "MindMap".
func (d *MindMapDiagram) Description() string { return "MindMap" }

// setDefaultDirection sets the side of the next "*" and "#" nodes: right
// and down are the right side.
func (d *MindMapDiagram) setDefaultDirection(direction klimt.Direction) {
	d.defaultDirection = direction == klimt.DirectionRight || direction == klimt.DirectionDown
}

func (d *MindMapDiagram) last() *MindMap { return d.mindmaps[len(d.mindmaps)-1] }

// addIdea adds a node on the default side.
func (d *MindMapDiagram) addIdea(backColor color.HColor, level int, label *klimt.Display,
	shape IdeaShape) command.Result {
	return d.addIdeaDirection(backColor, level, label, shape, d.defaultDirection)
}

// addIdeaDirection adds a node on a side, with the stereotype that ends its
// label, removed from the label.
func (d *MindMapDiagram) addIdeaDirection(backColor color.HColor, level int, label *klimt.Display,
	shape IdeaShape, direction bool) command.Result {
	stereotype := label.EndingStereotype()
	if stereotype != nil {
		label = label.RemoveEndingStereotype()
	}
	if d.last().isFull(level) {
		d.mindmaps = append(d.mindmaps, NewMindMap(d.SkinParam()))
	}
	return d.last().addIdeaInternal(stereotype, backColor, level, label, shape, direction)
}

// addIdeaStereotype adds a node on the default side with a stereotype.
func (d *MindMapDiagram) addIdeaStereotype(stereotype *klimt.Stereotype, backColor color.HColor, level int,
	label *klimt.Display, shape IdeaShape) command.Result {
	if d.last().isFull(level) {
		d.mindmaps = append(d.mindmaps, NewMindMap(d.SkinParam()))
	}
	return d.last().addIdeaInternal(stereotype, backColor, level, label, shape, d.defaultDirection)
}

// smartLevel is upstream's getSmartLevel: the depth of the org-mode markers
// typ, read against the markers of the first node. Markers it cannot read
// throw UnsupportedOperationException.
func (d *MindMapDiagram) smartLevel(typ string) int {
	if !d.hasFirst {
		d.first, d.hasFirst = typ, true
	}
	if strings.HasSuffix(typ, "**") {
		typ = text.JavaTrim(strings.ReplaceAll(typ, "\t", " "))
	}
	typ = strings.ReplaceAll(typ, "\t", " ")
	if !strings.Contains(typ, " ") {
		return len(typ) - 1
	}
	if strings.HasSuffix(typ, d.first) {
		return len(typ) - len(d.first)
	}
	if len(text.JavaTrim(typ)) == 1 {
		return len(typ) - 1
	}
	if strings.HasPrefix(typ, d.first) {
		return len(typ) - len(d.first)
	}
	uml.Throw(uml.UnsupportedOperationException, "type=<"+typ+">["+d.first+"]")
	return 0
}
