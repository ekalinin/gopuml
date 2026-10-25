// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandGanttArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandGanttArrow2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandColorTask.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandSeparator.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandWeekNumberStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandGroupStart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandGroupEnd.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandLanguage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandPrintScale.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandPrintBetween.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/CommandNoteBottom.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*GanttDiagram]

// single returns a single-line command that declares the first tokens of
// upstream (nil for any line).
func single(tokens []string, exec func(d *GanttDiagram, arg *regex.Result) command.Result,
	parts ...regex.Part) cmd {
	c := uml.Single(regex.NewConcat(parts...),
		func(d *GanttDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return exec(d, arg)
		})
	c.FirstTokens = tokens
	return c
}

// arrowStyle is the optional "[style]" of an arrow.
func arrowStyle() *regex.Leaf {
	return uml.Leaf(1, "ARROW_STYLE", `(?:\[(`+descdiagram.LineStyle+`)\])?`)
}

// commandGanttArrow is "A -> B" between task ids.
func commandGanttArrow() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "CODE1", `([%pLN_.]+)`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "", `(-+)`),
		arrowStyle(),
		uml.Leaf(1, "", `(-*)`),
		uml.Lit(`\>`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "CODE2", `([%pLN_.]+)`),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *GanttDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code1, code2 := arg.Str("CODE1", 0), arg.Str("CODE2", 0)
			task1 := d.ExistingTask(code1)
			if task1 == nil {
				return command.Error("No such task " + code1)
			}
			task2 := d.ExistingTask(code2)
			if task2 == nil {
				return command.Error("No such task " + code2)
			}
			style, ok := arg.Get("ARROW_STYLE", 0)
			d.ForceTaskOrder(task1, task2).ApplyStyle(style, ok)
			return command.Ok()
		})
	return &anyLine{c}
}

// anyLine declares that a command may accept any line, as upstream's
// FirstTokens.ANYTHING, whatever its pattern.
type anyLine struct {
	*command.SingleLine[*GanttDiagram]
}

// MandatoryFirstTokens implements command.FirstTokensCommand.
func (anyLine) MandatoryFirstTokens() []string { return nil }

// commandGanttArrow2 is "[A] -> [B]", creating the tasks.
func commandGanttArrow2() cmd {
	return single([]string{"["}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		task1 := d.OrCreateTask(TaskCodeFromID(arg.Str("TASK1", 0)), false)
		task2 := d.OrCreateTask(TaskCodeFromID(arg.Str("TASK2", 0)), false)
		style, ok := arg.Get("ARROW_STYLE", 0)
		d.ForceTaskOrder(task1, task2).ApplyStyle(style, ok)
		return command.Ok()
	},
		regex.Start(),
		uml.Lit(`\[`),
		uml.Leaf(1, "TASK1", `([^\[\]]+?)`),
		uml.Lit(`\]`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "", `(-+)`),
		arrowStyle(),
		uml.Leaf(1, "", `(-*)`),
		uml.Lit(`\>`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\[`),
		uml.Leaf(1, "TASK2", `([^\[\]]+?)`),
		uml.Lit(`\]`),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandColorTask is "[T] #center/#border", with strict colors.
func commandColorTask() cmd {
	return single([]string{"["}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		code := arg.Str("CODE", 0)
		task := d.ExistingTask(code)
		if task == nil {
			return command.Error("No such task " + code)
		}
		var colors [2]color.HColor
		for i := range colors {
			if s, ok := arg.Get("COLORS", i); ok {
				c, err := d.ColorSet().GetColor(s)
				if err != nil {
					return command.BadColor()
				}
				colors[i] = c
			}
		}
		task.SetColors(NewCenterBorderColor(colors[0], colors[1], ""))
		return command.Ok()
	},
		regex.Start(),
		uml.Leaf(1, "CODE", `\[([%pLN_.]+)\]`),
		regex.SpaceOneOrMore(),
		uml.Leaf(2, "COLORS", `#(\w+)(?:/(#?\w+))?`),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandSeparator is "-- text --" or "--".
func commandSeparator() cmd {
	return single([]string{"-"}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		comment, ok := arg.Get("COMMENT", 1)
		d.AddSeparator(comment, ok)
		return command.Ok()
	},
		regex.Start(),
		uml.Lit(`--`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(2, "COMMENT", `((.+?)[%s]*--)?`),
		regex.End())
}

// commandWeekNumberStrategy is "week starts on DAY and must have at least
// N days".
func commandWeekNumberStrategy() cmd {
	return single([]string{"week", "weeks"}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		weekDay := DayOfWeekFromString(arg.Str("WEEKDAY", 0))
		d.SetWeekNumberStrategy(weekDay, parseInt(arg.Str("NUM", 0)))
		return command.Ok()
	},
		regex.Start(),
		uml.Lit(`weeks?`),
		regex.SpaceOneOrMore(),
		uml.Lit(`starts?`),
		uml.Lit(`[^0-9]*?`),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WEEKDAY", `(`+DayOfWeekRegexString()+`)`),
		regex.SpaceOneOrMore(),
		uml.Lit(`[^0-9]*?`),
		uml.Leaf(1, "NUM", `([0-9]+)`),
		uml.Lit(`[^0-9]*?`),
		regex.End())
}

// commandGroupStart is "group [name]".
func commandGroupStart() cmd {
	return single([]string{"group"}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		return d.AddGroup(TaskCodeFromID(arg.Str("NAME", 0)))
	},
		regex.Start(),
		uml.Lit(`group`),
		regex.SpaceOneOrMore(),
		uml.Lit(`\[`),
		uml.Leaf(1, "NAME", `([^\[\]]+)`),
		uml.Lit(`\]`),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandGroupEnd is "end group".
func commandGroupEnd() cmd {
	return single([]string{"end", "endgroup"}, func(d *GanttDiagram, _ *regex.Result) command.Result {
		return d.EndGroup()
	},
		regex.Start(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`end`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`group`),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandLanguage is "language xx".
func commandLanguage() cmd {
	return single([]string{"language"}, func(d *GanttDiagram, arg *regex.Result) command.Result {
		return d.ChangeLanguage(arg.Str("LANG", 0))
	},
		regex.Start(),
		uml.Lit(`language`),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "LANG", `(\w+)`),
		regex.End())
}

// commandPrintScale is "printscale weekly with week numbering from 1 zoom
// 2".
func commandPrintScale() cmd {
	return single([]string{"ganttscale", "printscale", "projectscale"},
		func(d *GanttDiagram, arg *regex.Result) command.Result {
			d.SetPrintScale(PrintScaleFromString(arg.Str("SCALE", 0)))
			if zoom, ok := arg.Get("ZOOM", 0); ok {
				d.SetFactorScale(uml.JavaParseDouble(zoom))
			}
			if option, ok := arg.Get("OPTION", 0); ok {
				if strings.Contains(option, "date") {
					d.SetWeeklyHeaderStrategy(DayOfMonth, 0)
				} else if strings.Contains(option, "numbering") {
					d.SetWeeklyHeaderStrategy(FromN, parseInt(arg.Str("NUMBER", 0)))
				}
			}
			return command.Ok()
		},
		regex.Start(),
		regex.Or("", uml.Lit(`projectscale`), uml.Lit(`ganttscale`), uml.Lit(`printscale`)),
		regex.SpaceOneOrMore(),
		regex.Or("SCALE", uml.Lit(`yearly`), uml.Lit(`quarterly`), uml.Lit(`monthly`), uml.Lit(`daily`),
			uml.Lit(`weekly`)),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(),
			regex.Or("OPTION", uml.Leaf(1, "", `(with\s+calendar\s+date)`),
				uml.Leaf(1, "NUMBER", `(?:with\s+week\s+numbering\s+from\s+(-?\d+))`)))),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Lit(`zoom`), regex.SpaceOneOrMore(),
			uml.Leaf(1, "ZOOM", `([.\d]+)`))),
		regex.End())
}

