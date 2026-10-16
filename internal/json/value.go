// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonValue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"errors"
	"io"
	"strings"
)

// Value is a JSON value: an object, an array, a number, a string, or one of
// the literals true, false and null.
//
// The literals are the variables [Null], [True] and [False]. Objects and
// arrays are [*Object] and [*Array]; numbers, strings and booleans are
// created with the ValueOf* functions. The Is* methods tell the kind of a
// value, and the As* methods return it in the matching Go type.
//
// The As* methods panic with an [*UnsupportedOperationError] when the value
// is not of the requested kind, like upstream's UnsupportedOperationException.
//
// Value is implemented only by this package.
type Value interface {
	// IsObject reports whether the value is a JSON object, an *Object.
	IsObject() bool
	// IsArray reports whether the value is a JSON array, an *Array.
	IsArray() bool
	// IsNumber reports whether the value is a JSON number.
	IsNumber() bool
	// IsString reports whether the value is a JSON string.
	IsString() bool
	// IsBoolean reports whether the value is the literal true or false.
	IsBoolean() bool
	// IsTrue reports whether the value is the literal true.
	IsTrue() bool
	// IsFalse reports whether the value is the literal false.
	IsFalse() bool
	// IsNull reports whether the value is the literal null.
	IsNull() bool

	// AsObject returns the value as an object. It panics if the value is
	// not an object.
	AsObject() *Object
	// AsArray returns the value as an array. It panics if the value is not
	// an array.
	AsArray() *Array
	// AsInt returns the number as a Java int: it must have neither a
	// fraction nor an exponent and must be in the int32 range, otherwise a
	// *NumberFormatError is returned. It panics if the value is not a number.
	AsInt() (int, error)
	// AsLong returns the number as a Java long, with the rules of AsInt and
	// the int64 range. It panics if the value is not a number.
	AsLong() (int64, error)
	// AsFloat returns the number as a float32, ±Inf when it is out of
	// range. It panics if the value is not a number.
	AsFloat() float32
	// AsDouble returns the number as a float64, ±Inf when it is out of
	// range. It panics if the value is not a number.
	AsDouble() float64
	// AsString returns the string. It panics if the value is not a string.
	AsString() string
	// AsBoolean returns the boolean. It panics if the value is neither true
	// nor false.
	AsBoolean() bool

	// String returns the JSON text of the value in its minimal form,
	// without any additional whitespace.
	String() string
	// StringConfig returns the JSON text of the value formatted by config.
	// A nil config means the minimal form.
	StringConfig(config WriterConfig) string
	// WriteJSON writes the JSON text of the value formatted by config to w.
	// A nil config means the minimal form.
	WriteJSON(w io.Writer, config WriterConfig) error

	// Equal reports whether other represents the same JSON text: values of
	// the same kind with the same content. Objects with the same members
	// in a different order are not equal.
	Equal(other Value) bool
	// CloneMe returns a deep copy of the value. Immutable values return
	// themselves; arrays and objects are copied recursively. Added for
	// PlantUML.
	CloneMe() Value

	write(w jsonWriter)
}

// UnsupportedOperationError is the panic value of a wrong-type access, such
// as AsObject on a number, and of a modification of an unmodifiable view. It
// matches upstream's UnsupportedOperationException.
type UnsupportedOperationError struct {
	Msg string
}

func (e *UnsupportedOperationError) Error() string {
	if e.Msg == "" {
		return "unsupported operation"
	}
	return e.Msg
}

// Messages of the wrong-type accesses, completed with the JSON text of the
// value.
const (
	notAnObject = "Not an object: "
	notAnArray  = "Not an array: "
	notANumber  = "Not a number: "
	notAString  = "Not a string: "
	notABoolean = "Not a boolean: "
)

// unsupported panics with the upstream message for a wrong-type access on v.
func unsupported(prefix string, v Value) {
	panic(&UnsupportedOperationError{Msg: prefix + v.String()})
}

// errNilWriter is returned by WriteJSON for a nil writer.
var errNilWriter = errors.New("writer is null")

// writeJSON implements WriteJSON for every value.
func writeJSON(v Value, w io.Writer, config WriterConfig) error {
	if w == nil {
		return errNilWriter
	}
	if config == nil {
		config = Minimal
	}
	buffer := newWritingBuffer(w, 128)
	v.write(config.createWriter(buffer))
	return buffer.flush()
}

// stringConfig implements StringConfig for every value.
func stringConfig(v Value, config WriterConfig) string {
	var sb strings.Builder
	// A strings.Builder never fails.
	_ = writeJSON(v, &sb, config)
	return sb.String()
}

// base holds the defaults shared by every value: it is of no kind.
type base struct{}

func (base) IsObject() bool  { return false }
func (base) IsArray() bool   { return false }
func (base) IsNumber() bool  { return false }
func (base) IsString() bool  { return false }
func (base) IsBoolean() bool { return false }
func (base) IsTrue() bool    { return false }
func (base) IsFalse() bool   { return false }
func (base) IsNull() bool    { return false }
