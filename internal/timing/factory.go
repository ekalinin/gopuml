// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of timing diagrams, upstream's
// TimingDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Timing, commands, NewTimingDiagram)

// commands returns the commands of timing diagrams in upstream's order.
func commands() []cmd {
	cmds := uml.AddCommonCommands1([]cmd(nil))
	return append(cmds,
		abel.CommandFootboxIgnored[*TimingDiagram](),
		commandRobustConcise(),
		commandClock(),
		commandAnalog(),
		commandBinary(),
		commandDefineStateShort(),
		commandDefineStateLong(),
		commandChangeStateByPlayerCode(),
		commandChangeStateByTime(),
		commandAtTime(),
		commandAtPlayer(),
		commandTimeMessage(),
		commandNote(),
		commandNoteLong(),
		commandConstraint(),
		commandScalePixel(),
		commandHideTimeAxis(),
		commandHighlight(),
		commandModeCompact(),
		commandTicks(),
		commandPixelHeight(),
		commandUseDateFormat(),
	)
}
