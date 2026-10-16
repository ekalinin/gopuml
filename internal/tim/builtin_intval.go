// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/IntVal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// intVal is %intval(s).
type intVal struct{ simpleReturnFunction }

func (intVal) Signature() TFunctionSignature { return NewTFunctionSignature("%intval", 1) }

func (intVal) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (intVal) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	s := values[0].String()
	v, ok := javaParseInt(s)
	if !ok {
		return nil, newEaterException("Cannot convert "+s+" to integer.", location)
	}
	return TValueFromInt(v), nil
}
