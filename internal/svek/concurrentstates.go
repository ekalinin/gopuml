// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ConcurrentStates.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// concurrentSeparator is upstream's ConcurrentStates.Separator: '|' places
// the regions side by side (VERTICAL), '-' one above the other
// (HORIZONTAL).
type concurrentSeparator bool

const (
	separatorVertical   concurrentSeparator = true
	separatorHorizontal concurrentSeparator = false
)

func concurrentSeparatorFromChar(sep rune) concurrentSeparator {
	switch sep {
	case '|':
		return separatorVertical
	case '-':
		return separatorHorizontal
	}
	throwIllegalArgument("")
	return separatorHorizontal
}

// move is the translation from a region to the next one.
func (s concurrentSeparator) move(dim geom.XDimension2D) klimt.UTranslate {
	if s == separatorVertical {
		return klimt.UTranslateDx(dim.Width())
	}
	return klimt.UTranslateDy(dim.Height())
}

// add is the size of orig with the region other next to it.
func (s concurrentSeparator) add(orig, other geom.XDimension2D) geom.XDimension2D {
	if s == separatorVertical {
		return geom.NewXDimension2D(orig.Width()+other.Width(), max(orig.Height(), other.Height()))
	}
	return geom.NewXDimension2D(max(orig.Width(), other.Width()), orig.Height()+other.Height())
}

// drawSeparator draws the dashed line between two regions.
func (s concurrentSeparator) drawSeparator(ug klimt.UGraphic, dimTotal geom.XDimension2D) {
	const thicknessBorder = 1.5
	const dash = 8
	ug = ug.Apply(klimt.NewUStroke(dash, 10, thicknessBorder))
	if s == separatorVertical {
		ug.Draw(klimt.ULineV(dimTotal.Height() + dash))
	} else {
		ug.Draw(klimt.ULineH(dimTotal.Width() + dash))
	}
}

// ConcurrentStates is upstream's ConcurrentStates: the regions of a
// composite state, drawn next to each other with dashed separators.
type ConcurrentStates struct {
	klimt.TextBlockDefaults
	inners     []klimt.IEntityImage
	separator  concurrentSeparator
	skinParam  klimt.ISkinParam
	stereotype *klimt.Stereotype
}

var _ klimt.IEntityImage = (*ConcurrentStates)(nil)

func newConcurrentStates(inners []klimt.IEntityImage, separator rune, skinParam klimt.ISkinParam,
	stereotype *klimt.Stereotype) *ConcurrentStates {
	return &ConcurrentStates{separator: concurrentSeparatorFromChar(separator), skinParam: skinParam,
		stereotype: stereotype, inners: inners}
}

func (c *ConcurrentStates) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameStateDiagram,
		klimt.SNameState).MergedStyle(c.skinParam.CurrentStyleBuilder())
}

// DrawU draws the regions and the separators between them.
func (c *ConcurrentStates) DrawU(ug klimt.UGraphic) {
	borderColor := c.style().Value(klimt.PNameLineColor).AsColor(c.skinParam.IHtmlColorSet())
	stringBounder := ug.StringBounder()
	dimTotal := c.CalculateDimension(stringBounder)
	for i, inner := range c.inners {
		inner.DrawU(ug)
		dim := inner.CalculateDimension(stringBounder)
		ug = ug.Apply(c.separator.move(dim))
		if i < len(c.inners)-1 {
			c.separator.drawSeparator(ug.Apply(borderColor), dimTotal)
		}
	}
}

// CalculateDimension returns the size of the regions next to each other.
func (c *ConcurrentStates) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	result := geom.NewXDimension2D(0, 0)
	for _, inner := range c.inners {
		result = c.separator.add(result, inner.CalculateDimension(stringBounder))
	}
	return result
}

// ShapeType returns RECTANGLE.
func (c *ConcurrentStates) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// Shield returns no margin.
func (c *ConcurrentStates) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// IsHidden returns false.
func (c *ConcurrentStates) IsHidden() bool { return false }

// OverscanX returns 0.
func (c *ConcurrentStates) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (c *ConcurrentStates) IsCrash() bool { return false }
