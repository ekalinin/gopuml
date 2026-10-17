// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/AtomText.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/AtomTextUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/CharHidder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// delayedDouble is upstream's AtomText.DelayedDouble: a margin computed
// with the StringBounder.
type delayedDouble interface {
	double(stringBounder font.StringBounder) float64
}

// delayedZero is upstream's AtomTextUtils.ZERO, compared by identity.
type delayedZero struct{}

func (delayedZero) double(font.StringBounder) float64 { return 0 }

// delayedFunc is a margin computed by a function.
type delayedFunc struct {
	f func(stringBounder font.StringBounder) float64
}

func (d *delayedFunc) double(stringBounder font.StringBounder) float64 { return d.f(stringBounder) }

// AtomText is a run of text in one font configuration, possibly a link.
type AtomText struct {
	AtomBase
	fontConfiguration  *FontConfiguration
	text               string
	marginLeft         delayedDouble
	marginRight        delayedDouble
	url                *URL
	manageSpecialChars bool
}

var _ Atom = (*AtomText)(nil)

func newAtomText(text string, style *FontConfiguration, url *URL, marginLeft, marginRight delayedDouble,
	manageSpecialChars bool) *AtomText {
	s := CharHidderUnhide(text)
	if manageSpecialChars {
		s = atomTextManageSpecialChars(s)
	}
	return &AtomText{marginLeft: marginLeft, marginRight: marginRight, manageSpecialChars: manageSpecialChars,
		text: s, fontConfiguration: style, url: url}
}

// isAtomText is the marker that the SPECIALTXT graphics test.
func (a *AtomText) isAtomText() {}

// atomTextManageSpecialChars replaces "&#nnn;" (a decimal code point),
// "<U+hhhh>" (four or five hexadecimal digits), "~@start" and "\t".
func atomTextManageSpecialChars(str string) string {
	s := utf16Of(str)
	length := len(s)
	var sb []uint16 // allocated lazily on first replacement
	changed := false
	i := 0
	for i < length {
		c := s[i]
		consumed := 0
		var replacement []uint16
		found := false
		switch c {
		case '&':
			// Try to match &#[0-9]+;
			j := i + 1
			if j < length && s[j] == '#' {
				j++
				start := j
				for j < length && s[j] >= '0' && s[j] <= '9' {
					j++
				}
				if j > start && j < length && s[j] == ';' {
					// An overflow leaves the text as it is. So does a code
					// point above U+10FFFF, for which upstream's
					// Character.toChars throws IllegalArgumentException
					// (fixed per design D16).
					codePoint, err := strconv.ParseInt(stringOf(s[start:j]), 10, 32)
					if err == nil && codePoint <= unicode.MaxRune {
						replacement = javaToChars(int(codePoint))
						consumed = j + 1 - i
						found = true
					}
				}
			}
		case '<':
			// Try to match <U+[0-9a-fA-F]{4,5}>
			if i+6 < length && s[i+1] == 'U' && s[i+2] == '+' {
				j := i + 3
				start := j
				for j < length && j-start < 5 && isHexDigit(s[j]) {
					j++
				}
				if hexLen := j - start; hexLen >= 4 && j < length && s[j] == '>' {
					value, _ := strconv.ParseInt(stringOf(s[start:j]), 16, 32)
					replacement = javaToChars(int(value))
					consumed = j + 1 - i
					found = true
				}
			}
		case '~':
			// Try to match ~@start
			if i+7 <= length && stringOf(s[i+1:i+7]) == "@start" {
				replacement = utf16Of("@start")
				consumed = 7
				found = true
			}
		case '\\':
			// Try to match \t
			if i+1 < length && s[i+1] == 't' {
				replacement = []uint16{'\t'}
				consumed = 2
				found = true
			}
		}
		if found {
			if !changed {
				sb = append(make([]uint16, 0, length), s[:i]...)
				changed = true
			}
			sb = append(sb, replacement...)
			i += consumed
		} else {
			if changed {
				sb = append(sb, c)
			}
			i++
		}
	}
	if !changed {
		return str
	}
	return stringOf(sb)
}

