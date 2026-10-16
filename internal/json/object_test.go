// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestObjectDuplicates(t *testing.T) {
	o := NewObject().Add("a", ValueOfInt(1)).Add("b", ValueOfInt(2)).Add("a", ValueOfInt(3))
	steps := []struct {
		name string
		do   func()
		want string
	}{
		{"add", func() {}, `{"a":1,"b":2,"a":3}`},
		{"set existing", func() { o.Set("a", ValueOfInt(4)) }, `{"a":1,"b":2,"a":4}`},
		{"set new", func() { o.Set("c", ValueOfInt(5)) }, `{"a":1,"b":2,"a":4,"c":5}`},
		{"remove", func() { o.Remove("a") }, `{"a":1,"b":2,"c":5}`},
		{"remove again", func() { o.Remove("a") }, `{"b":2,"c":5}`},
		{"remove missing", func() { o.Remove("zz") }, `{"b":2,"c":5}`},
	}
	for _, s := range steps {
		if s.name == "add" {
			if got := o.Get("a").String(); got != "3" {
				t.Errorf("Get(a) = %s; want the last value 3", got)
			}
			if o.Get("z") != nil {
				t.Errorf("Get(z) = %v; want nil", o.Get("z"))
			}
		}
		s.do()
		if got := o.String(); got != s.want {
			t.Errorf("%s: %s; want %s", s.name, got, s.want)
		}
	}
	if !o.Contains("b") || o.Contains("a") {
		t.Errorf("Contains(b), Contains(a) = %v, %v; want true, false", o.Contains("b"), o.Contains("a"))
	}
	if got := o.Names(); !slices.Equal(got, []string{"b", "c"}) {
		t.Errorf("Names() = %v; want [b c]", got)
	}
	if o.Size() != 2 || o.IsEmpty() {
		t.Errorf("Size(), IsEmpty() = %d, %v; want 2, false", o.Size(), o.IsEmpty())
	}
	if !NewObject().IsEmpty() || !(&Object{}).IsEmpty() {
		t.Error("a new object is not empty")
	}
}

// TestObjectManyMembers goes past the 255 indexes of the hash table.
func TestObjectManyMembers(t *testing.T) {
	o := NewObject()
	for i := range 300 {
		o.Add("k"+strconv.Itoa(i%7), ValueOfInt(i))
	}
	if got := o.Get("k3").String(); got != "297" {
		t.Errorf("Get(k3) = %s; want 297", got)
	}
	o.Remove("k3")
	if got := o.Get("k3").String(); got != "290" {
		t.Errorf("Get(k3) after Remove = %s; want 290", got)
	}
	o.Set("k0", ValueOfString("x"))
	if got := o.Get("k0").String(); got != `"x"` {
		t.Errorf("Get(k0) after Set = %s; want \"x\"", got)
	}
	if o.Size() != 299 {
		t.Errorf("Size() = %d; want 299", o.Size())
	}
	var k3 []string
	for m := range o.Members() {
		if m.Name == "k3" {
			k3 = append(k3, m.Value.String())
		}
	}
	if got := strings.Join(k3[len(k3)-5:], ","); got != "262,269,276,283,290" {
		t.Errorf("last k3 values = %s; want 262,269,276,283,290", got)
	}
}

func TestObjectLookupAgainstLinearSearch(t *testing.T) {
	// Names that share hash slots, with duplicates, adds and removes.
	o := NewObject()
	names := []string{"a", "b", "a", "Aa", "BB", "c", "Aa", "a", "BB", "d"}
	for i, n := range names {
		o.Add(n, ValueOfInt(i))
	}
	for _, r := range []string{"a", "BB", "zz", "c", "a"} {
		o.Remove(r)
		for _, n := range []string{"a", "b", "Aa", "BB", "c", "d", "zz"} {
			want := -1
			for i, m := range o.names {
				if m == n {
					want = i
				}
			}
			if got := o.indexOf(n); got != want {
				t.Errorf("after Remove(%s): indexOf(%s) = %d; want %d", r, n, got, want)
			}
		}
	}
}

