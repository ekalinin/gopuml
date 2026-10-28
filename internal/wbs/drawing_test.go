// SPDX-License-Identifier: MIT
// Provenance: original

package wbs

import (
	"fmt"
	"testing"

	"github.com/ekalinin/gopuml/internal/activitydiagram3"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

var stringBounder = font.NewStringBounderSvg(nil)

func dimOf(tb klimt.TextBlock) string {
	dim := tb.CalculateDimension(stringBounder)
	return fmt.Sprintf("%.3fx%.3f", dim.Width(), dim.Height())
}

func TestDrawingDimensions(t *testing.T) {
	d := mustParse(t, "* root", "** a", "*** b", "***< c", "** d")
	f := d.drawingElement()
	if len(f.right) != 2 {
		t.Fatalf("%d children of the root", len(f.right))
	}
	a, ok := f.right[0].(*itfComposed)
	if !ok || len(a.left) != 1 || len(a.right) != 1 {
		t.Fatalf("a: %T", f.right[0])
	}
	leaf := a.right[0]
	leafDim := leaf.CalculateDimension(stringBounder)
	// A leaf is its box.
	if got, want := dimOf(leaf), dimOf(activitydiagram3.FtileBoxOldCreateWbs(nodeOf(d, "b").style(),
		d.SkinParam(), klimt.DisplayCreate("b"))); got != want {
		t.Errorf("leaf %s, want %s", got, want)
	}
	// A node with children: the left column, the vertical line in the
	// middle of the box and the right column, below the box.
	mainDim := a.main.CalculateDimension(stringBounder)
	leftDim := a.left[0].CalculateDimension(stringBounder)
	w := max(mainDim.Width()/2, 10+leftDim.Width()) + max(mainDim.Width()/2, 10+leafDim.Width())
	h := mainDim.Height() + max(a.marginBottom+leftDim.Height(), a.marginBottom+leafDim.Height())
	if got, want := dimOf(a), fmt.Sprintf("%.3fx%.3f", w, h); got != want {
		t.Errorf("node with children %s, want %s", got, want)
	}
	if got := a.t1(stringBounder).X; got != max(mainDim.Width()/2, 10+leftDim.Width()) {
		t.Errorf("t1 x %.3f", got)
	}
	// The root: its children side by side, 20 apart, 40 below it.
	rootDim := f.main.CalculateDimension(stringBounder)
	dDim := f.right[1].CalculateDimension(stringBounder)
	aDim := a.CalculateDimension(stringBounder)
	w = max(aDim.Width()+20+dDim.Width(), rootDim.Width())
	h = max(aDim.Height(), dDim.Height()) + rootDim.Height() + 40
	if got, want := dimOf(f), fmt.Sprintf("%.3fx%.3f", w, h); got != want {
		t.Errorf("root %s, want %s", got, want)
	}
}

func TestAutoWidth(t *testing.T) {
	d := mustParse(t, "<style>", "wbsDiagram {", "  node {", "    Width auto", "  }", "}", "</style>",
		"* root", "** a", "** longer node", "*** b", "*** much longer node")
	f := d.drawingElement()
	f.CalculateDimension(stringBounder)
	width := func(i itf) string { return fmt.Sprintf("%.3f", i.mainBoxWidth(stringBounder)) }
	if a, longer := width(f.right[0]), width(f.right[1]); a != longer {
		t.Errorf("a %s, longer node %s", a, longer)
	}
	c := f.right[1].(*itfComposed)
	if b, much := width(c.right[0]), width(c.right[1]); b != much {
		t.Errorf("b %s, much longer node %s", b, much)
	}
}

func TestPseudoNode(t *testing.T) {
	d := mustParse(t, "* root", "**_", "*** a", "*** b")
	p, ok := d.drawingElement().right[0].(*itfComposed)
	if !ok {
		t.Fatalf("pseudo node: %T", d.drawingElement().right[0])
	}
	if got := dimOf(p.main); got != "0.000x0.000" || p.marginBottom != 0 {
		t.Errorf("pseudo node: main %s, margin %v", got, p.marginBottom)
	}
}

func TestGeometry(t *testing.T) {
	d := mustParse(t, "* root", "**(x) a", "**(y) b", "x -> y")
	// The limit finder does not record the positions.
	d.drawingElement().DrawU(klimt.LimitFinderCreate(stringBounder, false))
	if d.codes["x"].position != nil {
		t.Error("position recorded by the limit finder")
	}
	// The SVG graphic does: a is left of b, on the same line, both below
	// the margins of the diagram.
	if _, err := svgOf(t, d); err != nil {
		t.Fatal(err)
	}
	a, b := d.codes["x"], d.codes["y"]
	if a.position == nil || b.position == nil {
		t.Fatal("no position recorded by the SVG graphic")
	}
	if !(a.position.Dx() >= 10 && a.position.Dx() < b.position.Dx() && a.position.Dy() == b.position.Dy()) {
		t.Errorf("positions a %v b %v", *a.position, *b.position)
	}
	if a.dimension == (geom.XDimension2D{}) {
		t.Error("no dimension recorded")
	}
}

func TestWithoutNodes(t *testing.T) {
	d := mustParse(t, "title Plan")
	if _, err := svgOf(t, d); err == nil {
		t.Error("a diagram without nodes draws")
	}
}

// nodeOf returns the node of d with that label.
func nodeOf(d *WBSDiagram, label string) *wElement {
	var find func(e *wElement) *wElement
	find = func(e *wElement) *wElement {
		if labelOf(e) == label {
			return e
		}
		for _, c := range append(append([]*wElement(nil), e.childrenRight...), e.childrenLeft...) {
			if r := find(c); r != nil {
				return r
			}
		}
		return nil
	}
	return find(d.root)
}
