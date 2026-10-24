// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/StringWithArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/LinkArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// stringWithArrow is upstream's StringWithArrow: a link label line with an
// arrow written as "<", ">", "< text", "text >"...
type stringWithArrow struct {
	label     string
	hasLabel  bool
	linkArrow abel.LinkArrow
}

func newStringWithArrow(completeLabel string) stringWithArrow {
	completeLabel = eventuallyRemoveStartingAndEndingDoubleQuote(completeLabel)
	if klimt.DisplayHasSeveralGuideLines(completeLabel) {
		return stringWithArrow{linkArrow: abel.LinkArrowNoneOrSeveral, label: completeLabel, hasLabel: true}
	}
	switch {
	case completeLabel == "<":
		return stringWithArrow{linkArrow: abel.LinkArrowBackward}
	case completeLabel == ">":
		return stringWithArrow{linkArrow: abel.LinkArrowDirectNormal}
	case strings.HasPrefix(completeLabel, "< "):
		return stringWithArrow{linkArrow: abel.LinkArrowBackward, label: trin(completeLabel[2:]), hasLabel: true}
	case strings.HasPrefix(completeLabel, "> "):
		return stringWithArrow{linkArrow: abel.LinkArrowDirectNormal, label: trin(completeLabel[2:]), hasLabel: true}
	case strings.HasSuffix(completeLabel, " >"):
		return stringWithArrow{linkArrow: abel.LinkArrowDirectNormal, label: trin(completeLabel[:len(completeLabel)-2]),
			hasLabel: true}
	case strings.HasSuffix(completeLabel, " <"):
		return stringWithArrow{linkArrow: abel.LinkArrowBackward, label: trin(completeLabel[:len(completeLabel)-2]),
			hasLabel: true}
	}
	return stringWithArrow{linkArrow: abel.LinkArrowNoneOrSeveral, label: completeLabel, hasLabel: true}
}

// eventuallyRemoveStartingAndEndingDoubleQuote is the StringUtils method
// with the format "\"": it removes one pair of surrounding double quotes.
func eventuallyRemoveStartingAndEndingDoubleQuote(s string) string {
	r := []rune(s)
	if len(r) > 1 && isDoubleQuote(r[0]) && isDoubleQuote(r[len(r)-1]) {
		return string(r[1 : len(r)-1])
	}
	return s
}

func isDoubleQuote(c rune) bool {
	return c == '"' || c == '“' || c == '”' || c == '«' || c == '»'
}

// trin is StringUtils.trin.
func trin(s string) string { return text.Trin(s) }

func (s stringWithArrow) display(pragma *klimt.Pragma) *klimt.Display {
	return klimt.DisplayWithNewlinesString(pragma, s.label, s.hasLabel)
}

// addMagicArrow puts the arrow of the guide line before the label.
func addMagicArrow(label klimt.TextBlock, guide klimt.GuideLine, fc *klimt.FontConfiguration) klimt.TextBlock {
	arrow := klimt.NewTextBlockArrow2(guide, fc)
	return klimt.TextBlockUtilsMergeLR(arrow, label, geom.VerticalAlignmentCenter)
}

// addSeveralMagicArrows draws each line of a label with its own arrow.
func addSeveralMagicArrows(label *klimt.Display, guide klimt.GuideLine, fc *klimt.FontConfiguration,
	alignment geom.HorizontalAlignment, skinParam klimt.ISkinParam) klimt.TextBlock {
	result := klimt.TextBlockUtilsEmptyTextBlock()
	for _, cs := range label.AsList() {
		tmp := newStringWithArrow(cs.String())
		block := tmp.display(skinParam.Pragma()).Create9(fc, alignment, skinParam, skinParam.MaxMessageSize())
		if tmp.linkArrow != abel.LinkArrowNoneOrSeveral {
			block = addMagicArrow(block, muteGuideLine(tmp.linkArrow, guide), fc)
		}
		result = klimt.TextBlockUtilsMergeTB(result, block, alignment)
	}
	return result
}

// muteGuideLine is LinkArrow.mute.
func muteGuideLine(a abel.LinkArrow, guide klimt.GuideLine) klimt.GuideLine {
	switch a {
	case abel.LinkArrowDirectNormal:
		return guide
	case abel.LinkArrowBackward:
		return backwardGuideLine{guide}
	}
	throwUnsupported("")
	return nil
}

type backwardGuideLine struct{ guide klimt.GuideLine }

func (b backwardGuideLine) ArrowDirection() klimt.Direction { return b.guide.ArrowDirection().Inv() }

func (b backwardGuideLine) ArrowDirectionInRadian() float64 {
	return math.Pi + b.guide.ArrowDirectionInRadian()
}
