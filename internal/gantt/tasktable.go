// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttTaskTableColumn.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

// GanttTaskTableColumn is upstream's GanttTaskTableColumn: a column of the
// table of tasks.
type GanttTaskTableColumn int

// The columns of the table of tasks, in upstream's order.
const (
	ColumnTask GanttTaskTableColumn = iota
	ColumnStart
	ColumnEnd
	ColumnDuration
)
