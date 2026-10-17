// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Dirpath.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Filedate.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Filename.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/FilenameNoExtension.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// environmentFunction is a builtin without argument that returns a value of
// the environment of the source file, read from the Defines when the
// function is created, or "" when it is not set: %dirpath, %filedate,
// %filename and %filename_no_extension.
type environmentFunction struct {
	simpleReturnFunction
	name  string
	value string
}

func (f environmentFunction) Signature() TFunctionSignature { return NewTFunctionSignature(f.name, 0) }

func (environmentFunction) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (f environmentFunction) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	return TValueFromString(f.value), nil
}

// newEnvironmentFunction creates the builtin name that returns the
// environment value key.
func newEnvironmentFunction(name, key string) func(*Defines) TFunction {
	return func(defines *Defines) TFunction {
		value, _ := defines.EnvironmentValue(key)
		return environmentFunction{name: name, value: value}
	}
}
