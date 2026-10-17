// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/ColorParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "github.com/ekalinin/gopuml/internal/command/regex"

// colorRegexp is a single color: '#', a word, then optionally a gradient
// separator and a second word. As upstream, it needs two word characters at
// least, so "#a" is not a color here.
const colorRegexp = `#\w+[-\\|/]?\w+`

// part2 is a color specification such as "#red;line:blue;line.dashed".
// Upstream ends each item with (?:;|(?![\w;:.])), a negative lookahead that
// RE2 does not have, so the pattern is rewritten by hand:
//   - The character before the lookahead is always a word character (the end
//     of a keyword or of a color), so (?!\w) is the word boundary \b, and the
//     terminator is written (?:;|\b).
//   - The lookahead also refuses '.', which makes the keyword "line" fail
//     before ".dashed" so that "line\.dashed" is tried. The keywords that
//     start with "line." are moved before "line": upstream can never match
//     both at the same place, so the order does not change its results, and
//     RE2 now finds "line.dashed" first whatever follows the color.
//   - What cannot be expressed is the refusal of ':', '.' or an unconsumed ';'
//     right after the last item. A match may end there where upstream
//     backtracks, usually to the single color alternative. This only changes
//     the result when the rest of the command pattern accepts that character
//     at that place, for example "#back:red:caption" before an optional
//     ":caption": upstream captures "#back" and Go "#back:red".
//
// Inside the repetition nothing changes, as the next item must start with a
// letter.
const part2 = `#(?:\w+[-\\|/]?\w+;)?(?:(?:text|back|header|line\.dashed|line\.dotted|line\.bold|line|shadowing)(?::\w+[-\\|/]?\w+)?(?:;|\b))+`

const colorsRegexp = "(?:" + part2 + ")|(?:" + colorRegexp + ")"

// ColorParser is the color part of command patterns: a regex leaf capturing
// the colors under a name, and the parsing of what it captured.
type ColorParser struct {
	regex    *regex.Leaf
	name     string
	mainType ColorType
}

// Color parses the colors captured under the parser's name, applying plain
// colors to the parser's main type. Without a capture it returns
// EmptyColors.
func (p *ColorParser) Color(arg *regex.Result, set *HColorSet) (*Colors, error) {
	data, ok := arg.Lazy(p.name, 0)
	if !ok {
		return EmptyColors(), nil
	}
	return NewColors(data, set, p.mainType)
}

// SimpleColor returns an optional color specification captured as "COLOR".
func SimpleColor(mainType ColorType) *ColorParser { return SimpleColorID(mainType, "COLOR") }

// SimpleColorID returns an optional color specification captured as id.
func SimpleColorID(mainType ColorType, id string) *ColorParser {
	return &ColorParser{name: id, regex: regex.NewNamedLeaf(1, id, "("+colorsRegexp+")?"), mainType: mainType}
}

// MandatoryColor returns a color specification captured as "COLOR".
func MandatoryColor(mainType ColorType) *ColorParser {
	return &ColorParser{name: "COLOR", regex: regex.NewNamedLeaf(1, "COLOR", "("+colorsRegexp+")"),
		mainType: mainType}
}

// SimpleColorPrefix returns an optional color specification preceded by
// optPrefix and a space, captured as "COLOR".
func SimpleColorPrefix(optPrefix string, mainType ColorType) *ColorParser {
	return &ColorParser{name: "COLOR",
		regex: regex.NewNamedLeaf(1, "COLOR", "(?:"+optPrefix+" ("+colorsRegexp+"))?"), mainType: mainType}
}

// Exp1 is the leaf of SimpleColor, for commands that parse the colors
// themselves.
func Exp1() *regex.Leaf { return regex.NewNamedLeaf(1, "COLOR", "("+colorsRegexp+")?") }

// Exp2 is an optional single color captured as "BACKCOLOR".
func Exp2() *regex.Leaf { return regex.NewNamedLeaf(1, "BACKCOLOR", "("+colorRegexp+")?") }

// Exp3 is an optional single color captured as "BACKCOLOR2".
func Exp3() *regex.Leaf { return regex.NewNamedLeaf(1, "BACKCOLOR2", "("+colorRegexp+")?") }

// Exp4 is an optional single color followed by ':', captured as "COLOR"
// without the ':'.
func Exp4() *regex.Leaf { return regex.NewNamedLeaf(1, "COLOR", "(?:("+colorRegexp+"):)?") }

// Exp6 is an optional single color followed by '|', captured as "COLOR"
// without the '|'.
func Exp6() *regex.Leaf { return regex.NewNamedLeaf(1, "COLOR", "(?:("+colorRegexp+")\\|)?") }

// Exp7 is an optional single color captured as "COLOR".
func Exp7() *regex.Leaf { return regex.NewNamedLeaf(1, "COLOR", "(?:("+colorRegexp+"))?") }

// Regex returns the leaf to put in the command pattern.
func (p *ColorParser) Regex() *regex.Leaf { return p.regex }
