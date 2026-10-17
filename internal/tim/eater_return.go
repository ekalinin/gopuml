// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterReturn.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterStartsub.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterUndef.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/text"
)

// EaterReturn is "!return expression".
type EaterReturn struct {
	Eater
	value *TValue
}

func newEaterReturn(sl text.StringLocated) *EaterReturn { return &EaterReturn{Eater: newEater(sl)} }

func (e *EaterReturn) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!return"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.value, err = e.eatExpression(context, memory)
	return err
}

func (e *EaterReturn) getValue2() *TValue { return e.value }

// EaterStartsub is "!startsub name".
type EaterStartsub struct {
	Eater
	subname string
}

func newEaterStartsub(sl text.StringLocated) *EaterStartsub {
	return &EaterStartsub{Eater: newEater(sl)}
}

var subnamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (e *EaterStartsub) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!startsub"); err != nil {
		return err
	}
	e.skipSpaces()
	e.subname = e.eatAllToEnd()
	if !subnamePattern.MatchString(e.subname) {
		return newEaterException("Bad sub name", e.getStringLocated())
	}
	return nil
}

func (e *EaterStartsub) getSubname() string { return e.subname }

// EaterUndef is "!undef name".
type EaterUndef struct{ Eater }

func newEaterUndef(sl text.StringLocated) *EaterUndef { return &EaterUndef{newEater(sl.Trimmed())} }

func (e *EaterUndef) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!undef"); err != nil {
		return err
	}
	e.skipSpaces()
	varname, err := e.eatAndGetVarname()
	if err != nil {
		return err
	}
	memory.RemoveVariable(varname)
	return nil
}
