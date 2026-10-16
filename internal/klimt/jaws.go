// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/jaws/Jaws.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/jaws/JawsFlags.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/text/BackSlash.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// The private-use characters that upstream's preprocessor (Jaws) puts in
// the text for newlines, tabulations and quotes. Upstream's Jaws line
// rewriting belongs to the preprocessor.
const (
	JawsBlockE1Newline           = ''
	JawsBlockE1NewlineLeftAlign  = ''
	JawsBlockE1NewlineRightAlign = ''
	JawsBlockE1Breakline         = ''
	JawsBlockE1RealBackslash     = ''
	JawsBlockE1RealTabulation    = ''
	JawsBlockE1InvisibleQuote    = ''
	JawsBlockE1StartTextblock    = ''
	JawsBlockE1EndTextblock      = ''
)

// The Jaws feature flags of upstream's JawsFlags.
const (
	jawsUseBlockE1InNewlineFunction = true
	jawsParseNewMultilineTripleMark = false
	jawsSpecialNewlineInDisplay     = false
)

// backSlashNewline is upstream's BackSlash.NEWLINE.
const backSlashNewline = "\n"

// hiddenNewLine is upstream's BackSlash.hiddenNewLine().
func hiddenNewLine() rune { return JawsBlockE1Newline }
