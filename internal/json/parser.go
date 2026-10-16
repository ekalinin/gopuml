// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/JsonParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"errors"
	"io"
	"unicode/utf16"
)

const (
	endOfText         = 0x0003
	maxNestingLevel   = 1000
	minBufferSize     = 10
	defaultBufferSize = 1024
)

// Parser is a streaming parser for JSON text. It reports every event to its
// [Handler].
//
// Beyond standard JSON, the parser accepts one line comment ("// ...") at
// the start of each object member and one after each member value, as
// PlantUML's fork does. Comments are not accepted anywhere else.
//
// The parser works on the UTF-16 form of the input and reads it in chunks
// into a buffer, exactly like upstream, so that error locations and the
// corner cases of the buffer handling match. In particular, a character
// U+0003 in the input ends the text, and a line comment that runs to the end
// of the input reads stale buffer content and usually fails with an
// [*IndexOutOfBoundsError] instead of a [*ParseError].
//
// A Parser may be reused for several inputs. Like upstream, a parse that
// failed can leave state behind (the nesting level and partially captured
// string content) that affects the next parse.
type Parser[A, O any] struct {
	handler Handler[A, O]
	reader  *unitReader
	buffer  []uint16
	// bufferOffset is the input offset of buffer[0].
	bufferOffset int
	index        int
	fill         int
	line         int
	lineOffset   int
	current      uint16
	// captureBuffer is non-empty only while a string spans several buffer
	// fills or contains escapes.
	captureBuffer []uint16
	captureStart  int
	nestingLevel  int
}

// NewParser returns a parser that reports its events to handler. The
// handler is attached to the parser so that it can query its location.
func NewParser[A, O any](handler Handler[A, O]) *Parser[A, O] {
	p := &Parser[A, O]{handler: handler}
	handler.setParser(p)
	return p
}

// Parse parses the input string, which must contain a single JSON value,
// optionally padded with whitespace. It returns a [*ParseError] if the input
// is not valid JSON.
func (p *Parser[A, O]) Parse(s string) error {
	units := utf16.Encode([]rune(s))
	bufferSize := max(minBufferSize, min(defaultBufferSize, len(units)))
	return p.parse(&unitReader{units: units}, bufferSize)
}

// ParseReader reads the entire input from r and parses it as JSON, like
// Parse, with a buffer of 1024 characters.
func (p *Parser[A, O]) ParseReader(r io.Reader) error {
	return p.ParseReaderSize(r, defaultBufferSize)
}

