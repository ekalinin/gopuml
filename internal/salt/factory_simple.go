// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryButton.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryCheckboxOn.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryCheckboxOff.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryRadioOn.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryRadioOff.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryDroplist.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryImage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryRetrieveFromDictonnary.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ElementFactoryTextField.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// simpleFactory is the state of the factories of the widgets.
type simpleFactory struct {
	dataSource dataSource
	dictionary *saltDictionary
}

// inner is s.substring(1, s.length() - 1): a token of one character throws,
// as Java's substring.
func inner(s string) string {
	if n := len(utf16.Encode([]rune(s))); n < 2 {
		uml.Throw(uml.StringIndexOutOfBoundsException, fmt.Sprintf("begin 1, end %d, length %d", n-1, n))
	}
	return s[1 : len(s)-1]
}

// elementFactoryButton makes a button: "[text]", but not a "[[...]]" link.
type elementFactoryButton struct{ simpleFactory }

func (f *elementFactoryButton) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementButton(inner(next.element), font.ByDefault(12), f.dictionary), next.terminator}
}

func (f *elementFactoryButton) ready() bool {
	s := f.dataSource.peek(0).element
	if strings.HasPrefix(s, "[[") {
		return false
	}
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

// extractedAfter is the label after the first sep, trimmed.
func extractedAfter(s string, sep byte) []string {
	x := strings.IndexByte(s, sep)
	return []string{text.Trin(s[x+1:])}
}

// elementFactoryCheckbox makes a checked ("[X]") or an unchecked ("[]" or
// "[ ]") checkbox.
type elementFactoryCheckbox struct {
	simpleFactory
	checked bool
}

func (f *elementFactoryCheckbox) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementRadioCheckbox(extractedAfter(next.element, ']'), font.ByDefault(12), false,
		f.checked, f.dictionary), next.terminator}
}

func (f *elementFactoryCheckbox) ready() bool {
	s := f.dataSource.peek(0).element
	if f.checked {
		return strings.HasPrefix(s, "[X]")
	}
	return strings.HasPrefix(s, "[]") || strings.HasPrefix(s, "[ ]")
}

// elementFactoryRadio makes a checked ("(X)") or an unchecked ("()" or
// "( )") radio button.
type elementFactoryRadio struct {
	simpleFactory
	checked bool
}

func (f *elementFactoryRadio) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementRadioCheckbox(extractedAfter(next.element, ')'), font.ByDefault(12), true,
		f.checked, f.dictionary), next.terminator}
}

func (f *elementFactoryRadio) ready() bool {
	s := f.dataSource.peek(0).element
	if f.checked {
		return strings.HasPrefix(s, "(X)")
	}
	return strings.HasPrefix(s, "()") || strings.HasPrefix(s, "( )")
}

// elementFactoryDroplist makes a droplist: "^text^".
type elementFactoryDroplist struct{ simpleFactory }

func (f *elementFactoryDroplist) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementDroplist(inner(next.element), font.ByDefault(12), f.dictionary),
		next.terminator}
}

func (f *elementFactoryDroplist) ready() bool {
	s := f.dataSource.peek(0).element
	return strings.HasPrefix(s, "^") && strings.HasSuffix(s, "^")
}

// The patterns of String.matches, anchored.
var (
	imageStartPattern     = regexp.MustCompile(`^<<\w+$`)
	imageReferencePattern = regexp.MustCompile(`^<<\w+>>$`)
)

// elementFactoryImage makes a pixel image: "<<" or "<<name", its rows, then
// ">>". A named image goes into the dictionary.
type elementFactoryImage struct{ simpleFactory }

func (f *elementFactoryImage) create() terminated[element] {
	mustBeReady(f)
	header := f.dataSource.next().element
	var name *string
	if len(header) > 2 {
		s := header[2:]
		name = &s
	}
	var img []string
	for f.dataSource.peek(0).element != ">>" {
		img = append(img, f.dataSource.next().element)
	}
	next := f.dataSource.next()
	e := &elementImage{img: img}
	if name != nil {
		f.dictionary.put(*name, e)
	}
	return terminated[element]{e, next.terminator}
}

