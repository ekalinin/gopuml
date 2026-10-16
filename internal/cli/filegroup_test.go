// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// tree creates files relative to a fresh working directory.
func tree(t *testing.T, files ...string) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, []byte("@startuml\n@enduml\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func inputs(t *testing.T, args ...string) []string {
	t.Helper()
	return parse(t, args...).InputFiles()
}

func TestInputDirectory(t *testing.T) {
	// Spec scenario "Directory argument".
	tree(t, "docs/a.puml", "docs/b.txt", "docs/c.plantuml", "docs/sub/d.puml", "docs/E.PU", "docs/f.wsd")
	got := inputs(t, "docs/")
	if want := []string{"docs/E.PU", "docs/a.puml", "docs/b.txt"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestInputRecursivePatternWithExclusion(t *testing.T) {
	// Spec scenario "Recursive pattern with exclusion".
	tree(t, "top.puml", "a/b.puml", "a/draft/c.puml", "draft/d.puml", "drafty/e.puml", "a/f.txt")
	got := inputs(t, "**/*.puml", "-x", "**/draft/**")
	if want := []string{"./a/b.puml", "./drafty/e.puml", "./top.puml"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestInputSimplePatterns(t *testing.T) {
	tree(t, "x.puml", "y.PUML", "z.txt", "dir/a.puml", "dir/b.pu", "dir/sub/c.puml")
	// The current directory is java.io.File("."): paths start with "./".
	if got, want := inputs(t, "*.puml"), []string{"./x.puml", "./y.PUML"}; !slices.Equal(got, want) {
		t.Errorf("*.puml: %v, want %v", got, want)
	}
	if got, want := inputs(t, "dir/*.p?"), []string{"dir/b.pu"}; !slices.Equal(got, want) {
		t.Errorf("dir/*.p?: %v, want %v", got, want)
	}
	if got, want := inputs(t, "dir/**.puml"), []string{"dir/a.puml", "dir/sub/c.puml"}; !slices.Equal(got, want) {
		t.Errorf("dir/**.puml: %v, want %v", got, want)
	}
	if got := inputs(t, "missing.puml"); len(got) != 0 {
		t.Errorf("missing file: %v", got)
	}
	if got, want := inputs(t, "dir/a.puml", "x.puml"), []string{"dir/a.puml", "x.puml"}; !slices.Equal(got, want) {
		t.Errorf("files keep argument order: %v", got)
	}
	if got, want := inputs(t, "dir//a.puml"), []string{"dir/a.puml"}; !slices.Equal(got, want) {
		t.Errorf("path normalization: %v", got)
	}
}

func TestGlobToRegexp(t *testing.T) {
	for glob, want := range map[string]string{
		"*.puml":      `(?i)^(\./)?[^/]{0,}\.puml$`,
		"**/draft/**": `(?i)^(\./)?.{0,}/draft/.{0,}$`, // no "/**/" inside
		"a/**/b":      `(?i)^(\./)?a(/|/.{0,}/)b$`,
		`a\b?.txt`:    `(?i)^(\./)?a/b[^/]\.txt$`,
	} {
		if got := globToRegexp(glob); got != want {
			t.Errorf("globToRegexp(%q) = %q, want %q", glob, got, want)
		}
	}
}
