// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"slices"
	"testing"
)

// TestUBrexBuilderMatchesString checks that an expression built from parts
// matches and captures as the same expression written as a string.
func TestUBrexBuilderMatchesString(t *testing.T) {
	built := UBrexConcat(
		UBrexLeaf("if"),
		UBrexSpaceOneOrMore(),
		UBrexOr(
			UBrexNamed("CODE", UBrexLeaf("〇+「〴an_.」")),
			UBrexLeaf("〴g 〶$QUOTED=〇*〴G 〴g")),
		UBrexOptional(UBrexConcat(UBrexSpaceOneOrMore(), UBrexLeaf("then"))),
		UBrexEnd())
	written := ubrexBuild("if 〇+〴s 【 〶$CODE=〇+「〴an_.」 ┇ 〴g 〶$QUOTED=〇*〴G 〴g 】 〇?〘 〇+〴s then 〙 〒$")
	for _, line := range []string{"if ok", `if "is valid?" then`, "if ok then", "if", "iff ok", `if "x" else`} {
		got, want := built.Match(line), written.matchString(line)
		if got.ExactMatch() != want.exactMatch() {
			t.Errorf("%q: exact %v, want %v", line, got.ExactMatch(), want.exactMatch())
		}
		for _, key := range []string{"CODE", "QUOTED"} {
			if g, w := got.FindValuesByKey(key), want.findValuesByKey(key); !slices.Equal(g, w) {
				t.Errorf("%q: %s = %q, want %q", line, key, g, w)
			}
		}
	}
	if m := built.Match("if ok"); !m.ExactMatch() || !slices.Equal(m.FindValuesByKey("CODE"), []string{"ok"}) {
		t.Errorf("if ok: exact %v, CODE %q", m.ExactMatch(), m.FindValuesByKey("CODE"))
	}
	if m := built.Match("iff ok"); m.ExactMatch() {
		t.Error("iff ok matches")
	}
	if m := built.Match(`if "is valid?" then`); !slices.Equal(m.FindFirstValuesByKeyPrefix("QUO"), []string{"is valid?"}) {
		t.Errorf("prefix lookup %q", m.FindFirstValuesByKeyPrefix("QUO"))
	}
}
