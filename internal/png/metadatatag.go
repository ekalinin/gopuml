// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/png/MetadataTag.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

// Package png reads and writes the PNG features upstream uses: the text
// chunk that embeds the diagram source.
package png

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

var signature = []byte("\x89PNG\r\n\x1a\n")

// ErrNotPNG reports data that is not a PNG image.
var ErrNotPNG = errors.New("not a PNG image")

// MetadataTag is upstream's MetadataTag.getData: the text stored under
// keyword in a tEXt, iTXt or zTXt chunk, found in the order of the Java
// image I/O metadata trees (tEXt, then iTXt, then zTXt). ok is false when
// no chunk has the keyword.
func MetadataTag(r io.Reader, keyword string) (text string, ok bool, err error) {
	br := bufio.NewReader(r)
	head := make([]byte, len(signature))
	if _, err := io.ReadFull(br, head); err != nil || !bytes.Equal(head, signature) {
		return "", false, ErrNotPNG
	}
	var found [3]*string // by kind: tEXt, iTXt, zTXt
	for {
		var hdr [8]byte
		if _, err := io.ReadFull(br, hdr[:]); err != nil {
			return "", false, ErrNotPNG
		}
		length := binary.BigEndian.Uint32(hdr[:4])
		kind := string(hdr[4:8])
		if length > 1<<30 {
			return "", false, ErrNotPNG
		}
		data := make([]byte, length+4) // with the CRC
		if _, err := io.ReadFull(br, data); err != nil {
			return "", false, ErrNotPNG
		}
		data = data[:length]
		if kind == "IEND" {
			break
		}
		i, isText := textSlots[kind]
		if !isText || found[i] != nil {
			continue
		}
		k, v, err := parseText(kind, data)
		if err != nil {
			return "", false, err
		}
		if k == keyword {
			found[i] = &v
		}
	}
	for _, v := range found {
		if v != nil {
			return *v, true, nil
		}
	}
	return "", false, nil
}

// textSlots order the text chunk kinds as Java's metadata trees do.
var textSlots = map[string]int{"tEXt": 0, "iTXt": 1, "zTXt": 2}

// latin1 decodes ISO-8859-1 text.
func latin1(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func inflate(b []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

// parseText returns the keyword and the text of a text chunk.
func parseText(kind string, data []byte) (keyword, value string, err error) {
	k, rest, found := bytes.Cut(data, []byte{0})
	if !found {
		return "", "", ErrNotPNG
	}
	keyword = latin1(k)
	switch kind {
	case "tEXt":
		return keyword, latin1(rest), nil
	case "zTXt":
		if len(rest) < 1 {
			return "", "", ErrNotPNG
		}
		b, err := inflate(rest[1:])
		if err != nil {
			return "", "", err
		}
		return keyword, latin1(b), nil
	}
	// iTXt: compression flag, method, language tag, translated keyword,
	// then UTF-8 text.
	if len(rest) < 2 {
		return "", "", ErrNotPNG
	}
	compressed := rest[0] == 1
	rest = rest[2:]
	for range 2 {
		if _, rest, found = bytes.Cut(rest, []byte{0}); !found {
			return "", "", ErrNotPNG
		}
	}
	if compressed {
		if rest, err = inflate(rest); err != nil {
			return "", "", err
		}
	}
	return keyword, text.DecodeUTF8(rest), nil
}
