// SPDX-License-Identifier: MIT
// Provenance: original

// Package asciiverse is upstream's asciiverse package: the character grid
// (InfinitePlan, InfiniteString) the text formats ATXT and UTXT are drawn
// on, and the blocks drawn on it (AsciiBlock and its decorators, note boxes,
// group frames, else separators and tables).
//
// Characters are Java chars (UTF-16 code units, uint16 here): column
// positions and widths count UTF-16 code units, as upstream does, and the
// grid is written as UTF-8 at the end, with Java's replacement of a lone
// surrogate by '?'.
//
// Upstream's Display implements AsciiBlock; Display lives in internal/klimt,
// which cannot import this package, so DisplayBlock wraps a Display into
// its AsciiBlock.
package asciiverse

import "github.com/ekalinin/plantuml-in-go/internal/klimt/color"

// Java exception classes thrown by this package.
const (
	illegalArgument      = "java.lang.IllegalArgumentException"
	unsupportedOperation = "java.lang.UnsupportedOperationException"
	nullPointer          = "java.lang.NullPointerException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }
