// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/ChenEerDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*ChenEerDiagram]

// Factory is the parser of Chen EER diagrams, upstream's
// ChenEerDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.ChenEER, commands, NewChenEerDiagram)

// commands returns the commands of Chen EER diagrams in upstream's order.
func commands() []cmd {
	cmds := uml.AddCommonCommands1([]cmd{})
	return append(cmds,
		abel.CommandRankDir[*ChenEerDiagram](),

		commandCreateEntity(),
		commandCreateAttribute(),
		commandAssociate(),
		commandEndGroup(),
		commandSimpleSubclass(),
		commandMultiSubclass(),
	)
}
