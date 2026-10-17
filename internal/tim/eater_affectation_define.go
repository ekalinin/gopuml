// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterAffectationDefine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterAffectationDefine is the legacy "!define NAME value": a global
// variable holding the text.
type EaterAffectationDefine struct{ Eater }

func newEaterAffectationDefine(sl text.StringLocated) *EaterAffectationDefine {
	return &EaterAffectationDefine{newEater(sl.Trimmed())}
}

func (e *EaterAffectationDefine) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!define"); err != nil {
		return err
	}
	e.skipSpaces()
	varname, err := e.eatAndGetVarname()
	if err != nil {
		return err
	}
	e.skipSpaces()
	tmp := e.eatAllToEnd()
	tmp2, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(tmp, e.getLineLocation()))
	if err != nil {
		return err
	}
	if tmp2 == nil {
		throwRuntime("NullPointerException", "")
	}
	return memory.PutVariable(varname, TValueFromString(*tmp2), scopeGlobal, e.getStringLocated())
}
