// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ActivityDiagramFactory3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of activity diagrams (new syntax), upstream's
// ActivityDiagramFactory3. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Activity, commands, NewActivityDiagram3)

// commands returns the commands of activity diagrams in upstream's order.
func commands() []cmd {
	cmds := []cmd{commandFootboxIgnored()}
	cmds = uml.AddCommonCommands1(cmds)
	return append(cmds,
		commandSwimlane(),
		commandSwimlane2(),
		commandPartition3(),
		commandCloseGroup3(),
		commandCloseGroupLegacy3(),
		commandArrow3(),
		commandArrowLong3(),
		commandRepeat3(),
		commandActivity3(),
		commandIf4(),
		commandIf2(),
		newDecoratorMultine(commandIf2(), 50),
		commandIfLegacy1(),
		commandElseIf3(),
		commandElseIf2(),
		commandElse3(),
		newDecoratorMultine(commandElse3(), 50),
		commandElseLegacy1(),
		commandEndif3(),

		commandSwitch(),
		commandCase(),
		commandEndSwitch(),

		commandRepeatWhile3(),
		commandRepeatWhile3Multilines(),
		commandBackward3(),
		commandBackwardLong3(),
		commandWhile3(),
		commandWhileEnd3(),

		commandFork3(),
		commandForkAgain3(),
		commandForkEnd3(),

		commandSplit3(),
		commandSplitAgain3(),
		commandSplitEnd3(),
		commandStart3(),
		commandStop3(),
		commandCircleSpot3(),
		commandBreak(),
		commandEnd3(),
		commandKill3(),
		commandLink3(),
		commandNote3(),
		commandNoteLong3(),

		commandActivityLong3(),
		commandActivityList(),

		commandLabel(),
		commandGoto(),
		newDecoratorMultine(commandElseIf2(), 50),
	)
}
