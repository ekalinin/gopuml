// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetAllStdlib.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetStdlib.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// The standard library builtins read the Stdlib index of the render. The
// embedded index is empty (design D14), so by default they list nothing,
// and naming a library throws UncheckedIOException, as upstream does for a
// library that does not exist.

// getAllStdlib is %get_all_stdlib() (the library names) and
// %get_all_stdlib(any) (the name, version and source of each library).
// Upstream deprecates it in favor of %get_stdlib.
type getAllStdlib struct{ simpleReturnFunction }

func (getAllStdlib) Signature() TFunctionSignature {
	return NewTFunctionSignature("%get_all_stdlib", 1)
}

func (getAllStdlib) CanCover(nbArg int, _ []string) bool { return nbArg == 0 || nbArg == 1 }

func (getAllStdlib) ExecuteReturnFunction(context *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	switch len(values) {
	case 0:
		result := json.NewArray()
		for _, name := range context.stdlib.getAllFolderNames() {
			result.Add(json.ValueOfString(name))
		}
		return TValueFromJSON(result), nil
	case 1:
		res := json.NewObject()
		for _, name := range context.stdlib.getAllFolderNames() {
			folder := context.stdlib.retrieve(name)
			version, okVersion := folder.getVersion()
			source, okSource := folder.getSource()
			object := json.NewObject().
				Add("name", json.ValueOfString(name)).
				Add("version", stringOrNull(version, okVersion)).
				Add("source", stringOrNull(source, okSource))
			res.Add(name, object)
		}
		return TValueFromJSON(res), nil
	}
	return nil, newEaterException("Error on get_all_stdlib: Too many arguments", location)
}

// stringOrNull is Json.value(String): JSON null for a Java null.
func stringOrNull(s string, ok bool) json.Value {
	if !ok {
		return json.Null
	}
	return json.ValueOfString(s)
}

// getStdlib is %get_stdlib() (the metadata of every library),
// %get_stdlib(name) (the metadata of a library) and %get_stdlib(name, key)
// (one metadata value, "" when missing). Metadata keys are in lower case.
type getStdlib struct{ simpleReturnFunction }

func (getStdlib) Signature() TFunctionSignature { return NewTFunctionSignature("%get_stdlib", 1) }

func (getStdlib) CanCover(nbArg int, _ []string) bool { return nbArg == 0 || nbArg == 1 || nbArg == 2 }

func (getStdlib) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	result := json.NewObject()
	switch len(values) {
	case 0:
		for _, folderName := range context.stdlib.getAllFolderNames() {
			folder := context.stdlib.retrieve(folderName)
			metadata := json.NewObject()
			for _, ent := range folder.info {
				metadata.Add(javaToLowerCase(ent.key), json.ValueOfString(ent.value))
			}
			result.Add(folderName, metadata)
		}
	case 1:
		folder := context.stdlib.retrieve(values[0].String())
		for _, ent := range folder.info {
			result.Add(javaToLowerCase(ent.key), json.ValueOfString(ent.value))
		}
	case 2:
		folder := context.stdlib.retrieve(values[0].String())
		key := javaToLowerCase(values[1].String())
		value, ok := folder.getMetadata(key)
		if !ok {
			// Temporary, we check the uppercase key if needed
			value, ok = folder.getMetadata(javaToUpperCase(key))
		}
		if !ok {
			value = ""
		}
		return TValueFromString(value), nil
	}
	return TValueFromJSON(result), nil
}
