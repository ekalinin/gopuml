// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/AlwaysTrue.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/AlwaysFalse.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// alwaysTrue is %true().
type alwaysTrue struct{ simpleReturnFunction }

func (alwaysTrue) Signature() TFunctionSignature { return NewTFunctionSignature("%true", 0) }

func (alwaysTrue) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (alwaysTrue) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	return TValueFromBoolean(true), nil
}

// alwaysFalse is %false().
type alwaysFalse struct{ simpleReturnFunction }

func (alwaysFalse) Signature() TFunctionSignature { return NewTFunctionSignature("%false", 0) }

func (alwaysFalse) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (alwaysFalse) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	return TValueFromBoolean(false), nil
}
