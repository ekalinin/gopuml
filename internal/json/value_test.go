// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// u2028Escape is the JSON escape of U+2028, spelled out to keep the source
// ASCII.
const u2028Escape = "\\" + "u2028"

// panicMessage runs f and returns the message of its UnsupportedOperationError
// panic, or reports a failure.
func panicMessage(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		e, ok := r.(*UnsupportedOperationError)
		if !ok {
			t.Errorf("panic = %#v; want *UnsupportedOperationError", r)
			return
		}
		msg = e.Msg
	}()
	f()
	return ""
}

func mustParse(t *testing.T, s string) Value {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return v
}

func TestKinds(t *testing.T) {
	type kinds struct{ object, array, number, str, boolean, isTrue, isFalse, null bool }
	tests := []struct {
		v    Value
		want kinds
	}{
		{Null, kinds{null: true}},
		{True, kinds{boolean: true, isTrue: true}},
		{False, kinds{boolean: true, isFalse: true}},
		{ValueOfInt(1), kinds{number: true}},
		{ValueOfString("x"), kinds{str: true}},
		{NewArray(), kinds{array: true}},
		{NewObject(), kinds{object: true}},
		{&Array{}, kinds{array: true}},
		{&Object{}, kinds{object: true}},
	}
	for _, tt := range tests {
		got := kinds{tt.v.IsObject(), tt.v.IsArray(), tt.v.IsNumber(), tt.v.IsString(), tt.v.IsBoolean(), tt.v.IsTrue(), tt.v.IsFalse(), tt.v.IsNull()}
		if got != tt.want {
			t.Errorf("%s: kinds = %+v; want %+v", tt.v, got, tt.want)
		}
	}
	if a := NewArray(); a.AsArray() != a {
		t.Error("AsArray() of an array is not itself")
	}
	if o := NewObject(); o.AsObject() != o {
		t.Error("AsObject() of an object is not itself")
	}
	if !True.AsBoolean() || False.AsBoolean() {
		t.Error("AsBoolean() of the literals is wrong")
	}
	if got := ValueOfString("a\nb").AsString(); got != "a\nb" {
		t.Errorf("AsString() = %q", got)
	}
}

// TestWrongTypeAccess checks the panics and messages of upstream's
// UnsupportedOperationException.
func TestWrongTypeAccess(t *testing.T) {
	str := mustParse(t, `"a\nb`+u2028Escape+`"`)
	accessors := []struct {
		name string
		call func(Value)
	}{
		{"AsObject", func(v Value) { v.AsObject() }},
		{"AsArray", func(v Value) { v.AsArray() }},
		{"AsInt", func(v Value) { _, _ = v.AsInt() }},
		{"AsLong", func(v Value) { _, _ = v.AsLong() }},
		{"AsFloat", func(v Value) { v.AsFloat() }},
		{"AsDouble", func(v Value) { v.AsDouble() }},
		{"AsString", func(v Value) { v.AsString() }},
		{"AsBoolean", func(v Value) { v.AsBoolean() }},
	}
	tests := []struct {
		v    Value
		want map[string]string // accessor name to message; absent means no panic
	}{
		{Null, map[string]string{
			"AsObject": "Not an object: null", "AsArray": "Not an array: null", "AsInt": "Not a number: null",
			"AsLong": "Not a number: null", "AsFloat": "Not a number: null", "AsDouble": "Not a number: null",
			"AsString": "Not a string: null", "AsBoolean": "Not a boolean: null",
		}},
		{True, map[string]string{
			"AsObject": "Not an object: true", "AsArray": "Not an array: true", "AsInt": "Not a number: true",
			"AsLong": "Not a number: true", "AsFloat": "Not a number: true", "AsDouble": "Not a number: true",
			"AsString": "Not a string: true",
		}},
		{str, map[string]string{
			"AsObject": `Not an object: "a\nb` + u2028Escape + `"`, "AsArray": `Not an array: "a\nb` + u2028Escape + `"`,
			"AsInt": `Not a number: "a\nb` + u2028Escape + `"`, "AsLong": `Not a number: "a\nb` + u2028Escape + `"`,
			"AsFloat": `Not a number: "a\nb` + u2028Escape + `"`, "AsDouble": `Not a number: "a\nb` + u2028Escape + `"`,
			"AsBoolean": `Not a boolean: "a\nb` + u2028Escape + `"`,
		}},
		{mustParse(t, "12.50"), map[string]string{
			"AsObject": "Not an object: 12.50", "AsArray": "Not an array: 12.50",
			"AsString": "Not a string: 12.50", "AsBoolean": "Not a boolean: 12.50",
		}},
		{mustParse(t, `[1,"x"]`), map[string]string{
			"AsObject": `Not an object: [1,"x"]`, "AsInt": `Not a number: [1,"x"]`, "AsLong": `Not a number: [1,"x"]`,
			"AsFloat": `Not a number: [1,"x"]`, "AsDouble": `Not a number: [1,"x"]`, "AsString": `Not a string: [1,"x"]`,
			"AsBoolean": `Not a boolean: [1,"x"]`,
		}},
		{mustParse(t, `{"k":[]}`), map[string]string{
			"AsArray": `Not an array: {"k":[]}`, "AsInt": `Not a number: {"k":[]}`, "AsLong": `Not a number: {"k":[]}`,
			"AsFloat": `Not a number: {"k":[]}`, "AsDouble": `Not a number: {"k":[]}`, "AsString": `Not a string: {"k":[]}`,
			"AsBoolean": `Not a boolean: {"k":[]}`,
		}},
	}
	for _, tt := range tests {
		for _, a := range accessors {
			want, panics := tt.want[a.name]
			if !panics {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("%s.%s() panicked: %v", tt.v, a.name, r)
						}
					}()
					a.call(tt.v)
				}()
				continue
			}
			if got := panicMessage(t, func() { a.call(tt.v) }); got != want {
				t.Errorf("%s.%s() panic = %q; want %q", tt.v, a.name, got, want)
			}
		}
	}
}

