// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TMemoryLocal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"maps"
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// TMemoryLocal is the memory of a function call: the arguments (overridden
// variables), the variables created by the call, and the global memory.
type TMemoryLocal struct {
	executionContexts
	memoryGlobal         *TMemoryGlobal
	overridenVariables00 *TrieImpl
	overridenVariables01 map[string]*TValue
	localVariables00     *TrieImpl
	localVariables01     map[string]*TValue
}

func newTMemoryLocal(global *TMemoryGlobal, input map[string]*TValue) *TMemoryLocal {
	return &TMemoryLocal{
		memoryGlobal:         global,
		overridenVariables01: maps.Clone(input),
		localVariables00:     newTrieImpl(),
		localVariables01:     map[string]*TValue{},
	}
}

func (m *TMemoryLocal) dumpDebug(message string, info func(string)) {
	info("[MemLocal] Start of memory_dump " + message)
	m.memoryGlobal.dumpMemoryInternal(info)
	info("[MemLocal] Number of overridden variable(s) : " + strconv.Itoa(len(m.overridenVariables01)))
	for _, name := range sortedNames(m.overridenVariables01) {
		info("[MemLocal] " + name + " = " + m.overridenVariables01[name].String())
	}
	info("[MemLocal] Number of local variable(s) : " + strconv.Itoa(len(m.localVariables01)))
	for _, name := range sortedNames(m.localVariables01) {
		info("[MemLocal] " + name + " = " + m.localVariables01[name].String())
	}
	info("[MemGlobal] End of memory_dump")
}

func (m *TMemoryLocal) PutVariable(varname string, value *TValue, scope TVariableScope, location text.StringLocated) error {
	if scope == scopeGlobal {
		return m.memoryGlobal.PutVariable(varname, value, scope, location)
	}
	if _, overridden := m.overridenVariables01[varname]; scope == scopeLocal || overridden {
		m.overridenVariables01[varname] = value
		if m.overridenVariables00 != nil {
			m.overridenVariables00.add(varname)
		}
	} else if m.memoryGlobal.GetVariable(varname) != nil {
		return m.memoryGlobal.PutVariable(varname, value, scope, location)
	} else {
		m.localVariables01[varname] = value
		m.localVariables00.add(varname)
	}
	return nil
}

func (m *TMemoryLocal) RemoveVariable(varname string) {
	if _, overridden := m.overridenVariables01[varname]; overridden {
		delete(m.overridenVariables01, varname)
		if m.overridenVariables00 != nil {
			m.overridenVariables00.remove(varname)
		}
	} else if m.memoryGlobal.GetVariable(varname) != nil {
		m.memoryGlobal.RemoveVariable(varname)
	} else {
		delete(m.localVariables01, varname)
		m.localVariables00.remove(varname)
	}
}

func (m *TMemoryLocal) GetVariable(varname string) *TValue {
	if result := m.overridenVariables01[varname]; result != nil {
		return result
	}
	if result := m.memoryGlobal.GetVariable(varname); result != nil {
		return result
	}
	return m.localVariables01[varname]
}

// localTrie looks names up in the global, overridden and local tries and
// keeps the longest match, preferring them in that order on ties.
type localTrie struct{ m *TMemoryLocal }

func (t localTrie) getLonguestMatchStartingIn(s []uint16, pos int) int {
	s1 := t.m.memoryGlobal.variablesNames3().getLonguestMatchStartingIn(s, pos)
	s2 := t.m.overridenVariables00.getLonguestMatchStartingIn(s, pos)
	s3 := t.m.localVariables00.getLonguestMatchStartingIn(s, pos)
	if s1 >= s2 && s1 >= s3 {
		return s1
	}
	if s2 >= s3 && s2 >= s1 {
		return s2
	}
	return s3
}

func (m *TMemoryLocal) variablesNames3() Trie {
	if m.overridenVariables00 == nil {
		m.overridenVariables00 = newTrieImpl()
		for name := range m.overridenVariables01 {
			m.overridenVariables00.add(name)
		}
	}
	return localTrie{m}
}

func (m *TMemoryLocal) IsEmpty() bool {
	return m.memoryGlobal.IsEmpty() && len(m.localVariables01) == 0 && len(m.overridenVariables01) == 0
}

func (m *TMemoryLocal) forkFromGlobal(input map[string]*TValue) TMemory {
	return newTMemoryLocal(m.memoryGlobal, input)
}
