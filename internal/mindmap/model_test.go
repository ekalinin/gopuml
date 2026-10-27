// SPDX-License-Identifier: MIT
// Provenance: original

package mindmap

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// parse parses the lines of a mindmap diagram; on an error it returns the
// message and the line of the error.
func parse(t *testing.T, lines ...string) (d *MindMapDiagram, errMsg, errLine string) {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@startmindmap\n" + strings.Join(lines, "\n") + "\n@endmindmap\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	if e, ok := result.(*core.ErrorDiagram); ok {
		line := ""
		if n := len(e.Trace); n > 0 {
			line = e.Trace[n-1].S
		}
		return nil, e.Error.Error, line
	}
	return result.(*MindMapDiagram), "", ""
}

// newDiagram returns an empty mindmap diagram.
func newDiagram() *MindMapDiagram {
	lines := []text.StringLocated{{S: "@startmindmap"}, {S: "@endmindmap"}}
	return NewMindMapDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), nil)
}

// mustParse is parse for lines without error.
func mustParse(t *testing.T, lines ...string) *MindMapDiagram {
	t.Helper()
	d, msg, line := parse(t, lines...)
	if d == nil {
		t.Fatalf("%q on %q", msg, line)
	}
	return d
}

// labelOf is the lines of the label of i, joined with "/".
func labelOf(i *idea) string {
	var lines []string
	for _, l := range i.label.AsList() {
		lines = append(lines, l.String())
	}
	return strings.Join(lines, "/")
}

// treeOf writes i and its children as "label(child,child)".
func treeOf(i *idea) string {
	s := labelOf(i)
	if len(i.children) > 0 {
		var children []string
		for _, c := range i.children {
			children = append(children, treeOf(c))
		}
		s += "(" + strings.Join(children, ",") + ")"
	}
	return s
}

// trees writes the mindmaps of d, each as "right tree | left tree", the
// mindmaps separated by " ; ".
func trees(d *MindMapDiagram) string {
	var result []string
	for _, m := range d.mindmaps {
		s := ""
		if m.regular.root != nil {
			s = treeOf(m.regular.root) + " | " + treeOf(m.reverse.root)
		}
		result = append(result, s)
	}
	return strings.Join(result, " ; ")
}

func TestIdeaShapeFromDesc(t *testing.T) {
	if IdeaShapeFromDesc("_") != IdeaShapeNone || IdeaShapeFromDesc("") != IdeaShapeBox ||
		IdeaShapeFromDesc("x") != IdeaShapeBox {
		t.Error("IdeaShapeFromDesc")
	}
}

func TestStyleQueries(t *testing.T) {
	// The SNames of a query are a set, listed in alphabetical order.
	names := func(s klimt.StyleSignatureBasic) string {
		var result []string
		for _, n := range s.SNames() {
			result = append(result, n.String())
		}
		return strings.Join(result, " ")
	}
	blue := klimt.StereotypeBuild("<<blue>>")
	root := createIdeaSimple(nil, nil, klimt.DisplayCreate("r"), IdeaShapeBox, nil)
	boxlessRoot := createIdeaSimple(nil, nil, klimt.DisplayCreate("r"), IdeaShapeNone, nil)
	node := root.createIdea(nil, nil, 1, klimt.DisplayCreate("a"), IdeaShapeBox, blue)
	boxless := root.createIdea(nil, nil, 1, klimt.DisplayCreate("b"), IdeaShapeNone, nil)
	node.createIdea(nil, nil, 2, klimt.DisplayCreate("c"), IdeaShapeBox, nil)
	boxless.createIdea(nil, nil, 2, klimt.DisplayCreate("d"), IdeaShapeBox, nil)
	leaf := root.createIdea(nil, nil, 1, klimt.DisplayCreate("e"), IdeaShapeBox, nil)
	boxlessLeaf := root.createIdea(nil, nil, 1, klimt.DisplayCreate("f"), IdeaShapeNone, nil)
	for _, tc := range []struct {
		name  string
		idea  *idea
		level int
		want  string
	}{
		{"root", root, 0, "element mindmapDiagram node root rootNode"},
		{"boxless root", boxlessRoot, 0, "boxless element mindmapDiagram node root rootNode"},
		{"node", node, 1, "element mindmapDiagram node root"},
		{"boxless node", boxless, 1, "boxless element mindmapDiagram node root"},
		{"leaf", leaf, 1, "element leafNode mindmapDiagram node root"},
		{"boxless leaf", boxlessLeaf, 1, "boxless element leafNode mindmapDiagram node root"},
		// The query of an ancestor at the level of a descendant.
		{"node at level 3", node, 3, "element mindmapDiagram node root"},
	} {
		q := tc.idea.defaultStyleDefinitionNode(tc.level)
		if got := names(q); got != tc.want {
			t.Errorf("%s: snames %q, want %q", tc.name, got, tc.want)
		}
		if q.Level() != tc.level || q.IsStarred() {
			t.Errorf("%s: level %d starred %v", tc.name, q.Level(), q.IsStarred())
		}
	}
	if got := node.defaultStyleDefinitionNode(1).Stereotypes(); len(got) != 1 || got[0] != "blue" {
		t.Errorf("stereotypes %q, want blue", got)
	}
}

