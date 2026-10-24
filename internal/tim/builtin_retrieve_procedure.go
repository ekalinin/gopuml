// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/RetrieveProcedure.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// retrieveProcedure is %retrieve_procedure(name, args...): the output of
// the procedure, as a string.
type retrieveProcedure struct{ simpleReturnFunction }

func (retrieveProcedure) Signature() TFunctionSignature {
	return NewTFunctionSignature("%retrieve_procedure", 1)
}

func (retrieveProcedure) CanCover(nbArg int, _ []string) bool { return nbArg > 0 }

func (retrieveProcedure) ExecuteReturnFunction(context *TContext, memory TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	fname := values[0].String()
	args := values[1:]
	fn := context.getFunctionSmart(NewTFunctionSignature(fname, len(args)))
	if fn == nil {
		throwRuntime("NullPointerException", "")
	}
	n1 := len(context.resultList)
	if err := fn.ExecuteProcedureInternal(context, memory, location, args, nil); err != nil {
		return nil, err
	}
	return TValueFromString(context.extractFromResultList(n1)), nil
}
