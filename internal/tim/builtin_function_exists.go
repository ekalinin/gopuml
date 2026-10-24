// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/FunctionExists.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// functionExists is %function_exists(name).
type functionExists struct{ simpleReturnFunction }

func (functionExists) Signature() TFunctionSignature {
	return NewTFunctionSignature("%function_exists", 1)
}

func (functionExists) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (functionExists) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromBoolean(context.doesFunctionExist(values[0].String())), nil
}
