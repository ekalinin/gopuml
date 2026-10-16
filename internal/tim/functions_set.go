// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/FunctionsSet.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// FunctionsSet holds the functions known to a TContext.
type FunctionsSet struct {
	functions map[signatureKey]TFunction
	// keys and order reproduce the iteration order of upstream's HashMap,
	// which decides which function getFunctionSmart picks when several
	// can cover a call. phantom keys stand for builtins that are not ported
	// yet: they take their place in the HashMap order but are never found.
	keys            []signatureKey
	phantom         []bool
	order           *javaHashOrder
	functionsByName map[string]map[signatureKey]TFunction
	functionsFinal  map[signatureKey]bool
	functions3      *TrieImpl
	pending         *TFunctionImpl
}

func newFunctionsSet() *FunctionsSet {
	return &FunctionsSet{
		functions:       map[signatureKey]TFunction{},
		order:           newJavaHashOrder(),
		functionsByName: map[string]map[signatureKey]TFunction{},
		functionsFinal:  map[signatureKey]bool{},
		functions3:      newTrieImpl(),
	}
}

// putFunction is functions.put: a new key takes its place in the HashMap
// order, an existing key keeps it.
func (fs *FunctionsSet) putFunction(fn TFunction) {
	sig := fn.Signature()
	key := sig.key()
	if _, ok := fs.functions[key]; !ok {
		fs.keys = append(fs.keys, key)
		fs.phantom = append(fs.phantom, false)
		fs.order.add(sig.hashCode())
	}
	fs.functions[key] = fn
}

// addPhantom reserves the HashMap slot of a builtin that is not ported yet.
func (fs *FunctionsSet) addPhantom(sig TFunctionSignature) {
	fs.keys = append(fs.keys, sig.key())
	fs.phantom = append(fs.phantom, true)
	fs.order.add(sig.hashCode())
}

func (fs *FunctionsSet) getFunctionSmart(searched TFunctionSignature) TFunction {
	if f := fs.functions[searched.key()]; f != nil {
		return f
	}
	for _, idx := range fs.order.order() {
		if fs.phantom[idx] {
			continue
		}
		candidate := fs.functions[fs.keys[idx]]
		if !candidate.Signature().sameFunctionNameAs(searched) {
			continue
		}
		if candidate.CanCover(searched.NbArg, searched.NamedArguments) {
			return candidate
		}
	}
	return nil
}

func (fs *FunctionsSet) size() int { return len(fs.functions) }

func (fs *FunctionsSet) getLonguestMatchStartingIn(s []uint16, pos int) int {
	return fs.functions3.getLonguestMatchStartingIn(s, pos)
}

func (fs *FunctionsSet) pendingFunction() *TFunctionImpl { return fs.pending }

func (fs *FunctionsSet) addFunction(fn TFunction) {
	if fn.FunctionType() == functionLegacyDefineLong {
		fn.(*TFunctionImpl).finalizeEnddefinelong()
	}
	fs.putFunction(fn)
	fs.functions3.add(fn.Signature().FunctionName + "(")
	fs.updateFunctionsByName(fn)
}

func (fs *FunctionsSet) updateFunctionsByName(fn TFunction) {
	name := fn.Signature().FunctionName
	m := fs.functionsByName[name]
	if m == nil {
		m = map[signatureKey]TFunction{}
		fs.functionsByName[name] = m
	}
	m[fn.Signature().key()] = fn
}

// doesFunctionExist reports whether at least one function has the name.
func (fs *FunctionsSet) doesFunctionExist(functionName string) bool {
	_, ok := fs.functionsByName[functionName]
	return ok
}

// getFunctionsByName returns the functions with the name.
func (fs *FunctionsSet) getFunctionsByName(functionName string) map[signatureKey]TFunction {
	return fs.functionsByName[functionName]
}

func (fs *FunctionsSet) executeEndfunction() {
	fs.addFunction(fs.pending)
	fs.pending = nil
}

func (fs *FunctionsSet) executeLegacyDefine(context *TContext, memory TMemory, s text.StringLocated) error {
	if fs.pending != nil {
		return newEaterException("already0048", s)
	}
	legacyDefine := newEaterLegacyDefine(s)
	if err := legacyDefine.analyze(context, memory); err != nil {
		return err
	}
	function := legacyDefine.getFunction()
	fs.putFunction(function)
	fs.functions3.add(function.Signature().FunctionName + "(")
	fs.updateFunctionsByName(function)
	return nil
}

func (fs *FunctionsSet) executeLegacyDefineLong(context *TContext, memory TMemory, s text.StringLocated) error {
	if fs.pending != nil {
		return newEaterException("already0068", s)
	}
	legacyDefineLong := newEaterLegacyDefineLong(s)
	if err := legacyDefineLong.analyze(context, memory); err != nil {
		return err
	}
	fs.pending = legacyDefineLong.getFunction()
	return nil
}

// declareFunction is the shared part of executeDeclareReturnFunction and
// executeDeclareProcedure.
func (fs *FunctionsSet) declareFunction(s text.StringLocated, function *TFunctionImpl, finalFlag bool) error {
	declaredSignature := function.Signature().key()
	previous := fs.functions[declaredSignature]
	if previous != nil && (finalFlag || fs.functionsFinal[declaredSignature]) {
		return newEaterException("This function is already defined", s)
	}
	if finalFlag {
		fs.functionsFinal[declaredSignature] = true
	}
	if function.hasBody() {
		fs.addFunction(function)
	} else {
		fs.pending = function
	}
	return nil
}

func (fs *FunctionsSet) executeDeclareReturnFunction(context *TContext, memory TMemory, s text.StringLocated) error {
	if fs.pending != nil {
		return newEaterException("already0068", s)
	}
	declareFunction := newEaterDeclareReturnFunction(s)
	if err := declareFunction.analyze(context, memory); err != nil {
		return err
	}
	return fs.declareFunction(s, declareFunction.getFunction(), declareFunction.getFinalFlag())
}

func (fs *FunctionsSet) executeDeclareProcedure(context *TContext, memory TMemory, s text.StringLocated) error {
	if fs.pending != nil {
		return newEaterException("already0068", s)
	}
	declareFunction := newEaterDeclareProcedure(s)
	if err := declareFunction.analyze(context, memory); err != nil {
		return err
	}
	return fs.declareFunction(s, declareFunction.getFunction(), declareFunction.getFinalFlag())
}
