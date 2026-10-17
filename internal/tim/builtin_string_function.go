// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/StringFunction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// stringFunction is %string(v).
type stringFunction struct{ simpleReturnFunction }

func (stringFunction) Signature() TFunctionSignature { return NewTFunctionSignature("%string", 1) }

func (stringFunction) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (stringFunction) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromString(values[0].String()), nil
}
