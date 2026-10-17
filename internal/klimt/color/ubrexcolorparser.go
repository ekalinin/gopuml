// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/UBrexColorParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package color

import "github.com/ekalinin/gopuml/internal/command/regex"

// ubrexColorRegexp translates upstream's UBrex pattern
// "# 〇+〴w 〇?〘 「-\|/」 〇+〴w 〙": '#', one or more word characters, then
// optionally one of '-', '\\', '|' or '/' followed by one or more word
// characters. Unlike colorRegexp, "#a" matches. UBrex quantifiers are greedy
// and never backtrack; RE2 may backtrack, but as the separators are not word
// characters both find the same match whenever what follows the color is not
// a word character or a separator, which holds in upstream's only user, the
// legacy activity partition command.
const ubrexColorRegexp = `#\w+(?:[-\\|/]\w+)?`

// UBrexColorParser is the color part of the commands upstream writes with its
// UBrex engine. That engine is not ported: its single color pattern is
// translated to a regex leaf.
type UBrexColorParser struct {
	regex    *regex.Leaf
	name     string
	mainType ColorType
}

// Color parses the color captured under the parser's name, applying it to
// the parser's main type. Without a capture it returns EmptyColors.
func (p *UBrexColorParser) Color(arg *regex.Result, set *HColorSet) (*Colors, error) {
	data, ok := arg.Lazy(p.name, 0)
	if !ok {
		return EmptyColors(), nil
	}
	return NewColors(data, set, p.mainType)
}

// UBrexSimpleColor returns a color captured as "COLOR".
func UBrexSimpleColor(mainType ColorType) *UBrexColorParser {
	return UBrexSimpleColorID(mainType, "COLOR")
}

// UBrexSimpleColorID returns a color captured as id. As upstream warns, the
// color is not optional.
func UBrexSimpleColorID(mainType ColorType, id string) *UBrexColorParser {
	return &UBrexColorParser{name: id, regex: regex.NewNamedLeaf(1, id, "("+ubrexColorRegexp+")"), mainType: mainType}
}

// Regex returns the leaf to put in the command pattern.
func (p *UBrexColorParser) Regex() *regex.Leaf { return p.regex }
