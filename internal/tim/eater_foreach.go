// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterForeach.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// EaterForeach is "!foreach $item in json".
type EaterForeach struct {
	Eater
	varname   string
	jsonValue json.Value
}

func newEaterForeach(sl text.StringLocated) *EaterForeach { return &EaterForeach{Eater: newEater(sl)} }

func (e *EaterForeach) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!foreach"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	if e.varname, err = e.eatAndGetVarname(); err != nil {
		return err
	}
	e.skipSpaces()
	if err := e.checkAndEatString("in"); err != nil {
		return err
	}
	e.skipSpaces()
	value, err := e.eatExpression(context, memory)
	if err != nil {
		return err
	}
	e.jsonValue = value.ToJSON()
	return nil
}

func (e *EaterForeach) isSkip() bool {
	if e.jsonValue == nil {
		return true
	}
	return foreachSize(e.jsonValue) == 0
}

// foreachSize is EaterForeach.size: the number of elements or members.
func foreachSize(value json.Value) int {
	switch v := value.(type) {
	case *json.Array:
		return v.Size()
	case *json.Object:
		return v.Size()
	}
	throwRuntime("IllegalArgumentException", "")
	return 0
}

func (e *EaterForeach) getVarname() string { return e.varname }

func (e *EaterForeach) getJSONValue() json.Value { return e.jsonValue }
