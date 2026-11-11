// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/YamlSyntaxException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

// The classes of the exceptions that upstream's parser throws.
const (
	YamlSyntaxException           = "net.sourceforge.plantuml.yaml.parser.YamlSyntaxException"
	IllegalStateException         = "java.lang.IllegalStateException"
	UnsupportedOperationException = "java.lang.UnsupportedOperationException"
	IndexOutOfBoundsException     = "java.lang.IndexOutOfBoundsException"
)

// Error is an exception of upstream's parser: YAML that is not of the subset
// it reads.
type Error struct {
	Class   string
	Message string
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Class
	}
	return e.Class + ": " + e.Message
}

// throw panics with the exception, which Parse and MonomorphToJSON return.
func throw(class, message string) { panic(&Error{Class: class, Message: message}) }

// catch returns the exception of a panic of throw as *err; other panics go
// on.
func catch(err *error) {
	if v := recover(); v != nil {
		e, ok := v.(*Error)
		if !ok {
			panic(v)
		}
		*err = e
	}
}
