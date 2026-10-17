// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/Style.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/Colors.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/LinkStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"fmt"
	"os"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// styleValues holds the properties of a style in PName order (upstream's
// EnumMap); nil is a missing property.
type styleValues [pnameCount]Value

// Style is a set of property values for a signature. The methods return
// new styles. As upstream, the styles a style sheet parser returns share
// the values of their parser context until the parsing ends.
type Style struct {
	values    *styleValues
	signature StyleSignatureBasic
}

// NewStyle returns the style of a signature with the given values.
func NewStyle(signature StyleSignatureBasic, values map[PName]Value) *Style {
	s := &Style{signature: signature, values: &styleValues{}}
	for k, v := range values {
		s.values[k] = v
	}
	return s
}

// newStyleValues returns a style with a copy of values.
func newStyleValues(signature StyleSignatureBasic, values styleValues) *Style {
	return &Style{signature: signature, values: &values}
}

// newStyleShared returns a style holding values themselves, as upstream's
// Style keeps the map it is given.
func newStyleShared(signature StyleSignatureBasic, values *styleValues) *Style {
	return &Style{signature: signature, values: values}
}

// DeltaPriority returns the starred style with the priority of its values
// increased by delta. A style that is not starred panics, as upstream throws
// UnsupportedOperationException.
func (s *Style) DeltaPriority(delta int) *Style {
	if !s.signature.IsStarred() {
		throw(unsupportedOperation, "")
	}
	var cp styleValues
	for k, v := range s.values {
		if v != nil {
			cp[k] = asValueImpl(v).AddPriority(delta)
		}
	}
	return newStyleValues(s.signature, cp)
}

// asValueImpl is upstream's cast to ValueImpl: other values panic with
// ClassCastException.
func asValueImpl(v Value) *ValueImpl {
	vi, ok := v.(*ValueImpl)
	if !ok {
		throw("java.lang.ClassCastException", fmt.Sprintf("%T cannot be cast to ValueImpl", v))
	}
	return vi
}

// PrintMe writes the style to the standard error, as upstream does.
func (s *Style) PrintMe() {
	if s.size() == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, s.signature.String()+" {")
	for k, v := range s.values {
		if v != nil {
			fmt.Fprintln(os.Stderr, "  "+PName(k).String()+": "+v.AsString())
		}
	}
	fmt.Fprintln(os.Stderr, "}")
}

func (s *Style) size() int {
	n := 0
	for _, v := range s.values {
		if v != nil {
			n++
		}
	}
	return n
}

// String is upstream's toString: the signature and the values.
func (s *Style) String() string {
	parts := make([]string, 0, s.size())
	for k, v := range s.values {
		if v != nil {
			parts = append(parts, PName(k).String()+"="+v.String())
		}
	}
	return s.signature.String() + " {" + strings.Join(parts, ", ") + "}"
}

// Value returns the value of a property, ValueNull when it is not set.
func (s *Style) Value(name PName) Value {
	if v := s.values[name]; v != nil {
		return v
	}
	return valueNull
}

// Shadowing returns the shadow size: 0 without the property.
func (s *Style) Shadowing() float64 {
	if s.values[PNameShadowing] == nil {
		return 0
	}
	return s.Value(PNameShadowing).AsDoubleDefaultTo(1.5)
}

// HasValue reports whether a property is set.
func (s *Style) HasValue(name PName) bool { return s.values[name] != nil }

// MergeWith returns the style with the values of other merged in (other
// nil gives s). With MergeKeepExistingValueOfStereotype, the values of s
// that come from a stereotype style are kept.
func (s *Style) MergeWith(other *Style, strategy MergeStrategy) *Style {
	if other == nil {
		return s
	}
	both := *s.values
	for k, v := range other.values {
		if v == nil {
			continue
		}
		previous := s.values[k]
		if previous != nil && previous.Priority() > StyleLoaderDeltaPriorityForStereotype &&
			strategy == MergeKeepExistingValueOfStereotype {
			continue
		}
		both[k] = asValueImpl(v).MergeWith(previous)
	}
	return newStyleValues(s.signature.MergeWith(other.Signature()), both)
}

// MergeNestedChildOver layers the style of a nested selector (for example
// "group.header") over s, keeping only the properties whose value differs
// from the nested selector's parent (nestedParent, which may be nil): the
// properties it only inherits do not overwrite those of s.
func (s *Style) MergeNestedChildOver(nested, nestedParent *Style, strategy MergeStrategy) *Style {
	if nested == nil {
		return s
	}
	var divergentOnly styleValues
	empty := true
	for k, v := range nested.values {
		if v == nil {
			continue
		}
		var ancestorValue Value
		if nestedParent != nil {
			ancestorValue = nestedParent.values[k]
		}
		if ancestorValue == nil {
			divergentOnly[k] = v
			empty = false
			continue
		}
		ancestor := valueStringNonNull(ancestorValue)
		if nestedString, ok := valueString(v); !ok || ancestor != nestedString {
			divergentOnly[k] = v
			empty = false
		}
	}
	if empty {
		return s
	}
	return s.MergeWith(newStyleValues(nested.signature, divergentOnly), strategy)
}

