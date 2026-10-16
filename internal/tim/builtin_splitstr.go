// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/SplitStr.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/SplitStrRegex.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/json"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// splitStr is %splitstr(str, separators): the tokens of a StringTokenizer,
// separated by any of the characters of separators, as a JSON array. Its
// signature has three arguments but, as upstream, it covers calls with two.
type splitStr struct{ simpleReturnFunction }

func (splitStr) Signature() TFunctionSignature { return NewTFunctionSignature("%splitstr", 3) }

func (splitStr) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (splitStr) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	result := json.NewArray()
	separator := values[1].String()
	isSeparator := func(r rune) bool { return strings.ContainsRune(separator, r) }
	for _, token := range strings.FieldsFunc(values[0].String(), isSeparator) {
		result.Add(json.ValueOfString(token))
	}
	return TValueFromJSON(result), nil
}

// splitStrRegex is %splitstr_regex(str, regex): String.split with a Java
// regular expression, as a JSON array. The regular expression runs with a
// timeout (design D4); upstream may run forever on a pathological pattern.
type splitStrRegex struct{ simpleReturnFunction }

func (splitStrRegex) Signature() TFunctionSignature {
	return NewTFunctionSignature("%splitstr_regex", 2)
}

func (splitStrRegex) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (splitStrRegex) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	re, err := compileJavaRegex(values[1].String())
	if err != nil {
		throwRuntime("PatternSyntaxException", err.Error())
	}
	parts, err := javaSplit(values[0].String(), re)
	if err != nil {
		return nil, newEaterException("Timeout in regular expression", location)
	}
	result := json.NewArray()
	for _, part := range parts {
		result.Add(json.ValueOfString(part))
	}
	return TValueFromJSON(result), nil
}
