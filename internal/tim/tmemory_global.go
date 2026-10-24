// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TMemoryGlobal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"maps"
	"slices"
	"strconv"

	"github.com/ekalinin/gopuml/internal/text"
)

// TMemoryGlobal is the memory of top-level code.
type TMemoryGlobal struct {
	executionContexts
	globalVariables map[string]*TValue
	variables       *TrieImpl
}

func newTMemoryGlobal() *TMemoryGlobal {
	return &TMemoryGlobal{globalVariables: map[string]*TValue{}, variables: newTrieImpl()}
}

func (m *TMemoryGlobal) GetVariable(varname string) *TValue { return m.globalVariables[varname] }

func (m *TMemoryGlobal) dumpDebug(message string, info func(string)) {
	info("[MemGlobal] Start of memory_dump " + message)
	m.dumpMemoryInternal(info)
	info("[MemGlobal] End of memory_dump")
}

// sortedNames returns the names in the order of a Java TreeMap<String>.
func sortedNames(variables map[string]*TValue) []string {
	return slices.SortedFunc(maps.Keys(variables), javaCompare)
}

func (m *TMemoryGlobal) dumpMemoryInternal(info func(string)) {
	info("[MemGlobal] Number of variable(s) : " + strconv.Itoa(len(m.globalVariables)))
	for _, name := range sortedNames(m.globalVariables) {
		info("[MemGlobal] " + name + " = " + m.globalVariables[name].String())
	}
}

func (m *TMemoryGlobal) PutVariable(varname string, value *TValue, scope TVariableScope, location text.StringLocated) error {
	if scope == scopeLocal {
		return newEaterException("Cannot use local variable here", location)
	}
	m.globalVariables[varname] = value
	m.variables.add(varname)
	return nil
}

func (m *TMemoryGlobal) RemoveVariable(varname string) {
	delete(m.globalVariables, varname)
	m.variables.remove(varname)
}

func (m *TMemoryGlobal) IsEmpty() bool { return len(m.globalVariables) == 0 }

// VariablesNames returns the names of the global variables.
func (m *TMemoryGlobal) VariablesNames() []string { return sortedNames(m.globalVariables) }

func (m *TMemoryGlobal) variablesNames3() Trie { return m.variables }

func (m *TMemoryGlobal) forkFromGlobal(input map[string]*TValue) TMemory {
	return newTMemoryLocal(m, input)
}