// EventuallyOverrideColor returns the style with a color property replaced
// by c, keeping the priority; a nil color gives s. Upstream reads the
// priority of the old value, which must be set.
func (s *Style) EventuallyOverrideColor(param PName, c color.HColor) *Style {
	if c == nil {
		return s
	}
	result := *s.values
	old := result[param]
	if old == nil {
		throw(nullPointer, "")
	}
	result[param] = NewValueColor(c, old.Priority())
	return newStyleValues(s.signature, result)
}

// EventuallyOverrideDouble returns the style with a property set to the
// text of a double, with the highest priority.
func (s *Style) EventuallyOverrideDouble(param PName, value float64) *Style {
	return s.EventuallyOverrideString(param, jstr(value))
}

// EventuallyOverrideString returns the style with a property set to value,
// with the highest priority.
func (s *Style) EventuallyOverrideString(param PName, value string) *Style {
	result := *s.values
	result[param] = ValueImplRegularPriority(value, maxInt32)
	return newStyleValues(s.signature, result)
}

const maxInt32 = 1<<31 - 1

// EventuallyOverrideColors returns the style with the back, line and text
// colors of colors (which may be nil).
func (s *Style) EventuallyOverrideColors(colors *color.Colors) *Style {
	result := s
	if colors != nil {
		if back := colors.Color(color.TypeBack); back != nil {
			result = result.EventuallyOverrideColor(PNameBackGroundColor, back)
		}
		if line := colors.Color(color.TypeLine); line != nil {
			result = result.EventuallyOverrideColor(PNameLineColor, line)
		}
		if text := colors.Color(color.TypeText); text != nil {
			result = result.EventuallyOverrideColor(PNameFontColor, text)
		}
	}
	return result
}

// EventuallyOverrideFashion returns the style with the background color of
// symbolContext (which may be nil).
func (s *Style) EventuallyOverrideFashion(symbolContext *Fashion) *Style {
	result := s
	if symbolContext != nil {
		if back := symbolContext.BackColor(); back != nil {
			result = result.EventuallyOverrideColor(PNameBackGroundColor, back)
		}
	}
	return result
}

// Signature returns the signature.
func (s *Style) Signature() StyleSignatureBasic { return s.signature }

// UFont returns the font of the FontName, FontStyle, FontWeight and
// FontSize properties: FontStyle gives the face, a FontWeight other than 400
// replaces its weight, and the size defaults to 14. A FontName set only for
// the dark mode gives the font of an empty name, where upstream builds a font
// of a null name that fails when measured.
func (s *Style) UFont() font.UFont {
	fontName := s.Value(PNameFontName).AsString()
	size := s.Value(PNameFontSize).AsIntButMinusOneIfError()
	if size == -1 {
		size = 14
	}
	face := s.Value(PNameFontStyle).AsFontFace()
	weightFace := s.Value(PNameFontWeight).AsFontFace()
	if weightFace.CssWeight() != 400 {
		face = face.WithWeight(weightFace.CssWeight())
	}
	return font.Build(fontName, face, size)
}

// FontConfiguration returns the font configuration of the style, the text
// color of colors (which may be nil) taking precedence.
func (s *Style) FontConfiguration(set *color.HColorSet, colors *color.Colors) *FontConfiguration {
	f := s.UFont()
	var c color.HColor
	if colors != nil {
		c = colors.Color(color.TypeText)
	}
	if c == nil {
		c = s.Value(PNameFontColor).AsColor(set)
	}
	hyperlinkColor := s.Value(PNameHyperLinkColor).AsColor(set)
	stroke := s.strokeOf(PNameHyperlinkUnderlineThickness, PNameHyperlinkUnderlineStyle)
	return NewFontConfiguration(f, c, hyperlinkColor, &stroke)
}

