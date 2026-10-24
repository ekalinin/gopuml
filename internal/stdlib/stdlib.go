// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/Stdlib.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/spm/SpmChannel.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/FutureImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package stdlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/brotli"
)

// Channel is upstream's SpmChannel: one Brotli-compressed file of a
// library folder.
type Channel int

// The channels of a library, in upstream's order.
const (
	ChannelInfo Channel = iota
	ChannelPuml
	ChannelJSON
	ChannelSprite
	ChannelSvg
	ChannelImage
)

var channelNames = [...]string{"info", "puml", "json", "sprite", "svg", "image"}

// fileName is SpmChannel.getFileName.
func (c Channel) fileName() string { return channelNames[c] + ".spm" }

// FileNotFoundError is the java.io.FileNotFoundException of a missing
// channel file: upstream looks for the file in its jar, then in the
// current directory, where stdlib/<name>/<channel>.spm does not exist
// either.
type FileNotFoundError struct {
	// Path is the upstream path, "stdlib/<name>/<channel>.spm".
	Path string
}

func (e *FileNotFoundError) Error() string { return e.Path + " (No such file or directory)" }

// IOExceptionString is the toString of the Java IOException of an error of
// this package.
func IOExceptionString(err error) string {
	if _, ok := err.(*FileNotFoundError); ok {
		return "java.io.FileNotFoundException: " + err.Error()
	}
	return "java.io.IOException: " + err.Error()
}

// Library is a folder of the standard library (an upstream Stdlib). Its
// channels are decoded on first use and kept, as upstream keeps them; it
// is safe for concurrent use.
type Library struct {
	name string
	info []byte

	puml, json lazy[map[string][]byte]
	sprites    lazy[map[string]*Sprite]
	svgs       lazy[map[string]string]
	images     lazy[[]*Image]
}

// lazy is a value decoded once.
type lazy[T any] struct {
	once sync.Once
	v    T
	err  error
}

func (l *lazy[T]) get(load func() (T, error)) (T, error) {
	l.once.Do(func() { l.v, l.err = load() })
	return l.v, l.err
}

// Sprite is a monochrome sprite of the sprite channel: 16 gray levels, two
// pixels per byte (the high nibble first), (Height+1)/2 rows of Width bytes
// (upstream's StdlibSprite data).
type Sprite struct {
	Width, Height int
	Data          []byte
}

// Image is an image of the image channel (upstream's FutureImage): Width
// times Height indices into the color table of the channel.
type Image struct {
	Width, Height int
	colors        []uint32
	data          []byte
	pixels        lazy[[]uint32]
}

// The libraries already read, by the name they were retrieved with
// (upstream's static Stdlib.all map).
var (
	allMu sync.Mutex
	all   = map[string]*Library{}
)

// Retrieve is Stdlib.retrieve: the library of a folder name, as given
// (folder names are lower case), following the link key of its info.spm.
// A folder without info.spm returns a *FileNotFoundError. Failures are not
// kept, as upstream's computeIfAbsent keeps nothing when it throws.
func Retrieve(name string) (*Library, error) {
	allMu.Lock()
	defer allMu.Unlock()
	return retrieveLocked(name, nil)
}

func retrieveLocked(name string, seen []string) (*Library, error) {
	if lib, ok := all[name]; ok {
		return lib, nil
	}
	// A link cycle overflows the stack upstream; here it is an error.
	if slices.Contains(seen, name) {
		return nil, errors.New("stdlib: link cycle through " + name)
	}
	info, err := readChannel(name, ChannelInfo)
	if err != nil {
		return nil, err
	}
	lib := &Library{name: name, info: info}
	if link, ok := infoValue(info, "link"); ok {
		lib, err = retrieveLocked(link, append(seen, name))
		if err != nil {
			return nil, err
		}
	}
	all[name] = lib
	return lib, nil
}

