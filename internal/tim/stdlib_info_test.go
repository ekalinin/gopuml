// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"slices"
	"testing"
)

func TestStdlibInfoVersion(t *testing.T) {
	if got := PrintStdLib(); len(got) == 0 || slices.Contains(got, "adaml") {
		t.Errorf("embedded index: %q", got)
	}
	s := NewStdlib(map[string]map[string][]byte{
		"awslib": {"info.spm": []byte("VERSION=1.2\nSOURCE=https://example.org\n")},
	})
	if got, want := s.infoVersion(true), []string{"<b>awslib", "Version 1.2", "Delivered by https://example.org", " "}; !slices.Equal(got, want) {
		t.Errorf("details: %q, want %q", got, want)
	}
	if got, want := s.infoVersion(false), []string{"* awslib (Version 1.2)"}; !slices.Equal(got, want) {
		t.Errorf("summary: %q, want %q", got, want)
	}
}