// SymbolContext returns the fashion of the style, the back and line colors
// of colors (which may be nil) taking precedence.
func (s *Style) SymbolContext(set *color.HColorSet, colors *color.Colors) *Fashion {
	var backColor, foreColor color.HColor
	if colors != nil {
		backColor = colors.Color(color.TypeBack)
	}
	if backColor == nil {
		backColor = s.Value(PNameBackGroundColor).AsColor(set)
	}
	if colors != nil {
		foreColor = colors.Color(color.TypeLine)
	}
	if foreColor == nil {
		foreColor = s.Value(PNameLineColor).AsColor(set)
	}
	deltaShadowing := s.Shadowing()
	roundCorner := s.Value(PNameRoundCorner).AsDouble()
	diagonalCorner := s.Value(PNameDiagonalCorner).AsDouble()
	return NewFashion(backColor, foreColor).WithStroke(s.Stroke()).WithDeltaShadow(deltaShadowing).
		WithCorner(roundCorner, diagonalCorner)
}

// EventuallyOverrideStroke returns the style with the thickness and dashes
// of stroke (nil gives s).
func (s *Style) EventuallyOverrideStroke(stroke *UStroke) *Style {
	if stroke == nil {
		return s
	}
	result := s.EventuallyOverrideDouble(PNameLineThickness, stroke.Thickness())
	space := stroke.DashSpace()
	visible := stroke.DashVisible()
	return result.EventuallyOverrideString(PNameLineStyle, jstr(visible)+"-"+jstr(space))
}

// Stroke returns the stroke of LineThickness and LineStyle.
func (s *Style) Stroke() UStroke { return s.strokeOf(PNameLineThickness, PNameLineStyle) }

