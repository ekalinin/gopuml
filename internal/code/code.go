// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/AsciiEncoder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/AsciiEncoderHex.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/CompressionZlib.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/CompressionHuffman.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/CompressionGZip.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/CompressionZip.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/deflate/OutputStreamProtected.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package code

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrNoCompression is upstream's NoPlantumlCompressionException: the text is
// not valid encoded diagram source.
var ErrNoCompression = errors.New("no plantuml compression")

func compressionError(err error) error { return fmt.Errorf("%w: %v", ErrNoCompression, err) }

var (
	encode6bit [64]byte
	decode6bit [128]byte
)

func init() {
	for b := range byte(64) {
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
		encode6bit[b] = c
		decode6bit[c] = b
	}
}

// encodeASCII writes data with PlantUML's 64-character alphabet, 3 bytes to
// 4 characters, padding the last group with zero bytes.
func encodeASCII(data []byte) string {
	var b strings.Builder
	at := func(i int) int {
		if i < len(data) {
			return int(data[i])
		}
		return 0
	}
	for i := 0; i < len(data); i += 3 {
		b1, b2, b3 := at(i), at(i+1), at(i+2)
		b.WriteByte(encode6bit[(b1>>2)&0x3f])
		b.WriteByte(encode6bit[((b1&0x3)<<4|b2>>4)&0x3f])
		b.WriteByte(encode6bit[((b2&0xf)<<2|b3>>6)&0x3f])
		b.WriteByte(encode6bit[b3&0x3f])
	}
	return b.String()
}

// decodeASCII reverses encodeASCII. Characters outside the alphabet but
// below 128 decode as 0, as upstream; others are an error.
func decodeASCII(s string) ([]byte, error) {
	n := len(s)
	if r := n % 4; r != 0 {
		n += 4 - r
	}
	out := make([]byte, (n*3+3)/4)
	ch := func(i int) (int, error) {
		if i >= len(s) {
			return int(decode6bit['0']), nil
		}
		if s[i] >= 128 {
			return 0, compressionError(fmt.Errorf("character %q outside the alphabet", s[i]))
		}
		return int(decode6bit[s[i]]), nil
	}
	pos := 0
	for i := 0; i < len(s); i += 4 {
		var c [4]int
		for k := range 4 {
			v, err := ch(i + k)
			if err != nil {
				return nil, err
			}
			c[k] = v
		}
		out[pos] = byte(c[0]<<2 | c[1]>>4)
		out[pos+1] = byte((c[1]&0x0f)<<4 | c[2]>>2)
		out[pos+2] = byte((c[2]&0x3)<<6 | c[3])
		pos += 3
	}
	return out, nil
}

// encodeHex and decodeHex are the uncompressed ~h form.
func encodeHex(data []byte) string {
	var b strings.Builder
	for _, x := range data {
		fmt.Fprintf(&b, "%02x", x)
	}
	return b.String()
}

func decodeHex(s string) ([]byte, error) {
	out := make([]byte, len(s)/2)
	for i := range out {
		v, err := strconv.ParseInt(s[i*2:i*2+2], 16, 32)
		if err != nil {
			return nil, compressionError(err)
		}
		out[i] = byte(v)
	}
	return out, nil
}

// MaxDeflateOutput is upstream's limit on decompressed deflate data.
const MaxDeflateOutput = 1024 * 1024

// maxArchiveOutput is upstream's limit for gzip and zip data.
const maxArchiveOutput = 200_000

// deflateRaw compresses with raw deflate at the best level. Empty input
// gives no data, as upstream.
func deflateRaw(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.BestCompression)
	w.Write(in)
	w.Close()
	return buf.Bytes()
}

func inflateRaw(in []byte) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(in)), MaxDeflateOutput+1))
	if err != nil {
		return nil, compressionError(err)
	}
	if len(data) > MaxDeflateOutput {
		return nil, compressionError(errors.New("Too big"))
	}
	return data, nil
}

// inflateZlib decompresses the legacy ~1 form: deflate with a zlib header.
func inflateZlib(in []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, compressionError(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, compressionError(err)
	}
	return data, nil
}

func readLimited(r io.Reader, what string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxArchiveOutput+1))
	if err != nil {
		return nil, compressionError(err)
	}
	if len(data) > maxArchiveOutput {
		return nil, compressionError(errors.New(what + " error"))
	}
	return data, nil
}

func gunzip(in []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, compressionError(err)
	}
	// Like Java's GZIPInputStream, ignore bytes after the first member (the
	// alphabet decoding leaves zero padding).
	r.Multistream(false)
	return readLimited(r, "Gzip")
}

// unzip returns the first entry of a zip archive, read from its local file
// header like Java's ZipInputStream, so trailing bytes do not matter.
func unzip(in []byte) ([]byte, error) {
	const headerLen = 30
	if len(in) < headerLen || !bytes.HasPrefix(in, []byte("PK\x03\x04")) {
		return nil, compressionError(errors.New("not a zip entry"))
	}
	u16 := func(i int) int { return int(in[i]) | int(in[i+1])<<8 }
	u32 := func(i int) int { return u16(i) | u16(i+2)<<16 }
	method, csize := u16(8), u32(18)
	start := headerLen + u16(26) + u16(28)
	if start > len(in) {
		return nil, compressionError(errors.New("truncated zip entry"))
	}
	body := in[start:]
	switch method {
	case 0: // stored
		if csize > len(body) {
			return nil, compressionError(errors.New("truncated zip entry"))
		}
		return readLimited(bytes.NewReader(body[:csize]), "Zip")
	case 8: // deflated
		return readLimited(flate.NewReader(bytes.NewReader(body)), "Zip")
	}
	return nil, compressionError(fmt.Errorf("unsupported zip method %d", method))
}
