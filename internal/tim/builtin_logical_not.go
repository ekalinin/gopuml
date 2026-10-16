// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalNot.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// logicalNot is %not(b).
type logicalNot struct{ simpleReturnFunction }

func (logicalNot) Signature() TFunctionSignature { return NewTFunctionSignature("%not", 1) }

func (logicalNot) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (logicalNot) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromBoolean(!values[0].ToBoolean()), nil
}
