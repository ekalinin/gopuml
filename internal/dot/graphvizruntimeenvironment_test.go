// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestFindDotPrefersGraphvizDotOverPath(t *testing.T) {
	// Fake host: the result is the GRAPHVIZ_DOT path.
	h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": "/custom/dot", "PATH": "/usr/bin:/fake/bin"},
		map[string]fs.FileMode{"/custom/dot": 0o755, "/usr/bin/dot": 0o755, "/fake/bin/dot": 0o755})
	if got := SearchDotExe("", h); got != "/custom/dot" {
		t.Fatalf("SearchDotExe = %q, want /custom/dot", got)
	}

	// Real files: the GRAPHVIZ_DOT script is the one that runs.
	custom := writeScript(t, workingDot("custom"))
	pathDir := t.TempDir()
	writeScriptIn(t, pathDir, workingDot("from-path"))
	e, err := Config{Host: envHost(map[string]string{"GRAPHVIZ_DOT": custom, "PATH": pathDir})}.NewExec()
	if err != nil {
		t.Fatal(err)
	}
	if e.Path != custom {
		t.Fatalf("Path = %q, want %q", e.Path, custom)
	}
	out, err := e.Layout(context.Background(), "digraph g { a -> b }")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<!-- custom -->") {
		t.Fatalf("the GRAPHVIZ_DOT executable did not run: %s", out)
	}
}

func TestFindDotSearchesPath(t *testing.T) {
	// The first PATH entry has a non-executable dot and a directory named
	// dot, both skipped as upstream does.
	h := fakeHost("darwin", map[string]string{"PATH": "/a:/b/dot-dir:/b:/c"},
		map[string]fs.FileMode{"/a/dot": 0o644, "/b/dot-dir/dot": fs.ModeDir | 0o755, "/b/dot": 0o755, "/c/dot": 0o755, "/usr/local/bin/dot": 0o755})
	if got := SearchDotExe("", h); got != "/b/dot" {
		t.Fatalf("SearchDotExe = %q, want /b/dot", got)
	}

	// Real files: the dot found on PATH is the one that runs.
	empty, pathDir := t.TempDir(), t.TempDir()
	want := writeScriptIn(t, pathDir, workingDot("from-path"))
	e, err := Config{Host: envHost(map[string]string{"PATH": empty + string(os.PathListSeparator) + pathDir})}.NewExec()
	if err != nil {
		t.Fatal(err)
	}
	if e.Path != want {
		t.Fatalf("Path = %q, want %q", e.Path, want)
	}
	out, err := e.Layout(context.Background(), "digraph g { a }")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<!-- from-path -->") {
		t.Fatalf("the PATH executable did not run: %s", out)
	}
}

func TestFindDotUsesGraphvizDotEvenWhenMissing(t *testing.T) {
	h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": "/missing/dot", "PATH": "/usr/bin"},
		map[string]fs.FileMode{"/usr/bin/dot": 0o755})
	if got := SearchDotExe("", h); got != "/missing/dot" {
		t.Fatalf("SearchDotExe = %q, want /missing/dot", got)
	}
	c := Config{Host: h}
	exe, err := c.Find()
	if exe != "/missing/dot" || !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find = %q, %v; want /missing/dot and ErrNotFound", exe, err)
	}
	nf, ok := errors.AsType[*NotFoundError](err)
	if !ok || nf.Path != "/missing/dot" || nf.State != DoesNotExist {
		t.Fatalf("error = %#v", err)
	}
	if e, err := c.NewExec(); e != nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("NewExec = %v, %v", e, err)
	}
	if got, want := nf.Error(), "cannot find Graphviz dot: File /missing/dot does not exist"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if got := nf.Lines()[:3]; !slices.Equal(got, []string{"Dot Executable: /missing/dot", "Dot executable does not exist", "Cannot find Graphviz. You should try"}) {
		t.Errorf("Lines() = %q", got)
	}
}

func TestFindDotExplicitPathWins(t *testing.T) {
	h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": "/custom/dot", "PATH": "/usr/bin"},
		map[string]fs.FileMode{"/custom/dot": 0o755, "/usr/bin/dot": 0o755})
	for _, explicit := range []string{"/explicit/dot", `"/explicit/dot"`, "  /explicit/dot\t", `" /explicit/dot "`, "\u00ab/explicit/dot\u00bb"} {
		if got := SearchDotExe(explicit, h); got != "/explicit/dot" {
			t.Errorf("SearchDotExe(%q) = %q, want /explicit/dot", explicit, got)
		}
	}
	// A blank explicit path is ignored.
	if got := SearchDotExe("  ", h); got != "/custom/dot" {
		t.Errorf("SearchDotExe(blank) = %q, want /custom/dot", got)
	}
}

func TestGetenvGraphvizDot(t *testing.T) {
	for _, tc := range []struct{ env, want string }{
		{"", ""},
		{" \t", ""},
		{`"C:\Program Files\Graphviz\bin\dot.exe"`, `C:\Program Files\Graphviz\bin\dot.exe`},
		{"(/x/dot)", "/x/dot"},
		{"[/x/dot]", "/x/dot"},
		{":/x/dot:", "/x/dot"},
		{`"/x/dot`, `"/x/dot`},
		{" /x/dot", " /x/dot"}, // the environment value is not trimmed
	} {
		h := fakeHost("linux", map[string]string{"GRAPHVIZ_DOT": tc.env}, nil)
		if got := GetenvGraphvizDot("", h); got != tc.want {
			t.Errorf("GetenvGraphvizDot with GRAPHVIZ_DOT=%q = %q, want %q", tc.env, got, tc.want)
		}
	}
}