// javaToChars is Java's Character.toChars: the UTF-16 units of a code
// point. An invalid code point panics, as Java throws
// IllegalArgumentException.
func javaToChars(codePoint int) []uint16 {
	if codePoint < 0 || codePoint > 0x10FFFF {
		throw(illegalArgument, "Not a valid Unicode code point: 0x"+strings.ToUpper(strconv.FormatInt(
			int64(uint32(codePoint)), 16)))
	}
	if codePoint >= 0x10000 {
		r1, r2 := utf16.EncodeRune(rune(codePoint))
		return []uint16{uint16(r1), uint16(r2)}
	}
	return []uint16{uint16(codePoint)}
}

func isHexDigit(c uint16) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func (a *AtomText) withText(text string) *AtomText {
	return newAtomText(text, a.fontConfiguration, a.url, a.marginLeft, a.marginRight, a.manageSpecialChars)
}

// CalculateDimension returns the size of the text with its margins, at
// least 10 high; tabulations advance to the next tab stop.
func (a *AtomText) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		rect := sb.CalculateDimension(a.fontConfiguration.Font(), a.text)
		h := rect.Height
		if h < 10 {
			h = 10
		}
		width := rect.Width
		if (strings.ContainsRune(a.text, '\t') || strings.ContainsRune(a.text, JawsBlockE1RealTabulation)) &&
			!sb.MatchesProperty("TIKZ") {
			width = a.width(sb, a.text)
		}
		left := a.marginLeft.double(sb)
		right := a.marginRight.double(sb)
		return geom.NewXDimension2D(width+left+right, h)
	})
}

// FontHeight returns the height of the font above the descent.
func (a *AtomText) FontHeight(stringBounder font.StringBounder) float64 {
	rect := stringBounder.CalculateDimension(a.fontConfiguration.Font(), a.text)
	descent := stringBounder.Descent(a.fontConfiguration.Font(), a.text)
	return rect.Height - descent
}

// DrawU draws the text on its baseline, inside its link if any.
func (a *AtomText) DrawU(ug UGraphic) {
	if a.url != nil {
		ug.StartURL(a.url)
	}
	if ug.MatchesProperty("SPECIALTXT") {
		ug.Draw(a)
	} else {
		useFontConfiguration := a.fontConfiguration.AdjustColorForBackground(ug)
		if a.marginLeft != (delayedZero{}) {
			ug = ug.Apply(UTranslateDx(a.marginLeft.double(ug.StringBounder())))
		}
		tokens := javaTokenizeWithDelims(a.text, "\t"+string(JawsBlockE1RealTabulation))
		rect := ug.StringBounder().CalculateDimension(a.fontConfiguration.Font(), a.text)
		descent := a.descent(ug.StringBounder())
		ypos := rect.Height - descent
		x := 0.0
		if len(tokens) > 0 {
			tabSize := a.tabSize(ug.StringBounder())
			for _, s := range tokens {
				if s == "\t" || s == string(JawsBlockE1RealTabulation) {
					remainder := math.Mod(x, tabSize)
					x += tabSize - remainder
				} else {
					dim := ug.StringBounder().CalculateDimension(a.fontConfiguration.Font(), s)
					utext := UTextBuild(s, useFontConfiguration)
					ug.Apply(NewUTranslate(x, ypos)).Draw(utext)
					x += dim.Width
				}
			}
		}
	}
	if a.url != nil {
		ug.CloseURL()
	}
}

func (a *AtomText) width(stringBounder font.StringBounder, text string) float64 {
	tabSize := a.tabSize(stringBounder)
	x := 0.0
	for _, s := range javaTokenizeWithDelims(text, "\t"+string(JawsBlockE1RealTabulation)) {
		if s == "\t" || s == string(JawsBlockE1RealTabulation) {
			remainder := math.Mod(x, tabSize)
			x += tabSize - remainder
		} else {
			x += stringBounder.CalculateDimension(a.fontConfiguration.Font(), s).Width
		}
	}
	return x
}

func (a *AtomText) tabString() string {
	if nb := a.fontConfiguration.TabSize(); nb >= 1 && nb < 7 {
		return "        "[:nb]
	}
	return "        "
}

func (a *AtomText) descent(stringBounder font.StringBounder) float64 {
	return stringBounder.Descent(a.fontConfiguration.Font(), a.text)
}

func (a *AtomText) tabSize(stringBounder font.StringBounder) float64 {
	width := stringBounder.CalculateDimension(a.fontConfiguration.Font(), a.tabString()).Width
	if width == 0 {
		return a.fontConfiguration.Font().Size2D() * 4
	}
	return width
}

