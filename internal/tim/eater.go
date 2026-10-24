// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/Eater.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/StringEater.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// Eater reads a line char by char (UTF-16 code units, as Java does). The
// directive eaters embed it.
type Eater struct {
	i  int
	sl text.StringLocated
	s  []uint16
}

func newEater(sl text.StringLocated) Eater {
	return Eater{sl: sl, s: toUTF16(sl.S)}
}

// newStringEater is StringEater: an eater over a string without location.
func newStringEater(s string) *Eater {
	e := newEater(text.StringLocated{S: s})
	return &e
}

func (e *Eater) getLineLocation() *text.LineLocation { return e.sl.Location }

func (e *Eater) getStringLocated() text.StringLocated { return e.sl }

func (e *Eater) getCurrentPosition() int { return e.i }

func (e *Eater) eatAllToEnd() string {
	result := fromUTF16(e.s[e.i:])
	e.i = len(e.s)
	return result
}

// eatExpression reads an expression, or a JSON literal when the expression
// starts with { or [. Invalid JSON throws upstream's ParseException, a
// runtime exception: it panics with the *json.ParseError.
func (e *Eater) eatExpression(context *TContext, memory TMemory) (*TValue, error) {
	ch := e.peekChar()
	if ch == '{' || ch == '[' {
		data := e.eatAllToEnd()
		j, err := json.Parse(data)
		if err != nil {
			panic(err)
		}
		return TValueFromJSON(j), nil
	}
	tokenStack, err := e.eatTokenStack()
	if err != nil {
		return nil, err
	}
	return tokenStack.getResult(e.getStringLocated(), context, memory)
}

func (e *Eater) eatTokenStack() (*TokenStack, error) {
	tokenStack := newTokenStack()
	if err := e.addIntoTokenStack(tokenStack, false); err != nil {
		return nil, err
	}
	if tokenStack.size() == 0 {
		return nil, newEaterException("Missing expression", e.sl)
	}
	return tokenStack, nil
}

func (e *Eater) eatExpressionStopAtColon(context *TContext, memory TMemory) (*TValue, error) {
	tokenStack := newTokenStack()
	if err := e.addIntoTokenStack(tokenStack, true); err != nil {
		return nil, err
	}
	return tokenStack.getResult(e.getStringLocated(), context, memory)
}

func (e *Eater) addIntoTokenStack(tokenStack *TokenStack, stopAtColon bool) error {
	var lastToken *Token
	for {
		token, err := eatOneToken(lastToken, e, stopAtColon)
		if err != nil {
			return err
		}
		if token == nil {
			return nil
		}
		tokenStack.add(token)
		if token.tokenType != tokenSpaces {
			lastToken = token
		}
	}
}

func (e *Eater) eatAndGetQuotedString() (string, error) {
	separator := e.peekChar()
	if !isQuote(separator) {
		return "", newEaterException("quote10", e.sl)
	}
	if err := e.checkAndEatChar(separator); err != nil {
		return "", err
	}
	var value []uint16
	e.addUpTo(separator, &value)
	if err := e.checkAndEatChar(separator); err != nil {
		return "", err
	}
	return fromUTF16(value), nil
}

func (e *Eater) eatAndGetOptionalQuotedString() (string, error) {
	if isQuote(e.peekChar()) {
		return e.eatAndGetQuotedString()
	}
	var value []uint16
	level := 0
	for {
		ch := e.peekChar()
		if ch == 0 {
			return "", newEaterException("until001", e.sl)
		}
		if level == 0 && (ch == ',' || ch == ')') {
			return text.Trin(fromUTF16(value)), nil
		}
		ch = e.eatOneChar()
		switch ch {
		case '(':
			level++
		case ')':
			level--
		}
		value = append(value, ch)
	}
}

func (e *Eater) eatAndGetNumber() string {
	var result []uint16
	for {
		ch := e.peekChar()
		if len(result) == 0 && ch == '-' {
			result = append(result, e.eatOneChar())
			continue
		}
		if ch == 0 || !isLatinDigit(ch) {
			return fromUTF16(result)
		}
		result = append(result, e.eatOneChar())
	}
}

func (e *Eater) eatAndGetSpaces() string {
	var result []uint16
	for {
		ch := e.peekChar()
		if ch == 0 || !isSpaceChar(ch) {
			return fromUTF16(result)
		}
		result = append(result, e.eatOneChar())
	}
}

func (e *Eater) eatAndGetVarname() (string, error) {
	varname := []uint16{e.eatOneChar()}
	if !isLetterOrEmojiOrUnderscoreOrDollar(varname[0]) {
		return "", newEaterException("a002", e.sl)
	}
	e.addUpToLastLetterOrEmojiOrUnderscoreOrDigit(&varname)
	return fromUTF16(varname), nil
}

func (e *Eater) eatAndGetFunctionName() (string, error) {
	varname := []uint16{e.eatOneChar()}
	if !isLetterOrEmojiOrUnderscoreOrDollar(varname[0]) {
		return "", newEaterException("a003", e.sl)
	}
	e.addUpToLastLetterOrEmojiOrUnderscoreOrDigit(&varname)
	return fromUTF16(varname), nil
}

func (e *Eater) skipSpaces() {
	for e.i < len(e.s) && isJavaWhitespace(e.s[e.i]) {
		e.i++
	}
}

