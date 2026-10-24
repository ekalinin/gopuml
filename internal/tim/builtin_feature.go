// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Feature.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// feature is %feature(name): 1 for the features "style" and "theme"
// (ignoring case), 0 otherwise.
type feature struct{ simpleReturnFunction }

func (feature) Signature() TFunctionSignature { return NewTFunctionSignature("%feature", 1) }

func (feature) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (feature) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	arg := values[0].String()
	if javaEqualsIgnoreCase("style", arg) || javaEqualsIgnoreCase("theme", arg) {
		return TValueFromInt(1), nil
	}
	return TValueFromInt(0), nil
}
