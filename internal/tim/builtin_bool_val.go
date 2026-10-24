// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/BoolVal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// boolVal is %boolval(s): 1 for "true" and "1", 0 for "false" and "0",
// ignoring case.
type boolVal struct{ simpleReturnFunction }

func (boolVal) Signature() TFunctionSignature { return NewTFunctionSignature("%boolval", 1) }

func (boolVal) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (boolVal) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	s := javaToLowerCase(values[0].String())
	switch s {
	case "true", "1":
		return TValueFromBoolean(true), nil
	case "false", "0":
		return TValueFromBoolean(false), nil
	}
	return nil, newEaterException("Cannot convert "+s+" to boolean.", location)
}
