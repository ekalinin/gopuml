// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleSizeChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleColorChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleColorAndSizeChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleFontFamilyChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleMonospaced.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleSvgAttributeChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleUrl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleSpace.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleImg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleQrcode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleOpenIcon.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleEmoji.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleSprite.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The patterns of the regular expression commands, compiled once.
var (
	creoleSizePattern            = creolePattern("^(" + splitterFontSizePattern + "(.*?)\\</size\\>)")
	creoleSizePatternEol         = creolePattern("^(" + splitterFontSizePattern + "(.*)$)")
	creoleColorPattern           = creolePattern("^(" + splitterFontColorPattern + "(.*?)\\</color\\>)")
	creoleColorPatternEol        = creolePattern("^(" + splitterFontColorPattern + "(.*)$)")
	creoleColorAndSizePattern    = creolePattern("^(" + commandCreoleColorAndSizeFontPattern + "(.*?)\\</font\\>)")
	creoleColorAndSizePatternEol = creolePattern("^(" + commandCreoleColorAndSizeFontPattern + "(.*))$")
	creoleFontFamilyPattern      = creolePattern("^(" + splitterFontFamilyPattern + "(.*?)\\</font\\>)")
	creoleFontFamilyPatternEol   = creolePattern("^(" + splitterFontFamilyPattern + "(.*)$)")
	creoleSvgAttributePattern    = creolePattern("^(" + splitterSvgAttributePattern + "(.*?)\\</text\\>)")
	creoleMonospacedPattern      = creolePattern("^(\"\"(.*?)\"\")")
	creoleURLPattern             = creolePattern("^(" + URLBuilderRegexp() + ")")
	creoleSpacePattern           = creolePattern("^(\\<space:(\\d+)/?\\>)")
	creoleImgPattern             = creolePattern("^(" + splitterImgPatternNoSrcColon + ")")
	creoleQrcodePattern          = creolePattern("^(" + splitterQrcodePattern + ")")
	creoleOpenIconPattern        = creolePattern("^(" + splitterOpeniconPattern + ")")
	creoleEmojiPattern           = creolePattern("^(" + splitterEmojiPattern + ")")
	creoleSpritePattern          = creolePattern("^(" + splitterSpritePattern + ")")
	creoleMathPattern            = creolePattern("^(" + splitterMathPattern + ")")
	creoleLatexPattern           = creolePattern("^(" + splitterLatexPattern + ")")
)

// commandCreoleSizeChange is upstream's CommandCreoleSizeChange:
// "<size:n>".
func commandCreoleSizeChange(p *regexp.Regexp) creoleCommand {
	return &creoleRegexCommand{starter: []string{"<s"}, pattern: p, sizeGroup: 2,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			size := javaParseInt(m.mustGroup(2))
			fc2 := stripe.ActualFontConfiguration().ChangeSize(float32(size))
			withFontConfiguration(stripe, fc2, m.mustGroup(3))
			return m.groupLen(1)
		}}
}

// commandCreoleColorChange is upstream's CommandCreoleColorChange:
// "<color:c>"; an unknown color is ignored.
func commandCreoleColorChange(p *regexp.Regexp) creoleCommand {
	return &creoleRegexCommand{starter: []string{"<c"}, pattern: p, sizeGroup: 2,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc1 := stripe.ActualFontConfiguration()
			if c, err := color.Instance().GetColor(m.mustGroup(2)); err == nil {
				stripe.SetActualFontConfiguration(fc1.ChangeColor(c))
			}
			// Too late for parsing error, so an unknown color is ignored.
			stripe.AnalyzeAndAdd(m.mustGroup(3))
			stripe.SetActualFontConfiguration(fc1)
			return m.groupLen(1)
		}}
}

// commandCreoleColorAndSizeChange is upstream's
// CommandCreoleColorAndSizeChange: "<font color=c size=n>". An unknown
// color panics with color.NoSuchColorRuntimeError, as upstream throws
// NoSuchColorRuntimeException.
func commandCreoleColorAndSizeChange(p *regexp.Regexp) creoleCommand {
	return &creoleRegexCommand{starter: []string{"<f"}, pattern: p, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc1 := stripe.ActualFontConfiguration()
			fc2 := fc1
			if size, ok := m.group(2); ok {
				fc2 = fc2.ChangeSize(float32(javaParseInt(size)))
			}
			if s, ok := m.group(3); ok {
				c, err := color.Instance().GetColor(s)
				if err != nil {
					panic(&color.NoSuchColorRuntimeError{})
				}
				fc2 = fc2.ChangeColor(c)
			}
			withFontConfiguration(stripe, fc2, m.mustGroup(4))
			return m.groupLen(1)
		}}
}

// commandCreoleFontFamilyChange is upstream's
// CommandCreoleFontFamilyChange: "<font:family>".
func commandCreoleFontFamilyChange(p *regexp.Regexp) creoleCommand {
	return &creoleRegexCommand{starter: []string{"<f"}, pattern: p, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc2 := stripe.ActualFontConfiguration().ChangeFamily(m.mustGroup(2))
			withFontConfiguration(stripe, fc2, m.mustGroup(3))
			return m.groupLen(1)
		}}
}

