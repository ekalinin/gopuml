// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterLegacyDefine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterLegacyDefineLong.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterLegacyDefine is the legacy macro "!define NAME(args) definition".
type EaterLegacyDefine struct {
	Eater
	function *TFunctionImpl
}

func newEaterLegacyDefine(sl text.StringLocated) *EaterLegacyDefine {
	return &EaterLegacyDefine{Eater: newEater(sl.Trimmed())}
}

func (e *EaterLegacyDefine) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!define"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.function, err = e.eatDeclareFunction(context, memory, true, e.getStringLocated(), false, functionLegacyDefine)
	if err != nil {
		return err
	}
	e.function.setLegacyDefinition(e.eatAllToEnd())
	return nil
}

func (e *EaterLegacyDefine) getFunction() TFunction { return e.function }

// EaterLegacyDefineLong is "!definelong NAME(args)", closed by
// "!enddefinelong".
type EaterLegacyDefineLong struct {
	Eater
	function *TFunctionImpl
}

func newEaterLegacyDefineLong(sl text.StringLocated) *EaterLegacyDefineLong {
	return &EaterLegacyDefineLong{Eater: newEater(sl.Trimmed())}
}

func (e *EaterLegacyDefineLong) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!definelong"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.function, err = e.eatDeclareFunction(context, memory, true, e.getStringLocated(), true, functionLegacyDefineLong)
	return err
}

func (e *EaterLegacyDefineLong) getFunction() *TFunctionImpl { return e.function }
