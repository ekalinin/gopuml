// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/VariableExists.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// variableExists is %variable_exists(name).
type variableExists struct{ simpleReturnFunction }

func (variableExists) Signature() TFunctionSignature {
	return NewTFunctionSignature("%variable_exists", 1)
}

func (variableExists) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (variableExists) ExecuteReturnFunction(_ *TContext, memory TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromBoolean(memory.GetVariable(values[0].String()) != nil), nil
}
