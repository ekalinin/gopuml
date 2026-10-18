// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/DescriptionDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of description diagrams, upstream's
// DescriptionDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Description, commands, NewDescriptionDiagram)

// newEmptyPage is the empty diagram of a "newpage", upstream's
// createEmptyDiagram with the source and images of the diagram.
func newEmptyPage(d *DescriptionDiagram) *DescriptionDiagram {
	return NewDescriptionDiagram(d.RenderContext(), d.Source(), d.MD5Map())
}

// noteOnEntityCode is the "desc" code of upstream's CommandFactoryNoteOnEntity
// for description diagrams. Upstream's negative lookahead, which keeps
// "(*)" for state diagrams, is rewritten for RE2 on the content of the
// parentheses.
func noteOnEntityCode() *regex.Leaf {
	return uml.Leaf(1, "CODE", "("+strings.Join([]string{
		`[%pLN_.]+`,
		`\(\)[%s]*[%pLN_.]+`,
		`\(\)[%s]*[%g][^%g]+[%g]`,
		`\[[^\]*]+[^\]]*\]`,
		`\((?:[^)*][^)]*|\*[^)]+)\)`,
		`:[^:]+:`,
		`[%g][^%g]+[%g]`,
	}, "|")+")")
}

// commands returns the commands of description diagrams in upstream's
// order, without the archimate commands, which are not supported yet.
func commands() []cmd {
	cmds := []cmd{
		abel.CommandFootboxIgnored[*DescriptionDiagram](),
		abel.CommandRankDir[*DescriptionDiagram](),
		abel.CommandNewpage(newEmptyPage),
	}
	cmds = uml.AddCommonCommands1(cmds)
	code := noteOnEntityCode()
	return append(cmds,
		commandLinkElement(),
		abel.CommandHideShow2[*DescriptionDiagram](),
		abel.CommandRemoveRestore[*DescriptionDiagram](),

		abel.CommandPackageWithUSymbol[*DescriptionDiagram](),
		abel.CommandTogether[*DescriptionDiagram](),

		abel.CommandEndPackage[*DescriptionDiagram](),
		abel.CommandNoteMulti[*DescriptionDiagram](),

		abel.CommandNoteOnLinkSingle[*DescriptionDiagram](),
		abel.CommandNoteOnLinkMulti[*DescriptionDiagram](),

		abel.CommandNoteOnEntitySingle[*DescriptionDiagram](code),

		abel.CommandNoteSingle[*DescriptionDiagram](),
		abel.CommandURL[*DescriptionDiagram](),
		commandCreateElementFull(),
		abel.CommandCreateElementMultilines[*DescriptionDiagram](0),
		abel.CommandCreateElementMultilines[*DescriptionDiagram](1),

		abel.CommandNoteOnEntityMulti[*DescriptionDiagram](code, true),
		abel.CommandNoteOnEntityMulti[*DescriptionDiagram](code, false),
		abel.CommandNoteMulti[*DescriptionDiagram](),

		abel.CommandCreateMap[*DescriptionDiagram](),
		abel.CommandCreateJSON[*DescriptionDiagram](),
		abel.CommandCreateJSONSingleLine[*DescriptionDiagram](),

		commandCreateDomain(),
	)
}
