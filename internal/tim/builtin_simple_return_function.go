// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/SimpleReturnFunction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// simpleReturnFunction is the base of the builtins that return a value:
// embed it and implement Signature, CanCover and ExecuteReturnFunction.
type simpleReturnFunction struct{}

func (simpleReturnFunction) FunctionType() TFunctionType { return functionReturn }

func (simpleReturnFunction) ExecuteProcedureInternal(*TContext, TMemory, text.StringLocated, []*TValue, map[string]*TValue) error {
	throwRuntime("UnsupportedOperationException", "")
	return nil
}

func (simpleReturnFunction) IsUnquoted() bool { return false }
