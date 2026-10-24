// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/ActivityDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of legacy activity diagrams, upstream's
// ActivityDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Activity, commands, NewActivityDiagram)

// commands returns the commands of legacy activity diagrams in upstream's
// order. Upstream does not register CommandInnerConcurrent.
func commands() []cmd {
	cmds := []cmd{abel.UBrexCommandFootboxIgnored[*ActivityDiagram]()}
	cmds = uml.AddCommonCommands1(cmds)
	return append(cmds,
		abel.UBrexCommandRankDir[*ActivityDiagram](),

		commandPartition(),
		commandEndPartition(),
		commandLinkLongActivity(),

		commandNoteActivitySingle(),
		commandNoteActivityMulti(),

		abel.CommandNoteOnLinkSingle[*ActivityDiagram](),
		abel.CommandNoteOnLinkMulti[*ActivityDiagram](),

		commandIf(),
		commandElse(),
		commandEndif(),

		commandLinkActivity(),
		abel.UBrexCommandHideShow2[*ActivityDiagram](),
	)
}
