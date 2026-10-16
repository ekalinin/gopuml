// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram/command/CommandLinkActivity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandCreateElementFull.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandArchimate.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/command/CommandCreateState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/command/CommandLinkState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandRobustConcise.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandClock.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandBinary.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandAnalog.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/timingdiagram/command/CommandAtTime.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"regexp"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// unsupportedType recognizes an @startuml diagram type this implementation
// cannot parse yet, from the leading part of its characteristic commands.
// Keywords that sequence, class or activity diagrams also accept are left
// out, so that their syntax errors are not reported as another type.
type unsupportedType struct {
	name     string
	patterns []*regexp.Regexp
}

func patterns(p ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(p))
	for i, s := range p {
		out[i] = regex.MustCompile(s)
	}
	return out
}

// unsupportedTypes are tried in upstream's factory order.
var unsupportedTypes = []unsupportedType{
	{"legacy activity", patterns(
		`^\(\*(top)?\)`,                     // (*) --> First
		`-+>[%s]*(\[[^\]]*\])?[%s]*\(\*\)$`, // ... --> (*)
	)},
	{"description (use case, component, deployment, archimate)", patterns(
		`^(usecase|component|artifact|agent|card|file|hexagon|label|node|person|process|rectangle|stack|storage|action|portin|portout)[%s]+[^{]*$`,
		`^archimate[%s]`,
		`^\([%pLN_][^()]*\)`,                    // (Use case) --> ...
		`-+>[%s]*\([%pLN_][^()]*\)[%s]*(:.*)?$`, // ... --> (Use case)
		`^:[^:]+:[^;]*$`,                        // :Actor: --> ... (not an activity ":Step 1: text;")
		`^\[[^\]*]+\]`,                          // [Component] --> ...
	)},
	{"state", patterns(
		`^state[%s]+`,
		`^\[(\*|H\*?)\]`,                    // [*] --> Idle
		`-+>[%s]*\[(\*|H\*?)\][%s]*(:.*)?$`, // Idle --> [*]
	)},
	{"timing", patterns(
		`^(compact[%s]+)?(robust|concise)[%s]+`,
		`^(compact[%s]+)?clock[%s]+.*[%s]with[%s]+period[%s]`,
		`^(compact[%s]+)?(binary|analog)[%s]+`,
		`^@[+-]?[0-9]`,
	)},
}

// detectUnsupportedUmlType returns a "not supported yet" error when an
// @startuml block uses the characteristic syntax of a diagram type this
// implementation does not parse yet, or nil.
func detectUnsupportedUmlType(src *UmlSource) Diagram {
	lines := src.Lines()
	for _, t := range unsupportedTypes {
		for i, l := range lines {
			s := text.Trin(l.S)
			for _, p := range t.patterns {
				if !p.MatchString(s) {
					continue
				}
				line := l
				return &ErrorDiagram{
					Source: lines,
					Trace:  lines[:i+1],
					Error:  ErrorUml{Type: ExecutionError, Error: t.name + " diagrams are not supported yet", Line: &line},

					NotSupportedYet: true,
				}
			}
		}
	}
	return nil
}

// descriptionElement matches a whole description diagram element or
// container line that activity v3 also reads as a partition, for example
// "rectangle Lane {" or `card "A" as a`: keyword, name, optional alias,
// stereotype, link, color and opening brace. Other text after it, which
// the description factory rejects, does not match.
var descriptionElement = regexp.MustCompile(`^(rectangle|card|package|folder|frame|cloud|node|database|storage|component|artifact|agent|file|hexagon|label|person|process|stack|usecase|actor|queue|collections)` +
	`[\t\n\x0B\f\r ]+("[^"]*"|[^\t\n\x0B\f\r "{}]+)([\t\n\x0B\f\r ]+as[\t\n\x0B\f\r ]+[^\t\n\x0B\f\r {}]+)?` +
	`([\t\n\x0B\f\r ]*<<[^>]*>>)?([\t\n\x0B\f\r ]*\[\[[^\]]*\]\])?([\t\n\x0B\f\r ]*#[\w#;:.-]+)?[\t\n\x0B\f\r ]*\{?$`)

// commonLine matches the lines every diagram type accepts.
var commonLine = regexp.MustCompile(`^(skinparam|title|caption|header|footer|scale|hide|show|left to right direction|top to bottom direction)\b`)

// detectDescriptionOnly returns the "not supported yet" error of a block
// whose content lines are only description elements, containers and their
// closing braces (with common commands): upstream's description factory,
// which comes before activity v3, accepts such a block. It returns nil
// otherwise.
func detectDescriptionOnly(src *UmlSource) Diagram {
	lines := src.Lines()
	if len(lines) < 3 {
		return nil
	}
	first := -1
	for i := 1; i < len(lines)-1; i++ {
		s := text.Trin(lines[i].S)
		switch {
		case s == "" || s == "}" || commonLine.MatchString(s):
		case descriptionElement.MatchString(s):
			if first == -1 {
				first = i
			}
		default:
			return nil
		}
	}
	if first == -1 {
		return nil
	}
	line := lines[first]
	return &ErrorDiagram{
		Source: lines,
		Trace:  lines[:first+1],
		Error: ErrorUml{Type: ExecutionError, Error: "description (use case, component, deployment, archimate) diagrams are not supported yet",
			Line: &line},
		NotSupportedYet: true,
	}
}
