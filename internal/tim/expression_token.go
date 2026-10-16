// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/Token.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/json"
)

// Token is one token of a preprocessor expression.
type Token struct {
	surface   string
	json      json.Value
	tokenType TokenType
}

func newToken(surface string, tokenType TokenType, j json.Value) *Token {
	return &Token{surface: surface, tokenType: tokenType, json: j}
}

func newTokenChar(surface uint16, tokenType TokenType, j json.Value) *Token {
	return newToken(fromUTF16([]uint16{surface}), tokenType, j)
}

func (t *Token) String() string { return t.tokenType.String() + "{" + t.surface + "}" }

// getTokenOperator returns the operator of an OPERATOR token, or nil.
func (t *Token) getTokenOperator() *TokenOperator {
	if t.tokenType != tokenOperator {
		throwRuntime("IllegalStateException", "")
	}
	u := utf16.Encode([]rune(t.surface))
	var ch2 uint16
	if len(u) > 1 {
		ch2 = u[1]
	}
	return getTokenOperator(u[0], ch2)
}

func (t *Token) getSurface() string { return t.surface }

func (t *Token) getTokenType() TokenType { return t.tokenType }

func (t *Token) muteToFunction() *Token {
	if t.tokenType != tokenPlainText {
		throwRuntime("IllegalStateException", "")
	}
	return newToken(t.surface, tokenFunctionName, nil)
}

func (t *Token) getJSON() json.Value {
	if t.tokenType != tokenJSONData {
		throwRuntime("IllegalStateException", "")
	}
	return t.json
}

func (t *Token) getPrecedence() int {
	if t.tokenType == tokenAffectation {
		return opEquals.precedence
	}
	op := t.getTokenOperator()
	if op == nil {
		throwRuntime("NullPointerException", "")
	}
	return op.precedence
}

func (t *Token) getLeftAssociativity() bool { return true }
