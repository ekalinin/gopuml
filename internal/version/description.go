// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/version/Version.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/version/CompilationInfo.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package version

// The build of the upstream release that the references come from
// (CompilationInfo.COMMIT, and COMPILE_TIMESTAMP 1788626773766 formatted as
// Version.compileTimeString does).
const (
	UpstreamCommit      = "149874a"
	UpstreamCompileTime = "2026-09-05 16:46:13 UTC"
)

// FullDescription is upstream's Version.fullDescription for that build, the
// version line of error images and of the unsupported-diagram image.
func FullDescription() string {
	return "PlantUML version " + Upstream + " / " + UpstreamCommit + " [" + UpstreamCompileTime + "]"
}
