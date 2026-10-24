// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TVariableScope.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "strings"

// TVariableScope is the scope modifier of an assignment. The zero value is
// Java's null: no modifier.
type TVariableScope int

// Scopes.
const (
	scopeNone TVariableScope = iota
	scopeLocal
	scopeGlobal
)

// lazzyParse is TVariableScope.lazzyParse.
func lazzyParse(value string) TVariableScope {
	switch {
	case strings.EqualFold(value, "local") && isASCII(value):
		return scopeLocal
	case strings.EqualFold(value, "global") && isASCII(value):
		return scopeGlobal
	}
	return scopeNone
}

// isASCII reports whether s holds only ASCII characters. Java's
// equalsIgnoreCase folds a few non-ASCII letters to ASCII ones differently
// from Go's EqualFold (the Kelvin sign), so only ASCII words are compared.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
