// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/TLineType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/FoxSignature.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// TLineType is the kind of a preprocessor line.
type TLineType int

// Line types, in upstream order.
const (
	linePlain TLineType = iota
	lineAffectationDefine
	lineAffectation
	lineAssert
	lineIf
	lineIfdef
	lineUndef
	lineIfndef
	lineElse
	lineElseif
	lineEndif
	lineWhile
	lineEndwhile
	lineForeach
	lineEndforeach
	lineDeclareReturnFunction
	lineDeclareProcedure
	lineEndFunction
	lineReturn
	lineLegacyDefine
	lineLegacyDefinelong
	lineTheme
	lineInclude
	lineIncludeDef
	lineImport
	lineStartsub
	lineEndsub
	lineIncludesub
	lineLog
	lineDumpMemory
	lineCommentSimple
	lineCommentLongStart
	lineOption
)

var tLineTypeNames = [...]string{"PLAIN", "AFFECTATION_DEFINE", "AFFECTATION", "ASSERT", "IF", "IFDEF", "UNDEF",
	"IFNDEF", "ELSE", "ELSEIF", "ENDIF", "WHILE", "ENDWHILE", "FOREACH", "ENDFOREACH", "DECLARE_RETURN_FUNCTION",
	"DECLARE_PROCEDURE", "END_FUNCTION", "RETURN", "LEGACY_DEFINE", "LEGACY_DEFINELONG", "THEME", "INCLUDE",
	"INCLUDE_DEF", "IMPORT", "STARTSUB", "ENDSUB", "INCLUDESUB", "LOG", "DUMP_MEMORY", "COMMENT_SIMPLE",
	"COMMENT_LONG_START", "OPTION"}

// String returns the upstream enum name, used in error messages.
func (t TLineType) String() string { return tLineTypeNames[t] }

// identifier is upstream's IDENTIFIER_WITH_UNICODE_SURROGATES_SUPPORT. Its
// Java surrogate ranges (\uD800-\uDBFF and \uDC00-\uDFFF) match the
// supplementary characters of the line (the reference jar accepts an emoji
// in "!$x<emoji> = 1"), so they become the code point range U+10000..U+10FFFF.
const identifier = `[\p{L}\x{10000}-\x{10FFFF}_][\p{L}\x{10000}-\x{10FFFF}_0-9]*`

// tLinePattern is one upstream RegexConcat. Java compiles the patterns with
// CASE_INSENSITIVE, which folds ASCII letters only; the patterns are written
// in lower case and matched against the line with ASCII letters lowered.
// needsSpace reproduces the FoxSignature pre-check of patterns that contain
// a [%s]+ leaf: the line must contain a space, \t, \r, \n, \f or U+00A0
// (but not only a vertical tab, which [%s] matches).
type tLinePattern struct {
	re         *regexp.Regexp
	needsSpace bool
}

func newTLinePattern(pattern string, needsSpace bool) tLinePattern {
	t, err := regex.Transform(pattern)
	if err != nil {
		panic(err)
	}
	return tLinePattern{re: regexp.MustCompile(t), needsSpace: needsSpace}
}

func (p tLinePattern) match(lowered string) bool {
	if p.needsSpace && !strings.ContainsAny(lowered, " \t\r\n\f\u00a0") {
		return false
	}
	return p.re.MatchString(lowered)
}

func simpleKeyword(word string) tLinePattern {
	return newTLinePattern(`^[%s]*`+word+`\b`, false)
}

// The patterns are compiled once and only read afterwards.
var (
	patternLegacyDefine      = newTLinePattern(`^[%s]*!define[%s]+`+identifier+`\(`, true)
	patternLegacyDefineLong  = newTLinePattern(`^[%s]*!definelong[%s]+`+identifier+`\b`, true)
	patternAffectationDefine = newTLinePattern(`^[%s]*!define[%s]+`+identifier+`\b`, true)
	patternAffectation       = newTLinePattern(`^[%s]*![%s]*(local|global)?[%s]*\$?`+identifier+`[%s]*\??=`, false)
	patternCommentSimple1    = newTLinePattern(`^[%s]*'`, false)
	patternCommentSimple2    = newTLinePattern(`^[%s]*/'.*'/[%s]*$`, false)
	patternCommentLongStart  = newTLinePattern(`^[%s]*/'`, false)
	patternIfdef             = simpleKeyword("!ifdef")
	patternUndef             = simpleKeyword("!undef")
	patternIfndef            = simpleKeyword("!ifndef")
	patternAssert            = simpleKeyword("!assert")
	patternIf                = simpleKeyword("!if")
	patternDeclareFunction   = newTLinePattern(`^[%s]*!(unquoted\s|final\s)*function[%s]+\$?`+identifier, true)
	patternDeclareProcedure  = newTLinePattern(`^[%s]*!(unquoted\s|final\s)*procedure[%s]+\$?`+identifier, true)
	patternElse              = simpleKeyword("!else")
	patternElseIf            = simpleKeyword("!elseif")
	patternEndIf             = simpleKeyword("!endif")
	patternWhile             = simpleKeyword("!while")
	patternEndWhile          = simpleKeyword("!endwhile")
	patternForeach           = simpleKeyword("!foreach")
	patternEndForeach        = simpleKeyword("!endfor")
	patternEndFunction       = newTLinePattern(`^[%s]*!end[%s]*(function|definelong|procedure)\b`, false)
	patternReturn            = simpleKeyword("!return")
	patternTheme             = simpleKeyword("!theme")
	patternInclude           = newTLinePattern(`^[%s]*!include[%s]*(?:(url|_many|_once))?\b`, false)
	patternIncludeDef        = simpleKeyword("!includedef")
	patternImport            = simpleKeyword("!import")
	patternStartSub          = simpleKeyword("!startsub")
	patternEndSub            = simpleKeyword("!endsub")
	patternIncludeSub        = simpleKeyword("!includesub")
	patternLog               = simpleKeyword("!log")
	patternDumpMemory        = simpleKeyword("!dump_memory")
	patternOption            = simpleKeyword("!option")
)

