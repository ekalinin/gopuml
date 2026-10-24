// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// svgsprite.golden is written by testdata/javagen/SpriteGolden.java from
// svgsprite_inputs.txt.

// spriteConfig is a configuration of the harness's drawings.
type spriteConfig struct {
	name                   string
	fontColor, forcedColor color.HColor
	scale                  float64
	backColor              color.HColor
	svg                    bool
}

var spriteConfigs = []spriteConfig{
	{"s1", nil, nil, 1, nil, false},
	{"s2", nil, nil, 2, nil, false},
	{"font", spriteColor("#3366CC"), nil, 1, nil, false},
	{"forced", spriteColor("#000000"), spriteColor("#FF8000"), 1.5, nil, false},
	{"gray", nil, spriteColor("#808080"), 2, nil, false},
	{"back", spriteColor("red"), nil, 1, spriteColor("#FFEEDD"), false},
	{"svg", nil, nil, 1, nil, true},
}

func spriteParserName(p ISvgSpriteParser) string {
	switch p.(type) {
	case *SvgNanoParser:
		return "SvgNanoParser"
	case *SvgSaxParser:
		return "SvgSaxParser"
	}
	return "?"
}

// The 2x2 PNGs of SpriteGolden.java.
const (
	spriteFourPng    = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAFElEQVR42mP4z8DwHwyBNBAwNAAAR0sIecMlh+sAAAAASUVORK5CYII="
	spriteUniformPng = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEUlEQVR42mMQUDD4D8IMMAYAK7QFfSBd8FwAAAAASUVORK5CYII="
)

func spriteDecodePng(t *testing.T, b64 string) PortableImage {
	t.Helper()
	data, err := base64Decode(b64)
	if err != nil {
		t.Fatal(err)
	}
	img, err := readRasterImage(data)
	if err != nil || img == nil {
		t.Fatal("cannot read the PNG", err)
	}
	return img
}

func TestGoldenSvgSprite(t *testing.T) {
	var out []string

	// The selection of the factory
	svg := `<svg viewBox="0 0 1 1"></svg>`
	out = append(out, "factory\tdefault\t"+spriteParserName(SvgSpriteParserFactoryCreateDefault(svg)))
	out = append(out, "factory\tnull\t"+spriteParserName(SvgSpriteParserFactoryCreate([]string{svg}, nil, nil)))
	for _, v := range []string{"<empty>", "sax", " SAX ", "Sax", "nano", "", "  ", "foo", "<novalue>"} {
		pragma := PragmaCreateEmpty()
		switch v {
		case "<novalue>":
			pragma.Define("svgparser", "", false)
		case "<empty>":
		default:
			pragma.Define("svgparser", v, true)
		}
		out = append(out, "factory\t"+v+"\t"+spriteParserName(SvgSpriteParserFactoryCreate([]string{svg}, pragma, nil)))
	}
	sax := PragmaCreateEmpty()
	sax.Define("svgparser", "sax", true)
	for _, m := range []map[string]string{{"abc": "QUJD", "def": "REVG"}, {}, nil} {
		parser := SvgSpriteParserFactoryCreate([]string{`<svg><image href="data:image/png;md5,abc"/>`,
			`<image href="data:image/png;md5,def"/><image href="data:image/png;md5,xyz"/></svg>`}, sax, m)
		size := "null"
		if m != nil {
			size = itoa(len(m))
		}
		out = append(out, "md5\t"+size+"\t"+esc(parser.(*SvgSaxParser).svg))
	}

	// The drawings
	for _, line := range spriteInputs(t, "svgsprite_inputs.txt") {
		name, raw, _ := strings.Cut(line, "\t")
		text := spriteUnesc(raw)
		for _, kind := range []string{"nano", "sax"} {
			for _, cfg := range spriteConfigs {
				spriteDrawing(&out, name+"\t"+kind+"\t"+cfg.name, cfg.svg, color.IdentityMapper, func(ug *spriteRec) {
					var parser ISvgSpriteParser
					if kind == "nano" {
						parser = NewSvgNanoParser(text)
					} else {
						parser = NewSvgSaxParser(text)
					}
					tb := parser.AsTextBlock(cfg.fontColor, cfg.forcedColor, cfg.scale, cfg.backColor)
					d := tb.CalculateDimension(fakeBounder{})
					*ug.log = append(*ug.log, "dim "+d2s(d.Width())+","+d2s(d.Height()))
					tb.DrawU(ug)
				})
			}
		}
	}

	// SpriteSvgNanoParser
	four := spriteDecodePng(t, spriteFourPng)
	uniform := spriteDecodePng(t, spriteUniformPng)
	spriteImage := func(name string, img PortableImage, forced color.HColor, scale float64, mapper *color.ColorMapper) {
		spriteDrawing(&out, "spriteimage\t"+name, false, mapper, func(ug *spriteRec) {
			tb := NewSpriteSvgNanoParser(img).AsTextBlock(nil, forced, scale, spriteColor("#FFEEDD"))
			d := tb.CalculateDimension(fakeBounder{})
			*ug.log = append(*ug.log, "dim "+d2s(d.Width())+","+d2s(d.Height()))
			tb.DrawU(ug)
		})
	}
	spriteImage("four-s1", four, nil, 1, color.IdentityMapper)
	spriteImage("four-forced", four, spriteColor("#FF8000"), 1, color.IdentityMapper)
	spriteImage("four-mono", four, spriteColor("#FF8000"), 1, color.MonochromeMapper)
	spriteImage("uniform-s2", uniform, nil, 2, color.IdentityMapper)
	if e := spriteTry(func() { NewSpriteSvgNanoParser(nil) }); e != "" {
		out = append(out, "spriteimage\tnull\t"+e)
	} else {
		out = append(out, "spriteimage\tnull\tok")
	}

	compareGoldenLines(t, "svgsprite.golden", withoutSvgSpriteDifferences(out),
		withoutSvgSpriteDifferences(readLines(t, "svgsprite.golden")))
}

// svgSpriteDifferences are the drawings where this port differs from
// upstream (design D16), by the prefix of their golden lines. The lines are
// skipped, and TestSvgSpriteDifferences checks the behavior.
var svgSpriteDifferences = map[string]string{
	"image-unknown\tsax\t": "upstream throws NullPointerException for image data in an unknown format, " +
		"which fails the sprite; the image is skipped, as data that does not decode",
}

func withoutSvgSpriteDifferences(lines []string) []string {
	return slices.DeleteFunc(lines, func(l string) bool {
		for prefix := range svgSpriteDifferences {
			if strings.HasPrefix(l, prefix) {
				return true
			}
		}
		return false
	})
}

// TestSvgSpriteDifferences checks the cases of svgSpriteDifferences: the
// SAX parser skips an image in an unknown format and draws the shapes
// around it.
func TestSvgSpriteDifferences(t *testing.T) {
	svg := `<svg viewBox="0 0 10 10"><rect width="1" height="1"/><image href="data:image/png;base64,AAAA"/>` +
		`<rect width="2" height="2"/></svg>`
	var out []string
	spriteDrawing(&out, "image-unknown", false, color.IdentityMapper, func(ug *spriteRec) {
		NewSvgSaxParser(svg).AsTextBlock(nil, nil, 1, nil).DrawU(ug)
	})
	paths := 0
	for _, l := range out {
		if strings.Contains(l, "ERR") {
			t.Errorf("error drawn: %s", l)
		}
		if strings.Contains(l, " path ") {
			paths++
		}
	}
	if paths != 2 {
		t.Errorf("%d paths, want the two rectangles: %q", paths, out)
	}
}
