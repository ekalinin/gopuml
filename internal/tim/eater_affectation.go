// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterAffectation.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterAffectation is "!$name = value", with an optional local or global
// modifier and the conditional form "?=".
type EaterAffectation struct{ Eater }

func newEaterAffectation(sl text.StringLocated) *EaterAffectation {
	return &EaterAffectation{newEater(sl.Trimmed())}
}

func (e *EaterAffectation) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!"); err != nil {
		return err
	}
	e.skipSpaces()
	varname, err := e.eatAndGetVarname()
	if err != nil {
		return err
	}
	scope := lazzyParse(varname)
	if scope != scopeNone {
		e.skipSpaces()
		if e.peekChar() == '?' || e.peekChar() == '=' {
			// The variable itself is "local" or "global", which is not a good
			// idea by the way
			scope = scopeNone
		} else if varname, err = e.eatAndGetVarname(); err != nil {
			return err
		}
	}
	e.skipSpaces()
	conditional := false
	if e.peekChar() == '?' {
		if err := e.checkAndEatChar('?'); err != nil {
			return err
		}
		conditional = true
	}
	if err := e.checkAndEatChar('='); err != nil {
		return err
	}
	if conditional && memory.GetVariable(varname) != nil {
		return nil
	}
	e.skipSpaces()
	value, err := e.eatExpression(context, memory)
	if err != nil {
		return err
	}
	return memory.PutVariable(varname, value, scope, e.getStringLocated())
}