// lineTypeChecks lists the directive patterns in upstream's order.
var lineTypeChecks = []struct {
	pattern *tLinePattern
	typ     TLineType
}{
	{&patternLegacyDefine, lineLegacyDefine},
	{&patternLegacyDefineLong, lineLegacyDefinelong},
	{&patternAffectationDefine, lineAffectationDefine},
	{&patternAffectation, lineAffectation},
	{&patternIfdef, lineIfdef},
	{&patternUndef, lineUndef},
	{&patternIfndef, lineIfndef},
	{&patternAssert, lineAssert},
	{&patternIf, lineIf},
	{&patternDeclareFunction, lineDeclareReturnFunction},
	{&patternDeclareProcedure, lineDeclareProcedure},
	{&patternElse, lineElse},
	{&patternElseIf, lineElseif},
	{&patternEndIf, lineEndif},
	{&patternWhile, lineWhile},
	{&patternEndWhile, lineEndwhile},
	{&patternForeach, lineForeach},
	{&patternEndForeach, lineEndforeach},
	{&patternEndFunction, lineEndFunction},
	{&patternReturn, lineReturn},
	{&patternTheme, lineTheme},
	{&patternInclude, lineInclude},
	{&patternIncludeDef, lineIncludeDef},
	{&patternImport, lineImport},
	{&patternStartSub, lineStartsub},
	{&patternEndSub, lineEndsub},
	{&patternIncludeSub, lineIncludesub},
	{&patternLog, lineLog},
	{&patternDumpMemory, lineDumpMemory},
	{&patternOption, lineOption},
}

// lowerASCII lowers the ASCII letters of s only, like Java's
// CASE_INSENSITIVE matching without UNICODE_CASE.
func lowerASCII(s string) string {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 'A' && c <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if b[j] >= 'A' && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return s
}

// getTLineType is TLineType.getFromLineInternal.
func getTLineType(s string) TLineType {
	lowered := lowerASCII(s)
	if patternCommentSimple1.match(lowered) || patternCommentSimple2.match(lowered) {
		return lineCommentSimple
	}
	if patternCommentLongStart.match(lowered) && !strings.Contains(s, "'/") {
		return lineCommentLongStart
	}
	// StringLocated.containsExclamationMark
	if !strings.Contains(s, "!") {
		return linePlain
	}
	for _, c := range lineTypeChecks {
		if c.pattern.match(lowered) {
			return c.typ
		}
	}
	return linePlain
}

// isQuote is TLineType.isQuote.
func isQuote(ch uint16) bool { return ch == '"' || ch == '\'' }

// isSpaceChar is TLineType.isSpaceChar.
func isSpaceChar(ch uint16) bool { return isJavaSpaceChar(ch) }

// isLatinDigit is TLineType.isLatinDigit.
func isLatinDigit(ch uint16) bool { return ch >= '0' && ch <= '9' }

// isLetterOrEmojiOrUnderscoreOrDigit is TLineType.isLetterOrEmojiOrUnderscoreOrDigit.
func isLetterOrEmojiOrUnderscoreOrDigit(ch uint16) bool {
	return isLetterOrUnderscore(ch) || isLatinDigit(ch) || isEmoji(ch)
}

// isLetterOrEmojiOrUnderscoreOrDollar is TLineType.isLetterOrEmojiOrUnderscoreOrDollar.
func isLetterOrEmojiOrUnderscoreOrDollar(ch uint16) bool {
	return isLetterOrUnderscore(ch) || ch == '$' || isEmoji(ch)
}

func isLetterOrUnderscore(ch uint16) bool { return isJavaLetter(ch) || ch == '_' }

// isEmoji reports whether ch is a surrogate, half of a supplementary
// character such as an emoji.
func isEmoji(ch uint16) bool { return ch >= 0xd800 && ch <= 0xdfff }
