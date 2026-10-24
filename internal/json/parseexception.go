// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/ParseException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

// ParseError is returned when the input is not valid JSON, like upstream's
// ParseException.
type ParseError struct {
	message string
	// Location is the position of the error in the input.
	Location Location
}

func newParseError(message string, location Location) *ParseError {
	return &ParseError{message: message, Location: location}
}

// Error returns upstream's message followed by the location, for example
// "Expected ':' at 1:6".
func (e *ParseError) Error() string {
	return e.message + " at " + e.Location.String()
}
