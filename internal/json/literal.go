// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonLiteral.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import "io"

// literal is one of the JSON literals null, true and false.
type literal struct {
	base
	value string
}

func (l literal) IsNull() bool    { return l.value == "null" }
func (l literal) IsTrue() bool    { return l.value == "true" }
func (l literal) IsFalse() bool   { return l.value == "false" }
func (l literal) IsBoolean() bool { return l.IsTrue() || l.IsFalse() }

func (l literal) AsBoolean() bool {
	if l.IsNull() {
		unsupported(notABoolean, l)
	}
	return l.IsTrue()
}

func (l literal) AsObject() *Object      { unsupported(notAnObject, l); return nil }
func (l literal) AsArray() *Array        { unsupported(notAnArray, l); return nil }
func (l literal) AsInt() (int, error)    { unsupported(notANumber, l); return 0, nil }
func (l literal) AsLong() (int64, error) { unsupported(notANumber, l); return 0, nil }
func (l literal) AsFloat() float32       { unsupported(notANumber, l); return 0 }
func (l literal) AsDouble() float64      { unsupported(notANumber, l); return 0 }
func (l literal) AsString() string       { unsupported(notAString, l); return "" }

func (l literal) String() string                          { return l.value }
func (l literal) StringConfig(config WriterConfig) string { return stringConfig(l, config) }
func (l literal) WriteJSON(w io.Writer, config WriterConfig) error {
	return writeJSON(l, w, config)
}

func (l literal) Equal(other Value) bool {
	o, ok := other.(literal)
	return ok && o.value == l.value
}

func (l literal) CloneMe() Value { return l }

func (l literal) write(w jsonWriter) { w.writeLiteral(l.value) }
