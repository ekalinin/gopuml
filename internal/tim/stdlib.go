// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/Stdlib.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/json"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Stdlib is an index of the standard library files that "!include
// <library/path>" reads. Library names and paths are matched in lower case
// and without the .puml extension, as upstream does. A Stdlib is read-only
// once built.
type Stdlib struct {
	libraries map[string]*stdlibLibrary
}

type stdlibLibrary struct {
	name  string
	files map[string][]byte
	// info is the metadata of the library, read from its info.spm file
	// ("key=value" lines), in HashMap order; json holds its .json files.
	info []metadataEntry
	json map[string][]byte
}

// embeddedStdlib is the index bundled with this build. It is empty: no
// standard library is bundled yet.
var embeddedStdlib = &Stdlib{}

// NewStdlib returns an index of the given libraries: library name, then
// file path (with or without .puml) to content. A file named info.spm holds
// the metadata of the library as "key=value" lines (upstream's INFO
// channel); files ending with .json are the JSON resources that %load_json
// reads.
func NewStdlib(libraries map[string]map[string][]byte) *Stdlib {
	s := &Stdlib{libraries: map[string]*stdlibLibrary{}}
	for name, files := range libraries {
		lib := &stdlibLibrary{name: strings.ToLower(name), files: map[string][]byte{}, json: map[string][]byte{}}
		for p, data := range files {
			lp := strings.ToLower(p)
			switch {
			case lp == "info.spm":
				lib.info = parseStdlibInfo(data)
			case strings.HasSuffix(lp, ".json"):
				lib.json[lp] = data
			default:
				lib.files[strings.ReplaceAll(lp, ".puml", "")] = data
			}
		}
		s.libraries[lib.name] = lib
	}
	return s
}

// parseStdlibInfo reads the "key=value" lines of an info.spm file into a
// HashMap, as the Stdlib constructor does: lines that do not split into
// exactly two parts are ignored.
func parseStdlibInfo(data []byte) []metadataEntry {
	var keys []string
	values := map[string]string{}
	// BufferedReader.lines ends lines at \n, \r or \r\n
	lines := strings.ReplaceAll(text.DecodeUTF8(data), "\r\n", "\n")
	for _, line := range strings.Split(strings.ReplaceAll(lines, "\r", "\n"), "\n") {
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.Split(line, "=")
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		if len(parts) != 2 {
			continue
		}
		if _, ok := values[parts[0]]; !ok {
			keys = append(keys, parts[0])
		}
		values[parts[0]] = parts[1]
	}
	order := newJavaHashOrder()
	for _, k := range keys {
		order.add(javaHashCode(k))
	}
	info := make([]metadataEntry, 0, len(keys))
	for _, i := range order.order() {
		info = append(info, metadataEntry{keys[i], values[keys[i]]})
	}
	return info
}

// getAllFolderNames is Stdlib.getAllFolderNames: the library names, sorted.
func (s *Stdlib) getAllFolderNames() []string {
	names := make([]string, 0, len(s.libraries))
	for name := range s.libraries {
		names = append(names, name)
	}
	slices.SortFunc(names, javaCompare)
	return names
}

// getMetadata returns the value of an info key, and whether it is set.
func (l *stdlibLibrary) getMetadata(key string) (string, bool) {
	for _, e := range l.info {
		if e.key == key {
			return e.value, true
		}
	}
	return "", false
}

// getVersion is Stdlib.getVersion: the VERSION or version info.
func (l *stdlibLibrary) getVersion() (string, bool) {
	if v, ok := l.getMetadata("VERSION"); ok {
		return v, true
	}
	return l.getMetadata("version")
}

// getSource is Stdlib.getSource: the SOURCE or source info.
func (l *stdlibLibrary) getSource() (string, bool) {
	if v, ok := l.getMetadata("SOURCE"); ok {
		return v, true
	}
	return l.getMetadata("source")
}

// getJSONResource is Stdlib.getJsonResource: the JSON file "library/path",
// nil when fullname has no slash. As upstream, the names are not converted
// to lower case, a missing library throws UncheckedIOException, a missing
// file NullPointerException, and invalid JSON returns the parse error.
func (s *Stdlib) getJSONResource(fullname string) (json.Value, error) {
	last := strings.IndexByte(fullname, '/')
	if last == -1 {
		return nil, nil
	}
	lib := s.libraries[fullname[:last]]
	if lib == nil {
		throwRuntime("UncheckedIOException", "stdlib/"+fullname[:last]+"/info.spm")
	}
	data, ok := lib.json[fullname[last+1:]]
	if !ok {
		throwRuntime("NullPointerException", "")
	}
	return json.Parse(text.DecodeUTF8(data))
}

// hasLibrary reports whether the library with the lower-case name is
// bundled.
func (s *Stdlib) hasLibrary(name string) bool {
	_, ok := s.libraries[name]
	return ok
}

// retrieve returns the library with the lower-case name. Like upstream,
// which fails to read the library's index, it throws
// UncheckedIOException when the library does not exist.
func (s *Stdlib) retrieve(name string) *stdlibLibrary {
	lib := s.libraries[name]
	if lib == nil {
		throwRuntime("UncheckedIOException", "stdlib/"+name+"/info.spm")
	}
	return lib
}

// getPumlResource is Stdlib.getPumlResource: the file "library/path", or
// nil.
func (s *Stdlib) getPumlResource(fullname string) []byte {
	fullname = strings.ReplaceAll(strings.ToLower(fullname), ".puml", "")
	last := strings.IndexByte(fullname, '/')
	if last == -1 {
		return nil
	}
	folder := s.retrieve(fullname[:last])
	if folder == nil {
		return nil
	}
	return folder.loadPumlResource(fullname[last+1:])
}

func (l *stdlibLibrary) loadPumlResource(name string) []byte { return l.files[name] }

// notBundledMessage is the error of an include from a library that this
// build does not bundle (upstream bundles the libraries and includes them).
func notBundledMessage(what, library string) string {
	return "cannot include " + what + " (library " + library + " is not bundled)"
}
