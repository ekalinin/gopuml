// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/Highlight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// highlight is an area over every lane between two times, between two
// lines, with a caption.
type highlight struct {
	tickFrom  *TimeTick
	tickTo    *TimeTick
	caption   *klimt.Display
	colors    *color.Colors
	skinParam klimt.ISkinParam
}

func newHighlight(skinParam klimt.ISkinParam, tickFrom, tickTo *TimeTick, caption *klimt.Display,
	colors *color.Colors) *highlight {
	return &highlight{tickFrom: tickFrom, tickTo: tickTo, caption: caption, colors: colors, skinParam: skinParam}
}

func (h *highlight) style() *klimt.Style {
	return timingStyle(klimt.SNameHighlight).MergedStyle(h.skinParam.CurrentStyleBuilder())
}

func (h *highlight) backColor() color.HColor {
	if result := h.colors.Color(color.TypeBack); result != nil {
		return result
	}
	return h.style().Value(klimt.PNameBackGroundColor).AsColor(h.skinParam.IHtmlColorSet())
}

func (h *highlight) lineColor() color.HColor {
	if result := h.colors.Color(color.TypeLine); result != nil {
		return result
	}
	return h.style().Value(klimt.PNameLineColor).AsColor(h.skinParam.IHtmlColorSet())
}

func (h *highlight) uStroke() klimt.UStroke { return h.style().Stroke() }

// captionBlock is the caption in the legacy font of timing diagrams, as
// upstream draws it.
func (h *highlight) captionBlock(skinParam klimt.ISkinParam) klimt.TextBlock {
	fc := klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamTiming, nil)
	return h.caption.Create(fc, geom.HorizontalAlignmentLeft, skinParam)
}

// drawHighlightsBack fills the area without a border.
func (h *highlight) drawHighlightsBack(ug klimt.UGraphic, ruler *TimingRuler, height float64) {
	ug = ug.Apply(color.None()).Apply(klimt.Bg(h.backColor()))
	start := ruler.getPosInPixel(h.tickFrom)
	end := ruler.getPosInPixel(h.tickTo)
	rect := klimt.URectangleBuild(end-start, height)
	ug.Apply(klimt.UTranslateDx(start)).Draw(rect)
}

// drawHighlightsLines draws the lines at the two times.
func (h *highlight) drawHighlightsLines(ug klimt.UGraphic, ruler *TimingRuler, height float64) {
	ug = ug.Apply(h.uStroke())
	ug = ug.Apply(h.lineColor())
	line := klimt.ULineV(height)
	start := ruler.getPosInPixel(h.tickFrom)
	end := ruler.getPosInPixel(h.tickTo)
	ug.Apply(klimt.UTranslateDx(start)).Draw(line)
	ug.Apply(klimt.UTranslateDx(end)).Draw(line)
}
