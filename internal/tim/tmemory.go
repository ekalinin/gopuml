// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TMemory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// TMemory holds the variables and the open control structures of the code
// being executed: the global memory, or the local memory of a function call.
type TMemory interface {
	GetVariable(varname string) *TValue
	PutVariable(varname string, value *TValue, scope TVariableScope, location text.StringLocated) error
	RemoveVariable(varname string)
	IsEmpty() bool
	variablesNames3() Trie
	forkFromGlobal(input map[string]*TValue) TMemory
	peekIf() *ExecutionContextIf
	areAllIfOk() bool
	addIf(value *ExecutionContextIf)
	addWhile(value *ExecutionContextWhile)
	addForeach(value *ExecutionContextForeach)
	pollIf() *ExecutionContextIf
	pollWhile() *ExecutionContextWhile
	peekWhile() *ExecutionContextWhile
	pollForeach() *ExecutionContextForeach
	peekForeach() *ExecutionContextForeach
	// dumpDebug writes the variables to the diagnostic log (Log.info).
	dumpDebug(message string, info func(string))
}
