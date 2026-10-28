// SPDX-License-Identifier: MIT
// Provenance: original

package wbs

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/mindmap"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commonFactory parses the common commands only, so that the model tests
// can read a <style> block and add the nodes themselves.
var commonFactory = uml.NewFactory(core.WBS, func() []command.Command[*WBSDiagram] {
	return uml.AddCommonCommands1([]command.Command[*WBSDiagram](nil))
}, NewWBSDiagram)

// styled returns an empty WBS diagram with the common commands of lines.
// Without lines, the diagram is built directly: an empty block is upstream's
// empty diagram.
func styled(t *testing.T, lines ...string) *WBSDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	if len(lines) == 0 {
		src := []text.StringLocated{{S: "@startwbs"}, {S: "@endwbs"}}
		return NewWBSDiagram(rc, core.NewUmlSource(src, src, true), nil)
	}
	src := "@startwbs\n" + strings.Join(lines, "\n") + "\n@endwbs\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	d, ok := commonFactory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true)).(*WBSDiagram)
	if !ok {
		t.Fatalf("%q is not a WBS diagram", lines)
	}
	return d
}

// addNode adds a node with the markers typ and its label, on the left side
// when typ contains "-", as a node command does; it returns the error
// message, or "".
func addNode(d *WBSDiagram, code, typ, label string) string {
	direction := klimt.DirectionRight
	if strings.Contains(typ, "-") {
		direction = klimt.DirectionLeft
	}
	return d.addIdeaLabel(code, nil, d.smartLevel(typ), label, direction, mindmap.IdeaShapeBox).Err()
}

// tree adds the nodes "typ label" of lines; it returns the error message of
// the first node that fails, or "".
func tree(d *WBSDiagram, lines ...string) string {
	for _, l := range lines {
		typ, label, _ := strings.Cut(l, " ")
		if msg := addNode(d, "", typ, label); msg != "" {
			return msg
		}
	}
	return ""
}

// labelOf is the lines of the label of e, joined with "/".
func labelOf(e *wElement) string {
	var lines []string
	for _, l := range e.label.AsList() {
		lines = append(lines, l.String())
	}
	return strings.Join(lines, "/")
}

// treeOf writes e and its children as "label(right,right)[left,left]".
func treeOf(e *wElement) string {
	s := labelOf(e)
	join := func(children []*wElement) string {
		var result []string
		for _, c := range children {
			result = append(result, treeOf(c))
		}
		return strings.Join(result, ",")
	}
	if len(e.childrenRight) > 0 {
		s += "(" + join(e.childrenRight) + ")"
	}
	if len(e.childrenLeft) > 0 {
		s += "[" + join(e.childrenLeft) + "]"
	}
	return s
}

func TestDiagram(t *testing.T) {
	d := styled(t)
	if d.Description() != "Work Breakdown Structure" || d.DiagramType() != core.WBS {
		t.Errorf("description %q type %v", d.Description(), d.DiagramType())
	}
}

func TestSmartLevel(t *testing.T) {
	for _, tc := range []struct {
		first, typ string
		want       int
	}{
		{"*", "**", 1}, {"*", " *", 1}, {"*", "  *", 2}, {"*", "\t*", 1}, {"*", "  +*", 3}, {"**", "**", 1},
	} {
		d := styled(t)
		if got := d.smartLevel(tc.first); got != 0 {
			t.Errorf("first %q: level %d, want 0", tc.first, got)
		}
		d.initRoot(nil, klimt.DisplayCreate("root"), nil, mindmap.IdeaShapeBox)
		if got := d.smartLevel(tc.typ); got != tc.want {
			t.Errorf("smartLevel(%q) after %q = %d, want %d", tc.typ, tc.first, got, tc.want)
		}
	}
	d := styled(t)
	d.smartLevel("*")
	d.initRoot(nil, klimt.DisplayCreate("root"), nil, mindmap.IdeaShapeBox)
	defer func() {
		re, _ := recover().(error)
		if re == nil || re.Error() != uml.UnsupportedOperationException+": type=<  ++>[*]" {
			t.Errorf("smartLevel(\"  ++\"): %v", re)
		}
	}()
	d.smartLevel("  ++")
}

