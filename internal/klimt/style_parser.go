// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/StyleParser.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/StyleToken.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/StyleTokenType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/StyleParsingException.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/Context.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/parser/CssVariables.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/CharInspector.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/CharInspectorImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
)

// StyleParsingError is upstream's StyleParsingException: a style sheet that
// cannot be parsed.
type StyleParsingError struct {
	Message string
}

func (e *StyleParsingError) Error() string { return e.Message }

// StyleTokenType is the kind of a style sheet token.
type StyleTokenType int

// Token kinds, in upstream's order.
const (
	StyleTokenOpenBracket StyleTokenType = iota
	StyleTokenCloseBracket
	StyleTokenString
	StyleTokenComma
	StyleTokenStar
	StyleTokenNewline
	StyleTokenSemicolon
	StyleTokenColon
	StyleTokenArobaseMedia
	styleTokenNone // stands for upstream's null token
)

var styleTokenTypeNames = []string{"OPEN_BRACKET", "CLOSE_BRACKET", "STRING", "COMMA", "STAR", "NEWLINE",
	"SEMICOLON", "COLON", "AROBASE_MEDIA"}

// String returns upstream's constant name.
func (t StyleTokenType) String() string {
	return enumName("StyleTokenType", styleTokenTypeNames, int(t))
}

// StyleToken is a token of a style sheet.
type StyleToken struct {
	Type StyleTokenType
	Data string
}

// String is upstream's toString, for example "STRING[root]".
func (t StyleToken) String() string { return t.Type.String() + "[" + t.Data + "]" }

// charInspector walks the characters of lines, upstream's CharInspectorImpl:
// each line is trimmed, a newline is added after each line when
// insertNewlines is set, empty lines are skipped, and peek does not look
// past the current line (0 there).
type charInspector struct {
	lines          [][]rune
	insertNewlines bool
	line           int
	pos            int
}

func newCharInspector(lines command.BlocLines, insertNewlines bool) *charInspector {
	ci := &charInspector{insertNewlines: insertNewlines}
	for _, l := range lines.Lines() {
		s := l.Trimmed().S
		if insertNewlines {
			s += "\n"
		}
		ci.lines = append(ci.lines, []rune(s))
	}
	if len(ci.lines) == 0 {
		// Upstream reads the first line unconditionally and throws.
		throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
	}
	return ci
}

func (ci *charInspector) peek(ahead int) rune {
	if ci.line == -1 {
		return 0
	}
	current := ci.lines[ci.line]
	if ci.pos+ahead >= len(current) {
		return 0
	}
	return current[ci.pos+ahead]
}

func (ci *charInspector) jump() {
	if ci.line == -1 {
		throw(illegalState, "")
	}
	ci.pos++
	if ci.pos >= len(ci.lines[ci.line]) {
		ci.line++
		ci.pos = 0
	}
	for ci.line < len(ci.lines) && len(ci.lines[ci.line]) == 0 {
		ci.line++
	}
	if ci.line >= len(ci.lines) {
		ci.line = -1
	}
}

// cssVariables are the "--name: value" variables of a style sheet.
type cssVariables struct {
	variables map[string]string
}

var (
	cssLearnPattern    = regexp.MustCompile(javaDots(`^--([_\w][-_\w]+)[ :]+(.*?);?$`))
	cssRetrievePattern = regexp.MustCompile(`^var\(-*([_\w][-_\w]+)\)$`)
)

func (c *cssVariables) learnLine(s string) {
	if m := cssLearnPattern.FindStringSubmatch(s); m != nil {
		c.variables[m[1]] = m[2]
	}
}

func (c *cssVariables) learn(name, value string) {
	name = strings.TrimPrefix(name, "--")
	c.variables[name] = value
}

func (c *cssVariables) value(v string) string {
	if strings.HasPrefix(v, "var(") {
		if m := cssRetrievePattern.FindStringSubmatch(v); m != nil {
			if result, ok := c.variables[m[1]]; ok {
				return result
			}
		}
	}
	return v
}

// styleContext is the stack of selectors being parsed, upstream's
// parser.Context.
type styleContext struct {
	signatures []StyleSignatureBasic
	values     styleValues
	parent     *styleContext
}

func styleContextEmpty() *styleContext {
	return &styleContext{signatures: []StyleSignatureBasic{StyleSignatureBasicEmpty()}}
}

var nonDigits = regexp.MustCompile(`[^0-9]`)

