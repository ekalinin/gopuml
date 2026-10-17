// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommonCommands.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandNope.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandPragma.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandAssumeTransparent.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandSkin.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandMinwidth.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandPage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandRotate.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

// AddCommonCommands1 adds the commands of every UML diagram, upstream's
// CommonCommands.addCommonCommands1: annotations, the common commands, the
// scale commands and the common hide commands.
func AddCommonCommands1[D Diagram](cmds []command.Command[D]) []command.Command[D] {
	cmds = AddTitleCommands(cmds)
	cmds = AddCommonCommands2(cmds)
	cmds = AddCommonScaleCommands(cmds)
	return AddCommonHides(cmds)
}

// AddCommonCommands2 adds upstream's CommonCommands.addCommonCommands2:
// empty lines, pragmas, skin parameters, skins, sprites and styles.
func AddCommonCommands2[D Diagram](cmds []command.Command[D]) []command.Command[D] {
	cmds = append(cmds,
		CommandNope[D](),
		CommandPragma[D](),
		CommandAssumeTransparent[D](),
		CommandSkinParam[D](),
		CommandSkinParamJaws[D](),
		CommandSkinParamMultilines[D](),
		CommandSkin[D](),
		CommandMinwidth[D](),
		CommandPage[D](),
		CommandRotate[D](),
	)
	cmds = append(cmds, CommandFactorySpriteMultiLine[D](), CommandFactorySpriteSingleLine[D]())
	cmds = append(cmds,
		CommandSpriteBase64[D](),
		CommandSpriteMd5[D](),
		CommandSpriteSvg[D](),
		CommandSpriteStdlib[D](),
		CommandSpriteStdlibSvg[D](),
		CommandSpriteSvgMultiline[D](),
		CommandSpriteFile[D](),
	)
	return append(cmds,
		klimt.NewCommandStyleSingleLineCSS[D](),
		klimt.NewCommandStyleMultilinesCSS[D](),
		klimt.NewCommandStyleImport[D](),
	)
}

// CommandNope accepts empty lines.
func CommandNope[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(), regex.SpaceZeroOrMore(), regex.End()),
		func(D, *text.LineLocation, *regex.Result, command.ParserPass) command.Result { return command.Ok() })
}

// CommandPragma is "!pragma name [value]". The preprocessor keeps these
// lines for the diagram.
func CommandPragma[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("!pragma"),
		regex.SpaceOneOrMore(),
		Leaf(1, "NAME", "([A-Za-z_][A-Za-z_0-9]*)"),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), Leaf(1, "VALUE", "(.*)"))),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			name := LowerCase(arg.Str("NAME", 0))
			value, hasValue := arg.Get("VALUE", 0)
			// valueIs is value.equalsIgnoreCase(other), which throws on a
			// null value as upstream.
			valueIs := func(other string) bool {
				if !hasValue {
					ThrowNPE(`Cannot invoke "String.equalsIgnoreCase(String)" because "value" is null`)
				}
				return EqualsIgnoreCase(value, other)
			}
			if EqualsIgnoreCase(name, "svgsize") {
				if !hasValue {
					ThrowNPE(`Cannot invoke "String.contains(java.lang.CharSequence)" because "value" is null`)
				}
				if strings.Contains(value, " ") {
					tokens := strings.FieldsFunc(value, isTokenizerSpace)
					b.Skin().SetSvgSize(tokens[0], tokens[1])
				}
				return command.Ok()
			}
			b.Pragma().Define(name, value, hasValue)
			if EqualsIgnoreCase(name, "layout") && valueIs("smetana") {
				b.SetUseSmetana(true)
			}
			if EqualsIgnoreCase(name, "graphviz_dot") && valueIs("jdot") {
				return command.Error("This directive has been renamed to '!pragma layout smetana'. Please update your diagram.")
			}
			if EqualsIgnoreCase(name, "graphviz_dot") {
				v := value
				if !hasValue {
					v = "null"
				}
				return command.Error("This directive has been renamed to '!pragma layout " + v + "'. Please update your diagram.")
			}
			if EqualsIgnoreCase(name, "layout") && valueIs("elk") {
				b.SetUseElk(true)
			}
			if EqualsIgnoreCase(name, "layout") && valueIs("vizjs") {
				b.Skin().SetUseVizJs(true)
			}
			return command.Ok()
		})
}

// isTokenizerSpace is the default delimiter set of Java's StringTokenizer.
func isTokenizerSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f'
}

// CommandAssumeTransparent is "!assume transparent dark|light", which
// upstream ignores.
func CommandAssumeTransparent[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("!assume"),
		regex.SpaceOneOrMore(),
		Lit("transparent"),
		regex.SpaceOneOrMore(),
		Leaf(1, "TYPE", "(dark|light)"),
		regex.End()),
		func(D, *text.LineLocation, *regex.Result, command.ParserPass) command.Result { return command.Ok() })
}

// CommandSkin is "skin name": the style sheet becomes name.skin.
func CommandSkin[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("skin"),
		regex.SpaceOneOrMore(),
		Leaf(1, "SKIN", `([\w.]+)`), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return d.Titled().LoadSkin(arg.Str("SKIN", 0))
		})
}

// CommandMinwidth is "minwidth n", which upstream ignores.
func CommandMinwidth[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("minwidth"),
		regex.SpaceOneOrMore(),
		Leaf(1, "VALUE", `(\d+)`), regex.End()),
		func(_ D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			JavaParseInt(arg.Str("VALUE", 0))
			return command.Ok()
		})
}

// CommandPage is "page HxV": the image is split in pages.
func CommandPage[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("page"),
		regex.SpaceOneOrMore(),
		Leaf(1, "NB1", `(\d+)`),
		regex.SpaceZeroOrMore(),
		Lit("x*"),
		regex.SpaceZeroOrMore(),
		Leaf(1, "NB2", `(\d+)`), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			horizontal := JavaParseInt(arg.Str("NB1", 0))
			vertical := JavaParseInt(arg.Str("NB2", 0))
			if horizontal <= 0 || vertical <= 0 {
				return command.Error("Argument must be positive")
			}
			d.Titled().SetSplitPages(horizontal, vertical)
			return command.Ok()
		})
}

// CommandRotate is "rotate", which upstream ignores.
func CommandRotate[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(), Lit("rotate"), regex.End()),
		func(D, *text.LineLocation, *regex.Result, command.ParserPass) command.Result { return command.Ok() })
}
