// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterIf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterElseIf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterIfdef.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterIfndef.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterWhile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterIf is "!if expression".
type EaterIf struct {
	Eater
	booleanValue bool
}

func newEaterIf(sl text.StringLocated) *EaterIf { return &EaterIf{Eater: newEater(sl)} }

func (e *EaterIf) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!if"); err != nil {
		return err
	}
	e.skipSpaces()
	value, err := e.eatExpression(context, memory)
	if err != nil {
		return err
	}
	e.booleanValue = value.ToBoolean()
	return nil
}

func (e *EaterIf) isTrue() bool { return e.booleanValue }

// EaterElseIf is "!elseif expression".
type EaterElseIf struct {
	Eater
	booleanValue bool
}

func newEaterElseIf(sl text.StringLocated) *EaterElseIf { return &EaterElseIf{Eater: newEater(sl)} }

func (e *EaterElseIf) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!elseif"); err != nil {
		return err
	}
	e.skipSpaces()
	value, err := e.eatExpression(context, memory)
	if err != nil {
		return err
	}
	e.booleanValue = value.ToBoolean()
	return nil
}

func (e *EaterElseIf) isTrue() bool { return e.booleanValue }

// EaterIfdef is "!ifdef expression": a boolean expression of names that
// are true when a variable or a function has the name.
type EaterIfdef struct {
	Eater
	expression string
}

func newEaterIfdef(sl text.StringLocated) *EaterIfdef { return &EaterIfdef{Eater: newEater(sl)} }

func (e *EaterIfdef) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!ifdef"); err != nil {
		return err
	}
	e.skipSpaces()
	e.expression = e.eatAllToEnd()
	return nil
}

// isTrue evaluates the expression; a syntax error throws
// IllegalArgumentException, as upstream does.
func (e *EaterIfdef) isTrue(context *TContext, memory TMemory) bool {
	eval := newEvalBoolean(e.expression, func(varname string) bool {
		return memory.GetVariable(varname) != nil || context.doesFunctionExist(varname)
	})
	return eval.eval()
}

// EaterIfndef is "!ifndef name".
type EaterIfndef struct {
	Eater
	varname string
}

func newEaterIfndef(sl text.StringLocated) *EaterIfndef { return &EaterIfndef{Eater: newEater(sl)} }

func (e *EaterIfndef) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!ifndef"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.varname, err = e.eatAndGetVarname()
	return err
}

func (e *EaterIfndef) isTrue(context *TContext, memory TMemory) bool {
	return memory.GetVariable(e.varname) == nil && !context.doesFunctionExist(e.varname)
}

// EaterWhile is "!while expression".
type EaterWhile struct {
	Eater
	expression *TokenStack
}

func newEaterWhile(sl text.StringLocated) *EaterWhile { return &EaterWhile{Eater: newEater(sl)} }

func (e *EaterWhile) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!while"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.expression, err = e.eatTokenStack()
	return err
}

func (e *EaterWhile) getWhileExpression() *TokenStack { return e.expression }