// infoValue returns a key of info.spm, read as the Stdlib constructor
// reads it: lines that split on "=" into exactly two parts, the last one
// winning.
func infoValue(info []byte, key string) (string, bool) {
	value, found := "", false
	for _, line := range lines(string(info)) {
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.Split(line, "=")
		// String.split drops the trailing empty strings.
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		if len(parts) == 2 && parts[0] == key {
			value, found = parts[1], true
		}
	}
	return value, found
}

// lines splits text as BufferedReader.lines does, at "\n", "\r" or "\r\n".
func lines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

// FolderNames is Stdlib.getAllFolderNames: the lines of home.spm, sorted
// and without duplicates (a TreeSet; the names are ASCII, so byte order is
// Java's order).
func FolderNames() []string {
	home, err := files.ReadFile(homeFile)
	if err != nil {
		return nil
	}
	names := lines(string(home))
	slices.Sort(names)
	return slices.Compact(names)
}

// readChannel returns the decoded content of a channel file.
func readChannel(name string, c Channel) ([]byte, error) {
	data, err := files.ReadFile(name + "/" + c.fileName())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrInvalid) {
			return nil, &FileNotFoundError{Path: "stdlib/" + name + "/" + c.fileName()}
		}
		return nil, err
	}
	return io.ReadAll(brotli.NewReader(bytes.NewReader(data)))
}

// Name returns the folder name of the library (after its link).
func (l *Library) Name() string { return l.name }

// Info returns the content of info.spm.
func (l *Library) Info() []byte { return l.info }

// Puml is loadPumlResource: the .puml file of a lower-case path without
// extension, nil when the library has none.
func (l *Library) Puml(name string) ([]byte, error) {
	m, err := l.puml.get(func() (map[string][]byte, error) { return l.readEntries(ChannelPuml) })
	return m[name], err
}

// JSON is loadJsonResource: the .json file of a lower-case path, nil when
// the library has none.
func (l *Library) JSON(name string) ([]byte, error) {
	m, err := l.json.get(func() (map[string][]byte, error) { return l.readEntries(ChannelJSON) })
	return m[name], err
}

// readEntries is initMapIfNeeded: a count, then for each entry a name and
// its bytes; names are kept in lower case.
func (l *Library) readEntries(c Channel) (map[string][]byte, error) {
	r, err := l.open(c)
	if err != nil {
		return nil, err
	}
	nb := r.readInt()
	m := map[string][]byte{}
	for i := 0; i < nb && r.err == nil; i++ {
		name := r.readUTF()
		data := r.readBytes(r.readInt())
		m[strings.ToLower(name)] = data
	}
	return m, r.err
}

// Sprite is readSprite: the monochrome sprite of a name, nil when the
// library has none.
func (l *Library) Sprite(name string) (*Sprite, error) {
	m, err := l.sprites.get(func() (map[string]*Sprite, error) {
		r, err := l.open(ChannelSprite)
		if err != nil {
			return nil, err
		}
		nb := r.readInt()
		m := map[string]*Sprite{}
		for i := 0; i < nb && r.err == nil; i++ {
			name := r.readUTF()
			width := r.readInt()
			height := r.readInt()
			nbLines := (height + 1) / 2
			data := r.readBytes(width * nbLines)
			m[name] = &Sprite{Width: width, Height: height, Data: data}
		}
		return m, r.err
	})
	return m[name], err
}

// SvgSprite is readSvgSprite: the SVG text of a sprite name, and whether
// the library has it. Upstream builds the sprite parser of every entry when
// it reads the channel; the caller parses the text instead.
func (l *Library) SvgSprite(name string) (string, bool, error) {
	m, err := l.svgs.get(func() (map[string]string, error) {
		r, err := l.open(ChannelSvg)
		if err != nil {
			return nil, err
		}
		nb := r.readInt()
		m := map[string]string{}
		for i := 0; i < nb && r.err == nil; i++ {
			name := r.readUTF()
			m[name] = r.readUTF()
		}
		return m, r.err
	})
	svg, ok := m[name]
	return svg, ok, err
}