func (f *elementFactoryImage) ready() bool {
	s := f.dataSource.peek(0).element
	return s == "<<" || imageStartPattern.MatchString(s)
}

// elementFactoryRetrieveFromDictionary draws an image again: "<<name>>".
type elementFactoryRetrieveFromDictionary struct{ simpleFactory }

func (f *elementFactoryRetrieveFromDictionary) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	name := next.element[2 : len(next.element)-2]
	return terminated[element]{f.dictionary.get(name), next.terminator}
}

func (f *elementFactoryRetrieveFromDictionary) ready() bool {
	return imageReferencePattern.MatchString(f.dataSource.peek(0).element)
}

// elementFactoryLine makes a separator: a token that starts and ends with
// "--", "==", "~~" or "..".
type elementFactoryLine struct{ simpleFactory }

func (f *elementFactoryLine) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{&elementLine{separator: next.element[0]}, next.terminator}
}

func (f *elementFactoryLine) ready() bool {
	s := f.dataSource.peek(0).element
	return isLine(s, '-') || isLine(s, '=') || isLine(s, '~') || isLine(s, '.')
}

func isLine(s string, c byte) bool {
	cc := string([]byte{c, c})
	return strings.HasPrefix(s, cc) && strings.HasSuffix(s, cc)
}

// elementFactoryText makes plain text, from any token that does not start a
// structure.
type elementFactoryText struct{ simpleFactory }

func (f *elementFactoryText) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementText([]string{next.element}, font.ByDefault(12), f.dictionary),
		next.terminator}
}

func (f *elementFactoryText) ready() bool {
	s := f.dataSource.peek(0).element
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "}") {
		return false
	}
	return len(text.Trin(s)) > 0
}

// elementFactoryTextField makes a text field: "\"text\"".
type elementFactoryTextField struct{ simpleFactory }

func (f *elementFactoryTextField) create() terminated[element] {
	mustBeReady(f)
	next := f.dataSource.next()
	return terminated[element]{newElementTextField(inner(next.element), font.ByDefault(12), f.dictionary),
		next.terminator}
}

func (f *elementFactoryTextField) ready() bool {
	s := f.dataSource.peek(0).element
	return strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)
}

// addSimpleFactory adds the factories of the widgets to a complex factory,
// in upstream's order.
func addSimpleFactory(cpx complexFactory, source dataSource, dictionary *saltDictionary,
	treeFactory *elementFactoryTree) {
	s := simpleFactory{dataSource: source, dictionary: dictionary}
	cpx.addFactory(newElementFactoryMenu(source, dictionary))
	cpx.addFactory(treeFactory)
	cpx.addFactory(newElementFactoryTab(source, dictionary))
	cpx.addFactory(&elementFactoryLine{s})
	cpx.addFactory(&elementFactoryTextField{s})
	cpx.addFactory(&elementFactoryCheckbox{s, true})
	cpx.addFactory(&elementFactoryCheckbox{s, false})
	cpx.addFactory(&elementFactoryButton{s})
	cpx.addFactory(&elementFactoryDroplist{s})
	cpx.addFactory(&elementFactoryRadio{s, true})
	cpx.addFactory(&elementFactoryRadio{s, false})
	cpx.addFactory(&elementFactoryImage{s})
	cpx.addFactory(&elementFactoryRetrieveFromDictionary{s})
	cpx.addFactory(&elementFactoryText{s})
}

// createElement is PSystemSalt.createElement: the element of the first
// complex factory ready for the tokens of data. A shared tree factory is
// both a widget of the others and a top-level candidate.
func createElement(data []string, dictionary *saltDictionary) element {
	source := newDataSourceImpl(data)
	treeFactory := newElementFactoryTree(source, dictionary)
	cpx := []complexFactory{
		newElementFactoryPyramid(source, dictionary),
		newElementFactoryScroll(source, dictionary),
		newElementFactoryBorder(source, dictionary),
		treeFactory,
	}
	for _, f := range cpx {
		addSimpleFactory(f, source, dictionary, treeFactory)
	}
	for _, f1 := range cpx {
		for _, f2 := range cpx {
			f1.addFactory(f2)
		}
	}
	for _, f := range cpx {
		if f.ready() {
			return f.create().element
		}
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return nil
}
