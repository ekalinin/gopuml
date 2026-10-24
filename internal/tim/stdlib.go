// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/Stdlib.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"slices"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/json"
	bundled "github.com/ekalinin/gopuml/internal/stdlib"
	"github.com/ekalinin/gopuml/internal/text"
)

// Stdlib is an index of the standard library files that "!include
// <library/path>" reads. Library names and paths are matched in lower case
// and without the .puml extension, as upstream does. A Stdlib is read-only
// once built.
type Stdlib struct {
	libraries map[string]*stdlibLibrary
	// embedded is set for the bundled index, which reads its libraries
	// from internal/stdlib on first use; mu guards libraries then.
	embedded bool
	mu       sync.Mutex
}

type stdlibLibrary struct {
	name  string
	files map[string][]byte
	// info is the metadata of the library, read from its info.spm file
	// ("key=value" lines), in HashMap order; json holds its .json files.
	info []metadataEntry
	json map[string][]byte
	// lib is the library of the bundled index, whose files and JSON
	// resources it reads.
	lib *bundled.Library
}

// embeddedStdlib is the index of the standard library bundled with this
// build (internal/stdlib).
var embeddedStdlib = &Stdlib{libraries: map[string]*stdlibLibrary{}, embedded: true}

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
	if s.embedded {
		names := bundled.FolderNames()
		slices.SortFunc(names, javaCompare)
		return names
	}
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
	lib := s.retrieve(fullname[:last])
	if len(lib.info) == 0 {
		return nil, nil
	}
	data, ok, err := lib.loadJSONResource(fullname[last+1:])
	if err != nil {
		// Upstream logs the IOException and returns null.
		return nil, nil
	}
	if !ok {
		throwRuntime("NullPointerException", "")
	}
	return json.Parse(text.DecodeUTF8(data))
}

// loadJSONResource is loadJsonResource: the JSON file of a name, and
// whether the library has it.
func (l *stdlibLibrary) loadJSONResource(name string) ([]byte, bool, error) {
	if l.lib != nil {
		data, err := l.lib.JSON(name)
		return data, data != nil, err
	}
	data, ok := l.json[name]
	return data, ok, nil
}

// retrieve returns the library with the lower-case name. Like upstream,
// which fails to read the library's index, it throws
// UncheckedIOException when the library does not exist.
func (s *Stdlib) retrieve(name string) *stdlibLibrary {
	if s.embedded {
		return s.retrieveEmbedded(name)
	}
	lib := s.libraries[name]
	if lib == nil {
		throwRuntime("UncheckedIOException", "stdlib/"+name+"/info.spm")
	}
	return lib
}

// retrieveEmbedded reads a library of the bundled index on first use,
// following its link, as Stdlib.retrieve does.
func (s *Stdlib) retrieveEmbedded(name string) *stdlibLibrary {
	s.mu.Lock()
	defer s.mu.Unlock()
	if lib := s.libraries[name]; lib != nil {
		return lib
	}
	lib, err := bundled.Retrieve(name)
	if err != nil {
		if nf, ok := err.(*bundled.FileNotFoundError); ok {
			throwRuntime("UncheckedIOException", nf.Path)
		}
		throwRuntime("UncheckedIOException", err.Error())
	}
	l := &stdlibLibrary{name: lib.Name(), info: parseStdlibInfo(lib.Info()), lib: lib}
	s.libraries[name] = l
	return l
}

// getPumlResource is Stdlib.getPumlResource: the file "library/path", or
// nil, also for a library without metadata.
func (s *Stdlib) getPumlResource(fullname string) []byte {
	fullname = strings.ReplaceAll(strings.ToLower(fullname), ".puml", "")
	last := strings.IndexByte(fullname, '/')
	if last == -1 {
		return nil
	}
	folder := s.retrieve(fullname[:last])
	if folder == nil || len(folder.info) == 0 {
		return nil
	}
	return folder.loadPumlResource(fullname[last+1:])
}

// loadPumlResource returns the file of a lower-case path without
// extension, nil when there is none. For a bundled library, a channel that
// cannot be read gives nil, as upstream logs the IOException.
func (l *stdlibLibrary) loadPumlResource(name string) []byte {
	if l.lib != nil {
		data, _ := l.lib.Puml(name)
		return data
	}
	return l.files[name]
}

// notBundledMessage is the error of an include from a library that this
// build excludes for its license (upstream bundles it and includes it).
func notBundledMessage(what, library string) string {
	return "cannot include " + what + " (library " + library + " is not bundled)"
}
