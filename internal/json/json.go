// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/Json.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"github.com/ekalinin/gopuml/internal/jfmt"

	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

var (
	// Null is the JSON literal null.
	Null Value = literal{value: "null"}
	// True is the JSON literal true.
	True Value = literal{value: "true"}
	// False is the JSON literal false.
	False Value = literal{value: "false"}
)

// errNotFinite is returned for NaN and infinite floating-point values.
var errNotFinite = errors.New("Infinite and NaN values not permitted in JSON")

// ValueOfInt returns a JSON number for v, written like Java's
// Integer.toString.
func ValueOfInt(v int) Value {
	return number{text: strconv.Itoa(v)}
}

// ValueOfLong returns a JSON number for v, written like Java's
// Long.toString.
func ValueOfLong(v int64) Value {
	return number{text: strconv.FormatInt(v, 10)}
}

// ValueOfFloat returns a JSON number for v, written like Java's
// Float.toString without a trailing ".0" (so 100 is "100" and 1e7 is
// "1.0E7"). It returns an error for NaN and infinite values.
func ValueOfFloat(v float32) (Value, error) {
	f := float64(v)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errNotFinite
	}
	return number{text: cutOffPointZero(javaFloatToString(f, 32))}, nil
}

// ValueOfDouble returns a JSON number for v, written like Java's
// Double.toString without a trailing ".0" (so 100 is "100", 0.001 is
// "0.001" and 1e7 is "1.0E7"). It returns an error for NaN and infinite
// values.
func ValueOfDouble(v float64) (Value, error) {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return nil, errNotFinite
	}
	return number{text: cutOffPointZero(javaFloatToString(v, 64))}, nil
}

// ValueOfString returns a JSON string for s.
func ValueOfString(s string) Value {
	return stringValue{s: s}
}

// ValueOfBool returns [True] or [False].
func ValueOfBool(b bool) Value {
	if b {
		return True
	}
	return False
}

// Parse parses s, which must contain a single JSON value optionally padded
// with whitespace. It returns a [*ParseError] if s is not valid JSON.
//
// Object members keep their order and duplicate names; numbers keep their
// text verbatim.
func Parse(s string) (Value, error) {
	handler := &DefaultHandler{}
	if err := NewParser(handler).Parse(s); err != nil {
		return nil, err
	}
	return handler.Value(), nil
}

// ParseReader reads the entire input from r and parses it like [Parse].
func ParseReader(r io.Reader) (Value, error) {
	handler := &DefaultHandler{}
	if err := NewParser(handler).ParseReader(r); err != nil {
		return nil, err
	}
	return handler.Value(), nil
}

func cutOffPointZero(s string) string {
	return strings.TrimSuffix(s, ".0")
}

// DefaultHandler is the [Handler] that builds a [Value] from the parser
// events. The zero value is ready to use.
type DefaultHandler struct {
	BaseHandler[*Array, *Object]
	value Value
}

func (h *DefaultHandler) StartArray() *Array   { return &Array{} }
func (h *DefaultHandler) StartObject() *Object { return &Object{} }
func (h *DefaultHandler) EndNull()             { h.value = Null }
func (h *DefaultHandler) EndBoolean(b bool)    { h.value = ValueOfBool(b) }
func (h *DefaultHandler) EndString(s string)   { h.value = stringValue{s: s} }
func (h *DefaultHandler) EndNumber(s string)   { h.value = number{text: s} }
func (h *DefaultHandler) EndArray(a *Array)    { h.value = a }
func (h *DefaultHandler) EndObject(o *Object)  { h.value = o }

func (h *DefaultHandler) EndArrayValue(a *Array) { a.Add(h.value) }

func (h *DefaultHandler) EndObjectValue(o *Object, name string) { o.Add(name, h.value) }

// Value returns the last value read, the whole input after a successful
// parse, or nil before any value.
func (h *DefaultHandler) Value() Value { return h.value }

// javaFloatToString returns the text of Java's Double.toString (bitSize 64)
// or Float.toString (bitSize 32) of v; see jfmt.
func javaFloatToString(v float64, bitSize int) string {
	if bitSize == 32 {
		return jfmt.FloatToString(float32(v))
	}
	return jfmt.DoubleToString(v)
}
