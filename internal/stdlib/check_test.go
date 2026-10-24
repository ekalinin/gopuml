// SPDX-License-Identifier: MIT
// Provenance: original

package stdlib

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// TestUpstreamStdlib checks that the bundled files are upstream's, as
// TestUpstreamThemes checks the themes. It needs the sources jar unpacked
// by tools/upstream/fetch.sh, which CI runs before the tests.
func TestUpstreamStdlib(t *testing.T) {
	dir := filepath.Join("..", "..", ".cache", "upstream", "src", "stdlib")
	if _, err := os.Stat(filepath.Join(dir, homeFile)); err != nil {
		t.Skipf("upstream sources not available: %v", err)
	}
	for _, p := range CheckUpstream(files, os.DirFS(dir)) {
		t.Error(p)
	}
}

// upstreamFS is a small stdlib/ folder: two bundled libraries and the
// excluded ones.
func upstreamFS() fstest.MapFS {
	up := fstest.MapFS{
		"c4/info.spm":     {Data: []byte("i")},
		"c4/puml.spm":     {Data: []byte("p")},
		"c4/puml.sha1":    {Data: []byte("x")},
		"office/info.spm": {Data: []byte("o")},
	}
	home := []string{"c4", "office"}
	for _, e := range Excluded {
		up[e.Name+"/info.spm"] = &fstest.MapFile{Data: []byte("e")}
		up[e.Name+"/info.sha1"] = &fstest.MapFile{Data: []byte("x")}
		home = append(home, e.Name)
	}
	slices.Sort(home)
	up["home.spm"] = &fstest.MapFile{Data: []byte(strings.Join(home, "\n"))}
	return up
}

// bundledFS is the bundled copy of upstreamFS.
func bundledFS() fstest.MapFS {
	return fstest.MapFS{
		"home.spm":        {Data: []byte("c4\noffice")},
		"c4/info.spm":     {Data: []byte("i")},
		"c4/puml.spm":     {Data: []byte("p")},
		"office/info.spm": {Data: []byte("o")},
	}
}

func TestCheckStdlibIdentical(t *testing.T) {
	if p := CheckUpstream(bundledFS(), upstreamFS()); p != nil {
		t.Errorf("CheckUpstream = %q, want nil", p)
	}
}

// TestCheckStdlibChangedFile is the scenario "Modified library file".
func TestCheckStdlibChangedFile(t *testing.T) {
	ours := bundledFS()
	ours["c4/puml.spm"] = &fstest.MapFile{Data: []byte("P")}
	want := []string{"c4/puml.spm differs from upstream"}
	if p := CheckUpstream(ours, upstreamFS()); !slices.Equal(p, want) {
		t.Errorf("CheckUpstream = %q, want %q", p, want)
	}
}

func TestCheckStdlibProblems(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(ours, up fstest.MapFS)
		want   []string
	}{
		{"neither bundled nor excluded", func(_, up fstest.MapFS) {
			up["azure/info.spm"] = &fstest.MapFile{Data: []byte("z")}
		}, []string{"library azure is neither bundled nor excluded"}},
		{"excluded library bundled", func(ours, _ fstest.MapFS) {
			ours["adaml/info.spm"] = &fstest.MapFile{Data: []byte("a")}
		}, []string{"excluded library adaml is bundled"}},
		{"home differs", func(ours, _ fstest.MapFS) {
			ours["home.spm"] = &fstest.MapFile{Data: []byte("c4\noffice\n")}
		}, []string{"home.spm differs from the filtered upstream list"}},
		{"missing file", func(ours, _ fstest.MapFS) {
			delete(ours, "c4/puml.spm")
		}, []string{"c4/puml.spm is missing"}},
		{"extra file", func(ours, _ fstest.MapFS) {
			ours["c4/svg.spm"] = &fstest.MapFile{Data: []byte("s")}
		}, []string{"c4/svg.spm is not upstream"}},
		{"library not upstream", func(ours, _ fstest.MapFS) {
			ours["mine/info.spm"] = &fstest.MapFile{Data: []byte("m")}
		}, []string{"library mine is not upstream"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ours, up := bundledFS(), upstreamFS()
			tc.change(ours, up)
			if p := CheckUpstream(ours, up); !slices.Equal(p, tc.want) {
				t.Errorf("CheckUpstream = %q, want %q", p, tc.want)
			}
		})
	}
}

func TestFilterHome(t *testing.T) {
	for in, want := range map[string]string{
		"adaml\nc4\noffice":     "c4\noffice",
		"adaml\nc4\noffice\n":   "c4\noffice\n",
		"c4\noffice\nosa2":      "c4\noffice",
		"c4\r\nadaml\r\noffice": "c4\r\noffice",
	} {
		if got := string(FilterHome([]byte(in))); got != want {
			t.Errorf("FilterHome(%q) = %q, want %q", in, got, want)
		}
	}
}
