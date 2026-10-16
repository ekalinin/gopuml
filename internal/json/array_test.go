// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"errors"
	"slices"
	"testing"
)

func TestArrayOperations(t *testing.T) {
	a := NewArray().Add(ValueOfInt(1)).Add(ValueOfString("x")).Add(True)
	if a.Size() != 3 || a.IsEmpty() {
		t.Errorf("Size(), IsEmpty() = %d, %v; want 3, false", a.Size(), a.IsEmpty())
	}
	errTests := []struct {
		name string
		err  error
		want string
	}{
		{"Get(3)", second(a.Get(3)), "Index 3 out of bounds for length 3"},
		{"Get(-1)", second(a.Get(-1)), "Index -1 out of bounds for length 3"},
		{"Set(3)", a.Set(3, ValueOfInt(1)), "Index 3 out of bounds for length 3"},
		{"Remove(5)", a.Remove(5), "Index 5 out of bounds for length 3"},
	}
	for _, tt := range errTests {
		if _, ok := errors.AsType[*IndexOutOfBoundsError](tt.err); !ok || tt.err.Error() != tt.want {
			t.Errorf("%s error = %v; want %q", tt.name, tt.err, tt.want)
		}
	}
	if err := a.Set(0, Null); err != nil {
		t.Fatal(err)
	}
	if err := a.Remove(1); err != nil {
		t.Fatal(err)
	}
	if got := a.String(); got != "[null,true]" {
		t.Errorf("String() = %s; want [null,true]", got)
	}
	v, err := a.Get(1)
	if err != nil || v != True {
		t.Errorf("Get(1) = %v, %v; want true", v, err)
	}
	if !NewArray().IsEmpty() || !(&Array{}).IsEmpty() {
		t.Error("a new array is not empty")
	}
	var zero Array
	zero.Add(ValueOfInt(9))
	if got := zero.String(); got != "[9]" {
		t.Errorf("zero value array = %s; want [9]", got)
	}
}

func second(_ Value, err error) error { return err }

func TestArrayIteration(t *testing.T) {
	a := mustParse(t, `[3,"b",null]`).AsArray()
	var got []string
	for i, v := range a.All() {
		got = append(got, string(rune('0'+i))+"="+v.String())
	}
	if want := []string{"0=3", `1="b"`, "2=null"}; !slices.Equal(got, want) {
		t.Errorf("All() = %v; want %v", got, want)
	}
	for i := range a.All() {
		if i != 0 {
			t.Errorf("early stop: got index %d", i)
		}
		break
	}
	values := a.Values()
	values[0] = Null
	if first, _ := a.Get(0); first.String() != "3" {
		t.Error("Values() is not a copy")
	}
}

func TestArrayCopyAndView(t *testing.T) {
	src := NewArray(ValueOfInt(1))
	view := UnmodifiableArray(src)
	src.Add(ValueOfInt(2))
	if got := view.String(); got != "[1,2]" {
		t.Errorf("view = %s; want [1,2]", got)
	}
	panicMessage(t, func() { view.Add(ValueOfInt(1)) })
	panicMessage(t, func() { _ = view.Set(5, ValueOfInt(1)) })
	panicMessage(t, func() { _ = view.Remove(9) })
	cp := CopyArray(view).Add(ValueOfInt(3))
	if got := cp.String(); got != "[1,2,3]" {
		t.Errorf("copy = %s; want [1,2,3]", got)
	}
	if got := view.String(); got != "[1,2]" {
		t.Errorf("view after copy = %s; want [1,2]", got)
	}
	if v, err := view.Get(1); err != nil || v.String() != "2" {
		t.Errorf("view.Get(1) = %v, %v", v, err)
	}
	if !view.Equal(src) || view.Size() != 2 {
		t.Error("view does not reflect its source")
	}
	// CopyArray is shallow.
	inner := NewArray()
	shallow := CopyArray(NewArray(inner))
	inner.Add(True)
	if got := shallow.String(); got != "[[true]]" {
		t.Errorf("shallow copy = %s; want [[true]]", got)
	}
}

func TestArrayNilValuePanics(t *testing.T) {
	for name, f := range map[string]func(){
		"Add":      func() { NewArray().Add(nil) },
		"Set":      func() { _ = NewArray(True).Set(0, nil) },
		"NewArray": func() { NewArray(nil) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s(nil) did not panic", name)
				}
			}()
			f()
		}()
	}
}
