// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyEnhanced1.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyEnhancedAbstract.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/Body3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// BodyFactoryCreate1 is upstream's BodyFactory.create1: the body of the raw
// lines of an entity, a separator line before its first block.
func BodyFactoryCreate1(align geom.HorizontalAlignment, rawBody []klimt.CharSequence, skinParam klimt.ISkinParam,
	stereotype *klimt.Stereotype, entity *Entity, style *klimt.Style) klimt.TextBlock {
	return newBodyEnhanced1Raw(align, rawBody, skinParam, entity, style)
}

// BodyFactoryCreate2 is upstream's BodyFactory.create2: the body of a
// display, as the description elements draw it.
func BodyFactoryCreate2(align geom.HorizontalAlignment, display *klimt.Display, skinParam klimt.ISkinParam,
	stereotype *klimt.Stereotype, entity *Entity, style *klimt.Style) klimt.TextBlock {
	return newBodyEnhanced1Display(align, display, skinParam, entity, style)
}

// BodyEnhanced1 is upstream's BodyEnhanced1 (with BodyEnhancedAbstract): a
// body split into blocks by the separator lines ("--", "==", "..", "__",
// possibly with a title), the members of each block in a
// MethodsOrFieldsArea, the creole trees and tables as their own blocks.
type BodyEnhanced1 struct {
	klimt.TextBlockDefaults

	align       geom.HorizontalAlignment
	titleConfig *klimt.FontConfiguration
	area        klimt.TextBlock
	style       *klimt.Style

	rawBody2  *klimt.Display
	skinParam klimt.ISkinParam
	lineFirst bool
	urls      []*klimt.URL
	entity    *Entity
	inEllipse bool
}

var _ klimt.WithPorts = (*BodyEnhanced1)(nil)

func newBodyEnhanced1Raw(align geom.HorizontalAlignment, rawBody []klimt.CharSequence, skinParam klimt.ISkinParam,
	entity *Entity, style *klimt.Style) *BodyEnhanced1 {
	return &BodyEnhanced1{
		align:       align,
		titleConfig: style.FontConfiguration(skinParam.IHtmlColorSet(), entity.Colors()),
		style:       style,
		rawBody2:    klimt.DisplayCreateSeq(rawBody),
		skinParam:   skinParam,
		lineFirst:   true,
		entity:      entity,
	}
}

func newBodyEnhanced1Display(align geom.HorizontalAlignment, display *klimt.Display, skinParam klimt.ISkinParam,
	entity *Entity, style *klimt.Style) *BodyEnhanced1 {
	b := &BodyEnhanced1{
		align:       align,
		titleConfig: style.FontConfiguration(skinParam.IHtmlColorSet(), entity.Colors()),
		style:       style,
		entity:      entity,
		skinParam:   skinParam,
	}
	leafType := entity.LeafType()
	b.inEllipse = leafType == LeafTypeUsecase || leafType == LeafTypeUsecaseBusiness
	if b.inEllipse && display.Size() > 0 && klimt.BodyEnhancedIsBlockSeparator(display.Get(0).String()) {
		display = display.Add(klimt.Text(""))
	}
	b.rawBody2 = display
	return b
}

func (b *BodyEnhanced1) marginX() float64 { return 6 }

// Style returns the style of the body.
func (b *BodyEnhanced1) Style() *klimt.Style { return b.style }

// CalculateDimension returns the size of the blocks.
func (b *BodyEnhanced1) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.getArea(stringBounder).CalculateDimension(stringBounder)
}

// DrawU draws the blocks.
func (b *BodyEnhanced1) DrawU(ug klimt.UGraphic) { b.getArea(ug.StringBounder()).DrawU(ug) }

