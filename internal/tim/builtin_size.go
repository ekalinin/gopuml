// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Size.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// size is %size(v): the length of a string in UTF-16 chars, the number of
// elements of a JSON array or members of a JSON object, otherwise 0.
type size struct{ simpleReturnFunction }

func (size) Signature() TFunctionSignature { return NewTFunctionSignature("%size", 1) }

func (size) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (size) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	value := values[0]
	if value.IsNumber() {
		return TValueFromInt(0), nil
	}
	if value.IsString() {
		return TValueFromInt(int32(javaLength(value.String()))), nil
	}
	json := value.ToJSON()
	if json.IsArray() {
		return TValueFromInt(int32(json.AsArray().Size())), nil
	}
	if json.IsObject() {
		return TValueFromInt(int32(json.AsObject().Size())), nil
	}
	return TValueFromInt(0), nil
}
