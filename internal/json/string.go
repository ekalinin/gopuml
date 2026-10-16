// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonString.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import "io"

// stringValue is a JSON string.
type stringValue struct {
	base
	s string
}

func (s stringValue) IsString() bool   { return true }
func (s stringValue) AsString() string { return s.s }

func (s stringValue) AsObject() *Object      { unsupported(notAnObject, s); return nil }
func (s stringValue) AsArray() *Array        { unsupported(notAnArray, s); return nil }
func (s stringValue) AsInt() (int, error)    { unsupported(notANumber, s); return 0, nil }
func (s stringValue) AsLong() (int64, error) { unsupported(notANumber, s); return 0, nil }
func (s stringValue) AsFloat() float32       { unsupported(notANumber, s); return 0 }
func (s stringValue) AsDouble() float64      { unsupported(notANumber, s); return 0 }
func (s stringValue) AsBoolean() bool        { unsupported(notABoolean, s); return false }

func (s stringValue) String() string                          { return stringConfig(s, Minimal) }
func (s stringValue) StringConfig(config WriterConfig) string { return stringConfig(s, config) }
func (s stringValue) WriteJSON(w io.Writer, config WriterConfig) error {
	return writeJSON(s, w, config)
}

func (s stringValue) Equal(other Value) bool {
	o, ok := other.(stringValue)
	return ok && o.s == s.s
}

func (s stringValue) CloneMe() Value { return s }

func (s stringValue) write(w jsonWriter) { w.writeString(s.s) }
