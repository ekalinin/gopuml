// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandFactorySprite.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteBase64.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteMd5.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteSvgMultiline.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteStdlib.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteStdlibSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSpriteFile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"archive/zip"
	"bytes"
	"image"
	_ "image/gif"  // sprite files
	_ "image/jpeg" // sprite files
	_ "image/png"  // sprite files and base64 sprites
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/stdlib"
	"github.com/ekalinin/gopuml/internal/text"
)

// executeSprite is CommandFactorySprite.executeInternal.
func executeSprite(b *Base, line0 *regex.Result, lines []string) command.Result {
	var sprite *SpriteDef
	if _, ok := line0.Get("DIM", 0); !ok {
		sprite = buildGraySprite(16, -1, -1, lines)
	} else {
		width := JavaParseInt(line0.Str("DIM", 0))
		height := JavaParseInt(line0.Str("DIM", 1))
		if _, isColor := line0.Get("DIM", 4); !isColor {
			nbLevel := JavaParseInt(line0.Str("DIM", 2))
			if nbLevel != 4 && nbLevel != 8 && nbLevel != 16 {
				return command.Error("Only 4, 8 or 16 graylevel are allowed.")
			}
			if _, z := line0.Get("DIM", 3); !z {
				sprite = buildGraySprite(nbLevel, width, height, lines)
			} else {
				var sb strings.Builder
				for _, s := range lines {
					sb.WriteString(text.Trin(s))
				}
				sprite = buildGraySpriteZ(nbLevel, width, height, sb.String())
				if sprite == nil {
					return command.Error("Cannot decode sprite.")
				}
			}
		} else {
			sprite = buildColorSprite(lines)
		}
	}
	b.AddSprite(line0.Str("NAME", 0), sprite)
	return command.Ok()
}

// CommandFactorySpriteMultiLine is "sprite $name [WxH/n] {" ... "}".
func CommandFactorySpriteMultiLine[D Diagram]() command.Command[D] {
	c := &Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			Lit("sprite"),
			regex.SpaceOneOrMore(),
			Lit(`\$?`),
			Leaf(1, "NAME", "([-.%pLN_]+)"),
			regex.SpaceZeroOrMore(),
			regex.Optional(Leaf(5, "DIM", `\[(\d+)x(\d+)/(?:(\d+)(z)?|(color))\]`)),
			regex.SpaceZeroOrMore(),
			Lit(`\{`), regex.End()),
		Strategy: RemoveStartingQuote,
		TrimEnd:  TrimBoth,
		End:      Compile(`^end[%s]?sprite|\}$`),
	}
	c.Exec = func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
		lines = lines.Trim().RemoveEmptyLines()
		line0 := c.StartArg(lines)
		lines = SubExtract(lines, 1, 1).RemoveEmptyColumns()
		if lines.Size() == 0 {
			return command.Error("No sprite defined.")
		}
		return executeSprite(d.Titled(), line0, lines.Strings())
	}
	return c
}

// CommandFactorySpriteSingleLine is "sprite $name [WxH/nz] data".
func CommandFactorySpriteSingleLine[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-.%pLN_]+)"),
		regex.SpaceZeroOrMore(),
		regex.Optional(Leaf(5, "DIM", `\[(\d+)x(\d+)/(?:(\d+)(z)|(color))\]`)),
		regex.SpaceOneOrMore(),
		Leaf(1, "DATA", "([-_A-Za-z0-9]+)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeSprite(d.Titled(), arg, []string{arg.Str("DATA", 0)})
		})
}

// CommandSpriteBase64 is "sprite $name data:image/png;base64,...". The base64
// patching of PSystemBuilder rewrites these lines before the parsers run,
// so CommandSpriteMd5 handles them in practice.
func CommandSpriteBase64[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-.%pLN_]+)"),
		regex.SpaceOneOrMore(),
		Lit("data:image/png;base64,"),
		Leaf(1, "DATA", "([A-Za-z0-9+/=]+)"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			sprite, ok := imageSprite(arg.Str("DATA", 0))
			if !ok {
				return command.Error("Cannot decode Base64 PNG sprite.")
			}
			d.Titled().AddSprite(arg.Str("NAME", 0), sprite)
			return command.Ok()
		})
}

