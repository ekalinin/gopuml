// SPDX-License-Identifier: MIT
// Provenance: original

package salt

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

func textOf(s string) element { return newElementText([]string{s}, font.ByDefault(12), testDictionary()) }

func dimString(d geom.XDimension2D) string { return fmt.Sprintf("[%v,%v]", d.Width(), d.Height()) }

// TestLayoutDimensions checks the preferred dimensions of groups, scroll
// areas, border layouts, trees, tabs and menus against the ones of
// upstream's elements, measured in the reference image.
func TestLayoutDimensions(t *testing.T) {
	f := font.ByDefault(12)
	d := testDictionary()

	var spanning positionner2
	spanning.add(terminated[element]{textOf("a"), newCol})
	spanning.add(terminated[element]{textOf("b"), newCol})
	spanning.add(terminated[element]{textOf("c"), newLine})
	spanning.add(terminated[element]{textOf("long text here"), newCol})
	spanning.mergeLeft(newCol)
	spanning.add(terminated[element]{textOf("x"), newLine})

	var titled positionner2
	titled.add(terminated[element]{textOf("Login"), newCol})
	titled.add(terminated[element]{newElementTextField("me", f, d), newLine})
	title := "My group"

	scrolled := func(s scrollStrategy) element {
		var p positionner2
		p.add(terminated[element]{textOf("Message"), newLine})
		return newElementPyramidScrolled(&p, d, s)
	}

	border := newElementBorder()
	border.north = newElementButton("Top", f, d)
	border.west = newElementButton("Left", f, d)
	border.center = newElementButton("Center", f, d)
	border.east = newElementButton("Right", f, d)
	border.south = newElementButton("Bottom", f, d)

	tree := newElementTree(f, d, drawNone)
	for _, s := range []string{"+ World", "++ America", "+++ Canada", "++ Europe"} {
		tree.addEntry(s)
	}
	treeTable := newElementTree(f, d, drawAll)
	treeTable.addEntry("+ World")
	treeTable.addCellToEntry(newElementTextField("Earth", f, d))
	treeTable.addEntry("++ America")
	treeTable.addCellToEntry(newElementButton("Visit", f, d))

	tabs := newElementTabBar(f, d)
	verticalTabs := newElementTabBar(f, d)
	verticalTabs.vertical = true
	for _, s := range []string{"<b>General", "Fullscreen", "Behavior"} {
		tabs.addTab(s)
		verticalTabs.addTab(s)
	}

	menuBar := newElementMenuBar(f, d)
	for _, s := range []string{"File", "Edit", "Source", "Refactor"} {
		menuBar.addEntry(s)
	}
	menuBar.addSubEntry("Refactor", "New")
	menuBar.addSubEntry("Refactor", "-")
	popup := newElementMenuPopup(f, d)
	for _, s := range []string{"New", "Open File", "-", "Close"} {
		popup.addEntry(s)
	}

	tests := []struct {
		name string
		e    element
		want string
	}{
		{"spanning", newElementPyramid(&spanning, drawAll, nil, d), "[95.828125,31.9375]"},
		{"titled", newElementPyramid(&titled, drawOutsideWithTitle, &title, d), "[61.65625,45.90625]"},
		{"scroll both", scrolled(scrollBoth), "[84.59375,45.96875]"},
		{"scroll vertical only", scrolled(scrollVerticalOnly), "[84.59375,15.96875]"},
		{"scroll horizontal only", scrolled(scrollHorizontalOnly), "[54.59375,45.96875]"},
		{"border", border, "[147,68.90625]"},
		{"tree", tree, "[77.662109375,55.875]"},
		{"tree table", treeTable, "[130.5,38.9375]"},
		{"tabs", tabs, "[212.4140625,13.96875]"},
		{"vertical tabs", verticalTabs, "[61.330078125,86.90625]"},
		{"menu bar", menuBar, "[176.546875,13.96875]"},
		{"menu popup", popup, "[56.818359375,46.90625]"},
	}
	for _, tt := range tests {
		if got := dimString(tt.e.preferredDimension(stringBounder, 0, 0)); got != tt.want {
			t.Errorf("%s: %s, want %s", tt.name, got, tt.want)
		}
	}
	if tree.entries[2].xDelta() != 30 {
		t.Errorf("the entry at level 3 is indented by %v, want 30", tree.entries[2].xDelta())
	}
}

