// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/DisplayPositioned.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// VerticalAlignmentNull stands for upstream's null vertical alignment (the
// header, footer and main frame have none).
const VerticalAlignmentNull geom.VerticalAlignment = -1

// DisplayPositioned is a text with its alignment and source line, such as
// the title or the legend of a diagram. The zero value is not valid; use
// DisplayPositionedNone or DisplayPositionedSingle.
type DisplayPositioned struct {
	display *klimt.Display
	// horizontal is klimt.HorizontalAlignmentNull for null.
	horizontal geom.HorizontalAlignment
	// vertical is VerticalAlignmentNull for null.
	vertical geom.VerticalAlignment
	location *text.LineLocation
}

// DisplayPositionedSingle returns a positioned text; location may be nil.
func DisplayPositionedSingle(location *text.LineLocation, display *klimt.Display,
	horizontal geom.HorizontalAlignment, vertical geom.VerticalAlignment) DisplayPositioned {
	return DisplayPositioned{display: display, horizontal: horizontal, vertical: vertical, location: location}
}

// DisplayPositionedNone returns the absent text (the null display).
func DisplayPositionedNone(horizontal geom.HorizontalAlignment, vertical geom.VerticalAlignment) DisplayPositioned {
	return DisplayPositioned{display: klimt.DisplayNULL(), horizontal: horizontal, vertical: vertical}
}

// WithPage returns the text with the page number variables replaced.
func (d DisplayPositioned) WithPage(page, lastpage int) DisplayPositioned {
	d.display = d.display.WithPage(page, lastpage)
	return d
}

// WithDisplay returns the positioned text with another text.
func (d DisplayPositioned) WithDisplay(display *klimt.Display) DisplayPositioned {
	d.display = display
	return d
}

// WithHorizontalAlignment returns the positioned text with another
// horizontal alignment.
func (d DisplayPositioned) WithHorizontalAlignment(horizontal geom.HorizontalAlignment) DisplayPositioned {
	d.horizontal = horizontal
	return d
}

// WithLocation returns the positioned text with another source line.
func (d DisplayPositioned) WithLocation(location *text.LineLocation) DisplayPositioned {
	d.location = location
	return d
}

// Display returns the text.
func (d DisplayPositioned) Display() *klimt.Display { return d.display }

// HorizontalAlignment returns the horizontal alignment,
// klimt.HorizontalAlignmentNull for none.
func (d DisplayPositioned) HorizontalAlignment() geom.HorizontalAlignment { return d.horizontal }

// VerticalAlignment returns the vertical alignment, VerticalAlignmentNull
// for none.
func (d DisplayPositioned) VerticalAlignment() geom.VerticalAlignment { return d.vertical }

// IsNull reports whether the text is absent.
func (d DisplayPositioned) IsNull() bool { return klimt.DisplayIsNull(d.display) }

// HasURL reports whether the text has a link.
func (d DisplayPositioned) HasURL() bool { return d.display.HasURL() }

// LineLocation returns the source line of the text, nil when unknown.
func (d DisplayPositioned) LineLocation() *text.LineLocation { return d.location }
