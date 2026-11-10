// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/hcl/HclParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package hcl

import (
	"strings"
	"unicode"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// The classes of the exceptions that upstream's parser throws.
const (
	IllegalStateException    = "java.lang.IllegalStateException"
	IllegalArgumentException = "java.lang.IllegalArgumentException"
	NoSuchElementException   = "java.util.NoSuchElementException"
)

// Error is an exception of upstream's parser: HCL that is not of the subset
// it reads.
type Error struct {
	Class   string
	Message string
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Class
	}
	return e.Class + ": " + e.Message
}

// throw panics with the exception, which the parser returns.
func throw(class, message string) { panic(&Error{Class: class, Message: message}) }

// catch returns the exception of a panic of throw as *err; other panics go
// on.
func catch(err *error) {
	if v := recover(); v != nil {
		e, ok := v.(*Error)
		if !ok {
			panic(v)
		}
		*err = e
	}
}

// hclParser reads HCL, after https://github.com/hashicorp/hcl, as
// upstream: the tokens of the source, then the blocks and their attributes.
type hclParser struct {
	terms []hclTerm
}

// newHclParser reads the tokens of source.
func newHclParser(source *hclSource) (p *hclParser, err error) {
	defer catch(&err)
	p = &hclParser{}
	p.parse(&iterator[rune]{list: source.all})
	return p, nil
}

// iterator is Java's Iterator on a list: next throws a
// NoSuchElementException past the end.
type iterator[O any] struct {
	list []O
	pos  int
}

func (it *iterator[O]) hasNext() bool { return it.pos < len(it.list) }

func (it *iterator[O]) next() O {
	if !it.hasNext() {
		throw(NoSuchElementException, "")
	}
	it.pos++
	return it.list[it.pos-1]
}

// parseMe returns the data of the blocks: the body of the block when there
// is one, otherwise an object whose keys are the headers of the blocks.
func (p *hclParser) parseMe() (result *json.Object, err error) {
	defer catch(&err)
	var names []string
	blocks := map[string]*json.Object{}
	it := &iterator[hclTerm]{list: p.terms}
	for it.hasNext() {
		name, block := getModuleOrSomething(it)
		if _, ok := blocks[name]; !ok {
			names = append(names, name)
		}
		blocks[name] = block
	}

	if len(names) == 1 {
		return blocks[names[0]], nil
	}

	result = json.NewObject()
	for _, name := range names {
		result.Add(name, blocks[name])
	}
	return result, nil
}

// getModuleOrSomething reads a block: its header, the names and the quoted
// labels before "{", then its body.
func getModuleOrSomething(it *iterator[hclTerm]) (string, *json.Object) {
	var name strings.Builder
	for {
		current := it.next()
		switch {
		case current.is(stringQuoted):
			name.WriteString(`"` + current.data + `" `)
		case current.is(stringSimple):
			name.WriteString(current.data + " ")
		case current.is(curlyBracketOpen):
			return text.JavaTrim(name.String()), getBracketData(it)
		default:
			throw(IllegalStateException, current.String())
		}
	}
}

// getFunctionData reads the arguments of a function call: the text "name()"
// without arguments, otherwise an object whose key "name()" holds the list
// of the arguments.
func getFunctionData(functionName string, it *iterator[hclTerm]) json.Value {
	args := json.NewArray()
	if !it.next().is(parenthesisOpen) {
		throw(IllegalStateException, "")
	}

	for {
		value := getValue(it)
		if t, ok := value.(hclTerm); ok && t.is(parenthesisClose) {
			if args.Size() == 0 {
				return json.ValueOfString(functionName + "()")
			}
			result := json.NewObject()
			result.Add(functionName+"()", args)
			return result
		}
		if t, ok := value.(hclTerm); ok && t.is(comma) {
			continue
		}

		switch v := value.(type) {
		case string:
			args.Add(json.ValueOfString(v))
		case json.Value:
			args.Add(v)
		default:
			throw(IllegalStateException, "")
		}
	}
}

