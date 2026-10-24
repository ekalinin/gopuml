// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIterator.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/AbstractCodeIterator.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodePosition.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// codeIterator walks the lines of a body. Each layer handles some
// directives (comments, subs, functions, conditions, loops, assignments)
// and lets the other lines through. peek returns nil at the end.
type codeIterator interface {
	peek() (*text.StringLocated, error)
	next() error
	getCodePosition() codePosition
	jumpToCodePosition(newPosition codePosition, location text.StringLocated) error
}

// codePosition is a position in the lines of a codeIteratorImpl.
type codePosition int

// abstractCodeIterator delegates next and the positions to its source.
type abstractCodeIterator struct {
	source codeIterator
}

func (a abstractCodeIterator) next() error { return a.source.next() }

func (a abstractCodeIterator) getCodePosition() codePosition { return a.source.getCodePosition() }

func (a abstractCodeIterator) jumpToCodePosition(newPosition codePosition, location text.StringLocated) error {
	return a.source.jumpToCodePosition(newPosition, location)
}

// codeIteratorImpl walks a list of lines.
type codeIteratorImpl struct {
	list      []text.StringLocated
	current   int
	countJump int
}

func newCodeIteratorImpl(list []text.StringLocated) *codeIteratorImpl {
	return &codeIteratorImpl{list: list}
}

func (it *codeIteratorImpl) peek() (*text.StringLocated, error) {
	if it.current == len(it.list) {
		return nil, nil
	}
	if it.current > len(it.list) {
		throwRuntime("IllegalStateException", "")
	}
	return &it.list[it.current], nil
}

func (it *codeIteratorImpl) next() error {
	if it.current >= len(it.list) {
		throwRuntime("IllegalStateException", "")
	}
	it.current++
	return nil
}

func (it *codeIteratorImpl) getCodePosition() codePosition { return codePosition(it.current) }

func (it *codeIteratorImpl) jumpToCodePosition(newPosition codePosition, location text.StringLocated) error {
	it.countJump++
	if it.countJump > 999 {
		return newEaterException("Infinite loop?", location)
	}
	it.current = int(newPosition)
	return nil
}
