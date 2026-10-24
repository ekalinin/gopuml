// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/Sprite.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteGrayLevel.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteMonochrome.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteColor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteColorBuilder4096.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/ColorPalette4096.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteImage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/AsciiEncoder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/CompressionZlib.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/Base64Coder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/security/SImageIO.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"bytes"
	"compress/flate"
	"errors"
	"image"
	_ "image/gif"  // sprite files
	_ "image/jpeg" // sprite files
	_ "image/png"  // sprite files and base64 sprites
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// SpriteKind is the kind of a sprite definition.
type SpriteKind int

var spriteKindNames = [...]string{"SpriteMonochrome", "SpriteColor", "SpriteImage", "SpriteSvg", "StdlibSprite"}

// SpriteDef is a sprite defined by a "sprite" command, decoded as upstream
// decodes it when the command runs. AsTextBlock draws it with the klimt
// port of the upstream class of its kind.
type SpriteDef struct {
	Kind          SpriteKind
	Width, Height int
	// GrayLevel is the number of gray levels of a monochrome sprite.
	GrayLevel int
	// Gray holds the gray level of each pixel of a monochrome sprite, row by
	// row; Color holds the RGB color of each pixel of a color sprite, -1 for
	// none.
	Gray, Color []int32
	// Image is the decoded raster image of an image sprite.
	Image image.Image
	// SVG holds the lines of an SVG sprite.
	SVG []string
	// Stdlib and Name name a stdlib sprite.
	Stdlib, Name string

	drawOnce sync.Once
	drawn    klimt.Sprite
}

// newGrid allocates the pixels of a sprite. Java allocates int[height][width]
// and fails with OutOfMemoryError for huge sprites; here the limit is fixed
// so that the process does not run out of memory.
func newGrid(width, height int) []int32 {
	const maxPixels = 1 << 26
	if width < 0 || height < 0 {
		Throw(NegativeArraySizeException, strconv.Itoa(min(width, height)))
	}
	if height != 0 && width > maxPixels/max(height, 1) {
		Throw(OutOfMemoryError, "Java heap space")
	}
	return make([]int32, width*height)
}

// newMonochrome is upstream's SpriteMonochrome constructor.
func newMonochrome(width, height, grayLevel int) *SpriteDef {
	if grayLevel != 2 && grayLevel != 4 && grayLevel != 8 && grayLevel != 16 {
		Throw(IllegalArgumentException, "")
	}
	return &SpriteDef{Kind: SpriteKindMonochrome, Width: width, Height: height, GrayLevel: grayLevel,
		Gray: newGrid(width, height)}
}

// utf16Of returns the UTF-16 units of s, Java's view of a string.
func utf16Of(s string) []uint16 { return utf16.Encode([]rune(s)) }

// javaHexDigit is Character.digit(c, 16) for ASCII and full-width digits and
// letters; other characters give -1.
func javaHexDigit(c uint16) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	case c >= 0xFF10 && c <= 0xFF19:
		return int(c - 0xFF10)
	case c >= 0xFF21 && c <= 0xFF26:
		return int(c-0xFF21) + 10
	case c >= 0xFF41 && c <= 0xFF46:
		return int(c-0xFF41) + 10
	}
	return -1
}

// decode6bitTable is AsciiEncoder's decode table: 0-9, A-Z, a-z, '-' and
// '_' are 0 to 63; other ASCII characters decode to 0.
var decode6bitTable = func() [128]byte {
	var t [128]byte
	for b := byte(0); b < 64; b++ {
		var c byte
		switch {
		case b < 10:
			c = '0' + b
		case b < 36:
			c = 'A' + b - 10
		case b < 62:
			c = 'a' + b - 36
		case b == 62:
			c = '-'
		default:
			c = '_'
		}
		t[c] = b
	}
	return t
}()

// decode6bit is AsciiEncoder.decode6bit; characters past ASCII throw as the
// array access does upstream.
func decode6bit(c uint16) int {
	if int(c) >= len(decode6bitTable) {
		Throw(ArrayIndexOutOfBoundsException, "Index "+strconv.Itoa(int(c))+" out of bounds for length 128")
	}
	return int(decode6bitTable[c])
}

