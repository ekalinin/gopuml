// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandDelay.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandDivider.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandFootbox.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandFootboxOld.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandGrouping.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandHideUnlinked.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandHSpace.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandLinkAnchor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandReferenceOverSeveral.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandReferenceMultilinesOverSeveral.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandDelay is "..." or "... text ...".
func commandDelay() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit(`(?:\.{3}|`+"…"+`)`),
		regex.Optional(uml.Leaf(1, "LABEL", `(.*)(?:\.{3}|`+"…"+`)`)), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			strings := klimt.DisplayEmpty()
			if label, ok := arg.Get("LABEL", 0); ok {
				strings = klimt.DisplayWithNewlines(d.Pragma(), label)
			}
			d.Delay(strings)
			return command.Ok()
		})
}

// commandDivider is "== text ==".
func commandDivider() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("=="),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LABEL", "(.*)"),
		regex.SpaceZeroOrMore(),
		uml.Lit("=="), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.Divider(klimt.DisplayWithNewlines(d.Pragma(), arg.Str("LABEL", 0)))
			return command.Ok()
		})
}

// commandFootbox is "hide footbox" and "show footbox".
func commandFootbox() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", "(hide|show)?"),
		regex.SpaceZeroOrMore(),
		uml.Lit("footbox"), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.SetShowFootbox(uml.EqualsIgnoreCase("show", arg.Str("TYPE", 0)))
			return command.Ok()
		})
}

// commandFootboxOld is "footbox on|off".
func commandFootboxOld() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("footbox"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "TYPE", "(on|off)?"),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typ, ok := arg.Get("TYPE", 0)
			if !ok {
				uml.ThrowNPE(`Cannot invoke "String.equalsIgnoreCase(String)" because the return value of ` +
					`"net.sourceforge.plantuml.regex.RegexResult.get(String, int)" is null`)
			}
			d.SetShowFootbox(uml.EqualsIgnoreCase(typ, "on"))
			return command.Ok()
		})
}

// trailingBracketContent splits "label [comment]" in a "group" line.
var trailingBracketContent = regexp.MustCompile(`^(` + uml.JavaDots(`.*\[\[.*\]\].*?|.*?`) + `)\[(` +
	uml.JavaDots(`.*`) + `)\]$`)

// commandGrouping is the fragment command: "alt", "else", "end" and the
// other keywords.
func commandGrouping() cmd {
	single := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(1, "TYPE", "(opt|alt|loop|par|par2|break|critical|else|end|also|group|partition)"),
		// Upstream's first group is ((?<!else)(?<!also)(?<!end)#\w+)?: RE2
		// has no lookbehind, so Reject below refuses the matches that take
		// a color right after these keywords.
		uml.Leaf(2, "COLORS", `(#\w+)?(?:[%s]+(#\w+))?`),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "COMMENT", "(.*?)"))),
		regex.End()), executeGrouping)
	return &uml.Guarded[*SequenceDiagram]{SingleLine: single, Reject: func(arg *regex.Result) bool {
		if _, ok := arg.Get("COLORS", 0); !ok {
			return false
		}
		typ := uml.LowerCase(arg.Str("TYPE", 0))
		return typ == "else" || typ == "also" || typ == "end"
	}}
}

func executeGrouping(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
	colorSet := d.Skin().IHtmlColorSet()
	var backColorElement, backColorGeneral color.HColor
	if s, ok := arg.Get("COLORS", 0); ok {
		c, err := colorSet.GetColor(s)
		if err != nil {
			return colorResult(err)
		}
		backColorElement = c
	}
	if s, ok := arg.Get("COLORS", 1); ok {
		c, err := colorSet.GetColor(s)
		if err != nil {
			return colorResult(err)
		}
		backColorGeneral = c
	}
	typ := uml.LowerCase(arg.Str("TYPE", 0))
	comment, hasComment := arg.Get("COMMENT", 0)
	groupingType := groupingTypeOf(typ)
	if typ == "group" {
		if uml.IsEmpty(comment, hasComment) {
			comment, hasComment = "group", true
		} else if m := trailingBracketContent.FindStringSubmatch(comment); m != nil {
			typ = m[1]
			comment = m[2]
		}
	}
	_, parallel := arg.Get("PARALLEL", 0)
	if !d.Grouping(typ, comment, hasComment, groupingType, backColorGeneral, backColorElement, parallel) {
		return command.Error("Cannot create group")
	}
	return command.Ok()
}

// commandHideUnlinked is "hide|show unlinked".
func commandHideUnlinked() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "HIDE", "(hide|show)"),
		regex.SpaceOneOrMore(),
		uml.Lit("@?unlinked"),
		regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.SetHideUnlinkedData(uml.EqualsIgnoreCase(arg.Str("HIDE", 0), "hide"))
			return command.Ok()
		})
}