func (e *Eater) skipUntilChar(ch uint16) {
	for e.i < len(e.s) && e.s[e.i] != ch {
		e.i++
	}
}

// peekChar returns the current char, or 0 at the end of the line.
func (e *Eater) peekChar() uint16 {
	if e.i >= len(e.s) {
		return 0
	}
	return e.s[e.i]
}

var affectationPattern = regexp.MustCompile(`^\$?[_\p{L}][_\p{L}0-9]*[\t\n\x0B\f\r ]*=`)

// matchAffectation reports whether the rest of the line starts with
// "name =".
func (e *Eater) matchAffectation() bool {
	return affectationPattern.MatchString(fromUTF16(e.s[e.i:]))
}

func (e *Eater) peekCharN2() uint16 {
	if e.i+1 >= len(e.s) {
		return 0
	}
	return e.s[e.i+1]
}

// eatOneChar returns the current char and moves on. Past the end, it
// throws StringIndexOutOfBoundsException, as Java does.
func (e *Eater) eatOneChar() uint16 {
	if e.i >= len(e.s) {
		throwRuntime("StringIndexOutOfBoundsException", "")
	}
	ch := e.s[e.i]
	e.i++
	return ch
}

func (e *Eater) checkAndEatChar(ch uint16) error {
	if e.i >= len(e.s) || e.s[e.i] != ch {
		return newEaterException("a001", e.sl)
	}
	e.i++
	return nil
}

func (e *Eater) safeCheckAndEatChar(ch uint16) bool {
	if e.i >= len(e.s) || e.s[e.i] != ch {
		return false
	}
	e.i++
	return true
}

func (e *Eater) checkAndEatString(s string) error {
	for _, ch := range toUTF16(s) {
		if err := e.checkAndEatChar(ch); err != nil {
			return err
		}
	}
	return nil
}

func (e *Eater) addUpToLastLetterOrEmojiOrUnderscoreOrDigit(sb *[]uint16) {
	for e.i < len(e.s) {
		ch := e.s[e.i]
		if !isLetterOrEmojiOrUnderscoreOrDigit(ch) {
			return
		}
		e.i++
		*sb = append(*sb, ch)
	}
}

func (e *Eater) addUpTo(separator uint16, sb *[]uint16) {
	for e.i < len(e.s) {
		ch := e.peekChar()
		if ch == separator {
			return
		}
		e.i++
		*sb = append(*sb, ch)
	}
}

func (e *Eater) eatDeclareFunction(context *TContext, memory TMemory, unquoted bool, location text.StringLocated,
	allowNoParenthesis bool, ftype TFunctionType) (*TFunctionImpl, error) {
	var args []TFunctionArgument
	functionName, err := e.eatAndGetFunctionName()
	if err != nil {
		return nil, err
	}
	e.skipSpaces()
	if !e.safeCheckAndEatChar('(') {
		if allowNoParenthesis {
			return newTFunctionImpl(functionName, args, unquoted, ftype), nil
		}
		return nil, newEaterException("Missing opening parenthesis", e.sl)
	}
	for {
		e.skipSpaces()
		ch := e.peekChar()
		switch {
		case isLetterOrEmojiOrUnderscoreOrDollar(ch):
			varname, err := e.eatAndGetVarname()
			if err != nil {
				return nil, err
			}
			e.skipSpaces()
			var defValue *TValue
			if e.peekChar() == '=' {
				e.eatOneChar()
				def, err := eatUntilCloseParenthesisOrComma(e)
				if err != nil {
					return nil, err
				}
				if err := def.guessFunctions(location); err != nil {
					return nil, err
				}
				if defValue, err = def.getResult(e.getStringLocated(), context, memory); err != nil {
					return nil, err
				}
			}
			args = append(args, TFunctionArgument{name: varname, def: defValue})
		case ch == ',':
			if err := e.checkAndEatChar(','); err != nil {
				return nil, err
			}
		case ch == ')':
			if err := e.checkAndEatString(")"); err != nil {
				return nil, err
			}
			e.skipSpaces()
			return newTFunctionImpl(functionName, args, unquoted, ftype), nil
		default:
			return nil, newEaterException("Error in function definition", e.sl)
		}
	}
}

func (e *Eater) eatDeclareReturnFunctionWithOptionalReturn(context *TContext, memory TMemory, unquoted bool,
	location text.StringLocated) (*TFunctionImpl, error) {
	result, err := e.eatDeclareFunction(context, memory, unquoted, location, false, functionReturn)
	if err != nil {
		return nil, err
	}
	switch e.peekChar() {
	case 'r':
		if err := e.checkAndEatString("return"); err != nil {
			return nil, err
		}
		e.skipSpaces()
		line := "!return " + e.eatAllToEnd()
		if err := result.addBody(context, text.NewStringLocated(line, location.Location)); err != nil {
			return nil, err
		}
	case '!':
		if err := e.checkAndEatString("!return"); err != nil {
			return nil, err
		}
		e.skipSpaces()
		line := "!return " + e.eatAllToEnd()
		if err := result.addBody(context, text.NewStringLocated(line, location.Location)); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (e *Eater) eatDeclareProcedure(context *TContext, memory TMemory, unquoted bool,
	location text.StringLocated) (*TFunctionImpl, error) {
	return e.eatDeclareFunction(context, memory, unquoted, location, false, functionProcedure)
}
