// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/Display.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// CharSequence is a line of a Display: a Text, a *Stereotype, or another
// element of the diagrams (a sequence message number, an embedded diagram).
type CharSequence interface {
	String() string
}

// Text is a plain line of a Display.
type Text string

// String returns the line.
func (t Text) String() string { return string(t) }

// embeddedEnd is upstream's EmbeddedDiagram.EMBEDDED_END.
const embeddedEnd = "}}"

// Display is the text of an element: lines, possibly with a stereotype, and
// the natural alignment given by "\l" or "\r". It is immutable: the methods
// return new displays.
type Display struct {
	displayData                []CharSequence
	naturalHorizontalAlignment geom.HorizontalAlignment // HorizontalAlignmentNull for none
	isNull                     bool
	defaultCreoleMode          CreoleMode
	showStereotype             bool
}

// displayNull is upstream's Display.NULL, a read-only value compared by
// identity.
var displayNull = &Display{showStereotype: true, isNull: true, defaultCreoleMode: CreoleModeFull,
	naturalHorizontalAlignment: HorizontalAlignmentNull}

// DisplayNULL returns upstream's Display.NULL, the missing display.
func DisplayNULL() *Display { return displayNull }

func newDisplayEmpty(showStereotype bool, naturalHorizontalAlignment geom.HorizontalAlignment, isNull bool,
	defaultCreoleMode CreoleMode) *Display {
	d := &Display{showStereotype: showStereotype, defaultCreoleMode: defaultCreoleMode, isNull: isNull,
		naturalHorizontalAlignment: HorizontalAlignmentNull}
	if !isNull {
		d.displayData = []CharSequence{}
		d.naturalHorizontalAlignment = naturalHorizontalAlignment
	}
	return d
}

// newDisplayCopy copies other; the null display panics, as upstream adds to
// its null list.
func newDisplayCopy(showStereotype bool, other *Display, mode CreoleMode) *Display {
	d := newDisplayEmpty(showStereotype, other.naturalHorizontalAlignment, other.isNull, mode)
	d.displayData = append(d.displayData, other.data()...)
	return d
}

// data returns the lines; the null display panics, as upstream uses its null
// list.
func (d *Display) data() []CharSequence {
	if d.isNull {
		throw(nullPointer, "")
	}
	return d.displayData
}

// at returns the line at i, panicking as Java's List.get when i is out of
// range.
func (d *Display) at(i int) CharSequence {
	data := d.data()
	if i < 0 || i >= len(data) {
		throw(indexOutOfBounds, "Index "+itoa(i)+" out of bounds for length "+itoa(len(data)))
	}
	return data[i]
}

func newDisplayFrom(showStereotype bool, other []CharSequence, naturalHorizontalAlignment geom.HorizontalAlignment,
	isNull bool, defaultCreoleMode CreoleMode) *Display {
	d := newDisplayEmpty(showStereotype, naturalHorizontalAlignment, isNull, defaultCreoleMode)
	if !isNull {
		d.displayData = append(d.displayData, other...)
	}
	return d
}

// HashCode is upstream's hashCode: 42 for the null display, else the list
// hash of the lines.
func (d *Display) HashCode() int32 {
	if d.isNull {
		return 42
	}
	h := int32(1)
	for _, cs := range d.displayData {
		h = 31*h + javaStringHashCode(cs.String())
	}
	return h
}

// Equals is upstream's equals: the lines are equal (plain lines by value,
// other elements by identity). The null display panics as receiver, and is
// equal to no display as argument.
func (d *Display) Equals(other *Display) bool {
	data := d.data()
	if other.isNull {
		return false
	}
	return slices.Equal(data, other.displayData)
}

// EqualsLike is Equals that also handles null displays.
func (d *Display) EqualsLike(other *Display) bool {
	if DisplayIsNull(d) {
		return DisplayIsNull(other)
	}
	if other.isNull {
		return false
	}
	return slices.Equal(d.displayData, other.displayData)
}