// nfe marks an expected NumberFormatError in TestNumberConversions.
const nfe = "!"

func TestNumberConversions(t *testing.T) {
	tests := []struct {
		text       string
		asInt      string
		asLong     string
		floatBits  uint32
		doubleBits uint64
	}{
		{"0", "0", "0", 0, 0},
		{"-0", "0", "0", 0x80000000, 0x8000000000000000},
		{"1.5", nfe, nfe, 0x3fc00000, 0x3ff8000000000000},
		{"2147483647", "2147483647", "2147483647", 0x4f000000, 0x41dfffffffc00000},
		{"2147483648", nfe, "2147483648", 0x4f000000, 0x41e0000000000000},
		{"-2147483648", "-2147483648", "-2147483648", 0xcf000000, 0xc1e0000000000000},
		{"-2147483649", nfe, "-2147483649", 0xcf000000, 0xc1e0000000200000},
		{"9223372036854775807", nfe, "9223372036854775807", 0x5f000000, 0x43e0000000000000},
		{"9223372036854775808", nfe, nfe, 0x5f000000, 0x43e0000000000000},
		{"1e2", nfe, nfe, 0x42c80000, 0x4059000000000000},
		{"1E2", nfe, nfe, 0x42c80000, 0x4059000000000000},
		{"12345678901234567890", nfe, nfe, 0x5f2b54aa, 0x43e56a95319d63e1},
		{"0.1", nfe, nfe, 0x3dcccccd, 0x3fb999999999999a},
		{"1e400", nfe, nfe, 0x7f800000, 0x7ff0000000000000},
		{"-1e400", nfe, nfe, 0xff800000, 0xfff0000000000000},
		{"1e-400", nfe, nfe, 0, 0},
		{"-1e-400", nfe, nfe, 0x80000000, 0x8000000000000000},
		{"3.4028236e38", nfe, nfe, 0x7f800000, 0x47effffff514a7bc},
		{"3.4028235e38", nfe, nfe, 0x7f7fffff, 0x47efffffe54daff8},
		{"1e39", nfe, nfe, 0x7f800000, 0x48078287f49c4a1d},
		{"1.0E7", nfe, nfe, 0x4b189680, 0x416312d000000000},
		{"4.9E-324", nfe, nfe, 0, 1},
	}
	for _, tt := range tests {
		v := mustParse(t, tt.text)
		i, err := v.AsInt()
		checkConversion(t, tt.text, "AsInt", fmt.Sprint(i), err, tt.asInt)
		l, err := v.AsLong()
		checkConversion(t, tt.text, "AsLong", fmt.Sprint(l), err, tt.asLong)
		if got := math.Float32bits(v.AsFloat()); got != tt.floatBits {
			t.Errorf("%s.AsFloat() bits = %#x; want %#x", tt.text, got, tt.floatBits)
		}
		if got := math.Float64bits(v.AsDouble()); got != tt.doubleBits {
			t.Errorf("%s.AsDouble() bits = %#x; want %#x", tt.text, got, tt.doubleBits)
		}
	}
}