// TestGridSegmentOrder checks that the lines of a group are drawn in the
// order of upstream's HashSet of segments.
func TestGridSegmentOrder(t *testing.T) {
	g := newGrid(make([]float64, 3), make([]float64, 3), drawAll, klimt.TextBlockUtilsEmpty(0, 0))
	for _, c := range []*cell{newCell(0, 0), newCell(0, 1), newCell(1, 0), newCell(1, 1)} {
		g.addCell(c)
	}
	var hashes []int32
	for _, s := range g.horizontals.all {
		hashes = append(hashes, s.hashCode())
	}
	var want []segment
	for _, i := range klimt.JavaHashOrder(hashes) {
		want = append(want, g.horizontals.all[i])
	}
	if got := g.horizontals.iterationOrder(); !slices.Equal(got, want) {
		t.Errorf("order %v, want %v", got, want)
	}
	// The outline comes first and the duplicates of the cells are dropped.
	if len(g.horizontals.all) != 6 || len(g.verticals.all) != 6 {
		t.Errorf("%d horizontal and %d vertical segments, want 6 and 6", len(g.horizontals.all), len(g.verticals.all))
	}
}

// TestGridStrategies checks the segments of each strategy for a 2 by 2
// group.
func TestGridStrategies(t *testing.T) {
	tests := []struct {
		strategy tableStrategy
		h, v     int
	}{
		{drawNone, 0, 0},
		{drawOutside, 4, 4},
		{drawOutsideWithTitle, 4, 4},
		{drawVertical, 0, 6},
		{drawHorizontal, 6, 0},
		{drawAll, 6, 6},
	}
	for _, tt := range tests {
		g := newGrid(make([]float64, 3), make([]float64, 3), tt.strategy, klimt.TextBlockUtilsEmpty(0, 0))
		for _, c := range []*cell{newCell(0, 0), newCell(0, 1), newCell(1, 0), newCell(1, 1)} {
			g.addCell(c)
		}
		if len(g.horizontals.all) != tt.h || len(g.verticals.all) != tt.v {
			t.Errorf("strategy %d: %d horizontal and %d vertical segments, want %d and %d", tt.strategy,
				len(g.horizontals.all), len(g.verticals.all), tt.h, tt.v)
		}
	}
}

// TestMenuBarPasses checks the lists of a menu bar in the order of their
// entries and the #BBB background that their entries keep after the pass 1.
func TestMenuBarPasses(t *testing.T) {
	f := font.ByDefault(12)
	m := newElementMenuBar(f, testDictionary())
	m.addEntry("File")
	m.addEntry("Edit")
	m.addSubEntry("Edit", "Undo")
	m.addSubEntry("File", "New")
	if len(m.popups) != 2 || m.popups[0].entry.text != "Edit" || m.popups[1].entry.text != "File" {
		t.Fatalf("lists %v, want Edit then File, in the order of their first item", m.popups)
	}
	if m.entries[0].background != nil {
		t.Error("an entry has a background before the pass 1")
	}
	ug := klimt.LimitFinderCreate(stringBounder, false)
	dim := geom.NewXDimension2D(100, 20)
	m.drawU(ug, 0, &dim)
	m.drawU(ug, 1, &dim)
	// Only the entries with items get the background.
	for _, e := range m.entries {
		if e.background == nil || e.background.ToRGB(color.IdentityMapper) != "#BBBBBB" {
			t.Errorf("entry %s has the background %v after the pass 1, want #BBBBBB", e.text, e.background)
		}
	}
}