// CommandSpriteMd5 is "sprite $name data:image/png;md5,...", the form of
// base64 images after PatchBase64.
func CommandSpriteMd5[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-.%pLN_]+)"),
		regex.SpaceOneOrMore(),
		Lit("data:image/png;md5,"),
		Leaf(1, "MD5", "([0-9a-f]+)"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			md5 := arg.Str("MD5", 0)
			b64, ok := d.Titled().Skin().FromMd5(md5)
			if !ok {
				return command.Error("Unknown MD5 sprite reference: " + md5)
			}
			sprite, ok := imageSprite(b64)
			if !ok {
				return command.Error("Cannot decode Base64 PNG sprite.")
			}
			d.Titled().AddSprite(arg.Str("NAME", 0), sprite)
			return command.Ok()
		})
}

// CommandSpriteSvg is "sprite $name <svg ...>...</svg>".
func CommandSpriteSvg[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-%pLN_]+)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "SVG", `(\<svg\b.*\</svg\>)`), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			parser := klimt.SvgSpriteParserFactoryCreate([]string{arg.Str("SVG", 0)}, d.Titled().Pragma(),
				d.Titled().MD5Map())
			d.Titled().AddSprite(arg.Str("NAME", 0), parser)
			return command.Ok()
		})
}

// CommandSpriteSvgMultiline is "sprite $name <svg ..." ... "</svg>".
func CommandSpriteSvgMultiline[D Diagram]() command.Command[D] {
	c := &Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			Lit("sprite"),
			regex.SpaceOneOrMore(),
			Lit(`\$?`),
			Leaf(1, "NAME", "([-%pLN_]+)"),
			regex.SpaceOneOrMore(),
			Leaf(1, "SVGSTART", `(\<svg\b.*)`),
			regex.End()),
		Strategy: KeepStartingQuote,
		TrimEnd:  TrimBoth,
		End:      Compile(`(.*\</svg\>)$`),
	}
	c.Exec = func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		svg := []string{line0.Str("SVGSTART", 0)}
		svg = append(svg, SubExtract(lines, 1, 0).Strings()...)
		parser := klimt.SvgSpriteParserFactoryCreate(svg, d.Titled().Pragma(), d.Titled().MD5Map())
		d.Titled().AddSprite(line0.Str("NAME", 0), parser)
		return command.Ok()
	}
	return c
}

// stdlibNotBundled is the error of a stdlib sprite of a library that is
// not bundled: Stdlib.retrieve throws UncheckedIOException, as upstream
// does for a library that does not exist.
func stdlibNotBundled(lib string) {
	Throw("java.io.UncheckedIOException", "java.io.FileNotFoundException: stdlib/"+lib+
		"/info.spm (No such file or directory)")
}

// retrieveStdlib is Stdlib.retrieve for the sprite commands.
func retrieveStdlib(name string) *stdlib.Library {
	lib, err := stdlib.Retrieve(name)
	if err != nil {
		if _, ok := err.(*stdlib.FileNotFoundError); ok {
			stdlibNotBundled(name)
		}
		Throw("java.io.UncheckedIOException", stdlib.IOExceptionString(err))
	}
	return lib
}

