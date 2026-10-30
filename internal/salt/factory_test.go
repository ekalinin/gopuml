// SPDX-License-Identifier: MIT
// Provenance: original

package salt

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// readyFactories returns the names of the widget factories ready for the
// first token of lines, in upstream's order.
func readyFactories(lines ...string) []string {
	source := newDataSourceImpl(lines)
	d := testDictionary()
	s := simpleFactory{dataSource: source, dictionary: d}
	named := []struct {
		name string
		f    elementFactory
	}{
		{"menu", newElementFactoryMenu(source, d)},
		{"tree", newElementFactoryTree(source, d)},
		{"tab", newElementFactoryTab(source, d)},
		{"line", &elementFactoryLine{s}},
		{"text field", &elementFactoryTextField{s}},
		{"checkbox on", &elementFactoryCheckbox{s, true}},
		{"checkbox off", &elementFactoryCheckbox{s, false}},
		{"button", &elementFactoryButton{s}},
		{"droplist", &elementFactoryDroplist{s}},
		{"radio on", &elementFactoryRadio{s, true}},
		{"radio off", &elementFactoryRadio{s, false}},
		{"image", &elementFactoryImage{s}},
		{"reuse", &elementFactoryRetrieveFromDictionary{s}},
		{"text", &elementFactoryText{s}},
		{"group", newElementFactoryPyramid(source, d)},
		{"scroll", newElementFactoryScroll(source, d)},
		{"border", newElementFactoryBorder(source, d)},
	}
	var result []string
	for _, n := range named {
		if n.f.ready() {
			result = append(result, n.name)
		}
	}
	return result
}

// TestFactoryReady checks the first factory that upstream's order picks
// for each syntax.
func TestFactoryReady(t *testing.T) {
	tests := []struct {
		lines []string
		first string
	}{
		{[]string{"[OK]"}, "button"},
		// A link is text, not a button.
		{[]string{"[[http://x.com]]"}, "text"},
		// A checked box is tried before the button.
		{[]string{"[X] a"}, "checkbox on"},
		{[]string{"[] a"}, "checkbox off"},
		{[]string{"[ ] a"}, "checkbox off"},
		{[]string{"(X) a"}, "radio on"},
		{[]string{"() a"}, "radio off"},
		{[]string{`"text"`}, "text field"},
		{[]string{"^Pick^"}, "droplist"},
		{[]string{".."}, "line"},
		{[]string{"=="}, "line"},
		{[]string{"~~"}, "line"},
		{[]string{"--"}, "line"},
		{[]string{"<<smile"}, "image"},
		{[]string{"<<"}, "image"},
		{[]string{"<<smile>>"}, "reuse"},
		{[]string{"{* File"}, "menu"},
		{[]string{"{/ a"}, "tab"},
		{[]string{"{T", "+ a", "}"}, "tree"},
		{[]string{"{", "N={ [a] }", "}"}, "border"},
		{[]string{"{#", "a", "}"}, "group"},
		{[]string{"{S", "a", "}"}, "scroll"},
		{[]string{"plain"}, "text"},
	}
	for _, tt := range tests {
		got := readyFactories(tt.lines...)
		if len(got) == 0 || got[0] != tt.first {
			t.Errorf("%q: ready factories %v, want %s first", tt.lines, got, tt.first)
		}
	}
}

func TestTreeFactory(t *testing.T) {
	e := createElement([]string{"{T#", "+ World | \"Earth\"", "++ America | [Visit]", "}"}, testDictionary())
	tree, ok := e.(*elementTree)
	if !ok {
		t.Fatalf("element %T, want a tree", e)
	}
	if tree.strategy != drawAll || len(tree.entries) != 2 {
		t.Fatalf("strategy %d and %d entries, want all lines and 2 entries", tree.strategy, len(tree.entries))
	}
	if tree.entries[0].level != 1 || tree.entries[1].level != 2 {
		t.Errorf("levels %d and %d, want 1 and 2", tree.entries[0].level, tree.entries[1].level)
	}
	if _, ok := tree.entries[0].otherElements[0].(*elementTextField); !ok {
		t.Errorf("cell %T, want a text field", tree.entries[0].otherElements[0])
	}
	if _, ok := tree.entries[1].otherElements[0].(*elementButton); !ok {
		t.Errorf("cell %T, want a button", tree.entries[1].otherElements[0])
	}
}

func TestMenuFactory(t *testing.T) {
	e := createElement([]string{"{", "{* File | Edit", " Edit | Undo | - | Redo }", "}"}, testDictionary())
	group, ok := e.(*elementPyramid)
	if !ok {
		t.Fatalf("element %T, want a group", e)
	}
	menu, ok := group.positions[0].element.(*elementMenuBar)
	if !ok {
		t.Fatalf("element %T, want a menu bar", group.positions[0].element)
	}
	if len(menu.entries) != 2 || menu.entries[0].text != "File" || menu.entries[1].text != "Edit" {
		t.Errorf("entries %v, want File and Edit", menu.entries)
	}
	if len(menu.popups) != 1 || menu.popups[0].entry.text != "Edit" {
		t.Fatalf("lists %v, want the list of Edit", menu.popups)
	}
	var items []string
	for _, item := range menu.popups[0].popup.entries {
		items = append(items, item.text)
	}
	if len(items) != 3 || items[0] != "Undo" || items[1] != "-" || items[2] != "Redo" {
		t.Errorf("items %q, want Undo, - and Redo", items)
	}
}

// TestFactoryErrors checks the exceptions that upstream throws while it
// builds the element tree.
func TestFactoryErrors(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		class string
	}{
		{"no ready factory", []string{"hello"}, "java.lang.IllegalArgumentException"},
		{"unknown menu", []string{"{* File | Edit", " View | Zoom }"}, "java.lang.IllegalArgumentException"},
		{"unknown position", []string{"{", "N={ [a] }", "X={ [b] }", "}"}, "java.lang.IllegalStateException"},
		{"unknown image", []string{"{", "<<nothere>>", "}"}, "java.lang.NullPointerException"},
		{"star first", []string{"{", "* | a", "}"}, "java.lang.NullPointerException"},
		{"unclosed group", []string{"{", "[OK]"}, "java.lang.IndexOutOfBoundsException"},
	}
	for _, tt := range tests {
		func() {
			defer func() {
				e, ok := recover().(*color.RuntimeError)
				if !ok || e.Class != tt.class {
					t.Errorf("%s: recovered %v, want %s", tt.name, e, tt.class)
				}
			}()
			createElement(tt.lines, testDictionary())
		}()
	}
}
