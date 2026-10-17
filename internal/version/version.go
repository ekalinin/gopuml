// SPDX-License-Identifier: MIT
// Provenance: original

// Package version describes this implementation and the upstream release it
// reproduces.
package version

import "runtime/debug"

// Upstream is the PlantUML release this implementation reproduces;
// %version() returns it.
const Upstream = "1.2026.8"

// Version is the version of this implementation. It can be set at build
// time with -ldflags "-X github.com/ekalinin/gopuml/internal/version.Version=v1.0.0";
// otherwise it comes from the module build information.
var Version = ""

// Semver returns the version of this implementation, or "devel" for a
// build without one.
func Semver() string {
	if Version != "" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "devel"
}

// Full is the line printed by --version (design D15).
func Full() string { return "gopuml " + Semver() + " (PlantUML " + Upstream + " compatible)" }
