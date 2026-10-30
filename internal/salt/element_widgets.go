// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementButton.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementTextField.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementRadioCheckbox.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementDroplist.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// elementText is plain text. A text "." is an empty cell: it takes the
// size of the dot and draws nothing.
type elementText struct {
	abstractElement
	block klimt.TextBlock
	show  bool
	text  string
}

func newElementText(lines []string, f font.UFont, spriteContainer klimt.ISkinSimple) *elementText {
	return &elementText{
		block: klimt.DisplayCreate(lines...).Create(blackBlueTrue(f), geom.HorizontalAlignmentLeft, spriteContainer),
		show:  lines[0] != ".",
		text:  lines[0],
	}
}

func (e *elementText) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	return e.block.CalculateDimension(stringBounder)
}

func (e *elementText) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	if e.show {
		e.block.DrawU(ug)
	}
}

// elementButton is "[text]": a rounded box filled with #EEE.
type elementButton struct {
	abstractElementText
}

const (
	buttonStroke  = 2.5
	buttonMarginX = 2
	buttonMarginY = 2
)

func newElementButton(s string, f font.UFont, spriteContainer klimt.ISkinSimple) *elementButton {
	return &elementButton{newAbstractElementText(s, f, true, spriteContainer)}
}

func (e *elementButton) preferredDimension(stringBounder font.StringBounder, x, _ float64) geom.XDimension2D {
	dim := e.textDimensionAt(stringBounder, x+buttonStroke+buttonMarginX)
	dim = dim.DeltaWH(2*buttonMarginX, 2*buttonMarginY)
	return dim.Delta(2 * buttonStroke)
}

func (e *elementButton) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	dim := e.preferredDimension(ug.StringBounder(), 0, 0)
	ug = ug.Apply(klimt.UStrokeWithThickness(buttonStroke))
	ug = ug.Apply(klimt.Bg(saltColorEE())).Apply(saltBlack())
	ug.Apply(klimt.NewUTranslate(buttonStroke, buttonStroke)).
		Draw(klimt.URectangleBuild(dim.Width()-2*buttonStroke, dim.Height()-2*buttonStroke).Rounded(10))
	dimPureText := e.pureTextDimension(ug.StringBounder())
	e.drawText(ug, (dim.Width()-dimPureText.Width())/2, buttonStroke+buttonMarginY)
}

// elementTextField is "\"text\"": the text with a line under it.
type elementTextField struct {
	abstractElementText
}

func newElementTextField(s string, f font.UFont, spriteContainer klimt.ISkinSimple) *elementTextField {
	return &elementTextField{newAbstractElementText(s, f, true, spriteContainer)}
}

func (e *elementTextField) preferredDimension(stringBounder font.StringBounder, x, _ float64) geom.XDimension2D {
	return e.textDimensionAt(stringBounder, x).DeltaWH(6, 2)
}

func (e *elementTextField) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	e.drawText(ug, 3, 0)
	dim := e.preferredDimension(ug.StringBounder(), 0, 0)
	textDim := e.textDimensionAt(ug.StringBounder(), 0)
	ug.Apply(klimt.NewUTranslate(1, textDim.Height())).Draw(klimt.ULineH(dim.Width() - 3))
	y3 := textDim.Height() - 3
	ug.Apply(klimt.NewUTranslate(1, y3)).Draw(klimt.ULineV(2))
	ug.Apply(klimt.NewUTranslate(3+textDim.Width()+1, y3)).Draw(klimt.ULineV(2))
}

// elementRadioCheckbox is a radio button "(X) text" or a checkbox
// "[X] text", checked or not.
type elementRadioCheckbox struct {
	abstractElement
	block   klimt.TextBlock
	radio   bool
	checked bool
}

const (
	radioCheckboxRectangle = 10
	radioCheckboxEllipse   = 10
	radioCheckboxEllipse2  = 4
	radioCheckboxMargin    = 20
	radioCheckboxStroke    = 1.5
)

