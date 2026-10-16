// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// This file holds the placeholders of the creole classes that the drawing
// model refers to, ported in the creole step: klimt/creole/SheetBuilder,
// Sheet, SheetBlock1, SheetBlock2 and klimt/creole/legacy/CreoleParser.
// Using them panics with the name of the missing class; the creole step
// replaces this file (see HANDOFF.md).

// SheetBuilder builds the sheet (the parsed creole) of a display. It is
// upstream's klimt/creole/SheetBuilder.
type SheetBuilder interface {
	CreateSheet(display *Display) *Sheet
}

// Sheet is a placeholder for upstream's klimt/creole/Sheet.
type Sheet struct{}

// creoleParserStub is a placeholder for upstream's
// klimt/creole/legacy/CreoleParser.
type creoleParserStub struct{}

func newCreoleParser(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	skinParam ISkinSimple, creoleMode CreoleMode, stereo *FontConfiguration) SheetBuilder {
	return creoleParserStub{}
}

// CreateSheet panics: creole parsing is not ported yet.
func (creoleParserStub) CreateSheet(*Display) *Sheet {
	notPorted("klimt/creole/legacy/CreoleParser", "creole parsing of a Display")
	return nil
}

// SheetBlock1 is a placeholder for upstream's klimt/creole/SheetBlock1.
type SheetBlock1 struct {
	TextBlockDefaults
}

// NewSheetBlock1 panics: creole rendering is not ported yet.
func NewSheetBlock1(sheet *Sheet, maxWidth LineBreakStrategy, padding ClockwiseTopRightBottomLeft,
	marginX1, marginX2 float64) *SheetBlock1 {
	notPorted("klimt/creole/SheetBlock1", "creole rendering of a Display")
	return nil
}

// CalculateDimension is not reachable: construction panics.
func (*SheetBlock1) CalculateDimension(font.StringBounder) geom.XDimension2D {
	notPorted("klimt/creole/SheetBlock1", "creole rendering")
	return geom.XDimension2D{}
}

// DrawU is not reachable: construction panics.
func (*SheetBlock1) DrawU(UGraphic) { notPorted("klimt/creole/SheetBlock1", "creole rendering") }

// StartingX is not reachable: construction panics.
func (*SheetBlock1) StartingX(font.StringBounder, float64) float64 {
	notPorted("klimt/creole/SheetBlock1", "creole rendering")
	return 0
}

// EndingX is not reachable: construction panics.
func (*SheetBlock1) EndingX(font.StringBounder, float64) float64 {
	notPorted("klimt/creole/SheetBlock1", "creole rendering")
	return 0
}

// SheetBlock2 is a placeholder for upstream's klimt/creole/SheetBlock2.
type SheetBlock2 struct {
	TextBlockDefaults
}

// NewSheetBlock2 panics: creole rendering is not ported yet.
func NewSheetBlock2(block *SheetBlock1, stencil Stencil, defaultStroke UStroke) *SheetBlock2 {
	notPorted("klimt/creole/SheetBlock2", "creole rendering of a Display")
	return nil
}

// EnlargeMe is not reachable: construction panics.
func (*SheetBlock2) EnlargeMe(left, right float64) TextBlock {
	notPorted("klimt/creole/SheetBlock2", "creole rendering")
	return nil
}

// CalculateDimension is not reachable: construction panics.
func (*SheetBlock2) CalculateDimension(font.StringBounder) geom.XDimension2D {
	notPorted("klimt/creole/SheetBlock2", "creole rendering")
	return geom.XDimension2D{}
}

// DrawU is not reachable: construction panics.
func (*SheetBlock2) DrawU(UGraphic) { notPorted("klimt/creole/SheetBlock2", "creole rendering") }