func (s *Style) strokeOf(thicknessParam, styleParam PName) UStroke {
	thickness := s.Value(thicknessParam).AsDouble()
	dash := valueStringNonNull(s.Value(styleParam))
	if dash == "" {
		return UStrokeWithThickness(thickness)
	}
	tokens := javaTokenize(dash, "-;,")
	if len(tokens) == 0 {
		return UStrokeWithThickness(thickness)
	}
	trim := func(s string) string { return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' }) }
	dashVisible, ok := javaTryParseDouble(trim(tokens[0]))
	if !ok {
		return UStrokeWithThickness(thickness)
	}
	dashSpace := dashVisible
	if len(tokens) > 1 {
		dashSpace, ok = javaTryParseDouble(trim(tokens[1]))
		if !ok {
			return UStrokeWithThickness(thickness)
		}
	}
	return NewUStroke(dashVisible, dashSpace, thickness)
}

// StrokeOf returns the line stroke of colors, or else the stroke of the
// style.
func (s *Style) StrokeOf(colors *color.Colors) UStroke {
	if stroke := ColorsSpecificLineStroke(colors); stroke != nil {
		return *stroke
	}
	return s.Stroke()
}

// WrapWidth returns the line break strategy of MaximumWidth.
func (s *Style) WrapWidth() LineBreakStrategy {
	return NewLineBreakStrategyOpt(valueString(s.Value(PNameMaximumWidth)))
}

// Padding returns the Padding property.
func (s *Style) Padding() ClockwiseTopRightBottomLeft {
	return ClockwiseRead(valueStringNonNull(s.Value(PNamePadding)))
}

// Margin returns the Margin property.
func (s *Style) Margin() ClockwiseTopRightBottomLeft {
	return ClockwiseRead(valueStringNonNull(s.Value(PNameMargin)))
}

// HorizontalAlignment returns the HorizontalAlignment property,
// HorizontalAlignmentNull when it is not an alignment.
func (s *Style) HorizontalAlignment() geom.HorizontalAlignment {
	return s.Value(PNameHorizontalAlignment).AsHorizontalAlignment()
}

// The identifiers of the document parts drawn with bordered text blocks.
const (
	StyleIDTitle   = "_title"
	StyleIDCaption = "_caption"
	StyleIDLegend  = "_legend"
)

// CreateTextBlockBordered returns note as a text in a box drawn with the
// style.
func (s *Style) CreateTextBlockBordered(note *Display, set *color.HColorSet, spriteContainer ISkinSimple,
	id string, lineBreak LineBreakStrategy) TextBlock {
	alignment := s.HorizontalAlignment()
	fc := s.FontConfiguration(set, nil)
	textBlock := note.Create0(fc, alignment, spriteContainer, lineBreak, CreoleModeFull, nil, nil)
	backgroundColor := s.Value(PNameBackGroundColor).AsColor(set)
	lineColor := s.Value(PNameLineColor).AsColor(set)
	stroke := s.Stroke()
	cornersize := s.Value(PNameRoundCorner).AsInt()
	margin := s.Margin()
	padding := s.Padding()
	result := TextBlockUtilsBordered(textBlock, &stroke, lineColor, backgroundColor, float64(cornersize), padding,
		id)
	return TextBlockUtilsWithMargins(result, margin)
}

// ApplyStrokeAndLineColor applies the line color and stroke to ug.
func (s *Style) ApplyStrokeAndLineColor(ug UGraphic, colorSet *color.HColorSet) UGraphic {
	c := s.Value(PNameLineColor).AsColor(colorSet)
	if c == nil {
		ug = ug.Apply(color.None())
	} else {
		ug = ug.Apply(c)
	}
	return ug.Apply(s.Stroke())
}

// ToColors returns the back, line, arrow and text colors of the style.
func (s *Style) ToColors(colorSet *color.HColorSet) *color.Colors {
	return ColorsApplyStyle(color.EmptyColors(), s, colorSet)
}

// ColorsApplyStyle is upstream's Colors.applyStyle: colors with the back,
// line (and arrow) and text colors of style.
func ColorsApplyStyle(colors *color.Colors, style *Style, colorSet *color.HColorSet) *color.Colors {
	result := colors
	if tmp := style.Value(PNameBackGroundColor); tmp != nil {
		result = result.Add(color.TypeBack, tmp.AsColor(colorSet))
	}
	if tmp := style.Value(PNameLineColor); tmp != nil {
		result = result.Add(color.TypeLine, tmp.AsColor(colorSet))
		result = result.Add(color.TypeArrow, tmp.AsColor(colorSet))
	}
	if tmp := style.Value(PNameFontColor); tmp != nil {
		result = result.Add(color.TypeText, tmp.AsColor(colorSet))
	}
	return result
}

// ColorsColorOf is upstream's Colors.getColor(Style, PName, HColorSet): the
// back or line color of colors for those properties, else the style's.
func ColorsColorOf(colors *color.Colors, style *Style, name PName, colorSet *color.HColorSet) color.HColor {
	if colors == nil {
		throw(nullPointer, "")
	}
	if name == PNameBackGroundColor {
		if c := colors.Color(color.TypeBack); c != nil {
			return c
		}
	}
	if name == PNameLineColor {
		if c := colors.Color(color.TypeLine); c != nil {
			return c
		}
	}
	return style.Value(name).AsColor(colorSet)
}

// lineStyleStroke is upstream's LinkStyle.getStroke3 for a line style
// without thickness.
func lineStyleStroke(ls color.LineStyle) UStroke {
	switch ls {
	case color.LineStyleDashed:
		return NewUStroke(7, 7, 1)
	case color.LineStyleDotted:
		return NewUStroke(1, 3, 1)
	case color.LineStyleBold:
		return UStrokeWithThickness(2)
	}
	return UStrokeWithThickness(1)
}

// ColorsSpecificLineStroke is upstream's Colors.getSpecificLineStroke: the
// stroke of the line style of colors, nil without one.
func ColorsSpecificLineStroke(colors *color.Colors) *UStroke {
	if colors == nil {
		throw(nullPointer, "")
	}
	if colors.LineStyle() == color.LineStyleNone {
		return nil
	}
	s := lineStyleStroke(colors.LineStyle())
	return &s
}

// ColorsApplyStroke is upstream's Colors.applyStroke: ug with the stroke of
// the line style of colors (which may be nil).
func ColorsApplyStroke(ug UGraphic, colors *color.Colors) UGraphic {
	if colors == nil || colors.LineStyle() == color.LineStyleNone {
		return ug
	}
	return ug.Apply(lineStyleStroke(colors.LineStyle()))
}

// ColorsMuteStroke is upstream's Colors.muteStroke: the stroke of a dashed,
// dotted or bold line style, else stroke.
func ColorsMuteStroke(colors *color.Colors, stroke UStroke) UStroke {
	if colors == nil {
		throw(nullPointer, "")
	}
	switch colors.LineStyle() {
	case color.LineStyleDashed, color.LineStyleDotted, color.LineStyleBold:
		return lineStyleStroke(colors.LineStyle())
	}
	return stroke
}

// ColorsMute is upstream's Colors.mute: the skin parameters with the colors
// of colors.
func ColorsMute(colors *color.Colors, skinParam ISkinParam) ISkinParam {
	return NewSkinParamColors(skinParam, colors)
}

// ColorsApplyStereotype is upstream's Colors.applyStereotype: colors with
// the color of param for the stereotype, unless colors already has one of
// its type. An unknown color is a *color.NoSuchColorError.
func ColorsApplyStereotype(colors *color.Colors, stereotype *Stereotype, skinParam ISkinParam,
	param ColorParam) (*color.Colors, error) {
	if stereotype == nil {
		throw(nullPointer, "")
	}
	colorType, ok := param.ColorType()
	if !ok {
		throw(nullPointer, "")
	}
	if colors.Color(colorType) != nil {
		return colors, nil
	}
	other, err := skinParam.Colors(param, stereotype)
	if err != nil {
		return nil, err
	}
	return colors.MergeWith(other), nil
}
