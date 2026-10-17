// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/file/FileGroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliOptions.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// directoryPattern selects the files of a directory argument, as upstream's
// CliOptions.getPattern.
var directoryPattern = regexp.MustCompile(`(?i)^.*\.(txt|tex|java|htm|html|c|h|cpp|apt|pu|puml|hpp|hh)$`)

var (
	predirPath        = regex.MustCompile(`^([^*?]*[/\\])?(.*)$`)
	noStarInDirectory = regex.MustCompile(`^(?:([^*?]*)[/\\])?([^/\\]*)$`)
)

// globToRegexp is upstream's FileGroup.toRegexp: the replacements are
// applied in this order, and other regular expression characters are not
// escaped.
func globToRegexp(pattern string) string {
	pattern = strings.ReplaceAll(pattern, `\`, "/")
	pattern = strings.ReplaceAll(pattern, ".", `\.`)
	pattern = strings.ReplaceAll(pattern, "?", "[^/]")
	pattern = strings.ReplaceAll(pattern, "/**/", "(/|/.{0,}/)")
	pattern = strings.ReplaceAll(pattern, "**", ".{0,}")
	pattern = strings.ReplaceAll(pattern, "*", "[^/]{0,}")
	return `(?i)^(\./)?` + pattern + "$"
}

// matchesGlob reports whether s matches upstream's translation of a glob.
// A glob that is not a valid expression matches nothing.
func matchesGlob(s, glob string) bool {
	re, err := regexp.Compile(globToRegexp(glob))
	return err == nil && re.MatchString(s)
}

// javaPath normalizes a path like java.io.File: repeated and trailing
// separators are removed.
func javaPath(p string) string {
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// childPath is java.io.File(parent, name).getPath().
func childPath(dir, name string) string {
	if dir == "" {
		return name
	}
	if strings.HasSuffix(dir, "/") {
		return dir + name
	}
	return dir + "/" + name
}

// javaCompare orders strings like Java's String.compareTo (UTF-16 units).
func javaCompare(a, b string) int {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	return slices.Compare(ua, ub)
}

// fileGroup selects the input files of one command line argument, as
// upstream's FileGroup: a file, a directory (its files matching the input
// pattern, not recursive), a pattern with * or ? (inside one directory), or
// a pattern with ** (recursive). Excluded patterns remove matches. The
// paths are returned sorted, as java.io.File paths.
func fileGroup(pattern string, excluded []string) []string {
	var result []string
	add := func(p string) {
		normalized := strings.ReplaceAll(p, `\`, "/")
		for _, x := range excluded {
			if matchesGlob(normalized, x) {
				return
			}
		}
		result = append(result, p)
	}
	addDirectory := func(dir string, match func(name string) bool) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if match(e.Name()) {
				add(childPath(dir, e.Name()))
			}
		}
	}
	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			child := childPath(dir, e.Name())
			switch {
			case e.IsDir():
				walk(child)
			case e.Type().IsRegular():
				if matchesGlob(strings.ReplaceAll(child, `\`, "/"), pattern) {
					add(child)
				}
			}
		}
	}
	recurse := func() {
		m := predirPath.FindStringSubmatch(pattern)
		parent := "."
		if m != nil && m[1] != "" {
			parent = javaPath(m[1])
		}
		walk(parent)
	}

	switch {
	case !strings.ContainsAny(pattern, "*?"):
		p := javaPath(pattern)
		if st, err := os.Stat(p); err == nil {
			if st.IsDir() {
				addDirectory(p, directoryPattern.MatchString)
			} else if st.Mode().IsRegular() {
				add(p)
			}
		}
	case strings.Contains(pattern, "**"):
		recurse()
	default:
		if m := noStarInDirectory.FindStringSubmatch(pattern); m != nil {
			dir := "."
			if m[1] != "" {
				dir = javaPath(m[1])
			}
			files := m[2]
			addDirectory(dir, func(name string) bool { return matchesGlob(name, files) })
		} else {
			recurse()
		}
	}
	slices.SortFunc(result, javaCompare)
	return result
}

// InputFiles returns the input files of all non-flag arguments, in argument
// order, each argument's files sorted, with the -x exclusions applied.
func (p *Parsed) InputFiles() []string {
	var excluded []string
	for _, v := range p.List(Exclude) {
		excluded = append(excluded, v.S)
	}
	var out []string
	for _, arg := range p.Args() {
		out = append(out, fileGroup(arg, excluded)...)
	}
	return out
}
