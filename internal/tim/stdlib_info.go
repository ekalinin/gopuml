// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/Stdlib.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "strings"

// javaNullString is Java's string conversion of a possibly null value.
func javaNullString(s string, ok bool) string {
	if !ok {
		return "null"
	}
	return s
}

// StdlibInfoVersion is Stdlib.addInfoVersion for the embedded standard
// library: with details, the name, version and source of each library;
// otherwise one "* name (Version v)" line each.
func StdlibInfoVersion(details bool) []string {
	return embeddedStdlib.infoVersion(details)
}

func (s *Stdlib) infoVersion(details bool) []string {
	var out []string
	for _, name := range s.getAllFolderNames() {
		lib := s.retrieve(name)
		version := javaNullString(lib.getVersion())
		if details {
			out = append(out, "<b>"+name, "Version "+version, "Delivered by "+javaNullString(lib.getSource()), " ")
		} else {
			out = append(out, "* "+name+" (Version "+version+")")
		}
	}
	return out
}

// PrintStdLib is Stdlib.printStdLib: the detailed lines without "<b>".
func PrintStdLib() []string {
	lines := StdlibInfoVersion(true)
	for i, l := range lines {
		lines[i] = strings.ReplaceAll(l, "<b>", "")
	}
	return lines
}