// CommandSpriteStdlib is "sprite $name #lib#sprite".
func CommandSpriteStdlib[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$`),
		Leaf(1, "NAME", "([-%pLN_]+)"),
		regex.SpaceZeroOrMore(),
		Leaf(2, "STDLIB", "#([^#]+)#([^%s]+)"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			lib := retrieveStdlib(arg.Str("STDLIB", 0))
			sprite, err := lib.Sprite(arg.Str("STDLIB", 1))
			if err != nil {
				return command.Error("Cannot read sprite " + stdlib.IOExceptionString(err))
			}
			// An unknown sprite is added as null, as upstream does.
			var s klimt.Sprite
			if sprite != nil {
				s = klimt.NewStdlibSprite(sprite.Width, sprite.Height, sprite.Data)
			}
			d.Titled().AddSprite(arg.Str("NAME", 0), s)
			return command.Ok()
		})
}

// CommandSpriteStdlibSvg is "sprite $name :lib:sprite".
func CommandSpriteStdlibSvg[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-%pLN_]+)"),
		regex.SpaceZeroOrMore(),
		Leaf(2, "STDLIB", ":([^:]+):([^%s]+)"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			lib := retrieveStdlib(arg.Str("STDLIB", 0))
			d.Titled().AddSprite(arg.Str("NAME", 0), klimt.NewStdlibSpriteSvg(lib, arg.Str("STDLIB", 1)))
			return command.Ok()
		})
}

// CommandSpriteFile is "sprite $name file", where file is a path, a
// "zipfile~entry" path or "jar:name" for upstream's internal sprites.
func CommandSpriteFile[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("sprite"),
		regex.SpaceOneOrMore(),
		Lit(`\$?`),
		Leaf(1, "NAME", "([-%pLN_]+)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "FILE", "([^<>%g#]*)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			src := arg.Str("FILE", 0)
			var sprite klimt.Sprite
			switch {
			case strings.HasPrefix(src, "jar:"):
				name := src[4:]
				sprite = klimt.SpriteImageFromInternal(name)
				if sprite == nil {
					return command.Error("No such internal sprite: " + name)
				}
			case strings.Contains(src, "~"):
				idx := strings.LastIndex(src, "~")
				data, ok := readSpriteFile(d.Titled().RenderContext(), src[:idx])
				if !ok {
					return command.Error("Cannot read: " + src)
				}
				s := spriteFromZip(data, src[idx+1:])
				if s == nil {
					return command.Error("Cannot read: " + src)
				}
				sprite = s
			default:
				data, ok := readSpriteFile(d.Titled().RenderContext(), src)
				if !ok {
					return command.Error("Cannot read: " + src)
				}
				s := spriteFromData(path.Base(src), data)
				if s == nil {
					return command.Error("Cannot read: " + src)
				}
				sprite = s
			}
			d.Titled().AddSprite(arg.Str("NAME", 0), sprite)
			return command.Ok()
		})
}

// readSpriteFile reads a file relative to the current directory of the
// render context (upstream's FileSystem.getFile), through its security
// profile, from the host files when the render has them and from the
// operating system otherwise; ok is false when it does not exist, cannot be
// read or is denied.
func readSpriteFile(ctx *core.RenderContext, name string) ([]byte, bool) {
	if ctx == nil {
		return nil, false
	}
	access := &security.Access{FS: ctx.FS}
	if ctx.Security != nil {
		access.Config = *ctx.Security
	} else {
		access.Config.Profile, _ = security.ParseProfile(ctx.SecurityProfile)
	}
	dir := ctx.CurrentDir
	if dir == "" {
		dir = "."
	}
	var p string
	if ctx.FS != nil {
		p = path.Clean("/" + path.Join(filepath.ToSlash(dir), name))
		if path.IsAbs(name) {
			p = path.Clean(name)
		}
	} else {
		p = name
		if !filepath.IsAbs(name) {
			p = filepath.Join(dir, name)
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
	}
	data, err := access.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return data, true
}

func isSvgName(name string) bool { return strings.HasSuffix(strings.ToLower(name), ".svg") }

// spriteFromData is a sprite read from a file: an SVG sprite for ".svg"
// files, an image sprite otherwise; nil when it cannot be read.
func spriteFromData(name string, data []byte) klimt.Sprite {
	if isSvgName(name) {
		return &SpriteDef{Kind: SpriteKindSvg, SVG: []string{string(data)}}
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	return &SpriteDef{Kind: SpriteKindImage, Width: bounds.Dx(), Height: bounds.Dy(), Image: img}
}

// spriteFromZip reads the entry name of a zip archive; nil when the entry
// is missing.
func spriteFromZip(data []byte, name string) klimt.Sprite {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil
		}
		if isSvgName(name) {
			return &SpriteDef{Kind: SpriteKindSvg, SVG: []string{string(content)}}
		}
		s, ok := spriteImage(content)
		if !ok {
			return nil
		}
		return s
	}
	return nil
}
