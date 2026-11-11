// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/TextBlockJson.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleQueries.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/HorizontalAlignment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const (
	textBlockJSONMinWidth  = 30
	textBlockJSONMinHeight = 15
)

// diagramSName is the style name of the diagram type: jsonDiagram for JSON
// diagrams, yamlDiagram for the others, as upstream.
func diagramSName(t core.DiagramType) klimt.SName {
	if t == core.JSON {
		return klimt.SNameJsonDiagram
	}
	return klimt.SNameYamlDiagram
}

// styleQuery is StyleQueries.ELEMENT with names, then the style name of the
// diagram type, as upstream's StyleQueries.NODE.add(diagramType) and the
// others.
func styleQuery(diagramType klimt.SName, names ...klimt.SName) klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(append(append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement}, names...),
		diagramType)...)
}

// jsonLine is a row of a box: the key and the value of an object member,
// or the value of an array item, and the highlight path of the row.
type jsonLine struct {
	b1, b2      klimt.TextBlock // b2 is nil for an array item
	highlighted *Highlighted
}

func (l *jsonLine) heightOfRow(stringBounder font.StringBounder) float64 {
	height := l.b1.CalculateDimension(stringBounder).Height()
	if l.b2 == nil {
		return height
	}
	return max(height, l.b2.CalculateDimension(stringBounder).Height())
}

// TextBlockJson is the box of an object or an array: one row per member or
// item, and the highlighted rows.
type TextBlockJson struct {
	klimt.TextBlockDefaults
	memo klimt.TextBlockMemoized

	lines       []*jsonLine
	skinParam   klimt.ISkinParam
	totalWidth  float64
	root        json.Value
	diagramType klimt.SName
}

// NewTextBlockJson returns the box of root, whose rows allHighlighteds may
// name.
func NewTextBlockJson(skinParam klimt.ISkinParam, diagramType klimt.SName, root json.Value,
	allHighlighteds []*Highlighted) *TextBlockJson {
	t := &TextBlockJson{skinParam: skinParam, root: root, diagramType: diagramType}
	switch {
	case root.IsObject():
		for key, member := range root.AsObject().All() {
			highlighted := isHighlighted(key, allHighlighteds)
			block1 := t.textBlock(t.styleToUse(true, highlighted), key)
			block2 := t.textBlock(t.styleToUse(false, highlighted), shortString(member))
			t.lines = append(t.lines, &jsonLine{b1: block1, b2: block2, highlighted: highlighted})
		}
	case root.IsArray():
		for i, value := range root.AsArray().All() {
			highlighted := isHighlighted(strconv.Itoa(i), allHighlighteds)
			block2 := t.textBlock(t.styleToUse(false, highlighted), shortString(value))
			t.lines = append(t.lines, &jsonLine{b1: block2, highlighted: highlighted})
		}
	}
	return t
}

func (t *TextBlockJson) styleToUse(header bool, highlighted *Highlighted) *klimt.Style {
	var query klimt.StyleSignatureBasic
	switch {
	case header && highlighted != nil:
		query = styleQuery(t.diagramType, klimt.SNameHeader, klimt.SNameNode, klimt.SNameHighlight).
			AddStereotypeOf(highlighted.Stereotype())
	case highlighted != nil:
		query = styleQuery(t.diagramType, klimt.SNameNode, klimt.SNameHighlight).AddStereotypeOf(highlighted.Stereotype())
	case header:
		query = styleQuery(t.diagramType, klimt.SNameHeader, klimt.SNameNode)
	default:
		query = styleQuery(t.diagramType, klimt.SNameNode)
	}
	return query.MergedStyle(t.skinParam.CurrentStyleBuilder())
}

func isHighlighted(key string, highlighted []*Highlighted) *Highlighted {
	for _, h := range highlighted {
		if h.isKeyHighlight(key) {
			return h
		}
	}
	return nil
}

// shortString is the text of a value in its row: three spaces for an
// object or an array, whose box is linked from the row.
func shortString(value json.Value) string {
	switch {
	case value.IsString():
		return value.AsString()
	case value.IsNull():
		return "<U+2400>"
	case value.IsNumber():
		return value.String()
	case value.IsTrue():
		return "<U+2611> true"
	case value.IsBoolean():
		return "<U+2610> false"
	}
	return "   "
}

// children returns the value of each row whose value is an object or an
// array, nil for the other rows.
func (t *TextBlockJson) children() []json.Value {
	var result []json.Value
	add := func(value json.Value) {
		if value.IsObject() || value.IsArray() {
			result = append(result, value)
		} else {
			result = append(result, nil)
		}
	}
	switch {
	case t.root.IsObject():
		for _, value := range t.root.AsObject().All() {
			add(value)
		}
	case t.root.IsArray():
		for _, value := range t.root.AsArray().All() {
			add(value)
		}
	}
	return result
}

// keys returns the key of each row, the index of an array item.
func (t *TextBlockJson) keys() []string {
	var result []string
	switch {
	case t.root.IsObject():
		for key := range t.root.AsObject().All() {
			result = append(result, key)
		}
	case t.root.IsArray():
		for i := range t.root.AsArray().All() {
			result = append(result, strconv.Itoa(i))
		}
	}
	return result
}

// CalculateDimension returns the width of the two columns and the height of
// the rows, 15 points without rows.
func (t *TextBlockJson) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(t.widthColA(sb)+t.widthColB(sb), t.totalHeight(sb))
	})
}

func (t *TextBlockJson) widthColA(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, line := range t.lines {
		width = max(width, line.b1.CalculateDimension(stringBounder).Width())
	}
	return width
}

