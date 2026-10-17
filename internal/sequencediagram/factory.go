// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/SequenceDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of sequence diagrams, upstream's
// SequenceDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Sequence, commands, NewSequenceDiagram)

// commands returns the commands of sequence diagrams in upstream's order.
func commands() []cmd {
	cmds := uml.AddCommonCommands1[*SequenceDiagram](nil)
	cmds = append(cmds,
		commandHideUnlinked(),

		commandActivate(),
		commandDeactivateShort(),

		commandParticipantA(),
		commandParticipantA2(),
		commandParticipantA3(),
		commandParticipantA4(),
		commandParticipantMultilines(),
		commandArrow(),
		commandExoArrowLeft(),
		commandExoArrowRight(),

		commandNoteSingle(),
		commandNoteOverSeveralSingle(),
		commandNoteAcrossSingle(),

		commandBoxStart(),
		commandBoxEnd(),
		commandGrouping(),

		commandActivate2(),
		commandReturn(),

		commandNoteOnArrowSingle(),

		commandNoteMulti(),
		commandNoteOverSeveralMulti(),
		commandNoteOnArrowMulti(),
		commandNoteAcrossMulti(),

		commandNewpage(),
		commandIgnoreNewpage(),
		commandAutoNewpage(),
		commandDivider(),
		commandHSpace(),
		commandReferenceOverSeveral(),
		commandReferenceMultilinesOverSeveral(),
		commandAutonumber(),
		commandAutonumberStop(),
		commandAutonumberResume(),
		commandAutonumberIncrement(),
		commandAutoactivate(),
		commandFootbox(),
		commandDelay(),
		commandFootboxOld(),
		commandURL(),
		commandLinkAnchor(),
	)
	return cmds
}
