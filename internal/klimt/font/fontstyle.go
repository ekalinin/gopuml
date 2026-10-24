// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/FontStyle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/FontPosition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import "strconv"

// FontStyle is a text style of creole and HTML-like markup.
type FontStyle int

// The font styles, in upstream's order.
const (
	FontStylePlain FontStyle = iota
	FontStyleItalic
	FontStyleBold
	FontStyleUnderline
	FontStyleStrike
	FontStyleWave
	FontStyleBackcolor
)

// String returns upstream's enum name.
func (s FontStyle) String() string {
	switch s {
	case FontStylePlain:
		return "PLAIN"
	case FontStyleItalic:
		return "ITALIC"
	case FontStyleBold:
		return "BOLD"
	case FontStyleUnderline:
		return "UNDERLINE"
	case FontStyleStrike:
		return "STRIKE"
	case FontStyleWave:
		return "WAVE"
	case FontStyleBackcolor:
		return "BACKCOLOR"
	}
	return "FontStyle(" + strconv.Itoa(int(s)) + ")"
}

// Starters returns the markup prefixes that can start the style, for pure
// creole or for HTML-like tags.
func (s FontStyle) Starters(isCreolePure bool) []string {
	switch s {
	case FontStylePlain:
		return []string{"<p", "<P"}
	case FontStyleItalic:
		if isCreolePure {
			return []string{"//"}
		}
		return []string{"<i", "<I"}
	case FontStyleBold:
		if isCreolePure {
			return []string{"**"}
		}
		return []string{"<b", "<B"}
	case FontStyleUnderline:
		if isCreolePure {
			return []string{"__"}
		}
		return []string{"<u", "<U"}
	case FontStyleStrike:
		if isCreolePure {
			return []string{"--"}
		}
		return []string{"<s", "<S", "<d", "<D"}
	case FontStyleWave:
		if isCreolePure {
			return []string{"~~"}
		}
		return []string{"<w"}
	case FontStyleBackcolor:
		return []string{"<b", "<B"}
	}
	panic("font: unknown " + s.String())
}

// MutateFont applies the style to a font: plain resets the face, italic and
// bold change one axis, the other styles keep the font.
func (s FontStyle) MutateFont(font UFont) UFont {
	switch s {
	case FontStylePlain:
		return font.WithFontFace(NormalFace())
	case FontStyleItalic:
		return font.WithFontFace(font.FontFace().WithStyle(StyleItalic))
	case FontStyleBold:
		return font.WithFontFace(font.FontFace().WithWeight(700))
	}
	return font
}

// RegexActivationPattern returns the regular expression of the opening tag.
// As in upstream, the strike pattern is written in ubrex syntax.
func (s FontStyle) RegexActivationPattern() string {
	switch s {
	case FontStylePlain:
		return `\<[pP][lL][aA][iI][nN]\>`
	case FontStyleItalic:
		return `\<[iI]\>`
	case FontStyleBold:
		return `\<[bB]\>`
	case FontStyleUnderline:
		return `\<[uU](?::(#[0-9a-fA-F]{6}|\w+))?\>`
	case FontStyleWave:
		return `\<[wW](?::(#[0-9a-fA-F]{6}|\w+))?\>`
	case FontStyleBackcolor:
		return `\<[bB][aA][cC][kK](?::(#?\w+(?:[-\\|/]#?\w+)?))?\>`
	case FontStyleStrike:
		return "<【strike┇STRIKE┇s┇S┇del┇DEL】〇?〘:〶$XC=【#〇{6}「0〜9a〜fA〜F」┇〇+〴w】〙>"
	}
	return ""
}

// UbrexActivationPattern returns the ubrex pattern of the opening tag.
func (s FontStyle) UbrexActivationPattern() string {
	switch s {
	case FontStylePlain:
		return "<「pP」「lL」「aA」「iI」「nN」>"
	case FontStyleItalic:
		return "<「iI」>"
	case FontStyleBold:
		return "<「bB」>"
	case FontStyleUnderline:
		return "<「uU」〇?〘:〶$XC=【#〇{6}「0〜9a〜fA〜F」┇〇+〴w】〙>"
	case FontStyleWave:
		return "<「wW」〇?〘:〶$XC=【#〇{6}「0〜9a〜fA〜F」┇〇+〴w】〙>"
	case FontStyleBackcolor:
		return "<「bB」「aA」「cC」「kK」〇?〘:〶$XC=〘" +
			"【#〇{6}「0〜9a〜fA〜F」┇〇+〴w 】 " +
			`〇?〘「-\|/」【〇{6}「0〜9a〜fA〜F」┇〇+〴w】 〙` +
			"〙 〙>"
	case FontStyleStrike:
		return "<【strike┇STRIKE┇s┇S┇del┇DEL】〇?〘:〶$XC=【#〇{6}「0〜9a〜fA〜F」┇〇+〴w】〙>"
	}
	return ""
}

