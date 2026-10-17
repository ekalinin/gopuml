// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/Command.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/Splitter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/AddStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleExposantChange.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleMath.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleLatex.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/CommandCreoleBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// The patterns of upstream's Splitter that the commands use. Splitter's
// tokenizer (the htmlTag pattern), purgeAllTag and the HtmlCommand classes
// (PlainText, TextLink, Img) are not used by upstream and not ported.
const (
	splitterFontColorPattern     = `\<color[\s:]+(#[0-9a-fA-F]{1,6}|#?\w+)[%s]*\>`
	splitterFontSizePattern      = `\<size[\s:]+(\d+)[%s]*\>`
	splitterQrcodePattern        = `\<qrcode[\s:]+([^>{}]+)` + `(\{scale=(?:[0-9.]+)\})?` + `\>`
	splitterImgPatternNoSrcColon = `\<img[\s:]+([^>{}]+)` + `(\{scale=(?:[0-9.]+)\})?` + `\>`
	splitterFontFamilyPattern    = `\<font[\s:]+([^>]+)/?\>`
	splitterSvgAttributePattern  = `\<text[\s:]+([^>]+)/?\>`
	splitterScaleOrColor         = `(` + `[\{,]?` + `(?:(?:scale=|\*)[0-9.]+)?` +
		`(?:,?color[= :](?:#[0-9a-fA-F]{1,8}|\w+))?` + `\}?` + `)?`
	splitterEmojiPattern    = `\<(#\w+)?:([0-9a-z][0-9_a-z]*):` + splitterScaleOrColor + `\>`
	splitterOpeniconPattern = `\<(#\w+)?&([-\w]+)` + splitterScaleOrColor + `\>`
	splitterSpritePattern   = `\<(#\w+)?\$(` + SpriteUtilsSpriteName + `)` + splitterScaleOrColor + `\>`
	splitterMathPattern     = `\<math\>(.+?)\</math\>`
	splitterLatexPattern    = `\<latex\>(.+?)\</latex\>`
	// commandCreoleColorAndSizeFontPattern is
	// CommandCreoleColorAndSizeChange.fontPattern.
	commandCreoleColorAndSizeFontPattern = `\<font(?:[%s]+size[%s]*=[%s]*[%g]?(\d+)[%g]?|[%s]+color[%s]*=[%s]*` +
		`[%g]?(#[0-9a-fA-F]{6}|\w+)[%g]?)+[%s]*\>`
)

// creoleCommand is upstream's creole Command: markup recognized at a
// position of a line (in UTF-16 units) and applied to the stripe.
type creoleCommand interface {
	// starters returns the two-character prefixes that can start the markup.
	starters() []string
	// matchingSize returns a non-zero value when the markup starts at pos.
	matchingSize(line []uint16, pos int) int
	// executeAndAdvance applies the markup and returns its length.
	executeAndAdvance(skinSimple ISkinSimple, line []uint16, pos int, stripe *StripeSimple) int
}

// creolePattern is upstream's Pattern2 compiled for the commands:
// case-insensitive, with Pattern2's macros, and Java's ".".
func creolePattern(p string) *regexp.Regexp { return regex.MustCompile(javaDots(p)) }

// creoleMatch is a match of a creolePattern at a position of a line: the
// pattern is applied to the rest of the line, as Java's Matcher.region.
type creoleMatch struct {
	s   string
	loc []int
}

func creoleFind(p *regexp.Regexp, line []uint16, pos int) *creoleMatch {
	s := stringOf(line[pos:])
	loc := p.FindStringSubmatchIndex(s)
	if loc == nil {
		return nil
	}
	return &creoleMatch{s: s, loc: loc}
}

// group returns group i; ok is false when it did not participate (null
// upstream).
func (m *creoleMatch) group(i int) (string, bool) {
	if m.loc[2*i] < 0 {
		return "", false
	}
	return m.s[m.loc[2*i]:m.loc[2*i+1]], true
}

// groupLen returns the length of group i in UTF-16 units; a missing group
// panics, as calling length() on null.
func (m *creoleMatch) groupLen(i int) int {
	g, ok := m.group(i)
	if !ok {
		throw(nullPointer, "")
	}
	return javaLen(g)
}

func (m *creoleMatch) mustGroup(i int) string {
	g, ok := m.group(i)
	if !ok {
		throw(nullPointer, "")
	}
	return g
}

// creoleRegexCommand is a command whose markup is a regular expression:
// matchingSize is the length of group sizeGroup and execute applies it.
type creoleRegexCommand struct {
	starter   []string
	pattern   *regexp.Regexp
	sizeGroup int
	execute   func(skinSimple ISkinSimple, m *creoleMatch, stripe *StripeSimple) int
}

