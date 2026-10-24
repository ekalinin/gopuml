// SPDX-License-Identifier: MIT
// Provenance: original

// Package stdlib holds the standard library bundled with the program: the
// stdlib/ folder of the plantuml-mit 1.2026.9 jar, without the libraries
// whose license does not allow bundling them (see NOTICE). Each library is
// a folder of Brotli-compressed channel files (info.spm, puml.spm,
// json.spm, sprite.spm, svg.spm, image.spm), embedded unchanged, and
// home.spm lists the library names. tools/stdlib imports the files.
package stdlib

import (
	"embed"
	"slices"
)

// The library files, unchanged from upstream's stdlib/ folder, except
// home.spm, which lists the bundled libraries only.
//
//go:embed home.spm */*.spm
var files embed.FS

// Exclusion is a library of upstream's stdlib/ folder that is not bundled.
type Exclusion struct {
	Name, Reason string
}

// Excluded lists the libraries of upstream v1.2026.9 that are not bundled,
// in name order, for the license reasons given in NOTICE.
var Excluded = []Exclusion{
	{"adaml", "GPL-3.0 (the source repository and its info.spm)"},
	{"classy", "no license could be verified"},
	{"classy-c4", "no license could be verified"},
	{"cloudogu", "no license could be verified"},
	{"edgy", "no license could be verified"},
	{"elastic", "no license could be verified"},
	{"gcp", "no license could be verified"},
	{"osa", "no license could be verified"},
	{"osa2", "no license could be verified"},
}

// IsExcluded reports whether name, in lower case, is a library of
// upstream's stdlib/ folder that this build does not bundle.
func IsExcluded(name string) bool {
	return slices.ContainsFunc(Excluded, func(e Exclusion) bool { return e.Name == name })
}