func (b *BodyEnhanced1) title(s string, spriteContainer klimt.ISkinSimple) klimt.TextBlock {
	u := utf16.Encode([]rune(s))
	if len(u) <= 4 {
		return nil
	}
	s = text.Trin(string(utf16.Decode(u[2 : len(u)-2])))
	return klimt.DisplayWithNewlines(spriteContainer.Pragma(), s).Create(b.titleConfig,
		geom.HorizontalAlignmentLeft, spriteContainer)
}

func (b *BodyEnhanced1) defaultThickness() float64 {
	return b.style.Value(klimt.PNameLineThickness).AsDouble()
}

func (b *BodyEnhanced1) decorate(block klimt.TextBlock, separator rune, title klimt.TextBlock,
	stringBounder font.StringBounder) klimt.TextBlock {
	marginX := b.marginX()
	if separator == 0 {
		return klimt.TextBlockUtilsWithMargin(block, marginX, 0)
	}
	if title == nil {
		return klimt.NewTextBlockLineBefore(b.defaultThickness(), klimt.TextBlockUtilsWithMargin(block, marginX, 4),
			separator, nil)
	}
	dimTitle := title.CalculateDimension(stringBounder)
	raw := klimt.NewTextBlockLineBefore(b.defaultThickness(),
		klimt.TextBlockUtilsWithMargin4(block, marginX, 6, dimTitle.Height()/2, 4), separator, title)
	return klimt.TextBlockUtilsWithMargin4(raw, 0, 0, dimTitle.Height()/2, 0)
}

func bodyIsTreeOrTable(s string) bool {
	s = text.Trin(s)
	return klimt.ParserIsTreeStart(s) || klimt.CreoleParserIsTableLine(s)
}

func (b *BodyEnhanced1) getArea(stringBounder font.StringBounder) klimt.TextBlock {
	if b.area != nil {
		return b.area
	}
	b.urls = nil
	var blocks []klimt.TextBlock
	var separator rune
	if b.lineFirst {
		separator = '_'
	}
	var title klimt.TextBlock
	var display *klimt.Display
	lines := b.rawBody2.AsList()
	for i := 0; i < len(lines); i++ {
		cs := lines[i]
		s := cs.String()
		switch {
		case klimt.BodyEnhancedIsBlockSeparator(s):
			if display == nil {
				display = klimt.DisplayEmpty()
			}
			blocks = append(blocks, b.buildTextBlock(display, separator, title, stringBounder))
			separator = rune(utf16.Encode([]rune(s))[0])
			title = b.title(s, b.skinParam)
			display = nil
		case bodyIsTreeOrTable(s):
			isTable := klimt.CreoleParserIsTableLine(s)
			if display == nil {
				display = klimt.DisplayEmpty()
			}
			blocks = append(blocks, b.buildTextBlock(display, separator, title, stringBounder))
			separator = 0
			title = nil
			display = nil
			var allTree []klimt.CharSequence
			allTree, i = buildTreeOrTable(s, lines, i)
			fontConfiguration := b.style.FontConfiguration(b.skinParam.IHtmlColorSet(), nil)
			block := klimt.DisplayCreateSeq(allTree).Create7(fontConfiguration, b.align, b.skinParam,
				klimt.CreoleModeFull)
			if isTable {
				block = klimt.TextBlockUtilsWithMargin4(block, 10, 10, 0, 5)
			}
			blocks = append(blocks, block)
		default:
			if display == nil {
				display = klimt.DisplayEmpty()
			}
			display = display.Add(cs)
			if m, ok := cs.(*Member); ok && m.URL() != nil {
				b.urls = append(b.urls, m.URL())
			}
		}
	}
	if display == nil {
		display = klimt.DisplayEmpty()
	}
	if b.inEllipse && display.Size() == 0 {
		display = display.Add(klimt.Text(""))
	}
	blocks = append(blocks, b.buildTextBlock(display, separator, title, stringBounder))
	if len(blocks) == 1 {
		b.area = blocks[0]
	} else {
		b.area = klimt.NewTextBlockVertical(blocks, b.align)
	}
	if minClassWidth := b.style.Value(klimt.PNameMinimumWidth).AsDouble(); minClassWidth > 0 {
		b.area = klimt.TextBlockUtilsWithMinWidth(b.area, minClassWidth, b.align)
	}
	return b.area
}

