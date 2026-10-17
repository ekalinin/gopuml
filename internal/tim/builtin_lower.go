// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Lower.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// lower is %lower(s).
type lower struct{ simpleReturnFunction }

func (lower) Signature() TFunctionSignature { return NewTFunctionSignature("%lower", 1) }

func (lower) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (lower) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromString(javaToLowerCase(values[0].String())), nil
}
