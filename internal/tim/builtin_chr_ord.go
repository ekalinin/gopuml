// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Chr.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Ord.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"unicode/utf8"

	"github.com/ekalinin/gopuml/internal/text"
)

// chr is %chr(codePoint): the character of a Unicode code point, or the NUL
// character when the int is not a code point. A surrogate code point gives
// U+FFFD, the Go form of a lone surrogate.
type chr struct{ simpleReturnFunction }

func (chr) Signature() TFunctionSignature { return NewTFunctionSignature("%chr", 1) }

func (chr) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (chr) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	cp := values[0].ToInt()
	if cp < 0 || cp > utf8.MaxRune {
		// Character.toChars throws IllegalArgumentException
		return TValueFromString("\x00"), nil
	}
	return TValueFromString(string(rune(cp))), nil
}

// ord is %ord(s): the code point at the start of s, 0 for an empty string.
type ord struct{ simpleReturnFunction }

func (ord) Signature() TFunctionSignature { return NewTFunctionSignature("%ord", 1) }

func (ord) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (ord) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	s := values[0].String()
	if s == "" {
		// codePointAt throws StringIndexOutOfBoundsException
		return TValueFromInt(0), nil
	}
	r, _ := utf8.DecodeRuneInString(s)
	return TValueFromInt(int32(r)), nil
}
