// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"bytes"
	"embed"
	"image"
	"strings"
	"sync"
)

// internalSprites are upstream's /sprites resources of the jar, the
// ArchiMate icons that upstream takes from Archi: sprites/archimate/<name>.svg
// or .png, unchanged (NOTICE).
//
//go:embed sprites/archimate
var internalSprites embed.FS

// internalSpriteCache holds the jar sprites already read, by name: the
// parsed SVG sprite (ISvgSpriteParser) or the decoded PNG image
// (PortableImage). Names without a sprite are not kept, so it holds at most
// the embedded sprites.
var internalSpriteCache sync.Map

// SpriteImageFromInternal returns the sprite of upstream's jar resources
// (sprites/<name>.svg or .png, the archimate sprites), nil when there is
// none, as upstream's SpriteImage.fromInternal: an SVG file is parsed with
// the default SVG sprite parser, without pragma, and a PNG file is an image
// sprite. A file that cannot be read gives nil, as upstream catches the
// exception. Upstream's TeaVM branch does not apply.
//
// Upstream reads the file at each call. Here each sprite is read once for
// the process: the SVG sprite is shared, as the parser is safe for
// concurrent use, and a PNG sprite is a new image sprite around the shared
// decoded image, which image sprites only read.
func SpriteImageFromInternal(name string) Sprite {
	if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".svg") {
		throw(illegalArgument, "")
	}
	v, ok := internalSpriteCache.Load(name)
	if !ok {
		if v = readInternalSprite(name); v == nil {
			return nil
		}
		v, _ = internalSpriteCache.LoadOrStore(name, v)
	}
	switch v := v.(type) {
	case ISvgSpriteParser:
		return v
	case PortableImage:
		return NewSpriteImage(v)
	}
	return nil
}

// readInternalSprite reads the jar sprite of a name: its parsed SVG sprite,
// its decoded PNG image, or nil.
func readInternalSprite(name string) any {
	// A name that is not a valid path (with "..") is not embedded
	if data, err := internalSprites.ReadFile("sprites/" + name + ".svg"); err == nil {
		return SvgSpriteParserFactoryCreateDefault(string(data))
	}
	data, err := internalSprites.ReadFile("sprites/" + name + ".png")
	if err != nil {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return PortableImageFromImage(img)
}