func TestObjectMerge(t *testing.T) {
	m1 := mustParse(t, `{"a":1,"b":{"x":1,"y":2},"c":[1],"a":2}`).AsObject()
	m2 := mustParse(t, `{"b":{"y":3,"z":4},"c":[2],"d":null,"a":5}`).AsObject()
	if got, want := m1.CloneMe().AsObject().Merge(m2).String(), `{"a":1,"b":{"y":3,"z":4},"c":[2],"a":5,"d":null}`; got != want {
		t.Errorf("Merge = %s; want %s", got, want)
	}
	if got, want := m1.CloneMe().AsObject().DeepMerge(m2).String(), `{"a":1,"b":{"x":1,"y":3,"z":4},"c":[2],"a":5,"d":null}`; got != want {
		t.Errorf("DeepMerge = %s; want %s", got, want)
	}
	self := mustParse(t, `{"a":1,"a":2,"o":{"p":1}}`).AsObject()
	if got, want := self.Merge(self).String(), `{"a":1,"a":1,"o":{"p":1}}`; got != want {
		t.Errorf("self Merge = %s; want %s", got, want)
	}
	self = mustParse(t, `{"a":1,"a":2,"o":{"p":1}}`).AsObject()
	if got, want := self.DeepMerge(self).String(), `{"a":1,"a":1,"o":{"p":1}}`; got != want {
		t.Errorf("self DeepMerge = %s; want %s", got, want)
	}
}

func TestObjectGetters(t *testing.T) {
	o := mustParse(t, `{"i":7,"f":1.5,"b":true,"s":"x","big":3000000000}`).AsObject()
	if v, err := o.GetInt("i", 1); v != 7 || err != nil {
		t.Errorf("GetInt(i) = %d, %v", v, err)
	}
	if v, err := o.GetInt("missing", 1); v != 1 || err != nil {
		t.Errorf("GetInt(missing) = %d, %v", v, err)
	}
	if _, err := o.GetInt("big", 1); err == nil || err.Error() != `For input string: "3000000000"` {
		t.Errorf("GetInt(big) error = %v", err)
	}
	if v, err := o.GetLong("big", 1); v != 3000000000 || err != nil {
		t.Errorf("GetLong(big) = %d, %v", v, err)
	}
	if v, err := o.GetLong("missing", -1); v != -1 || err != nil {
		t.Errorf("GetLong(missing) = %d, %v", v, err)
	}
	if v := o.GetFloat("f", 0); v != 1.5 {
		t.Errorf("GetFloat(f) = %v", v)
	}
	if v := o.GetFloat("missing", 2.5); v != 2.5 {
		t.Errorf("GetFloat(missing) = %v", v)
	}
	if v := o.GetDouble("f", 0); v != 1.5 {
		t.Errorf("GetDouble(f) = %v", v)
	}
	if v := o.GetDouble("missing", 2.5); v != 2.5 {
		t.Errorf("GetDouble(missing) = %v", v)
	}
	if !o.GetBoolean("b", false) || !o.GetBoolean("missing", true) {
		t.Error("GetBoolean is wrong")
	}
	if o.GetString("s", "d") != "x" || o.GetString("missing", "d") != "d" {
		t.Error("GetString is wrong")
	}
	if got := panicMessage(t, func() { o.GetString("i", "") }); got != "Not a string: 7" {
		t.Errorf("GetString(i) panic = %q", got)
	}
}