func TestCreateElement(t *testing.T) {
	root := newRootElement(nil, klimt.DisplayCreate("r"), nil, nil, mindmap.IdeaShapeBox)
	a := root.createElement(nil, 1, klimt.DisplayCreate("a"), nil, klimt.DirectionRight, mindmap.IdeaShapeBox, nil)
	// A left node at level one goes first in the right children and into
	// the left children.
	root.createElement(nil, 1, klimt.DisplayCreate("b"), nil, klimt.DirectionLeft, mindmap.IdeaShapeBox, nil)
	// A left node at level two goes into the left children only.
	a.createElement(nil, 2, klimt.DisplayCreate("c"), nil, klimt.DirectionLeft, mindmap.IdeaShapeBox, nil)
	if got := treeOf(root); got != "r(b,a[c])[b]" {
		t.Errorf("tree %q", got)
	}
	if root.isLeaf() || a.isLeaf() || !a.childrenLeft[0].isLeaf() {
		t.Error("isLeaf")
	}
	if len(root.children(klimt.DirectionLeft)) != 1 || len(root.children(klimt.DirectionDown)) != 2 {
		t.Error("children")
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
	right := klimt.DirectionRight
	box, none := mindmap.IdeaShapeBox, mindmap.IdeaShapeNone
	root := newRootElement(nil, klimt.DisplayCreate("r"), nil, nil, box)
	boxlessRoot := newRootElement(nil, klimt.DisplayCreate("r"), nil, nil, none)
	node := root.createElement(nil, 1, klimt.DisplayCreate("a"), blue, right, box, nil)
	boxless := root.createElement(nil, 1, klimt.DisplayCreate("b"), nil, right, none, nil)
	node.createElement(nil, 2, klimt.DisplayCreate("c"), nil, right, box, nil)
	boxless.createElement(nil, 2, klimt.DisplayCreate("d"), nil, right, box, nil)
	leaf := root.createElement(nil, 1, klimt.DisplayCreate("e"), nil, right, box, nil)
	boxlessLeaf := root.createElement(nil, 1, klimt.DisplayCreate("f"), nil, right, none, nil)
	for _, tc := range []struct {
		name  string
		e     *wElement
		level int
		want  string
	}{
		{"root", root, 0, "element node root rootNode wbsDiagram"},
		{"boxless root", boxlessRoot, 0, "boxless element node root rootNode wbsDiagram"},
		{"node", node, 1, "element node root wbsDiagram"},
		{"boxless node", boxless, 1, "boxless element node root wbsDiagram"},
		{"leaf", leaf, 1, "element leafNode node root wbsDiagram"},
		{"boxless leaf", boxlessLeaf, 1, "boxless element leafNode node root wbsDiagram"},
		// The query of an ancestor at the level of a descendant.
		{"node at level 3", node, 3, "element node root wbsDiagram"},
	} {
		q := tc.e.defaultStyleDefinitionNode(tc.level)
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

func TestStyleCascade(t *testing.T) {
	// Scenario "Nearest ancestor": the starred style of the nearest
	// ancestor wins, and applies to the node itself.
	d := styled(t, "<style>", "wbsDiagram {", "  .outer * {", "    node {", "      FontColor blue", "    }", "  }",
		"  .inner * {", "    node {", "      FontColor red", "    }", "  }", "}", "</style>")
	if msg := tree(d, "* root", "** a <<outer>>", "*** b <<inner>>", "**** c", "*** e"); msg != "" {
		t.Fatal(msg)
	}
	var find func(e *wElement, label string) *wElement
	find = func(e *wElement, label string) *wElement {
		if labelOf(e) == label {
			return e
		}
		for _, c := range e.childrenRight {
			if r := find(c, label); r != nil {
				return r
			}
		}
		return nil
	}
	for label, want := range map[string]string{"root": "#000000", "a": "#0000FF", "b": "#FF0000", "c": "#FF0000",
		"e": "#0000FF"} {
		e := find(d.root, label)
		if e == nil {
			t.Fatalf("no node %q in %s", label, treeOf(d.root))
		}
		if got := e.style().Value(klimt.PNameFontColor).AsColor(d.SkinParam().IHtmlColorSet()).AsString(); got != want {
			t.Errorf("%s: font color %s, want %s", label, got, want)
		}
	}
}

func TestTree(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"siblings and ancestors", []string{"* r", "** a", "*** b", "** c", "*** d", "**** e", "** f"},
			"r(a(b),c(d(e)),f)"},
		{"root with several markers", []string{"** root", "** a"}, "root(a)"},
		{"left node", []string{"+ r", "++ a", "+++ b", "++- c"}, "r(a(b)[c])"},
	} {
		d := styled(t)
		if msg := tree(d, tc.lines...); msg != "" {
			t.Fatalf("%s: %s", tc.name, msg)
		}
		if got := treeOf(d.root); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
	for _, tc := range []struct {
		lines []string
		msg   string
	}{
		{[]string{"* root", "** a", "**** b"}, "Bad tree structure"},
		{[]string{"* a", "** b", "* c"}, "WBS diagram only accepts one root first declared."},
	} {
		if got := tree(styled(t), tc.lines...); got != tc.msg {
			t.Errorf("%q: error %q, want %q", tc.lines, got, tc.msg)
		}
	}

	// The stereotype that ends a label is removed from it.
	d := styled(t)
	tree(d, "* root")
	addNode(d, "x", "**", "a <<blue>>")
	a := d.codes["x"]
	if a == nil || labelOf(a) != "a" || a.stereotype == nil || strings.Join(a.stereotype.StyleNames(), ",") != "blue" {
		t.Errorf("node %+v", a)
	}
}

func TestLink(t *testing.T) {
	d := styled(t, "<style>", "wbsDiagram {", "  .hot {", "    LineColor red", "  }", "}", "</style>")
	tree(d, "* root")
	addNode(d, "x", "**", "a")
	addNode(d, "y", "**", "b")
	if msg := d.link("x", "z", color.EmptyColors(), nil).Err(); msg != "No such node z" {
		t.Errorf("unknown alias: %q", msg)
	}
	if msg := d.link("w", "y", color.EmptyColors(), nil).Err(); msg != "No such node w" {
		t.Errorf("unknown alias: %q", msg)
	}
	d.link("x", "y", color.EmptyColors(), nil)
	d.link("x", "y", color.EmptyColors(), klimt.StereotypeBuild("<<hot>>"))
	d.link("x", "y", color.EmptyColors().Add(color.TypeLine, color.Blue), nil)
	var got []string
	for _, l := range d.links {
		got = append(got, labelOf(l.element1)+">"+labelOf(l.element2)+" "+l.color.AsString())
	}
	if want := "a>b #181818|a>b #FF0000|a>b #0000FF"; strings.Join(got, "|") != want {
		t.Errorf("links %q, want %q", strings.Join(got, "|"), want)
	}
}
