// SPDX-License-Identifier: MIT
// Provenance: original

package yamlparser

import (
	"slices"
	"testing"
)

func TestBuildYamlLine(t *testing.T) {
	for _, tc := range []struct {
		line       string
		typ        YamlLineType
		indent     int
		key, value string
		listItem   bool
	}{
		{"fruit: Apple", KeyAndValue, 0, "fruit", "Apple", false},
		{"  - Red", PlainElementList, 4, "", "Red", true},
		{"-", PlainDash, 1, "", "", true},
		{"text: |", KeyAndBlockStyle, 0, "text", "|", false},
		{"color:", KeyOnly, 0, "color", "", false},
		{"just some text", NoKeyOnlyText, 0, "", "just some text", false},
		{"a: 1 # trailing", KeyAndValue, 0, "a", "1", false},
		{"# a comment", EmptyLine, 0, "", "", false},
		{`quoted: "hello: world"`, KeyAndValue, 0, "quoted", "hello: world", false},
		{"\tb: 2", KeyAndValue, 4, "b", "2", false},
		{"  - name: Mark", KeyAndValue, 4, "name", "Mark", true},
	} {
		l := BuildYamlLine(tc.line)
		if l.Type() != tc.typ || l.Indent() != tc.indent || l.Key() != tc.key || l.value != tc.value ||
			l.IsListItem() != tc.listItem {
			t.Errorf("%q: %v, want %v indent %d key %q value %q list item %v", tc.line, l, tc.typ, tc.indent,
				tc.key, tc.value, tc.listItem)
		}
	}
}

func TestBuildYamlLineFlowSequence(t *testing.T) {
	l := BuildYamlLine("tags: [a, b, c]")
	if l.Type() != KeyAndFlowSequence || l.Key() != "tags" || !slices.Equal(l.Values(), []string{"a", "b", "c"}) {
		t.Errorf("%v with the values %q", l, l.values)
	}
	// A quoted field keeps what follows its closing quote, untrimmed, as
	// upstream does.
	l = BuildYamlLine(`t: [ "a, b" , c\,d, ]`)
	if !slices.Equal(l.Values(), []string{"a, b ", "c,d", ""}) {
		t.Errorf("quoted and escaped fields: %q", l.Values())
	}
}

func TestYamlLineValueThrows(t *testing.T) {
	defer func() {
		if e, ok := recover().(*Error); !ok || e.Class != IllegalStateException || e.Message != "KEY_ONLY" {
			t.Errorf("recovered %v, want an IllegalStateException for KEY_ONLY", e)
		}
	}()
	BuildYamlLine("color:").Value()
}

func TestPeeker(t *testing.T) {
	p := newPeeker([]string{"a", "b"})
	if v, ok := p.peek(1); !ok || v != "b" {
		t.Errorf("peek(1) = %q, %v", v, ok)
	}
	p.jump()
	if _, ok := p.peek(1); ok {
		t.Error("peek past the end is ok")
	}
}
