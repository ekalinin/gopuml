// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/json/PrettyPrint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import (
	"errors"
	"strings"
)

// PrettyPrint is a [WriterConfig] that adds whitespace to the JSON text:
// either every value on a separate line with a given indentation, or a
// single line with a space after each separator.
type PrettyPrint struct {
	indentChars string
	// newLines is false for the single-line mode, where upstream's
	// indentChars is null.
	newLines bool
}

func newPrettyPrint(indentChars string, newLines bool) *PrettyPrint {
	return &PrettyPrint{indentChars: indentChars, newLines: newLines}
}

// SingleLine does not break lines, but still inserts a space after the
// separators between values and between a name and its value.
func SingleLine() *PrettyPrint {
	return newPrettyPrint("", false)
}

// IndentWithSpaces prints every value on a separate line, indented with the
// given number of spaces per level. It returns an error if number is
// negative.
func IndentWithSpaces(number int) (*PrettyPrint, error) {
	if number < 0 {
		return nil, errors.New("number is negative")
	}
	return newPrettyPrint(strings.Repeat(" ", number), true), nil
}

// IndentWithTabs prints every value on a separate line, indented with one
// tab per level.
func IndentWithTabs() *PrettyPrint {
	return newPrettyPrint("\t", true)
}

func (p *PrettyPrint) createWriter(buffer *writingBuffer) jsonWriter {
	return &prettyPrintWriter{minimalWriter: minimalWriter{writer: buffer}, indentChars: p.indentChars, newLines: p.newLines}
}

// prettyPrintWriter overrides the structural events of minimalWriter.
type prettyPrintWriter struct {
	minimalWriter
	indentChars string
	newLines    bool
	indent      int
}

func (w *prettyPrintWriter) writeArrayOpen() {
	w.indent++
	w.writer.writeByte('[')
	w.writeNewLine()
}

func (w *prettyPrintWriter) writeArrayClose() {
	w.indent--
	w.writeNewLine()
	w.writer.writeByte(']')
}

func (w *prettyPrintWriter) writeArraySeparator() {
	w.writer.writeByte(',')
	if !w.writeNewLine() {
		w.writer.writeByte(' ')
	}
}

func (w *prettyPrintWriter) writeObjectOpen() {
	w.indent++
	w.writer.writeByte('{')
	w.writeNewLine()
}

func (w *prettyPrintWriter) writeObjectClose() {
	w.indent--
	w.writeNewLine()
	w.writer.writeByte('}')
}

func (w *prettyPrintWriter) writeMemberSeparator() {
	w.writer.writeByte(':')
	w.writer.writeByte(' ')
}

func (w *prettyPrintWriter) writeObjectSeparator() {
	w.writer.writeByte(',')
	if !w.writeNewLine() {
		w.writer.writeByte(' ')
	}
}

// writeNewLine starts a new line indented to the current level and reports
// whether it did, which it does not in the single-line mode.
func (w *prettyPrintWriter) writeNewLine() bool {
	if !w.newLines {
		return false
	}
	w.writer.writeByte('\n')
	for range w.indent {
		w.writer.writeString(w.indentChars)
	}
	return true
}
