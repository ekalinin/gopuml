// SPDX-License-Identifier: MIT
// Provenance: original

package stdlib

import (
	"bytes"
	"io/fs"
	"path"
	"slices"
	"strings"
)

// homeFile is upstream's list of the library names, one per line.
const homeFile = "home.spm"

// CheckUpstream compares the bundled files with upstream's stdlib/ folder
// and returns one message per difference, nil when there is none: every
// .spm file of a bundled library must be byte-identical to upstream's, every
// upstream library must be either bundled or excluded, no excluded library
// may be bundled, and home.spm must hold upstream's lines without the
// excluded names, in upstream's order.
func CheckUpstream(bundled, upstream fs.FS) []string {
	var problems []string
	upDirs, err := libraryNames(upstream)
	if err != nil {
		return []string{"upstream: " + err.Error()}
	}
	ourDirs, err := libraryNames(bundled)
	if err != nil {
		return []string{"bundled: " + err.Error()}
	}
	for _, name := range upDirs {
		bundledDir := slices.Contains(ourDirs, name)
		switch {
		case IsExcluded(name):
			if bundledDir {
				problems = append(problems, "excluded library "+name+" is bundled")
			}
		case !bundledDir:
			problems = append(problems, "library "+name+" is neither bundled nor excluded")
		default:
			problems = append(problems, compareLibrary(bundled, upstream, name)...)
		}
	}
	for _, name := range ourDirs {
		if !slices.Contains(upDirs, name) {
			problems = append(problems, "library "+name+" is not upstream")
		}
	}
	for _, e := range Excluded {
		if !slices.Contains(upDirs, e.Name) {
			problems = append(problems, "excluded library "+e.Name+" is not upstream")
		}
	}
	upHome, err := fs.ReadFile(upstream, homeFile)
	if err != nil {
		return append(problems, "upstream: "+err.Error())
	}
	ourHome, err := fs.ReadFile(bundled, homeFile)
	if err != nil {
		return append(problems, "bundled: "+err.Error())
	}
	if !bytes.Equal(ourHome, FilterHome(upHome)) {
		problems = append(problems, homeFile+" differs from the filtered upstream list")
	}
	return problems
}

// FilterHome returns upstream's home.spm without the lines of the excluded
// libraries, keeping the order and the line ends of the other lines.
func FilterHome(home []byte) []byte {
	var out []byte
	for line := range bytes.Lines(home) {
		if !IsExcluded(strings.TrimRight(string(line), "\r\n")) {
			out = append(out, line...)
		}
	}
	// The last kept line ends as upstream's last line does.
	if len(out) > 0 && !bytes.HasSuffix(home, []byte("\n")) {
		out = bytes.TrimRight(out, "\r\n")
	}
	return out
}

// compareLibrary compares the .spm files of one library.
func compareLibrary(bundled, upstream fs.FS, name string) []string {
	var problems []string
	upFiles, err := spmFiles(upstream, name)
	if err != nil {
		return []string{"upstream: " + err.Error()}
	}
	ourFiles, err := spmFiles(bundled, name)
	if err != nil {
		return []string{"bundled: " + err.Error()}
	}
	for _, f := range upFiles {
		p := path.Join(name, f)
		if !slices.Contains(ourFiles, f) {
			problems = append(problems, p+" is missing")
			continue
		}
		want, err := fs.ReadFile(upstream, p)
		if err != nil {
			problems = append(problems, "upstream: "+err.Error())
			continue
		}
		got, err := fs.ReadFile(bundled, p)
		if err != nil {
			problems = append(problems, "bundled: "+err.Error())
			continue
		}
		if !bytes.Equal(got, want) {
			problems = append(problems, p+" differs from upstream")
		}
	}
	for _, f := range ourFiles {
		if !slices.Contains(upFiles, f) {
			problems = append(problems, path.Join(name, f)+" is not upstream")
		}
	}
	return problems
}

// libraryNames returns the folder names of a stdlib/ tree, sorted.
func libraryNames(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// spmFiles returns the .spm file names of a library folder, sorted.
func spmFiles(fsys fs.FS, name string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".spm") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