// fontColorOf is the font color of the node of that label, without the
// space that upstream leaves before a removed stereotype.
func fontColorOf(d *MindMapDiagram, label string) string {
	var find func(i *idea) *idea
	find = func(i *idea) *idea {
		if strings.TrimSpace(labelOf(i)) == label {
			return i
		}
		for _, c := range i.children {
			if r := find(c); r != nil {
				return r
			}
		}
		return nil
	}
	i := find(d.mindmaps[0].regular.root)
	return i.style().Value(klimt.PNameFontColor).AsColor(d.SkinParam().IHtmlColorSet()).AsString()
}

func TestStyleCascade(t *testing.T) {
	// Scenario "Nearest ancestor": the starred style of the nearest
	// ancestor wins, and applies to the node itself.
	d := mustParse(t, "<style>", "mindmapDiagram {", "  .outer * {", "    node {", "      FontColor blue",
		"    }", "  }", "  .inner * {", "    node {", "      FontColor red", "    }", "  }", "}", "</style>",
		"* root", "** a <<outer>>", "*** b <<inner>>", "**** c", "*** e")
	for label, want := range map[string]string{"root": "#000000", "a": "#0000FF", "b": "#FF0000", "c": "#FF0000",
		"e": "#0000FF"} {
		if got := fontColorOf(d, label); got != want {
			t.Errorf("%s: font color %s, want %s", label, got, want)
		}
	}
}

func TestIndentation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"first node is the root", []string{"** root", "*** a"}, "root(a) | root"},
		{"multiplier", []string{"* root", "*** a", "***** b"}, "root(a(b)) | root"},
		{"siblings and ancestors", []string{"* r", "** a", "*** b", "** c", "*** d", "**** e", "** f"},
			"r(a(b),c(d(e)),f) | r"},
		{"left branch", []string{"* r", "** a", "left side", "** b", "*** c"}, "r(a) | r(b(c))"},
		{"new mindmap", []string{"* a", "** b", "* c", "** d"}, "a(b) | a ; c(d) | c"},
	} {
		if got := trees(mustParse(t, tc.lines...)); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
	for _, tc := range []struct {
		lines         []string
		msg, lastLine string
	}{
		{[]string{"** root", "*** a", "**** b"}, "Bad indentation", "**** b"},
		{[]string{"* root", "** a", "**** b"}, "error42L", "**** b"},
	} {
		if _, msg, line := parse(t, tc.lines...); msg != tc.msg || line != tc.lastLine {
			t.Errorf("%q: error %q on %q, want %q on %q", tc.lines, msg, line, tc.msg, tc.lastLine)
		}
	}
	m := NewMindMap(newDiagram().SkinParam())
	if m.isFull(0) {
		t.Error("an empty mindmap is full")
	}
	m.addIdeaInternal(nil, nil, 0, klimt.DisplayCreate("r"), IdeaShapeBox, true)
	if !m.isFull(0) || m.isFull(1) {
		t.Error("isFull after the root")
	}
}

func TestSmartLevel(t *testing.T) {
	for _, tc := range []struct {
		typ  string
		want int
	}{
		{"*", 0}, {"**", 1}, {" *", 1}, {"  *", 2}, {"\t*", 1}, {" **", 1}, {"#", 0}, {"##", 1},
	} {
		d := newDiagram()
		d.smartLevel("*")
		if got := d.smartLevel(tc.typ); got != tc.want {
			t.Errorf("smartLevel(%q) = %d, want %d", tc.typ, got, tc.want)
		}
	}
	d := newDiagram()
	d.smartLevel("*")
	defer func() {
		re, _ := recover().(error)
		if re == nil || re.Error() != uml.UnsupportedOperationException+": type=<  ##>[*]" {
			t.Errorf("smartLevel(\"  ##\"): %v", re)
		}
	}()
	d.smartLevel("  ##")
}

