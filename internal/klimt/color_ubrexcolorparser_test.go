// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// TestUBrexColorParser checks the color of upstream's legacy partition
// command, its only user.
func TestUBrexColorParser(t *testing.T) {
	p := UBrexSimpleColor(color.TypeBack)
	cmd := UBrexConcat(UBrexLeaf("partition"), UBrexSpaceOneOrMore(), UBrexNamed("NAME", UBrexLeaf("〇+〴S")),
		UBrexSpaceZeroOrMore(), p.Regex(), UBrexOptional(UBrexNamed("STEREOTYPE", UBrexLeaf("<<  〄+〴. ->〘 >>〙"))),
		UBrexSpaceZeroOrMore(), UBrexOptional(UBrexLeaf("{")), UBrexEnd())
	for line, want := range map[string]string{
		"partition P #red {":         "#red",
		"partition P #a":             "#a",
		"partition P #red-blue<<s>>": "#red-blue",
		"partition P #red/blue {":    "#red/blue",
	} {
		m := cmd.Match(line)
		if !m.ExactMatch() {
			t.Errorf("%q must match", line)
			continue
		}
		r := regex.NewResultFromCaptures(m)
		if got := r.Str("COLOR", 0); got != want {
			t.Errorf("%q: COLOR = %q, want %q", line, got, want)
		}
		colors, err := p.Color(r, color.Instance())
		if err != nil || colors.Color(color.TypeBack) == nil {
			t.Errorf("%q: Color = %v, %v", line, colors, err)
		}
	}
	if cmd.Match("partition P {").ExactMatch() {
		t.Error("the UBrex color is mandatory")
	}
}