func TestFindDotStandardLocations(t *testing.T) {
	all := []string{"/usr/local/bin/dot", "/opt/homebrew/bin/dot", "/opt/homebrew/opt/graphviz/bin/dot", "/usr/bin/dot"}
	for i, want := range all {
		files := map[string]fs.FileMode{"/opt/local/bin/dot": 0o755}
		for _, f := range all[i:] {
			files[f] = 0o755
		}
		// Earlier locations hold files that are not executable.
		for _, f := range all[:i] {
			files[f] = 0o644
		}
		h := fakeHost("darwin", map[string]string{"PATH": "/nowhere"}, files)
		if got := SearchDotExe("", h); got != want {
			t.Errorf("SearchDotExe = %q, want %q", got, want)
		}
	}

	// Without any, upstream returns /opt/local/bin/dot unchecked.
	h := fakeHost("linux", nil, nil)
	if got := SearchDotExe("", h); got != "/opt/local/bin/dot" {
		t.Fatalf("SearchDotExe = %q, want /opt/local/bin/dot", got)
	}
	exe, err := Config{Host: h}.Find()
	if nf, ok := errors.AsType[*NotFoundError](err); exe != "/opt/local/bin/dot" || !ok || nf.State != DoesNotExist {
		t.Fatalf("Find = %q, %v", exe, err)
	}
}

func TestFindDotWindows(t *testing.T) {
	// dot.exe on PATH, with ';' as the separator.
	h := fakeHost("windows", map[string]string{"PATH": `C:\Windows;C:\Graphviz\bin`},
		map[string]fs.FileMode{`C:\Windows\dot`: 0o666, `C:\Graphviz\bin\dot.exe`: 0o666})
	if got, want := SearchDotExe("", h), `C:\Graphviz\bin\dot.exe`; got != want {
		t.Errorf("SearchDotExe = %q, want %q", got, want)
	}

	// Install directories under c:/: the first top-level directory with a
	// Graphviz wins, and in it the highest path.
	h = fakeHost("windows", nil, map[string]fs.FileMode{
		"c:/Apps/readme.txt":                                0o666,
		"c:/Program Files/Graphviz2.38/bin/dot.exe":         0o666,
		"c:/Program Files/graphviz2.44/release/bin/dot.exe": 0o666,
		"c:/Program Files/Other/bin/dot.exe":                0o666,
		"c:/Tools/Graphviz/bin/dot.exe":                     0o666,
	})
	if got, want := SearchDotExe("", h), `c:\Program Files\graphviz2.44\release\bin\dot.exe`; got != want {
		t.Errorf("SearchDotExe = %q, want %q", got, want)
	}

	// Nothing installed: upstream's null.
	h = fakeHost("windows", nil, map[string]fs.FileMode{"c:/Windows/notepad.exe": 0o666})
	if got := SearchDotExe("", h); got != "" {
		t.Errorf("SearchDotExe = %q, want empty", got)
	}
	exe, err := Config{Host: h}.Find()
	nf, ok := errors.AsType[*NotFoundError](err)
	if exe != "" || !ok || nf.State != NullUndefined {
		t.Fatalf("Find = %q, %v", exe, err)
	}
	if got := nf.Lines()[:2]; !slices.Equal(got, []string{"Dot Executable: null", "No dot executable found"}) {
		t.Errorf("Lines() = %q", got)
	}
}

func TestCheckFile(t *testing.T) {
	h := fakeHost("linux", nil, map[string]fs.FileMode{
		"/ok/dot":         0o755,
		"/dir/dot":        fs.ModeDir | 0o755,
		"/pipe/dot":       fs.ModeNamedPipe | 0o644,
		"/unreadable/dot": 0o000,
	})
	for path, want := range map[string]ExeState{
		"":                NullUndefined,
		"/ok/dot":         OK,
		"/missing/dot":    DoesNotExist,
		"/dir/dot":        IsADirectory,
		"/pipe/dot":       NotAFile,
		"/unreadable/dot": CannotBeRead,
	} {
		if got := CheckFile(path, h); got != want {
			t.Errorf("CheckFile(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestExeStateMessages(t *testing.T) {
	for _, tc := range []struct {
		s           ExeState
		text, named string
	}{
		{OK, "Dot executable OK", "File /x/dot OK"},
		{NullUndefined, "No dot executable found", "No dot executable found"},
		{DoesNotExist, "Dot executable does not exist", "File /x/dot does not exist"},
		{IsADirectory, "Dot executable should be an executable, not a directory", "File /x/dot should be an executable, not a directory"},
		{NotAFile, "Dot executable is not a valid file", "File /x/dot is not a valid file"},
		{CannotBeRead, "Dot executable cannot be read", "File /x/dot cannot be read"},
	} {
		if got := tc.s.TextMessage(); got != tc.text {
			t.Errorf("%v.TextMessage() = %q, want %q", tc.s, got, tc.text)
		}
		if got := tc.s.TextMessageFor("/x/dot"); got != tc.named {
			t.Errorf("%v.TextMessageFor() = %q, want %q", tc.s, got, tc.named)
		}
	}
	rel := DoesNotExist.TextMessageFor("dot")
	if want := "File " + filepath.Join(mustGetwd(t), "dot") + " does not exist"; rel != want {
		t.Errorf("relative path: %q, want %q", rel, want)
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}

func TestNotFoundLines(t *testing.T) {
	want := []string{
		"Dot Executable: /opt/local/bin/dot",
		"Dot executable does not exist",
		"Cannot find Graphviz. You should try",
		" ",
		"@startuml",
		"testdot",
		"@enduml",
		" ",
		" or ",
		" ",
		"java -jar plantuml.jar -testdot",
		" ",
	}
	if got := NotFoundLines("/opt/local/bin/dot", DoesNotExist); !slices.Equal(got, want) {
		t.Errorf("NotFoundLines = %q", got)
	}
}