func TestDiagram(t *testing.T) {
	d := newDiagram()
	if d.Description() != "MindMap" || d.DiagramType() != core.MindMap || d.SkinParam().Rankdir() != geom.RankdirLeftToRight {
		t.Errorf("description %q type %v rankdir %v", d.Description(), d.DiagramType(), d.SkinParam().Rankdir())
	}
	for _, tc := range []struct {
		direction klimt.Direction
		want      bool
	}{
		{klimt.DirectionRight, true}, {klimt.DirectionDown, true}, {klimt.DirectionLeft, false}, {klimt.DirectionUp, false},
	} {
		d.setDefaultDirection(tc.direction)
		if d.defaultDirection != tc.want {
			t.Errorf("direction %v: right side %v, want %v", tc.direction, d.defaultDirection, tc.want)
		}
	}
	// The ending stereotype is removed from the label, which keeps the
	// space before it, as upstream's does.
	d = newDiagram()
	d.addIdea(nil, 0, klimt.DisplayCreate("a <<blue>>"), IdeaShapeBox)
	root := d.mindmaps[0].regular.root
	if labels := root.stereotype.Labels(klimt.GuillemetNone()); labelOf(root) != "a " || len(labels) != 1 ||
		labels[0] != "blue" {
		t.Errorf("label %q stereotype %v", labelOf(root), root.stereotype)
	}
	// A new root starts a second mindmap.
	d.addIdea(nil, 0, klimt.DisplayCreate("b"), IdeaShapeBox)
	if len(d.mindmaps) != 2 {
		t.Errorf("%d mindmaps, want 2", len(d.mindmaps))
	}
}

func TestCommands(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"color and no box", []string{"* r", "**[#Orange]_ a"}, "r(a) | r"},
		{"brackets after a space", []string{"* r", "** [#Orange] a"}, "r([#Orange] a) | r"},
		{"indented markers", []string{"* r", " * a", "  * b", "\t* c"}, "r(a(b),c) | r"},
		{"plus and minus", []string{"+ r", "++ a", "-- b", "+- c"}, "r(a) | r(b,c)"},
		{"root command", []string{"0 Root", "** a"}, "Root(a) | Root"},
		{"multiline", []string{"* r", "**:first line", "second line; <<blue>>"}, "r(first line/second line) | r"},
		{"multiline hash", []string{"# r", "##:x", "y;"}, "r(x/y) | r"},
		{"multiline trailing space", []string{"* r", "**:a", "b; "}, "r(a/b;) | r"},
		{"line break", []string{"* r", `** a\nb`}, "r(a/b) | r"},
		{"sides", []string{"* r", "left side", "** a", "right side", "** b", "top side", "** c", "bottom side", "** d"},
			"r(b,d) | r(a,c)"},
	} {
		if got := trees(mustParse(t, tc.lines...)); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}

	d := mustParse(t, "* r", "**[#Orange]_ a")
	a := d.mindmaps[0].regular.root.children[0]
	if a.shape != IdeaShapeNone || a.backColor == nil || a.backColor.AsString() != "#FFA500" {
		t.Errorf("shape %v color %v", a.shape, a.backColor)
	}
	d = mustParse(t, "* r", "**:first line", "second line; <<blue>>")
	if s := d.mindmaps[0].regular.root.children[0].stereotype; s == nil || s.Labels(klimt.GuillemetNone())[0] != "blue" {
		t.Errorf("multiline stereotype %v", s)
	}
	if d := mustParse(t, "top to bottom direction", "* r"); d.SkinParam().Rankdir() != geom.RankdirTopToBottom {
		t.Errorf("rankdir %v", d.SkinParam().Rankdir())
	}
	if d := mustParse(t, "title Ideas", "* r"); d.Title().Display().String() != "[Ideas]" {
		t.Errorf("title %q", d.Title().Display().String())
	}

	for _, tc := range []struct {
		lines         []string
		msg, lastLine string
	}{
		{[]string{"* r", "**[#ZZZ] a"}, "No such color", "**[#ZZZ] a"},
		{[]string{"* r", "** a", "  ## x"},
			"You should send a mail to plantuml@gmail.com with this log (V1.2026.9) " +
				uml.UnsupportedOperationException + ": type=<  ##>[*]", "  ## x"},
		{[]string{"* r", "**:a", "b; <<st>> "},
			"You should send a mail to plantuml@gmail.com with this log (V1.2026.9) " + uml.NullPointerException +
				`: Cannot invoke "java.util.List.get(int)" because "lineLast" is null`, "b; <<st>> "},
	} {
		if _, msg, line := parse(t, tc.lines...); msg != tc.msg || line != tc.lastLine {
			t.Errorf("%q: error %q on %q, want %q on %q", tc.lines, msg, line, tc.msg, tc.lastLine)
		}
	}
}
