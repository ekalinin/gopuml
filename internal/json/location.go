// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/Location.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import "strconv"

// Location is a character position in a JSON input. Offset and Column count
// UTF-16 code units, like Java string indexes.
type Location struct {
	// Offset is the absolute character index, starting at 0.
	Offset int
	// Line is the line number, starting at 1.
	Line int
	// Column is the column number, starting at 1.
	Column int
}

// String returns "line:column", for example "1:5".
func (l Location) String() string {
	return strconv.Itoa(l.Line) + ":" + strconv.Itoa(l.Column)
}
