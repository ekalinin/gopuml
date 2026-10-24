// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/Parser.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/CreoleMode.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"unicode/utf16"
)

// CreoleMode is how much creole markup a text accepts.
type CreoleMode int

// Creole modes, in upstream's order.
const (
	CreoleModeFull CreoleMode = iota
	CreoleModeSimpleLine
	CreoleModeNoCreole
	CreoleModeFullButUnderscore
)

// String returns upstream's constant name.
func (m CreoleMode) String() string {
	return enumName("CreoleMode", []string{"FULL", "SIMPLE_LINE", "NO_CREOLE", "FULL_BUT_UNDERSCORE"}, int(m))
}

// ParserMonospaced is the family name of the monospaced font in creole.
const ParserMonospaced = "monospaced"

// ParserIsLatexStart reports whether line opens a LaTeX block.
func ParserIsLatexStart(line string) bool { return line == "<latex>" }

// ParserIsLatexEnd reports whether line closes a LaTeX block.
func ParserIsLatexEnd(line string) bool { return line == "</latex>" }

// ParserIsCodeStart reports whether line opens a code block.
func ParserIsCodeStart(line string) bool { return line == "<code>" }

// ParserIsCodeEnd reports whether line closes a code block.
func ParserIsCodeEnd(line string) bool { return line == "</code>" }

// ParserIsTreeStart reports whether line starts a tree ("|_").
func ParserIsTreeStart(line string) bool {
	u := utf16.Encode([]rune(line))
	if len(u) < 2 {
		return false
	}
	return u[0] == '|' && u[1] == '_'
}

var (
	parserScale = regexp.MustCompile(`(?:scale=|\*)([0-9.]+)`)
	parserColor = regexp.MustCompile(`color[= :](#[0-9a-fA-F]{1,6}|\w+)`)
)

// ParserScale returns the scale given by "scale=x" or "*x" in s, or def
// (also when s is missing, ok false). A malformed number such as "1.2.3"
// panics, as upstream throws NumberFormatException.
func ParserScale(s string, ok bool, def float64) float64 {
	if !ok {
		return def
	}
	if m := parserScale.FindStringSubmatch(s); m != nil {
		return javaParseDouble(m[1])
	}
	return def
}

// ParserColor returns the color given by "color=x" in s; ok is false when
// there is none.
func ParserColor(s string) (c string, ok bool) {
	if m := parserColor.FindStringSubmatch(s); m != nil {
		return m[1], true
	}
	return "", false
}
