// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/JsonWriter.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

// jsonWriter receives the serialization events of a value. minimalWriter
// writes the minimal form; prettyPrintWriter adds whitespace.
type jsonWriter interface {
	writeLiteral(value string)
	writeNumber(s string)
	writeString(s string)
	writeArrayOpen()
	writeArrayClose()
	writeArraySeparator()
	writeObjectOpen()
	writeObjectClose()
	writeMemberName(name string)
	writeMemberSeparator()
	writeObjectSeparator()
}

const controlCharactersEnd = 0x001f

// Replacements of the escaped characters.
const (
	quotChars = `\"`
	bsChars   = `\\`
	lfChars   = `\n`
	crChars   = `\r`
	tabChars  = `\t`
	// In JavaScript, U+2028 and U+2029 characters count as line endings and
	// must be encoded.
	// http://stackoverflow.com/questions/2965293/javascript-parse-error-on-u2028-unicode-character
	unicode2028Chars = "\\u2028"
	unicode2029Chars = "\\u2029"
)

const hexDigits = "0123456789abcdef"

// UTF-8 encodings of U+2028 and U+2029.
const (
	lineSeparator      = "\xe2\x80\xa8"
	paragraphSeparator = "\xe2\x80\xa9"
)

// minimalWriter writes JSON in its minimal form, without any additional
// whitespace.
type minimalWriter struct {
	writer *writingBuffer
}

func (w *minimalWriter) writeLiteral(value string) { w.writer.writeString(value) }
func (w *minimalWriter) writeNumber(s string)      { w.writer.writeString(s) }

func (w *minimalWriter) writeString(s string) {
	w.writer.writeByte('"')
	w.writeJSONString(s)
	w.writer.writeByte('"')
}

func (w *minimalWriter) writeArrayOpen()      { w.writer.writeByte('[') }
func (w *minimalWriter) writeArrayClose()     { w.writer.writeByte(']') }
func (w *minimalWriter) writeArraySeparator() { w.writer.writeByte(',') }
func (w *minimalWriter) writeObjectOpen()     { w.writer.writeByte('{') }
func (w *minimalWriter) writeObjectClose()    { w.writer.writeByte('}') }

func (w *minimalWriter) writeMemberName(name string) {
	w.writer.writeByte('"')
	w.writeJSONString(name)
	w.writer.writeByte('"')
}

func (w *minimalWriter) writeMemberSeparator() { w.writer.writeByte(':') }
func (w *minimalWriter) writeObjectSeparator() { w.writer.writeByte(',') }

// writeJSONString writes s with upstream's escapes: quote, backslash and the
// control characters, plus U+2028 and U+2029. Every other character,
// including DEL and invalid UTF-8 bytes, is written as is.
func (w *minimalWriter) writeJSONString(s string) {
	start := 0
	for index := 0; index < len(s); index++ {
		replacement, size := replacementChars(s[index:])
		if replacement != "" {
			w.writer.writeString(s[start:index])
			w.writer.writeString(replacement)
			index += size - 1
			start = index + 1
		}
	}
	w.writer.writeString(s[start:])
}

// replacementChars returns the escape of the character that s starts with
// and its length in bytes, or "" if the character is written as is.
// Upstream decides per UTF-16 unit; only U+2028 and U+2029 are escaped above
// the ASCII range, so looking at UTF-8 bytes gives the same result.
func replacementChars(s string) (string, int) {
	ch := s[0]
	if ch > '\\' {
		switch {
		case len(s) >= 3 && s[:3] == lineSeparator:
			return unicode2028Chars, 3
		case len(s) >= 3 && s[:3] == paragraphSeparator:
			return unicode2029Chars, 3
		}
		// The lower range contains 'a' .. 'z'. Only 2 checks required.
		return "", 1
	}
	if ch == '\\' {
		return bsChars, 1
	}
	if ch > '"' {
		// This range contains '0' .. '9' and 'A' .. 'Z'. Need 3 checks to get
		// here.
		return "", 1
	}
	if ch == '"' {
		return quotChars, 1
	}
	if ch > controlCharactersEnd {
		return "", 1
	}
	switch ch {
	case '\n':
		return lfChars, 1
	case '\r':
		return crChars, 1
	case '\t':
		return tabChars, 1
	}
	return `\u00` + string(hexDigits[ch>>4&0x000f]) + string(hexDigits[ch&0x000f]), 1
}
