// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/UBrexColorParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// ubrexColorRegexp is upstream's COLOR_REGEXP: '#', word characters, and
// optionally one of '-', '\', '|' or '/' followed by word characters.
const ubrexColorRegexp = `# 〇+〴w 〇?〘 「-\|/」 〇+〴w 〙`

// UBrexColorParser is upstream's klimt.color.UBrexColorParser: the color
// part of the commands upstream writes with UBrex. It lives here, not in
// package color, because a UBrex part is a type of this package.
type UBrexColorParser struct {
	part     UBrexPart
	name     string
	mainType color.ColorType
}

// Color parses the color captured under the parser's name, applying it to
// the parser's main type. Without a capture it returns EmptyColors.
func (p *UBrexColorParser) Color(arg *regex.Result, set *color.HColorSet) (*color.Colors, error) {
	data, ok := arg.Lazy(p.name, 0)
	if !ok {
		return color.EmptyColors(), nil
	}
	return color.NewColors(data, set, p.mainType)
}

// UBrexSimpleColor returns a color captured as "COLOR".
func UBrexSimpleColor(mainType color.ColorType) *UBrexColorParser {
	return UBrexSimpleColorID(mainType, "COLOR")
}

// UBrexSimpleColorID returns a color captured as id. As upstream warns, the
// color is not optional.
func UBrexSimpleColorID(mainType color.ColorType, id string) *UBrexColorParser {
	return &UBrexColorParser{name: id, part: UBrexNamed(id, UBrexLeaf(ubrexColorRegexp)), mainType: mainType}
}

// Regex returns the part to put in the command pattern.
func (p *UBrexColorParser) Regex() UBrexPart { return p.part }
