// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/AtomicParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"unicode"
)

// The parser, upstream's AtomicParser.

func ubrexClosingBracket(input *ubrexText, open, closing uint16) int {
	level := 0
	for i := 1; i < input.length(); i++ {
		switch input.charAt(i) {
		case open:
			level++
		case closing:
			if level == 0 {
				return i
			}
			level--
			if level < 0 {
				throw(illegalArgument, "")
			}
		}
	}
	return -1
}

func ubrexParseSingle(input *ubrexText) ubrexChallenge {
	result := ubrexParse(input)
	if len(result) != 1 {
		throw(unsupportedOperation, "")
	}
	return result[0]
}

// ubrexParse parses the next element of input and consumes it.
func ubrexParse(input *ubrexText) []ubrexChallenge {
	switch input.charAt(0) {
	case '┇':
		throw(illegalArgument, "")
	case '〒':
		return []ubrexChallenge{ubrexManageLookAround(input)}
	case '【':
		return []ubrexChallenge{ubrexManageAlternative(input)}
	case '〇':
		return ubrexManageQuantifier(input)
	case '〄':
		return ubrexManageUpTo(input)
	case '〘':
		return []ubrexChallenge{ubrexManageGroup(input)}
	case '「':
		return []ubrexChallenge{ubrexManageCharacterSet(input)}
	case '〴':
		return []ubrexChallenge{ubrexManageClass(input)}
	case '〶':
		return ubrexManageNamed(input)
	case '〃':
		input.jump(1)
		return []ubrexChallenge{newUbrexSingleChar('"')}
	}
	return []ubrexChallenge{ubrexManageRegularCharacter(input)}
}

func ubrexManageLookAround(input *ubrexText) ubrexChallenge {
	input.jump(1)
	look, ok := ubrexLookAroundFrom(input)
	if !ok {
		throw(unsupportedOperation, "Syntax error")
	}
	input.jump(look.definitionSize())
	if look == ubrexEndOfTextLook {
		return ubrexEndOfText{}
	}
	return &ubrexLook{origin: ubrexParseSingle(input), look: look}
}

func ubrexManageClass(input *ubrexText) ubrexChallenge {
	input.jump(1)
	negative := unicode.IsUpper(rune(input.charAt(0)))
	raw := ubrexCharClassRawFromDefinition(input)
	input.jump(raw.definitionLength())
	return &ubrexCharClass{raw: raw, negative: negative}
}

func ubrexManageQuantifier(input *ubrexText) []ubrexChallenge {
	operator := input.charAt(1)
	input.jump(2)
	switch operator {
	case '{':
		repetition := ubrexRepetitionParse(input)
		return []ubrexChallenge{&ubrexRepeat{origin: ubrexParseSingle(input), repetition: repetition}}
	case 'l':
		input.jump(1)
		origin := ubrexParseSingle(input)
		remaining := ubrexParseAndBuildFromText(input)
		// The lazzy quantifier only peeks at remaining as a stop condition.
		// The remaining challenges are returned as siblings, so they get
		// matched normally by the parent CompositeList/CompositeNamed.
		return append([]ubrexChallenge{&ubrexLazyOneOrMore{origin: origin, stop: remaining}},
			remaining.challenges...)
	}
	origin := ubrexParseSingle(input)
	switch operator {
	case '+':
		return []ubrexChallenge{&ubrexOneOrMore{origin}}
	case '*':
		return []ubrexChallenge{&ubrexZeroOrMore{origin}}
	case '?':
		return []ubrexChallenge{&ubrexOptional{origin}}
	}
	throw(unsupportedOperation, "wip01")
	return nil
}