func (c *styleContext) push(newString string) *styleContext {
	newString = strings.TrimPrefix(newString, ":")
	result := &styleContext{parent: c}
	star := false
	if strings.HasSuffix(newString, StyleSignatureStar) {
		newString = strings.TrimFunc(newString[:len(newString)-1], func(r rune) bool { return r <= ' ' })
		star = true
	}
	for _, s := range javaSplit(newString, ",") {
		for _, ssb := range c.signatures {
			switch {
			case strings.HasPrefix(s, "."):
				ssb = ssb.AddStereotype(s)
			case strings.HasPrefix(s, "depth("):
				ssb = ssb.AddLevel(javaParseInt(nonDigits.ReplaceAllString(s, "")))
			default:
				if sname, ok := SNameRetrieve(s); ok {
					ssb = ssb.AddSName(sname)
				} else {
					ssb = ssb.AddStereotype(s)
				}
			}
			if star {
				ssb = ssb.AddStar()
			}
			result.signatures = append(result.signatures, ssb)
		}
	}
	return result
}

func (c *styleContext) pop() *styleContext {
	if c.isEmpty() {
		throw(illegalState, "")
	}
	return c.parent
}

func (c *styleContext) isEmpty() bool {
	if len(c.signatures) == 0 {
		throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
	}
	return c.signatures[0].IsEmpty()
}

func (c *styleContext) putInContext(key PName, value Value) { c.values[key] = value }

// toStyles returns the styles of the context. As upstream, a style without
// stereotype holds the values of the context itself: the root context and
// top-level "depth(n)" contexts are never popped, so declarations written
// after their "}" still land in the styles already returned.
func (c *styleContext) toStyles() []*Style {
	var result []*Style
	for _, signature := range c.signatures {
		if signature.IsWithDot() {
			tmp := addPriorityForStereotype(c.values)
			if countValues(tmp) > 0 {
				result = append(result, newStyleValues(signature, tmp))
			}
		} else if countValues(c.values) > 0 {
			result = append(result, newStyleShared(signature, &c.values))
		}
	}
	return result
}

func countValues(v styleValues) int {
	n := 0
	for _, x := range v {
		if x != nil {
			n++
		}
	}
	return n
}

// StyleParser parses CSS-like style sheets.
type StyleParser struct {
	variables cssVariables
	scheme    StyleScheme
	context   *styleContext
	counter   AutomaticCounter
}

// NewStyleParser returns a parser numbering the values from 1.
func NewStyleParser() *StyleParser { return NewStyleParserCounter(&AutomaticCounterBasic{}) }

// NewStyleParserCounter returns a parser numbering the values with counter
// (usually the StyleBuilder the styles are for).
func NewStyleParserCounter(counter AutomaticCounter) *StyleParser {
	return &StyleParser{variables: cssVariables{variables: map[string]string{}}, scheme: StyleSchemeRegular,
		context: styleContextEmpty(), counter: counter}
}

// ParseSingleLine parses "prop: value; ..." into a style without signature.
func (p *StyleParser) ParseSingleLine(s string) (*Style, error) {
	tokens, err := p.tokenize(newCharInspector(command.SingleString(s), false))
	if err != nil {
		return nil, err
	}
	if _, err := p.parseNow(tokens); err != nil {
		return nil, err
	}
	return newStyleShared(StyleSignatureBasicEmpty(), &p.context.values), nil
}

// Parse parses the lines of a style sheet into styles, in declaration
// order.
func (p *StyleParser) Parse(lines command.BlocLines) ([]*Style, error) {
	if lines.Size() == 0 {
		return nil, nil
	}
	tokens, err := p.tokenize(newCharInspector(lines, true))
	if err != nil {
		return nil, err
	}
	return p.parseNow(tokens)
}

// styleTokenPeeker walks tokens; past the end it returns the none token,
// which stands for upstream's null.
type styleTokenPeeker struct {
	tokens []StyleToken
	pos    int
}

func (pk *styleTokenPeeker) peek(i int) StyleToken {
	if pk.pos+i >= len(pk.tokens) {
		return StyleToken{Type: styleTokenNone}
	}
	return pk.tokens[pk.pos+i]
}

func (pk *styleTokenPeeker) jump() { pk.pos++ }

// typeOf is upstream's peek(i).getType(), which throws NullPointerException
// past the end.
func (pk *styleTokenPeeker) typeOf(i int) StyleTokenType {
	t := pk.peek(i)
	if t.Type == styleTokenNone {
		throw(nullPointer, "")
	}
	return t.Type
}

