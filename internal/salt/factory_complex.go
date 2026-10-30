// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/AbstractElementFactoryComplex.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryPyramid.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryScroll.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryBorder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryTree.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryTab.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryMenu.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// elementFactory makes an element from the next tokens, when it is ready
// for them.
type elementFactory interface {
	create() terminated[element]
	ready() bool
}

// complexFactory is a factory of an element that holds other elements.
type complexFactory interface {
	elementFactory
	addFactory(factory elementFactory)
}

// abstractElementFactoryComplex holds the factories of the elements inside
// the element of a complex factory, tried in the order they were added.
type abstractElementFactoryComplex struct {
	dataSource dataSource
	factories  []elementFactory
	dictionary *saltDictionary
}

func (f *abstractElementFactoryComplex) addFactory(factory elementFactory) {
	f.factories = append(f.factories, factory)
}

func (f *abstractElementFactoryComplex) nextElement() terminated[element] {
	for _, factory := range f.factories {
		if factory.ready() {
			return factory.create()
		}
	}
	uml.Throw(uml.IllegalStateException, f.dataSource.peek(0).element)
	return terminated[element]{}
}

// mustBeReady panics as the create methods of upstream, which throw an
// IllegalStateException when the factory is not ready.
func mustBeReady(f elementFactory) {
	if !f.ready() {
		uml.Throw(uml.IllegalStateException, "")
	}
}

// borderPositionPattern is the "[NSEW]=" of String.matches.
var borderPositionPattern = regexp.MustCompile(`^[NSEW]=$`)

// elementFactoryPyramid makes a group: "{", "{+", "{^", "{#", "{!" or "{-",
// when the next token is neither a position of a border layout nor a tree
// marker.
type elementFactoryPyramid struct{ abstractElementFactoryComplex }