// commandPrintBetween is "print between DATE and DATE".
func commandPrintBetween() cmd {
	date := complementDate{dateAny}
	c := uml.UBrexTokens([]string{"print"}, klimt.UBrexConcat(
		klimt.UBrexLeaf("print"), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("between"),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexNamed("START", date.complementUbrex()), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexLeaf("and"), klimt.UBrexSpaceOneOrMore(), klimt.UBrexNamed("END", date.complementUbrex()),
		klimt.UBrexEnd()),
		func(d *GanttDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			m := arg.Captures().(*klimt.UBrexMatcher)
			start := date.me(d, m.ExtractByPrefix("START")).Get().(jtime.LocalDate)
			end := date.me(d, m.ExtractByPrefix("END")).Get().(jtime.LocalDate)
			d.SetPrintInterval(start, end)
			return command.Ok()
		})
	return c
}

// endNote is the end of "note bottom".
var endNote = uml.Compile(`^end[%s]*note$`)

// commandNoteBottom is "note bottom" ... "end note", the note of the last
// task.
func commandNoteBottom() cmd {
	c := &uml.Multilines2[*GanttDiagram]{
		Start: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", `(note)`),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "POSITION", `(bottom)`),
			regex.SpaceZeroOrMore(),
			uml.Leaf(4, "TAGS", klimt.StereotagPattern()+"?"),
			klimt.StereotypePatternOptional("STEREO"),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      endNote,
	}
	c.Exec = func(d *GanttDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		strings := klimt.DisplayCreateFoo(lines.Lines())
		if strings.Size() == 0 {
			return command.Error("No note defined")
		}
		var stereotype *klimt.Stereotype
		if s, ok := line0.Get("STEREO", 0); ok {
			stereotype = klimt.StereotypeBuild(s)
		}
		return d.AddNote(strings, stereotype)
	}
	return &noteFirstTokens{c}
}

// noteFirstTokens declares the first tokens of upstream's CommandNoteBottom.
type noteFirstTokens struct {
	*uml.Multilines2[*GanttDiagram]
}

// MandatoryFirstTokens implements command.FirstTokensCommand.
func (noteFirstTokens) MandatoryFirstTokens() []string { return []string{"note", "notebottom"} }
