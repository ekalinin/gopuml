// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Dec2hex.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Hex2dec.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/text"
)

// dec2hex is %dec2hex(i): Integer.toHexString, the unsigned hexadecimal
// form of the int (negative ints give eight digits). A string argument
// counts as 0.
type dec2hex struct{ simpleReturnFunction }

func (dec2hex) Signature() TFunctionSignature { return NewTFunctionSignature("%dec2hex", 1) }

func (dec2hex) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (dec2hex) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromString(strconv.FormatUint(uint64(uint32(values[0].ToInt())), 16)), nil
}

// hex2dec is %hex2dec(s): Integer.parseInt(s, 16), or 0 when s is not a
// hexadecimal int.
type hex2dec struct{ simpleReturnFunction }

func (hex2dec) Signature() TFunctionSignature { return NewTFunctionSignature("%hex2dec", 1) }

func (hex2dec) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (hex2dec) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	v, ok := javaParseIntRadix(values[0].String(), 16)
	if !ok {
		return TValueFromInt(0), nil
	}
	return TValueFromInt(v), nil
}
