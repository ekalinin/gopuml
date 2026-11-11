// SPDX-License-Identifier: MIT
// Provenance: original

package yamlparser

import (
	"errors"
	"testing"
)

// parseJSON reads the lines and returns the JSON text of their value.
func parseJSON(t *testing.T, lines ...string) (string, error) {
	t.Helper()
	m, err := Parse(lines)
	if err != nil {
		return "", err
	}
	v, err := MonomorphToJSON(m)
	if err != nil || v == nil {
		return "", err
	}
	return v.String(), nil
}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"Mapping with nested values", []string{"fruit: Apple", "size: Large", "color:", "  - Red", "  - Green", "info:",
			"  origin: Spain", "  weight: 120"},
			`{"fruit":"Apple","size":"Large","color":["Red","Green"],"info":{"origin":"Spain","weight":"120"}}`},
		{"Sequence of mappings", []string{"- name: Mark", "  hr: 65", "- name: Sammy", "  hr: 63"},
			`[{"name":"Mark","hr":"65"},{"name":"Sammy","hr":"63"}]`},
		{"Mappings after lone dashes", []string{"-", "  name: Mark", "-", "  name: Sammy"},
			`[{"name":"Mark"},{"name":"Sammy"}]`},
		{"Flow sequence, empty and quoted values", []string{"tags: [a, b, c]", "empty:", `quoted: "hello: world"`},
			`{"tags":["a","b","c"],"empty":"","quoted":"hello: world"}`},
		{"Block value", []string{"text: |", "  line one", "  line two", "next: value"},
			`{"text":"line one\nline two\n","next":"value"}`},
		{"Value continued on the next lines", []string{"desc:", "  first part", "  second part", "other: 1"},
			`{"desc":"first part second part","other":"1"}`},
		{"YAML comments", []string{"# a comment", "a: 1 # trailing", "b: 2"}, `{"a":"1","b":"2"}`},
		{"Highlighted YAML key", []string{"info:", "  origin: Spain", "  weight: 120"},
			`{"info":{"origin":"Spain","weight":"120"}}`},
	} {
		got, err := parseJSON(t, tc.lines...)
		if err != nil || got != tc.want {
			t.Errorf("%s: %s, error %v, want %s", tc.name, got, err, tc.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, tc := range []struct {
		lines []string
		class string
	}{
		{[]string{"just some text"}, YamlSyntaxException},
		{[]string{"tags: [a, b, c]", "- x"}, IllegalStateException},
		{[]string{"text: >", "  folded"}, UnsupportedOperationException},
		// A list item under a lone dash leaves an undetermined value in the
		// list, as upstream does.
		{[]string{"-", "  - a"}, UnsupportedOperationException},
	} {
		_, err := parseJSON(t, tc.lines...)
		if e, ok := errors.AsType[*Error](err); !ok || e.Class != tc.class {
			t.Errorf("%q: error %v, want a %s", tc.lines, err, tc.class)
		}
	}
}

func TestParseDashDash(t *testing.T) {
	// "- - a" is the item "- a", as upstream reads it.
	if got, err := parseJSON(t, "- - a"); err != nil || got != `["- a"]` {
		t.Errorf("- - a: %s, error %v", got, err)
	}
}

func TestParseEmpty(t *testing.T) {
	m, err := Parse(nil)
	if err != nil || m.Type() != Undeterminate {
		t.Fatalf("Parse(nil) = %v, %v", m, err)
	}
	if v, err := MonomorphToJSON(m); v != nil || err != nil {
		t.Errorf("MonomorphToJSON of nothing = %v, %v, want nil", v, err)
	}
}
