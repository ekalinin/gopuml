// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/StateDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of state diagrams, upstream's StateDiagramFactory.
// It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.State, commands, NewStateDiagram)

// noteOnEntityCode is the "state" code of upstream's
// CommandFactoryNoteOnEntity for state diagrams.
func noteOnEntityCode() regex.Part {
	return regex.Or("CODE", uml.Lit(`[%pLN_.]+`), uml.Lit(`[%g][^%g]+[%g]`))
}

// commands returns the commands of state diagrams in upstream's order. The
// notes on states run in the third pass and the notes on transitions in
// the second, once the states and transitions exist.
func commands() []cmd {
	code := noteOnEntityCode()
	cmds := []cmd{
		abel.CommandFootboxIgnored[*StateDiagram](),
		abel.CommandRankDir[*StateDiagram](),
		abel.CommandRemoveRestore[*StateDiagram](),
		commandCreateState(),
		commandLinkState(),
		commandLinkStateReverse(),
		commandCreatePackageState(),
		commandCreatePackage2(),
		commandEndState(),
		commandAddField(),
		commandConcurrentState(),

		command.InPasses(abel.CommandNoteOnEntityMulti[*StateDiagram](code, true), command.Three),
		command.InPasses(abel.CommandNoteOnEntityMulti[*StateDiagram](code, false), command.Three),
		command.InPasses(abel.CommandNoteOnEntitySingle[*StateDiagram](code), command.Three),

		command.InPasses(abel.CommandNoteOnLinkSingle[*StateDiagram](), command.Two),
		command.InPasses(abel.CommandNoteOnLinkMulti[*StateDiagram](), command.Two),
		abel.CommandURL[*StateDiagram](),

		abel.CommandNoteSingle[*StateDiagram](),
		abel.CommandNoteMulti[*StateDiagram](),

		abel.CommandCreateMap[*StateDiagram](),
		abel.CommandCreateJSON[*StateDiagram](),
		abel.CommandCreateJSONSingleLine[*StateDiagram](),
	}
	cmds = uml.AddCommonCommands1(cmds)
	return append(cmds, abel.CommandHideShow2[*StateDiagram]())
}
