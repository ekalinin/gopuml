// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/packetdiag/PacketDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of packet diagrams, upstream's PacketDiagramFactory.
// It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Packet, commands, NewPacketDiagram)

// commands returns the commands of packet diagrams in upstream's order: the
// common commands, then the packetdiag directives and fields.
func commands() []cmd {
	cmds := uml.AddCommonCommands1([]cmd(nil))
	return append(cmds,
		commandPacketDiagStart(),
		commandColWidth(),
		commandNodeHeight(),
		commandScaleDirection(),
		commandScaleInterval(),
		commandSameHeight(),
		commandNumRange(),
		commandPacketDiagEnd(),
	)
}
