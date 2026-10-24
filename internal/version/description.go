// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/version/Version.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/version/CompilationInfo.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package version

// The build of the upstream release that the references come from
// (CompilationInfo.COMMIT, and COMPILE_TIMESTAMP 1792226081956 formatted as
// Version.compileTimeString does).
const (
	UpstreamCommit      = "18d0aae"
	UpstreamCompileTime = "2026-10-17 08:34:41 UTC"
)

// FullDescription is upstream's Version.fullDescription for that build, the
// version line of error images and of the unsupported-diagram image.
func FullDescription() string {
	return "PlantUML version " + Upstream + " / " + UpstreamCommit + " [" + UpstreamCompileTime + "]"
}