// commandHSpace is "|||" or "||n||".
func commandHSpace() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit(`\|\|`),
		uml.Leaf(1, "VALUE", `(\d+)?`),
		uml.Lit(`\|+`), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if size, ok := arg.Get("VALUE", 0); ok && !uml.IsEmpty(size, true) {
				d.HSpace(uml.JavaParseInt(size))
			} else {
				d.HSpace(25)
			}
			return command.Ok()
		})
}

// commandLinkAnchor is "{a} <-> {b} : text".
func commandLinkAnchor() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "ANCHOR1", `\{([%pLN_]+)\}`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(0, "LINK", `\<-\>`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "ANCHOR2", `\{([%pLN_]+)\}`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "MESSAGE", "(?::[%s]*(.*))?"), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			message, ok := arg.Get("MESSAGE", 0)
			return d.LinkAnchor(arg.Str("ANCHOR1", 0), arg.Str("ANCHOR2", 0), message, ok)
		})
}

// referenceParticipants returns the participants of a "ref over" line.
func referenceParticipants(d *SequenceDiagram, location *text.LineLocation, parts string) []*Participant {
	var p []*Participant
	for _, s := range uml.SplitComma(parts) {
		p = append(p, d.OrCreateParticipant(location, uml.RemoveQuotes(s)))
	}
	return p
}

// commandReferenceOverSeveral is "ref over A, B : text".
func commandReferenceOverSeveral() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("ref"),
		uml.Leaf(1, "REF", `(#\w+)?`),
		regex.SpaceOneOrMore(),
		uml.Lit("over"),
		regex.SpaceOneOrMore(),
		uml.Leaf(4, "PARTS", "(([%pLN_.@]+|[%g][^%g]+[%g])([%s]*,[%s]*([%pLN_.@]+|[%g][^%g]+[%g]))*)"),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, klimt.URLBuilderKey, `(\[\[.*?\]\])`)),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "TEXT", "(.*)"), regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			backColorElement, err := optionalColor(arg, "REF")
			if err != nil {
				return colorResult(err)
			}
			url, hasURL := arg.Get(klimt.URLBuilderKey, 0)
			text := text.Trin(arg.Str("TEXT", 0))
			p := referenceParticipants(d, location, arg.Str("PARTS", 0))
			strings := klimt.DisplayWithNewlines(d.Pragma(), text)
			b := urlBuilder(d)
			var u *klimt.URL
			if hasURL {
				u = b.URL(url)
			}
			d.AddReference(newReference(p, u, strings, nil, backColorElement, d.CurrentStyleBuilder()))
			return command.Ok()
		})
}

// referenceMultilinesStart is the first line of "ref over A, B" ... "end ref".
var referenceMultilinesStart = regex.NewConcat(regex.Start(),
	uml.Lit("ref"),
	uml.Leaf(1, "REF", `(#\w+)?`),
	regex.SpaceOneOrMore(),
	uml.Lit("over"),
	regex.SpaceOneOrMore(),
	uml.Leaf(1, "PARTS", "((?:[%pLN_.@]+|[%g][^%g]+[%g])(?:[%s]*,[%s]*(?:[%pLN_.@]+|[%g][^%g]+[%g]))*)"),
	regex.SpaceZeroOrMore(),
	regex.Optional(uml.Leaf(1, klimt.URLBuilderKey, `(\[\[.*?\]\])`)),
	regex.SpaceZeroOrMore(),
	uml.Leaf(1, "UNUSED", `(#\w+)?`),
	regex.End())

// commandReferenceMultilinesOverSeveral is "ref over A, B" ... "end ref".
func commandReferenceMultilinesOverSeveral() cmd {
	return &uml.Multilines[*SequenceDiagram]{
		Start: uml.Compile(referenceMultilinesStart.Pattern()),
		End:   uml.Compile("^end[%s]?(ref)?$"),
		Exec: func(d *SequenceDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
			firstLine := lines.At(0).Trimmed().S
			arg := referenceMultilinesStart.Matcher(firstLine)
			if arg == nil {
				return command.Error("Cannot parse line " + firstLine)
			}
			backColorElement, err := optionalColor(arg, "REF")
			if err != nil {
				return colorResult(err)
			}
			p := referenceParticipants(d, uml.Location(lines), arg.Str("PARTS", 0))
			lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
			strings := klimt.DisplayCreateFoo(lines.Lines())
			url, hasURL := arg.Get(klimt.URLBuilderKey, 0)
			b := urlBuilder(d)
			var u *klimt.URL
			if hasURL {
				u = b.URL(url)
			}
			d.AddReference(newReference(p, u, strings, nil, backColorElement, d.CurrentStyleBuilder()))
			return command.Ok()
		},
	}
}
