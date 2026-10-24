// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svg/parser/ISvgSpriteParser.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svg/parser/SvgSpriteParserFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// ISvgSpriteParser is an SVG sprite: a sprite that draws the shapes of an
// SVG text. Upstream's package svg.parser joins this one (design D2): the
// parsers draw klimt shapes, and klimt sprites create parsers.
type ISvgSpriteParser interface {
	Sprite
	// DrawU draws the sprite at scale, with the default font color and the
	// optional forced color (nil for none) of monochrome sprites.
	DrawU(ug UGraphic, scale float64, fontColor, forcedColor color.HColor)
}

// SvgSpriteParserFactoryCreateDefault is upstream's
// SvgSpriteParserFactory.create(String): the parser of one SVG text without
// pragma, the Nano parser.
func SvgSpriteParserFactoryCreateDefault(svg string) ISvgSpriteParser {
	return SvgSpriteParserFactoryCreate([]string{svg}, nil, nil)
}

// SvgSpriteParserFactoryCreate is upstream's SvgSpriteParserFactory.create
// of SVG fragments, a pragma and the MD5 map of the preprocessor (both may be
// nil). The fragments are joined, each "data:image/png;md5,<key>" replaced by
// "data:image/png;base64,<data>" of the map. The pragma "svgparser" selects
// the parser: "sax" (ignoring case and surrounding spaces) for SvgSaxParser,
// SvgNanoParser otherwise. Upstream's TeaVM branch, which always uses Nano,
// does not apply.
func SvgSpriteParserFactoryCreate(svg []string, pragma *Pragma, md5map map[string]string) ISvgSpriteParser {
	parser := svgSpriteParserSelector(pragma)

	// the replacements do not overlap (base64 data has no ";"), so the order of
	// upstream's HashMap does not matter
	keys := make([]string, 0, len(md5map))
	for k := range md5map {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	for _, s := range svg {
		for _, k := range keys {
			s = strings.ReplaceAll(s, umlSourceBase64TagReplacement+k, AtomImgDataImagePngBase64+md5map[k])
		}
		sb.WriteString(s)
	}

	if parser == "sax" {
		return NewSvgSaxParser(sb.String())
	}

	// Default parser: nano unless explicitly set to sax.
	return NewSvgNanoParser(sb.String())
}

// svgSpriteParserSelector returns the value of the "svgparser" pragma,
// trimmed and in lower case, "nano" when it is not set or empty.
func svgSpriteParserSelector(pragma *Pragma) string {
	if pragma != nil && pragma.IsDefine(PragmaSvgParser) {
		value, ok := pragma.Value(PragmaSvgParser)
		if !ok {
			return "nano"
		}
		value = javaLower(javaTrim(value))
		if value == "" {
			return "nano"
		}
		return value
	}
	return "nano"
}
