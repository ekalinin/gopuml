// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the lanes, the time axis, the
// messages and the highlights, with the annotations around them (title,
// header, footer, caption, legend, main frame). A panic while drawing is
// returned as an error.
func (d *TimingDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: the drawing with a margin of 10.
func (d *TimingDiagram) TextBlock(int, export.FileFormatOption) klimt.TextBlock {
	return klimt.TextBlockUtilsWithMargin(&timingTextBlock{d: d}, 10, 10)
}

// timingTextBlock is the drawing of the diagram. Its dimension is computed
// before the drawing adds a time at 0 to an empty ruler, as upstream's.
type timingTextBlock struct {
	klimt.TextBlockDefaults
	d *TimingDiagram
}

func (t *timingTextBlock) DrawU(ug klimt.UGraphic) { t.d.drawInternal(ug) }

func (t *timingTextBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	withBeforeRuler := t.d.part1MaxWidth(stringBounder)
	totalWith := withBeforeRuler + t.d.ruler.getWidth() + marginX1 + marginX2
	return geom.NewXDimension2D(totalWith, t.d.heightTotal(stringBounder))
}

// black is the line color of the timingDiagram style.
func (d *TimingDiagram) black() color.HColor {
	style := timingStyle().MergedStyle(d.SkinParam().CurrentStyleBuilder())
	return style.Value(klimt.PNameLineColor).AsColor(d.SkinParam().IHtmlColorSet())
}

func (d *TimingDiagram) drawInternal(ug klimt.UGraphic) {
	d.ruler.ensureNotEmpty()
	stringBounder := ug.StringBounder()
	part1MaxWidth := d.part1MaxWidth(stringBounder)
	for _, p := range d.playersInOrder() {
		generalBackgroundColor := p.base().generalBackgroundColor
		if generalBackgroundColor != nil {
			ugPlayer := ug.Apply(d.translateForPlayerFrame(p, stringBounder))
			fullHeight := p.base().frameHeight(stringBounder) + panelsOf(p).getFullHeight(stringBounder)
			ugPlayer = ugPlayer.Apply(generalBackgroundColor).Apply(klimt.Bg(generalBackgroundColor))
			ugPlayer.Draw(klimt.URectangleBuild(d.widthTotal(stringBounder), fullHeight))
		}
	}
	widthPart1 := klimt.UTranslateDx(part1MaxWidth)
	if !d.compactByDefault {
		d.drawBorder(ug)
	}
	ug = ug.Apply(klimt.UTranslateDx(marginX1))
	d.drawHighlightsBack(ug.Apply(widthPart1))
	d.ruler.drawVlines(ug.Apply(widthPart1), d.heightInner(stringBounder))
	for _, p := range d.playersInOrder() {
		ugPlayer := ug.Apply(d.translateForPlayer(p, stringBounder))
		ugFrame := ug.Apply(d.translateForPlayerFrame(p, stringBounder))
		if !p.base().compact {
			d.drawHorizontalSeparator(ugFrame)
		}
		p.base().drawFrameTitle(ugFrame)
		panels := panelsOf(p)
		panels.drawLeftPanel(ugPlayer, part1MaxWidth)
		panels.drawRightPanel(ugPlayer.Apply(widthPart1))
	}
	ug = ug.Apply(widthPart1)
	d.ruler.drawTimeAxis(ug.Apply(d.lastTranslate(stringBounder)), d.timeAxisStategy, d.namedTimes())
	for _, message := range d.messages {
		d.drawMessages(ug, message)
	}
	d.drawHighlightsLines(ug)
}

func (d *TimingDiagram) drawHorizontalSeparator(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug = ug.Apply(d.black())
	ug = ug.Apply(d.borderStroke())
	ug = ug.Apply(klimt.UTranslateDx(-marginX1))
	ug.Draw(klimt.ULineH(d.widthTotal(stringBounder)))
}

func (d *TimingDiagram) drawBorder(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	border := klimt.ULineV(d.lastTranslate(stringBounder).Dy())
	ug = ug.Apply(d.black()).Apply(d.borderStroke())
	ug.Draw(border)
	ug.Apply(klimt.UTranslateDx(d.widthTotal(stringBounder))).Draw(border)
}

func (d *TimingDiagram) borderStroke() klimt.UStroke {
	return timingStyle().MergedStyle(d.SkinParam().CurrentStyleBuilder()).Stroke()
}

func (d *TimingDiagram) lastTranslate(stringBounder font.StringBounder) klimt.UTranslate {
	return d.translateForPlayer(nil, stringBounder)
}

func (d *TimingDiagram) drawHighlightsBack(ug klimt.UGraphic) {
	height := d.heightInner(ug.StringBounder())
	for _, h := range d.highlights {
		h.drawHighlightsBack(ug, d.ruler, height)
	}
}

func (d *TimingDiagram) drawHighlightsLines(ug klimt.UGraphic) {
	height := d.heightInner(ug.StringBounder())
	for _, h := range d.highlights {
		h.drawHighlightsLines(ug, d.ruler, height)
		start := d.ruler.getPosInPixel(h.tickFrom)
		h.captionBlock(d.SkinParam()).DrawU(ug.Apply(klimt.NewUTranslate(start+3, 2)))
	}
}

func (d *TimingDiagram) heightTotal(stringBounder font.StringBounder) float64 {
	return d.heightInner(stringBounder) + d.ruler.getHeight(stringBounder)
}

func (d *TimingDiagram) heightInner(stringBounder font.StringBounder) float64 {
	return d.lastTranslate(stringBounder).Dy()
}

func (d *TimingDiagram) widthTotal(stringBounder font.StringBounder) float64 {
	return d.part1MaxWidth(stringBounder) + d.ruler.getWidth() + marginX1 + marginX2
}

// part1MaxWidth is the width of the widest left panel.
func (d *TimingDiagram) part1MaxWidth(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, p := range d.playersInOrder() {
		width = max(width, panelsOf(p).getLeftPanelWidth(stringBounder))
	}
	return width
}

// drawMessages draws a message, unless one of its players has no point at
// its time (a message without a time before any time).
func (d *TimingDiagram) drawMessages(ug klimt.UGraphic, message *timeMessage) {
	player1 := message.tickInPlayer1.player
	player2 := message.tickInPlayer2.player
	stringBounder := ug.StringBounder()
	translate1 := d.translateForPlayer(player1, stringBounder)
	translate2 := d.translateForPlayer(player2, stringBounder)
	pt1 := panelsOf(player1).getTimeProjection(stringBounder, message.tickInPlayer1.tick)
	pt2 := panelsOf(player2).getTimeProjection(stringBounder, message.tickInPlayer2.tick)
	if pt1 == nil || pt2 == nil {
		return
	}
	arrow := newTimeArrow(pt1.translated(translate1), pt2.translated(translate2), message.label, d.SkinParam(),
		message)
	arrow.DrawU(ug)
}

// translateForPlayer is the position of the panels of a player, below its
// title, or the position below the last player for nil.
func (d *TimingDiagram) translateForPlayer(candidate player, stringBounder font.StringBounder) klimt.UTranslate {
	y := 0.0
	for _, p := range d.playersInOrder() {
		y += p.base().frameHeight(stringBounder)
		if candidate == p {
			return klimt.UTranslateDy(y)
		}
		y += panelsOf(p).getFullHeight(stringBounder)
	}
	if candidate == nil {
		return klimt.UTranslateDy(y)
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}

// translateForPlayerFrame is the position of the title of a player.
func (d *TimingDiagram) translateForPlayerFrame(candidate player,
	stringBounder font.StringBounder) klimt.UTranslate {
	y := 0.0
	for _, p := range d.playersInOrder() {
		if candidate == p {
			return klimt.UTranslateDy(y)
		}
		y += p.base().frameHeight(stringBounder)
		y += panelsOf(p).getFullHeight(stringBounder)
	}
	if candidate == nil {
		return klimt.UTranslateDy(y)
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}
