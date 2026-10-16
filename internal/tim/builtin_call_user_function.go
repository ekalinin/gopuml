// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/CallUserFunction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// callUserFunction is %call_user_func(name, args...).
type callUserFunction struct{ simpleReturnFunction }

func (callUserFunction) Signature() TFunctionSignature {
	return NewTFunctionSignature("%call_user_func", 1)
}

func (callUserFunction) CanCover(nbArg int, _ []string) bool { return nbArg > 0 }

func (callUserFunction) ExecuteReturnFunction(context *TContext, memory TMemory, location text.StringLocated,
	values []*TValue, named map[string]*TValue) (*TValue, error) {
	fname := values[0].String()
	args := values[1:]
	fn := context.getFunctionSmart(NewTFunctionSignature(fname, len(args)))
	if fn == nil {
		return nil, newEaterException("Cannot find void function "+fname, location)
	}
	return fn.ExecuteReturnFunction(context, memory, location, args, named)
}
