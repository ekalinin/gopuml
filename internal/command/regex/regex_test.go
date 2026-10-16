// SPDX-License-Identifier: MIT
// Provenance: original

package regex

import "testing"

func TestTransform(t *testing.T) {
	for in, want := range map[string]string{
		`^[%s]*x$`: "^[\\t\\n\\x0B\\f\\r \\x{00A0}]*x$",
		`a\sb`:     "a[\\t\\n\\x0B\\f\\r ]b",
		`a\Sb`:     "a[^\\t\\n\\x0B\\f\\r ]b",
		`[%g]`:     "[\"\\x{201C}\\x{201D}]",
		`[%q]`:     `['\x{2018}\x{2019}]`,
		`[%pLN_]+`: `[\p{L}\p{N}_]+`,
		`\[x\]\d`:  `\[x\]\d`,
		`100%`:     `100%`,
	} {
		got, err := Transform(in)
		if err != nil || got != want {
			t.Errorf("Transform(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := Transform(`[\S]`); err == nil {
		t.Error(`\S inside a class must be rejected`)
	}
}

func TestCompileSemantics(t *testing.T) {
	re := MustCompile(`^start[%s]+(\w+)$`)
	for in, want := range map[string]bool{
		"start x":  true,
		"START x":  true, // case-insensitive
		"start\vx": true, // Java's \s includes vertical tab
		"start x":  true, // %s includes no-break space
		"startx":   false,
	} {
		if got := re.MatchString(in); got != want {
			t.Errorf("match %q = %v, want %v", in, got, want)
		}
	}
	if first, second := Cached(`a+`), Cached(`a+`); first != second {
		t.Error("Cached must return the same compiled pattern")
	}
}
