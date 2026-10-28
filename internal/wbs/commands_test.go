// SPDX-License-Identifier: MIT
// Provenance: original

package wbs

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/mindmap"
)

// parse parses the lines of a WBS diagram with the WBS factory; on an error
// it returns the message and the line of the error.
func parse(t *testing.T, lines ...string) (d *WBSDiagram, errMsg, errLine string) {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@startwbs\n" + strings.Join(lines, "\n") + "\n@endwbs\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	if e, ok := result.(*core.ErrorDiagram); ok {
		line := ""
		if n := len(e.Trace); n > 0 {
			line = e.Trace[n-1].S
		}
		return nil, e.Error.Error, line
	}
	return result.(*WBSDiagram), "", ""
}

// mustParse is parse for lines without error.
func mustParse(t *testing.T, lines ...string) *WBSDiagram {
	t.Helper()
	d, msg, line := parse(t, lines...)
	if d == nil {
		t.Fatalf("%q on %q", msg, line)
	}
	return d
}

// lastNode is the node of the last node line.
func lastNode(d *WBSDiagram) *wElement { return d.last }

// warnings is the warnings of d, one per line.
func warnings(d *WBSDiagram) []string {
	var result []string
	for _, w := range d.Warnings() {
		result = append(result, w.AsSingleLine())
	}
	return result
}

// orderMessage is upstream's warning on the legacy order for prefix.
func orderMessage(prefix string) string {
	return "Shape (_) and direction (< >) must come before color and id: start the line with '" + prefix + "'"
}

func TestNodeCommands(t *testing.T) {
	// "_" then ">": no box, the right side and the label.
	d := mustParse(t, "* root", "** a", "***_> b")
	if b := lastNode(d); labelOf(b) != "b" || b.shape != mindmap.IdeaShapeNone || len(d.root.childrenRight[0].childrenRight) != 1 {
		t.Errorf("***_> b: %s, shape %v", treeOf(d.root), b.shape)
	}
	// ">" then "_" is not read, as upstream.
	if _, msg, line := parse(t, "* root", "** a", "***>_ b"); msg != "Syntax Error?" || line != "***>_ b" {
		t.Errorf("***>_ b: error %q on %q", msg, line)
	}
	// The color and the alias, in either order.
	for _, l := range []string{"**[#Orange](x) a", "**(x)[#Orange] a"} {
		d := mustParse(t, "* root", l)
		a := d.codes["x"]
		if a == nil || labelOf(a) != "a" || a.backColor == nil || a.backColor.AsString() != "#FFA500" {
			t.Errorf("%s: node %v", l, a)
		}
		if len(warnings(d)) != 0 {
			t.Errorf("%s: warnings %q", l, warnings(d))
		}
	}
	// A quoted label with an alias.
	d = mustParse(t, "* root", `** "First node" as f`)
	if f := d.codes["f"]; f == nil || labelOf(f) != "First node" {
		t.Errorf("quoted label: %s", treeOf(d.root))
	}
	if _, msg, line := parse(t, "* root", "**[#ZZZ] a"); msg != "No such color" || line != "**[#ZZZ] a" {
		t.Errorf("unknown color: %q on %q", msg, line)
	}
	// "_" without a label is a pseudo node; no label without "_" fails.
	d = mustParse(t, "* root", "**_", "*** a")
	if p := d.root.childrenRight[0]; p.shape != mindmap.IdeaShapePseudo || labelOf(p) != "" || treeOf(d.root) != "root((a))" {
		t.Errorf("pseudo node: %s, shape %v", treeOf(d.root), p.shape)
	}
	if _, msg, line := parse(t, "* root", "** a", "**"); msg != "Missing label for WBS node." || line != "**" {
		t.Errorf("missing label: %q on %q", msg, line)
	}
	// The spaces and tabs before the markers reach getSmartLevel.
	d = mustParse(t, "* root", " * a", "  * b", "\t* c")
	if got := treeOf(d.root); got != "root(a(b),c)" {
		t.Errorf("indentation: %q", got)
	}
	// The sides.
	d = mustParse(t, "+ root", "++ a", "+++ b", "++- c", "+++< d", "+++> e")
	if got := treeOf(d.root); got != "root(a(b,e)[c,d])" {
		t.Errorf("sides: %q", got)
	}
	// "\n" starts a new line.
	d = mustParse(t, "* root", `** a\nb`)
	if got := labelOf(lastNode(d)); got != "a/b" {
		t.Errorf(`a\nb: %q`, got)
	}
	d = mustParse(t, "title Plan", "* root")
	if got := d.TitleDisplay().String(); got != "[Plan]" {
		t.Errorf("title %q", got)
	}
}

func TestOldOrder(t *testing.T) {
	d := mustParse(t, "* root", "**[#red](x)_ a", "**(y) b", "x -> y")
	a := d.codes["x"]
	if a == nil || labelOf(a) != "a" || a.shape != mindmap.IdeaShapeNone || a.backColor.AsString() != "#FF0000" {
		t.Errorf("old order: node %v", a)
	}
	if got := warnings(d); len(got) != 1 || got[0] != orderMessage("**_[#red](x)") {
		t.Errorf("old order: warnings %q", got)
	}
	if len(d.links) != 1 {
		t.Errorf("old order: %d links", len(d.links))
	}
	// The alias of a quoted label is not in the warning.
	d = mustParse(t, "* root", `**[#red]> "Lbl" as b`)
	if got := warnings(d); len(got) != 1 || got[0] != orderMessage("**>[#red]") {
		t.Errorf("quoted old order: warnings %q", got)
	}
}

func TestMultilineCommands(t *testing.T) {
	d := mustParse(t, "* root", "**:first line", "second line; <<blue>>")
	a := lastNode(d)
	if labelOf(a) != "first line/second line" || a.stereotype == nil ||
		strings.Join(a.stereotype.StyleNames(), ",") != "blue" {
		t.Errorf("multiline: %q, stereotype %v", labelOf(a), a.stereotype)
	}
	if len(warnings(d)) != 0 {
		t.Errorf("multiline: warnings %q", warnings(d))
	}
	d = mustParse(t, "* root", "**[#red]>:first", "second;")
	if got := labelOf(lastNode(d)); got != "first/second" {
		t.Errorf("old multiline: %q", got)
	}
	if got := warnings(d); len(got) != 1 || got[0] != orderMessage("**>[#red]") {
		t.Errorf("old multiline: warnings %q", got)
	}
}

func TestLinkCommand(t *testing.T) {
	for _, tc := range []struct {
		line, want string
	}{
		{"x -> y", "a>b #181818"},
		{"x ..> y", "a>b #181818"},
		{"x -> y #red", "a>b #FF0000"},
		{"x -> y <<hot>>", "a>b #FF0000"},
		{"x -> y : hello", "a>b #181818"},
	} {
		d := mustParse(t, "<style>", "wbsDiagram {", "  .hot {", "    LineColor red", "  }", "}", "</style>",
			"* root", "**(x) a", "**(y) b", tc.line)
		if len(d.links) != 1 {
			t.Fatalf("%s: %d links", tc.line, len(d.links))
		}
		l := d.links[0]
		if got := labelOf(l.element1) + ">" + labelOf(l.element2) + " " + l.color.AsString(); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.line, got, tc.want)
		}
	}
	if _, msg, line := parse(t, "* root", "**(x) a", "x -> y"); msg != "No such node y" || line != "x -> y" {
		t.Errorf("unknown alias: %q on %q", msg, line)
	}
}
