// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram

import "testing"

// TestCleanID checks upstream's DescriptionDiagram.cleanId: its own "()",
// ":name:/" and "(name)/" forms, then CucaDiagram's cleanId, which also
// removes the quotes, parentheses, brackets and colons of the short forms.
func TestCleanID(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"() REST", "REST"},
		{"()REST", "REST"},
		{`() "REST api"`, "REST api"},
		{":Admin:/", "Admin"},
		{"(Login)/", "Login"},
		{`"Web server"`, "Web server"},
		{"API", "API"},
		{":Admin:", "Admin"},
		{"(Login)", "Login"},
		{"[API]", "API"},
	} {
		if got := cleanID(c.in); got != c.want {
			t.Errorf("cleanID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