func (t *TextBlockJson) widthColB(stringBounder font.StringBounder) float64 {
	width := 0.0
	for _, line := range t.lines {
		if line.b2 != nil {
			width = max(width, line.b2.CalculateDimension(stringBounder).Width())
		}
	}
	return width
}

// DrawU draws the box, the highlighted rows, the lines between the rows and
// between the columns, the texts, then the border of the box.
func (t *TextBlockJson) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	set := t.skinParam.IHtmlColorSet()

	fullDim := t.CalculateDimension(stringBounder)
	trueWidth := max(fullDim.Width(), t.totalWidth)
	widthColA := t.widthColA(stringBounder)
	widthColB := t.widthColB(stringBounder)

	y := 0.0
	styleNode := styleQuery(t.diagramType, klimt.SNameNode).MergedStyle(t.skinParam.CurrentStyleBuilder())
	ugNode := styleNode.ApplyStrokeAndLineColor(ug, set)
	for _, line := range t.lines {
		y += line.heightOfRow(stringBounder)
	}
	if y == 0 {
		y = textBlockJSONMinHeight
	}
	if trueWidth == 0 {
		trueWidth = textBlockJSONMinWidth
	}

	round := styleNode.Value(klimt.PNameRoundCorner).AsDouble()
	fullNodeRectangle := klimt.URectangleBuild(trueWidth, y).Rounded(round)
	backColor := styleNode.Value(klimt.PNameBackGroundColor).AsColor(set)
	ugNode.Apply(klimt.Bg(backColor)).Apply(backColor).Draw(fullNodeRectangle)

	styleSeparator := styleQuery(t.diagramType, klimt.SNameNode, klimt.SNameSeparator).
		MergedStyle(t.skinParam.CurrentStyleBuilder())
	ugSeparator := styleSeparator.ApplyStrokeAndLineColor(ug, set)

	y = 0
	for _, line := range t.lines {
		ugline := ugSeparator.Apply(klimt.UTranslateDy(y))
		heightOfRow := line.heightOfRow(stringBounder)
		if line.highlighted != nil {
			back := klimt.URectangleBuild(trueWidth-2, heightOfRow).Rounded(4)
			styleNodeHighlight := styleQuery(t.diagramType, klimt.SNameNode, klimt.SNameHighlight).
				AddStereotypeOf(line.highlighted.Stereotype()).MergedStyle(t.skinParam.CurrentStyleBuilder())
			cellBackColor := styleNodeHighlight.Value(klimt.PNameBackGroundColor).AsColor(set)
			ugline.Apply(cellBackColor).Apply(klimt.Bg(cellBackColor)).Apply(klimt.NewUTranslate(1.5, 0)).Draw(back)
		}

		if y > 0 {
			ugline.Draw(klimt.ULineH(trueWidth))
		}

		horizontalAlignment := styleNode.HorizontalAlignment()
		drawAligned(horizontalAlignment, ugline, line.b1, 0, 0, widthColA)

		if line.b2 != nil {
			uglineColB := ugline.Apply(klimt.UTranslateDx(widthColA))
			drawAligned(horizontalAlignment, uglineColB, line.b2, 0, 0, widthColB)
			uglineColB.Draw(klimt.ULineV(heightOfRow))
		}

		y += heightOfRow
	}
	ugNode.Draw(fullNodeRectangle)
}

func (t *TextBlockJson) totalHeight(stringBounder font.StringBounder) float64 {
	height := 0.0
	for _, line := range t.lines {
		height += line.heightOfRow(stringBounder)
	}
	if height == 0 {
		return textBlockJSONMinHeight
	}
	return height
}

// allHeights returns the height of each row.
func (t *TextBlockJson) allHeights(stringBounder font.StringBounder) []float64 {
	result := make([]float64, len(t.lines))
	for i, line := range t.lines {
		result[i] = line.heightOfRow(stringBounder)
	}
	return result
}

func (t *TextBlockJson) textBlock(style *klimt.Style, key string) klimt.TextBlock {
	display := klimt.DisplayWithNewlines(t.skinParam.Pragma(), key)
	fontConfiguration := style.FontConfiguration(t.skinParam.IHtmlColorSet(), nil)
	wrap := style.WrapWidth()
	horizontalAlignment := style.HorizontalAlignment()
	result := display.Create0(fontConfiguration, horizontalAlignment, t.skinParam, wrap, klimt.CreoleModeNoCreole,
		nil, nil)
	return klimt.TextBlockUtilsWithMargin(result, 5, 2)
}

// SetTotalWidth sets the width of the box, when it is wider than its rows.
func (t *TextBlockJson) SetTotalWidth(totalWidth float64) { t.totalWidth = totalWidth }

// drawAligned is upstream's HorizontalAlignment.draw: tb drawn at
// ypadding, aligned in width with xpadding.
func drawAligned(alignment geom.HorizontalAlignment, ug klimt.UGraphic, tb klimt.TextBlock,
	xpadding, ypadding, width float64) {
	switch alignment {
	case geom.HorizontalAlignmentLeft:
		tb.DrawU(ug.Apply(klimt.NewUTranslate(xpadding, ypadding)))
	case geom.HorizontalAlignmentRight:
		dimTb := tb.CalculateDimension(ug.StringBounder())
		tb.DrawU(ug.Apply(klimt.NewUTranslate(width-dimTb.Width()-xpadding, ypadding)))
	case geom.HorizontalAlignmentCenter:
		dimTb := tb.CalculateDimension(ug.StringBounder())
		tb.DrawU(ug.Apply(klimt.NewUTranslate((width-dimTb.Width())/2, ypadding)))
	}
}
