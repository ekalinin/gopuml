// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag

import (
	"maps"
	"slices"
	"testing"
)

func TestToSet(t *testing.T) {
	for _, tc := range []struct {
		definition string
		ok         bool
		want       map[string]string
	}{
		{"len = 4", true, map[string]string{"len": "4"}},
		{"len=4,height=2", true, map[string]string{"len": "4", "height": "2"}},
		{`description = "foo *bar*, baz"`, true, map[string]string{"description": "foo *bar*, baz"}},
		{`color="palegreen", textcolor="blue", linecolor="red"`, true,
			map[string]string{"color": "palegreen", "textcolor": "blue", "linecolor": "red"}},
		{"label = label_1 rotate = 270", true, map[string]string{"label": "label_1", "rotate": "270"}},
		{"len = 1, len = 2", true, map[string]string{"len": "2"}},
		{"", true, map[string]string{}},
		{"len", true, map[string]string{}},
		{"", false, map[string]string{}},
	} {
		if got := toSet(tc.definition, tc.ok); !maps.Equal(got, tc.want) {
			t.Errorf("toSet(%q, %v) = %v, want %v", tc.definition, tc.ok, got, tc.want)
		}
	}
}

func TestVerticalLine(t *testing.T) {
	r := newRec()
	newVerticalLine(32, 16, nil).DrawU(r)
	if want := []string{"0.0,0.0 path SEG_MOVETO [0.0, 16.0] SEG_LINETO [0.0, 32.0]"}; !slices.Equal(*r.log, want) {
		t.Errorf("drawn %q, want %q", *r.log, want)
	}
}