// Images is the list that readDataImagePng reads: a count, the color
// table, then the width, height and color indices of each image.
func (l *Library) Images() ([]*Image, error) {
	return l.images.get(func() ([]*Image, error) {
		r, err := l.open(ChannelImage)
		if err != nil {
			return nil, err
		}
		nb := r.readInt()
		size := r.read2bytes()
		colors := make([]uint32, 0, size)
		for i := 0; i < size && r.err == nil; i++ {
			argb := r.readBytes(4)
			if len(argb) == 4 {
				colors = append(colors, binary.BigEndian.Uint32(argb))
			}
		}
		var images []*Image
		for i := 0; i < nb && r.err == nil; i++ {
			width := r.readInt()
			height := r.readInt()
			data := r.readBytes(width * height * 2)
			images = append(images, &Image{Width: width, Height: height, colors: colors, data: data})
		}
		return images, r.err
	})
}

// ARGB is FutureImage.getNow: the pixels of the image, row by row, as
// 0xAARRGGBB. A color index outside the color table is an error, as
// upstream's List.get throws.
func (im *Image) ARGB() ([]uint32, error) {
	return im.pixels.get(func() ([]uint32, error) {
		pixels := make([]uint32, im.Width*im.Height)
		for i := range pixels {
			index := int(im.data[2*i])<<8 | int(im.data[2*i+1])
			if index >= len(im.colors) {
				return nil, &IndexError{Index: index, Length: len(im.colors)}
			}
			pixels[i] = im.colors[index]
		}
		im.data = nil // free memory, as upstream does
		return pixels, nil
	})
}

// IndexError is the java.lang.IndexOutOfBoundsException of a list index.
type IndexError struct{ Index, Length int }

func (e *IndexError) Error() string {
	return "Index " + strconv.Itoa(e.Index) + " out of bounds for length " + strconv.Itoa(e.Length)
}

// open decodes a channel of the library for reading.
func (l *Library) open(c Channel) (*dataReader, error) {
	data, err := readChannel(l.name, c)
	if err != nil {
		return nil, err
	}
	return &dataReader{data: data}, nil
}

// errEOF is the java.io.EOFException of a truncated channel.
var errEOF = errors.New("java.io.EOFException")

// dataReader reads the big-endian values of Java's DataInputStream. After
// the first error every read returns zero values and err keeps the error.
type dataReader struct {
	data []byte
	pos  int
	err  error
}

func (r *dataReader) readBytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || n > len(r.data)-r.pos {
		r.err = errEOF
		return nil
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *dataReader) readInt() int {
	b := r.readBytes(4)
	if b == nil {
		return 0
	}
	return int(int32(binary.BigEndian.Uint32(b)))
}

func (r *dataReader) read2bytes() int {
	b := r.readBytes(2)
	if b == nil {
		return 0
	}
	return int(binary.BigEndian.Uint16(b))
}

// readUTF is DataInputStream.readUTF: a length, then the text in Java's
// modified UTF-8 (supplementary characters as two 3-byte surrogates).
func (r *dataReader) readUTF() string {
	b := r.readBytes(r.read2bytes())
	if r.err != nil {
		return ""
	}
	units := make([]uint16, 0, len(b))
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c < 0x80:
			units = append(units, uint16(c))
			i++
		case c>>5 == 0x6 && i+1 < len(b) && b[i+1]>>6 == 0x2:
			units = append(units, uint16(c&0x1F)<<6|uint16(b[i+1]&0x3F))
			i += 2
		case c>>4 == 0xE && i+2 < len(b) && b[i+1]>>6 == 0x2 && b[i+2]>>6 == 0x2:
			units = append(units, uint16(c&0x0F)<<12|uint16(b[i+1]&0x3F)<<6|uint16(b[i+2]&0x3F))
			i += 3
		default:
			r.err = errors.New("java.io.UTFDataFormatException: malformed input around byte " + strconv.Itoa(i))
			return ""
		}
	}
	return string(utf16.Decode(units))
}