// commandCreoleMonospaced is upstream's CommandCreoleMonospaced:
// `""text""`.
func commandCreoleMonospaced() creoleCommand {
	return &creoleRegexCommand{starter: []string{`""`}, pattern: creoleMonospacedPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc2 := stripe.ActualFontConfiguration().ChangeFamily(stripe.SkinParam().MonospacedFamily())
			withFontConfiguration(stripe, fc2, m.mustGroup(2))
			return m.groupLen(1)
		}}
}

// commandCreoleSvgAttributeChange is upstream's
// CommandCreoleSvgAttributeChange: "<text attributes>".
func commandCreoleSvgAttributeChange(p *regexp.Regexp) creoleCommand {
	return &creoleRegexCommand{starter: []string{"<t"}, pattern: p, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc2 := stripe.ActualFontConfiguration()
			if attributes, ok := m.group(2); ok {
				fc2 = fc2.ChangeAttributes(SvgAttributesBuild(attributes))
			}
			withFontConfiguration(stripe, fc2, m.mustGroup(3))
			return m.groupLen(1)
		}}
}

// commandCreoleURL is upstream's CommandCreoleUrl: "[[url label]]".
func commandCreoleURL() creoleCommand {
	return &creoleRegexCommand{starter: []string{"[["}, pattern: creoleURLPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			topurl, ok := stripe.SkinParam().Value("topurl")
			urlBuilder := NewURLBuilder(topurl, ok, URLModeStrict)
			url := urlBuilder.URL(m.mustGroup(1))
			stripe.AddURL(url)
			return m.groupLen(1)
		}}
}

// commandCreoleSpace is upstream's CommandCreoleSpace: "<space:n>".
func commandCreoleSpace() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<s"}, pattern: creoleSpacePattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			stripe.AddSpace(javaParseInt(m.mustGroup(2)))
			return m.groupLen(1)
		}}
}

// commandCreoleImg is upstream's CommandCreoleImg: "<img:src{scale=x}>".
func commandCreoleImg() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<i"}, pattern: creoleImgPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			src := m.mustGroup(2)
			scaleGroup, ok := m.group(3)
			scale := ParserScale(scaleGroup, ok, 1)
			if strings.HasPrefix(javaLower(src), "src=") {
				src = src[4:]
			}
			src = eventuallyRemoveStartingAndEndingDoubleQuoteFormat(src, "\"")
			stripe.AddImage(src, scale)
			return m.groupLen(1)
		}}
}

// commandCreoleQrcode is upstream's CommandCreoleQrcode:
// "<qrcode:text{scale=x}>".
func commandCreoleQrcode() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<q"}, pattern: creoleQrcodePattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			scaleGroup, ok := m.group(3)
			stripe.AddQrcode(m.mustGroup(2), ParserScale(scaleGroup, ok, 3))
			return m.groupLen(1)
		}}
}

// commandCreoleOpenIcon is upstream's CommandCreoleOpenIcon: "<&name>".
func commandCreoleOpenIcon() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<#", "<&"}, pattern: creoleOpenIconPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			colorName1, ok1 := m.group(2)
			src := m.mustGroup(3)
			scaleOrColor, ok4 := m.group(4)
			scale := ParserScale(scaleOrColor, ok4, 1)
			colorName, ok := ParserColor(scaleOrColor)
			if !ok4 {
				ok = false
			}
			if ok1 {
				colorName, ok = colorName1, true
			}
			var c color.HColor
			if ok {
				c = stripe.SkinParam().IHtmlColorSet().GetColorOrWhite(colorName)
			}
			stripe.AddOpenIcon(src, scale, c, m.mustGroup(1))
			return m.groupLen(1)
		}}
}

// commandCreoleEmoji is upstream's CommandCreoleEmoji: "<:name:>".
func commandCreoleEmoji() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<#", "<:"}, pattern: creoleEmojiPattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			if stripe != nil {
				colorName1, ok1 := m.group(2)
				emoji := m.mustGroup(3)
				scaleOrColor, ok4 := m.group(4)
				scale := ParserScale(scaleOrColor, ok4, 1)
				colorName, ok := ParserColor(scaleOrColor)
				if !ok4 {
					ok = false
				}
				if ok1 {
					colorName, ok = colorName1, true
				}
				stripe.AddEmoji(emoji, scale, colorName, ok, m.mustGroup(1))
			}
			return m.groupLen(1)
		}}
}

// commandCreoleSprite is upstream's CommandCreoleSprite: "<$name>".
func commandCreoleSprite() creoleCommand {
	return &creoleRegexCommand{starter: []string{"<#", "<$"}, pattern: creoleSpritePattern, sizeGroup: 1,
		execute: func(_ ISkinSimple, m *creoleMatch, stripe *StripeSimple) int {
			fc := stripe.ActualFontConfiguration()
			src := m.mustGroup(3)
			scaleOrColor, ok4 := m.group(4)
			scale := float64(ParserScale(scaleOrColor, ok4, 1)*fc.Size2D()) / 13.0
			colorName, ok := m.group(2)
			if !ok && ok4 {
				colorName, ok = ParserColor(scaleOrColor)
			}
			var c color.HColor
			if ok {
				c = stripe.SkinParam().IHtmlColorSet().GetColorOrWhite(colorName)
			}
			stripe.AddSprite(src, scale, c)
			return m.groupLen(1)
		}}
}
