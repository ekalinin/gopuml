// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonObject.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"io"
	"iter"
	"slices"
	"unicode/utf16"
)

// Object is a JSON object: an ordered sequence of members, each a name and
// a value. The zero value is an empty object ready to use.
//
// Members keep their insertion order, which is also the order of the JSON
// text. Names may repeat: [Object.Add] always appends a member, while
// [Object.Set] replaces the value of the last member with the name, if any.
// [Object.Get] and [Object.Remove] use the last member with the name too.
//
// An unmodifiable view created by [UnmodifiableObject] panics with an
// [*UnsupportedOperationError] on any modification.
type Object struct {
	base
	names  []string
	values []Value
	table  hashIndexTable
	// src is the object an unmodifiable view reads through, nil otherwise.
	src *Object
}

// Member is a member of an object: a name and its value.
type Member struct {
	Name  string
	Value Value
}

// Equal reports whether both members have the same name and equal values.
func (m Member) Equal(other Member) bool {
	return m.Name == other.Name && m.Value.Equal(other.Value)
}

// NewObject returns a new empty object.
func NewObject() *Object { return &Object{} }

// CopyObject returns a new object with the members of o. The values
// themselves are shared; use CloneMe for a deep copy.
func CopyObject(o *Object) *Object {
	copied := &Object{names: slices.Clone(o.nameList()), values: slices.Clone(o.valueList())}
	copied.updateHashIndex()
	return copied
}

// UnmodifiableObject returns an unmodifiable view of o. The view reflects
// later changes of o; modifying it panics with an
// [*UnsupportedOperationError].
func UnmodifiableObject(o *Object) *Object {
	view := &Object{src: o}
	view.updateHashIndex()
	return view
}

// CloneMe returns a deep copy of the object. Added for PlantUML.
func (o *Object) CloneMe() Value {
	clone := &Object{}
	names, values := o.nameList(), o.valueList()
	for i := range names {
		clone.Add(names[i], values[i].CloneMe())
	}
	return clone
}

// nameList returns the member names, read through the source of a view.
func (o *Object) nameList() []string {
	if o.src != nil {
		return o.src.nameList()
	}
	return o.names
}

// valueList returns the member values, read through the source of a view.
func (o *Object) valueList() []Value {
	if o.src != nil {
		return o.src.valueList()
	}
	return o.values
}

// checkModifiable panics for an unmodifiable view.
func (o *Object) checkModifiable() {
	if o.src != nil {
		panic(&UnsupportedOperationError{})
	}
}

// Add appends a member with the given name and value, even if the object
// already has a member with this name, and returns the object. It panics if
// value is nil.
func (o *Object) Add(name string, value Value) *Object {
	if value == nil {
		panic("json: value is null")
	}
	o.table.add(name, len(o.nameList()))
	o.checkModifiable()
	o.names = append(o.names, name)
	o.values = append(o.values, value)
	return o
}

// Set replaces the value of the last member with the given name, or appends
// a new member if there is none, and returns the object. It panics if value
// is nil.
func (o *Object) Set(name string, value Value) *Object {
	if value == nil {
		panic("json: value is null")
	}
	index := o.indexOf(name)
	if index != -1 {
		o.checkModifiable()
		o.values[index] = value
	} else {
		o.table.add(name, len(o.nameList()))
		o.checkModifiable()
		o.names = append(o.names, name)
		o.values = append(o.values, value)
	}
	return o
}

// Remove removes the last member with the given name, if any, and returns
// the object.
func (o *Object) Remove(name string) *Object {
	index := o.indexOf(name)
	if index != -1 {
		o.table.remove(index)
		o.checkModifiable()
		o.names = slices.Delete(o.names, index, index+1)
		o.values = slices.Delete(o.values, index, index+1)
	}
	return o
}

// Contains reports whether the object has a member with the given name.
func (o *Object) Contains(name string) bool {
	return slices.Contains(o.nameList(), name)
}

