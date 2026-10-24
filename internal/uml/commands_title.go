// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandTitle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilinesTitle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandCaption.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilinesCaption.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandLegend.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilinesLegend.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandHeader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilinesHeader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandFooter.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilinesFooter.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMainframe.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandNamespaceSeparator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// AddTitleCommands adds the annotation commands, upstream's
// CommonCommands.addTitleCommands.
func AddTitleCommands[D Diagram](cmds []command.Command[D]) []command.Command[D] {
	return append(cmds,
		CommandTitle[D](),
		CommandMainframe[D](),
		CommandCaption[D](),
		CommandMultilinesCaption[D](),
		CommandMultilinesTitle[D](),
		CommandMultilinesLegend[D](),
		CommandLegend[D](),
		CommandFooter[D](),
		CommandMultilinesFooter[D](),
		CommandHeader[D](),
		CommandMultilinesHeader[D](),
		CommandNamespaceSeparator[D](),
	)
}

// titleLike is the pattern of "title", "caption" and "legend" on one line:
// the keyword, ":" or spaces, then a quoted text or a text with at least one
// letter, digit, '_' or '.'.
func titleLike(keyword, name string) *regex.Concat {
	return regex.NewConcat(regex.Start(),
		Lit(keyword),
		Lit("(?:[%s]*:[%s]*|[%s]+)"),
		regex.Or("",
			Leaf(1, name+"1", "[%g](.*)[%g]"),
			Leaf(1, name+"2", "(.*[%pLN_.].*)")),
		regex.End())
}

// CommandTitle is "title text".
func CommandTitle[D Diagram]() command.Command[D] {
	return Single(titleLike("title", "TITLE"),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			s, ok := arg.Lazy("TITLE", 0)
			display := klimt.DisplayWithNewlinesString(b.Pragma(), s, ok)
			b.SetTitle(DisplayPositionedSingle(location, display, geom.HorizontalAlignmentCenter, geom.VerticalAlignmentTop))
			return command.Ok()
		})
}

// CommandCaption is "caption text".
func CommandCaption[D Diagram]() command.Command[D] {
	return Single(titleLike("caption", "DISPLAY"),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			s, ok := arg.Lazy("DISPLAY", 0)
			display := klimt.DisplayWithNewlinesString(b.Pragma(), s, ok)
			b.SetCaption(DisplayPositionedSingle(location, display, geom.HorizontalAlignmentCenter, geom.VerticalAlignmentBottom))
			return command.Ok()
		})
}

// CommandLegend is "legend text"; unlike the other annotations, upstream
// does not record its source line.
func CommandLegend[D Diagram]() command.Command[D] {
	return Single(titleLike("legend", "LEGEND"),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			s, ok := arg.Lazy("LEGEND", 0)
			display := klimt.DisplayWithNewlinesString(b.Pragma(), s, ok)
			setLegend(d, DisplayPositionedSingle(nil, display, geom.HorizontalAlignmentCenter, geom.VerticalAlignmentBottom))
			return command.Ok()
		})
}

// bodyDisplay is the text between the first and the last line of a block,
// with the shared leading spaces removed.
func bodyDisplay(lines command.BlocLines) *klimt.Display {
	lines = SubExtract(lines, 1, 1).RemoveEmptyColumns()
	return klimt.DisplayCreateFoo(lines.Lines())
}

// CommandMultilinesTitle is "title" ... "end title".
func CommandMultilinesTitle[D Diagram]() command.Command[D] {
	return &Multilines[D]{
		Pattern: "^title$",
		Start:   Compile("^title$"),
		End:     Compile("^end[%s]?title$"),
		Exec: func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
			location := Location(lines)
			strings := bodyDisplay(lines)
			if strings.Size() > 0 {
				d.Titled().SetTitle(DisplayPositionedSingle(location, strings.ReplaceBackslashT(),
					geom.HorizontalAlignmentCenter, geom.VerticalAlignmentTop))
				return command.Ok()
			}
			return command.Error("No title defined")
		},
	}
}