func (p *StyleParser) parseNow(tokens []StyleToken) ([]*Style, error) {
	var result []*Style
	peeker := &styleTokenPeeker{tokens: tokens}
	for peeker.peek(0).Type != styleTokenNone {
		token := peeker.peek(0)
		peeker.jump()
		if token.Type == StyleTokenNewline || token.Type == StyleTokenSemicolon {
			continue
		}
		if token.Type == StyleTokenString && javaEqualsIgnoreCase(token.Data, "<style>") {
			continue
		}
		if token.Type == StyleTokenString && javaEqualsIgnoreCase(token.Data, "</style>") {
			continue
		}
		if peeker.typeOf(0) == StyleTokenComma {
			full := token.Data + readWithComma(peeker)
			skipNewLines(peeker)
			if peeker.typeOf(0) == StyleTokenOpenBracket {
				p.context = p.context.push(full)
				peeker.jump()
				continue
			}
			throw(illegalState, "")
		}
		switch {
		case token.Type == StyleTokenString:
			full := token.Data
			if peeker.typeOf(0) == StyleTokenStar {
				peeker.jump()
				full += StyleSignatureStar
			}
			skipNewLines(peeker)
			if peeker.typeOf(0) == StyleTokenOpenBracket {
				p.context = p.context.push(full)
				peeker.jump()
				continue
			}
			skipColon(peeker)
			if strings.HasPrefix(token.Data, "--") {
				v, err := readValue(peeker)
				if err != nil {
					return nil, err
				}
				p.variables.learn(token.Data, v)
			} else if peeker.typeOf(0) == StyleTokenString {
				v, err := readValue(peeker)
				if err != nil {
					return nil, err
				}
				valueString := p.variables.value(v)
				keyString := token.Data
				if key, ok := PNameFromName(keyString, p.scheme); ok {
					var value Value
					if p.scheme == StyleSchemeRegular {
						value = ValueImplRegular(valueString, p.counter)
					} else {
						value = ValueImplDark(valueString, p.counter)
					}
					p.context.putInContext(key, value)
				}
			} else {
				return nil, &StyleParsingError{Message: "parsing"}
			}
		case token.Type == StyleTokenCloseBracket:
			result = append(result, p.context.toStyles()...)
			if !p.context.isEmpty() {
				p.context = p.context.pop()
			}
		case token.Type == StyleTokenArobaseMedia:
			p.scheme = StyleSchemeDark
			continue
		case token.Type == StyleTokenColon && peeker.typeOf(0) == StyleTokenString &&
			peeker.typeOf(1) == StyleTokenOpenBracket:
			full := token.Data + peeker.peek(0).Data
			p.context = p.context.push(full)
			peeker.jump()
			peeker.jump()
			continue
		case token.Type == StyleTokenColon && peeker.typeOf(0) == StyleTokenString &&
			peeker.typeOf(1) == StyleTokenStar && peeker.typeOf(2) == StyleTokenOpenBracket:
			full := token.Data + peeker.peek(0).Data + peeker.peek(1).Data
			p.context = p.context.push(full)
			peeker.jump()
			peeker.jump()
			peeker.jump()
			continue
		case token.Type == StyleTokenOpenBracket:
			return nil, &StyleParsingError{Message: "Invalid open bracket"}
		default:
			throw(illegalState, token.String())
		}
	}
	return result, nil
}

func readWithComma(ins *styleTokenPeeker) string {
	var result strings.Builder
	for ins.peek(0).Type != styleTokenNone {
		current := ins.peek(0)
		if current.Type != StyleTokenString && current.Type != StyleTokenComma {
			return result.String()
		}
		result.WriteString(current.Data)
		ins.jump()
	}
	return result.String()
}

func readValue(ins *styleTokenPeeker) (string, error) {
	var result strings.Builder
	for ins.peek(0).Type != styleTokenNone {
		current := ins.peek(0)
		switch current.Type {
		case StyleTokenNewline, StyleTokenSemicolon, StyleTokenCloseBracket:
			return result.String(), nil
		case StyleTokenString:
			if result.Len() > 0 {
				result.WriteByte(' ')
			}
			result.WriteString(current.Data)
			ins.jump()
		case StyleTokenComma:
			result.WriteString(current.Data)
			ins.jump()
		case StyleTokenColon:
			result.WriteString(current.Data)
			ins.jump()
			if ins.typeOf(0) == StyleTokenString {
				result.WriteString(ins.peek(0).Data)
				ins.jump()
			} else {
				return "", &StyleParsingError{Message: "bad definition"}
			}
		default:
			return "", &StyleParsingError{Message: "bad definition"}
		}
	}
	return result.String(), nil
}

func skipNewLines(ins *styleTokenPeeker) {
	for {
		token := ins.peek(0)
		if token.Type == styleTokenNone || token.Type != StyleTokenNewline {
			return
		}
		ins.jump()
	}
}

