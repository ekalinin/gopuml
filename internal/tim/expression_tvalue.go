// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/expression/TValue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/json"
)

// TValue is a preprocessor value: a Java int, a string or a JSON value. A
// nil *TValue is upstream's null (for example an undefined variable).
// Values are immutable.
type TValue struct {
	intValue    int32
	stringValue string
	isString    bool
	jsonValue   json.Value
}

// TValueFromInt returns an int value.
func TValueFromInt(v int32) *TValue { return &TValue{intValue: v} }

// TValueFromBoolean returns 1 or 0.
func TValueFromBoolean(b bool) *TValue {
	if b {
		return TValueFromInt(1)
	}
	return TValueFromInt(0)
}

// TValueFromJSON returns a JSON value. A nil JSON value gives the int 0, as
// upstream's TValue(null) does.
func TValueFromJSON(v json.Value) *TValue { return &TValue{jsonValue: v} }

// TValueFromString returns a string value.
func TValueFromString(s string) *TValue { return &TValue{stringValue: s, isString: true} }

// tValueFromStringToken is TValue.fromString(Token).
func tValueFromStringToken(token *Token) *TValue {
	if token.tokenType != tokenQuotedString {
		throwRuntime("IllegalArgumentException", "")
	}
	return TValueFromString(token.surface)
}

// tValueFromNumberToken is TValue.fromNumber(Token): Integer.parseInt of the
// token, which throws NumberFormatException on overflow.
func tValueFromNumberToken(token *Token) *TValue {
	if token.tokenType != tokenNumber {
		throwRuntime("IllegalArgumentException", "")
	}
	v, ok := javaParseInt(token.surface)
	if !ok {
		throwRuntime("NumberFormatException", `For input string: "`+token.surface+`"`)
	}
	return TValueFromInt(v)
}

// String is TValue.toString.
func (v *TValue) String() string {
	if v.jsonValue != nil && v.jsonValue.IsString() {
		return v.jsonValue.AsString()
	}
	if v.jsonValue != nil {
		return v.jsonValue.String()
	}
	if !v.isString {
		return strconv.Itoa(int(v.intValue))
	}
	return v.stringValue
}

// Add is the + operator: int addition (wrapping like Java) or concatenation.
func (v *TValue) Add(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromInt(v.intValue + v2.intValue)
	}
	return TValueFromString(v.String() + v2.String())
}

// Minus is the - operator; on strings it concatenates, as upstream does.
func (v *TValue) Minus(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromInt(v.intValue - v2.intValue)
	}
	return TValueFromString(v.String() + v2.String())
}

// Multiply is the * operator.
func (v *TValue) Multiply(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromInt(v.intValue * v2.intValue)
	}
	return TValueFromString(v.String() + "*" + v2.String())
}

// DividedBy is the / operator. A division by zero throws
// ArithmeticException, as in Java.
func (v *TValue) DividedBy(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		if v2.intValue == 0 {
			throwRuntime("ArithmeticException", "/ by zero")
		}
		return TValueFromInt(v.intValue / v2.intValue)
	}
	return TValueFromString(v.String() + "/" + v2.String())
}

// IsNumber reports whether the value is an int.
func (v *TValue) IsNumber() bool { return v.jsonValue == nil && !v.isString }

// IsJSON reports whether the value is a JSON value.
func (v *TValue) IsJSON() bool { return v.jsonValue != nil }

// IsString reports whether the value is a string.
func (v *TValue) IsString() bool { return v.isString }

// toToken is TValue.toToken.
func (v *TValue) toToken() *Token {
	if v.IsNumber() {
		return newToken(v.String(), tokenNumber, nil)
	}
	if v.IsJSON() {
		return newToken(v.String(), tokenJSONData, v.jsonValue)
	}
	return newToken(v.String(), tokenQuotedString, nil)
}

func (v *TValue) compare(v2 *TValue) int { return javaCompare(v.String(), v2.String()) }

// GreaterThanOrEquals is the >= operator.
func (v *TValue) GreaterThanOrEquals(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue >= v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) >= 0)
}

// GreaterThan is the > operator.
func (v *TValue) GreaterThan(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue > v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) > 0)
}

// LessThanOrEquals is the <= operator.
func (v *TValue) LessThanOrEquals(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue <= v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) <= 0)
}

// LessThan is the < operator.
func (v *TValue) LessThan(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue < v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) < 0)
}

// EqualsOperation is the == operator.
func (v *TValue) EqualsOperation(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue == v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) == 0)
}

// NotEquals is the != operator.
func (v *TValue) NotEquals(v2 *TValue) *TValue {
	if v.IsNumber() && v2.IsNumber() {
		return TValueFromBoolean(v.intValue != v2.intValue)
	}
	return TValueFromBoolean(v.compare(v2) != 0)
}

// ToBoolean is TValue.toBoolean: a non-zero int or a non-empty string.
func (v *TValue) ToBoolean() bool {
	if v.IsNumber() {
		return v.intValue != 0
	}
	return v.String() != ""
}

// ToInt returns the int, 0 for strings and JSON values.
func (v *TValue) ToInt() int32 { return v.intValue }

// LogicalAnd is the && operator.
func (v *TValue) LogicalAnd(v2 *TValue) *TValue {
	return TValueFromBoolean(v.ToBoolean() && v2.ToBoolean())
}

// LogicalOr is the || operator.
func (v *TValue) LogicalOr(v2 *TValue) *TValue {
	return TValueFromBoolean(v.ToBoolean() || v2.ToBoolean())
}

// ToJSON returns the JSON value, nil when the value is not JSON.
func (v *TValue) ToJSON() json.Value { return v.jsonValue }

// ToJSONValue converts the value to JSON: ints and strings become JSON
// numbers and strings.
func (v *TValue) ToJSONValue() json.Value {
	if v.IsNumber() {
		return json.ValueOfInt(int(v.intValue))
	}
	if v.IsString() {
		return json.ValueOfString(v.stringValue)
	}
	return v.jsonValue
}
