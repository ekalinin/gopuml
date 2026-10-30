// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/PSystemSaltFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/CommandAnything.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*PSystemSalt]

// Factory is the parser of salt diagrams, upstream's PSystemSaltFactory. It
// is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Salt, commands, NewPSystemSalt)

// commands returns the commands of salt diagrams in upstream's order: the
// common commands, then CommandAnything, which keeps every other line for
// the drawing. A line that a common command matches is that command, even
// inside a group.
func commands() []cmd {
	cmds := uml.AddCommonCommands2([]cmd(nil))
	cmds = uml.AddCommonScaleCommands(cmds)
	cmds = uml.AddTitleCommands(cmds)
	return append(cmds, commandAnything())
}

// commandAnything keeps any line, untrimmed. A first line "salt" fails, as
// the keyword is not needed anymore.
func commandAnything() cmd {
	return &command.SingleLine[*PSystemSalt]{
		Pattern: regex.NewConcat(regex.Start(), uml.Leaf(1, "ALL", "(.*)"), regex.End()),
		NoTrim:  true,
		Exec: func(d *PSystemSalt, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			s := arg.Str("ALL", 0)
			if d.size() == 0 && text.JavaTrim(s) == "salt" {
				return command.Error("This is not needed anymore.")
			}
			d.add(s)
			return command.Ok()
		},
	}
}
