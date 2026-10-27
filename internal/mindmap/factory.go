// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/MindMapDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of mindmap diagrams, upstream's
// MindMapDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.MindMap, commands, NewMindMapDiagram)

// commands returns the commands of mindmap diagrams in upstream's order:
// the common commands, the rank direction, then the nodes and the side.
func commands() []cmd {
	cmds := uml.AddCommonCommands1([]cmd(nil))
	return append(cmds,
		abel.CommandRankDir[*MindMapDiagram](),
		commandMindMapOrgmodeMultiline(),
		commandMindMapOrgmode(),
		commandMindMapRoot(),
		commandMindMapPlus(),
		commandMindMapDirection(),
	)
}
