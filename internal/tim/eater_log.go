// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterDumpMemory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterLog.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterDumpMemory is "!dump_memory [message]": it writes the variables to
// the diagnostic log.
type EaterDumpMemory struct{ Eater }

func newEaterDumpMemory(sl text.StringLocated) *EaterDumpMemory {
	return &EaterDumpMemory{newEater(sl)}
}

func (e *EaterDumpMemory) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!dump_memory"); err != nil {
		return err
	}
	e.skipSpaces()
	remain := e.eatAllToEnd()
	memory.dumpDebug(remain, context.logInfo)
	return nil
}

// EaterLog is "!log message": the message, with variables and functions
// applied, goes to the user log (stderr), not to the diagram.
type EaterLog struct{ Eater }

func newEaterLog(sl text.StringLocated) *EaterLog { return &EaterLog{newEater(sl)} }

func (e *EaterLog) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!log"); err != nil {
		return err
	}
	e.skipSpaces()
	logData, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(e.eatAllToEnd(), e.getLineLocation()))
	if err != nil {
		return err
	}
	data := "null"
	if logData != nil {
		data = *logData
	}
	context.userLog = append(context.userLog, "[Log] "+data)
	return nil
}