// Neutrons splits the text into words, runs of spaces and CJK ideographs,
// with a break opportunity around each space run and ideograph.
func (a *AtomText) Neutrons() []*Neutron {
	var result []*Neutron
	var pending []uint16
	for _, ch := range utf16Of(a.text) {
		if len(pending) == 0 {
			pending = append(pending, ch)
			continue
		}
		pendingType := NeutronTypeFromChar(rune(pending[0]))
		currentType := NeutronTypeFromChar(rune(ch))
		if pendingType != currentType || pendingType == NeutronTypeCjkIdeograph {
			result = a.addPending(result, stringOf(pending))
			pending = pending[:0]
		}
		pending = append(pending, ch)
	}
	if len(pending) > 0 {
		result = a.addPending(result, stringOf(pending))
	}
	return result
}

func (a *AtomText) addPending(result []*Neutron, pending string) []*Neutron {
	tmp := NeutronCreate(a.withText(pending))
	breakable := tmp.Type() == NeutronTypeWhitespace || tmp.Type() == NeutronTypeCjkIdeograph
	if breakable {
		result = append(result, NeutronZwspSeparator())
	}
	result = append(result, tmp)
	if breakable {
		result = append(result, NeutronZwspSeparator())
	}
	return result
}

// Text returns the text.
func (a *AtomText) Text() string { return a.text }

// StartingAltitude returns the vertical shift of the font position
// (superscript, subscript).
func (a *AtomText) StartingAltitude(font.StringBounder) float64 {
	return float64(a.fontConfiguration.Space())
}

// String is upstream's toString.
func (a *AtomText) String() string { return a.text + " " + a.fontConfiguration.String() }

// FontConfiguration returns the font configuration.
func (a *AtomText) FontConfiguration() *FontConfiguration { return a.fontConfiguration }

// URL returns the link, nil for none.
func (a *AtomText) URL() *URL { return a.url }

// AtomTextUtilsCreateLegacy returns a text atom that replaces the special
// character notations ("&#nnn;", "<U+hhhh>", "~@start", "\t").
func AtomTextUtilsCreateLegacy(text string, fontConfiguration *FontConfiguration) Atom {
	return newAtomText(text, fontConfiguration, nil, delayedZero{}, delayedZero{}, true)
}

// AtomTextUtilsCreate returns a text atom.
func AtomTextUtilsCreate(text string, fontConfiguration *FontConfiguration) Atom {
	return newAtomText(text, fontConfiguration, nil, delayedZero{}, delayedZero{}, false)
}

// AtomTextUtilsCreateURL returns the atom of a link: its label in the
// hyperlink style, one atom per line of the label.
func AtomTextUtilsCreateURL(url *URL, fontConfiguration *FontConfiguration, skinSimple ISkinSimple) Atom {
	fontConfiguration = fontConfiguration.Hyperlink()
	display := DisplayWithNewlines(skinSimple.Pragma(), url.Label())
	if display.Size() > 1 {
		var all []Atom
		for _, s := range display.AsList() {
			all = append(all, createAtomTextForURL(s.String(), url, fontConfiguration, skinSimple))
		}
		return NewAtomVerticalTexts(all)
	}
	return createAtomTextForURL(url.Label(), url, fontConfiguration, skinSimple)
}

// atomTextURLPattern is upstream's AtomTextUtils.p (a java.util.regex
// pattern, case-sensitive): icons, sprites, images and emoji in a link
// label.
var atomTextURLPattern = func() *regexp.Regexp {
	t, err := regex.Transform(splitterOpeniconPattern + "|" + splitterSpritePattern + "|" +
		splitterImgPatternNoSrcColon + "|" + splitterEmojiPattern)
	if err != nil {
		panic(err)
	}
	return regexp.MustCompile(t)
}()

