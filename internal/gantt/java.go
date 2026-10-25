// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/uml"
)

// utf16Encode returns the UTF-16 code units of s, Java's chars.
func utf16Encode(s string) []uint16 { return utf16.Encode([]rune(s)) }

// utf16Decode returns the string of UTF-16 code units.
func utf16Decode(u []uint16) string { return string(utf16.Decode(u)) }

// panicNPE panics with Java's NullPointerException, for a value upstream
// leaves null.
func panicNPE() { uml.ThrowNPE("") }
