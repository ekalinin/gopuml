// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetVariableValue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// getVariableValue is %get_variable_value(name).
type getVariableValue struct{ simpleReturnFunction }

func (getVariableValue) Signature() TFunctionSignature {
	return NewTFunctionSignature("%get_variable_value", 1)
}

func (getVariableValue) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (getVariableValue) ExecuteReturnFunction(_ *TContext, memory TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	if variable := memory.GetVariable(values[0].String()); variable != nil {
		return variable, nil
	}
	return TValueFromString(""), nil
}
