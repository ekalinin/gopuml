// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/YamlBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

import "strconv"

// yamlBuilder builds the Monomorph of the lines: a stack of the values that
// the lines go into, and of their indentations.
type yamlBuilder struct {
	indents []int
	nodes   []*Monomorph
}

func newYamlBuilder() *yamlBuilder { return &yamlBuilder{nodes: []*Monomorph{{}}} }

// adjustIndentation is called before a line is processed: it adjusts the
// depth of the stack to the indentation of the line.
func (b *yamlBuilder) adjustIndentation(indent int) {
	if len(b.indents) == 0 {
		b.indents = append(b.indents, indent)
		return
	}
	if indent > b.indents[len(b.indents)-1] {
		b.indents = append(b.indents, indent)
		return
	}
	for len(b.indents) > 0 && indent < b.indents[len(b.indents)-1] {
		b.indents = b.indents[:len(b.indents)-1]
		b.removeLast()
		if b.last().Type() == List {
			b.removeLast()
		}
	}
}

func (b *yamlBuilder) result() *Monomorph { return b.nodes[0] }

// last returns the top of the stack. Upstream's list throws an
// IndexOutOfBoundsException when the stack is empty.
func (b *yamlBuilder) last() *Monomorph {
	if len(b.nodes) == 0 {
		throw(IndexOutOfBoundsException, "Index -1 out of bounds for length 0")
	}
	return b.nodes[len(b.nodes)-1]
}

func (b *yamlBuilder) removeLast() {
	if len(b.nodes) == 0 {
		throw(IndexOutOfBoundsException, "Index -1 out of bounds for length 0")
	}
	b.nodes = b.nodes[:len(b.nodes)-1]
}

func (b *yamlBuilder) onListItemPlainDash() {
	if b.isArrayAlreadyThere() {
		b.removeLast()
	}
	newElement := &Monomorph{}
	b.last().addInList(newElement)
	b.nodes = append(b.nodes, newElement)
}

func (b *yamlBuilder) isArrayAlreadyThere() bool {
	if len(b.nodes) < 2 {
		return false
	}
	potentialList := b.nodes[len(b.nodes)-2]
	if potentialList.Type() != List {
		return false
	}
	return potentialList.ElementAt(potentialList.Size()-1) == b.nodes[len(b.nodes)-1]
}

func (b *yamlBuilder) onKeyAndValue(key, value string) { b.last().putInMap(key, scalar(value)) }

func (b *yamlBuilder) onKeyAndFlowSequence(key string, values []string) {
	b.last().putInMap(key, list(values))
}

func (b *yamlBuilder) onOnlyKey(key string) {
	newElement := &Monomorph{}
	b.last().putInMap(key, newElement)
	b.nodes = append(b.nodes, newElement)
}

func (b *yamlBuilder) onListItemOnlyKey(key string) {
	if b.isArrayAlreadyThere() {
		b.removeLast()
	}
	newElement := &Monomorph{}
	b.last().addInList(newElement)
	b.nodes = append(b.nodes, newElement)
	newElement2 := &Monomorph{}
	b.last().putInMap(key, newElement2)
	b.nodes = append(b.nodes, newElement2)
}

func (b *yamlBuilder) onListItemOnlyValue(value string) {
	if b.isArrayAlreadyThere() {
		b.removeLast()
	}
	b.last().addInList(scalar(value))
}

func (b *yamlBuilder) onListItemKeyAndValue(key, value string) {
	if b.isArrayAlreadyThere() {
		b.removeLast()
	}
	newElement := &Monomorph{}
	b.last().addInList(newElement)
	b.nodes = append(b.nodes, newElement)
	b.last().putInMap(key, scalar(value))
}

func (b *yamlBuilder) onListItemKeyAndFlowSequence(key string, values []string) {
	if b.isArrayAlreadyThere() {
		b.removeLast()
	}
	newElement := &Monomorph{}
	b.last().addInList(newElement)
	b.nodes = append(b.nodes, newElement)
	b.last().putInMap(key, list(values))
}

// String lists the depth of the stack, for the messages of the tests.
func (b *yamlBuilder) String() string { return "yamlBuilder(" + strconv.Itoa(len(b.nodes)) + " nodes)" }
