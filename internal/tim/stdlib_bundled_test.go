// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/security"
	bundled "github.com/ekalinin/gopuml/internal/stdlib"
)

// TestGetAllStdlibBundled is the scenario "Excluded library not listed":
// %get_all_stdlib() lists the bundled libraries, without the excluded ones.
func TestGetAllStdlibBundled(t *testing.T) {
	got, err := preprocessBuiltins(t, nil, Options{}, nil, "A : %get_all_stdlib()")
	if err != nil || len(got) != 1 {
		t.Fatalf("got %q, %v", got, err)
	}
	var names []string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(got[0], "A : ")), &names); err != nil {
		t.Fatal(err)
	}
	if !slices.IsSorted(names) || !slices.Contains(names, "c4") || !slices.Contains(names, "awslib14") {
		t.Errorf("names %q", names)
	}
	for _, e := range bundled.Excluded {
		if slices.Contains(names, e.Name) {
			t.Errorf("excluded library %s is listed", e.Name)
		}
	}
	if want := bundled.FolderNames(); !slices.Equal(names, want) {
		t.Errorf("names %q, want %q", names, want)
	}
}

// TestStdlibBundledMetadata checks the scenarios "Version of a library",
// "Version of an alias" and "JSON resource of a library".
func TestStdlibBundledMetadata(t *testing.T) {
	got, err := preprocessBuiltins(t, nil, Options{}, nil,
		`A : %get_stdlib("c4", "version")`,
		`A : %get_stdlib("awslib", "version")`,
		`!$o = %load_json("<office/all.json>")`,
		`A : %size($o)`)
	checkLines(t, got, err, "A : 2.14.0", "A : 14.0.0", "A : 10")

	// %get_stdlib does not change the case of the name, as upstream.
	for _, line := range []string{`A : %get_stdlib("C4", "version")`, `A : %get_stdlib("adaml")`, `A : %get_stdlib("foo")`} {
		if _, err := preprocessBuiltins(t, nil, Options{}, nil, line); err == nil || err.Error() != "Fatal parsing error" {
			t.Errorf("%s: error %v, want Fatal parsing error", line, err)
		}
	}
}

// TestStdlibBundledIncludes checks the scenarios "Library include",
// "Library include in the sandbox", "Library name in another case",
// "Missing file in a library" and "Missing library".
func TestStdlibBundledIncludes(t *testing.T) {
	for _, profile := range []security.Profile{security.Legacy, security.Sandbox} {
		for _, lib := range []string{"C4", "c4"} {
			src := "@startuml\n!include <" + lib + "/C4_Container>\nContainer(api, \"API\", \"Go\")\n@enduml\n"
			res, err := preprocessBlock(t, src, memAccess(profile, nil), "")
			if err != nil {
				t.Fatalf("%v %s: %v", profile, lib, err)
			}
			// C4_Container includes C4_Context, which includes C4.
			lines := res.PreprocLines()
			if !slices.ContainsFunc(lines, func(l string) bool { return strings.Contains(l, "<<container>>") }) {
				t.Errorf("%v %s: no container element in %d lines", profile, lib, len(lines))
			}
		}
	}
	for _, tt := range []struct{ include, message string }{
		{"<C4/Nope>", "cannot include <C4/Nope>"},
		{"<foo/bar>", "Fatal parsing error"},
	} {
		src := "@startuml\nAlice -> Bob : hi\n!include " + tt.include + "\n@enduml\n"
		_, err := preprocessBlock(t, src, memAccess(security.Legacy, nil), "")
		wantError(t, err, tt.message, 2)
	}
}
