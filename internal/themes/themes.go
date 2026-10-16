// SPDX-License-Identifier: MIT
// Provenance: original

// Package themes holds the themes bundled with the program: the theme files
// of PlantUML v1.2026.8 whose license allows bundling them (see NOTICE for
// the themes that are left out and why). Bundled implements
// tim.ThemeSource and tim.ThemeLister, so that "!theme name" finds these
// themes first, as upstream finds the files in its jar's /themes/ folder,
// and %get_all_theme() lists them.
package themes

import (
	"embed"
	"io/fs"
	"slices"
	"strings"
)

// The theme files, unchanged from upstream's themes/ folder.
//
//go:embed puml-theme-*.puml
var files embed.FS

const (
	filePrefix = "puml-theme-"
	fileSuffix = ".puml"
)

// Source is the set of bundled themes. Its zero value is ready to use and
// safe for concurrent use.
type Source struct{}

// Bundled is the source of the bundled themes, for tim.Options.Themes.
var Bundled Source

// Theme returns the content of the bundled file puml-theme-<name>.puml.
// The name must match the file exactly (upstream reads the jar entry
// /themes/puml-theme-<name>.puml, which is case-sensitive and not
// normalized), so names with path elements such as ".." find nothing.
func (Source) Theme(name string) ([]byte, bool) {
	data, err := files.ReadFile(filePrefix + name + fileSuffix)
	if err != nil {
		return nil, false
	}
	return data, true
}

// ThemeNames returns the names of the bundled themes, without the
// puml-theme- prefix and the .puml suffix, sorted (upstream's
// ThemeUtils.getAllThemeNames).
func (Source) ThemeNames() []string {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		f := e.Name()
		if !e.IsDir() && strings.HasPrefix(f, filePrefix) && strings.HasSuffix(f, fileSuffix) {
			names = append(names, f[len(filePrefix):len(f)-len(fileSuffix)])
		}
	}
	slices.Sort(names)
	return names
}
