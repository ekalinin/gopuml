// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/ExecutionContexts.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/ExecutionContextIf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/ExecutionContextWhile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/ExecutionContextForeach.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// executionContexts holds the open !if, !while and !foreach of a memory.
type executionContexts struct {
	allIfs      []*ExecutionContextIf
	allWhiles   []*ExecutionContextWhile
	allForeachs []*ExecutionContextForeach
}

func (e *executionContexts) addIf(value *ExecutionContextIf) { e.allIfs = append(e.allIfs, value) }
func (e *executionContexts) addWhile(value *ExecutionContextWhile) {
	e.allWhiles = append(e.allWhiles, value)
}
func (e *executionContexts) addForeach(value *ExecutionContextForeach) {
	e.allForeachs = append(e.allForeachs, value)
}

func peekLast[T any](s []*T) *T {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

func pollLast[T any](s *[]*T) *T {
	if len(*s) == 0 {
		return nil
	}
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

func (e *executionContexts) peekIf() *ExecutionContextIf           { return peekLast(e.allIfs) }
func (e *executionContexts) peekWhile() *ExecutionContextWhile     { return peekLast(e.allWhiles) }
func (e *executionContexts) peekForeach() *ExecutionContextForeach { return peekLast(e.allForeachs) }
func (e *executionContexts) pollIf() *ExecutionContextIf           { return pollLast(&e.allIfs) }
func (e *executionContexts) pollWhile() *ExecutionContextWhile     { return pollLast(&e.allWhiles) }
func (e *executionContexts) pollForeach() *ExecutionContextForeach { return pollLast(&e.allForeachs) }

func (e *executionContexts) areAllIfOk() bool {
	for _, conditionalContext := range e.allIfs {
		if !conditionalContext.conditionIsOkHere() {
			return false
		}
	}
	return true
}

// ExecutionContextIf is an open !if.
type ExecutionContextIf struct {
	isTrue      bool
	hasBeenBurn bool
}

func executionContextIfFromValue(isTrue bool) *ExecutionContextIf {
	return &ExecutionContextIf{isTrue: isTrue, hasBeenBurn: isTrue}
}

func (c *ExecutionContextIf) conditionIsOkHere() bool { return c.isTrue }

func (c *ExecutionContextIf) enteringElseIf() { c.isTrue = false }

func (c *ExecutionContextIf) nowInElse() { c.isTrue = !c.hasBeenBurn }

func (c *ExecutionContextIf) nowInSomeElseIf() {
	c.isTrue = true
	c.hasBeenBurn = true
}

// ExecutionContextWhile is an open !while.
type ExecutionContextWhile struct {
	whileExpression *TokenStack
	codePosition    codePosition
	skipMe          bool
}

func executionContextWhileFromValue(whileExpression *TokenStack, position codePosition) *ExecutionContextWhile {
	return &ExecutionContextWhile{whileExpression: whileExpression, codePosition: position}
}

func (c *ExecutionContextWhile) conditionValue(location text.StringLocated, context *TContext, memory TMemory) (*TValue, error) {
	return c.whileExpression.getResult(location, context, memory)
}

// ExecutionContextForeach is an open !foreach.
type ExecutionContextForeach struct {
	varname      string
	jsonValue    json.Value
	codePosition codePosition
	skipMe       bool
	currentIndex int
}

func executionContextForeachFromValue(varname string, jsonValue json.Value, position codePosition) *ExecutionContextForeach {
	return &ExecutionContextForeach{varname: varname, jsonValue: jsonValue, codePosition: position}
}

func (c *ExecutionContextForeach) currentValue() json.Value {
	switch v := c.jsonValue.(type) {
	case *json.Array:
		value, err := v.Get(c.currentIndex)
		if err != nil {
			panic(err)
		}
		return value
	case *json.Object:
		return json.ValueOfString(v.Names()[c.currentIndex])
	}
	throwRuntime("IllegalStateException", "")
	return nil
}

func (c *ExecutionContextForeach) inc() {
	c.currentIndex++
	if c.currentIndex >= foreachSize(c.jsonValue) {
		c.skipMe = true
	}
}
