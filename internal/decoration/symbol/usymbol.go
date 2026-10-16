// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbol.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbols.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// USymbol is a shape drawn around a text.
type USymbol interface {
	// SNames returns the style names of the symbol.
	SNames() []klimt.SName
	// AsSmall returns the symbol around a label and a stereotype, sized to
	// fit them; name is the title of the symbols that show one.
	AsSmall(name, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
		stereoAlignment geom.HorizontalAlignment) klimt.TextBlock
	// AsBig returns the symbol of a given size with a title and a
	// stereotype at its top, as a container.
	AsBig(label klimt.TextBlock, labelAlignment geom.HorizontalAlignment, stereotype klimt.TextBlock,
		width, height float64, symbolContext *klimt.Fashion, stereoAlignment geom.HorizontalAlignment) klimt.TextBlock
	// SuppHeightBecauseOfShape returns the extra height of the shape.
	SuppHeightBecauseOfShape() int
	// SuppWidthBecauseOfShape returns the extra width of the shape.
	SuppWidthBecauseOfShape() int
}

// usymbolBase holds upstream's default methods of USymbol.
type usymbolBase struct{}

// SuppHeightBecauseOfShape returns 0, the default.
func (usymbolBase) SuppHeightBecauseOfShape() int { return 0 }

// SuppWidthBecauseOfShape returns 0, the default.
func (usymbolBase) SuppWidthBecauseOfShape() int { return 0 }

// margin is upstream's USymbol.Margin: the space around the text of a
// symbol.
type margin struct {
	x1, x2, y1, y2 float64
}

func newMargin(x1, x2, y1, y2 float64) margin { return margin{x1: x1, x2: x2, y1: y1, y2: y2} }

func (m margin) addDimension(dim geom.XDimension2D) geom.XDimension2D {
	return geom.NewXDimension2D(dim.Width()+m.x1+m.x2, dim.Height()+m.y1+m.y2)
}

// The symbols of upstream's USymbols that are ported. They are stateless
// values, shared by every render.
var (
	// Database is upstream's USymbols.DATABASE.
	Database USymbol = usymbolDatabase{}
	// Queue is upstream's USymbols.QUEUE.
	Queue USymbol = usymbolQueue{}
)
