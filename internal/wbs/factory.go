// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of WBS diagrams, upstream's WBSDiagramFactory. It
// is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.WBS, commands, NewWBSDiagram)

// commands returns the commands of WBS diagrams in upstream's order: the
// common commands, the nodes in the expected order, the nodes in the legacy
// order, then the links.
func commands() []cmd {
	cmds := uml.AddCommonCommands1([]cmd(nil))
	return append(cmds,
		commandWBSItemNew(1),
		commandWBSItemNew(0),
		commandWBSItemMultiline(false),
		commandWBSItemOld(1),
		commandWBSItemOld(0),
		commandWBSItemMultiline(true),
		commandWBSLink(),
	)
}
