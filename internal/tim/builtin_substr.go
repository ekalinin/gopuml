// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Substr.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/text"
)

// substr is %substr(full, pos[, len]), with positions and lengths in UTF-16
// chars. As in Java, a negative position or length throws
// StringIndexOutOfBoundsException, which becomes "Fatal parsing error".
type substr struct{ simpleReturnFunction }

func (substr) Signature() TFunctionSignature { return NewTFunctionSignature("%substr", 3) }

func (substr) CanCover(nbArg int, _ []string) bool { return nbArg == 2 || nbArg == 3 }

func (substr) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	full := toUTF16(values[0].String())
	pos := int(values[1].ToInt())
	if pos >= len(full) {
		return TValueFromString(""), nil
	}
	if pos < 0 {
		throwRuntime("StringIndexOutOfBoundsException", "begin "+strconv.Itoa(pos)+", end "+strconv.Itoa(len(full))+
			", length "+strconv.Itoa(len(full)))
	}
	result := full[pos:]
	if len(values) == 3 {
		length := int(values[2].ToInt())
		if length < len(result) {
			if length < 0 {
				throwRuntime("StringIndexOutOfBoundsException", "begin 0, end "+strconv.Itoa(length)+
					", length "+strconv.Itoa(len(result)))
			}
			result = result[:length]
		}
	}
	return TValueFromString(fromUTF16(result)), nil
}