func checkConversion(t *testing.T, text, method, got string, err error, want string) {
	t.Helper()
	if want == nfe {
		nfe, ok := errors.AsType[*NumberFormatError](err)
		if !ok || nfe.Error() != `For input string: "`+text+`"` {
			t.Errorf("%s.%s() error = %v; want NumberFormatError", text, method, err)
		}
		return
	}
	if err != nil || got != want {
		t.Errorf("%s.%s() = %s, %v; want %s", text, method, got, err, want)
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		a, b Value
		want bool
	}{
		{mustParse(t, "1"), mustParse(t, "1.0"), false},
		{mustParse(t, "1"), ValueOfInt(1), true},
		{mustParse(t, `"1"`), mustParse(t, "1"), false},
		{mustParse(t, `{"a":1,"b":2}`), mustParse(t, `{"b":2,"a":1}`), false},
		{mustParse(t, `{"a":1,"b":[2]}`), mustParse(t, `{ "a" : 1 , "b" : [ 2 ] }`), true},
		{mustParse(t, `{"a":1}`), mustParse(t, `{"a":1,"a":1}`), false},
		{mustParse(t, `[1,[true]]`), NewArray(ValueOfInt(1), NewArray(True)), true},
		{mustParse(t, `[1]`), mustParse(t, `[1,2]`), false},
		{Null, mustParse(t, "null"), true},
		{Null, False, false},
		{ValueOfString("x"), ValueOfString("x"), true},
		{NewArray(), NewObject(), false},
		{NewObject(), nil, false},
		{ValueOfInt(1), nil, false},
	}
	for _, tt := range tests {
		if got := tt.a.Equal(tt.b); got != tt.want {
			t.Errorf("%v.Equal(%v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCloneMe(t *testing.T) {
	orig := mustParse(t, `{"a":[1,{"b":2}]}`).AsObject()
	clone := orig.CloneMe().AsObject()
	inner, err := clone.Get("a").AsArray().Get(1)
	if err != nil {
		t.Fatal(err)
	}
	inner.AsObject().Set("b", ValueOfInt(3))
	clone.Get("a").AsArray().Add(ValueOfInt(4))
	if got, want := orig.String(), `{"a":[1,{"b":2}]}`; got != want {
		t.Errorf("original = %s; want %s", got, want)
	}
	if got, want := clone.String(), `{"a":[1,{"b":3},4]}`; got != want {
		t.Errorf("clone = %s; want %s", got, want)
	}
	// Immutable values are their own clones.
	for _, v := range []Value{Null, True, ValueOfInt(1), ValueOfString("s")} {
		if v.CloneMe() != v {
			t.Errorf("%v.CloneMe() is not itself", v)
		}
	}
	// The clone of a view is a modifiable copy.
	view := UnmodifiableArray(NewArray(ValueOfInt(1)))
	c := view.CloneMe().AsArray()
	c.Add(ValueOfInt(2))
	if got := c.String(); got != "[1,2]" {
		t.Errorf("clone of view = %s; want [1,2]", got)
	}
}

func TestUnsupportedOperationErrorMessage(t *testing.T) {
	if got := (&UnsupportedOperationError{}).Error(); got != "unsupported operation" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&UnsupportedOperationError{Msg: "Not a string: 1"}).Error(); got != "Not a string: 1" {
		t.Errorf("Error() = %q", got)
	}
}
