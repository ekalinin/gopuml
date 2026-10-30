// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/DataSource.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/DataSourceImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/Terminated.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/Terminator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// terminator is what follows a token: a new line or a new column.
type terminator int

const (
	newLine terminator = iota
	newCol
)

func (t terminator) String() string {
	if t == newCol {
		return "NEWCOL"
	}
	return "NEWLINE"
}

// terminated is an element with the terminator that follows it.
type terminated[O any] struct {
	element    O
	terminator terminator
}

// dataSource is upstream's DataSource: the tokens of a salt diagram, read
// one after the other with a lookahead.
type dataSource interface {
	peek(i int) terminated[string]
	hasNext() bool
	next() terminated[string]
}

// structuredBlockStartPattern is DataSourceImpl.STRUCTURED_BLOCK_START_PATTERN,
// a Pattern2, so case-insensitive.
var structuredBlockStartPattern = regex.MustCompile(`\{(?:[-+^#!*/]|S-|SI|S)?`)

// The placeholders of the escaped "\|", "\{" and "\}", from the Unicode
// Private Use Area, as upstream.
const (
	escapedPipe  = "\uE000"
	escapedOpen  = "\uE001"
	escapedClose = "\uE002"
)

// dataSourceImpl is upstream's DataSourceImpl.
type dataSourceImpl struct {
	i    int
	data []terminated[string]
}

// newDataSourceImpl cuts the lines into tokens: "|" and "}" end a token, and
// the start of a structured block ("{", "{#", "{SI", ...) is a token of its
// own.
func newDataSourceImpl(lines []string) *dataSourceImpl {
	d := &dataSourceImpl{}
	for _, rawLine := range lines {
		s := strings.ReplaceAll(rawLine, `\|`, escapedPipe)
		s = strings.ReplaceAll(s, `\{`, escapedOpen)
		s = strings.ReplaceAll(s, `\}`, escapedClose)
		tokens := klimt.JavaTokenizeWithDelims(s, "|}")
		for n, raw := range tokens {
			token := text.Trin(raw)
			if token == "|" {
				continue
			}
			term := newLine
			if n+1 < len(tokens) {
				term = newCol
			}
			matches := structuredBlockStartPattern.FindAllStringIndex(token, -1)
			if len(matches) == 0 {
				d.addInternal(token, term)
				continue
			}
			lastStart := 0
			end := 0
			for _, m := range matches {
				start := m[0]
				if start > lastStart {
					d.addInternal(token[lastStart:start], newCol)
				}
				end = m[1]
				t := newCol
				if end == len(token) {
					t = term
				}
				d.addInternal(token[start:end], t)
				lastStart = end
			}
			if end < len(token) {
				d.addInternal(token[end:], term)
			}
		}
	}
	return d
}

func (d *dataSourceImpl) addInternal(s string, t terminator) {
	s = text.Trin(s)
	s = strings.ReplaceAll(s, escapedPipe, "|")
	s = strings.ReplaceAll(s, escapedOpen, "{")
	s = strings.ReplaceAll(s, escapedClose, "}")
	if len(s) > 0 {
		d.data = append(d.data, terminated[string]{element: s, terminator: t})
	}
}

// peek returns the token nb places ahead. Past the end it panics as Java's
// ArrayList.get.
func (d *dataSourceImpl) peek(nb int) terminated[string] {
	if d.i+nb >= len(d.data) {
		uml.ThrowIndexOutOfBounds(d.i+nb, len(d.data))
	}
	return d.data[d.i+nb]
}

func (d *dataSourceImpl) hasNext() bool { return d.i < len(d.data) }

func (d *dataSourceImpl) next() terminated[string] {
	if d.i >= len(d.data) {
		uml.ThrowIndexOutOfBounds(d.i, len(d.data))
	}
	result := d.data[d.i]
	d.i++
	return result
}