func newElementRadioCheckbox(lines []string, f font.UFont, radio, checked bool,
	spriteContainer klimt.ISkinSimple) *elementRadioCheckbox {
	return &elementRadioCheckbox{
		block: klimt.DisplayCreate(lines...).Create(blackBlueTrue(f), geom.HorizontalAlignmentLeft, spriteContainer),
		radio: radio, checked: checked,
	}
}

func (e *elementRadioCheckbox) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	return e.block.CalculateDimension(stringBounder).DeltaWH(radioCheckboxMargin, 0)
}

func (e *elementRadioCheckbox) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	ug = ug.Apply(saltBlack())
	e.block.DrawU(ug.Apply(klimt.UTranslateDx(radioCheckboxMargin)))
	height := e.preferredDimension(ug.StringBounder(), 0, 0).Height()
	ug = ug.Apply(klimt.UStrokeWithThickness(radioCheckboxStroke))
	if e.radio {
		e.drawRadio(ug, height)
	} else {
		e.drawOther(ug, height)
	}
}

func (e *elementRadioCheckbox) drawOther(ug klimt.UGraphic, height float64) {
	ug.Apply(klimt.NewUTranslate(2, (height-radioCheckboxRectangle)/2)).
		Draw(klimt.URectangleBuild(radioCheckboxRectangle, radioCheckboxRectangle))
	if e.checked {
		poly := klimt.NewUPolygon()
		poly.AddPoint(0, 0)
		poly.AddPoint(3, 3)
		poly.AddPoint(10, -6)
		poly.AddPoint(3, 1)
		ug = ug.Apply(klimt.ChangeBack(ug))
		ug = ug.Apply(klimt.NewUTranslate(3, 6))
		ug.Draw(poly)
	}
}

func (e *elementRadioCheckbox) drawRadio(ug klimt.UGraphic, height float64) {
	ug.Apply(klimt.NewUTranslate(2, (height-radioCheckboxEllipse)/2)).
		Draw(klimt.UEllipseBuild(radioCheckboxEllipse, radioCheckboxEllipse))
	if e.checked {
		ug = ug.Apply(klimt.ChangeBack(ug))
		// (ELLIPSE - ELLIPSE2) / 2 is an int division upstream.
		ug = ug.Apply(klimt.NewUTranslate(2+(radioCheckboxEllipse-radioCheckboxEllipse2)/2,
			(height-radioCheckboxEllipse2)/2))
		ug.Draw(klimt.UEllipseBuild(radioCheckboxEllipse2, radioCheckboxEllipse2))
	}
}

// elementDroplist is "^text^": a box filled with #EEE with a down arrow.
// The items after the text, separated by "^", make it open: they are drawn
// under it in the pass 1, over what follows.
type elementDroplist struct {
	abstractElementText
	openDrop klimt.TextBlock
}

const droplistBox = 12

func newElementDroplist(s string, f font.UFont, spriteContainer klimt.ISkinSimple) *elementDroplist {
	e := &elementDroplist{abstractElementText: newAbstractElementText(droplistExtract(s), f, true, spriteContainer)}
	drop := klimt.JavaTokenize(s, "^")
	if len(drop) > 0 {
		drop = drop[1:]
	}
	if len(drop) > 0 {
		e.openDrop = klimt.DisplayCreate(drop...).Create(e.config, geom.HorizontalAlignmentLeft, spriteContainer)
	}
	return e
}

