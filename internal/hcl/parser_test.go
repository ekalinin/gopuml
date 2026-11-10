// SPDX-License-Identifier: MIT
// Provenance: original

package hcl

import (
	"errors"
	"slices"
	"testing"
)

// sourceOf returns the source of the lines.
func sourceOf(lines ...string) *hclSource {
	s := &hclSource{}
	for _, l := range lines {
		s.add(l)
	}
	return s
}

// terms returns the tokens of the lines, as text.
func terms(t *testing.T, lines ...string) []string {
	t.Helper()
	p, err := newHclParser(sourceOf(lines...))
	if err != nil {
		t.Fatal(err)
	}
	var result []string
	for _, term := range p.terms {
		result = append(result, term.String())
	}
	return result
}

func TestHclSource(t *testing.T) {
	if got := string(sourceOf("a {", "  # a comment line", "  b = 1").all); got != "a {  b = 1" {
		t.Errorf("source %q, want the lines joined without the comment line", got)
	}
}

func TestHclTokens(t *testing.T) {
	want := []string{"STRING_SIMPLE(b)", "EQUALS", "FUNCTION_NAME(lower)", "PARENTHESIS_OPEN", "STRING_QUOTED(ABC)",
		"COMMA", "STRING_QUOTED(d)", "PARENTHESIS_CLOSE"}
	if got := terms(t, `b = lower("ABC", "d")`); !slices.Equal(got, want) {
		t.Errorf("tokens %q, want %q", got, want)
	}
	// The lines are joined without line breaks: "1" and "e" are one value.
	// The characters after the last symbol are dropped, as upstream does.
	if got := terms(t, "d = 1", "e = 2"); !slices.Equal(got, []string{"STRING_SIMPLE(d)", "EQUALS", "STRING_SIMPLE(1e)",
		"EQUALS"}) {
		t.Errorf("tokens %q", got)
	}
	if _, err := newHclParser(sourceOf("a = (")); err == nil {
		t.Error("a parenthesis without a function name gives no error")
	}
}

// parseJSON parses the lines and returns the JSON text of their data.
func parseJSON(lines ...string) (string, error) {
	p, err := newHclParser(sourceOf(lines...))
	if err != nil {
		return "", err
	}
	data, err := p.parseMe()
	if err != nil {
		return "", err
	}
	return data.String(), nil
}

func TestHclParse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"Single block", []string{`resource "aws_instance" "web" {`, `  ami           = "ami-a1b2c3d4"`,
			`  instance_type = "t2.micro"`, `  count         = 2`, `  enabled       = true`, `}`},
			`{"ami":"ami-a1b2c3d4","instance_type":"t2.micro","count":"2","enabled":"true"}`},
		{"Two blocks", []string{`variable "region" {`, `  default = "us-east-1"`, `}`, `provider aws {`,
			`  region = "x"`, `}`},
			`{"variable \"region\"":{"default":"us-east-1"},"provider aws":{"region":"x"}}`},
		{"Lists, objects and function calls", []string{`module "m" {`, `  tags = ["a", "b"]`, `  settings = {`,
			`    level = "high"`, `  }`, `  name = lower("ABC", "d")`, `  empty = list()`, `}`},
			`{"tags":["a","b"],"settings":{"level":"high"},"name":{"lower()":["ABC","d"]},"empty":"list()"}`},
		{"Attribute with a colon", []string{`a {`, `  b : "c"`, `  d = 1.5`, `}`}, `{"b":"c","d":"1.5"}`},
		{"Comment line", []string{`a {`, `  # a comment line`, `  b = "c"`, `}`}, `{"b":"c"}`},
		{"Unquoted values in a list", []string{`a {`, `  l = [1, "x"]`, `}`}, `{"l":["x"]}`},
		{"Empty source", nil, `{}`},
	} {
		got, err := parseJSON(tc.lines...)
		if err != nil || got != tc.want {
			t.Errorf("%s: %s, error %v, want %s", tc.name, got, err, tc.want)
		}
	}
}

func TestHclParseErrors(t *testing.T) {
	for _, tc := range []struct {
		lines []string
		class string
	}{
		{[]string{`a = "b"`}, IllegalStateException},
		{[]string{`module "m" {`, `  nested {`, `    x = "y"`, `  }`, `}`}, IllegalStateException},
		{[]string{`a {`, `  b = "c" # trailing`, `}`}, IllegalStateException},
		{[]string{`a {`, `b = "c"`, `d = 1`, `e = 2`, `}`}, IllegalStateException},
		{[]string{`a {`}, NoSuchElementException},
	} {
		_, err := parseJSON(tc.lines...)
		if e, ok := errors.AsType[*Error](err); !ok || e.Class != tc.class {
			t.Errorf("%q: error %v, want a %s", tc.lines, err, tc.class)
		}
	}
}
