// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/Defines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/Define.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/DefinitionsContainer.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"maps"
	"strings"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Version is the upstream version this preprocessor targets.
const Version = "1.2026.8"

// Defines holds the definitions given outside the source (-D) and the
// environment of the source file (%filename, %dirpath, ...).
type Defines struct {
	environment map[string]string
	names       []string // definition names, in insertion order
	values      map[string]*string
}

// NewDefines returns empty definitions (Defines.createEmpty).
func NewDefines() *Defines {
	return &Defines{environment: map[string]string{"PLANTUML_VERSION": Version}, values: map[string]*string{}}
}

// DefinesWithFileName returns the definitions of a source file
// (Defines.createWithFileName): its name, its modification date and, when
// dirpath is not empty, its directory (upstream sets it in the INSECURE
// profile only).
func DefinesWithFileName(name string, modTime time.Time, dirpath string) *Defines {
	d := NewDefines()
	d.OverrideFilename(name)
	d.environment["filedate"] = javaDateToString(modTime, zoneOfLocation(modTime.Location()))
	if dirpath != "" {
		d.environment["dirpath"] = strings.ReplaceAll(dirpath, `\`, "/")
	}
	return d
}

// OverrideFilename sets %filename and %filename_no_extension.
func (d *Defines) OverrideFilename(filename string) {
	if filename != "" {
		d.environment["filename"] = filename
		d.environment["filenameNoExtension"] = nameNoExtension(filename)
	}
}

// OverrideDirPath sets %dirpath.
func (d *Defines) OverrideDirPath(fileDir string) {
	if fileDir != "" {
		d.environment["dirpath"] = strings.ReplaceAll(fileDir, `\`, "/")
	}
}

// EnvironmentValue returns an environment value, "" when it is not set.
func (d *Defines) EnvironmentValue(key string) (string, bool) {
	v, ok := d.environment[key]
	return v, ok
}

func nameNoExtension(name string) string {
	if x := strings.LastIndex(name, "."); x != -1 {
		return name[:x]
	}
	return name
}

// Define adds a definition; its lines are joined with newlines.
func (d *Defines) Define(name string, value []string) {
	if _, ok := d.values[name]; !ok {
		d.names = append(d.names, name)
	}
	if value == nil {
		d.values[name] = nil
		return
	}
	v := strings.Join(value, "\n")
	d.values[name] = &v
}

// Undefine removes a definition.
func (d *Defines) Undefine(name string) {
	delete(d.values, name)
	for i, n := range d.names {
		if n == name {
			d.names = append(d.names[:i:i], d.names[i+1:]...)
			break
		}
	}
}

// CloneMe returns an independent copy.
func (d *Defines) CloneMe() *Defines {
	return &Defines{environment: maps.Clone(d.environment), names: append([]string(nil), d.names...),
		values: maps.Clone(d.values)}
}

// copyTo puts the definitions into the memory as global variables.
func (d *Defines) copyTo(memory TMemory, location text.StringLocated) error {
	for _, name := range d.names {
		def := d.values[name]
		if def == nil {
			throwRuntime("NullPointerException", "")
		}
		if err := memory.PutVariable(name, TValueFromString(*def), scopeGlobal, location); err != nil {
			return err
		}
	}
	return nil
}

// DefinitionsContainer gives !includedef the lines of the @startdef blocks
// that precede the block in the same source.
type DefinitionsContainer interface {
	// GetDefinition returns the lines of the definition, without its start
	// and end lines, or nil.
	GetDefinition(name string) []string
}
