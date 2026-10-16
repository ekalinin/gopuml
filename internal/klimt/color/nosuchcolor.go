// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/NoSuchColorException.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/NoSuchColorRuntimeException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

// NoSuchColorError is upstream's NoSuchColorException: a string is not a
// color.
type NoSuchColorError struct{}

func (*NoSuchColorError) Error() string { return "NoSuchColorException" }

// NoSuchColorRuntimeError is upstream's NoSuchColorRuntimeException, the
// panic value of creole color changes with an invalid color.
type NoSuchColorRuntimeError struct{}

func (*NoSuchColorRuntimeError) Error() string {
	return "net.sourceforge.plantuml.klimt.color.NoSuchColorRuntimeException"
}
