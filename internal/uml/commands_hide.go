// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandHideEmptyDescription.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandHideShowByVisibility.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandHideShowByGender.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
)

// AddCommonHides adds the "hide" commands that every diagram accepts.
func AddCommonHides[D Diagram](cmds []command.Command[D]) []command.Command[D] {
	return append(cmds,
		CommandHideEmptyDescription[D](),
		CommandHideShowByVisibility[D](),
		CommandHideShowByGender[D](),
	)
}

// CommandHideEmptyDescription is "hide|show empty description"; only the
// diagrams implementing HideEmptyDescriptionSetter use it.
func CommandHideEmptyDescription[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Leaf(1, "HIDE", "(hide|show)"),
		regex.SpaceOneOrMore(),
		Lit("empty"),
		regex.SpaceOneOrMore(),
		Lit("description"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if s, ok := any(d).(HideEmptyDescriptionSetter); ok {
				s.SetHideEmptyDescription(EqualsIgnoreCase(arg.Str("HIDE", 0), "hide"))
			}
			return command.Ok()
		})
}

// VisibilityHider is implemented by the class diagram, the only diagram
// that honors "hide private members" and the like (upstream's
// CommandHideShowByVisibility ignores the line elsewhere).
type VisibilityHider interface {
	HideShowByVisibility(arg *regex.Result) command.Result
}

// CommandHideShowByVisibility is "hide|show [visibilities] members". The
// groups of the match are COMMAND, VISIBILITY and PORTION.
func CommandHideShowByVisibility[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Leaf(1, "COMMAND", "(hide|show)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "VISIBILITY",
			"((?:public|private|protected|package)?(?:[,%s]+(?:public|private|protected|package))*)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "PORTION", "(members?|attributes?|fields?|methods?)"), regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if h, ok := any(d).(VisibilityHider); ok {
				return h.HideShowByVisibility(arg)
			}
			return command.Ok()
		})
}

// GenderHider is implemented by the diagrams that honor "hide|show
// [gender] [empty] portion" (upstream's CommandHideShowByGender ignores the
// line elsewhere). The groups of the match are COMMAND, GENDER, EMPTY and
// PORTION.
type GenderHider interface {
	HideShowByGender(arg *regex.Result) command.Result
}

// CommandHideShowByGender is "hide|show [gender] [empty] portion".
func CommandHideShowByGender[D Diagram]() command.Command[D] {
	return Single(regex.NewConcat(regex.Start(),
		Leaf(1, "COMMAND", "(hide|show)"),
		regex.SpaceOneOrMore(),
		Leaf(1, "GENDER",
			`(?:(class|object|interface|enum|annotation|dataclass|record|abstract|[%pLN_.]+|[%g][^%g]+[%g]|\<\<.*\>\>)[%s]+)*?`),
		regex.Optional(regex.NewConcat(Leaf(1, "EMPTY", "(empty)"), regex.SpaceOneOrMore())),
		Leaf(1, "PORTION", "(members?|attributes?|fields?|methods?|circles?|circled?|stereotypes?)"),
		regex.End()),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if h, ok := any(d).(GenderHider); ok {
				return h.HideShowByGender(arg)
			}
			return command.Ok()
		})
}

// GenderPortion is CommandHideShowByGender.getEntityPortion: the portion
// named by the first three letters of s.
func GenderPortion(s string) EntityPortion {
	if JavaLength(s) < 3 {
		Throw(StringIndexOutOfBoundsException, "")
	}
	switch LowerCase(s[:3]) {
	case "met":
		return PortionMethod
	case "mem":
		return PortionMember
	case "att", "fie":
		return PortionField
	case "cir":
		return PortionCircledCharacter
	case "ste":
		return PortionStereotype
	}
	Throw(IllegalArgumentException, "")
	return 0
}