func TestObjectIteration(t *testing.T) {
	o := mustParse(t, `{"b":1,"a":2,"b":3}`).AsObject()
	var got []string
	for name, v := range o.All() {
		got = append(got, name+"="+v.String())
	}
	if want := []string{"b=1", "a=2", "b=3"}; !slices.Equal(got, want) {
		t.Errorf("All() = %v; want %v", got, want)
	}
	for name := range o.All() {
		if name != "b" {
			t.Errorf("early stop: got %s", name)
		}
		break
	}
	for m := range o.Members() {
		if !m.Equal(Member{Name: "b", Value: ValueOfInt(1)}) {
			t.Errorf("first member = %+v", m)
		}
		break
	}
	names := o.Names()
	names[0] = "changed"
	if o.Names()[0] != "b" {
		t.Error("Names() is not a copy")
	}
}

func TestObjectCopy(t *testing.T) {
	orig := mustParse(t, `{"a":[1,{"b":2}]}`).AsObject()
	cp := CopyObject(orig)
	cp.Get("a").AsArray().Add(ValueOfInt(5))
	cp.Add("z", ValueOfInt(0))
	if got, want := orig.String(), `{"a":[1,{"b":2},5]}`; got != want {
		t.Errorf("original = %s; want %s", got, want)
	}
	if got, want := cp.String(), `{"a":[1,{"b":2},5],"z":0}`; got != want {
		t.Errorf("copy = %s; want %s", got, want)
	}
	if got := cp.Get("a").String(); got != "[1,{\"b\":2},5]" {
		t.Errorf("Get(a) on the copy = %s", got)
	}
}

// TestUnmodifiableObject checks the view semantics, including upstream's
// quirks: the view indexes the names when it is created, and failed
// modifications still update that index.
func TestUnmodifiableObject(t *testing.T) {
	src := NewObject().Add("a", ValueOfInt(1))
	view := UnmodifiableObject(src)
	src.Add("a", ValueOfInt(2)).Add("b", ValueOfInt(3))
	if got, want := view.String(), `{"a":1,"a":2,"b":3}`; got != want {
		t.Errorf("view = %s; want %s", got, want)
	}
	if got := view.Get("a").String(); got != "1" {
		t.Errorf("view.Get(a) = %s; want the stale 1", got)
	}
	if got := src.Get("a").String(); got != "2" {
		t.Errorf("src.Get(a) = %s; want 2", got)
	}
	if got := view.Get("b").String(); got != "3" {
		t.Errorf("view.Get(b) = %s; want 3", got)
	}
	panicMessage(t, func() { view.Add("c", ValueOfInt(1)) })
	panicMessage(t, func() { view.Set("a", ValueOfInt(1)) })
	panicMessage(t, func() { view.Remove("a") })
	view.Remove("nope")
	panicMessage(t, func() { view.Merge(NewObject().Add("x", True)) })
	if got, want := view.String(), `{"a":1,"a":2,"b":3}`; got != want {
		t.Errorf("view after failures = %s; want %s", got, want)
	}
	if got := view.Get("a").String(); got != "2" {
		t.Errorf("view.Get(a) after failures = %s; want 2", got)
	}
	if !view.Equal(src) || view.Size() != 3 || !view.Contains("b") {
		t.Error("view does not reflect its source")
	}
	cp := CopyObject(view)
	cp.Add("c", Null)
	if got, want := cp.String(), `{"a":1,"a":2,"b":3,"c":null}`; got != want {
		t.Errorf("copy of view = %s; want %s", got, want)
	}
}

func TestObjectNilValuePanics(t *testing.T) {
	for name, f := range map[string]func(){
		"Add": func() { NewObject().Add("a", nil) },
		"Set": func() { NewObject().Set("a", nil) },
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

func TestJavaStringHash(t *testing.T) {
	tests := map[string]int32{
		"":   0,
		"a":  97,
		"Aa": 2112,
		"BB": 2112,
		// "hello" and a supplementary character, hashed over UTF-16 units.
		"hello":            99162322,
		"\U0001f600":       1772899,
		"abcdefghijklmnop": -2093879032,
	}
	for s, want := range tests {
		if got := javaStringHash(s); got != want {
			t.Errorf("javaStringHash(%q) = %d; want %d", s, got, want)
		}
	}
}
