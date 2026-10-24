// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/BlocLines.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/version/IteratorCounter2Impl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/text"
)

// BlocLines is an immutable list of source lines given to a command.
type BlocLines struct {
	lines []text.StringLocated
}

// NewBlocLines returns lines from a slice, which it does not copy.
func NewBlocLines(lines []text.StringLocated) BlocLines { return BlocLines{lines: lines} }

// Single returns one line.
func Single(l text.StringLocated) BlocLines { return BlocLines{lines: []text.StringLocated{l}} }

// SingleString returns one line without a location.
func SingleString(s string) BlocLines { return Single(text.NewStringLocated(s, nil)) }

// Size is the number of lines.
func (b BlocLines) Size() int { return len(b.lines) }

// At returns line i.
func (b BlocLines) At(i int) text.StringLocated { return b.lines[i] }

// First returns the first line; ok is false when there is none.
func (b BlocLines) First() (text.StringLocated, bool) {
	if len(b.lines) == 0 {
		return text.StringLocated{}, false
	}
	return b.lines[0], true
}

// Last returns the last line.
func (b BlocLines) Last() text.StringLocated { return b.lines[len(b.lines)-1] }

// Lines returns the lines; the slice must not be modified.
func (b BlocLines) Lines() []text.StringLocated { return b.lines }

// Strings returns the text of every line.
func (b BlocLines) Strings() []string {
	out := make([]string, len(b.lines))
	for i, l := range b.lines {
		out[i] = l.S
	}
	return out
}

// Add returns the lines with l appended.
func (b BlocLines) Add(l text.StringLocated) BlocLines {
	return BlocLines{lines: append(slices.Clip(b.lines), l)}
}

// Trim returns the lines with characters <= ' ' removed at both ends.
func (b BlocLines) Trim() BlocLines {
	out := make([]text.StringLocated, len(b.lines))
	for i, l := range b.lines {
		out[i] = l.Trimmed()
	}
	return BlocLines{lines: out}
}

// RemoveEmptyLines drops lines without characters.
func (b BlocLines) RemoveEmptyLines() BlocLines {
	var out []text.StringLocated
	for _, l := range b.lines {
		if l.S != "" {
			out = append(out, l)
		}
	}
	return BlocLines{lines: out}
}

// RemoveEmptyColumns removes leading space and tab columns shared by every
// non-empty line.
func (b BlocLines) RemoveEmptyColumns() BlocLines {
	if !firstColumnRemovable(b.lines) {
		return b
	}
	out := slices.Clone(b.lines)
	for firstColumnRemovable(out) {
		for i, l := range out {
			if l.S != "" {
				out[i] = l.Substring(1, len(l.S))
			}
		}
	}
	return BlocLines{lines: out}
}

func firstColumnRemovable(lines []text.StringLocated) bool {
	allEmpty := true
	for _, l := range lines {
		if l.S == "" {
			continue
		}
		allEmpty = false
		if c := l.S[0]; c != ' ' && c != '\t' {
			return false
		}
	}
	return !allEmpty
}

// RemoveStartingAndEnding replaces the first line with data and removes
// removeAtEnd bytes from the end of the last line.
func (b BlocLines) RemoveStartingAndEnding(data string, removeAtEnd int) BlocLines {
	if len(b.lines) == 0 {
		return b
	}
	out := slices.Clone(b.lines)
	out[0] = text.NewStringLocated(data, nil)
	if removeAtEnd > 0 {
		n := len(out) - 1
		out[n] = out[n].Substring(0, len(out[n].S)-removeAtEnd)
	}
	return BlocLines{lines: out}
}

// OverrideLastLine replaces the text of the last line, keeping its location.
func (b BlocLines) OverrideLastLine(last string) BlocLines {
	if len(b.lines) == 0 {
		return b
	}
	out := slices.Clone(b.lines)
	n := len(out) - 1
	out[n] = text.NewStringLocated(last, out[n].Location)
	return BlocLines{lines: out}
}

// iterator walks source lines and records the lines it has consumed, which
// become the trace of an error.
type iterator struct {
	data  []text.StringLocated
	pos   int
	trace []text.StringLocated
}

func newIterator(data []text.StringLocated) *iterator { return &iterator{data: data} }

func (it *iterator) hasNext() bool { return it.pos < len(it.data) }

func (it *iterator) peek() text.StringLocated { return it.data[it.pos] }

func (it *iterator) next() text.StringLocated {
	l := it.data[it.pos]
	it.pos++
	it.trace = append(it.trace, l)
	return l
}

func (it *iterator) clone() *iterator {
	return &iterator{data: it.data, pos: it.pos, trace: slices.Clone(it.trace)}
}

func (it *iterator) copyStateFrom(other *iterator) { *it = *other }
