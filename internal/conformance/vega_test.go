// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestLoadVegaCases(t *testing.T) {
	fsys := fstest.MapFS{
		// Vega headers are not strict YAML: the value below has an unquoted ": ".
		"crash/finalizing.puml": {Data: []byte("---\noutput: debug, svg\nexpected-exception: IllegalStateException - boom: here\n---\n@startcrash\ncrash\n@endcrash\n")},
		"sequence/ascii.puml":   {Data: []byte("---\noutput: svg, debug, atxt, utxt\n---\n@startuml\nA -> B\n@enduml\n")},
		"preproc/p.puml":        {Data: []byte("---\noutput: preproc\n---\n\\startuml\n!$a = 1\n\\enduml\n")},
		"xmi/x.puml":            {Data: []byte("---\noutput: xmi_star\n---\n@startuml\nclass A\n@enduml\n")},
		"mvp/plain.puml":        {Data: []byte("@startmindmap\n* root\n@endmindmap\n")},
		"mvp/notes.txt":         {Data: []byte("ignored")},
	}
	cases, err := LoadVegaCases(fsys, ".")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct {
		family  string
		formats []string
	}{
		"vega/crash/finalizing": {"crash", []string{"svg"}},
		"vega/sequence/ascii":   {"uml", []string{"svg", "txt", "utxt"}},
		"vega/preproc/p":        {"uml", []string{"preproc"}},
		"vega/xmi/x":            {"uml", []string{"svg"}},
		"vega/mvp/plain":        {"mindmap", []string{"svg"}},
	}
	if len(cases) != len(want) {
		t.Fatalf("loaded %d cases, want %d", len(cases), len(want))
	}
	for _, c := range cases {
		w, ok := want[c.ID]
		if !ok {
			t.Errorf("unexpected case %s", c.ID)
			continue
		}
		if c.Family != w.family || !slices.Equal(c.Formats, w.formats) {
			t.Errorf("%s: family %q formats %v, want %q %v", c.ID, c.Family, c.Formats, w.family, w.formats)
		}
	}
	if cases[0].ID != "vega/crash/finalizing" || cases[0].Source != "@startcrash\ncrash\n@endcrash\n" {
		t.Errorf("header not stripped from source: %q", cases[0].Source)
	}
}
