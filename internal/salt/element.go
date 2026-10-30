// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Element.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/AbstractElement.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/AbstractElementText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementEmpty.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/WrappedElement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"regexp"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// element is a widget or a group of a salt diagram. It is drawn in two
// passes: the z-index 0 for the widgets, 1 for what floats over them (open
// droplists and menus). dimToUse is the size of its cell, nil where upstream
// passes null (the cells of a tree).
type element interface {
	preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D
	drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D)
	// mayDrawBeyondPreferredDimension reports whether the element, or an
	// element inside it, draws past its preferred dimension (an open
	// droplist), so that the canvas is measured from a dry run.
	mayDrawBeyondPreferredDimension() bool
}

// abstractElement gives the default of mayDrawBeyondPreferredDimension.
type abstractElement struct{}

func (abstractElement) mayDrawBeyondPreferredDimension() bool { return false }

// The colors of the widgets, each with its dark mode color.
func saltBlack() color.HColor   { return color.Black.WithDark(color.White) }
func saltWhite() color.HColor   { return color.White.WithDark(color.Black) }
func saltColor88() color.HColor { return buildColor("#8", "#8") }
func saltColorAA() color.HColor { return buildColor("#A", "#6") }
func saltColorBB() color.HColor { return buildColor("#B", "#5") }
func saltColorDD() color.HColor { return buildColor("#D", "#3") }
func saltColorEE() color.HColor { return buildColor("#E", "#2") }

func buildColor(color1, color2 string) color.HColor {
	return color.Instance().GetColorOrWhite(color1).WithDark(color.Instance().GetColorOrWhite(color2))
}

func blackBlueTrue(f font.UFont) *klimt.FontConfiguration {
	return klimt.FontConfigurationBlackBlueTrue(f)
}

// dimOf returns dimToUse, on which upstream calls method: when it is null,
// the call throws a NullPointerException.
func dimOf(dimToUse *geom.XDimension2D, method string) geom.XDimension2D {
	if dimToUse == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.klimt.geom.XDimension2D.` + method +
			`()" because "dimToUse" is null`)
	}
	return *dimToUse
}

// openIconPattern is the "<&[-\\w]+>" of getCharNumber, a plain Java pattern.
var openIconPattern = regexp.MustCompile(`<&[-\w]+>`)

// abstractElementText is the text of a button, a text field or a droplist.
// With manageLength, the width is at least 8 pixels per character of the
// text, counted before it is trimmed.
type abstractElementText struct {
	abstractElement
	block      klimt.TextBlock
	config     *klimt.FontConfiguration
	charLength int
}

func newAbstractElementText(s string, f font.UFont, manageLength bool,
	spriteContainer klimt.ISkinSimple) abstractElementText {
	t := abstractElementText{config: blackBlueTrue(f)}
	if manageLength {
		t.charLength = charNumber(s)
		s = text.Trin(s)
	}
	t.block = klimt.DisplayCreate(s).Create(t.config, geom.HorizontalAlignmentLeft, spriteContainer)
	return t
}

// charNumber is getCharNumber: an icon counts as two characters, a creole
// tag as none, and the length is Java's, in UTF-16 code units.
func charNumber(s string) int {
	s = openIconPattern.ReplaceAllString(s, "00")
	s = klimt.SplitterPurgeAllTag(s)
	return len(utf16.Encode([]rune(s)))
}

func (t *abstractElementText) drawText(ug klimt.UGraphic, x, y float64) {
	t.block.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
}

func (t *abstractElementText) pureTextDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.block.CalculateDimension(stringBounder)
}

func (t *abstractElementText) textDimensionAt(stringBounder font.StringBounder, _ float64) geom.XDimension2D {
	result := t.block.CalculateDimension(stringBounder)
	if t.charLength == 0 {
		return result
	}
	dimSpace := singleSpace()
	return geom.NewXDimension2D(max(result.Width(), float64(t.charLength)*dimSpace), result.Height())
}

// singleSpace is upstream's getSingleSpace: a fixed 8 pixels.
func singleSpace() float64 { return 8 }

// elementEmpty is an empty position of a border layout.
type elementEmpty struct{ abstractElement }

func (*elementEmpty) preferredDimension(font.StringBounder, float64, float64) geom.XDimension2D {
	return geom.NewXDimension2D(1, 1)
}

func (*elementEmpty) drawU(klimt.UGraphic, int, *geom.XDimension2D) {}

// wrappedElement is an image drawn again by its name.
type wrappedElement struct{ wrapped element }

func (w *wrappedElement) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	return w.wrapped.preferredDimension(stringBounder, 0, 0)
}

func (w *wrappedElement) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	w.wrapped.drawU(ug, zIndex, dimToUse)
}

func (w *wrappedElement) mayDrawBeyondPreferredDimension() bool {
	return w.wrapped.mayDrawBeyondPreferredDimension()
}
