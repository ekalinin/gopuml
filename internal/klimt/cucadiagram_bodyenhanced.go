// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyEnhanced2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyEnhancedAbstract.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// This file holds the part of upstream's cucadiagram package that the skin
// components need (AbstractTextualComponent draws enhanced notes with it):
// BodyFactory.create3, BodyEnhanced2 and BodyEnhancedAbstract. The class
// bodies (BodyEnhanced1, Bodier) belong to the class diagram step.

// BodyFactoryCreate3 is upstream's BodyFactory.create3: a text whose
// separator lines ("--", "==", "..", "__", possibly with a title between the
// marks) split it into blocks with a line before each.
func BodyFactoryCreate3(rawBody *Display, skinParam ISkinSimple, align geom.HorizontalAlignment,
	fc *FontConfiguration, lineBreakStrategy LineBreakStrategy, style *Style) TextBlock {
	return &BodyEnhanced2{rawBody: rawBody, skinParam: skinParam, align: align, titleConfig: fc,
		lineBreakStrategy: lineBreakStrategy, style: style}
}

// BodyEnhancedIsBlockSeparator reports whether a line separates blocks:
// it starts and ends with "--", "==", ".." (but is not "...") or "__".
func BodyEnhancedIsBlockSeparator(cs string) bool {
	switch {
	case strings.HasPrefix(cs, "--") && strings.HasSuffix(cs, "--"):
		return true
	case strings.HasPrefix(cs, "==") && strings.HasSuffix(cs, "=="):
		return true
	case strings.HasPrefix(cs, "..") && strings.HasSuffix(cs, "..") && cs != "...":
		return true
	case strings.HasPrefix(cs, "__") && strings.HasSuffix(cs, "__"):
		return true
	}
	return false
}

// BodyEnhanced2 is upstream's BodyEnhanced2 (with BodyEnhancedAbstract).
type BodyEnhanced2 struct {
	TextBlockDefaults
	align             geom.HorizontalAlignment
	titleConfig       *FontConfiguration
	area              TextBlock
	style             *Style
	rawBody           *Display
	skinParam         ISkinSimple
	lineBreakStrategy LineBreakStrategy
}

// CalculateDimension returns the size of the blocks.
func (b *BodyEnhanced2) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.getArea(stringBounder).CalculateDimension(stringBounder)
}

// DrawU draws the blocks.
func (b *BodyEnhanced2) DrawU(ug UGraphic) { b.getArea(ug.StringBounder()).DrawU(ug) }

// Style returns the style.
func (b *BodyEnhanced2) Style() *Style { return b.style }

func (b *BodyEnhanced2) title(s string, spriteContainer ISkinSimple) TextBlock {
	u := utf16Of(s)
	if len(u) <= 4 {
		return nil
	}
	s = trin(stringOf(u[2 : len(u)-2]))
	return DisplayWithNewlines(spriteContainer.Pragma(), s).Create(b.titleConfig, geom.HorizontalAlignmentLeft,
		spriteContainer)
}

func (b *BodyEnhanced2) decorate(block TextBlock, separator rune, title TextBlock,
	stringBounder font.StringBounder) TextBlock {
	marginX := 0.0
	if separator == 0 {
		return TextBlockUtilsWithMargin(block, marginX, 0)
	}
	if title == nil {
		return NewTextBlockLineBefore(b.defaultThickness(), TextBlockUtilsWithMargin(block, marginX, 4), separator,
			nil)
	}
	dimTitle := title.CalculateDimension(stringBounder)
	raw := NewTextBlockLineBefore(b.defaultThickness(),
		TextBlockUtilsWithMargin4(block, marginX, 6, dimTitle.Height()/2, 4), separator, title)
	return TextBlockUtilsWithMargin4(raw, 0, 0, dimTitle.Height()/2, 0)
}

func (b *BodyEnhanced2) defaultThickness() float64 {
	return b.style.Value(PNameLineThickness).AsDouble()
}

func (b *BodyEnhanced2) getArea(stringBounder font.StringBounder) TextBlock {
	if b.area != nil {
		return b.area
	}
	var blocks []TextBlock
	var separator rune
	var title TextBlock
	display := DisplayEmpty()
	lines := b.rawBody.AsList()
	for i := 0; i < len(lines); i++ {
		s := lines[i]
		if _, ok := EmbeddedDiagramType(s.String()); ok {
			display = display.Add(s)
			display, i = bodyEnhancedAddOneSingleLineManageEmbedded2(lines, i+1, display)
		} else if BodyEnhancedIsBlockSeparator(s.String()) {
			blocks = append(blocks, b.decorate(b.textBlock(display), separator, title, stringBounder))
			separator = javaCharAt(s.String(), 0)
			title = b.title(s.String(), b.skinParam)
			display = DisplayEmpty()
		} else {
			display = display.Add(s)
		}
	}
	blocks = append(blocks, b.decorate(b.textBlock(display), separator, title, stringBounder))
	if len(blocks) == 1 {
		b.area = blocks[0]
	} else {
		b.area = NewTextBlockVertical(blocks, b.align)
	}
	if minClassWidth := b.style.Value(PNameMinimumWidth).AsDouble(); minClassWidth > 0 {
		b.area = TextBlockUtilsWithMinWidth(b.area, minClassWidth, b.align)
	}
	return b.area
}

// bodyEnhancedAddOneSingleLineManageEmbedded2 adds the lines of an embedded
// diagram up to its closing "}}"; it returns the display and the index of
// the last line read.
func bodyEnhancedAddOneSingleLineManageEmbedded2(lines []CharSequence, start int, display *Display) (*Display, int) {
	nested := 1
	i := start
	for ; i < len(lines); i++ {
		s := lines[i]
		display = display.Add(s)
		if _, ok := EmbeddedDiagramType(s.String()); ok {
			nested++
		} else if trin(s.String()) == EmbeddedDiagramEnd {
			nested--
			if nested == 0 {
				return display, i
			}
		}
	}
	return display, i - 1
}

func (b *BodyEnhanced2) textBlock(display *Display) TextBlock {
	return display.Create9(b.titleConfig, b.align, b.skinParam, b.lineBreakStrategy)
}
