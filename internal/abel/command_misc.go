// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandFootboxIgnored.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandRankDir.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandNewpage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandHideShow2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandRemoveRestore.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandUrl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/CommandConstraintOnLinks.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// CommandFootboxIgnored is "[hide|show] footbox", accepted and ignored by
// the diagrams without footbox.
func CommandFootboxIgnored[D uml.Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.NewNamedLeaf(1, "", "(hide|show)?"),
		regex.SpaceZeroOrMore(),
		uml.Lit("footbox"), regex.End()),
		func(D, *text.LineLocation, *regex.Result, command.ParserPass) command.Result { return command.Ok() })
}

// CommandRankDir is "left to right direction" and "top to bottom
// direction".
func CommandRankDir[D uml.Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "DIRECTION", "(left[%s]to[%s]right|top[%s]to[%s]bottom)"),
		regex.SpaceOneOrMore(),
		uml.Lit("direction"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			s := strings.ReplaceAll(uml.UpperCase(arg.Str("DIRECTION", 0)), " ", "_")
			rankdir, ok := geom.RankdirValueOf(s)
			if !ok {
				uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.klimt.geom.Rankdir."+s)
			}
			d.Titled().Skin().SetRankdir(rankdir)
			return command.Ok()
		})
}

// CommandNewpage is "newpage": the following lines form a new empty
// diagram, made by newEmpty from the diagram (only the dpi setting is
// carried over). The result is a NewpagedDiagram.
func CommandNewpage[D uml.Diagram](newEmpty func(d D) D) command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(), uml.Lit("newpage"), regex.End()),
		func(d D, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			dpi := d.Titled().Skin().Dpi()
			emptyDiagram := newEmpty(d)
			if dpi != 96 {
				emptyDiagram.Titled().SetParam("dpi", strconv.Itoa(dpi))
			}
			return command.NewDiagramResult(NewNewpagedDiagram(d.Titled().Source(), d, emptyDiagram))
		})
}

// CommandHideShow2 is "hide|show [-class] what": what is a name pattern,
// a "<<stereotype>>", a "$tag" or "@unlinked".
func CommandHideShow2[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "COMMAND", "(hide|hide-class|show|show-class)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WHAT", `([^%s]+|\<\<.*\>\>)`),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			tmp := arg.Str("COMMAND", 0)[0]
			show := tmp == 's' || tmp == 'S'
			what := text.Trin(arg.Str("WHAT", 0))
			d.Cuca().HideOrShow2(what, show)
			return command.Ok()
		})
}

// CommandRemoveRestore is "remove|restore what".
func CommandRemoveRestore[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "COMMAND", "(remove|restore)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WHAT", "(.+)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			show := uml.EqualsIgnoreCase(arg.Str("COMMAND", 0), "restore")
			what := text.Trin(arg.Str("WHAT", 0))
			d.Cuca().RemoveOrRestore(what, show)
			return command.Ok()
		})
}

// CommandURL is "url [of|for] code [is] [[link]]".
func CommandURL[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("url"),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Lit("of|for")),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.]+|[%g][^%g]+[%g])"),
		regex.SpaceOneOrMore(),
		regex.Optional(uml.Lit("is")),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderMandatory(),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			quark := c.QuarkInContext(true, c.CleanID(arg.Str("CODE", 0)))
			entity := quark.Data()
			if entity == nil {
				return command.Error(quark.Name() + " does not exist")
			}
			entity.AddURL(URLBuilderOf(c).URL(arg.Str(klimt.URLBuilderKey, 0)))
			return command.Ok()
		})
}

// CommandConstraintOnLinks is "constraint on links : text", drawn between
// the two last links.
func CommandConstraintOnLinks[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("constraint"),
		regex.SpaceZeroOrMore(),
		uml.Lit("on"),
		regex.SpaceOneOrMore(),
		uml.Lit("links"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			links := c.TwoLastLinks()
			if links == nil {
				return command.Error("Cannot put constraint on two last links")
			}
			var note []text.StringLocated
			for _, cs := range klimt.DisplayWithNewlines3(arg.Str("NOTE", 0), true) {
				note = append(note, text.NewStringLocated(cs, nil))
			}
			return c.ConstraintOnLinks(links[0], links[1], klimt.DisplayCreateFoo(note))
		})
}
