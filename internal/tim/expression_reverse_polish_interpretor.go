// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/ReversePolishInterpretor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/Knowledge.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Knowledge gives an expression access to variables and functions.
type Knowledge interface {
	GetVariable(name string) (*TValue, error)
	GetFunction(signature TFunctionSignature) TFunction
}

// valueDeque is the ArrayDeque of the interpreter: "first" is the top.
type valueDeque []*TValue

func (d *valueDeque) addFirst(v *TValue) {
	if v == nil {
		throwRuntime("NullPointerException", "")
	}
	*d = append(*d, v)
}

func (d *valueDeque) removeFirst() *TValue {
	if len(*d) == 0 {
		throwRuntime("NoSuchElementException", "")
	}
	v := (*d)[len(*d)-1]
	*d = (*d)[:len(*d)-1]
	return v
}

// reversePolishInterpretor evaluates a queue in reverse Polish notation.
func reversePolishInterpretor(location text.StringLocated, queue *TokenStack, knowledge Knowledge, memory TMemory,
	context *TContext) (*TValue, error) {
	named := map[string]*TValue{}
	var stack valueDeque
	for it := queue.tokenIterator(); it.hasMoreTokens(); {
		token := it.nextToken()
		switch token.tokenType {
		case tokenNumber:
			stack.addFirst(tValueFromNumberToken(token))
		case tokenQuotedString:
			stack.addFirst(tValueFromStringToken(token))
		case tokenJSONData:
			stack.addFirst(TValueFromJSON(token.getJSON()))
		case tokenAffectation:
			v2 := stack.removeFirst()
			v1 := stack.removeFirst()
			named[v1.String()] = v2
		case tokenOperator:
			v2 := stack.removeFirst()
			v1 := stack.removeFirst()
			op := token.getTokenOperator()
			if op == nil {
				return nil, newEaterException("bad op", location)
			}
			stack.addFirst(op.operate(v1, v2))
		case tokenOpenParenFunc:
			n, ok := javaParseInt(token.surface)
			if !ok {
				throwRuntime("NumberFormatException", `For input string: "`+token.surface+`"`)
			}
			nb := int(n) - len(named)
			token2 := it.nextToken()
			if token2 == nil {
				throwRuntime("NullPointerException", "")
			}
			if token2.tokenType != tokenFunctionName {
				return nil, newEaterException("rpn43", location)
			}
			function := knowledge.GetFunction(NewTFunctionSignature(token2.surface, nb))
			if function == nil {
				return nil, newEaterException("Unknown built-in function "+token2.surface, location)
			}
			if !function.CanCover(nb, nil) {
				return nil, newEaterException("Bad number of arguments for "+function.Signature().FunctionName, location)
			}
			args := make([]*TValue, max(nb, 0))
			for i := nb - 1; i >= 0; i-- {
				args[i] = stack.removeFirst()
			}
			r, err := function.ExecuteReturnFunction(context, memory, location, args, named)
			if err != nil {
				return nil, err
			}
			named = map[string]*TValue{}
			stack.addFirst(r)
		default:
			return nil, newEaterException("rpn41", location)
		}
	}
	return stack.removeFirst(), nil
}
