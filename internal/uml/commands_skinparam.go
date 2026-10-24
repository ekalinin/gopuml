// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSkinParam.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSkinParamJaws.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandSkinParamMultilines.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/SkinLoader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

// CommandSkinParam is "skinparam name value".
func CommandSkinParam[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Leaf(1, "TYPE", "(skinparam|skinparamlocked)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "NAME", `([\w.]*(?:\<\<[^<>]*\>\>)?[\w.]*)`),
		regex.SpaceOneOrMore(),
		Leaf(1, "VALUE", "([^{}]*)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			name := arg.Str("NAME", 0)
			if EqualsIgnoreCase("handwritten", name) {
				b.AddWarning(klimt.NewWarning("Please use '!option handwritten true' to enable handwritten "))
			}
			if EqualsIgnoreCase("ParticipantPadding", name) {
				b.AddWarning(klimt.NewWarning("Please use CSS style instead of skinparam ParticipantPadding"))
			}
			if EqualsIgnoreCase("padding", name) {
				b.AddWarning(klimt.NewWarning("Please use CSS style instead of skinparam padding"))
			}
			b.SetParam(name, arg.Str("VALUE", 0))
			b.SetSkinParamUsed(true)
			return command.Ok()
		})
}

// CommandSkinParamJaws is a whole "skinparam {" block that the
// preprocessor joined on one line with its newline character.
func CommandSkinParamJaws[D Diagram]() command.Command[D] {
	newline := string(klimt.JawsBlockE1Newline)
	multi := commandSkinParamMultilines[D]()
	return Single(regex.NewConcat(regex.Start(),
		regex.OneOrMore("FULL", regex.NewConcat(
			Leaf(1, "TYPE", "(skinparam|skinparamlocked)"),
			regex.SpaceOneOrMore(),
			Lit(".*"+newline+".*"))),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result {
			var lines command.BlocLines
			for _, line := range JavaSplit(arg.Str("FULL", 0), newline) {
				lines = lines.Add(text.NewStringLocated(line, nil))
				if text.JavaTrim(line) == "}" {
					if res := multi.Execute(d, lines, pass); !res.IsOK() {
						return res
					}
					lines = command.BlocLines{}
				}
			}
			d.Titled().SetSkinParamUsed(true)
			return command.Ok()
		})
}

// JavaSplit is Java's String.split with a literal separator and no limit:
// trailing empty strings are removed, and a string without the separator
// gives itself.
func JavaSplit(s, sep string) []string {
	if !strings.Contains(s, sep) {
		return []string{s}
	}
	parts := strings.Split(s, sep)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// commentSingleLine is upstream's CommandMultilinesComment.COMMENT_SINGLE_LINE.
var commentSingleLine = Compile(`^[%s]*([%q].*||/[%q].*[%q]/[%s]*)$`)

// skinLoaderLine is upstream's SkinLoader.p1.
var skinLoaderLine = Compile(`^([\w.]*(?:\<\<.*\>\>)?[\w.]*)[%s]+(?:(\{)|(.*))$|^\}?$`)

// CommandSkinParamMultilines is "skinparam [name] {" ... "}".
func CommandSkinParamMultilines[D Diagram]() command.Command[D] {
	return commandSkinParamMultilines[D]()
}

func commandSkinParamMultilines[D Diagram]() *MultilinesBracket[D] {
	start := Compile(`^skinparam[%s]*(?:[%s]+([\w.]*(?:\<\<.*\>\>)?[\w.]*))?[%s]*\{$`)
	return &MultilinesBracket[D]{
		Start: start,
		Consistent: func(line string, _ int) bool {
			line = text.Trin(line)
			if commentSingleLine.MatchString(line) {
				return true
			}
			return skinLoaderLine.MatchString(line)
		},
		Exec: func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
			lines = expandsNewline(lines)
			first := lines.At(0).Trimmed().S
			m := start.FindStringSubmatchIndex(first)
			if m == nil {
				return command.Error("Bad syntax for skinparam")
			}
			group1, ok1 := group(first, m, 1)
			d.Titled().SetSkinParamUsed(true)
			return skinLoaderExecute(d.Titled(), lines, group1, ok1)
		},
	}
}

// expandsNewline is upstream's BlocLines.expandsNewline(true): each line is
// split at the preprocessor's newline character.
func expandsNewline(lines command.BlocLines) command.BlocLines {
	newline := string(klimt.JawsBlockE1Newline)
	var out []text.StringLocated
	for _, l := range lines.Lines() {
		for _, s := range JavaSplit(l.S, newline) {
			out = append(out, text.StringLocated{S: s, Location: l.Location, PreprocessorError: l.PreprocessorError})
		}
	}
	return command.NewBlocLines(out)
}

// skinLoaderExecute is upstream's SkinLoader.execute: the inner lines of a
// skinparam block set "context + name" parameters, nested blocks extend
// the context.
func skinLoaderExecute(b *Base, lines command.BlocLines, group1 string, hasGroup1 bool) command.Result {
	var context []string
	if hasGroup1 {
		context = append(context, group1)
	}
	lines = SubExtract(lines, 1, 1).Trim().RemoveEmptyLines()
	for _, s := range lines.Lines() {
		if s.S == "}" {
			if len(context) == 0 {
				ThrowIndexOutOfBounds(-1, 0)
			}
			context = context[:len(context)-1]
			continue
		}
		m := skinLoaderLine.FindStringSubmatchIndex(s.S)
		if m == nil {
			Throw(IllegalStateException, "")
		}
		g1, _ := group(s.S, m, 1)
		if _, ok := group(s.S, m, 2); ok {
			context = append(context, g1)
		} else if g3, ok := group(s.S, m, 3); ok {
			b.SetParam(strings.Join(context, "")+g1, g3)
		} else {
			Throw(IllegalStateException, "."+s.S+".")
		}
	}
	return command.Ok()
}
