// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeMessage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// timeMessage is a message from a time in the lane of a player to a time in
// the lane of another one.
type timeMessage struct {
	decoration.WithLinkType
	tickInPlayer1 tickInPlayer
	tickInPlayer2 tickInPlayer
	label         *klimt.Display
	skinParam     klimt.ISkinParam
	styleBuilder  *klimt.StyleBuilder
}

// newTimeMessage builds a message in the line color of the arrow style of
// the time of the message; label is the message, hasLabel false for none.
func newTimeMessage(tickInPlayer1, tickInPlayer2 tickInPlayer, label string, hasLabel bool,
	skinParam klimt.ISkinParam) *timeMessage {
	m := &timeMessage{skinParam: skinParam, styleBuilder: skinParam.CurrentStyleBuilder(),
		tickInPlayer1: tickInPlayer1, tickInPlayer2: tickInPlayer2,
		label: klimt.DisplayWithNewlinesString(skinParam.Pragma(), label, hasLabel)}
	m.WithLinkType = decoration.InitWithLinkType(decoration.NewLinkType(decoration.LinkDecorNone,
		decoration.LinkDecorNone), func() {})
	m.SetSpecificColor(m.color(), 0)
	return m
}

// uStroke is always the stroke of the arrow style: the dashed, dotted, bold
// and thickness styles of the message change nothing, as upstream does.
func (m *timeMessage) uStroke() klimt.UStroke {
	if m.styleBuilder == nil {
		return klimt.UStrokeWithThickness(1.5)
	}
	return m.style().Stroke()
}

func (m *timeMessage) color() color.HColor {
	if m.styleBuilder == nil {
		return color.MyRed
	}
	return m.style().Value(klimt.PNameLineColor).AsColor(m.skinParam.IHtmlColorSet())
}

func (m *timeMessage) style() *klimt.Style {
	return timingStyle(klimt.SNameArrow).MergedStyle(m.styleBuilder)
}
