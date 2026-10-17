// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterFunctionCall.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterFunctionCall reads the arguments of a call: "name(a, b, c=1)".
type EaterFunctionCall struct {
	Eater
	values         []*TValue
	namedArguments map[string]*TValue
	isLegacyDefine bool
	unquoted       bool
}

func newEaterFunctionCall(sl text.StringLocated, isLegacyDefine, unquoted bool) *EaterFunctionCall {
	return &EaterFunctionCall{Eater: newEater(sl), namedArguments: map[string]*TValue{}, isLegacyDefine: isLegacyDefine,
		unquoted: unquoted}
}

// readUnquoted reads an unquoted argument and applies variables and
// functions to it.
func (e *EaterFunctionCall) readUnquoted(context *TContext, memory TMemory) (*TValue, error) {
	read, err := e.eatAndGetOptionalQuotedString()
	if err != nil {
		return nil, err
	}
	value, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(read, e.getLineLocation()))
	if err != nil {
		return nil, err
	}
	if value == nil {
		throwRuntime("NullPointerException", "")
	}
	return TValueFromString(*value), nil
}

// readExpression reads an argument expression and evaluates it.
func (e *EaterFunctionCall) readExpression(context *TContext, memory TMemory) (*TValue, error) {
	stack, err := eatUntilCloseParenthesisOrComma(&e.Eater)
	if err != nil {
		return nil, err
	}
	tokens := stack.withoutSpace()
	if err := tokens.guessFunctions(e.getStringLocated()); err != nil {
		return nil, err
	}
	return tokens.getResult(e.getStringLocated(), context, memory)
}

func (e *EaterFunctionCall) readName() (string, error) {
	varname, err := e.eatAndGetVarname()
	if err != nil {
		return "", err
	}
	e.skipSpaces()
	if err := e.checkAndEatChar('='); err != nil {
		return "", err
	}
	e.skipSpaces()
	return varname, nil
}

func (e *EaterFunctionCall) analyze(context *TContext, memory TMemory) error {
	e.skipUntilChar('(')
	if err := e.checkAndEatChar('('); err != nil {
		return err
	}
	e.skipSpaces()
	if e.peekChar() == ')' {
		return e.checkAndEatChar(')')
	}
	for {
		e.skipSpaces()
		switch {
		case e.isLegacyDefine:
			result, err := e.readUnquoted(context, memory)
			if err != nil {
				return err
			}
			e.values = append(e.values, result)
		case e.unquoted:
			if e.matchAffectation() {
				varname, err := e.readName()
				if err != nil {
					return err
				}
				result, err := e.readUnquoted(context, memory)
				if err != nil {
					return err
				}
				e.namedArguments[varname] = result
			} else {
				result, err := e.readUnquoted(context, memory)
				if err != nil {
					return err
				}
				e.values = append(e.values, result)
			}
		default:
			if e.matchAffectation() {
				varname, err := e.readName()
				if err != nil {
					return err
				}
				result, err := e.readExpression(context, memory)
				if err != nil {
					return err
				}
				e.namedArguments[varname] = result
			} else {
				result, err := e.readExpression(context, memory)
				if err != nil {
					return err
				}
				e.values = append(e.values, result)
			}
		}
		e.skipSpaces()
		ch := e.eatOneChar()
		if ch == ',' {
			continue
		}
		if ch == ')' {
			return nil
		}
		if e.unquoted {
			return newEaterException("unquoted function/procedure cannot use expression.", e.getStringLocated())
		}
		return newEaterException("call001", e.getStringLocated())
	}
}

func (e *EaterFunctionCall) getValues() []*TValue { return e.values }

func (e *EaterFunctionCall) getNamedArguments() map[string]*TValue { return e.namedArguments }

// namedArgumentNames returns the names of the named arguments.
func (e *EaterFunctionCall) namedArgumentNames() []string {
	names := make([]string, 0, len(e.namedArguments))
	for name := range e.namedArguments {
		names = append(names, name)
	}
	return names
}
