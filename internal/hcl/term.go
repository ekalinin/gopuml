// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/hcl/HclTerm.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/hcl/SymbolType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package hcl

// symbolType is the kind of an HCL token.
type symbolType int

// The kinds of HCL tokens.
const (
	stringSimple symbolType = iota
	stringQuoted
	functionName
	squareBracketOpen
	squareBracketClose
	curlyBracketOpen
	curlyBracketClose
	parenthesisOpen
	parenthesisClose
	twoPoints
	equals
	comma
	space
)

var symbolTypeNames = [...]string{"STRING_SIMPLE", "STRING_QUOTED", "FUNCTION_NAME", "SQUARE_BRACKET_OPEN",
	"SQUARE_BRACKET_CLOSE", "CURLY_BRACKET_OPEN", "CURLY_BRACKET_CLOSE", "PARENTHESIS_OPEN", "PARENTHESIS_CLOSE",
	"TWO_POINTS", "EQUALS", "COMMA", "SPACE"}

func (t symbolType) String() string { return symbolTypeNames[t] }

// hclTerm is an HCL token: a symbol, or a string or a function name with its
// text.
type hclTerm struct {
	typ     symbolType
	data    string
	hasData bool
}

func newTerm(typ symbolType) hclTerm { return hclTerm{typ: typ} }

func newTermData(typ symbolType, data string) hclTerm {
	return hclTerm{typ: typ, data: data, hasData: true}
}

// String is upstream's toString: the kind, and the text in parentheses.
func (t hclTerm) String() string {
	if !t.hasData {
		return t.typ.String()
	}
	return t.typ.String() + "(" + t.data + ")"
}

func (t hclTerm) is(types ...symbolType) bool {
	for _, typ := range types {
		if t.typ == typ {
			return true
		}
	}
	return false
}
