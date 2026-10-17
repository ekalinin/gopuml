// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/AsciiSnake.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// AsciiSnake is upstream's placeholder for the text counterpart of an
// activity diagram snake: every method throws
// UnsupportedOperationException, as upstream. Nothing uses it.
type AsciiSnake struct{}

// AsciiSnakeCreate panics with UnsupportedOperationException.
func AsciiSnakeCreate() *AsciiSnake {
	throw(unsupportedOperation, "")
	return nil
}

// Move panics with UnsupportedOperationException.
func (s *AsciiSnake) Move(dx, dy int) *AsciiSnake {
	throw(unsupportedOperation, "")
	return nil
}

// Translate panics with UnsupportedOperationException.
func (s *AsciiSnake) Translate(translate klimt.UTranslate) *AsciiSnake {
	throw(unsupportedOperation, "")
	return nil
}

// AddPoint panics with UnsupportedOperationException.
func (s *AsciiSnake) AddPoint(x, y float64) { throw(unsupportedOperation, "") }

// DrawInternal panics with UnsupportedOperationException.
func (s *AsciiSnake) DrawInternal(ug klimt.UGraphic) { throw(unsupportedOperation, "") }

// Merge panics with UnsupportedOperationException.
func (s *AsciiSnake) Merge(other *AsciiSnake, stringBounder font.StringBounder) *AsciiSnake {
	throw(unsupportedOperation, "")
	return nil
}
