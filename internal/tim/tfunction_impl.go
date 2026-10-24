// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TFunctionImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/text"
)

// TFunctionImpl is a function or procedure defined by the diagram, or a
// legacy !define / !definelong macro.
type TFunctionImpl struct {
	signature        TFunctionSignature
	args             []TFunctionArgument
	body             []text.StringLocated
	unquoted         bool
	functionType     TFunctionType
	legacyDefinition *string
	containsReturn   bool
}

func newTFunctionImpl(functionName string, args []TFunctionArgument, unquoted bool, functionType TFunctionType) *TFunctionImpl {
	var names []string
	for _, a := range args {
		if !slices.Contains(names, a.name) {
			names = append(names, a.name)
		}
	}
	return &TFunctionImpl{
		signature:    TFunctionSignature{FunctionName: functionName, NbArg: len(args), NamedArguments: names},
		args:         args,
		unquoted:     unquoted,
		functionType: functionType,
	}
}

func (f *TFunctionImpl) CanCover(nbArg int, namedArguments []string) bool {
	for _, n := range namedArguments {
		if !slices.Contains(f.signature.NamedArguments, n) {
			return false
		}
	}
	if nbArg > len(f.args) {
		return false
	}
	neededArgument := 0
	for _, arg := range f.args {
		if slices.Contains(namedArguments, arg.name) {
			continue
		}
		if arg.def == nil {
			neededArgument++
		}
	}
	return nbArg >= neededArgument
}

func (f *TFunctionImpl) getNewMemory(memory TMemory, values []*TValue, namedArguments map[string]*TValue) TMemory {
	result := map[string]*TValue{}
	ivalue := 0
	for _, arg := range f.args {
		var value *TValue
		if v, ok := namedArguments[arg.name]; ok {
			value = v
		} else if ivalue < len(values) {
			value = values[ivalue]
			ivalue++
		} else {
			value = arg.def
		}
		if value == nil {
			throwRuntime("IllegalStateException", "")
		}
		result[arg.name] = value
	}
	return memory.forkFromGlobal(result)
}

func (f *TFunctionImpl) String() string { return "FUNCTION " + f.signature.String() }

func (f *TFunctionImpl) addBody(context *TContext, s text.StringLocated) error {
	f.body = append(f.body, s)
	if context.lineType(s) == lineReturn {
		f.containsReturn = true
		if f.functionType == functionProcedure {
			return newEaterException("A procedure cannot have !return directive. Declare it as a function instead ?", s)
		}
	}
	return nil
}

func (f *TFunctionImpl) ExecuteProcedureInternal(context *TContext, memory TMemory, location text.StringLocated,
	args []*TValue, named map[string]*TValue) error {
	if f.functionType != functionProcedure && f.functionType != functionLegacyDefineLong {
		throwRuntime("IllegalStateException", "")
	}
	leave, err := context.enterCall(location)
	if err != nil {
		return err
	}
	defer leave()
	local := f.getNewMemory(memory, args, named)
	_, err = context.executeLines(local, f.body, functionProcedure, false)
	return err
}

func (f *TFunctionImpl) ExecuteReturnFunction(context *TContext, memory TMemory, location text.StringLocated,
	args []*TValue, named map[string]*TValue) (*TValue, error) {
	if f.functionType == functionLegacyDefine {
		return f.executeReturnLegacyDefine(location.Location, context, memory, args)
	}
	if f.functionType != functionReturn {
		return nil, newEaterException("Illegal call here. Is there a return directive in your function?", location)
	}
	leave, err := context.enterCall(location)
	if err != nil {
		return nil, err
	}
	defer leave()
	local := f.getNewMemory(memory, args, named)
	result, err := context.executeLines(local, f.body, functionReturn, true)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, newEaterException("No return directive found in your function", location)
	}
	return result, nil
}

func (f *TFunctionImpl) executeReturnLegacyDefine(location *text.LineLocation, context *TContext, memory TMemory,
	args []*TValue) (*TValue, error) {
	if f.legacyDefinition == nil {
		throwRuntime("IllegalStateException", "")
	}
	local := f.getNewMemory(memory, args, nil)
	tmp, err := context.applyFunctionsAndVariables(local, text.NewStringLocated(*f.legacyDefinition, location))
	if err != nil {
		return nil, err
	}
	if tmp == nil {
		return TValueFromString(""), nil
	}
	return TValueFromString(*tmp), nil
}

func (f *TFunctionImpl) FunctionType() TFunctionType { return f.functionType }

func (f *TFunctionImpl) Signature() TFunctionSignature { return f.signature }

func (f *TFunctionImpl) setLegacyDefinition(legacyDefinition string) {
	f.legacyDefinition = &legacyDefinition
}

func (f *TFunctionImpl) IsUnquoted() bool { return f.unquoted }

func (f *TFunctionImpl) hasBody() bool { return len(f.body) > 0 }

func (f *TFunctionImpl) finalizeEnddefinelong() {
	if f.functionType != functionLegacyDefineLong {
		throwRuntime("UnsupportedOperationException", "")
	}
	if len(f.body) == 1 {
		f.functionType = functionLegacyDefine
		f.setLegacyDefinition(f.body[0].S)
	}
}

func (f *TFunctionImpl) doesContainReturn() bool { return f.containsReturn }
