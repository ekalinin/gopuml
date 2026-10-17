// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Display.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import "github.com/ekalinin/gopuml/internal/klimt"

// DisplayBlock returns a Display as an AsciiBlock, upstream's
// Display.asciiDimension and Display.asciiDraw: one row per line, drawn as
// is (creole markup included), left aligned.
func DisplayBlock(display *klimt.Display) AsciiBlock { return displayBlock{display: display} }

type displayBlock struct {
	display *klimt.Display
}

// AsciiDimension returns the longest line (in UTF-16 code units) by the
// number of lines.
func (b displayBlock) AsciiDimension() ADimension2D {
	return NewADimension2D(DisplayContentWidth(b.display), b.display.Size())
}

// AsciiDraw draws each line on its row.
func (b displayBlock) AsciiDraw(plan *InfinitePlan) {
	for i := 0; i < b.display.Size(); i++ {
		plan.Move(0, i).DrawString(displayLine(b.display, i))
	}
}

// DisplayContentWidth is upstream's Display.contentWidth: the length of the
// longest line in UTF-16 code units.
func DisplayContentWidth(display *klimt.Display) int {
	width := 0
	for i := 0; i < display.Size(); i++ {
		if l := len(javaChars(displayLine(display, i))); l > width {
			width = l
		}
	}
	return width
}

// displayLine is get(i).toString(); a null line panics with
// NullPointerException, as upstream.
func displayLine(display *klimt.Display, i int) string {
	cs := display.Get(i)
	if cs == nil {
		throw(nullPointer, "")
	}
	return cs.String()
}
