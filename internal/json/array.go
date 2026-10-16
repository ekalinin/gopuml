// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonArray.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"io"
	"iter"
	"slices"
	"strconv"
)

// Array is a JSON array: an ordered sequence of values. The zero value is an
// empty array ready to use.
//
// Elements are added with [Array.Add], replaced with [Array.Set] and removed
// with [Array.Remove]. The methods that take an index return an
// [*IndexOutOfBoundsError] for an index out of range. An unmodifiable view
// created by [UnmodifiableArray] panics with an [*UnsupportedOperationError]
// on any modification.
type Array struct {
	base
	values []Value
	// src is the array an unmodifiable view reads through, nil otherwise.
	src *Array
}

// IndexOutOfBoundsError is returned for an index out of range, like Java's
// IndexOutOfBoundsException.
type IndexOutOfBoundsError struct {
	Index  int
	Length int
}

// Error returns Java's message, for example "Index 3 out of bounds for
// length 2".
func (e *IndexOutOfBoundsError) Error() string {
	return "Index " + strconv.Itoa(e.Index) + " out of bounds for length " + strconv.Itoa(e.Length)
}

// NewArray returns a new array holding the given values, in order. It panics
// if a value is nil.
func NewArray(values ...Value) *Array {
	a := &Array{}
	for _, v := range values {
		a.Add(v)
	}
	return a
}

// CopyArray returns a new array with the elements of a. The elements
// themselves are shared; use CloneMe for a deep copy.
func CopyArray(a *Array) *Array {
	return &Array{values: slices.Clone(a.list())}
}

// UnmodifiableArray returns an unmodifiable view of a. The view reflects
// later changes of a; modifying it panics with an
// [*UnsupportedOperationError].
func UnmodifiableArray(a *Array) *Array {
	return &Array{src: a}
}

// CloneMe returns a deep copy of the array. Added for PlantUML.
func (a *Array) CloneMe() Value {
	values := a.list()
	clone := &Array{values: make([]Value, 0, len(values))}
	for _, v := range values {
		clone.values = append(clone.values, v.CloneMe())
	}
	return clone
}

// list returns the elements, read through the source of a view.
func (a *Array) list() []Value {
	if a.src != nil {
		return a.src.list()
	}
	return a.values
}

// checkModifiable panics for an unmodifiable view.
func (a *Array) checkModifiable() {
	if a.src != nil {
		panic(&UnsupportedOperationError{})
	}
}

// checkIndex returns an error unless index is in [0, length).
func checkIndex(index, length int) error {
	if index < 0 || index >= length {
		return &IndexOutOfBoundsError{Index: index, Length: length}
	}
	return nil
}

// Add appends value to the end of the array and returns the array. It
// panics if value is nil.
func (a *Array) Add(value Value) *Array {
	if value == nil {
		panic("json: value is null")
	}
	a.checkModifiable()
	a.values = append(a.values, value)
	return a
}

// Set replaces the element at index with value. It panics if value is nil.
func (a *Array) Set(index int, value Value) error {
	if value == nil {
		panic("json: value is null")
	}
	a.checkModifiable()
	if err := checkIndex(index, len(a.values)); err != nil {
		return err
	}
	a.values[index] = value
	return nil
}

// Remove removes the element at index.
func (a *Array) Remove(index int) error {
	a.checkModifiable()
	if err := checkIndex(index, len(a.values)); err != nil {
		return err
	}
	a.values = slices.Delete(a.values, index, index+1)
	return nil
}

// Size returns the number of elements.
func (a *Array) Size() int { return len(a.list()) }

// IsEmpty reports whether the array has no elements.
func (a *Array) IsEmpty() bool { return len(a.list()) == 0 }

// Get returns the element at index.
func (a *Array) Get(index int) (Value, error) {
	values := a.list()
	if err := checkIndex(index, len(values)); err != nil {
		return nil, err
	}
	return values[index], nil
}

// Values returns a copy of the elements, in order.
func (a *Array) Values() []Value { return slices.Clone(a.list()) }

// All returns an iterator over the indexes and elements, in order.
func (a *Array) All() iter.Seq2[int, Value] {
	return func(yield func(int, Value) bool) {
		for i, v := range a.list() {
			if !yield(i, v) {
				return
			}
		}
	}
}

func (a *Array) write(w jsonWriter) {
	w.writeArrayOpen()
	for i, v := range a.list() {
		if i > 0 {
			w.writeArraySeparator()
		}
		v.write(w)
	}
	w.writeArrayClose()
}

func (a *Array) IsArray() bool   { return true }
func (a *Array) AsArray() *Array { return a }

func (a *Array) AsObject() *Object      { unsupported(notAnObject, a); return nil }
func (a *Array) AsInt() (int, error)    { unsupported(notANumber, a); return 0, nil }
func (a *Array) AsLong() (int64, error) { unsupported(notANumber, a); return 0, nil }
func (a *Array) AsFloat() float32       { unsupported(notANumber, a); return 0 }
func (a *Array) AsDouble() float64      { unsupported(notANumber, a); return 0 }
func (a *Array) AsString() string       { unsupported(notAString, a); return "" }
func (a *Array) AsBoolean() bool        { unsupported(notABoolean, a); return false }

func (a *Array) String() string                          { return stringConfig(a, Minimal) }
func (a *Array) StringConfig(config WriterConfig) string { return stringConfig(a, config) }
func (a *Array) WriteJSON(w io.Writer, config WriterConfig) error {
	return writeJSON(a, w, config)
}

// Equal reports whether other is an array with equal elements in the same
// order.
func (a *Array) Equal(other Value) bool {
	o, ok := other.(*Array)
	if !ok || o == nil {
		return false
	}
	if a == o {
		return true
	}
	return slices.EqualFunc(a.list(), o.list(), func(x, y Value) bool { return x.Equal(y) })
}
