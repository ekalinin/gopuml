// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/EvalBoolean.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/Truth.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// evalBoolean evaluates the expression of !ifdef: names combined with !, &
// (or &&), | (or ||) and parentheses
// (http://stackoverflow.com/questions/3422673/evaluating-a-math-expression-given-in-string-form).
type evalBoolean struct {
	str   []uint16
	pos   int
	ch    uint16
	truth func(name string) bool
}

func newEvalBoolean(str string, truth func(name string) bool) *evalBoolean {
	return &evalBoolean{str: toUTF16(str), pos: -1, truth: truth}
}

func (e *evalBoolean) nextChar() {
	e.pos++
	if e.pos < len(e.str) {
		e.ch = e.str[e.pos]
	} else {
		e.ch = 0
	}
}

func (e *evalBoolean) eat(charToEat uint16) bool {
	for e.ch == ' ' {
		e.nextChar()
	}
	if e.ch == charToEat {
		e.nextChar()
		return true
	}
	return false
}

func (e *evalBoolean) parseExpression() bool {
	x := e.parseTerm()
	for e.eat('|') {
		e.eat('|')
		x = e.parseTerm() || x
	}
	return x
}

func (e *evalBoolean) parseTerm() bool {
	x := e.parseFactor()
	for e.eat('&') {
		e.eat('&')
		x = e.parseFactor() && x
	}
	return x
}

func (e *evalBoolean) parseFactor() bool {
	if e.eat('!') {
		return !e.parseFactor()
	}
	startPos := e.pos
	var x bool
	switch {
	case e.eat('('):
		x = e.parseExpression()
		e.eat(')')
	case e.isIdentifier():
		for e.isIdentifier() {
			e.nextChar()
		}
		x = e.truth(fromUTF16(e.str[startPos:e.pos]))
	default:
		throwRuntime("IllegalArgumentException", "Unexpected: "+fromUTF16([]uint16{e.ch}))
	}
	return x
}

func (e *evalBoolean) isIdentifier() bool {
	return e.ch == '_' || e.ch == '$' || isJavaLetterOrDigit(e.ch)
}

func (e *evalBoolean) eval() bool {
	e.nextChar()
	x := e.parseExpression()
	if e.pos < len(e.str) {
		throwRuntime("IllegalArgumentException", "Unexpected: "+fromUTF16([]uint16{e.ch}))
	}
	return x
}
