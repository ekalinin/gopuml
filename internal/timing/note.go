// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingNote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// timingNote is a note above or below the lane of a player, at a time; when
// is nil for a note given before any time.
type timingNote struct {
	when      *TimeTick
	note      *klimt.Display
	position  abel.Position
	skinParam klimt.ISkinParam
	style     *klimt.Style
}

func newTimingNote(when *TimeTick, note *klimt.Display, position abel.Position, skinParam klimt.ISkinParam,
	style *klimt.Style) *timingNote {
	return &timingNote{style: style, note: note, when: when, skinParam: skinParam, position: position}
}

func (n *timingNote) drawU(ug klimt.UGraphic) {
	if n.position == abel.PositionBottom {
		ug = ug.Apply(klimt.UTranslateDy(noteMarginY / 2))
	}
	n.createOpale().DrawU(ug)
}

func (n *timingNote) createOpale() *klimt.Opale {
	fc := klimt.FontConfigurationFromStyle(n.skinParam, n.style, nil)
	shadowing := n.style.Shadowing()
	borderColor := n.style.Value(klimt.PNameLineColor).AsColor(n.skinParam.IHtmlColorSet())
	noteBackgroundColor := n.style.Value(klimt.PNameBackGroundColor).AsColor(n.skinParam.IHtmlColorSet())
	stroke := n.style.Stroke()
	sheet := n.skinParam.Sheet(fc, n.skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft),
		klimt.CreoleModeFull).CreateSheet(n.note)
	sheet1 := klimt.NewSheetBlock1Padding(sheet, klimt.LineBreakStrategyNone, n.skinParam.Padding())
	return klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, sheet1, false, &stroke)
}

func (n *timingNote) height(stringBounder font.StringBounder) float64 {
	return n.createOpale().CalculateDimension(stringBounder).Height() + noteMarginY
}

const noteMarginY = 10