func newElementFactoryPyramid(dataSource dataSource, dictionary *saltDictionary) *elementFactoryPyramid {
	return &elementFactoryPyramid{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryPyramid) create() terminated[element] {
	mustBeReady(f)
	tmp := f.dataSource.next()
	header := tmp.element
	var title *string
	strategy := drawNone
	if len(header) == 2 {
		strategy = tableStrategyFromChar(header[1])
	}
	if strategy == drawOutsideWithTitle && tmp.terminator == newCol {
		s := text.RemoveQuotesFormat(f.dataSource.next().element, `"`)
		title = &s
	}
	positionner := &positionner2{}
	for f.dataSource.peek(0).element != "}" {
		next := f.nextElement()
		if isStar(next.element) {
			positionner.mergeLeft(next.terminator)
		} else {
			positionner.add(next)
		}
	}
	next := f.dataSource.next()
	return terminated[element]{newElementPyramid(positionner, strategy, title, f.dictionary), next.terminator}
}

func isStar(e element) bool {
	t, ok := e.(*elementText)
	return ok && t.text == "*"
}

func (f *elementFactoryPyramid) ready() bool {
	switch f.dataSource.peek(0).element {
	case "{", "{+", "{^", "{#", "{!", "{-":
		text1 := f.dataSource.peek(1).element
		return !borderPositionPattern.MatchString(text1) && !isTreeMarker(text1)
	}
	return false
}

// elementFactoryScroll makes a scroll area: "{S", "{S-" or "{SI".
type elementFactoryScroll struct{ abstractElementFactoryComplex }

func newElementFactoryScroll(dataSource dataSource, dictionary *saltDictionary) *elementFactoryScroll {
	return &elementFactoryScroll{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryScroll) create() terminated[element] {
	mustBeReady(f)
	header := f.dataSource.next().element
	positionner := &positionner2{}
	for f.dataSource.peek(0).element != "}" {
		positionner.add(f.nextElement())
	}
	next := f.dataSource.next()
	return terminated[element]{newElementPyramidScrolled(positionner, f.dictionary, scrollStrategyFromDesc(header)),
		next.terminator}
}

func (f *elementFactoryScroll) ready() bool {
	switch f.dataSource.peek(0).element {
	case "{S", "{S-", "{SI":
		return true
	}
	return false
}

// elementFactoryBorder makes a border layout: "{", "{+", "{#", "{!" or "{-"
// followed by "N=", "S=", "E=" or "W=".
type elementFactoryBorder struct{ abstractElementFactoryComplex }

func newElementFactoryBorder(dataSource dataSource, dictionary *saltDictionary) *elementFactoryBorder {
	return &elementFactoryBorder{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryBorder) create() terminated[element] {
	mustBeReady(f)
	f.dataSource.next()
	result := newElementBorder()
	for f.dataSource.peek(0).element != "}" {
		pos := f.dataSource.next().element
		switch pos[0] {
		case 'N':
			result.north = f.nextElement().element
		case 'S':
			result.south = f.nextElement().element
		case 'E':
			result.east = f.nextElement().element
		case 'W':
			result.west = f.nextElement().element
		case 'C':
			result.center = f.nextElement().element
		default:
			uml.Throw(uml.IllegalStateException, "")
		}
	}
	next := f.dataSource.next()
	return terminated[element]{result, next.terminator}
}

func (f *elementFactoryBorder) ready() bool {
	switch f.dataSource.peek(0).element {
	case "{", "{+", "{#", "{!", "{-":
		return borderPositionPattern.MatchString(f.dataSource.peek(1).element)
	}
	return false
}

// elementFactoryTree makes a tree: "{" followed by a tree marker. The first
// cell of a line is the label of an entry, read as one raw token; the other
// cells are widgets.
type elementFactoryTree struct{ abstractElementFactoryComplex }

func newElementFactoryTree(dataSource dataSource, dictionary *saltDictionary) *elementFactoryTree {
	return &elementFactoryTree{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryTree) create() terminated[element] {
	mustBeReady(f)
	f.dataSource.next()
	textT := f.dataSource.next().element
	strategy := drawNone
	if len(textT) == 2 {
		strategy = tableStrategyFromChar(textT[1])
	}
	result := newElementTree(font.ByDefault(12), f.dictionary, strategy)
	takeMe := true
	// A nested group in the first cell must not end the tree.
	depth := 0
	for depth > 0 || f.dataSource.peek(0).element != "}" {
		if takeMe {
			t := f.dataSource.next()
			s := t.element
			if s == "}" {
				depth--
			} else if strings.HasPrefix(s, "{") {
				depth++
			}
			result.addEntry(s)
			takeMe = t.terminator == newLine
		} else {
			next := f.nextElement()
			result.addCellToEntry(next.element)
			takeMe = next.terminator == newLine
		}
	}
	next := f.dataSource.next()
	return terminated[element]{result, next.terminator}
}

func (f *elementFactoryTree) ready() bool {
	if f.dataSource.peek(0).element == "{" {
		return isTreeMarker(f.dataSource.peek(1).element)
	}
	return false
}

// isTreeMarker reports whether text1 is "T", or "T" followed by the
// character of a table strategy.
func isTreeMarker(text1 string) bool {
	if text1 == "T" {
		return true
	}
	return len(text1) == 2 && text1[0] == 'T' && tableStrategyFromChar(text1[1]) != noStrategy
}

// elementFactoryTab makes tabs: "{/".
type elementFactoryTab struct{ abstractElementFactoryComplex }

func newElementFactoryTab(dataSource dataSource, dictionary *saltDictionary) *elementFactoryTab {
	return &elementFactoryTab{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryTab) create() terminated[element] {
	mustBeReady(f)
	f.dataSource.next()
	result := newElementTabBar(font.ByDefault(12), f.dictionary)
	for f.dataSource.peek(0).element != "}" {
		t := f.dataSource.next()
		result.addTab(t.element)
		if t.terminator == newLine {
			result.vertical = true
		}
	}
	next := f.dataSource.next()
	return terminated[element]{result, next.terminator}
}

func (f *elementFactoryTab) ready() bool { return f.dataSource.peek(0).element == "{/" }

// elementFactoryMenu makes a menu bar: "{*". The cells of the first line
// are the entries; on each following line, the first cell names an entry
// and the others are its items.
type elementFactoryMenu struct{ abstractElementFactoryComplex }

func newElementFactoryMenu(dataSource dataSource, dictionary *saltDictionary) *elementFactoryMenu {
	return &elementFactoryMenu{abstractElementFactoryComplex{dataSource: dataSource, dictionary: dictionary}}
}

func (f *elementFactoryMenu) create() terminated[element] {
	mustBeReady(f)
	f.dataSource.next()
	result := newElementMenuBar(font.ByDefault(12), f.dictionary)
	var subentry *string
	for f.dataSource.peek(0).element != "}" {
		t := f.dataSource.next()
		s := t.element
		switch {
		case subentry == nil:
			result.addEntry(s)
		case *subentry == "":
			subentry = &s
		default:
			result.addSubEntry(*subentry, s)
		}
		if t.terminator == newLine {
			empty := ""
			subentry = &empty
		}
	}
	next := f.dataSource.next()
	return terminated[element]{result, next.terminator}
}

func (f *elementFactoryMenu) ready() bool { return f.dataSource.peek(0).element == "{*" }
