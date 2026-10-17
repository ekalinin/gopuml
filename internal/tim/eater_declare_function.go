// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterDeclareProcedure.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterDeclareReturnFunction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// eaterDeclare reads "![unquoted] [final] procedure|function $name(args)".
type eaterDeclare struct {
	Eater
	function  *TFunctionImpl
	location  text.StringLocated
	finalFlag bool
	keyword   string
}

// EaterDeclareProcedure is "!procedure $name(args)".
type EaterDeclareProcedure struct{ eaterDeclare }

// EaterDeclareReturnFunction is "!function $name(args)", with an optional
// one-line return.
type EaterDeclareReturnFunction struct{ eaterDeclare }

func newEaterDeclareProcedure(s text.StringLocated) *EaterDeclareProcedure {
	return &EaterDeclareProcedure{eaterDeclare{Eater: newEater(s.Trimmed()), location: s, keyword: "procedure"}}
}

func newEaterDeclareReturnFunction(s text.StringLocated) *EaterDeclareReturnFunction {
	return &EaterDeclareReturnFunction{eaterDeclare{Eater: newEater(s.Trimmed()), location: s, keyword: "function"}}
}

func (e *eaterDeclare) analyzeModifiers() (unquoted bool, err error) {
	e.skipSpaces()
	if err := e.checkAndEatString("!"); err != nil {
		return false, err
	}
	for e.peekUnquoted() || e.peekFinal() {
		if e.peekUnquoted() {
			if err := e.checkAndEatString("unquoted"); err != nil {
				return false, err
			}
			e.skipSpaces()
			unquoted = true
		} else if e.peekFinal() {
			if err := e.checkAndEatString("final"); err != nil {
				return false, err
			}
			e.skipSpaces()
			e.finalFlag = true
		}
	}
	if err := e.checkAndEatString(e.keyword); err != nil {
		return false, err
	}
	e.skipSpaces()
	return unquoted, nil
}

func (e *EaterDeclareProcedure) analyze(context *TContext, memory TMemory) error {
	unquoted, err := e.analyzeModifiers()
	if err != nil {
		return err
	}
	e.function, err = e.eatDeclareProcedure(context, memory, unquoted, e.location)
	return err
}

func (e *EaterDeclareReturnFunction) analyze(context *TContext, memory TMemory) error {
	unquoted, err := e.analyzeModifiers()
	if err != nil {
		return err
	}
	e.function, err = e.eatDeclareReturnFunctionWithOptionalReturn(context, memory, unquoted, e.location)
	return err
}

func (e *eaterDeclare) peekUnquoted() bool { return e.peekChar() == 'u' }

func (e *eaterDeclare) peekFinal() bool { return e.peekChar() == 'f' && e.peekCharN2() == 'i' }

func (e *eaterDeclare) getFunction() *TFunctionImpl { return e.function }

func (e *eaterDeclare) getFinalFlag() bool { return e.finalFlag }