// CommandMultilinesCaption is "caption" ... "end caption".
func CommandMultilinesCaption[D Diagram]() command.Command[D] {
	return &Multilines[D]{
		Pattern: "^caption$",
		Start:   Compile("^caption$"),
		End:     Compile("^end[%s]?caption$"),
		Exec: func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
			location := Location(lines)
			strings := bodyDisplay(lines)
			if strings.Size() > 0 {
				d.Titled().SetCaption(DisplayPositionedSingle(location, strings.ReplaceBackslashT(),
					geom.HorizontalAlignmentCenter, geom.VerticalAlignmentBottom))
				return command.Ok()
			}
			return command.Error("No caption defined")
		},
	}
}

// CommandMultilinesLegend is "legend [top|bottom] [left|right|center]" ...
// "end legend".
func CommandMultilinesLegend[D Diagram]() command.Command[D] {
	c := &Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			Lit("legend"),
			regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), Leaf(1, "VALIGN", "(top|bottom)"))),
			regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), Leaf(1, "ALIGN", "(left|right|center)"))),
			regex.End()),
		Strategy: RemoveStartingQuote,
		TrimEnd:  TrimBoth,
		End:      Compile("^end[%s]?legend$"),
	}
	c.Exec = func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
		location := Location(lines)
		lines = TrimSmart(lines, 1)
		line0 := c.StartArg(lines)
		align, alignOK := line0.Get("ALIGN", 0)
		valign, _ := line0.Get("VALIGN", 0)
		strings := bodyDisplay(lines)
		if strings.Size() > 0 {
			valignment := geom.VerticalAlignmentFromString(valign)
			alignment := geom.HorizontalAlignmentCenter
			if alignOK {
				if a, ok := geom.HorizontalAlignmentFromString(align); ok {
					alignment = a
				}
			}
			setLegend(d, DisplayPositionedSingle(location, strings.ReplaceBackslashT(), alignment, valignment))
			return command.Ok()
		}
		return command.Error("No legend defined")
	}
	return c
}

// headerFooterLike is the pattern of "header" and "footer" on one line.
func headerFooterLike(keyword string) *regex.Concat {
	return regex.NewConcat(regex.Start(),
		regex.Optional(Leaf(1, "POSITION", "(left|right|center)")),
		regex.SpaceZeroOrMore(),
		Lit(keyword),
		regex.Or("",
			regex.NewConcat(regex.SpaceZeroOrMore(), Lit(":"), regex.SpaceZeroOrMore()),
			regex.SpaceOneOrMore()),
		regex.Or("",
			Leaf(1, "LABEL1", "[%g](.*)[%g]"),
			Leaf(1, "LABEL2", "(.*[%pLN_.].*)")),
		regex.End())
}

// annotationAlignment is the alignment of a header or footer: the given
// position, else the one of the style sheet.
func annotationAlignment(b *Base, align string, alignOK bool, param klimt.FontParam,
	defaultValue geom.HorizontalAlignment) geom.HorizontalAlignment {
	if !alignOK {
		return param.StyleDefinition(0).MergedStyle(b.CurrentStyleBuilder()).HorizontalAlignment()
	}
	return geom.HorizontalAlignmentFromStringDefault(align, defaultValue)
}

// CommandHeader is "[left|right|center] header text".
func CommandHeader[D Diagram]() command.Command[D] {
	return Single(headerFooterLike("header"),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			align, alignOK := arg.Get("POSITION", 0)
			ha := annotationAlignment(b, align, alignOK, klimt.FontParamHeader, geom.HorizontalAlignmentRight)
			s, ok := arg.Lazy("LABEL", 0)
			b.UpdateHeader(location, klimt.DisplayWithNewlinesString(b.Pragma(), s, ok), ha)
			return command.Ok()
		})
}

// CommandFooter is "[left|right|center] footer text".
func CommandFooter[D Diagram]() command.Command[D] {
	return Single(headerFooterLike("footer"),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			align, alignOK := arg.Get("POSITION", 0)
			ha := annotationAlignment(b, align, alignOK, klimt.FontParamFooter, geom.HorizontalAlignmentCenter)
			s, ok := arg.Lazy("LABEL", 0)
			b.UpdateFooter(location, klimt.DisplayWithNewlinesString(b.Pragma(), s, ok), ha)
			return command.Ok()
		})
}

