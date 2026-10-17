// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetJsonKey.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetJsonType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/JsonAdd.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/JsonKeyExists.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/JsonMerge.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/JsonRemove.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/JsonSet.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Str2Json.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// The JSON builtins never modify their argument: they work on a deep copy.

// getJSONKey is %get_json_keys(data): the member names of an object, or of
// the objects of an array, as a JSON array.
type getJSONKey struct{ simpleReturnFunction }

func (getJSONKey) Signature() TFunctionSignature { return NewTFunctionSignature("%get_json_keys", 1) }

func (getJSONKey) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (getJSONKey) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data := values[0]
	if !data.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	j := data.ToJSON()
	if j.IsObject() {
		result := json.NewArray()
		for _, key := range j.AsObject().Names() {
			result.Add(json.ValueOfString(key))
		}
		return TValueFromJSON(result), nil
	}
	if j.IsArray() {
		result := json.NewArray()
		for _, tmp := range j.AsArray().All() {
			if tmp.IsObject() {
				for _, key := range tmp.AsObject().Names() {
					result.Add(json.ValueOfString(key))
				}
			}
		}
		return TValueFromJSON(result), nil
	}
	return nil, newEaterException("Bad JSON type", location)
}

// getJSONType is %get_json_type(data): "string", "number", "array",
// "object", "boolean" or "json" (null).
type getJSONType struct{ simpleReturnFunction }

func (getJSONType) Signature() TFunctionSignature { return NewTFunctionSignature("%get_json_type", 1) }

func (getJSONType) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (getJSONType) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data := values[0]
	switch {
	case data.IsString():
		return TValueFromString("string"), nil
	case data.IsNumber():
		return TValueFromString("number"), nil
	case !data.IsJSON():
		return TValueFromString("not_json"), nil
	}
	j := data.ToJSON()
	switch {
	case j.IsArray():
		return TValueFromString("array"), nil
	case j.IsObject():
		return TValueFromString("object"), nil
	case j.IsBoolean():
		return TValueFromString("boolean"), nil
	case j.IsNumber():
		return TValueFromString("number"), nil
	case j.IsString():
		return TValueFromString("string"), nil
	}
	return TValueFromString("json"), nil
}

// jsonAdd is %json_add(array, value) and %json_add(object, name, value):
// a copy with the value appended.
type jsonAdd struct{ simpleReturnFunction }

func (jsonAdd) Signature() TFunctionSignature { return NewTFunctionSignature("%json_add", 3) }

func (jsonAdd) CanCover(nbArg int, _ []string) bool { return nbArg == 2 || nbArg == 3 }

func (jsonAdd) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data := values[0]
	if !data.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	j := data.ToJSON().CloneMe()
	if !j.IsArray() && !j.IsObject() {
		return data, nil
	}
	if j.IsArray() {
		j.AsArray().Add(values[1].ToJSONValue())
		return TValueFromJSON(j), nil
	}
	name := values[1].String()
	j.AsObject().Add(name, values[2].ToJSONValue())
	return TValueFromJSON(j), nil
}

// jsonKeyExists is %json_key_exists(object, name). Its signature has one
// argument but, as upstream, it covers calls with two: a call with one
// argument in plain text reaches it and fails on an object.
type jsonKeyExists struct{ simpleReturnFunction }

func (jsonKeyExists) Signature() TFunctionSignature {
	return NewTFunctionSignature("%json_key_exists", 1)
}

func (jsonKeyExists) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (jsonKeyExists) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	arg0 := values[0]
	if !arg0.IsJSON() {
		return TValueFromBoolean(false), nil
	}
	j := arg0.ToJSON()
	if !j.IsObject() {
		return TValueFromBoolean(false), nil
	}
	arg1 := values[1]
	if arg1.IsString() || (arg1.IsJSON() && arg1.ToJSON().IsString()) {
		if j.AsObject().Contains(arg1.String()) {
			return TValueFromBoolean(true), nil
		}
	}
	return TValueFromBoolean(false), nil
}