// ShowStereotype reports whether the stereotype is shown.
func (d *Display) ShowStereotype() bool { return d.showStereotype }

// WithoutStereotypeIfNeeded hides the stereotype when the style sets
// ShowStereotype to false.
func (d *Display) WithoutStereotypeIfNeeded(usedStyle *Style) *Display {
	if d == displayNull {
		return displayNull
	}
	showStereotype := usedStyle.Value(PNameShowStereotype)
	if _, isNull := showStereotype.(*ValueNull); isNull || showStereotype.AsBoolean() {
		return d
	}
	return newDisplayCopy(false, d, d.defaultCreoleMode)
}

// StereotypeIfAny returns the first stereotype line, nil for none.
func (d *Display) StereotypeIfAny() *Stereotype {
	for _, cs := range d.data() {
		if st, ok := cs.(*Stereotype); ok {
			return st
		}
	}
	return nil
}

// ReplaceBackslashT replaces "\t" by tabulations.
func (d *Display) ReplaceBackslashT() *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	for i, s := range d.data() {
		if strings.Contains(s.String(), "\\t") {
			result.displayData[i] = Text(strings.ReplaceAll(s.String(), "\\t", "\t"))
		}
	}
	return result
}

// Replace replaces src by dest in the lines (turning them into plain
// lines).
func (d *Display) Replace(src, dest string) *Display {
	var newDisplay []CharSequence
	for _, cs := range d.data() {
		if strings.Contains(cs.String(), src) {
			cs = Text(strings.ReplaceAll(cs.String(), src, dest))
		}
		newDisplay = append(newDisplay, cs)
	}
	return newDisplayFrom(d.showStereotype, newDisplay, d.naturalHorizontalAlignment, d.isNull,
		d.defaultCreoleMode)
}

var onlyWhitespace = regexp.MustCompile(`^[\t\n\v\f\r ]*$`)

// IsWhite reports whether the display is empty or a single blank line.
func (d *Display) IsWhite() bool {
	return d.displayData == nil || len(d.displayData) == 0 ||
		(len(d.displayData) == 1 && onlyWhitespace.MatchString(d.displayData[0].String()))
}

// DisplayEmpty returns a display without lines.
func DisplayEmpty() *Display {
	return newDisplayEmpty(true, HorizontalAlignmentNull, false, CreoleModeFull)
}

// DisplayCreate returns the display of plain lines.
func DisplayCreate(lines ...string) *Display {
	cs := make([]CharSequence, len(lines))
	for i, l := range lines {
		cs[i] = Text(l)
	}
	return DisplayCreateSeq(cs)
}

// DisplayCreateSeq returns the display of lines.
func DisplayCreateSeq(other []CharSequence) *Display {
	return newDisplayFrom(true, other, HorizontalAlignmentNull, false, CreoleModeFull)
}

// DisplayCreateFoo returns the display of source lines, without the empty
// line after a closing "}}" of an embedded diagram.
func DisplayCreateFoo(data []text.StringLocated) *Display {
	var tmp []string
	for _, s := range data {
		tmp = append(tmp, s.S)
	}
	if len(tmp) > 2 && tmp[len(tmp)-1] == "" && tmp[len(tmp)-2] == embeddedEnd {
		tmp = tmp[:len(tmp)-1]
	}
	return DisplayCreate(tmp...)
}

// DisplayWithNewlines3 splits s at "\n", turning "\t" into tabulations and
// "\\" into backslashes; other escapes are dropped. ok false gives nil.
func DisplayWithNewlines3(s string, ok bool) []string {
	if !ok {
		return nil
	}
	u := utf16.Encode([]rune(s))
	var result []string
	var current []uint16
	for i := 0; i < len(u); i++ {
		c := u[i]
		if c == '\\' && i < len(u)-1 {
			c2 := u[i+1]
			i++
			switch c2 {
			case 'n':
				result = append(result, string(utf16.Decode(current)))
				current = current[:0]
			case 't':
				current = append(current, '\t')
			case '\\':
				current = append(current, c2)
			}
		} else {
			current = append(current, c)
		}
	}
	return append(result, string(utf16.Decode(current)))
}

