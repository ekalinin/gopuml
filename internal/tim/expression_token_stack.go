// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/TokenStack.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/TokenIterator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// TokenIterator reads tokens one by one.
type TokenIterator interface {
	nextToken() *Token
	peekToken() *Token
	hasMoreTokens() bool
}

// TokenStack is a list of tokens. A sub stack shares the tokens of its
// parent, like Java's List.subList view.
type TokenStack struct {
	tokens []*Token
}

func newTokenStack() *TokenStack { return &TokenStack{} }

func (ts *TokenStack) size() int { return len(ts.tokens) }

func (ts *TokenStack) subTokenStack(i int) *TokenStack {
	return &TokenStack{tokens: ts.tokens[i:len(ts.tokens):len(ts.tokens)]}
}

func (ts *TokenStack) String() string {
	parts := make([]string, len(ts.tokens))
	for i, t := range ts.tokens {
		parts[i] = t.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (ts *TokenStack) add(token *Token) { ts.tokens = append(ts.tokens, token) }

func (ts *TokenStack) withoutSpace() *TokenStack {
	result := newTokenStack()
	for _, token := range ts.tokens {
		if token.tokenType != tokenSpaces {
			result.add(token)
		}
	}
	return result
}

// eatUntilCloseParenthesisOrComma reads the tokens of one function argument.
func eatUntilCloseParenthesisOrComma(eater *Eater) (*TokenStack, error) {
	result := newTokenStack()
	level := 0
	var lastToken *Token
	for {
		eater.skipSpaces()
		ch := eater.peekChar()
		if ch == 0 {
			return nil, newEaterException("until001", eater.getStringLocated())
		}
		if level == 0 && (ch == ',' || ch == ')') {
			return result, nil
		}
		token, err := eatOneToken(lastToken, eater, false)
		if err != nil {
			return nil, err
		}
		switch token.tokenType {
		case tokenOpenParenMath:
			level++
		case tokenCloseParenMath:
			level--
		}
		if token.tokenType != tokenSpaces {
			lastToken = token
		}
		result.add(token)
	}
}

// skipUntilCloseParenthesisOrComma is the TokenIterator variant of
// eatUntilCloseParenthesisOrComma.
func skipUntilCloseParenthesisOrComma(it TokenIterator, location text.StringLocated) error {
	level := 0
	for {
		ch := it.peekToken()
		if ch == nil {
			return newEaterException("until002", location)
		}
		typech := ch.tokenType
		if level == 0 && (typech == tokenComma || typech == tokenCloseParenMath) || typech == tokenCloseParenFunc {
			return nil
		}
		token := it.nextToken()
		switch token.tokenType {
		case tokenOpenParenMath, tokenOpenParenFunc:
			level++
		case tokenCloseParenMath, tokenCloseParenFunc:
			level--
		}
	}
}

func (ts *TokenStack) countFunctionArg(it TokenIterator, location text.StringLocated) (int, error) {
	type1 := it.peekToken().tokenType
	if type1 == tokenCloseParenMath || type1 == tokenCloseParenFunc {
		return 0, nil
	}
	result := 1
	for it.hasMoreTokens() {
		if err := skipUntilCloseParenthesisOrComma(it, location); err != nil {
			return 0, err
		}
		token := it.nextToken()
		switch token.tokenType {
		case tokenCloseParenMath, tokenCloseParenFunc:
			return result, nil
		case tokenComma:
			result++
		default:
			return 0, newEaterException("count13", location)
		}
	}
	return 0, newEaterException("count12", location)
}

// guessFunctions turns the plain text before a parenthesis into a function
// name. The parentheses are visited in the iteration order of upstream's
// HashMap<Integer, Integer>, which decides how nested calls are counted.
func (ts *TokenStack) guessFunctions(location text.StringLocated) error {
	const null = -1
	var open []int
	type pair struct{ open, close int }
	var parens []pair
	order := newJavaHashOrder()
	put := func(key, value int) {
		for i := range parens {
			if parens[i].open == key {
				parens[i].close = value
				return
			}
		}
		parens = append(parens, pair{key, value})
		if key == null {
			order.add(0)
		} else {
			order.add(int32(key))
		}
	}
	for i, token := range ts.tokens {
		switch token.tokenType {
		case tokenOpenParenMath:
			open = append(open, i)
		case tokenCloseParenMath:
			key := null
			if len(open) > 0 {
				key = open[len(open)-1]
				open = open[:len(open)-1]
			}
			put(key, i)
		}
	}
	for _, idx := range order.order() {
		ids := parens[idx]
		if ids.open == null {
			throwRuntime("NullPointerException", "")
		}
		iopen, iclose := ids.open, ids.close
		if iopen > 0 && ts.tokens[iopen-1].tokenType == tokenPlainText {
			ts.tokens[iopen-1] = newToken(ts.tokens[iopen-1].surface, tokenFunctionName, nil)
			nbArg, err := ts.countFunctionArg(ts.subTokenStack(iopen+1).tokenIterator(), location)
			if err != nil {
				return err
			}
			ts.tokens[iopen] = newToken(strconv.Itoa(nbArg), tokenOpenParenFunc, nil)
			ts.tokens[iclose] = newToken(")", tokenCloseParenFunc, nil)
		}
	}
	return nil
}

type tokenStackIterator struct {
	ts  *TokenStack
	pos int
}

// peekToken returns the next token; past the end it throws, like
// List.get.
func (it *tokenStackIterator) peekToken() *Token { return it.ts.tokens[it.pos] }

func (it *tokenStackIterator) nextToken() *Token {
	if !it.hasMoreTokens() {
		return nil
	}
	t := it.ts.tokens[it.pos]
	it.pos++
	return t
}

func (it *tokenStackIterator) hasMoreTokens() bool { return it.pos < len(it.ts.tokens) }

func (ts *TokenStack) tokenIterator() TokenIterator { return &tokenStackIterator{ts: ts} }

// getResult evaluates the expression.
func (ts *TokenStack) getResult(location text.StringLocated, context *TContext, memory TMemory) (*TValue, error) {
	knowledge := context.asKnowledge(memory, location.Location)
	tmp := ts.withoutSpace()
	if err := tmp.guessFunctions(location); err != nil {
		return nil, err
	}
	it := tmp.tokenIterator()
	shuntingYard, err := newShuntingYard(it, knowledge, location)
	if err != nil {
		return nil, err
	}
	return reversePolishInterpretor(location, shuntingYard.getQueue(), knowledge, memory, context)
}
