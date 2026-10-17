// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Xargs.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// xargs is %xargs(): the data given after @startuml. Its signature has no
// argument but, as upstream, it only covers calls with one argument.
type xargs struct{ simpleReturnFunction }

func (xargs) Signature() TFunctionSignature { return NewTFunctionSignature("%xargs", 0) }

func (xargs) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (xargs) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, _ []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	if x, ok := context.getXargs(); ok {
		return TValueFromString(x), nil
	}
	return TValueFromString(""), nil
}
