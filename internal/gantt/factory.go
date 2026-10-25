// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory is the parser of gantt diagrams, upstream's GanttDiagramFactory.
// It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Gantt, commands, NewGanttDiagram)

// commands returns the commands of gantt diagrams in upstream's order: the
// title, common and scale commands, the sentences, then the other gantt
// commands.
func commands() []cmd {
	cmds := uml.AddTitleCommands([]cmd(nil))
	cmds = uml.AddCommonCommands2(cmds)
	cmds = uml.AddCommonScaleCommands(cmds)
	for _, s := range []subject{subjectGantt{}, subjectToday{}, subjectTask{}, subjectResource{},
		subjectDaysAsDates{}, subjectDayOfWeek{}, subjectDayAsDate{}, subjectSeparator{}, subjectWorkingHours{}} {
		cmds = append(cmds, newNaturalGanttCommand(s))
	}
	return append(cmds,
		commandGanttArrow(),
		commandGanttArrow2(),
		commandColorTask(),
		commandSeparator(),
		commandWeekNumberStrategy(),
		commandGroupStart(),
		commandGroupEnd(),
		commandLanguage(),
		commandPrintScale(),
		commandPrintBetween(),
		commandNoteBottom(),
		commandFootboxGantt(),
		commandLabelOnColumn(),
		commandHideResourceName(),
		commandHideResourceFootbox(),
		commandHideShowColumns(),
		commandHideClosed(),
		commandTaskCompleteDefault(),
	)
}