func ubrexManageUpTo(input *ubrexText) []ubrexChallenge {
	operator := input.charAt(1)
	if operator == '>' {
		input.jump(2)
		p2 := ubrexParseSingle(input)
		// This is a terrible hack, because if this is used with named group,
		// we do not want p2 to be captured
		return []ubrexChallenge{&ubrexUpTo{p2}, p2}
	}
	if operator != '+' {
		throw(unsupportedOperation, "manageQuantifierUpTo1")
	}
	input.jump(2)
	ubrexSkipSpaces(input)
	p1 := ubrexParseSingle(input)
	ubrexSkipSpaces(input)
	if input.charAt(0) != '-' || input.charAt(1) != '>' {
		throw(unsupportedOperation, "manageQuantifierUpTo2")
	}
	input.jump(2)
	ubrexSkipSpaces(input)
	p2 := ubrexParseSingle(input)
	// This is a terrible hack, because if this is used with named group, we
	// do not want p2 to be captured
	return []ubrexChallenge{&ubrexOneOrMoreUpTo{origin: p1, end: p2}, p2}
}

func ubrexSkipSpaces(input *ubrexText) {
	for input.charAt(0) == ' ' {
		input.jump(1)
	}
}

func ubrexManageGroup(input *ubrexText) ubrexChallenge {
	end := ubrexClosingBracket(input, '〘', '〙')
	if end == -1 {
		throw(unsupportedOperation, "wip99")
	}
	result := ubrexParseAndBuild(input.subSequence(1, end))
	input.jump(end + 1)
	return result
}

func ubrexManageAlternative(input *ubrexText) ubrexChallenge {
	result := &ubrexAlternative{}
	start := 1
	level := 0
	for i := 1; i < input.length(); i++ {
		switch ch := input.charAt(i); {
		case ch == '【':
			level++
		case level == 0 && ch == '┇':
			result.alternatives = append(result.alternatives, ubrexParseAndBuild(input.subSequence(start, i)))
			start = i + 1
		case ch == '】':
			if level == 0 {
				result.alternatives = append(result.alternatives, ubrexParseAndBuild(input.subSequence(start, i)))
				input.jump(i + 1)
				return result
			}
			level--
			if level < 0 {
				throw(illegalArgument, "")
			}
		}
	}
	throw(unsupportedOperation, "wip32")
	return nil
}

func ubrexManageNamed(input *ubrexText) []ubrexChallenge {
	if input.charAt(1) != '$' {
		throw(unsupportedOperation, "varname must have a $")
	}
	input.jump(2)
	var name []rune
	for {
		ch := input.charAt(0)
		input.jump(1)
		if ch == '=' {
			if len(name) == 0 {
				throw(unsupportedOperation, "no name!")
			}
			parsed := ubrexParse(input)
			// The first element is the named group's content. Any additional
			// elements (e.g. siblings from a lazzy quantifier) must live
			// outside the named group, at the parent level.
			named := &ubrexNamed{name: string(name), challenges: []ubrexChallenge{parsed[0]}}
			return append([]ubrexChallenge{named}, parsed[1:]...)
		}
		if !javaIsIdentifierPart(rune(ch)) {
			throw(unsupportedOperation, "Unsupported name!")
		}
		name = append(name, rune(ch))
	}
}

// javaIsIdentifierPart is Java's Character.isJavaIdentifierPart for the
// characters of variable names: letters, digits, '_', '$' and the
// connecting punctuation.
func javaIsIdentifierPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$' || unicode.Is(unicode.Pc, r) ||
		unicode.Is(unicode.Nl, r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) ||
		unicode.Is(unicode.Sc, r)
}

func ubrexManageCharacterSet(input *ubrexText) ubrexChallenge {
	end := input.indexOf('」')
	if end == -1 {
		throw(unsupportedOperation, "wip80")
	}
	result := ubrexCharSetBuild(input.subSequence(1, end))
	input.jump(end + 1)
	return result
}

func ubrexManageRegularCharacter(input *ubrexText) ubrexChallenge {
	ch := input.charAt(0)
	if ch == ' ' {
		throw(illegalState, "no space allowed")
	}
	if ch == '∙' {
		ch = ' '
	}
	input.jump(1)
	return newUbrexSingleChar(ch)
}
