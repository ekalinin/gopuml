// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/expression/ShuntingYard.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/text"
)

// shuntingYard turns an infix token list into reverse Polish notation
// (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
type shuntingYard struct {
	ouputQueue    *TokenStack
	operatorStack tokenDeque
}

// tokenDeque is the ArrayDeque used as a stack: "first" is the top.
type tokenDeque []*Token

func (d *tokenDeque) addFirst(t *Token) { *d = append(*d, t) }

func (d *tokenDeque) peekFirst() *Token {
	if len(*d) == 0 {
		return nil
	}
	return (*d)[len(*d)-1]
}

// removeFirst pops the top; on an empty deque it throws
// NoSuchElementException, as Java does.
func (d *tokenDeque) removeFirst() *Token {
	if len(*d) == 0 {
		throwRuntime("NoSuchElementException", "")
	}
	t := (*d)[len(*d)-1]
	*d = (*d)[:len(*d)-1]
	return t
}

func newShuntingYard(it TokenIterator, knowledge Knowledge, location text.StringLocated) (*shuntingYard, error) {
	sy := &shuntingYard{ouputQueue: newTokenStack()}
	for it.hasMoreTokens() {
		token := it.nextToken()
		switch {
		case token.tokenType == tokenNumber || token.tokenType == tokenQuotedString:
			sy.ouputQueue.add(token)
		case token.tokenType == tokenFunctionName:
			sy.operatorStack.addFirst(token)
		case token.tokenType == tokenPlainText:
			name := token.surface
			variable, err := knowledge.GetVariable(name)
			if err != nil {
				return nil, err
			}
			if variable == nil {
				if !isVariableName(name) {
					return nil, newEaterException("Parsing syntax error about "+name, location)
				}
				sy.ouputQueue.add(newToken(name, tokenQuotedString, nil))
			} else {
				sy.ouputQueue.add(variable.toToken())
			}
		case isOperatorOrAffectation(token):
			for (sy.thereIsAFunctionAtTheTopOfTheOperatorStack() ||
				sy.thereIsAnOperatorAtTheTopOfTheOperatorStackWithGreaterPrecedence(token) ||
				sy.theOperatorAtTheTopOfTheOperatorStackHasEqualPrecedenceAndIsLeftAssociative(token)) &&
				sy.theOperatorAtTheTopOfTheOperatorStackIsNotALeftParenthesis(token) {
				sy.ouputQueue.add(sy.operatorStack.removeFirst())
			}
			// push it onto the operator stack.
			sy.operatorStack.addFirst(token)
		case token.tokenType == tokenOpenParenFunc, token.tokenType == tokenOpenParenMath:
			sy.operatorStack.addFirst(token)
		case token.tokenType == tokenCloseParenFunc:
			for sy.operatorStack.peekFirst() != nil && sy.operatorStack.peekFirst().tokenType != tokenOpenParenFunc {
				sy.ouputQueue.add(sy.operatorStack.removeFirst())
			}
			sy.ouputQueue.add(sy.operatorStack.removeFirst())
		case token.tokenType == tokenCloseParenMath:
			for mustPeek(&sy.operatorStack).tokenType != tokenOpenParenMath {
				sy.ouputQueue.add(sy.operatorStack.removeFirst())
			}
			if mustPeek(&sy.operatorStack).tokenType == tokenOpenParenMath {
				sy.operatorStack.removeFirst()
			}
		case token.tokenType == tokenComma:
			for sy.operatorStack.peekFirst() != nil && sy.operatorStack.peekFirst().tokenType != tokenOpenParenFunc {
				sy.ouputQueue.add(sy.operatorStack.removeFirst())
			}
		default:
			throwRuntime("UnsupportedOperationException", token.String())
		}
	}
	for len(sy.operatorStack) > 0 {
		sy.ouputQueue.add(sy.operatorStack.removeFirst())
	}
	return sy, nil
}

// mustPeek is peekFirst followed by a dereference: an empty stack throws
// NullPointerException.
func mustPeek(d *tokenDeque) *Token {
	t := d.peekFirst()
	if t == nil {
		throwRuntime("NullPointerException", "")
	}
	return t
}

var variableNamePattern = regexp.MustCompile(`^[a-zA-Z0-9.$_]+$`)

func isVariableName(name string) bool { return variableNamePattern.MatchString(name) }

func (sy *shuntingYard) thereIsAFunctionAtTheTopOfTheOperatorStack() bool {
	top := sy.operatorStack.peekFirst()
	return top != nil && top.tokenType == tokenFunctionName
}

func (sy *shuntingYard) thereIsAnOperatorAtTheTopOfTheOperatorStackWithGreaterPrecedence(token *Token) bool {
	top := sy.operatorStack.peekFirst()
	return top != nil && isOperatorOrAffectation(top) && top.getPrecedence() > token.getPrecedence()
}

func (sy *shuntingYard) theOperatorAtTheTopOfTheOperatorStackHasEqualPrecedenceAndIsLeftAssociative(token *Token) bool {
	top := sy.operatorStack.peekFirst()
	return top != nil && isOperatorOrAffectation(top) && top.getLeftAssociativity() &&
		top.getPrecedence() == token.getPrecedence()
}

func isOperatorOrAffectation(token *Token) bool {
	return token.tokenType == tokenOperator || token.tokenType == tokenAffectation
}

// theOperatorAtTheTopOfTheOperatorStackIsNotALeftParenthesis always returns
// true, as upstream does.
func (sy *shuntingYard) theOperatorAtTheTopOfTheOperatorStackIsNotALeftParenthesis(*Token) bool {
	return true
}

func (sy *shuntingYard) getQueue() *TokenStack { return sy.ouputQueue }
