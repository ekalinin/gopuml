// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/VariableManager.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// variableManager replaces the variables of a line by their values.
type variableManager struct {
	memory   TMemory
	context  *TContext
	location text.StringLocated
}

// replaceVariables appends the value of the variable at str[i:] to result
// and returns the index of the last char consumed.
func (vm variableManager) replaceVariables(str []uint16, i int, result *[]uint16) (int, error) {
	presentVariable := vm.getVarnameAt(str, i)
	if n := len(*result); n >= 2 && (*result)[n-1] == '#' && (*result)[n-2] == '#' {
		*result = (*result)[:n-2]
	}
	value := vm.memory.GetVariable(fromUTF16(presentVariable))
	i += len(presentVariable) - 1
	if value.IsJSON() {
		j := value.ToJSON()
		switch {
		case j.IsString():
			*result = append(*result, toUTF16(j.AsString())...)
		case j.IsNumber():
			*result = append(*result, toUTF16(j.String())...)
		default:
			if !j.IsArray() && !j.IsObject() {
				throwRuntime("ClassCastException", "")
			}
			i++
			next, err := vm.replaceJSON(j, str, i, result)
			if err != nil {
				return 0, err
			}
			i = next - 1
		}
	} else {
		*result = append(*result, toUTF16(value.String())...)
	}
	if i+2 < len(str) && str[i+1] == '#' && str[i+2] == '#' {
		i += 2
	}
	return i, nil
}

func (vm variableManager) replaceJSON(jsonValue json.Value, str []uint16, i int, result *[]uint16) (int, error) {
	for i < len(str) {
		n := str[i]
		if n == '.' {
			i++
			var fieldName []uint16
			for i < len(str) {
				if !isJavaIdentifierPart(str[i]) {
					break
				}
				fieldName = append(fieldName, str[i])
				i++
			}
			obj, ok := jsonValue.(*json.Object)
			if !ok {
				throwRuntime("ClassCastException", "")
			}
			jsonValue = obj.Get(fromUTF16(fieldName))
		} else if n == '[' {
			i++
			var inBracket []uint16
			level := 0
			for {
				if i >= len(str) {
					throwRuntime("StringIndexOutOfBoundsException", "")
				}
				if str[i] == '[' {
					level++
				}
				if str[i] == ']' {
					if level == 0 {
						break
					}
					level--
				}
				inBracket = append(inBracket, str[i])
				i++
			}
			nbString, err := vm.context.applyFunctionsAndVariables(vm.memory,
				text.NewStringLocated(fromUTF16(inBracket), vm.location.Location))
			if err != nil {
				return 0, err
			}
			if nbString == nil {
				throwRuntime("NullPointerException", "")
			}
			switch v := jsonValue.(type) {
			case *json.Array:
				nb, ok := javaParseInt(*nbString)
				if !ok {
					throwRuntime("NumberFormatException", `For input string: "`+*nbString+`"`)
				}
				value, err := v.Get(int(nb))
				if err != nil {
					panic(err)
				}
				jsonValue = value
			case *json.Object:
				jsonValue = v.Get(*nbString)
			default:
				return 0, newEaterException("Major parsing error", vm.location)
			}
			if jsonValue == nil {
				return 0, newEaterException("Data parsing error", vm.location)
			}
			i++
		} else {
			break
		}
	}
	if jsonValue != nil {
		if jsonValue.IsString() {
			*result = append(*result, toUTF16(jsonValue.AsString())...)
		} else {
			*result = append(*result, toUTF16(jsonValue.String())...)
		}
	}
	return i, nil
}

// getVarnameAt returns the name of the variable at s[pos:], or nil.
func (vm variableManager) getVarnameAt(s []uint16, pos int) []uint16 {
	justAfterALetter := pos > 0 && isLetterOrEmojiOrUnderscoreOrDigit(s[pos-1]) && !justAfterBackslashN(s, pos)
	if justAfterALetter && s[pos] != '$' {
		return nil
	}
	n := vm.memory.variablesNames3().getLonguestMatchStartingIn(s, pos)
	if n == 0 {
		return nil
	}
	if pos+n == len(s) || !isLetterOrEmojiOrUnderscoreOrDigit(s[pos+n]) {
		return s[pos : pos+n]
	}
	return nil
}

func justAfterBackslashN(s []uint16, pos int) bool {
	return pos > 1 && s[pos-2] == '\\' && s[pos-1] == 'n'
}