// jsonMerge is %json_merge(a, b): the elements of two arrays, or the
// members of two objects (those of b replacing those of a).
type jsonMerge struct{ simpleReturnFunction }

func (jsonMerge) Signature() TFunctionSignature { return NewTFunctionSignature("%json_merge", 2) }

func (jsonMerge) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (jsonMerge) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data0 := values[0]
	if !data0.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	data1 := values[1]
	if !data1.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	json0 := data0.ToJSON().CloneMe()
	json1 := data1.ToJSON()
	if (!json0.IsArray() && !json0.IsObject() && !json1.IsArray() && !json1.IsObject()) ||
		(json0.IsArray() && json1.IsObject()) || (json0.IsObject() && json1.IsArray()) {
		return data0, nil
	}
	if json0.IsArray() && json1.IsArray() {
		for _, j1 := range json1.AsArray().Values() {
			json0.AsArray().Add(j1)
		}
		return TValueFromJSON(json0), nil
	}
	if json0.IsObject() && json1.IsObject() {
		json0.AsObject().Merge(json1.AsObject())
		return TValueFromJSON(json0), nil
	}
	return nil, newEaterException("Bad JSON type", location)
}

// jsonRemove is %json_remove(array, index) and %json_remove(object, name).
type jsonRemove struct{ simpleReturnFunction }

func (jsonRemove) Signature() TFunctionSignature { return NewTFunctionSignature("%json_remove", 2) }

func (jsonRemove) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (jsonRemove) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data := values[0]
	if !data.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	j := data.ToJSON().CloneMe()
	if !j.IsArray() && !j.IsObject() {
		return data, nil
	}
	if j.IsArray() {
		if values[1].IsNumber() {
			index := int(values[1].ToInt())
			if 0 <= index && index < j.AsArray().Size() {
				_ = j.AsArray().Remove(index)
			}
		}
		return TValueFromJSON(j), nil
	}
	j.AsObject().Remove(values[1].String())
	return TValueFromJSON(j), nil
}

// jsonSet is %json_set(object, object2) (deep merge),
// %json_set(array, index, value) and %json_set(object, name, value).
type jsonSet struct{ simpleReturnFunction }

func (jsonSet) Signature() TFunctionSignature { return NewTFunctionSignature("%json_set", 3) }

func (jsonSet) CanCover(nbArg int, _ []string) bool { return nbArg == 2 || nbArg == 3 }

func (jsonSet) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	data := values[0]
	if !data.IsJSON() {
		return nil, newEaterException("Not JSON data", location)
	}
	j := data.ToJSON().CloneMe()
	if !j.IsArray() && !j.IsObject() {
		return data, nil
	}
	// The cases of upstream's switch, with the fall through from 2 to 3.
	if len(values) == 2 && j.IsObject() {
		value := values[1].ToJSONValue()
		if value.IsObject() {
			j.AsObject().DeepMerge(value.AsObject())
		}
		return TValueFromJSON(j), nil
	}
	if len(values) == 2 || len(values) == 3 {
		if j.IsArray() {
			if values[1].IsNumber() {
				index := int(values[1].ToInt())
				value := values[2].ToJSONValue()
				if 0 <= index && index < j.AsArray().Size() {
					_ = j.AsArray().Set(index, value)
				}
			}
			return TValueFromJSON(j), nil
		}
		if j.IsObject() {
			j.AsObject().Set(values[1].String(), values[2].ToJSONValue())
			return TValueFromJSON(j), nil
		}
	}
	return nil, newEaterException("Error on json_set: Too many arguments", location)
}

// str2JSON is %str2json(s): the JSON value of s, or "" when s is not JSON.
type str2JSON struct{ simpleReturnFunction }

func (str2JSON) Signature() TFunctionSignature { return NewTFunctionSignature("%str2json", 1) }

func (str2JSON) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (str2JSON) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	j, err := json.Parse(values[0].String())
	if err != nil {
		return TValueFromString(""), nil
	}
	return TValueFromJSON(j), nil
}
