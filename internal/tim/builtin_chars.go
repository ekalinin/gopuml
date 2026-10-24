// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Backslash.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Breakline.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Dollar.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/LeftAlign.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Newline.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/NewlineShort.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Percent.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/RightAlign.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/Tabulation.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jaws/Jaws.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jaws/JawsFlags.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// More private-use characters of Jaws. JawsFlags.USE_BLOCK_E1_IN_NEWLINE_FUNCTION
// is true, so %newline, %n and %breakline return them instead of "\n".
const (
	jawsNewlineLeftAlign  = '\uE101'
	jawsNewlineRightAlign = '\uE102'
	jawsRealBackslash     = '\uE110'
	jawsRealTabulation    = '\uE111'
)

// charFunction is a builtin without argument that returns a fixed string:
// %backslash, %breakline, %dollar, %left_align, %newline, %n, %percent,
// %right_align and %tab.
type charFunction struct {
	simpleReturnFunction
	name  string
	value string
}

func (f charFunction) Signature() TFunctionSignature { return NewTFunctionSignature(f.name, 0) }

func (charFunction) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (f charFunction) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	return TValueFromString(f.value), nil
}

func newCharFunction(name string, value rune) func(*Defines) TFunction {
	return func(*Defines) TFunction { return charFunction{name: name, value: string(value)} }
}
