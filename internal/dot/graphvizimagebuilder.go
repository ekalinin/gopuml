// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/GraphvizImageBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

// NotFoundLines returns the text of upstream's error diagram for a dot
// executable that cannot be used (GraphvizImageBuilder.error). dotExe is the
// path that was tried ("" prints as "null", as upstream) and state its state.
func NotFoundLines(dotExe string, state ExeState) []string {
	return []string{
		"Dot Executable: " + javaFileString(dotExe),
		state.TextMessage(),
		"Cannot find Graphviz. You should try",
		" ",
		"@startuml",
		"testdot",
		"@enduml",
		" ",
		" or ",
		" ",
		"java -jar plantuml.jar -testdot",
		" ",
	}
}
