// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/MonomorphToJson.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

import "github.com/ekalinin/gopuml/internal/json"

// MonomorphToJSON converts a Monomorph into the JSON value of its data: a
// string, an array or an object; nil for an undetermined Monomorph. A list
// in a list, or an undetermined value in a container, gives the
// UnsupportedOperationException of upstream as an *Error.
func MonomorphToJSON(input *Monomorph) (result json.Value, err error) {
	defer catch(&err)
	switch input.Type() {
	case Scalar:
		return json.ValueOfString(input.Value()), nil
	case List:
		return convertToArray(input), nil
	case Map:
		return convertToObject(input), nil
	}
	return nil, nil
}

func convertToArray(input *Monomorph) *json.Array {
	result := json.NewArray()
	for i := range input.Size() {
		element := input.ElementAt(i)
		switch element.Type() {
		case Scalar:
			result.Add(json.ValueOfString(element.Value()))
		case Map:
			result.Add(convertToObject(element))
		default:
			throw(UnsupportedOperationException, "")
		}
	}
	return result
}

func convertToObject(input *Monomorph) *json.Object {
	result := json.NewObject()
	for _, key := range input.Keys() {
		element := input.MapValue(key)
		switch element.Type() {
		case Scalar:
			result.Add(key, json.ValueOfString(element.Value()))
		case Map:
			result.Add(key, convertToObject(element))
		case List:
			result.Add(key, convertToArray(element))
		default:
			throw(UnsupportedOperationException, "")
		}
	}
	return result
}
