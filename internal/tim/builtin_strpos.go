// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Strpos.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// strpos is %strpos(full, searched): the index in UTF-16 chars of the first
// occurrence, or -1.
type strpos struct{ simpleReturnFunction }

func (strpos) Signature() TFunctionSignature { return NewTFunctionSignature("%strpos", 2) }

func (strpos) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (strpos) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromInt(int32(javaIndexOf(values[0].String(), values[1].String()))), nil
}
