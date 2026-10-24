// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScale.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScaleWidthAndHeight.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScaleWidthOrHeight.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScaleMaxWidth.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScaleMaxHeight.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandScaleMaxWidthAndHeight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
)

// AddCommonScaleCommands adds the "scale" commands.
func AddCommonScaleCommands[D Diagram](cmds []command.Command[D]) []command.Command[D] {
	return append(cmds,
		CommandScale[D](),
		CommandScaleWidthAndHeight[D](),
		CommandScaleWidthOrHeight[D](),
		CommandScaleMaxWidth[D](),
		CommandScaleMaxHeight[D](),
		CommandScaleMaxWidthAndHeight[D](),
	)
}

// CommandScale is "scale f" or "scale a/b".
func CommandScale[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Leaf(1, "SCALE", "([0-9.]+)"),
		regex.Optional(regex.NewConcat(regex.SpaceZeroOrMore(), Lit("/"), regex.SpaceZeroOrMore(),
			Leaf(1, "DIV", "([0-9.]+)"))),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			scale := JavaParseDouble(arg.Str("SCALE", 0))
			if scale == 0 {
				return command.Error("Scale cannot be zero")
			}
			if div, ok := arg.Get("DIV", 0); ok {
				v := JavaParseDouble(div)
				if v == 0 {
					return command.Error("Scale cannot be zero")
				}
				scale /= v
			}
			d.Titled().SetScale(&Scale{Kind: ScaleSimple, A: scale})
			return command.Ok()
		})
}

// CommandScaleWidthAndHeight is "scale WxH".
func CommandScaleWidthAndHeight[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Leaf(1, "WIDTH", "([0-9.]+)"),
		regex.SpaceZeroOrMore(),
		Lit("[*x]"),
		regex.SpaceZeroOrMore(),
		Leaf(1, "HEIGHT", "([0-9.]+)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			width := JavaParseDouble(arg.Str("WIDTH", 0))
			height := JavaParseDouble(arg.Str("HEIGHT", 0))
			d.Titled().SetScale(&Scale{Kind: ScaleWidthAndHeight, A: width, B: height})
			return command.Ok()
		})
}

// CommandScaleWidthOrHeight is "scale n width" or "scale n height".
func CommandScaleWidthOrHeight[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Leaf(1, "VALUE", "([0-9.]+)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "WIDTH", "(width|height)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			size := JavaParseDouble(arg.Str("VALUE", 0))
			if EqualsIgnoreCase("width", arg.Str("WIDTH", 0)) {
				d.Titled().SetScale(&Scale{Kind: ScaleWidth, A: size})
			} else {
				d.Titled().SetScale(&Scale{Kind: ScaleHeight, A: size})
			}
			return command.Ok()
		})
}

// CommandScaleMaxWidth is "scale max n width".
func CommandScaleMaxWidth[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Lit("max"),
		regex.SpaceOneOrMore(),
		Leaf(1, "WIDTH", "([0-9.]+)"),
		regex.SpaceOneOrMore(),
		Lit("width"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.Titled().SetScale(&Scale{Kind: ScaleMaxWidth, A: JavaParseDouble(arg.Str("WIDTH", 0))})
			return command.Ok()
		})
}

// CommandScaleMaxHeight is "scale max n height".
func CommandScaleMaxHeight[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Lit("max"),
		regex.SpaceOneOrMore(),
		Leaf(1, "HEIGHT", "([0-9.]+)"),
		regex.SpaceOneOrMore(),
		Lit("height"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.Titled().SetScale(&Scale{Kind: ScaleMaxHeight, A: JavaParseDouble(arg.Str("HEIGHT", 0))})
			return command.Ok()
		})
}

// CommandScaleMaxWidthAndHeight is "scale max WxH".
func CommandScaleMaxWidthAndHeight[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("scale"),
		regex.SpaceOneOrMore(),
		Lit("max"),
		regex.SpaceOneOrMore(),
		Leaf(1, "WIDTH", "([0-9.]+)"),
		regex.SpaceZeroOrMore(),
		Lit("[*x]"),
		regex.SpaceZeroOrMore(),
		Leaf(1, "HEIGHT", "([0-9.]+)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			width := JavaParseDouble(arg.Str("WIDTH", 0))
			height := JavaParseDouble(arg.Str("HEIGHT", 0))
			d.Titled().SetScale(&Scale{Kind: ScaleMaxWidthAndHeight, A: width, B: height})
			return command.Ok()
		})
}