// getBracketData reads the body of a block or an object: attributes up to
// "}".
func getBracketData(it *iterator[hclTerm]) *json.Object {
	result := json.NewObject()
	for {
		current := it.next()
		if current.is(curlyBracketClose) {
			return result
		}
		if !current.is(stringSimple, stringQuoted) {
			throw(IllegalStateException, current.String())
		}
		fieldName := current.data
		if next := it.next(); !next.is(equals, twoPoints) {
			throw(IllegalStateException, current.String())
		}
		switch v := getValue(it).(type) {
		case string:
			result.Add(fieldName, json.ValueOfString(v))
		case json.Value:
			result.Add(fieldName, v)
		default:
			throw(IllegalStateException, "")
		}
	}
}

// getValue reads a value: a comma or a closing parenthesis as the term
// itself, a string, or the JSON value of a list, an object or a function
// call.
func getValue(it *iterator[hclTerm]) any {
	current := it.next()
	switch {
	case current.is(comma, parenthesisClose):
		return current
	case current.is(stringQuoted, stringSimple):
		return current.data
	case current.is(squareBracketOpen):
		return getArray(it)
	case current.is(curlyBracketOpen):
		return getBracketData(it)
	case current.is(functionName):
		return getFunctionData(current.data, it)
	}
	throw(IllegalStateException, current.String())
	return nil
}

// getArray reads a list up to "]": its objects and quoted strings; the other
// values are dropped, as upstream does.
func getArray(it *iterator[hclTerm]) json.Value {
	result := json.NewArray()
	for {
		current := it.next()
		if current.is(curlyBracketOpen) {
			result.Add(getBracketData(it))
		}
		if current.is(squareBracketClose) {
			return result
		}
		if current.is(comma) {
			continue
		}
		if current.is(stringQuoted) {
			result.Add(json.ValueOfString(current.data))
		}
	}
}

// String lists the tokens, as upstream's toString.
func (p *hclParser) String() string {
	parts := make([]string, len(p.terms))
	for i, t := range p.terms {
		parts[i] = t.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// parse reads the tokens: the symbols, the quoted strings, the function
// names before "(" and the other runs of characters, which only a symbol
// or a space ends.
func (p *hclParser) parse(it *iterator[rune]) {
	var pendingString strings.Builder
	for it.hasNext() {
		c := it.next()
		typ, isSymbol := getType(c)
		if isSymbol && typ == parenthesisOpen {
			if pendingString.Len() == 0 {
				throw(IllegalArgumentException, "")
			}
			p.terms = append(p.terms, newTermData(functionName, pendingString.String()))
			pendingString.Reset()
		} else if isSymbol && pendingString.Len() > 0 {
			p.terms = append(p.terms, newTermData(stringSimple, pendingString.String()))
			pendingString.Reset()
		}

		if isSymbol && typ == space {
			continue
		}

		if isSymbol {
			p.terms = append(p.terms, newTerm(typ))
			continue
		}
		if c == '"' {
			p.terms = append(p.terms, newTermData(stringQuoted, eatUntilDoubleQuote(it)))
			continue
		}
		pendingString.WriteRune(c)
	}
}

// eatUntilDoubleQuote reads a quoted string after its opening quote; a
// backslash escapes the next character.
func eatUntilDoubleQuote(it *iterator[rune]) string {
	var sb strings.Builder
	for it.hasNext() {
		c := it.next()
		if c == '\\' {
			sb.WriteRune(it.next())
			continue
		}
		if c == '"' {
			return sb.String()
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

// getType returns the symbol of c; ok is false for the other characters.
// Spaces are Java's space characters: the Unicode space, line and
// paragraph separators, not the tab.
func getType(c rune) (typ symbolType, ok bool) {
	switch {
	case unicode.In(c, unicode.Zs, unicode.Zl, unicode.Zp):
		return space, true
	case c == '{':
		return curlyBracketOpen, true
	case c == '}':
		return curlyBracketClose, true
	case c == '[':
		return squareBracketOpen, true
	case c == ']':
		return squareBracketClose, true
	case c == '(':
		return parenthesisOpen, true
	case c == ')':
		return parenthesisClose, true
	case c == '=':
		return equals, true
	case c == ',':
		return comma, true
	case c == ':':
		return twoPoints, true
	}
	return 0, false
}