// Merge copies every member of other into this object with [Object.Set], in
// order, and returns this object.
func (o *Object) Merge(other *Object) *Object {
	// Index over the live lists, like the Java iterator, so that merging an
	// object into itself sees the values already replaced.
	for i := 0; i < len(other.nameList()); i++ {
		o.Set(other.nameList()[i], other.valueList()[i])
	}
	return o
}

// DeepMerge copies every member of other into this object, in order, and
// returns this object. When both the new and the existing value of a name are
// objects, the existing object is deep-merged with the new one instead of
// being replaced.
func (o *Object) DeepMerge(other *Object) *Object {
	for i := 0; i < len(other.nameList()); i++ {
		name, value := other.nameList()[i], other.valueList()[i]
		if obj, ok := value.(*Object); ok {
			if existing, ok := o.Get(name).(*Object); ok {
				value = existing.DeepMerge(obj)
			}
		}
		o.Set(name, value)
	}
	return o
}

// Get returns the value of the last member with the given name, or nil if
// there is none.
func (o *Object) Get(name string) Value {
	index := o.indexOf(name)
	if index == -1 {
		return nil
	}
	return o.valueList()[index]
}

// GetInt returns the value of the last member with the given name as an
// int (see [Value.AsInt]), or defaultValue if there is none.
func (o *Object) GetInt(name string, defaultValue int) (int, error) {
	if v := o.Get(name); v != nil {
		return v.AsInt()
	}
	return defaultValue, nil
}

// GetLong returns the value of the last member with the given name as an
// int64 (see [Value.AsLong]), or defaultValue if there is none.
func (o *Object) GetLong(name string, defaultValue int64) (int64, error) {
	if v := o.Get(name); v != nil {
		return v.AsLong()
	}
	return defaultValue, nil
}

// GetFloat returns the value of the last member with the given name as a
// float32, or defaultValue if there is none. It panics if the value is not
// a number.
func (o *Object) GetFloat(name string, defaultValue float32) float32 {
	if v := o.Get(name); v != nil {
		return v.AsFloat()
	}
	return defaultValue
}

// GetDouble returns the value of the last member with the given name as a
// float64, or defaultValue if there is none. It panics if the value is not
// a number.
func (o *Object) GetDouble(name string, defaultValue float64) float64 {
	if v := o.Get(name); v != nil {
		return v.AsDouble()
	}
	return defaultValue
}

// GetBoolean returns the value of the last member with the given name as a
// bool, or defaultValue if there is none. It panics if the value is neither
// true nor false.
func (o *Object) GetBoolean(name string, defaultValue bool) bool {
	if v := o.Get(name); v != nil {
		return v.AsBoolean()
	}
	return defaultValue
}

// GetString returns the value of the last member with the given name as a
// string, or defaultValue if there is none. It panics if the value is not a
// string.
func (o *Object) GetString(name string, defaultValue string) string {
	if v := o.Get(name); v != nil {
		return v.AsString()
	}
	return defaultValue
}

// Size returns the number of members.
func (o *Object) Size() int { return len(o.nameList()) }

// IsEmpty reports whether the object has no members.
func (o *Object) IsEmpty() bool { return len(o.nameList()) == 0 }

// Names returns a copy of the member names, in order, duplicates included.
func (o *Object) Names() []string { return slices.Clone(o.nameList()) }

// All returns an iterator over the member names and values, in order.
func (o *Object) All() iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		names, values := o.nameList(), o.valueList()
		for i := range names {
			if !yield(names[i], values[i]) {
				return
			}
		}
	}
}

// Members returns an iterator over the members, in order.
func (o *Object) Members() iter.Seq[Member] {
	return func(yield func(Member) bool) {
		for name, value := range o.All() {
			if !yield(Member{Name: name, Value: value}) {
				return
			}
		}
	}
}

