// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/InfiniteString.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import "unicode/utf16"

// InfiniteString is one line of the grid, growing to the right (positions
// >= 0, in positiveChars) and to the left (positions < 0, in negativeChars,
// where index i holds position -i-1). Unwritten positions read as spaces.
type InfiniteString struct {
	positiveChars []uint16
	negativeChars []uint16
}

func (s *InfiniteString) ensureLength(length int) {
	for len(s.positiveChars) < length {
		s.positiveChars = append(s.positiveChars, ' ')
	}
}

func (s *InfiniteString) ensureNegativeLength(length int) {
	for len(s.negativeChars) < length {
		s.negativeChars = append(s.negativeChars, ' ')
	}
}

// SetCharAt writes a character at a position.
func (s *InfiniteString) SetCharAt(position int, ch uint16) {
	if position >= 0 {
		s.ensureLength(position + 1)
		s.positiveChars[position] = ch
	} else {
		negIndex := -position - 1
		s.ensureNegativeLength(negIndex + 1)
		s.negativeChars[negIndex] = ch
	}
}

// SetStringAt writes the UTF-16 code units of str from a position on.
func (s *InfiniteString) SetStringAt(position int, str string) {
	s.setCharsAt(position, utf16.Encode([]rune(str)))
}

func (s *InfiniteString) setCharsAt(position int, chars []uint16) {
	for i, ch := range chars {
		s.SetCharAt(position+i, ch)
	}
}

// CharAt returns the character at a position, a space if none was written.
func (s *InfiniteString) CharAt(position int) uint16 {
	if position >= 0 {
		if position >= len(s.positiveChars) {
			return ' '
		}
		return s.positiveChars[position]
	}
	negIndex := -position - 1
	if negIndex >= len(s.negativeChars) {
		return ' '
	}
	return s.negativeChars[negIndex]
}

// LeftmostPosition returns the leftmost position ever written, 0 or less.
func (s *InfiniteString) LeftmostPosition() int { return -len(s.negativeChars) }

// Chars returns the line from startingPosition (inclusive, possibly
// negative) to its right end, as UTF-16 code units (upstream's
// getString).
func (s *InfiniteString) Chars(startingPosition int) []uint16 {
	if startingPosition >= 0 {
		if startingPosition > len(s.positiveChars) {
			// StringBuilder.substring throws for a start past the end
			throw("java.lang.StringIndexOutOfBoundsException", "")
		}
		return append([]uint16(nil), s.positiveChars[startingPosition:]...)
	}
	negLength := -startingPosition
	result := make([]uint16, 0, negLength+len(s.positiveChars))
	// From the leftmost requested position up to -1, in reading order
	for negIndex := negLength - 1; negIndex >= 0; negIndex-- {
		if negIndex < len(s.negativeChars) {
			result = append(result, s.negativeChars[negIndex])
		} else {
			result = append(result, ' ')
		}
	}
	return append(result, s.positiveChars...)
}

// String returns the line from startingPosition as a string; a lone
// surrogate becomes U+FFFD (ExportTxt writes '?' for it, as Java).
func (s *InfiniteString) String(startingPosition int) string {
	return string(utf16.Decode(s.Chars(startingPosition)))
}
