// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"bytes"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// compareSprites returns the differences between the bundled sprites and
// upstream's sprites/ folder: the files that differ, are not bundled or are
// not upstream's.
func compareSprites(bundled, upstream fs.FS) []string {
	want, err := spriteFiles(upstream)
	if err != nil {
		return []string{"upstream: " + err.Error()}
	}
	got, err := spriteFiles(bundled)
	if err != nil {
		return []string{"bundled: " + err.Error()}
	}
	var problems []string
	for _, name := range slices.Sorted(maps.Keys(want)) {
		data, ok := got[name]
		switch {
		case !ok:
			problems = append(problems, name+": not bundled")
		case !bytes.Equal(data, want[name]):
			problems = append(problems, name+": differs from upstream")
		}
	}
	for _, name := range slices.Sorted(maps.Keys(got)) {
		if _, ok := want[name]; !ok {
			problems = append(problems, name+": not in upstream")
		}
	}
	return problems
}

// spriteFiles reads every file of a tree by path.
func spriteFiles(fsys fs.FS) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		files[p] = data
		return err
	})
	return files, err
}

// TestUpstreamInternalSprites checks that the bundled sprites are
// upstream's, as TestUpstreamStdlib checks the standard library. It needs
// the sources jar unpacked by tools/upstream/fetch.sh, which CI runs before
// the tests.
func TestUpstreamInternalSprites(t *testing.T) {
	dir := filepath.Join("..", "..", ".cache", "upstream", "src", "sprites")
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("upstream sources not available: %v", err)
	}
	bundled, err := fs.Sub(internalSprites, "sprites")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range compareSprites(bundled, os.DirFS(dir)) {
		t.Error(p)
	}
}

func TestSpriteImageFromInternal(t *testing.T) {
	if _, ok := SpriteImageFromInternal("archimate/actor").(ISvgSpriteParser); !ok {
		t.Error("archimate/actor is not an SVG sprite")
	}
	for _, name := range []string{"archimate/access", "archimate/assignment", "archimate/association-unidirect"} {
		s, ok := SpriteImageFromInternal(name).(*SpriteImage)
		if !ok {
			t.Errorf("%s is not an image sprite", name)
			continue
		}
		if img := s.img.image.Image(); img.Width() != 16 || img.Height() != 16 {
			t.Errorf("%s: %dx%d", name, img.Width(), img.Height())
		}
	}
	err := fs.WalkDir(internalSprites, "sprites", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := strings.TrimSuffix(strings.TrimPrefix(p, "sprites/"), path.Ext(p))
		if SpriteImageFromInternal(name) == nil {
			t.Errorf("%s: no sprite", name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"archimate/nope", "nothing", "../stdlib/home"} {
		if SpriteImageFromInternal(name) != nil {
			t.Errorf("%s: a sprite", name)
		}
	}
	if class, _ := expectPanic(t, func() { SpriteImageFromInternal("archimate/actor.svg") }); !strings.Contains(class, "IllegalArgumentException") {
		t.Errorf("class %q", class)
	}
}

func TestCompareSprites(t *testing.T) {
	upstream := fstest.MapFS{
		"archimate/actor.svg":  {Data: []byte("<svg/>")},
		"archimate/access.png": {Data: []byte{0x89, 'P', 'N', 'G'}},
	}
	if got := compareSprites(upstream, upstream); got != nil {
		t.Errorf("identical trees: %q", got)
	}
	bundled := fstest.MapFS{
		"archimate/actor.svg": {Data: []byte("<svg />")},
		"archimate/extra.svg": {Data: []byte("<svg/>")},
	}
	want := []string{
		"archimate/access.png: not bundled",
		"archimate/actor.svg: differs from upstream",
		"archimate/extra.svg: not in upstream",
	}
	if got := compareSprites(bundled, upstream); !slices.Equal(got, want) {
		t.Errorf("compareSprites = %q, want %q", got, want)
	}
}
