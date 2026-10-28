// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSTextBlock.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/activitydiagram3"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/mindmap"
)

// wbsTextBlock is the part that Fork and ITFComposed share: a memoized text
// block for a node, which draws the lines of the tree.
type wbsTextBlock struct {
	klimt.TextBlockDefaults
	memo         klimt.TextBlockMemoized
	skinParam    klimt.ISkinParam
	stereotype   *klimt.Stereotype
	styleBuilder *klimt.StyleBuilder
	level        int
	idea         *wElement
}

func newWBSTextBlock(skinParam klimt.ISkinParam, idea *wElement) wbsTextBlock {
	return wbsTextBlock{idea: idea, skinParam: skinParam, styleBuilder: idea.styleBuilder, level: idea.level,
		stereotype: idea.stereotype}
}

// isAutoWidth reports whether the style of the node has "Width auto".
func (t *wbsTextBlock) isAutoWidth() bool {
	return strings.EqualFold(t.idea.style().Value(klimt.PNameWidth).AsString(), "auto")
}

// styleUsed is the style of the lines below the node.
func (t *wbsTextBlock) styleUsed() *klimt.Style {
	return wbsArrow().AddLevel(t.level).AddStereotypeOf(t.stereotype).MergedStyle(t.styleBuilder)
}

// drawLine draws a line of the tree from the smaller x to the larger one.
func (t *wbsTextBlock) drawLine(ug klimt.UGraphic, x1, y1, x2, y2 float64) {
	ug = t.styleUsed().ApplyStrokeAndLineColor(ug, t.skinParam.IHtmlColorSet())
	p1 := geom.NewXPoint2D(min(x1, x2), y1)
	p2 := geom.NewXPoint2D(max(x1, x2), y2)
	line := klimt.ULineCreate(p1, p2)
	ug = ug.Apply(klimt.UTranslatePoint(p1))
	ug.Draw(line)
}

// buildMain returns the drawing of the node itself: nothing for a pseudo
// node, a box, or the text without a box.
func (t *wbsTextBlock) buildMain(idea *wElement) klimt.TextBlock {
	label := idea.label
	style := idea.style()
	if idea.shape == mindmap.IdeaShapePseudo {
		return klimt.TextBlockUtilsEmptyTextBlock()
	}
	if idea.shape == mindmap.IdeaShapeBox {
		return activitydiagram3.FtileBoxOldCreateWbs(style, idea.withBackColor(t.skinParam), label)
	}
	text := label.Create0(style.FontConfiguration(t.skinParam.IHtmlColorSet(), nil), style.HorizontalAlignment(),
		t.skinParam, style.WrapWidth(), klimt.CreoleModeFull, nil, nil)
	return klimt.TextBlockUtilsWithMargin4(text, 0, 3, 1, 1)
}

// translateOf returns the translation of ug, upstream's getTranslate of an
// AbstractCommonUGraphic; ok is false for the other graphics, such as the
// limit finder.
func translateOf(ug klimt.UGraphic) (translate klimt.UTranslate, ok bool) {
	c, ok := ug.(klimt.CommonUGraphic)
	if !ok {
		return klimt.UTranslate{}, false
	}
	return c.Common().Translate(), true
}