func (c *creoleRegexCommand) starters() []string { return c.starter }

func (c *creoleRegexCommand) matchingSize(line []uint16, pos int) int {
	m := creoleFind(c.pattern, line, pos)
	if m == nil {
		return 0
	}
	return m.groupLen(c.sizeGroup)
}

func (c *creoleRegexCommand) executeAndAdvance(skinSimple ISkinSimple, line []uint16, pos int,
	stripe *StripeSimple) int {
	m := creoleFind(c.pattern, line, pos)
	if m == nil {
		throw(illegalState, "")
	}
	return c.execute(skinSimple, m, stripe)
}

// withFontConfiguration analyzes text with fc, then restores the font
// configuration of the stripe.
func withFontConfiguration(stripe *StripeSimple, fc *FontConfiguration, text string) {
	fc1 := stripe.ActualFontConfiguration()
	stripe.SetActualFontConfiguration(fc)
	stripe.AnalyzeAndAdd(text)
	stripe.SetActualFontConfiguration(fc1)
}

// addStyle is upstream's AddStyle: a font style with an optional color.
type addStyle struct {
	style         font.FontStyle
	extendedColor color.HColor
}

func (a addStyle) apply(initial *FontConfiguration) *FontConfiguration {
	initial = initial.Add(a.style)
	if a.extendedColor != nil {
		initial = initial.ChangeExtendedColor(a.extendedColor)
	}
	return initial
}

// commandCreoleStyle is upstream's CommandCreoleStyle: a font style in
// creole ("**bold**") or HTML-like ("<b>bold</b>", "<b>to the end of the
// line") markup, written as a UBrex expression.
type commandCreoleStyle struct {
	style            font.FontStyle
	tryExtendedColor bool
	isCreolePure     bool
	ubrex            *ubrexExpression
}

func commandCreoleStyleCreole(style font.FontStyle) creoleCommand {
	expression := style.UbrexCreoleSyntax() + "〶$V=〄+〴.->〘" + style.UbrexCreoleSyntax() + "〙"
	return &commandCreoleStyle{ubrex: ubrexBuild(expression), style: style, isCreolePure: true}
}

func commandCreoleStyleLegacy(style font.FontStyle) creoleCommand {
	expression := style.UbrexActivationPattern() + "〶$V=〄>〘" + style.UbrexDeactivationPattern() + "〙"
	return &commandCreoleStyle{ubrex: ubrexBuild(expression), style: style,
		tryExtendedColor: style.CanHaveExtendedColor()}
}

func commandCreoleStyleLegacyEol(style font.FontStyle) creoleCommand {
	expression := style.UbrexActivationPattern() + "〶$V=〇+〴."
	return &commandCreoleStyle{ubrex: ubrexBuild(expression), style: style,
		tryExtendedColor: style.CanHaveExtendedColor()}
}

func (c *commandCreoleStyle) starters() []string { return c.style.Starters(c.isCreolePure) }

func (c *commandCreoleStyle) extendedColor(matcher *ubrexMatcher) color.HColor {
	if c.tryExtendedColor {
		if extendedColor := matcher.findValuesByKey("XC"); len(extendedColor) > 0 {
			return color.Instance().GetColorOrWhite(extendedColor[0])
		}
	}
	return nil
}

func (c *commandCreoleStyle) executeAndAdvance(skinSimple ISkinSimple, line []uint16, pos int,
	stripe *StripeSimple) int {
	matcher := c.ubrex.match(line, pos)
	value := matcher.findValuesByKey("V")
	accepted := matcher.acceptedMatch()
	if len(value) == 0 {
		throw(illegalState, "")
	}
	fc1 := stripe.ActualFontConfiguration()
	fc2 := addStyle{c.style, c.extendedColor(matcher)}.apply(fc1)
	withFontConfiguration(stripe, fc2, value[0])
	return javaLen(accepted)
}

func (c *commandCreoleStyle) matchingSize(line []uint16, pos int) int {
	value := c.ubrex.match(line, pos).findValuesByKey("V")
	if len(value) == 0 {
		return 0
	}
	return javaLen(value[0])
}

// commandCreoleExposantChange is upstream's CommandCreoleExposantChange:
// "<sup>" and "<sub>".
type commandCreoleExposantChange struct {
	position font.FontPosition
	ubrex    *ubrexExpression
}

func newCommandCreoleExposantChange(position font.FontPosition) creoleCommand {
	htmlTag := position.HtmlTag()
	expression := "<" + htmlTag + ">〶$V=〄>〘</" + htmlTag + ">〙"
	return &commandCreoleExposantChange{position: position, ubrex: ubrexBuild(expression)}
}