// droplistExtract is the text before the first "^".
func droplistExtract(s string) string {
	if idx := strings.IndexByte(s, '^'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func (e *elementDroplist) mayDrawBeyondPreferredDimension() bool { return e.openDrop != nil }

func (e *elementDroplist) preferredDimension(stringBounder font.StringBounder, x, _ float64) geom.XDimension2D {
	return e.textDimensionAt(stringBounder, x+2).DeltaWH(4+droplistBox, 4)
}

func (e *elementDroplist) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	dim := e.preferredDimension(ug.StringBounder(), 0, 0)
	ug = ug.Apply(saltBlack())
	if zIndex == 0 {
		ug.Apply(klimt.Bg(saltColorEE())).Draw(klimt.URectangleBuild(dim.Width()-1, dim.Height()-1))
		e.drawText(ug, 2, 2)
		xline := dim.Width() - droplistBox
		ug.Apply(klimt.UTranslateDx(xline)).Draw(klimt.ULineV(dim.Height() - 1))
		poly := klimt.NewUPolygon()
		poly.AddPoint(0, 0)
		poly.AddPoint(droplistBox-6, 0)
		dimText := e.pureTextDimension(ug.StringBounder())
		// (box - 6) / 2 is an int division upstream.
		poly.AddPoint((droplistBox-6)/2, dimText.Height()-8)
		ug.Apply(klimt.ChangeBack(ug)).Apply(klimt.NewUTranslate(xline+3, 6)).Draw(poly)
	}
	if e.openDrop != nil && zIndex == 1 {
		dimOpen := e.openDrop.CalculateDimension(ug.StringBounder()).AtLeast(dim.Width()-1, 0)
		ug = ug.Apply(klimt.UTranslateDy(dim.Height() - 1))
		ug.Apply(klimt.Bg(saltColorEE())).Draw(klimt.URectangleBuild(dimOpen.Width()-1, dimOpen.Height()-1))
		e.openDrop.DrawU(ug)
	}
}

// elementLine is a separator: "..", "==", "~~" or "--".
type elementLine struct {
	abstractElement
	separator byte
}

func (*elementLine) preferredDimension(font.StringBounder, float64, float64) geom.XDimension2D {
	return geom.NewXDimension2D(10, 6)
}

func (e *elementLine) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	ug = ug.Apply(saltColorAA())
	y2 := dimOf(dimToUse, "getHeight").Height() / 2
	if e.separator == '=' {
		y2 = y2 - 1
	}
	drawSeparatorLine(ug, 0, y2, dimToUse.Width(), e.separator)
}

func drawSeparatorLine(ug klimt.UGraphic, x, y, widthToUse float64, separator byte) {
	switch separator {
	case '=':
		ug.Apply(klimt.UStrokeSimple()).Apply(klimt.NewUTranslate(x, y)).Draw(klimt.ULineH(widthToUse))
		ug.Apply(klimt.UStrokeSimple()).Apply(klimt.NewUTranslate(x, y+2)).Draw(klimt.ULineH(widthToUse))
	case '.':
		ug.Apply(klimt.NewUStroke(1, 2, 1)).Apply(klimt.NewUTranslate(x, y)).Draw(klimt.ULineH(widthToUse))
	case '-':
		ug.Apply(klimt.UStrokeSimple()).Apply(klimt.NewUTranslate(x, y)).Draw(klimt.ULineH(widthToUse))
	default:
		ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.NewUTranslate(x, y)).Draw(klimt.ULineH(widthToUse))
	}
}

// elementImage is a pixel image: each "X" is a black pixel.
type elementImage struct {
	abstractElement
	img []string
}

func (e *elementImage) preferredDimension(font.StringBounder, float64, float64) geom.XDimension2D {
	if len(e.img) == 0 {
		// An image without rows throws, as List.get.
		uml.ThrowIndexOutOfBounds(0, 0)
	}
	return geom.NewXDimension2D(float64(len(utf16.Encode([]rune(e.img[0])))), float64(len(e.img)))
}

func (e *elementImage) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	ug = ug.Apply(saltBlack())
	rows := make([][]uint16, len(e.img))
	for j, s := range e.img {
		rows[j] = utf16.Encode([]rune(s))
	}
	w := len(rows[0])
	h := len(rows)
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			// A row shorter than the first one throws, as String.charAt.
			if i >= len(rows[j]) {
				uml.Throw(uml.StringIndexOutOfBoundsException,
					fmt.Sprintf("Index %d out of bounds for length %d", i, len(rows[j])))
			}
			if rows[j][i] == 'X' {
				ug.Apply(klimt.NewUTranslate(float64(i), float64(j))).Draw(&klimt.UPixel{})
			}
		}
	}
}