func createAtomTextForURL(text string, url *URL, fontConfiguration *FontConfiguration,
	skinSimple ISkinSimple) Atom {
	var result []Atom
	last := 0
	for _, m := range atomTextURLPattern.FindAllStringSubmatchIndex(text, -1) {
		if m[0] > last {
			result = append(result, newAtomText(text[last:m[0]], fontConfiguration, url, delayedZero{},
				delayedZero{}, true))
		}
		last = m[1]
		group := func(i int) (string, bool) {
			if m[2*i] < 0 {
				return "", false
			}
			return text[m[2*i]:m[2*i+1]], true
		}
		literal := func() Atom {
			return newAtomText(text[m[0]:m[1]], fontConfiguration, url, delayedZero{}, delayedZero{}, true)
		}
		_, isOpenicon := group(2)
		valSprite, isSprite := group(5)
		valImg, isImg := group(7)
		switch {
		case isOpenicon:
			// OpenIconic is not ported: the reference renders as its
			// literal text (see addOpenIcon).
			result = append(result, literal())
		case isSprite:
			// Note: sprites are probably not working in URL
			if sprite := skinSimple.Sprite(valSprite); sprite != nil {
				scaleGroup, ok := group(6)
				scale := ParserScale(scaleGroup, ok, 1)
				result = append(result, NewAtomSprite(fontConfiguration.Color(), nil, scale, sprite, url, nil))
			} else {
				result = append(result, literal())
			}
		case isImg:
			scaleGroup, ok := group(8)
			scale := ParserScale(scaleGroup, ok, 1)
			result = append(result, AtomImgCreate(skinSimple, valImg, geom.ImgValignTop, 0, scale, url))
		default:
			// An emoji. Upstream drops it, or throws
			// UnsupportedOperationException when it has a color; emoji are
			// not ported and render as their literal text (see addEmoji).
			result = append(result, literal())
		}
	}
	if last < len(text) {
		result = append(result, newAtomText(text[last:], fontConfiguration, url, delayedZero{}, delayedZero{},
			true))
	}
	if len(result) == 1 {
		return result[0]
	}
	return NewAtomHorizontalTexts(result)
}

// AtomTextUtilsCreateListNumber returns the number of a numbered list item
// ("1."), indented by its level.
func AtomTextUtilsCreateListNumber(fontConfiguration *FontConfiguration, order, localNumber int) Atom {
	left := &delayedFunc{func(stringBounder font.StringBounder) float64 {
		dim := stringBounder.CalculateDimension(fontConfiguration.Font(), "9. ")
		return dim.Width * float64(order)
	}}
	right := &delayedFunc{func(stringBounder font.StringBounder) float64 {
		return stringBounder.CalculateDimension(fontConfiguration.Font(), ".").Width
	}}
	return newAtomText(itoa(localNumber+1)+".", fontConfiguration, nil, left, right, true)
}

// CharHidderAddTileAtBegin returns s after a "~".
func CharHidderAddTileAtBegin(s string) string { return "~" + s }

// CharHidderHide replaces the escaped characters ("~" followed by one of
// _-"#][*./<, and "\~") by private-use characters, so that the creole
// markup ignores them; "~" followed by another character is kept.
func CharHidderHide(s string) string {
	runes := []rune(s)
	n := len(runes)
	var sb strings.Builder
	changed := false
	start := func(i int) {
		if !changed {
			sb.Grow(len(s))
			sb.WriteString(string(runes[:i]))
			changed = true
		}
	}
	for i := 0; i < n; {
		c := runes[i]
		switch {
		case c == '\\' && i+1 < n && runes[i+1] == '~':
			start(i)
			sb.WriteRune(0xE000 + '~')
			i += 2
		case c == '~' && i+1 < n:
			c2 := runes[i+1]
			if charHidderIsToBeHidden(c2) {
				start(i)
				sb.WriteRune(0xE000 + c2)
			} else if changed {
				sb.WriteRune(c)
				sb.WriteRune(c2)
			}
			i += 2
		default:
			if changed {
				sb.WriteRune(c)
			}
			i++
		}
	}
	if !changed {
		return s
	}
	return sb.String()
}

func charHidderIsToBeHidden(c rune) bool {
	switch c {
	case '_', '-', '"', '#', ']', '[', '*', '.', '/', '<':
		return true
	}
	return false
}

// CharHidderUnhide restores the characters hidden by CharHidderHide (any
// character from U+E000 to U+E0FF).
func CharHidderUnhide(s string) string {
	if !strings.ContainsFunc(s, func(r rune) bool { return r >= 0xE000 && r <= 0xE0FF }) {
		return s
	}
	runes := []rune(s)
	for i, r := range runes {
		if r >= 0xE000 && r <= 0xE0FF {
			runes[i] = r - 0xE000
		}
	}
	return string(runes)
}
