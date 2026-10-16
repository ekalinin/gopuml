// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Strlen.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// strlen is %strlen(s): the length in UTF-16 chars.
type strlen struct{ simpleReturnFunction }

func (strlen) Signature() TFunctionSignature { return NewTFunctionSignature("%strlen", 1) }

func (strlen) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (strlen) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromInt(int32(javaLength(values[0].String()))), nil
}
