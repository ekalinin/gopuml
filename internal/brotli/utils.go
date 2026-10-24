// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Utils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

import "io"

// maxConsecutiveEmptyReads is the number of reads of the source that return
// no byte and no error after which readInput gives up, as bufio does.
const maxConsecutiveEmptyReads = 100

// fillBytesWithZeroes fills dest[start:end] with zeroes.
func fillBytesWithZeroes(dest []byte, start, end int) {
	clear(dest[start:end])
}

// fillIntsWithZeroes fills dest[start:end] with zeroes.
func fillIntsWithZeroes(dest []int, start, end int) {
	clear(dest[start:end])
}

// copyBytesWithin copies bytes[start:end] to bytes[target:], as
// System.arraycopy does when the ranges overlap.
func copyBytesWithin(bytes []byte, target, start, end int) {
	copy(bytes[target:], bytes[start:end])
}

// readInput reads at most length bytes of the source of s into
// dst[offset:] and returns their number, or -1 at the end of the source, as
// Java's InputStream.read does. It reads s.input instead of taking the
// stream, so that it can keep the error that comes with the last bytes of a
// read (s.inputErr), and it retries a read that returns nothing.
func readInput(s *state, dst []byte, offset, length int) int {
	for range maxConsecutiveEmptyReads {
		if s.inputErr != nil {
			if s.inputErr == io.EOF {
				return -1
			}
			throwCause("Failed to read input", s.inputErr)
		}
		n, err := s.input.Read(dst[offset : offset+length])
		s.inputErr = err
		if n > 0 {
			return n
		}
	}
	throwCause("Failed to read input", io.ErrNoProgress)
	return -1
}