func (b *BodyEnhanced1) buildTextBlock(display *klimt.Display, separator rune, title klimt.TextBlock,
	stringBounder font.StringBounder) klimt.TextBlock {
	var result klimt.TextBlock = NewMethodsOrFieldsAreaAligned(display, b.skinParam, b.align, b.entity, b.style)
	return b.decorate(result, separator, title, stringBounder)
}

// bodyLeadingSpaces is upstream's BodyEnhanced1.p: the leading Java white
// space of a line.
var bodyLeadingSpaces = regexp.MustCompile(`^([\t\n\x0B\f\r ]+)`)

// buildTreeOrTable returns the lines of the tree or table starting at line
// index (whose text is init), without the indentation of its first line,
// and the index of its last line.
func buildTreeOrTable(init string, lines []klimt.CharSequence, index int) ([]klimt.CharSequence, int) {
	start := ""
	if m := bodyLeadingSpaces.FindStringSubmatch(init); m != nil {
		start = m[1]
	}
	result := []klimt.CharSequence{klimt.Text(purgeStart(init, start))}
	for index+1 < len(lines) {
		s := lines[index+1].String()
		if !bodyIsTreeOrTable(s) {
			return result, index
		}
		result = append(result, klimt.Text(purgeStart(s, start)))
		index++
	}
	return result, index
}

func purgeStart(s, start string) string {
	if strings.HasPrefix(s, start) {
		return s[len(start):]
	}
	return s
}

// Ports implements klimt.WithPorts: the ports of the blocks.
func (b *BodyEnhanced1) Ports(stringBounder font.StringBounder) *klimt.Ports {
	if wp, ok := b.getArea(stringBounder).(klimt.WithPorts); ok {
		return wp.Ports(stringBounder)
	}
	return klimt.NewPorts()
}

// URLs returns the links of the members, known once the body is measured
// or drawn.
func (b *BodyEnhanced1) URLs() []*klimt.URL { return b.urls }

// InnerPosition returns the position of a member in the blocks.
func (b *BodyEnhanced1) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return b.getArea(stringBounder).InnerPosition(member, stringBounder)
}

// Body3 is upstream's Body3, an alternative body that upstream never uses
// (BodyFactory.BODY3 is false): the raw lines with their visibility
// characters replaced by icons characters.
type Body3 struct {
	klimt.TextBlockDefaults
	memo      klimt.TextBlockMemoized
	rawBody   []string
	skinParam klimt.ISkinParam
	style     *klimt.Style
}

var _ klimt.WithPorts = (*Body3)(nil)

// NewBody3 returns the body of the raw lines.
func NewBody3(rawBody []klimt.CharSequence, skinParam klimt.ISkinParam, style *klimt.Style) *Body3 {
	b := &Body3{skinParam: skinParam, style: style}
	for _, s := range rawBody {
		b.rawBody = append(b.rawBody, klimt.VisibilityReplaceVisibilityModifierByUnicodeChar(s.String(), true))
	}
	return b
}

// DrawU draws the lines.
func (b *Body3) DrawU(ug klimt.UGraphic) { b.textBlock().DrawU(ug) }

func (b *Body3) textBlock() klimt.TextBlock {
	display := klimt.DisplayCreate(b.rawBody...)
	config := klimt.FontConfigurationFromStyle(b.skinParam, b.style, nil)
	return display.Create(config, geom.HorizontalAlignmentLeft, b.skinParam)
}

// Ports implements klimt.WithPorts: none.
func (b *Body3) Ports(font.StringBounder) *klimt.Ports { return klimt.NewPorts() }

// CalculateDimension returns the size of the lines, memoized.
func (b *Body3) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return b.textBlock().CalculateDimension(sb)
	})
}
