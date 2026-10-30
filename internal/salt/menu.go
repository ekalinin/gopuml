// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementMenuBar.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementMenuEntry.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementMenuPopup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// menuPopup is an entry of a menu bar with the list of its items.
type menuPopup struct {
	entry *elementMenuEntry
	popup *elementMenuPopup
}

// elementMenuBar is a menu bar: its entries on #DDD in the pass 0, and the
// lists of the entries with items in the pass 1. Upstream keeps the lists in
// a HashMap keyed by the entries, whose order follows Java's identity hash
// codes; they are kept here in the order of the entries.
type elementMenuBar struct {
	abstractElement
	entries         []*elementMenuEntry
	popups          []menuPopup
	font            font.UFont
	spriteContainer klimt.ISkinSimple
}

func newElementMenuBar(f font.UFont, spriteContainer klimt.ISkinSimple) *elementMenuBar {
	return &elementMenuBar{font: f, spriteContainer: spriteContainer}
}

func (e *elementMenuBar) addEntry(s string) {
	e.entries = append(e.entries, newElementMenuEntry(s, e.font, e.spriteContainer))
}

func (e *elementMenuBar) addSubEntry(s, sub string) {
	e.popupOf(e.elementMenuEntry(s)).addEntry(sub)
}

func (e *elementMenuBar) popupOf(entry *elementMenuEntry) *elementMenuPopup {
	for _, p := range e.popups {
		if p.entry == entry {
			return p.popup
		}
	}
	popup := newElementMenuPopup(e.font, e.spriteContainer)
	e.popups = append(e.popups, menuPopup{entry: entry, popup: popup})
	return popup
}

// elementMenuEntry returns the entry of a name. An unknown name throws an
// IllegalArgumentException.
func (e *elementMenuBar) elementMenuEntry(n string) *elementMenuEntry {
	for _, entry := range e.entries {
		if entry.text == n {
			return entry
		}
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return nil
}

func (e *elementMenuBar) preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	w := 0.0
	h := 0.0
	for _, entry := range e.entries {
		dim := entry.preferredDimension(stringBounder, x, y)
		w += dim.Width() + 10
		h = max(h, dim.Height())
	}
	return geom.NewXDimension2D(w, h)
}

func (e *elementMenuBar) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	preferred := e.preferredDimension(ug.StringBounder(), 0, 0)
	ug = ug.Apply(saltBlack())
	x1 := 0.0
	if zIndex == 0 {
		dim := dimOf(dimToUse, "getWidth")
		ug.Apply(klimt.Bg(saltColorDD())).Draw(klimt.URectangleBuild(dim.Width(), dim.Height()))
		for _, entry := range e.entries {
			entry.drawU(ug.Apply(klimt.UTranslateDx(x1)), zIndex, dimToUse)
			w := entry.preferredDimension(ug.StringBounder(), x1, 0).Width()
			entry.x = x1
			x1 += w + 10
		}
		return
	}
	if zIndex == 1 {
		// The entries keep this background: a later drawing of the same
		// menu draws them with it, as upstream does.
		for _, p := range e.popups {
			p.entry.background = saltColorBB()
		}
		y1 := preferred.Height()
		for _, p := range e.popups {
			xpopup := p.entry.x
			dim := p.popup.preferredDimension(ug.StringBounder(), xpopup, y1)
			p.popup.drawU(ug.Apply(klimt.NewUTranslate(xpopup, y1)), zIndex, &dim)
		}
	}
}

// elementMenuEntry is an entry of a menu bar or an item of its list. "-" is
// a separator.
type elementMenuEntry struct {
	abstractElement
	block      klimt.TextBlock
	text       string
	background color.HColor
	x          float64
}

func newElementMenuEntry(s string, f font.UFont, spriteContainer klimt.ISkinSimple) *elementMenuEntry {
	return &elementMenuEntry{
		block: klimt.DisplayCreate(s).Create(blackBlueTrue(f), geom.HorizontalAlignmentLeft, spriteContainer),
		text:  s,
	}
}

func (e *elementMenuEntry) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	if e.text == "-" {
		return geom.NewXDimension2D(10, 5)
	}
	return e.block.CalculateDimension(stringBounder)
}

func (e *elementMenuEntry) drawU(ug klimt.UGraphic, _ int, _ *geom.XDimension2D) {
	ug = ug.Apply(saltBlack())
	if e.background != nil {
		dim := e.preferredDimension(ug.StringBounder(), 0, 0)
		ug.Apply(klimt.Bg(e.background)).Draw(klimt.URectangleBuild(dim.Width(), dim.Height()))
	}
	e.block.DrawU(ug)
}

// elementMenuPopup is the list of the items of a menu entry, drawn in the
// pass 1.
type elementMenuPopup struct {
	abstractElement
	entries         []*elementMenuEntry
	font            font.UFont
	spriteContainer klimt.ISkinSimple
}

func newElementMenuPopup(f font.UFont, spriteContainer klimt.ISkinSimple) *elementMenuPopup {
	return &elementMenuPopup{font: f, spriteContainer: spriteContainer}
}

func (e *elementMenuPopup) addEntry(s string) {
	e.entries = append(e.entries, newElementMenuEntry(s, e.font, e.spriteContainer))
}

func (e *elementMenuPopup) preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	w := 0.0
	h := 0.0
	for _, entry := range e.entries {
		dim := entry.preferredDimension(stringBounder, x, y)
		w = max(w, dim.Width())
		h += dim.Height()
	}
	return geom.NewXDimension2D(w, h)
}

func (e *elementMenuPopup) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	if zIndex != 1 {
		return
	}
	ug = ug.Apply(saltBlack())
	dim := dimOf(dimToUse, "getWidth")
	ug.Apply(klimt.Bg(saltColorDD())).Draw(klimt.URectangleBuild(dim.Width(), dim.Height()))
	y1 := 0.0
	for _, entry := range e.entries {
		h := entry.preferredDimension(ug.StringBounder(), 0, y1).Height()
		if entry.text == "-" {
			ug.Apply(klimt.UTranslateDy(y1 + h/2)).Draw(klimt.ULineH(dim.Width()))
		} else {
			entry.drawU(ug.Apply(klimt.UTranslateDy(y1)), zIndex, dimToUse)
		}
		y1 += h
	}
}
