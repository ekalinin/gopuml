// SPDX-License-Identifier: MIT
// Provenance: original

package text

import (
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

// DecodeUTF8 converts bytes like Java's UTF-8 decoder: each malformed
// sequence becomes one U+FFFD. A malformed sequence is the longest prefix
// that could start a valid sequence, at least one byte; as in Java, an
// encoded surrogate counts as one malformed sequence of three bytes.
func DecodeUTF8(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	var b strings.Builder
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			size = malformedLength(data)
		}
		b.WriteRune(r)
		data = data[size:]
	}
	return b.String()
}

// malformedLength is the length of the malformed sequence at the start of
// data, as Java's UTF-8 decoder reports it.
func malformedLength(data []byte) int {
	n, lo, hi := 0, byte(0x80), byte(0xBF)
	switch c := data[0]; {
	case c >= 0xC2 && c <= 0xDF:
		n = 2
	case c == 0xE0:
		n, lo = 3, 0xA0
	case c >= 0xE1 && c <= 0xEF:
		n = 3
	case c == 0xF0:
		n, lo = 4, 0x90
	case c >= 0xF1 && c <= 0xF3:
		n = 4
	case c == 0xF4:
		n, hi = 4, 0x8F
	default:
		return 1
	}
	i := 1
	for ; i < n && i < len(data); i++ {
		if c := data[i]; (i == 1 && (c < lo || c > hi)) || c < 0x80 || c > 0xBF {
			break
		}
	}
	return i
}

// javaCharsetAliases are Java names of common charsets that are not IANA
// names.
var javaCharsetAliases = map[string]string{
	"utf8": "UTF-8", "iso8859_1": "ISO-8859-1", "ascii": "US-ASCII",
	"cp1250": "windows-1250", "cp1251": "windows-1251", "cp1252": "windows-1252",
}

// Charset decodes bytes like a Java charset.
type Charset struct {
	enc encoding.Encoding // nil for UTF-8
}

// LookupCharset returns the charset Java's Charset.forName(name) would
// return, or ok false for an unknown charset.
func LookupCharset(name string) (c Charset, ok bool) {
	if alias, found := javaCharsetAliases[strings.ToLower(name)]; found {
		name = alias
	}
	enc, err := ianaindex.IANA.Encoding(name)
	if err != nil || enc == nil {
		return Charset{}, false
	}
	if canonical, _ := ianaindex.IANA.Name(enc); canonical == "UTF-8" {
		return Charset{}, true
	}
	return Charset{enc: enc}, true
}

// Decode converts bytes to a string.
func (c Charset) Decode(data []byte) string {
	if c.enc == nil {
		return DecodeUTF8(data)
	}
	out, err := c.enc.NewDecoder().Bytes(data)
	if err != nil {
		return DecodeUTF8(data)
	}
	return string(out)
}

// NewReader returns a reader of r converted to UTF-8. For UTF-8 it returns
// r: invalid bytes are left to DecodeUTF8.
func (c Charset) NewReader(r io.Reader) io.Reader {
	if c.enc == nil {
		return r
	}
	return c.enc.NewDecoder().Reader(r)
}
