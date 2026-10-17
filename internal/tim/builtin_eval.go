// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Eval.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// eval is %eval(expression): the int value of the expression.
type eval struct{ simpleReturnFunction }

func (eval) Signature() TFunctionSignature { return NewTFunctionSignature("%eval", 1) }

func (eval) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (eval) ExecuteReturnFunction(context *TContext, memory TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	eater := newStringEater(values[0].String())
	value, err := eater.eatExpression(context, memory)
	if err != nil {
		return nil, err
	}
	return TValueFromInt(value.ToInt()), nil
}