var displayMoreInfo = NewWarning("More info on https://plantuml.com/newline")

// The markers of the raw parts of a text, as UTF-16 units.
var (
	utf16Math     = utf16.Encode([]rune("<math>"))
	utf16Latex    = utf16.Encode([]rune("<latex>"))
	utf16OpenURL  = utf16.Encode([]rune("[["))
	utf16EndMath  = utf16.Encode([]rune("</math>"))
	utf16EndLatex = utf16.Encode([]rune("</latex>"))
	utf16CloseURL = utf16.Encode([]rune("]]"))
)

// DisplayWithNewlinesString is upstream's getWithNewlines(Pragma, String):
// the display of a text whose lines are separated by "\n" (outside
// "<math>", "<latex>" and "[[...]]") and by the newline characters of the
// preprocessor; "\l" and "\r" also give the natural alignment. ok false
// gives the null display.
func DisplayWithNewlinesString(pragma *Pragma, s string, ok bool) *Display {
	if !ok {
		return displayNull
	}
	u := utf16.Encode([]rune(s))
	var result []string
	var current []uint16
	naturalHorizontalAlignment := HorizontalAlignmentNull
	rawMode := false
	flush := func() {
		result = append(result, string(utf16.Decode(current)))
		current = current[:0]
	}
	hasPrefixAt := func(i int, pu []uint16) bool {
		if i+len(pu) > len(u) {
			return false
		}
		return slices.Equal(u[i:i+len(pu)], pu)
	}
	for i := 0; i < len(u); i++ {
		c := rune(u[i])
		if hasPrefixAt(i, utf16Math) || hasPrefixAt(i, utf16Latex) || hasPrefixAt(i, utf16OpenURL) {
			rawMode = true
		} else if hasPrefixAt(i, utf16EndMath) || hasPrefixAt(i, utf16EndLatex) || hasPrefixAt(i, utf16CloseURL) {
			rawMode = false
		}
		switch {
		case jawsSpecialNewlineInDisplay && hasPrefixAt(i, utf16.Encode([]rune("%newline()"))):
			flush()
			i += 9
		case jawsSpecialNewlineInDisplay && hasPrefixAt(i, utf16.Encode([]rune("%n()"))):
			flush()
			i += 3
		case PragmaLegacyReplaceBackslashNByNewline() && !rawMode && c == '\\' && i < len(u)-1:
			c2 := rune(u[i+1])
			i++
			switch c2 {
			case 'n', 'r', 'l':
				if c2 == 'r' {
					naturalHorizontalAlignment = geom.HorizontalAlignmentRight
					addDisplayWarning(pragma, JawsWarningBackslashRight)
				} else if c2 == 'l' {
					naturalHorizontalAlignment = geom.HorizontalAlignmentLeft
					addDisplayWarning(pragma, JawsWarningBackslashLeft)
				} else {
					addDisplayWarning(pragma, JawsWarningBackslashNewline)
				}
				flush()
			case 't':
				current = append(current, '\t')
				addDisplayWarning(pragma, JawsWarningBackslashTabulation)
			case '\\':
				current = append(current, uint16(c2))
				addDisplayWarning(pragma, JawsWarningBackslashBackslash)
			default:
				current = append(current, uint16(c), uint16(c2))
			}
		case c == JawsBlockE1RealTabulation:
			current = append(current, uint16(c))
		case c == JawsBlockE1RealBackslash:
			current = append(current, '\\')
		case c == JawsBlockE1NewlineLeftAlign:
			naturalHorizontalAlignment = geom.HorizontalAlignmentLeft
			flush()
		case c == JawsBlockE1InvisibleQuote:
			// No thing to do, just ignore this character
		case c == JawsBlockE1NewlineRightAlign:
			naturalHorizontalAlignment = geom.HorizontalAlignmentRight
			flush()
		case !rawMode && c == JawsBlockE1Newline:
			flush()
		case c == JawsBlockE1Breakline:
			// Because of embedded diagrams
			flush()
		case !rawMode && c == hiddenNewLine():
			flush()
		default:
			current = append(current, uint16(c))
		}
	}
	flush()
	cs := make([]CharSequence, len(result))
	for i, l := range result {
		cs[i] = Text(l)
	}
	return newDisplayFrom(true, cs, naturalHorizontalAlignment, false, CreoleModeFull)
}