// buildGraySprite is SpriteGrayLevel.buildSprite.
func buildGraySprite(nbColor, width, height int, lines []string) *SpriteDef {
	switch nbColor {
	case 16:
		if len(lines) == 0 {
			ThrowIndexOutOfBounds(0, 0)
		}
		result := newMonochrome(len(utf16Of(lines[0])), len(lines), 16)
		units := make([][]uint16, len(lines))
		for i, l := range lines {
			units[i] = utf16Of(l)
		}
		for col := 0; col < result.Width; col++ {
			for line := 0; line < result.Height; line++ {
				if col >= len(units[line]) {
					continue
				}
				if c := units[line][col]; c != '0' {
					result.setGray(col, line, javaHexDigit(c))
				}
			}
		}
		return result
	case 8, 4:
		result := newMonochrome(width, height, nbColor)
		units := make([][]uint16, len(lines))
		for i, l := range lines {
			units[i] = utf16Of(l)
		}
		for col := 0; col < result.Width; col++ {
			for line := range units {
				if col >= len(units[line]) {
					continue
				}
				v := decode6bit(units[line][col])
				if nbColor == 8 {
					result.setGray(col, line*2, v/8)
					result.setGray(col, line*2+1, v%8)
				} else {
					result.setGray(col, line*3, v/16)
					result.setGray(col, line*3+1, v%16/4)
					result.setGray(col, line*3+2, v%4)
				}
			}
		}
		return result
	}
	Throw(UnsupportedOperationException, "")
	return nil
}

// asciiDecode is AsciiEncoder.decode: four characters give three bytes,
// missing characters count as '0'.
func asciiDecode(s string) []byte {
	u := utf16Of(s)
	n := len(u)
	if r := n % 4; r != 0 {
		n += 4 - r
	}
	data := make([]byte, (n*3+3)/4)
	at := func(i int) uint16 {
		if i >= len(u) {
			return '0'
		}
		return u[i]
	}
	pos := 0
	for i := 0; i < len(u); i += 4 {
		c1, c2, c3, c4 := decode6bit(at(i)), decode6bit(at(i+1)), decode6bit(at(i+2)), decode6bit(at(i+3))
		data[pos] = byte(c1<<2 | c2>>4)
		data[pos+1] = byte((c2&0x0F)<<4 | c3>>2)
		data[pos+2] = byte((c3&0x3)<<6 | c4)
		pos += 3
	}
	return data
}

// buildGraySpriteZ is SpriteGrayLevel.buildSpriteZ: nil when the data does
// not inflate.
func buildGraySpriteZ(nbColor, width, height int, compressed string) *SpriteDef {
	comp := asciiDecode(compressed)
	padded := append(append([]byte(nil), comp...), make([]byte, 256)...)
	img, err := io.ReadAll(flate.NewReader(bytes.NewReader(padded)))
	if err != nil {
		return nil
	}
	result := newMonochrome(width, height, nbColor)
	cpt := 0
	for line := 0; line < result.Height; line++ {
		for col := 0; col < result.Width; col++ {
			if cpt >= len(img) {
				Throw(ArrayIndexOutOfBoundsException, "Index "+strconv.Itoa(cpt)+" out of bounds for length "+
					strconv.Itoa(len(img)))
			}
			result.setGray(col, line, int(int8(img[cpt])))
			cpt++
		}
	}
	return result
}

// colorValue is ColorPalette4096's alphabet.
const colorValue = ".,$%&*+-:;<=>?@^_~GHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// buildColorSprite is SpriteColorBuilder4096.buildSprite.
func buildColorSprite(lines []string) *SpriteDef {
	if len(lines) == 0 {
		ThrowIndexOutOfBounds(0, 0)
	}
	units := make([][]uint16, len(lines))
	for i, l := range lines {
		units[i] = utf16Of(l)
	}
	width, height := len(units[0])/2, len(lines)
	result := &SpriteDef{Kind: SpriteKindColor, Width: width, Height: height, Gray: newGrid(width, height),
		Color: newGrid(width, height)}
	for col := 0; col < width; col++ {
		for line := 0; line < height; line++ {
			if col*2 >= len(units[line]) {
				continue
			}
			v1 := paletteIndex(units[line][col*2])
			v2 := -1
			if col*2+1 < len(units[line]) {
				v2 = paletteIndex(units[line][col*2+1])
			} else {
				Throw(StringIndexOutOfBoundsException, "begin "+strconv.Itoa(col*2)+", end "+strconv.Itoa(col*2+2)+
					", length "+strconv.Itoa(len(units[line])))
			}
			if v1 == -1 || v2 == -1 {
				ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.klimt.color.XColor.getRGB()" because "rgb" is null`)
			}
			code := v1*64 + v2
			dup := func(v int) int { return v*16 + v }
			rgb := dup(code/256%16)<<16 | dup(code/16%16)<<8 | dup(code%16)
			result.Gray[line*width+col] = -1
			result.Color[line*width+col] = int32(rgb)
		}
	}
	return result
}

func paletteIndex(c uint16) int {
	switch c {
	case '!':
		c = '.'
	case '#':
		c = ','
	}
	return strings.IndexRune(colorValue, rune(c))
}

// imageSprite decodes a base64 PNG image as upstream's SImageIO.read: data
// that is not an image gives a sprite without image, as ImageIO returns
// null; ok is false for an image that cannot be read.
func imageSprite(b64 string) (*SpriteDef, bool) {
	return spriteImage(base64Decode(b64))
}