// UbrexDeactivationPattern returns the ubrex pattern of the closing tag.
func (s FontStyle) UbrexDeactivationPattern() string {
	switch s {
	case FontStylePlain:
		return "</「pP」「lL」「aA」「iI」「nN」>"
	case FontStyleItalic:
		return "</「iI」>"
	case FontStyleBold:
		return "</「bB」>"
	case FontStyleUnderline:
		return "</「uU」>"
	case FontStyleWave:
		return "</「wW」>"
	case FontStyleBackcolor:
		return "</「bB」「aA」「cC」「kK」>"
	case FontStyleStrike:
		return "</【strike┇STRIKE┇s┇S┇del┇DEL】>"
	}
	return ""
}

// RegexDeactivationPattern returns the regular expression of the closing
// tag.
func (s FontStyle) RegexDeactivationPattern() string {
	switch s {
	case FontStylePlain:
		return `\</[pP][lL][aA][iI][nN]\>`
	case FontStyleItalic:
		return `\</[iI]\>`
	case FontStyleBold:
		return `\</[bB]\>`
	case FontStyleUnderline:
		return `\</[uU]\>`
	case FontStyleWave:
		return `\</[wW]\>`
	case FontStyleBackcolor:
		return `\</[bB][aA][cC][kK]\>`
	case FontStyleStrike:
		return `\</(?:s|S|strike|STRIKE|del|DEL)\>`
	}
	return ""
}

// CanHaveExtendedColor reports whether the style takes a color: underline,
// wave, backcolor and strike.
func (s FontStyle) CanHaveExtendedColor() bool {
	switch s {
	case FontStyleUnderline, FontStyleWave, FontStyleBackcolor, FontStyleStrike:
		return true
	}
	return false
}

// UbrexCreoleSyntax returns the creole markup of the style. Upstream throws
// for plain and backcolor; this panics.
func (s FontStyle) UbrexCreoleSyntax() string {
	switch s {
	case FontStyleItalic:
		return "//"
	case FontStyleBold:
		return "**"
	case FontStyleUnderline:
		return "__"
	case FontStyleWave:
		return "~~"
	case FontStyleStrike:
		return "--"
	}
	panic("font: no creole syntax for " + s.String())
}

// FontPosition is the vertical position of text: normal, superscript
// (exposant) or subscript (indice).
type FontPosition int

// The font positions.
const (
	FontPositionNormal FontPosition = iota
	FontPositionExposant
	FontPositionIndice
)

// String returns upstream's enum name.
func (p FontPosition) String() string {
	switch p {
	case FontPositionNormal:
		return "NORMAL"
	case FontPositionExposant:
		return "EXPOSANT"
	case FontPositionIndice:
		return "INDICE"
	}
	return "FontPosition(" + strconv.Itoa(int(p)) + ")"
}

// Space returns the vertical shift of the text.
func (p FontPosition) Space() int {
	switch p {
	case FontPositionExposant:
		return -6
	case FontPositionIndice:
		return 3
	}
	return 0
}

// Mute returns the font for the position: superscript and subscript text
// is 3 points smaller, with a minimum of 2.
func (p FontPosition) Mute(font UFont) UFont {
	if p == FontPositionNormal {
		return font
	}
	return font.WithSize(float32(max(font.Size()-3, 2)))
}

// HtmlTag returns the HTML tag of the position, "sup" or "sub". Upstream
// throws for the normal position; this panics.
func (p FontPosition) HtmlTag() string {
	switch p {
	case FontPositionExposant:
		return "sup"
	case FontPositionIndice:
		return "sub"
	}
	panic("font: no HTML tag for " + p.String())
}
