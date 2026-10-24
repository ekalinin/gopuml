// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizRuntimeEnvironment.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/AbstractGraphviz.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizLinux.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizWindowsOld.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"context"
	"io/fs"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"
)

// Host gives access to the parts of the host system that are used to find
// dot. Nil functions and an empty GOOS default to the operating system, so the
// zero Host is the real host. Tests replace the fields with fakes.
type Host struct {
	// GOOS selects the search rules: "windows" looks for dot.exe and scans the
	// Graphviz install directories under c:/. Empty means runtime.GOOS.
	GOOS string
	// Getenv returns an environment variable, "" when unset (os.Getenv).
	Getenv func(key string) string
	// Stat returns information about a file, following links (os.Stat).
	Stat func(name string) (fs.FileInfo, error)
	// ReadDir lists a directory (os.ReadDir). Only the Windows search uses it.
	ReadDir func(name string) ([]fs.DirEntry, error)
}

func (h Host) getenv(key string) string {
	if h.Getenv != nil {
		return h.Getenv(key)
	}
	return os.Getenv(key)
}

func (h Host) stat(name string) (fs.FileInfo, error) {
	if h.Stat != nil {
		return h.Stat(name)
	}
	return os.Stat(name)
}

func (h Host) readDir(name string) ([]fs.DirEntry, error) {
	if h.ReadDir != nil {
		return h.ReadDir(name)
	}
	return os.ReadDir(name)
}

func (h Host) isWindows() bool {
	goos := h.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	return goos == "windows"
}

// canRead approximates java.io.File.canRead with the permission bits. On
// Windows every existing file is readable.
func (h Host) canRead(fi fs.FileInfo) bool {
	return h.isWindows() || fi.Mode().Perm()&0o444 != 0
}

// canExecute approximates java.io.File.canExecute with the permission bits.
// On Windows every existing file counts as executable, as in Java.
func (h Host) canExecute(fi fs.FileInfo) bool {
	return h.isWindows() || fi.Mode().Perm()&0o111 != 0
}

// pathSeparator is java.io.File.pathSeparator of the host.
func (h Host) pathSeparator() string {
	if h.isWindows() {
		return ";"
	}
	return ":"
}

