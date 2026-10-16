// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonNumber.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"io"
	"strconv"
)

// number is a JSON number. Its text is kept verbatim as parsed or as
// produced by a ValueOf* function.
type number struct {
	base
	text string
}

// NumberFormatError is returned by AsInt and AsLong when the number text is
// not a Java int or long, like upstream's NumberFormatException.
type NumberFormatError struct {
	Input string
}

// Error returns Java's message, for example `For input string: "1.5"`.
func (e *NumberFormatError) Error() string {
	return `For input string: "` + e.Input + `"`
}

func (n number) IsNumber() bool { return true }

// AsInt parses the text like Java's Integer.parseInt(text, 10).
func (n number) AsInt() (int, error) {
	v, err := strconv.ParseInt(n.text, 10, 32)
	if err != nil {
		return 0, &NumberFormatError{Input: n.text}
	}
	return int(v), nil
}

// AsLong parses the text like Java's Long.parseLong(text, 10).
func (n number) AsLong() (int64, error) {
	v, err := strconv.ParseInt(n.text, 10, 64)
	if err != nil {
		return 0, &NumberFormatError{Input: n.text}
	}
	return v, nil
}

// AsFloat parses the text like Java's Float.parseFloat: an out of range
// value gives ±Inf.
func (n number) AsFloat() float32 {
	// The text is valid JSON number syntax, so the only possible error is
	// ErrRange, which comes with the ±Inf or zero that Java returns too.
	v, _ := strconv.ParseFloat(n.text, 32)
	return float32(v)
}

// AsDouble parses the text like Java's Double.parseDouble: an out of range
// value gives ±Inf.
func (n number) AsDouble() float64 {
	v, _ := strconv.ParseFloat(n.text, 64)
	return v
}

func (n number) AsObject() *Object { unsupported(notAnObject, n); return nil }
func (n number) AsArray() *Array   { unsupported(notAnArray, n); return nil }
func (n number) AsString() string  { unsupported(notAString, n); return "" }
func (n number) AsBoolean() bool   { unsupported(notABoolean, n); return false }

func (n number) String() string                          { return n.text }
func (n number) StringConfig(config WriterConfig) string { return stringConfig(n, config) }
func (n number) WriteJSON(w io.Writer, config WriterConfig) error {
	return writeJSON(n, w, config)
}

func (n number) Equal(other Value) bool {
	o, ok := other.(number)
	return ok && o.text == n.text
}

func (n number) CloneMe() Value { return n }

func (n number) write(w jsonWriter) { w.writeNumber(n.text) }
