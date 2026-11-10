// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/YamlParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// Parse reads the YAML lines into a Monomorph. YAML outside upstream's
// subset gives the *Error of the exception that upstream throws.
func Parse(lines []string) (result *Monomorph, err error) {
	defer catch(&err)
	yamlBuilder := newYamlBuilder()

	for p := newPeeker(lines); ; p.jump() {
		line, ok := p.peek(0)
		if !ok {
			break
		}
		yamlLine := BuildYamlLine(line)
		if yamlLine.Type() == EmptyLine {
			continue
		}

		if yamlLine.Type() == NoKeyOnlyText {
			throw(YamlSyntaxException, "YamlLineType.NO_KEY_ONLY_TEXT")
		}

		if yamlLine.Type() == PlainDash {
			yamlBuilder.onListItemPlainDash()
			continue
		}

		yamlBuilder.adjustIndentation(yamlLine.Indent())

		switch {
		case yamlLine.IsListItem():
			switch yamlLine.Type() {
			case KeyOnly:
				yamlBuilder.onListItemOnlyKey(yamlLine.Key())
			case PlainElementList:
				yamlBuilder.onListItemOnlyValue(yamlLine.Value())
			case KeyAndValue:
				yamlBuilder.onListItemKeyAndValue(yamlLine.Key(), yamlLine.Value())
			case KeyAndFlowSequence:
				yamlBuilder.onListItemKeyAndFlowSequence(yamlLine.Key(), yamlLine.Values())
			default:
				throw(UnsupportedOperationException, "wip3 "+yamlLine.String())
			}
		case yamlLine.Type() == KeyOnly:
			next, hasNext := peekNext(p)
			switch {
			case !hasNext || next.Indent() <= yamlLine.Indent():
				yamlBuilder.onKeyAndValue(yamlLine.Key(), "")
			case next.Type() == NoKeyOnlyText:
				yamlBuilder.onKeyAndValue(yamlLine.Key(), peekNextOnlyText(p))
			default:
				yamlBuilder.onOnlyKey(yamlLine.Key())
			}
		case yamlLine.Type() == KeyAndValue:
			yamlBuilder.onKeyAndValue(yamlLine.Key(), yamlLine.Value())
		case yamlLine.Type() == KeyAndBlockStyle:
			yamlBuilder.onKeyAndValue(yamlLine.Key(), blockStyleString(yamlLine.Indent(), p))
		case yamlLine.Type() == KeyAndFlowSequence:
			yamlBuilder.onKeyAndFlowSequence(yamlLine.Key(), yamlLine.Values())
		default:
			throw(UnsupportedOperationException, "wip4 "+yamlLine.String())
		}
	}

	return yamlBuilder.result(), nil
}

// peekNextOnlyText reads the lines of text that follow, joined with spaces,
// skipping the empty lines.
func peekNextOnlyText(p *peeker[string]) string {
	var result strings.Builder
	for {
		peek, ok := p.peek(1)
		if !ok {
			return result.String()
		}
		next := BuildYamlLine(peek)
		if next.Type() == EmptyLine {
			p.jump()
			continue
		}
		if next.Type() == NoKeyOnlyText {
			if result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString(next.Value())
			p.jump()
			continue
		}
		return result.String()
	}
}

// peekNext returns the next line that is not empty; ok is false when there
// is none.
func peekNext(p *peeker[string]) (next *YamlLine, ok bool) {
	for i := 1; ; i++ {
		peek, ok := p.peek(i)
		if !ok {
			return nil, false
		}
		if next := BuildYamlLine(peek); next.Type() != EmptyLine {
			return next, true
		}
	}
}

// blockStyleString reads the lines of a "|" value: the lines of text, the
// empty lines and the lines indented more than the key, each trimmed and
// followed by a line break.
func blockStyleString(indent int, p *peeker[string]) string {
	var result strings.Builder
	for {
		line, ok := p.peek(1)
		if !ok {
			return result.String()
		}
		yamlLine := BuildYamlLine(line)
		if yamlLine.Type() != NoKeyOnlyText && yamlLine.Type() != EmptyLine && yamlLine.Indent() <= indent {
			return result.String()
		}

		result.WriteString(cleanBlockStyle(line))
		result.WriteString("\n")
		p.jump()
	}
}

// cleanBlockStyle is upstream's unfinished cleaning of a block line: it
// trims it.
func cleanBlockStyle(s string) string { return text.JavaTrim(s) }
