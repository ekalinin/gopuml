// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PlayerFrame.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// playerFrame is the title of a player, with the line under it unless the
// player is compact.
type playerFrame struct {
	skinParam klimt.ISkinParam
	title     klimt.TextBlock
	compact   bool
}

func newPlayerFrame(title klimt.TextBlock, skinParam klimt.ISkinParam, compact bool) *playerFrame {
	return &playerFrame{title: title, skinParam: skinParam, compact: compact}
}

func (f *playerFrame) style() *klimt.Style {
	return timingStyle().MergedStyle(f.skinParam.CurrentStyleBuilder())
}

func (f *playerFrame) lineColor() color.HColor {
	return f.style().Value(klimt.PNameLineColor).AsColor(f.skinParam.IHtmlColorSet())
}

func (f *playerFrame) uStroke() klimt.UStroke { return f.style().Stroke() }

// drawFrameTitle draws the title and, unless the player is compact, the
// line under it that rises to its right; a player without a title draws
// nothing.
func (f *playerFrame) drawFrameTitle(ug klimt.UGraphic) {
	if f.title == klimt.TextBlockUtilsEmptyTextBlock() {
		return
	}
	f.title.DrawU(ug)
	if f.compact {
		return
	}
	ug = ug.Apply(f.lineColor()).Apply(f.uStroke())
	dimTitle := f.title.CalculateDimension(ug.StringBounder())
	widthTmp := dimTitle.Width() + 1
	height := dimTitle.Height() + 1
	drawLine(ug, -marginX1, height, widthTmp, height, widthTmp+10, 0)
}

func (f *playerFrame) height(stringBounder font.StringBounder) float64 {
	height := f.title.CalculateDimension(stringBounder).Height()
	if f.compact {
		return height - 1
	}
	return height + 1
}

func drawLine(ug klimt.UGraphic, coord ...float64) {
	for i := 0; i < len(coord)-2; i += 2 {
		x1 := coord[i]
		y1 := coord[i+1]
		x2 := coord[i+2]
		y2 := coord[i+3]
		ug.Apply(klimt.NewUTranslate(x1, y1)).Draw(klimt.NewULine(x2-x1, y2-y1))
	}
}