// DisplayWithNewlines is upstream's getWithNewlines(Pragma, String) for a
// present text.
func DisplayWithNewlines(pragma *Pragma, s string) *Display {
	return DisplayWithNewlinesString(pragma, s, true)
}

func addDisplayWarning(pragma *Pragma, warning JawsWarning) {
	if pragma.IsTrue(PragmaShowDeprecation) {
		pragma.AddWarning(displayMoreInfo)
		pragma.AddWarning(warning.ToWarning())
	}
}

// ManageGuillemet writes the stereotypes of the lines with guillemets,
// removing the visibility character of the first line when asked.
func (d *Display) ManageGuillemet(manageVisibilityModifier bool) *Display {
	var result []CharSequence
	first := true
	for _, line := range d.data() {
		lineString := line.String()
		if manageVisibilityModifier && first && VisibilityIsVisibilityCharacter(lineString) {
			u := utf16.Encode([]rune(lineString))
			lineString = text.JavaTrim(string(utf16.Decode(u[1:])))
		}
		result = append(result, Text(GuillemetGuillemet().ManageGuillemet(lineString)))
		first = false
	}
	return newDisplayFrom(d.showStereotype, result, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
}

// WithPage replaces "%page%" and "%lastpage%".
func (d *Display) WithPage(page, lastpage int) *Display {
	if d.displayData == nil {
		return d
	}
	var result []CharSequence
	for _, line := range d.displayData {
		s := strings.ReplaceAll(line.String(), "%page%", itoa(page))
		s = strings.ReplaceAll(s, "%lastpage%", itoa(lastpage))
		result = append(result, Text(s))
	}
	return newDisplayFrom(d.showStereotype, result, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
}

// DisplayPatternStereotype is the pattern of a line ending with a
// stereotype.
var DisplayPatternStereotype = regex.MustCompile(javaDots(`^(.*?)(\<\<\s*(.*)\s*\>\>)\s*$`))

// RemoveEndingStereotype removes the stereotype that ends the last line.
func (d *Display) RemoveEndingStereotype() *Display {
	m := DisplayPatternStereotype.FindStringSubmatch(d.at(len(d.data()) - 1).String())
	if m != nil {
		result := slices.Clone(d.displayData)
		result[len(result)-1] = Text(m[1])
		return newDisplayFrom(d.showStereotype, result, d.naturalHorizontalAlignment, d.isNull,
			d.defaultCreoleMode)
	}
	return d
}

// EndingStereotype returns the stereotype that ends the last line, nil for
// none.
func (d *Display) EndingStereotype() *Stereotype {
	m := DisplayPatternStereotype.FindStringSubmatch(d.at(len(d.data()) - 1).String())
	if m != nil {
		return StereotypeBuild(m[2])
	}
	return nil
}

// Underlined returns the lines with "<u>" before each.
func (d *Display) Underlined() *Display {
	var result []CharSequence
	for _, line := range d.data() {
		result = append(result, Text("<u>"+line.String()))
	}
	return newDisplayFrom(d.showStereotype, result, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
}

var underlinedNamePattern = regexp.MustCompile(javaDots(`^([^:]+?)([\t\n\v\f\r ]*:.+)$`))

// UnderlinedName underlines the name before ":" of the first line, and the
// other lines.
func (d *Display) UnderlinedName() *Display {
	var result []CharSequence
	for _, line := range d.data() {
		if len(result) == 0 {
			if m := underlinedNamePattern.FindStringSubmatch(line.String()); m != nil {
				result = append(result, Text("<u>"+m[1]+"</u>"+m[2]))
			} else {
				result = append(result, Text("<u>"+line.String()))
			}
		} else {
			result = append(result, Text("<u>"+line.String()))
		}
	}
	return newDisplayFrom(d.showStereotype, result, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
}

// WithCreoleMode returns the display with another creole mode. The null
// display panics, as upstream throws IllegalArgumentException.
func (d *Display) WithCreoleMode(mode CreoleMode) *Display {
	if d.isNull {
		throw(illegalArgument, "")
	}
	return newDisplayCopy(d.showStereotype, d, mode)
}

// String is upstream's toString: "NULL", or the list of lines.
func (d *Display) String() string {
	if d.isNull {
		return "NULL"
	}
	parts := make([]string, len(d.displayData))
	for i, cs := range d.displayData {
		parts[i] = cs.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// AddAll returns the display followed by the lines of other.
func (d *Display) AddAll(other *Display) *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	result.displayData = append(result.displayData, other.data()...)
	return result
}

// AddFirst returns the display with a line first.
func (d *Display) AddFirst(s CharSequence) *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	result.displayData = slices.Insert(result.displayData, 0, s)
	return result
}

// AppendFirstLine returns the display with text before its first line.
func (d *Display) AppendFirstLine(appended string) *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	result.displayData[0] = Text(appended + result.at(0).String())
	return result
}

// Add returns the display with one more line.
func (d *Display) Add(s CharSequence) *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	result.displayData = append(result.displayData, s)
	return result
}

// AddGeneric returns the display with "<s>" appended to its last line.
func (d *Display) AddGeneric(s string) *Display {
	result := newDisplayCopy(d.showStereotype, d, d.defaultCreoleMode)
	size := len(d.displayData)
	if size == 0 {
		result.displayData = append(result.displayData, Text("<"+s+">"))
	} else {
		result.displayData[size-1] = Text(d.displayData[size-1].String() + "<" + s + ">")
	}
	return result
}

// Size returns the number of lines, 0 for the null display.
func (d *Display) Size() int {
	if d.isNull {
		return 0
	}
	return len(d.displayData)
}

// Get returns the line at i; it panics as Java's List.get when i is out of
// range.
func (d *Display) Get(i int) CharSequence { return d.at(i) }

// SubList returns the lines from i to size (excluded); it panics as Java's
// List.subList for a wrong range.
func (d *Display) SubList(i, size int) *Display {
	data := d.data()
	switch {
	case i < 0:
		throw(indexOutOfBounds, "fromIndex = "+itoa(i))
	case size > len(data):
		throw(indexOutOfBounds, "toIndex = "+itoa(size))
	case i > size:
		throw(illegalArgument, "fromIndex("+itoa(i)+") > toIndex("+itoa(size)+")")
	}
	return newDisplayFrom(d.showStereotype, data[i:size], d.naturalHorizontalAlignment, d.isNull,
		d.defaultCreoleMode)
}

// AsList returns the lines; the slice must not be modified.
func (d *Display) AsList() []CharSequence { return d.displayData }

// HasURL reports whether a line has a link.
func (d *Display) HasURL() bool {
	urlBuilder := NewURLBuilder("", false, URLModeAnywhere)
	for _, s := range d.data() {
		if urlBuilder.URL(s.String()) != nil {
			return true
		}
	}
	return false
}

// NaturalHorizontalAlignment returns the alignment given by "\l" or "\r",
// HorizontalAlignmentNull for none.
func (d *Display) NaturalHorizontalAlignment() geom.HorizontalAlignment {
	return d.naturalHorizontalAlignment
}

// SplitMultiline splits the display at the first match of separator in
// each line.
func (d *Display) SplitMultiline(separator *regexp.Regexp) []*Display {
	pending := newDisplayEmpty(d.showStereotype, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
	result := []*Display{pending}
	for _, line := range d.data() {
		s := line.String()
		if loc := separator.FindStringIndex(s); loc != nil {
			pending.displayData = append(pending.displayData, Text(s[:loc[0]]))
			pending = newDisplayEmpty(d.showStereotype, d.naturalHorizontalAlignment, d.isNull, d.defaultCreoleMode)
			result = append(result, pending)
			pending.displayData = append(pending.displayData, Text(s[loc[1]:]))
		} else {
			pending.displayData = append(pending.displayData, line)
		}
	}
	return result
}

// ToTooltipText returns the first line, "" without lines.
func (d *Display) ToTooltipText() string {
	if d.Size() == 0 {
		return ""
	}
	return d.Get(0).String()
}

// DisplayIsNull reports whether d is nil or the null display.
func DisplayIsNull(d *Display) bool { return d == nil || d.isNull }

// Create returns the text block of the display, with full creole.
func (d *Display) Create(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple) TextBlock {
	return d.Create7(fontConfiguration, horizontalAlignment, spriteContainer, CreoleModeFull)
}

// Create7 returns the text block of the display in a creole mode.
func (d *Display) Create7(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, creoleMode CreoleMode) TextBlock {
	return d.Create0(fontConfiguration, horizontalAlignment, spriteContainer, LineBreakStrategyNone, creoleMode,
		nil, nil)
}

// Create8 returns the text block of the display with a maximum width.
func (d *Display) Create8(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, modeSimpleLine CreoleMode, maxMessageSize LineBreakStrategy) TextBlock {
	return d.Create0(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize, modeSimpleLine, nil,
		nil)
}

// Create9 returns the text block of the display with a maximum width, in
// its default creole mode.
func (d *Display) Create9(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, maxMessageSize LineBreakStrategy) TextBlock {
	return d.Create0(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize, d.defaultCreoleMode,
		nil, nil)
}

// Create0 returns the text block of the display; fontForStereotype and
// htmlColorForStereotype may be nil.
func (d *Display) Create0(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, maxMessageSize LineBreakStrategy, creoleMode CreoleMode,
	fontForStereotype *font.UFont, htmlColorForStereotype color.HColor) TextBlock {
	return d.Create0Margins(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize, creoleMode,
		fontForStereotype, htmlColorForStereotype, 0, 0)
}

// Create0Margins is Create0 with horizontal margins.
func (d *Display) Create0Margins(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, maxMessageSize LineBreakStrategy, creoleMode CreoleMode,
	fontForStereotype *font.UFont, htmlColorForStereotype color.HColor, marginX1, marginX2 float64) TextBlock {
	if d.NaturalHorizontalAlignment() != HorizontalAlignmentNull {
		horizontalAlignment = d.NaturalHorizontalAlignment()
	}
	stereotypeConfiguration := fontConfiguration.ForceFont(fontForStereotype, htmlColorForStereotype)
	if d.Size() > 0 {
		if _, ok := d.Get(0).(*Stereotype); ok {
			return d.createStereotype(fontConfiguration, horizontalAlignment, spriteContainer, 0, fontForStereotype,
				htmlColorForStereotype, maxMessageSize, creoleMode, marginX1, marginX2)
		}
		if _, ok := d.Get(d.Size() - 1).(*Stereotype); ok {
			return d.createStereotype(fontConfiguration, horizontalAlignment, spriteContainer, d.Size()-1,
				fontForStereotype, htmlColorForStereotype, maxMessageSize, creoleMode, marginX1, marginX2)
		}
		if _, ok := d.Get(0).(MessageNumber); ok {
			return d.createMessageNumber(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize,
				stereotypeConfiguration, marginX1, marginX2)
		}
	}
	return d.creole(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize, creoleMode,
		stereotypeConfiguration, marginX1, marginX2)
}

// MessageNumber is a sequence message number line of a Display, upstream's
// sequencediagram.MessageNumber, which implements it with a
// MessageNumberMarker method.
type MessageNumber interface {
	CharSequence
	MessageNumberMarker()
}

func (d *Display) createStereotype(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, position int, fontForStereotype *font.UFont, htmlColorForStereotype color.HColor,
	maxMessageSize LineBreakStrategy, creoleMode CreoleMode, marginX1, marginX2 float64) TextBlock {
	stereotype := d.Get(position).(*Stereotype)
	var circledCharacter TextBlock
	if stereotype.IsSpotted() {
		var f font.UFont
		if stereotype.CircledFont() != nil {
			f = *stereotype.CircledFont()
		}
		circledCharacter = NewCircledCharacter(stereotype.Character(), stereotype.Radius(), f,
			stereotype.HtmlColor(), nil, fontConfiguration.Color())
	} else if sprite := stereotype.Sprite(spriteContainer); sprite != nil {
		circledCharacter = sprite
	}
	stereotypeConfiguration := fontConfiguration.ForceFont(fontForStereotype, htmlColorForStereotype)
	result := d.creole(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize, creoleMode,
		stereotypeConfiguration, marginX1, marginX2)
	if circledCharacter != nil {
		return NewTextBlockSprited(circledCharacter, result)
	}
	return result
}

func (d *Display) creole(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, maxMessageSize LineBreakStrategy, creoleMode CreoleMode,
	stereotypeConfiguration *FontConfiguration, marginX1, marginX2 float64) TextBlock {
	sheet := spriteContainer.SheetStereo(fontConfiguration, horizontalAlignment, creoleMode,
		stereotypeConfiguration).CreateSheet(d)
	padding := ClockwiseNone()
	if spriteContainer != nil {
		padding = spriteContainer.Padding()
	}
	sheetBlock1 := NewSheetBlock1(sheet, maxMessageSize, padding, marginX1, marginX2)
	return NewSheetBlock2(sheetBlock1, sheetBlock1, UStrokeWithThickness(1.5))
}

func (d *Display) createMessageNumber(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	spriteContainer ISkinSimple, maxMessageSize LineBreakStrategy, stereotypeConfiguration *FontConfiguration,
	marginX1, marginX2 float64) TextBlock {
	tb1 := d.SubList(0, 1).creole(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize,
		CreoleModeFull, stereotypeConfiguration, marginX1, marginX2)
	tb1 = TextBlockUtilsWithMargin4(tb1, 0, 4, 0, 0)
	tb2 := d.SubList(1, d.Size()).creole(fontConfiguration, horizontalAlignment, spriteContainer, maxMessageSize,
		CreoleModeFull, stereotypeConfiguration, marginX1, marginX2)
	return TextBlockUtilsMergeLR(tb1, tb2, geom.VerticalAlignmentCenter)
}

// HasSeveralGuideLines reports whether the display has several lines, one
// of which starts with "< " or "> " or ends with " <" or " >".
func (d *Display) HasSeveralGuideLines() bool {
	lines := make([]string, len(d.data()))
	for i, cs := range d.displayData {
		lines[i] = cs.String()
	}
	return hasSeveralGuideLines(lines)
}

// DisplayHasSeveralGuideLines is HasSeveralGuideLines for a text whose
// lines are separated by "\n".
func DisplayHasSeveralGuideLines(s string) bool {
	return hasSeveralGuideLines(javaSplit(s, "\\n"))
}

func hasSeveralGuideLines(all []string) bool {
	if len(all) <= 1 {
		return false
	}
	for _, s := range all {
		if strings.HasPrefix(s, "< ") || strings.HasPrefix(s, "> ") || strings.HasSuffix(s, " <") ||
			strings.HasSuffix(s, " >") {
			return true
		}
	}
	return false
}

// ContentWidth returns the length of the longest line, in UTF-16 units.
func (d *Display) ContentWidth() int {
	width := 0
	for i := 0; i < d.Size(); i++ {
		if l := javaLen(d.Get(i).String()); l > width {
			width = l
		}
	}
	return width
}