func spriteImage(data []byte) (*SpriteDef, bool) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return &SpriteDef{Kind: SpriteKindImage}, true
		}
		return nil, false
	}
	bounds := img.Bounds()
	return &SpriteDef{Kind: SpriteKindImage, Width: bounds.Dx(), Height: bounds.Dy(), Image: img}, true
}

// base64Map is Base64Coder's decoding table, -1 for other characters.
var base64Map = func() [128]int8 {
	var m [128]int8
	for i := range m {
		m[i] = -1
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(alphabet); i++ {
		m[alphabet[i]] = int8(i)
	}
	return m
}()

// base64Decode is upstream's Base64Coder.decode.
func base64Decode(s string) []byte {
	in := utf16Of(s)
	n := len(in)
	if n%4 != 0 {
		Throw(IllegalArgumentException, "Length of Base64 encoded input string is not a multiple of 4.")
	}
	for n > 0 && in[n-1] == '=' {
		n--
	}
	oLen := n * 3 / 4
	out := make([]byte, oLen)
	illegal := func() { Throw(IllegalArgumentException, "Illegal character in Base64 encoded data.") }
	ip, op := 0, 0
	for ip < n {
		i0 := in[ip]
		i1 := in[ip+1]
		ip += 2
		var i2, i3 uint16 = 'A', 'A'
		if ip < n {
			i2 = in[ip]
			ip++
		}
		if ip < n {
			i3 = in[ip]
			ip++
		}
		if i0 > 127 || i1 > 127 || i2 > 127 || i3 > 127 {
			illegal()
		}
		b0, b1, b2, b3 := base64Map[i0], base64Map[i1], base64Map[i2], base64Map[i3]
		if b0 < 0 || b1 < 0 || b2 < 0 || b3 < 0 {
			illegal()
		}
		out[op] = byte(int(b0)<<2 | int(b1)>>4)
		op++
		if op < oLen {
			out[op] = byte((int(b1)&0xf)<<4 | int(b2)>>2)
			op++
		}
		if op < oLen {
			out[op] = byte((int(b2)&3)<<6 | int(b3))
			op++
		}
	}
	return out
}

// Sprite kinds, after upstream's Sprite implementations.
const (
	SpriteKindMonochrome SpriteKind = iota // SpriteMonochrome (gray levels)
	SpriteKindColor                        // SpriteColor (4096 colors)
	SpriteKindImage                        // SpriteImage (PNG and other raster images)
	SpriteKindSvg                          // SVG sprites
	SpriteKindStdlib                       // stdlib sprites
)

// String returns the upstream class name of the kind.
func (k SpriteKind) String() string { return spriteKindNames[k] }

// AsTextBlock implements klimt.Sprite.
func (s *SpriteDef) AsTextBlock(fontColor, forcedColor color.HColor, scale float64, backColor color.HColor) klimt.TextBlock {
	s.drawOnce.Do(func() { s.drawn = s.klimtSprite() })
	return s.drawn.AsTextBlock(fontColor, forcedColor, scale, backColor)
}

// klimtSprite returns the klimt sprite of the definition. SVG definitions
// are SVG files (SpriteSvg): the commands of inline SVG add the sprite of
// the SVG parser instead.
func (s *SpriteDef) klimtSprite() klimt.Sprite {
	switch s.Kind {
	case SpriteKindMonochrome:
		return &klimt.SpriteMonochrome{Width: s.Width, Height: s.Height, GrayLevel: s.GrayLevel, Gray: s.Gray}
	case SpriteKindColor:
		return &klimt.SpriteColor{Width: s.Width, Height: s.Height, Gray: s.Gray, Color: s.Color}
	case SpriteKindImage:
		// SpriteImage requires an image, as Objects.requireNonNull does.
		if s.Image == nil {
			Throw(NullPointerException, "")
		}
		return klimt.NewSpriteImage(klimt.PortableImageFromImage(s.Image))
	case SpriteKindSvg:
		return klimt.NewSpriteSvg(strings.Join(s.SVG, ""))
	}
	panic(&color.RuntimeError{Class: UnsupportedOperationException,
		Message: "not ported yet: klimt/sprite/" + s.Kind.String() + " (drawing of sprites)"})
}

// setGray is SpriteMonochrome.setGray: pixels outside the sprite are
// ignored, a level out of range throws.
func (s *SpriteDef) setGray(x, y, level int) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	if level < 0 || level >= s.GrayLevel {
		Throw(IllegalArgumentException, "level="+strconv.Itoa(level)+" grayLevel="+strconv.Itoa(s.GrayLevel))
	}
	s.Gray[y*s.Width+x] = int32(level)
}
