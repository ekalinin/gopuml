// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/YamlLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// YamlLine is a YAML line: its kind, its indentation, and its key and value.
type YamlLine struct {
	indent   int
	key      string
	value    string
	values   []string
	listItem bool
	typ      YamlLineType
}

// BuildYamlLine reads a line. A tab counts as four spaces. The indentation
// of a list item includes its "- ".
func BuildYamlLine(line string) *YamlLine {
	count := 0
	line = strings.ReplaceAll(line, "\t", "    ")
	for count < len(line) && line[count] == ' ' {
		count++
	}

	trimmedLine := removeYamlComment(text.JavaTrim(line[count:]))

	if trimmedLine == "" {
		return &YamlLine{typ: EmptyLine}
	}

	if trimmedLine == "-" {
		return &YamlLine{typ: PlainDash, indent: count + 1, listItem: true}
	}

	listItem := strings.HasPrefix(trimmedLine, "- ")

	if listItem {
		count += 2
		trimmedLine = trimmedLine[2:]
	}

	colonIndex := findColonSeparator(trimmedLine)
	if colonIndex == -1 {
		if listItem {
			return &YamlLine{typ: PlainElementList, indent: count, value: unquote(trimmedLine), listItem: listItem}
		}
		return &YamlLine{typ: NoKeyOnlyText, indent: count, value: unquote(trimmedLine)}
	}

	rawKey := text.JavaTrim(trimmedLine[:colonIndex])
	rawValue := text.JavaTrim(trimmedLine[colonIndex+1:])

	typ := KeyAndValue

	switch {
	case rawValue == "":
		typ = KeyOnly
	case rawValue == "|":
		typ = KeyAndBlockStyle
	case rawValue == ">":
		typ = KeyAndFoldedStyle
	case strings.HasPrefix(rawValue, "[") && strings.HasSuffix(rawValue, "]"):
		return &YamlLine{typ: KeyAndFlowSequence, indent: count, key: unquote(rawKey),
			values: toList(rawValue[1 : len(rawValue)-1]), listItem: listItem}
	}

	return &YamlLine{typ: typ, indent: count, key: unquote(rawKey), value: unquote(rawValue), listItem: listItem}
}

// findColonSeparator returns the index of the colon that separates the key
// from the value, skipping the colons inside quoted strings, or -1.
func findColonSeparator(line string) int {
	var inQuote byte
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inQuote != 0:
			if c == inQuote {
				inQuote = 0
			}
		case c == '"' || c == '\'':
			inQuote = c
		case c == ':':
			return i
		}
	}
	return -1
}

// toList splits the inside of a flow sequence at its commas: a field that
// starts with a quote is kept as it is up to its closing quote, the others
// are trimmed, and a backslash escapes the next character.
func toList(rawValue string) []string {
	var result []string

	var current strings.Builder

	// Zero if we are not in a quoted string, otherwise the quote character.
	var inQuotedString byte

	// Whether the current field started with a quote.
	fieldStartWithQuote := false

	for i := 0; i < len(rawValue); i++ {
		c := rawValue[i]

		if inQuotedString != 0 {
			switch {
			case c == '\\':
				if i+1 < len(rawValue) {
					current.WriteByte(rawValue[i+1])
					i++
				}
			case c == inQuotedString:
				inQuotedString = 0
			default:
				current.WriteByte(c)
			}
			continue
		}
		switch {
		case !fieldStartWithQuote && text.JavaTrim(current.String()) == "" && (c == '\'' || c == '"'):
			// A field of spaces followed by a quote starts with the quote.
			inQuotedString = c
			fieldStartWithQuote = true
			current.Reset()
		case c == ',':
			result = append(result, field(current.String(), fieldStartWithQuote))
			current.Reset()
			fieldStartWithQuote = false
		case c == '\\':
			if i+1 < len(rawValue) {
				current.WriteByte(rawValue[i+1])
				i++
			}
		default:
			current.WriteByte(c)
		}
	}
	// The last field, even when it is empty.
	return append(result, field(current.String(), fieldStartWithQuote))
}

// field is a field of a flow sequence: as it is when it was quoted, trimmed
// otherwise.
func field(s string, quoted bool) string {
	if quoted {
		return s
	}
	return text.JavaTrim(s)
}

func unquote(str string) string {
	if len(str) < 2 {
		return str
	}
	first := str[0]
	last := str[len(str)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return str[1 : len(str)-1]
	}
	return str
}

// Indent returns the indentation of the line.
func (l *YamlLine) Indent() int { return l.indent }

// removeYamlComment removes a comment: the whole line when it starts with
// "#", otherwise from a " #" outside quotes.
func removeYamlComment(s string) string {
	if s == "" {
		return s
	}

	var inQuoteChar byte

	if s[0] == '#' {
		return ""
	}

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '\'' || c == '"' {
			if inQuoteChar == 0 {
				inQuoteChar = c
			} else if c == inQuoteChar {
				inQuoteChar = 0
			}
		}

		if inQuoteChar == 0 && i < len(s)-1 && c == ' ' && s[i+1] == '#' {
			return s[:i]
		}
	}

	return s
}

// Key returns the key of the line.
func (l *YamlLine) Key() string { return l.key }

// Value returns the value of the line. Upstream throws an
// IllegalStateException for the kinds of lines without a value.
func (l *YamlLine) Value() string {
	switch l.typ {
	case KeyAndValue, KeyAndFlowSequence, PlainElementList, NoKeyOnlyText:
		return l.value
	}
	throw(IllegalStateException, l.typ.String())
	return ""
}

// IsListItem reports whether the line is a list item.
func (l *YamlLine) IsListItem() bool { return l.listItem }

// Type returns the kind of the line.
func (l *YamlLine) Type() YamlLineType { return l.typ }

// Values returns the values of a flow sequence. Upstream throws an
// IllegalStateException for the other kinds of lines.
func (l *YamlLine) Values() []string {
	if l.typ != KeyAndFlowSequence {
		throw(IllegalStateException, l.typ.String())
	}
	return l.values
}

// String is upstream's toString, which names a missing key or value null.
func (l *YamlLine) String() string {
	return "YamlLine(" + l.typ.String() + " indent=" + strconv.Itoa(l.indent) + ", key=" + orNull(l.key) +
		", value=" + orNull(l.value) + ")"
}

func orNull(s string) string {
	if s == "" {
		return "null"
	}
	return s
}