func (c *commandCreoleExposantChange) starters() []string { return []string{"<s"} }

func (c *commandCreoleExposantChange) matchingSize(line []uint16, pos int) int {
	value := c.ubrex.match(line, pos).findValuesByKey("V")
	if len(value) == 0 {
		return 0
	}
	return javaLen(value[0])
}

func (c *commandCreoleExposantChange) executeAndAdvance(skinSimple ISkinSimple, line []uint16, pos int,
	stripe *StripeSimple) int {
	matcher := c.ubrex.match(line, pos)
	value := matcher.findValuesByKey("V")
	accepted := matcher.acceptedMatch()
	if len(value) == 0 {
		throw(illegalState, "")
	}
	fc2 := stripe.ActualFontConfiguration().ChangeFontPosition(c.position)
	withFontConfiguration(stripe, fc2, value[0])
	return javaLen(accepted)
}

// commandCreoleMath is upstream's CommandCreoleMath: "<math>...</math>".
func commandCreoleMath() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<m"}, pattern: creoleMathPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			stripe.AddMath("AsciiMath", m.mustGroup(2))
			return m.groupLen(1)
		}}
}

// commandCreoleLatex is upstream's CommandCreoleLatex: "<latex>...</latex>".
func commandCreoleLatex() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<l"}, pattern: creoleLatexPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			stripe.AddMath("LaTeX", m.mustGroup(2))
			return m.groupLen(1)
		}}
}

// creoleCommands is upstream's CommandCreoleBuilder: the commands by
// starter (two UTF-16 units), in the order they are tried.
type creoleCommands map[[2]uint16][]creoleCommand

func (c creoleCommands) add(cmd creoleCommand) {
	for _, starter := range cmd.starters() {
		u := utf16Of(starter)
		key := [2]uint16{u[0], u[1]}
		c[key] = append(c[key], cmd)
	}
}

func newCreoleCommands(full bool) creoleCommands {
	c := creoleCommands{}
	c.add(commandCreoleStyleCreole(font.FontStyleBold))
	c.add(commandCreoleStyleLegacy(font.FontStyleBold))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleBold))

	c.add(commandCreoleStyleCreole(font.FontStyleItalic))
	c.add(commandCreoleStyleLegacy(font.FontStyleItalic))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleItalic))
	c.add(commandCreoleStyleLegacy(font.FontStylePlain))
	c.add(commandCreoleStyleLegacyEol(font.FontStylePlain))
	if full {
		c.add(commandCreoleStyleCreole(font.FontStyleUnderline))
	}
	c.add(commandCreoleStyleLegacy(font.FontStyleUnderline))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleUnderline))
	c.add(commandCreoleStyleCreole(font.FontStyleStrike))
	c.add(commandCreoleStyleLegacy(font.FontStyleStrike))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleStrike))
	c.add(commandCreoleStyleCreole(font.FontStyleWave))
	c.add(commandCreoleStyleLegacy(font.FontStyleWave))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleWave))
	c.add(commandCreoleStyleLegacy(font.FontStyleBackcolor))
	c.add(commandCreoleStyleLegacyEol(font.FontStyleBackcolor))
	c.add(commandCreoleSizeChange(creoleSizePattern))
	c.add(commandCreoleSizeChange(creoleSizePatternEol))
	c.add(commandCreoleColorChange(creoleColorPattern))
	c.add(commandCreoleColorChange(creoleColorPatternEol))
	c.add(commandCreoleColorAndSizeChange(creoleColorAndSizePattern))
	c.add(commandCreoleColorAndSizeChange(creoleColorAndSizePatternEol))
	c.add(newCommandCreoleExposantChange(font.FontPositionExposant))
	c.add(newCommandCreoleExposantChange(font.FontPositionIndice))
	c.add(commandCreoleImg())
	c.add(commandCreoleQrcode())
	c.add(commandCreoleOpenIcon())
	c.add(commandCreoleEmoji())
	c.add(commandCreoleMath())
	c.add(commandCreoleLatex())
	c.add(commandCreoleSprite())
	c.add(commandCreoleSpace())
	c.add(commandCreoleFontFamilyChange(creoleFontFamilyPattern))
	c.add(commandCreoleFontFamilyChange(creoleFontFamilyPatternEol))
	c.add(commandCreoleMonospaced())
	c.add(commandCreoleURL())
	// SecurityUtils.allowSvgText() is always true.
	c.add(commandCreoleSvgAttributeChange(creoleSvgAttributePattern))
	return c
}

// The command tables of upstream's CommandCreoleBuilder.FULL and OTHER,
// read-only once built.
var (
	creoleCommandsFull  = newCreoleCommands(true)
	creoleCommandsOther = newCreoleCommands(false)
)
