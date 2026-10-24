// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/SetVariableValue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// setVariableValue is %set_variable_value(name, value): it sets a global
// variable and returns "".
type setVariableValue struct{ simpleReturnFunction }

func (setVariableValue) Signature() TFunctionSignature {
	return NewTFunctionSignature("%set_variable_value", 2)
}

func (setVariableValue) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (setVariableValue) ExecuteReturnFunction(_ *TContext, memory TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	if err := memory.PutVariable(values[0].String(), values[1], scopeGlobal, location); err != nil {
		return nil, err
	}
	return TValueFromString(""), nil
}
