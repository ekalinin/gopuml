// SPDX-License-Identifier: MIT
// Provenance: original

package stdlib

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestNotice checks that NOTICE names every bundled and every excluded
// library.
func TestNotice(t *testing.T) {
	notice, err := os.ReadFile(filepath.Join("..", "..", "NOTICE"))
	if err != nil {
		t.Fatal(err)
	}
	named := func(name string) bool {
		return regexp.MustCompile(`(^|[^a-z0-9.-])` + regexp.QuoteMeta(name) + `([^a-z0-9-]|\.?$|\.[^0-9])`).Match(notice)
	}
	for _, name := range FolderNames() {
		if !named(name) {
			t.Errorf("NOTICE does not name the bundled library %s", name)
		}
	}
	for _, e := range Excluded {
		if !named(e.Name) {
			t.Errorf("NOTICE does not name the excluded library %s", e.Name)
		}
	}
}
