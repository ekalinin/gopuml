// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/InvokeProcedure.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// invokeProcedure is %invoke_procedure(name, args...).
type invokeProcedure struct{}

func (invokeProcedure) Signature() TFunctionSignature {
	return NewTFunctionSignature("%invoke_procedure", 1)
}

func (invokeProcedure) CanCover(nbArg int, _ []string) bool { return nbArg > 0 }

func (invokeProcedure) FunctionType() TFunctionType { return functionProcedure }

func (invokeProcedure) ExecuteProcedureInternal(context *TContext, memory TMemory, location text.StringLocated,
	args []*TValue, named map[string]*TValue) error {
	fname := args[0].String()
	sublist := args[1:]
	fn := context.getFunctionSmart(NewTFunctionSignature(fname, len(sublist)))
	if fn == nil {
		return newEaterException("Cannot find void function "+fname, location)
	}
	return fn.ExecuteProcedureInternal(context, memory, location, sublist, named)
}

func (invokeProcedure) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	throwRuntime("UnsupportedOperationException", "")
	return nil, nil
}

func (invokeProcedure) IsUnquoted() bool { return false }