func skipColon(ins *styleTokenPeeker) {
	for {
		token := ins.peek(0)
		if token.Type == styleTokenNone || token.Type != StyleTokenColon {
			return
		}
		ins.jump()
	}
}

func (p *StyleParser) tokenize(ins *charInspector) ([]StyleToken, error) {
	var result []StyleToken
	for {
		current := ins.peek(0)
		if current == 0 {
			break
		}
		switch {
		case current == ' ' || current == '\t':
			ins.jump()
		case current == '/' && ins.peek(1) == '/':
			jumpUntil(ins, '\n')
		case current == '/' && ins.peek(1) == '\'':
			jumpUntil2(ins, '\'', '/')
		case current == '/' && ins.peek(1) == '*':
			jumpUntil2(ins, '*', '/')
		case current == ',':
			result = append(result, StyleToken{StyleTokenComma, ","})
			ins.jump()
		case current == ';':
			result = append(result, StyleToken{StyleTokenSemicolon, ";"})
			ins.jump()
		case current == '\n' || current == '\r':
			result = append(result, StyleToken{StyleTokenNewline, "NEWLINE"})
			ins.jump()
		case current == '*':
			result = append(result, StyleToken{StyleTokenStar, StyleSignatureStar})
			ins.jump()
		case current == ':':
			result = append(result, StyleToken{StyleTokenColon, ":"})
			ins.jump()
		case current == '{':
			result = append(result, StyleToken{StyleTokenOpenBracket, "{"})
			ins.jump()
		case current == '}':
			result = append(result, StyleToken{StyleTokenCloseBracket, "}"})
			ins.jump()
		case current == '@':
			result = append(result, StyleToken{StyleTokenArobaseMedia, readArobaseMedia(ins)})
		case current == '"':
			result = append(result, StyleToken{StyleTokenString, readQuotedString(ins)})
		default:
			result = append(result, StyleToken{StyleTokenString, readString(ins)})
		}
	}
	return result, nil
}

func jumpUntil(ins *charInspector, ch1 rune) {
	for ins.peek(0) != 0 {
		if ins.peek(0) == ch1 {
			ins.jump()
			return
		}
		ins.jump()
	}
}

func jumpUntil2(ins *charInspector, ch1, ch2 rune) {
	for ins.peek(0) != 0 {
		if ins.peek(0) == ch1 && ins.peek(1) == ch2 {
			ins.jump()
			ins.jump()
			return
		}
		ins.jump()
	}
}

func readArobaseMedia(ins *charInspector) string {
	if ins.peek(0) != '@' {
		throw(illegalState, "")
	}
	ins.jump()
	var result strings.Builder
	for ins.peek(0) != 0 {
		ch := ins.peek(0)
		ins.jump()
		if ch == '{' || ch == '}' || ch == ';' {
			break
		}
		result.WriteRune(ch)
	}
	return result.String()
}

func readQuotedString(ins *charInspector) string {
	var result strings.Builder
	if ins.peek(0) != '"' {
		throw(illegalState, "")
	}
	ins.jump()
	for ins.peek(0) != 0 && ins.peek(0) != '"' {
		ch := ins.peek(0)
		ins.jump()
		result.WriteRune(ch)
	}
	if ins.peek(0) == '"' {
		ins.jump()
	}
	return result.String()
}

func readString(ins *charInspector) string {
	var result []rune
	for ins.peek(0) != 0 {
		ch := ins.peek(0)
		if ch == '\n' || ch == '\r' {
			break
		}
		if ch == ' ' && result[0] != '.' {
			break
		}
		if ch == '{' || ch == '}' || ch == ';' || ch == ',' || ch == ':' || ch == '\t' {
			break
		}
		ins.jump()
		result = append(result, ch)
	}
	if len(result) == 0 {
		throw("java.lang.StringIndexOutOfBoundsException", "index 0,length 0")
	}
	if result[0] == '.' {
		return strings.TrimFunc(string(result), func(r rune) bool { return r <= ' ' })
	}
	return string(result)
}

// StyleLoaderDeltaPriorityForStereotype is added to the priority of the
// values of stereotype styles, so that they win over the others.
const StyleLoaderDeltaPriorityForStereotype = 1000

// addPriorityForStereotype is upstream's
// StyleLoader.addPriorityForStereotype.
func addPriorityForStereotype(tmp styleValues) styleValues {
	var result styleValues
	for k, v := range tmp {
		if v != nil {
			result[k] = asValueImpl(v).AddPriority(StyleLoaderDeltaPriorityForStereotype)
		}
	}
	return result
}
