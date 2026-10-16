// SPDX-License-Identifier: MIT
// Provenance: original

package regex

import (
	"strconv"
	"strings"
	"testing"
)

// The expected orders come from java.util.HashMap on JDK 25, with the
// insertions upstream's createPartialMatch makes.
func TestJavaHashMapOrder(t *testing.T) {
	seq := map[int]string{
		1:  "KEY0",
		2:  "KEY1,KEY0",
		3:  "KEY2,KEY1,KEY0",
		5:  "KEY2,KEY1,KEY0,KEY4,KEY3",
		8:  "KEY2,KEY1,KEY0,KEY6,KEY5,KEY4,KEY3,KEY7",
		12: "KEY2,KEY1,KEY0,KEY6,KEY5,KEY4,KEY3,KEY9,KEY8,KEY7,KEY11,KEY10",
		13: "KEY12,KEY2,KEY1,KEY0,KEY6,KEY5,KEY4,KEY3,KEY9,KEY8,KEY7,KEY11,KEY10",
		20: "KEY13,KEY12,KEY15,KEY14,KEY17,KEY16,KEY19,KEY18,KEY2,KEY1,KEY0,KEY6,KEY5,KEY4,KEY3,KEY9,KEY8,KEY7," +
			"KEY11,KEY10",
		24: "KEY13,KEY12,KEY23,KEY15,KEY14,KEY17,KEY16,KEY19,KEY18,KEY2,KEY1,KEY0,KEY6,KEY5,KEY4,KEY3,KEY9,KEY8," +
			"KEY7,KEY20,KEY11,KEY22,KEY10,KEY21",
		25: "KEY24,KEY23,KEY11,KEY10,KEY13,KEY12,KEY15,KEY14,KEY17,KEY16,KEY19,KEY18,KEY2,KEY1,KEY0,KEY6,KEY5," +
			"KEY4,KEY3,KEY9,KEY8,KEY7,KEY20,KEY22,KEY21",
		40: "KEY24,KEY23,KEY26,KEY25,KEY28,KEY27,KEY29,KEY31,KEY30,KEY11,KEY33,KEY10,KEY32,KEY13,KEY35,KEY12," +
			"KEY34,KEY15,KEY37,KEY14,KEY36,KEY17,KEY39,KEY16,KEY38,KEY19,KEY18,KEY2,KEY1,KEY0,KEY6,KEY5,KEY4," +
			"KEY3,KEY9,KEY8,KEY7,KEY20,KEY22,KEY21",
	}
	for n, want := range seq {
		m := &javaHashMap{}
		for i := range n {
			m.putAll([]string{"KEY" + strconv.Itoa(i)})
		}
		if got := strings.Join(m.keys(), ","); got != want {
			t.Errorf("seq%d:\n got  %s\n want %s", n, got, want)
		}
	}
	leaf := func(name string) Part { return NewNamedLeaf(0, name, "") }
	or := Or("NAME", leaf("CODE1"), leaf("DISPLAY1"), leaf("CODE2"), leaf("DISPLAY2"), leaf("CODE3"),
		leaf("CODE4"))
	if got := strings.Join(javaNames(or), ","); got != "CODE2,CODE1,CODE4,DISPLAY2,CODE3,DISPLAY1,NAME" {
		t.Errorf("or: %s", got)
	}
	top := NewConcat(Start(), leaf("TYPE"), or, leaf("GENERIC"), leaf("TAGS1"), leaf("STEREO"), leaf("TAGS2"),
		leaf("URL"), leaf("COLOR"), leaf("LINECOLOR"), leaf("EXTENDS"), leaf("IMPLEMENTS"))
	want := "EXTENDS,TAGS1,TAGS2,IMPLEMENTS,DISPLAY1,URL,NAME,LINECOLOR,CODE2,CODE1,CODE4,DISPLAY2,CODE3,GENERIC," +
		"COLOR,STEREO,TYPE"
	if got := strings.Join(javaNames(top), ","); got != want {
		t.Errorf("top:\n got  %s\n want %s", got, want)
	}
}

func TestLazyFollowsJavaOrder(t *testing.T) {
	// TAGS2 comes before TAGS1 in a HashMap of these two names and NAME.
	c := NewConcat(Start(), NewNamedLeaf(1, "TAGS1", "(a)?"), NewNamedLeaf(1, "NAME", "(b)"),
		NewNamedLeaf(1, "TAGS2", "(c)?"), End())
	r := c.Matcher("abc")
	got, _ := r.Lazy("TAGS", 0)
	m := &javaHashMap{}
	for _, k := range []string{"TAGS1", "NAME", "TAGS2"} {
		m.putAll([]string{k})
	}
	first := "a"
	for _, k := range m.keys() {
		if k == "TAGS2" {
			first = "c"
			break
		}
		if k == "TAGS1" {
			break
		}
	}
	if got != first {
		t.Errorf("Lazy = %q, want %q (order %v)", got, first, m.keys())
	}
}
