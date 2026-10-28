// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/mindmap"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// WBSDiagram is upstream's WBSDiagram: a parsed WBS diagram, a tree of
// elements under one root, and the links between aliased elements.
type WBSDiagram struct {
	uml.Base

	root *wElement
	last *wElement
	// first is the markers of the root, which getSmartLevel compares the
	// others with.
	first string
	codes map[string]*wElement
	links []*wbsLink
}

// NewWBSDiagram returns an empty WBS diagram for src; md5map holds the
// base64 images of the source.
func NewWBSDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *WBSDiagram {
	return &WBSDiagram{Base: uml.NewBase(ctx, src, core.WBS, md5map), codes: map[string]*wElement{}}
}

var _ core.TypedDiagram = (*WBSDiagram)(nil)

// Description implements core.Diagram: "Work Breakdown Structure".
func (d *WBSDiagram) Description() string { return "Work Breakdown Structure" }

// patternStereotype is a label that ends with a stereotype.
var patternStereotype = uml.Compile(`^\s*(.*?)\s*(\<\<\s*(.*)\s*\>\>)\s*$`)

// addIdeaLabel adds a node from its label, with the stereotype that ends
// the label removed from it.
func (d *WBSDiagram) addIdeaLabel(code string, backColor color.HColor, level int, label string,
	direction klimt.Direction, shape mindmap.IdeaShape) command.Result {
	var stereotype *klimt.Stereotype
	if m := patternStereotype.FindStringSubmatch(label); m != nil {
		label = m[1]
		stereotype = klimt.StereotypeBuild(m[2])
	}
	display := klimt.DisplayWithNewlines(d.Pragma(), label)
	return d.addIdea(code, backColor, level, display, stereotype, direction, shape)
}

// addIdea adds a node; code is its alias, or "" for none. The first node at
// level zero is the root.
func (d *WBSDiagram) addIdea(code string, backColor color.HColor, level int, display *klimt.Display,
	stereotype *klimt.Stereotype, direction klimt.Direction, shape mindmap.IdeaShape) command.Result {
	if level == 0 {
		if d.root != nil {
			return command.Error("WBS diagram only accepts one root first declared.")
		}
		d.initRoot(backColor, display, stereotype, shape)
		return command.Ok()
	}
	return d.add(code, backColor, level, display, stereotype, direction, shape)
}

func (d *WBSDiagram) initRoot(backColor color.HColor, display *klimt.Display, stereotype *klimt.Stereotype,
	shape mindmap.IdeaShape) {
	d.root = newRootElement(backColor, display, stereotype, d.SkinParam().CurrentStyleBuilder(), shape)
	d.last = d.root
}

// parentOfLast returns the ancestor nb levels above the last node.
func (d *WBSDiagram) parentOfLast(nb int) *wElement {
	result := d.last
	for range nb {
		result = result.parent
	}
	return result
}

// smartLevel is upstream's getSmartLevel: the level of the markers typ. The
// root is at level zero and records its markers, which the markers of the
// other nodes are read against. Markers it cannot read throw
// UnsupportedOperationException.
func (d *WBSDiagram) smartLevel(typ string) int {
	if d.root == nil {
		d.first = typ
		return 0
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

// add adds a node below the root: a child of the last node one level below
// it, or a child of the ancestor of the last node one level above it.
func (d *WBSDiagram) add(code string, backColor color.HColor, level int, display *klimt.Display,
	stereotype *klimt.Stereotype, direction klimt.Direction, shape mindmap.IdeaShape) command.Result {
	var parent *wElement
	switch {
	case level == d.last.level+1:
		parent = d.last
	case level <= d.last.level:
		parent = d.parentOfLast(d.last.level - level + 1)
	default:
		return command.Error("Bad tree structure")
	}
	newIdea := parent.createElement(backColor, level, display, stereotype, direction, shape,
		d.SkinParam().CurrentStyleBuilder())
	d.last = newIdea
	if code != "" {
		d.codes[code] = newIdea
	}
	return command.Ok()
}

// link adds a link from the node with the alias code1 to the node with the
// alias code2, in the line color of colors or, without one, of the arrow
// style with the stereotype.
func (d *WBSDiagram) link(code1, code2 string, colors *color.Colors, stereotype *klimt.Stereotype) command.Result {
	element1 := d.codes[code1]
	if element1 == nil {
		return command.Error("No such node " + code1)
	}
	element2 := d.codes[code2]
	if element2 == nil {
		return command.Error("No such node " + code2)
	}
	c := colors.Color(color.TypeLine)
	if c == nil {
		style := wbsArrow().AddStereotypeOf(stereotype).MergedStyle(d.SkinParam().CurrentStyleBuilder())
		c = style.Value(klimt.PNameLineColor).AsColor(d.SkinParam().IHtmlColorSet())
	}
	d.links = append(d.links, newWBSLink(element1, element2, c))
	return command.Ok()
}