// multilinesHeaderFooter is "[position] header" ... "end header" and the
// footer equivalent.
func multilinesHeaderFooter[D Diagram](keyword, empty string, param klimt.FontParam,
	defaultValue geom.HorizontalAlignment, update func(b *Base, location *text.LineLocation,
		display *klimt.Display, ha geom.HorizontalAlignment)) command.Command[D] {
	pattern := "^(?:(left|right|center)?[%s]*)" + keyword + "$"
	start := Compile(pattern)
	return &Multilines[D]{
		Pattern: pattern,
		Start:   start,
		End:     Compile("^end[%s]?" + keyword + "$"),
		Exec: func(d D, lines command.BlocLines, _ command.ParserPass) command.Result {
			b := d.Titled()
			location := Location(lines)
			lines = lines.Trim()
			m := start.FindStringSubmatchIndex(lines.At(0).Trimmed().S)
			if m == nil {
				Throw(IllegalStateException, "")
			}
			align, alignOK := group(lines.At(0).Trimmed().S, m, 1)
			strings := klimt.DisplayCreateFoo(SubExtract(lines, 1, 1).Lines())
			if strings.Size() > 0 {
				ha := annotationAlignment(b, align, alignOK, param, defaultValue)
				update(b, location, strings, ha)
				return command.Ok()
			}
			return command.Error(empty)
		},
	}
}

// group returns group i of a match by index.
func group(s string, m []int, i int) (string, bool) {
	if m[2*i] < 0 {
		return "", false
	}
	return s[m[2*i]:m[2*i+1]], true
}

// CommandMultilinesHeader is "[position] header" ... "end header".
func CommandMultilinesHeader[D Diagram]() command.Command[D] {
	return multilinesHeaderFooter[D]("header", "Empty header", klimt.FontParamHeader, geom.HorizontalAlignmentRight,
		(*Base).UpdateHeader)
}

// CommandMultilinesFooter is "[position] footer" ... "end footer".
func CommandMultilinesFooter[D Diagram]() command.Command[D] {
	return multilinesHeaderFooter[D]("footer", "Empty footer", klimt.FontParamFooter, geom.HorizontalAlignmentCenter,
		(*Base).UpdateFooter)
}

// CommandMainframe is "mainframe text".
func CommandMainframe[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Lit("mainframe"),
		regex.Or("",
			regex.NewConcat(regex.SpaceZeroOrMore(), Lit(":"), regex.SpaceZeroOrMore()),
			regex.SpaceOneOrMore()),
		Leaf(1, "LABEL", "(.*[%pLN_.].*)"), regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			b := d.Titled()
			s, ok := arg.Get("LABEL", 0)
			label := klimt.DisplayWithNewlinesString(b.Pragma(), s, ok)
			b.SetMainFrame(DisplayPositionedSingle(location, label, klimt.HorizontalAlignmentNull, VerticalAlignmentNull))
			return command.Ok()
		})
}

// separatorCharDouble is upstream's CommandLinkClass.SEPARATOR_CHAR_DOUBLE.
const separatorCharDouble = `[\\]{2}|::`

// CommandNamespaceSeparator is "set separator x" (or "set
// namespaceseparator x"): "none" or "null" disable namespaces.
func CommandNamespaceSeparator[D Diagram]() command.Command[D] {
	separator := `((?:none|null)|` + separatorCharDouble + `|[^%pLN%s_$#\\{}<>%g])`
	return Single(regex.NewConcat(regex.Start(),
		Lit("set"),
		regex.SpaceOneOrMore(),
		regex.Or("", Lit("separator"), Lit("namespaceseparator")),
		regex.SpaceOneOrMore(),
		Leaf(1, "SEPARATOR", separator),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			s := arg.Str("SEPARATOR", 0)
			if EqualsIgnoreCase("none", s) || EqualsIgnoreCase("null", s) {
				d.Titled().SetNamespaceSeparator("", false)
			} else {
				d.Titled().SetNamespaceSeparator(s, true)
			}
			return command.Ok()
		})
}
