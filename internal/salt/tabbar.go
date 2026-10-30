// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementTabBar.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// elementTabBar is tabs, side by side or one below the other.
type elementTabBar struct {
	abstractElement
	tabs            []element
	font            font.UFont
	spriteContainer klimt.ISkinSimple
	vertical        bool
}

const (
	tabMargin1 = 2
	tabMargin2 = 3
	tabMargin3 = 10
)

func newElementTabBar(f font.UFont, spriteContainer klimt.ISkinSimple) *elementTabBar {
	return &elementTabBar{font: f, spriteContainer: spriteContainer}
}

func (e *elementTabBar) addTab(tab string) {
	e.tabs = append(e.tabs, newElementText([]string{tab}, e.font, e.spriteContainer))
}

func (e *elementTabBar) preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	if e.vertical {
		return e.preferredDimensionVertical(stringBounder, x, y)
	}
	return e.preferredDimensionHorizontal(stringBounder, x, y)
}

func (e *elementTabBar) preferredDimensionHorizontal(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	w := 0.0
	h := 0.0
	for _, elt := range e.tabs {
		dim := elt.preferredDimension(stringBounder, x, y)
		w += dim.Width() + tabMargin1 + tabMargin2 + tabMargin3
		h = max(h, dim.Height())
	}
	return geom.NewXDimension2D(w, h)
}

func (e *elementTabBar) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	ug = ug.Apply(saltBlack())
	if e.vertical {
		e.drawUVertical(ug, 0, 0, zIndex, dimToUse)
	} else {
		e.drawUHorizontal(ug, 0, 0, zIndex, dimToUse)
	}
}

func (e *elementTabBar) drawUHorizontal(ug klimt.UGraphic, x, y float64, zIndex int, dimToUse *geom.XDimension2D) {
	x1 := x
	for _, elt := range e.tabs {
		elt.drawU(ug.Apply(klimt.NewUTranslate(x1+tabMargin1, y)), zIndex, dimToUse)
		dimText := elt.preferredDimension(ug.StringBounder(), x1, y)
		w := dimText.Width()
		ug.Apply(klimt.NewUTranslate(x1, y)).Draw(klimt.ULineV(dimText.Height()))
		ug.Apply(klimt.NewUTranslate(x1, y)).Draw(klimt.ULineH(w + tabMargin1 + tabMargin2))
		ug.Apply(klimt.NewUTranslate(x1+w+tabMargin1+tabMargin2, y)).Draw(klimt.ULineV(dimText.Height()))
		ug.Apply(klimt.NewUTranslate(x1+w+tabMargin1+tabMargin2, y+dimText.Height())).Draw(klimt.ULineH(tabMargin3))
		x1 += w + tabMargin1 + tabMargin2 + tabMargin3
	}
}

func (e *elementTabBar) preferredDimensionVertical(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	w := 0.0
	h := 0.0
	for _, elt := range e.tabs {
		dim := elt.preferredDimension(stringBounder, x, y)
		h += dim.Height() + tabMargin1 + tabMargin2 + tabMargin3
		w = max(w, dim.Width())
	}
	return geom.NewXDimension2D(w, h)
}

func (e *elementTabBar) drawUVertical(ug klimt.UGraphic, x, y float64, zIndex int, dimToUse *geom.XDimension2D) {
	preferred := e.preferredDimension(ug.StringBounder(), x, y)
	ug = ug.Apply(klimt.NewUTranslate(x, y))
	// Upstream starts from x, which is 0 as y.
	y1 := x
	for _, elt := range e.tabs {
		elt.drawU(ug.Apply(klimt.UTranslateDy(y1+tabMargin1)), zIndex, dimToUse)
		dimText := elt.preferredDimension(ug.StringBounder(), x, y1)
		h := dimText.Height()
		ug.Apply(klimt.UTranslateDy(y1)).Draw(klimt.ULineH(preferred.Width()))
		ug.Apply(klimt.UTranslateDy(y1)).Draw(klimt.ULineV(h + tabMargin1 + tabMargin2))
		ug.Apply(klimt.UTranslateDy(y1 + h + tabMargin1 + tabMargin2)).Draw(klimt.ULineH(preferred.Width()))
		ug.Apply(klimt.NewUTranslate(preferred.Width(), y1+h+tabMargin1+tabMargin2)).Draw(klimt.ULineV(tabMargin3))
		y1 += h + tabMargin1 + tabMargin2 + tabMargin3
	}
}
