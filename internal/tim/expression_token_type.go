// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/expression/TokenType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// TokenType is the kind of a Token.
type TokenType int

// Token types, in upstream order.
const (
	tokenQuotedString TokenType = iota
	tokenJSONData
	tokenOperator
	tokenOpenParenMath
	tokenComma
	tokenCloseParenMath
	tokenNumber
	tokenPlainText
	tokenSpaces
	tokenFunctionName
	tokenOpenParenFunc
	tokenCloseParenFunc
	// tokenAffectation is upstream's TokenType.AFFECTATION, renamed because
	// the line type TLineType.AFFECTATION has the same name.
	tokenAffectation
)

var tokenTypeNames = [...]string{"QUOTED_STRING", "JSON_DATA", "OPERATOR", "OPEN_PAREN_MATH", "COMMA",
	"CLOSE_PAREN_MATH", "NUMBER", "PLAIN_TEXT", "SPACES", "FUNCTION_NAME", "OPEN_PAREN_FUNC", "CLOSE_PAREN_FUNC",
	"AFFECTATION"}

func (t TokenType) String() string { return tokenTypeNames[t] }

// commercialMinusSign marks a subtraction, to tell it from a negative
// number.
const commercialMinusSign = '\u2052'

func (t TokenType) isSingleChar1() bool {
	return t == tokenOpenParenMath || t == tokenComma || t == tokenCloseParenMath
}

func isPlainTextBreak(ch, ch2 uint16) bool {
	tmp := tokenTypeFromChar(ch, ch2)
	return tmp.isSingleChar1() || tmp == tokenOperator || tmp == tokenSpaces || tmp == tokenAffectation
}

func tokenTypeFromChar(ch, ch2 uint16) TokenType {
	switch {
	case isQuote(ch):
		return tokenQuotedString
	case ch == '=':
		return tokenAffectation
	case ch == '(':
		return tokenOpenParenMath
	case ch == ')':
		return tokenCloseParenMath
	case ch == ',':
		return tokenComma
	case isLatinDigit(ch):
		return tokenNumber
	case isSpaceChar(ch):
		return tokenSpaces
	case ch == '-' || getTokenOperator(ch, ch2) != nil:
		return tokenOperator
	}
	return tokenPlainText
}

// eatOneToken reads the next token of the eater, or returns nil at the end
// of the line (or at a colon, with manageColon).
func eatOneToken(lastToken *Token, eater *Eater, manageColon bool) (*Token, error) {
	ch := eater.peekChar()
	if ch == 0 {
		return nil, nil
	}
	if manageColon && ch == ':' {
		return nil, nil
	}
	if ch == '-' && isSubtractionOperator(lastToken) {
		ch = commercialMinusSign
	}
	op := getTokenOperator(ch, eater.peekCharN2())
	switch {
	case isQuote(ch):
		s, err := eater.eatAndGetQuotedString()
		if err != nil {
			return nil, err
		}
		return newToken(s, tokenQuotedString, nil), nil
	case op != nil:
		if op.displayLength() == 1 {
			eater.eatOneChar()
			return newTokenChar(ch, tokenOperator, nil), nil
		}
		c1 := eater.eatOneChar()
		c2 := eater.eatOneChar()
		return newToken(fromUTF16([]uint16{c1, c2}), tokenOperator, nil), nil
	case ch == '=':
		return newTokenChar(eater.eatOneChar(), tokenAffectation, nil), nil
	case ch == '(':
		return newTokenChar(eater.eatOneChar(), tokenOpenParenMath, nil), nil
	case ch == ')':
		return newTokenChar(eater.eatOneChar(), tokenCloseParenMath, nil), nil
	case ch == ',':
		return newTokenChar(eater.eatOneChar(), tokenComma, nil), nil
	case isLatinDigit(ch) || ch == '-':
		return newToken(eater.eatAndGetNumber(), tokenNumber, nil), nil
	case isSpaceChar(ch):
		return newToken(eater.eatAndGetSpaces(), tokenSpaces, nil), nil
	}
	return newToken(eatAndGetTokenPlainText(eater), tokenPlainText, nil), nil
}

func isSubtractionOperator(lastToken *Token) bool {
	if lastToken == nil {
		return false
	}
	switch lastToken.tokenType {
	case tokenOperator, tokenOpenParenMath, tokenComma, tokenAffectation:
		return false
	}
	return true
}

func eatAndGetTokenPlainText(eater *Eater) string {
	var result []uint16
	for {
		ch := eater.peekChar()
		if ch == 0 || isPlainTextBreak(ch, eater.peekCharN2()) {
			return fromUTF16(result)
		}
		result = append(result, eater.eatOneChar())
	}
}
