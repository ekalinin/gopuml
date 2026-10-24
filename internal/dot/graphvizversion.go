// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizVersion.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizVersionFinder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizRuntimeEnvironment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"regexp"
	"strconv"
)

// GraphvizVersion tells the DOT generation which workarounds the installed
// Graphviz needs.
type GraphvizVersion interface {
	UseShieldForQuantifier() bool
	UseProtectionWhenThereALinkFromOrToGroup() bool
	UseXLabelInsteadOfLabel() bool
	IsVizjs() bool
	IgnoreHorizontalLinks() bool
}

// DefaultGraphvizVersion is used when the dot version cannot be determined
// (upstream's GraphvizVersionFinder.DEFAULT).
var DefaultGraphvizVersion GraphvizVersion = defaultGraphvizVersion{}

type defaultGraphvizVersion struct{}

func (defaultGraphvizVersion) UseShieldForQuantifier() bool                   { return true }
func (defaultGraphvizVersion) UseProtectionWhenThereALinkFromOrToGroup() bool { return true }
func (defaultGraphvizVersion) UseXLabelInsteadOfLabel() bool                  { return false }
func (defaultGraphvizVersion) IsVizjs() bool                                  { return false }
func (defaultGraphvizVersion) IgnoreHorizontalLinks() bool                    { return false }

// numberedGraphvizVersion is 100 * major + minor.
type numberedGraphvizVersion int

func (v numberedGraphvizVersion) UseShieldForQuantifier() bool { return v <= 228 }

func (v numberedGraphvizVersion) UseProtectionWhenThereALinkFromOrToGroup() bool {
	return v != 239 && v != 240
}

func (numberedGraphvizVersion) UseXLabelInsteadOfLabel() bool { return false }
func (numberedGraphvizVersion) IsVizjs() bool                 { return false }

func (v numberedGraphvizVersion) IgnoreHorizontalLinks() bool { return v == 230 }

var versionFinderPattern = regexp.MustCompile(`(\d+)\.(\d\d?)`)

// NewGraphvizVersion returns the workarounds for the dot version described by
// dotVersion, the output of `dot -V` (upstream's
// GraphvizVersionFinder.getVersion). Unparsable text gives
// DefaultGraphvizVersion.
func NewGraphvizVersion(dotVersion string) GraphvizVersion {
	m := versionFinderPattern.FindStringSubmatch(dotVersion)
	if m == nil {
		return DefaultGraphvizVersion
	}
	major, err1 := strconv.Atoi(m[1])
	minor, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return DefaultGraphvizVersion
	}
	return numberedGraphvizVersion(100*major + minor)
}

var retrieveVersionPattern = regexp.MustCompile(`\s(\d+)\.(\d\d?)\D`)

// RetrieveVersion extracts 100 * major + minor from `dot -V` output, or -1
// when s has no version (upstream's GraphvizRuntimeEnvironment.retrieveVersion).
// Unlike NewGraphvizVersion it needs a space before the version and a
// non-digit after it.
func RetrieveVersion(s string) int {
	m := retrieveVersionPattern.FindStringSubmatch(s)
	if m == nil {
		return -1
	}
	major, err1 := strconv.Atoi(m[1])
	minor, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return -1
	}
	return 100*major + minor
}