func (o *Object) write(w jsonWriter) {
	w.writeObjectOpen()
	names, values := o.nameList(), o.valueList()
	for i := range names {
		if i > 0 {
			w.writeObjectSeparator()
		}
		w.writeMemberName(names[i])
		w.writeMemberSeparator()
		values[i].write(w)
	}
	w.writeObjectClose()
}

func (o *Object) IsObject() bool    { return true }
func (o *Object) AsObject() *Object { return o }

func (o *Object) AsArray() *Array        { unsupported(notAnArray, o); return nil }
func (o *Object) AsInt() (int, error)    { unsupported(notANumber, o); return 0, nil }
func (o *Object) AsLong() (int64, error) { unsupported(notANumber, o); return 0, nil }
func (o *Object) AsFloat() float32       { unsupported(notANumber, o); return 0 }
func (o *Object) AsDouble() float64      { unsupported(notANumber, o); return 0 }
func (o *Object) AsString() string       { unsupported(notAString, o); return "" }
func (o *Object) AsBoolean() bool        { unsupported(notABoolean, o); return false }

func (o *Object) String() string                          { return stringConfig(o, Minimal) }
func (o *Object) StringConfig(config WriterConfig) string { return stringConfig(o, config) }
func (o *Object) WriteJSON(w io.Writer, config WriterConfig) error {
	return writeJSON(o, w, config)
}

// Equal reports whether other is an object with the same names and equal
// values in the same order.
func (o *Object) Equal(other Value) bool {
	x, ok := other.(*Object)
	if !ok || x == nil {
		return false
	}
	if o == x {
		return true
	}
	return slices.Equal(o.nameList(), x.nameList()) &&
		slices.EqualFunc(o.valueList(), x.valueList(), func(a, b Value) bool { return a.Equal(b) })
}

// indexOf returns the index of the last member with the given name, or -1.
func (o *Object) indexOf(name string) int {
	names := o.nameList()
	index := o.table.get(name)
	// Upstream reads names[index] unchecked; the bound check only matters for
	// a view whose source has since shrunk, where Java would throw.
	if index != -1 && index < len(names) && name == names[index] {
		return index
	}
	for i := len(names) - 1; i >= 0; i-- {
		if names[i] == name {
			return i
		}
	}
	return -1
}

// updateHashIndex indexes every current member name.
func (o *Object) updateHashIndex() {
	for i, name := range o.nameList() {
		o.table.add(name, i)
	}
}

// hashIndexTable maps the hash slot of a name to the index, plus one, of the
// last member added with a name in that slot, as long as the index is below
// 255. Zero stands for empty. It only speeds up the lookups: indexOf checks
// every hit and falls back to a linear search.
type hashIndexTable [32]byte // the length must be a power of two

func (t *hashIndexTable) add(name string, index int) {
	slot := hashSlotFor(name)
	if index < 0xff {
		// increment by 1, 0 stands for empty
		t[slot] = byte(index + 1)
	} else {
		t[slot] = 0
	}
}

func (t *hashIndexTable) remove(index int) {
	for i := range t {
		switch {
		case int(t[i]) == index+1:
			t[i] = 0
		case int(t[i]) > index+1:
			t[i]--
		}
	}
}

func (t *hashIndexTable) get(name string) int {
	// subtract 1, 0 stands for empty
	return int(t[hashSlotFor(name)]) - 1
}

func hashSlotFor(name string) int {
	return int(javaStringHash(name) & (int32(len(hashIndexTable{})) - 1))
}

// javaStringHash returns Java's String.hashCode of s, computed over its
// UTF-16 code units.
func javaStringHash(s string) int32 {
	var h int32
	for _, r := range s {
		if r >= 0x10000 {
			r1, r2 := utf16.EncodeRune(r)
			h = 31*h + r1
			h = 31*h + r2
			continue
		}
		h = 31*h + r
	}
	return h
}
