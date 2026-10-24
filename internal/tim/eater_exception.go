// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// EaterException is a preprocessing error: the message and the line it is
// reported on. Its Location.Location.Position is the 0-based line number in
// the file that holds the line.
type EaterException struct {
	Message  string
	Location text.StringLocated
}

func newEaterException(message string, location text.StringLocated) *EaterException {
	return &EaterException{Message: message, Location: location}
}

func (e *EaterException) Error() string { return e.Message }

// Java separates checked exceptions (EaterException, returned as errors
// here) from runtime exceptions (NullPointerException, ArithmeticException,
// NumberFormatException...), which upstream catches only in a few places,
// mainly TContext.executeOneLineSafe, which turns them into "Fatal parsing
// error". Runtime exceptions are Go panics: either a javaRuntimeException
// raised explicitly, or a Go runtime error (index out of range, nil
// dereference) at the same place where Java would throw.

// javaRuntimeException is the panic value of a Java runtime exception.
type javaRuntimeException struct {
	Name    string
	Message string
}

func (e *javaRuntimeException) Error() string {
	if e.Message == "" {
		return "java.lang." + e.Name
	}
	return "java.lang." + e.Name + ": " + e.Message
}

// throwRuntime panics with a Java runtime exception.
func throwRuntime(name, message string) {
	panic(&javaRuntimeException{Name: name, Message: message})
}
