// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Upper.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// upper is %upper(s).
type upper struct{ simpleReturnFunction }

func (upper) Signature() TFunctionSignature { return NewTFunctionSignature("%upper", 1) }

func (upper) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (upper) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromString(javaToUpperCase(values[0].String())), nil
}