// ParseReaderSize reads the entire input from r and parses it as JSON, like
// Parse, with a buffer of bufferSize characters. The input is read
// completely before parsing starts.
func (p *Parser[A, O]) ParseReaderSize(r io.Reader, bufferSize int) error {
	if r == nil {
		return errors.New("reader is null")
	}
	if bufferSize <= 0 {
		return errors.New("buffersize is zero or negative")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return p.parse(&unitReader{units: utf16.Encode([]rune(string(data)))}, bufferSize)
}

// parseFailure carries an error out of the recursive descent; parse
// recovers it.
type parseFailure struct {
	err error
}

// fail aborts the current parse with err.
func (p *Parser[A, O]) fail(err error) {
	panic(parseFailure{err: err})
}

func (p *Parser[A, O]) parse(reader *unitReader, bufferSize int) (err error) {
	p.reader = reader
	p.buffer = make([]uint16, bufferSize)
	p.bufferOffset = 0
	p.index = 0
	p.fill = 0
	p.line = 1
	p.lineOffset = 0
	p.current = 0
	p.captureStart = -1
	defer func() {
		if r := recover(); r != nil {
			f, ok := r.(parseFailure)
			if !ok {
				panic(r)
			}
			err = f.err
		}
	}()
	p.read()
	p.skipWhiteSpace()
	p.readValue()
	p.skipWhiteSpace()
	if !p.isEndOfText() {
		p.fail(p.error("Unexpected character"))
	}
	return nil
}

func (p *Parser[A, O]) readValue() {
	switch p.current {
	case 'n':
		p.readNull()
	case 't':
		p.readTrue()
	case 'f':
		p.readFalse()
	case '"':
		p.readString()
	case '[':
		p.readArray()
	case '{':
		p.readObject()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.readNumber()
	default:
		p.fail(p.expected("value"))
	}
}

func (p *Parser[A, O]) readArray() {
	array := p.handler.StartArray()
	p.read()
	p.nestingLevel++
	if p.nestingLevel > maxNestingLevel {
		p.fail(p.error("Nesting too deep"))
	}
	p.skipWhiteSpace()
	if p.readChar(']') {
		p.nestingLevel--
		p.handler.EndArray(array)
		return
	}
	for {
		p.skipWhiteSpace()
		p.handler.StartArrayValue(array)
		p.readValue()
		p.handler.EndArrayValue(array)
		p.skipWhiteSpace()
		if !p.readChar(',') {
			break
		}
	}
	if !p.readChar(']') {
		p.fail(p.expected("',' or ']'"))
	}
	p.nestingLevel--
	p.handler.EndArray(array)
}

func (p *Parser[A, O]) readObject() {
	object := p.handler.StartObject()
	p.read()
	p.nestingLevel++
	if p.nestingLevel > maxNestingLevel {
		p.fail(p.error("Nesting too deep"))
	}
	p.skipWhiteSpace()
	if p.readChar('}') {
		p.nestingLevel--
		p.handler.EndObject(object)
		return
	}
	for {
		p.skipWhiteSpace()
		if p.readChar('/') {
			p.skipComment()
		}
		p.handler.StartObjectName(object)
		name := p.readName()
		p.handler.EndObjectName(object, name)
		p.skipWhiteSpace()
		if !p.readChar(':') {
			p.fail(p.expected("':'"))
		}
		p.skipWhiteSpace()
		p.handler.StartObjectValue(object, name)
		p.readValue()
		p.handler.EndObjectValue(object, name)
		p.skipWhiteSpace()
		if p.readChar('/') {
			p.skipComment()
		}
		if !p.readChar(',') {
			break
		}
	}
	if !p.readChar('}') {
		p.fail(p.expected("',' or '}'"))
	}
	p.nestingLevel--
	p.handler.EndObject(object)
}

// skipComment skips a line comment whose first slash has been read.
func (p *Parser[A, O]) skipComment() {
	if !p.readChar('/') {
		p.fail(p.expected("Error in comment"))
	}
	for p.current != '\n' && p.current != '\r' {
		p.read()
	}
	p.skipWhiteSpace()
}

func (p *Parser[A, O]) readName() string {
	if p.current != '"' {
		p.fail(p.expected("name"))
	}
	return p.readStringInternal()
}

func (p *Parser[A, O]) readNull() {
	p.handler.StartNull()
	p.read()
	p.readRequiredChar('u')
	p.readRequiredChar('l')
	p.readRequiredChar('l')
	p.handler.EndNull()
}

func (p *Parser[A, O]) readTrue() {
	p.handler.StartBoolean()
	p.read()
	p.readRequiredChar('r')
	p.readRequiredChar('u')
	p.readRequiredChar('e')
	p.handler.EndBoolean(true)
}

func (p *Parser[A, O]) readFalse() {
	p.handler.StartBoolean()
	p.read()
	p.readRequiredChar('a')
	p.readRequiredChar('l')
	p.readRequiredChar('s')
	p.readRequiredChar('e')
	p.handler.EndBoolean(false)
}

func (p *Parser[A, O]) readRequiredChar(ch uint16) {
	if !p.readChar(ch) {
		p.fail(p.expected("'" + string(rune(ch)) + "'"))
	}
}

func (p *Parser[A, O]) readString() {
	p.handler.StartString()
	p.handler.EndString(p.readStringInternal())
}

func (p *Parser[A, O]) readStringInternal() string {
	p.read()
	p.startCapture()
	for p.current != '"' {
		switch {
		case p.current == '\\':
			p.pauseCapture()
			p.readEscape()
			p.startCapture()
		case p.current < 0x20:
			p.fail(p.expected("valid string character"))
		default:
			p.read()
		}
	}
	s := p.endCapture()
	p.read()
	return s
}

func (p *Parser[A, O]) readEscape() {
	p.read()
	switch p.current {
	case '"', '/', '\\':
		p.captureBuffer = append(p.captureBuffer, p.current)
	case 'b':
		p.captureBuffer = append(p.captureBuffer, '\b')
	case 'f':
		p.captureBuffer = append(p.captureBuffer, '\f')
	case 'n':
		p.captureBuffer = append(p.captureBuffer, '\n')
	case 'r':
		p.captureBuffer = append(p.captureBuffer, '\r')
	case 't':
		p.captureBuffer = append(p.captureBuffer, '\t')
	case 'u':
		var ch uint16
		for range 4 {
			p.read()
			if !p.isHexDigit() {
				p.fail(p.expected("hexadecimal digit"))
			}
			ch = ch<<4 | hexValue(p.current)
		}
		p.captureBuffer = append(p.captureBuffer, ch)
	default:
		p.fail(p.expected("valid escape sequence"))
	}
	p.read()
}

// hexValue returns the value of a hexadecimal digit.
func hexValue(c uint16) uint16 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

func (p *Parser[A, O]) readNumber() {
	p.handler.StartNumber()
	p.startCapture()
	p.readChar('-')
	firstDigit := p.current
	if !p.readDigit() {
		p.fail(p.expected("digit"))
	}
	if firstDigit != '0' {
		for p.readDigit() {
		}
	}
	p.readFraction()
	p.readExponent()
	p.handler.EndNumber(p.endCapture())
}

func (p *Parser[A, O]) readFraction() bool {
	if !p.readChar('.') {
		return false
	}
	if !p.readDigit() {
		p.fail(p.expected("digit"))
	}
	for p.readDigit() {
	}
	return true
}

func (p *Parser[A, O]) readExponent() bool {
	if !p.readChar('e') && !p.readChar('E') {
		return false
	}
	if !p.readChar('+') {
		p.readChar('-')
	}
	if !p.readDigit() {
		p.fail(p.expected("digit"))
	}
	for p.readDigit() {
	}
	return true
}

func (p *Parser[A, O]) readChar(ch uint16) bool {
	if p.current != ch {
		return false
	}
	p.read()
	return true
}

func (p *Parser[A, O]) readDigit() bool {
	if !p.isDigit() {
		return false
	}
	p.read()
	return true
}

func (p *Parser[A, O]) skipWhiteSpace() {
	for p.isWhiteSpace() {
		p.read()
	}
}

// read advances to the next character, refilling the buffer when it is
// exhausted. At the end of the input current becomes endOfText.
func (p *Parser[A, O]) read() {
	if p.index == p.fill {
		if p.captureStart != -1 {
			p.captureBuffer = append(p.captureBuffer, p.buffer[p.captureStart:p.fill]...)
			p.captureStart = 0
		}
		p.bufferOffset += p.fill
		p.fill = p.reader.read(p.buffer)
		p.index = 0
		if p.fill == -1 {
			p.current = endOfText
			p.index++
			return
		}
	}
	if p.current == '\n' {
		p.line++
		p.lineOffset = p.bufferOffset + p.index
	}
	// Past the end of the input, index runs over stale buffer content
	// (fill is -1) until it leaves the array, where Java throws.
	if p.index >= len(p.buffer) {
		p.fail(&IndexOutOfBoundsError{Index: p.index, Length: len(p.buffer)})
	}
	p.current = p.buffer[p.index]
	p.index++
}

func (p *Parser[A, O]) startCapture() {
	p.captureStart = p.index - 1
}

func (p *Parser[A, O]) pauseCapture() {
	end := p.index - 1
	if p.current == endOfText {
		end = p.index
	}
	p.captureBuffer = append(p.captureBuffer, p.buffer[p.captureStart:end]...)
	p.captureStart = -1
}

func (p *Parser[A, O]) endCapture() string {
	start := p.captureStart
	end := p.index - 1
	p.captureStart = -1
	if len(p.captureBuffer) > 0 {
		p.captureBuffer = append(p.captureBuffer, p.buffer[start:end]...)
		captured := string(utf16.Decode(p.captureBuffer))
		p.captureBuffer = p.captureBuffer[:0]
		return captured
	}
	return string(utf16.Decode(p.buffer[start:end]))
}

// location returns the position of the current character.
func (p *Parser[A, O]) location() Location {
	offset := p.bufferOffset + p.index - 1
	column := offset - p.lineOffset + 1
	return Location{Offset: offset, Line: p.line, Column: column}
}

func (p *Parser[A, O]) expected(expected string) *ParseError {
	if p.isEndOfText() {
		return p.error("Unexpected end of input")
	}
	return p.error("Expected " + expected)
}

func (p *Parser[A, O]) error(message string) *ParseError {
	return newParseError(message, p.location())
}

func (p *Parser[A, O]) isWhiteSpace() bool {
	return p.current == ' ' || p.current == '\t' || p.current == '\n' || p.current == '\r'
}

func (p *Parser[A, O]) isDigit() bool {
	return p.current >= '0' && p.current <= '9'
}

func (p *Parser[A, O]) isHexDigit() bool {
	return p.current >= '0' && p.current <= '9' || p.current >= 'a' && p.current <= 'f' ||
		p.current >= 'A' && p.current <= 'F'
}

func (p *Parser[A, O]) isEndOfText() bool {
	return p.current == endOfText
}

// unitReader reads UTF-16 code units from a slice, like Java's StringReader
// reads chars from a string.
type unitReader struct {
	units []uint16
	next  int
}

// read copies the next units into buf and returns their number, or -1 at
// the end of the input.
func (r *unitReader) read(buf []uint16) int {
	if len(buf) == 0 {
		return 0
	}
	if r.next >= len(r.units) {
		return -1
	}
	n := copy(buf, r.units[r.next:])
	r.next += n
	return n
}
