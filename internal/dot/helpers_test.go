// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

// fakeHost returns a Host whose environment is env and whose files are the
// keys of files, with the given modes. Parent directories are implied.
func fakeHost(goos string, env map[string]string, files map[string]fs.FileMode) Host {
	m := fstest.MapFS{}
	for name, mode := range files {
		m[fsName(name)] = &fstest.MapFile{Mode: mode}
	}
	return Host{
		GOOS:    goos,
		Getenv:  func(key string) string { return env[key] },
		Stat:    func(name string) (fs.FileInfo, error) { return m.Stat(fsName(name)) },
		ReadDir: func(name string) ([]fs.DirEntry, error) { return m.ReadDir(fsName(name)) },
	}
}

// fsName maps a host path ("/usr/bin/dot", `C:\Graphviz\bin\dot.exe`) to a
// MapFS name.
func fsName(name string) string {
	name = strings.Trim(strings.ReplaceAll(name, `\`, "/"), "/")
	if name == "" {
		return "."
	}
	return name
}

// envHost returns the real host with a fake environment.
func envHost(env map[string]string) Host {
	return Host{Getenv: func(key string) string { return env[key] }}
}

// writeScript writes an executable shell script named dot in a new temporary
// directory and returns its path.
func writeScript(t *testing.T, body string) string {
	t.Helper()
	return writeScriptIn(t, t.TempDir(), body)
}

func writeScriptIn(t *testing.T, dir, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake dot scripts need a POSIX shell")
	}
	path := filepath.Join(dir, "dot")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

const testVersion = "dot - graphviz version 14.0.1 (20251006.0113)"

const testSVG = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="62pt" height="44pt" viewBox="0.00 0.00 62.00 44.00" xmlns="http://www.w3.org/2000/svg"></svg>`

// workingDot is a fake dot that prints testVersion for -V, as dot does on
// stderr, and a minimal SVG otherwise. The marker is printed in an XML
// comment so tests can tell which script ran.
func workingDot(marker string) string {
	return `if [ "$1" = "-V" ]; then echo "` + testVersion + `" >&2; exit 0; fi
cat >/dev/null
echo '` + testSVG + `'
echo '<!-- ` + marker + ` -->'
`
}
