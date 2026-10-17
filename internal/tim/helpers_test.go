// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/text"
)

func TestJavaParseInt(t *testing.T) {
	tests := []struct {
		in   string
		want int32
		ok   bool
	}{
		{"0", 0, true}, {"42", 42, true}, {"-42", -42, true}, {"+7", 7, true},
		{"2147483647", 2147483647, true}, {"-2147483648", -2147483648, true},
		{"2147483648", 0, false}, {"-2147483649", 0, false}, {"99999999999", 0, false},
		{"", 0, false}, {"-", 0, false}, {"+", 0, false}, {"1a", 0, false}, {" 1", 0, false},
		{"١٢٣", 123, true}, {"１２", 12, true},
	}
	for _, tt := range tests {
		got, ok := javaParseInt(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("javaParseInt(%q) = %d, %v; want %d, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestJavaStrings(t *testing.T) {
	if got := javaSplitNewline("a\nb\n\n"); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("split %q", got)
	}
	if got := javaSplitNewline("\n"); len(got) != 0 {
		t.Errorf("split newline %q", got)
	}
	if got := javaSplitNewline(""); !slices.Equal(got, []string{""}) {
		t.Errorf("split empty %q", got)
	}
	if got := javaSplitNewline("\na"); !slices.Equal(got, []string{"", "a"}) {
		t.Errorf("split leading %q", got)
	}
	// UTF-16 order: a supplementary character (surrogates) sorts before U+FFFF.
	if javaCompare("😀", "￿") >= 0 || javaCompare("b", "a") <= 0 || javaCompare("a", "ab") >= 0 {
		t.Error("javaCompare")
	}
	if got := javaHashCode("abc"); got != 96354 {
		t.Errorf("hash %d", got)
	}
	if got := javaLength("a😀"); got != 3 {
		t.Errorf("length %d", got)
	}
	if got := javaToUpperCase("straße ﬁ é"); got != "STRASSE FI É" {
		t.Errorf("upper %q", got)
	}
	if got := fromUTF16(toUTF16("aé😀")); got != "aé😀" {
		t.Errorf("round trip %q", got)
	}
}

func TestJavaHashOrder(t *testing.T) {
	order := func(keys ...int32) []int32 {
		o := newJavaHashOrder()
		for _, k := range keys {
			o.add(k)
		}
		var out []int32
		for _, i := range o.order() {
			out = append(out, keys[i])
		}
		return out
	}
	// Small keys are visited in increasing order.
	if got := order(5, 3, 9); !slices.Equal(got, []int32{3, 5, 9}) {
		t.Errorf("got %v", got)
	}
	// 17 and 1 share bucket 1 of 16: insertion order.
	if got := order(17, 20, 1); !slices.Equal(got, []int32{17, 1, 20}) {
		t.Errorf("got %v", got)
	}
	// The 13th key doubles the table to 32 buckets: 17 moves after 1..16.
	keys := []int32{17, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	if got := order(keys...); !slices.Equal(got, []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 17}) {
		t.Errorf("got %v", got)
	}
	// The high half of the hash is folded into the low half.
	if javaSpread(0x10001) != 0x10000 {
		t.Errorf("spread %x", javaSpread(0x10001))
	}
}

func TestTrie(t *testing.T) {
	trie := newTrieImpl()
	trie.add("$a")
	trie.add("$abc")
	match := func(s string) int { return trie.getLonguestMatchStartingIn(toUTF16(s), 0) }
	tests := map[string]int{"$a": 2, "$a b": 2, "$abc": 4, "$abcd": 4, "$ab": 0, "$abx": 0, "$": 0, "x": 0}
	for s, want := range tests {
		if got := match(s); got != want {
			t.Errorf("%q: %d, want %d", s, got, want)
		}
	}
	if !trie.remove("$a") || trie.remove("$a") || match("$a") != 0 || match("$abc") != 4 {
		t.Error("remove")
	}
}

func TestTValue(t *testing.T) {
	i := TValueFromInt
	s := TValueFromString
	tests := []struct {
		got  *TValue
		want string
	}{
		{i(2147483647).Add(i(1)), "-2147483648"},
		{i(7).DividedBy(i(-2)), "-3"},
		{i(-2147483648).DividedBy(i(-1)), "-2147483648"},
		{s("a").Add(i(1)), "a1"},
		{s("a").Minus(s("b")), "ab"},
		{s("a").Multiply(i(2)), "a*2"},
		{i(1).DividedBy(s("b")), "1/b"},
		{s("10").LessThan(s("9")), "1"},
		{i(10).LessThan(s("9")), "1"},
		{s("x").EqualsOperation(s("x")), "1"},
		{TValueFromJSON(nil), "0"},
	}
	for n, tt := range tests {
		if tt.got.String() != tt.want {
			t.Errorf("%d: %q, want %q", n, tt.got.String(), tt.want)
		}
	}
	if s("").ToBoolean() || !s("0").ToBoolean() || i(0).ToBoolean() || !i(-1).ToBoolean() {
		t.Error("ToBoolean")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("division by zero did not throw")
		}
	}()
	i(1).DividedBy(i(0))
}

func TestEvalBoolean(t *testing.T) {
	truth := func(name string) bool { return name == "A" || name == "$b" }
	tests := map[string]bool{"A": true, "C": false, "$b": true, "!A": false, "A && C": false, "A & $b": true,
		"C || A": true, "C | C": false, "(A || C) && !C": true, " A ": true}
	for expr, want := range tests {
		if got := newEvalBoolean(expr, truth).eval(); got != want {
			t.Errorf("%q: %v, want %v", expr, got, want)
		}
	}
	for _, bad := range []string{"(", "A B", "", "&"} {
		func() {
			defer func() { _ = recover() }()
			newEvalBoolean(bad, truth).eval()
			t.Errorf("%q did not throw", bad)
		}()
	}
}

func TestMutateExpands1(t *testing.T) {
	loc := &text.LineLocation{Desc: "x", Position: 3}
	in := []text.StringLocated{
		{S: "ab", Location: loc},
		{S: "{{xy}}z", Location: loc},
		{S: "plain", Location: loc},
		{S: "", Location: loc, PreprocessorError: "e"},
	}
	var got []string
	for _, l := range mutateExpands1(in) {
		got = append(got, l.S)
	}
	if want := []string{"a", "b", "{{xy}}", "z", "plain", "", ""}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLineTypes(t *testing.T) {
	tests := map[string]TLineType{
		"A -> B":            linePlain,
		"' x":               lineCommentSimple,
		"/' x":              lineCommentLongStart,
		"!$x = 1":           lineAffectation,
		"!if 1":             lineIf,
		"!IF 1":             lineIf,
		"!ifdef X":          lineIfdef,
		"!define X(a) a":    lineLegacyDefine,
		"!define X":         lineAffectationDefine,
		"!define\vX(a)":     linePlain,
		"!define\vX(a) a b": lineLegacyDefine,
		"!ſtartsub X":       linePlain,
		"!include_once x":   lineInclude,
		"!endfunction":      lineEndFunction,
		"!end procedure":    lineEndFunction,
		"!$x😀 = 1":          lineAffectation,
	}
	for s, want := range tests {
		if got := getTLineType(s); got != want {
			t.Errorf("%q: %s, want %s", s, got, want)
		}
	}
}

// TestStandardFunctions checks the builtin table against upstream's 74
// builtins.
func TestStandardFunctions(t *testing.T) {
	if len(standardFunctions) != 74 {
		t.Errorf("%d builtins, want 74", len(standardFunctions))
	}
	seen := map[signatureKey]bool{}
	for _, f := range standardFunctions {
		if seen[f.signature.key()] {
			t.Errorf("duplicate %v", f.signature)
		}
		seen[f.signature.key()] = true
		if f.build == nil {
			continue
		}
		if got := f.build(NewDefines()).Signature(); got.key() != f.signature.key() {
			t.Errorf("builtin %v has the signature %v", f.signature, got)
		}
	}
}