// join mirrors new File(parent, child): child is appended with the host
// separator. On Windows forward slashes become backslashes, as Java
// normalizes them.
func (h Host) join(parent, child string) string {
	sep := "/"
	if h.isWindows() {
		sep = `\`
	}
	var p string
	if strings.HasSuffix(parent, "/") || strings.HasSuffix(parent, sep) {
		p = parent + child
	} else {
		p = parent + sep + child
	}
	if h.isWindows() {
		p = strings.ReplaceAll(p, "/", `\`)
	}
	return p
}

// exeName is the dot executable name searched on PATH.
func (h Host) exeName() string {
	if h.isWindows() {
		return "dot.exe"
	}
	return "dot"
}

// Config holds the Graphviz settings of one render. It replaces upstream's
// GraphvizRuntimeEnvironment singleton and the global TIMEOUT_MS setting.
type Config struct {
	// DotPath is the dot executable chosen by the caller (the --dot-path CLI
	// flag or the library option). Empty means search for it.
	DotPath string
	// Timeout limits each run of dot. Zero or negative means DefaultTimeout.
	Timeout time.Duration
	// Host is the system to search. The zero value is the real host.
	Host Host
}

// Find returns the dot executable to use, searched as SearchDotExe does. When
// that path is not usable it also returns a *NotFoundError, which carries the
// path and its state for the error diagram (see NotFoundLines).
func (c Config) Find() (string, error) {
	exe := SearchDotExe(c.DotPath, c.Host)
	if state := CheckFile(exe, c.Host); state != OK {
		return exe, &NotFoundError{Path: exe, State: state}
	}
	return exe, nil
}

// NewExec finds dot (see Find) and returns the dotexec engine running it. It
// returns a *NotFoundError when no usable dot executable is found.
func (c Config) NewExec() (*Exec, error) {
	exe, err := c.Find()
	if err != nil {
		return nil, err
	}
	return &Exec{Path: exe, Timeout: c.Timeout}, nil
}

// DotVersion returns the output of `dot -V` on one line, "?" when dot cannot
// be run, or "Error:" followed by the reason when the executable is not usable.
func (c Config) DotVersion(ctx context.Context) string {
	exe := SearchDotExe(c.DotPath, c.Host)
	if state := CheckFile(exe, c.Host); state != OK {
		return "Error:" + state.TextMessageFor(exe)
	}
	return (&Exec{Path: exe, Timeout: c.Timeout}).DotVersion(ctx)
}

// GetenvGraphvizDot returns the dot path configured by the caller or by the
// GRAPHVIZ_DOT environment variable, in that order, without checking that a
// file exists there. Surrounding quotes are removed. It returns "" when
// neither is set.
func GetenvGraphvizDot(explicit string, h Host) string {
	// Run.java removes the quotes of --dot-path, then setDotExecutable trims
	// the value, then the quotes are removed once more below.
	if v := strings.TrimFunc(eventuallyRemoveStartingAndEndingDoubleQuote(explicit), func(r rune) bool { return r <= ' ' }); isNotEmpty(v) {
		return eventuallyRemoveStartingAndEndingDoubleQuote(v)
	}
	if v := h.getenv("GRAPHVIZ_DOT"); isNotEmpty(v) {
		return eventuallyRemoveStartingAndEndingDoubleQuote(v)
	}
	return ""
}

// SearchDotExe returns the dot executable path upstream would use, without
// checking that it is usable:
//
//  1. the explicit path
//  2. the GRAPHVIZ_DOT environment variable, even if no file exists there
//  3. the first executable dot (dot.exe on Windows) in the PATH directories
//  4. the standard locations of the host
//
// On Unix the result is never empty: when no standard location has dot,
// "/opt/local/bin/dot" is returned, as upstream does. On Windows "" (upstream's
// null) is returned when no Graphviz install directory has dot.exe.
func SearchDotExe(explicit string, h Host) string {
	exe := GetenvGraphvizDot(explicit, h)
	if exe == "" {
		exe = findExecutableOnPath(h, h.exeName())
	}
	if exe == "" {
		return specificDotExe(h)
	}
	return exe
}

// findExecutableOnPath returns the first file named name that is executable
// in the PATH directories, or "".
func findExecutableOnPath(h Host, name string) string {
	path := h.getenv("PATH")
	if path == "" {
		return ""
	}
	for dirname := range strings.SplitSeq(path, h.pathSeparator()) {
		// Upstream resolves an empty entry against the root directory; the
		// root is never a sensible place for dot, so empty entries are skipped.
		if dirname == "" {
			continue
		}
		file := h.join(dirname, name)
		if fi, err := h.stat(file); err == nil && fi.Mode().IsRegular() && h.canExecute(fi) {
			return file
		}
	}
	return ""
}

// specificDotExe returns the dot executable in the standard locations of the
// host.
func specificDotExe(h Host) string {
	if h.isWindows() {
		return specificDotExeSlow(h)
	}
	// GraphvizLinux.specificDotExe
	for _, f := range []string{"/usr/local/bin/dot", "/opt/homebrew/bin/dot", "/opt/homebrew/opt/graphviz/bin/dot", "/usr/bin/dot"} {
		if fi, err := h.stat(f); err == nil && h.canRead(fi) && h.canExecute(fi) {
			return f
		}
	}
	return "/opt/local/bin/dot"
}

// specificDotExeSlow searches c:/<dir>/graphviz*/bin/dot.exe and
// c:/<dir>/graphviz*/release/bin/dot.exe, as GraphvizWindowsOld does. The
// first top-level directory with a match wins; within it the highest path
// wins.
func specificDotExeSlow(h Host) string {
	entries, err := h.readDir("c:/")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		tmp := h.join("c:/", e.Name())
		if fi, err := h.stat(tmp); err != nil || !fi.IsDir() || !h.canRead(fi) {
			continue
		}
		if result := searchInDir(h, tmp); result != "" {
			return result
		}
	}
	return ""
}

func searchInDir(h Host, dir string) string {
	files, err := h.readDir(dir)
	if err != nil {
		return ""
	}
	var dots []string
	for _, e := range files {
		if !strings.HasPrefix(strings.ToLower(e.Name()), "graphviz") {
			continue
		}
		f := h.join(dir, e.Name())
		if fi, err := h.stat(f); err != nil || !fi.IsDir() {
			continue
		}
		for _, result := range []string{h.join(h.join(f, "bin"), "dot.exe"), h.join(h.join(h.join(f, "release"), "bin"), "dot.exe")} {
			if fi, err := h.stat(result); err == nil && h.canRead(fi) {
				dots = append(dots, result)
			}
		}
	}
	return higherVersion(dots)
}

// higherVersion returns the greatest path, compared without case as Java
// compares Windows paths.
func higherVersion(dots []string) string {
	if len(dots) == 0 {
		return ""
	}
	return slices.MaxFunc(dots, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
}

// trin is StringUtils.trin: it removes leading and trailing characters up to
// and including the space, and U+00A0.
func trin(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' || r == '\u00A0' })
}

// isNotEmpty is StringUtils.isNotEmpty: s has a character other than space,
// tab, CR, LF and NUL.
func isNotEmpty(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool {
		return r != ' ' && r != '\t' && r != '\r' && r != '\n' && r != 0
	})
}

// eventuallyRemoveStartingAndEndingDoubleQuote is the one-argument form of
// the StringUtils method (format "\"([:"): it removes one pair of surrounding
// quotes, parentheses, brackets or colons.
func eventuallyRemoveStartingAndEndingDoubleQuote(s string) string {
	r := []rune(s)
	if len(r) < 2 {
		return s
	}
	first, last := r[0], r[len(r)-1]
	if isDoubleQuote(first) && isDoubleQuote(last) ||
		first == '(' && last == ')' ||
		first == '[' && last == ']' ||
		first == ':' && last == ':' {
		return string(r[1 : len(r)-1])
	}
	return s
}

func isDoubleQuote(c rune) bool {
	return c == '"' || c == '\u201c' || c == '\u201d' || c == '\u00ab' || c == '\u00bb'
}
