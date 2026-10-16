// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/TokenOperator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// TokenOperator is a binary operator, with C-like precedence
// (https://en.cppreference.com/w/c/language/operator_precedence).
type TokenOperator struct {
	name       string
	precedence int
	display    string
	operate    func(v1, v2 *TValue) *TValue
}

// The operators are constants of the program.
var (
	opMultiplication      = &TokenOperator{"MULTIPLICATION", 100 - 3, "*", (*TValue).Multiply}
	opDivision            = &TokenOperator{"DIVISION", 100 - 3, "/", (*TValue).DividedBy}
	opAddition            = &TokenOperator{"ADDITION", 100 - 4, "+", (*TValue).Add}
	opSubstraction        = &TokenOperator{"SUBSTRACTION", 100 - 4, string(rune(commercialMinusSign)), (*TValue).Minus}
	opLessThan            = &TokenOperator{"LESS_THAN", 100 - 6, "<", (*TValue).LessThan}
	opGreaterThan         = &TokenOperator{"GREATER_THAN", 100 - 6, ">", (*TValue).GreaterThan}
	opLessThanOrEquals    = &TokenOperator{"LESS_THAN_OR_EQUALS", 100 - 6, "<=", (*TValue).LessThanOrEquals}
	opGreaterThanOrEquals = &TokenOperator{"GREATER_THAN_OR_EQUALS", 100 - 6, ">=", (*TValue).GreaterThanOrEquals}
	opEquals              = &TokenOperator{"EQUALS", 100 - 7, "==", (*TValue).EqualsOperation}
	opNotEquals           = &TokenOperator{"NOT_EQUALS", 100 - 7, "!=", (*TValue).NotEquals}
	opLogicalAnd          = &TokenOperator{"LOGICAL_AND", 100 - 11, "&&", (*TValue).LogicalAnd}
	opLogicalOr           = &TokenOperator{"LOGICAL_OR", 100 - 12, "||", (*TValue).LogicalOr}
)

// getTokenOperator returns the operator starting with ch (followed by ch2),
// or nil.
func getTokenOperator(ch, ch2 uint16) *TokenOperator {
	switch ch {
	case '*':
		return opMultiplication
	case '/':
		return opDivision
	case '+':
		return opAddition
	case commercialMinusSign:
		return opSubstraction
	case '<':
		if ch2 == '=' {
			return opLessThanOrEquals
		}
		return opLessThan
	case '>':
		if ch2 == '=' {
			return opGreaterThanOrEquals
		}
		return opGreaterThan
	case '=':
		if ch2 == '=' {
			return opEquals
		}
	case '!':
		if ch2 == '=' {
			return opNotEquals
		}
	case '&':
		if ch2 == '&' {
			return opLogicalAnd
		}
	case '|':
		if ch2 == '|' {
			return opLogicalOr
		}
	}
	return nil
}

// displayLength is the length of the operator text in chars.
func (op *TokenOperator) displayLength() int { return javaLength(op.display) }
