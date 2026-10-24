// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/UBrexSingleLineCommand2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

// UBrexSingle is upstream's UBrexSingleLineCommand2: a command of one line,
// trimmed, that its UBrex pattern matches exactly. The command reads the
// captures of the pattern through a regex.Result.
type UBrexSingle[D command.Diagram] struct {
	Pattern klimt.UBrexPart
	// Exec runs the command with the captures of the match.
	Exec func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result
	// Passes are the passes the command runs in; One by default.
	Passes []command.ParserPass
	// Tokens are the first tokens the command declares, as upstream writes
	// them by hand for a UBrex pattern; nil for any line.
	Tokens []string
}

// UBrex returns a UBrex command of one line.
func UBrex[D command.Diagram](pattern klimt.UBrexPart,
	exec func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result) *UBrexSingle[D] {
	return &UBrexSingle[D]{Pattern: pattern, Exec: exec}
}

// UBrexTokens returns a UBrex command of one line that declares the first
// tokens of the lines it accepts.
func UBrexTokens[D command.Diagram](tokens []string, pattern klimt.UBrexPart,
	exec func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result) *UBrexSingle[D] {
	c := UBrex(pattern, exec)
	c.Tokens = tokens
	return c
}

// MandatoryFirstTokens implements command.FirstTokensCommand.
func (c *UBrexSingle[D]) MandatoryFirstTokens() []string { return c.Tokens }

// IsValid implements command.Command.
func (c *UBrexSingle[D]) IsValid(lines command.BlocLines) command.Control {
	if lines.Size() != 1 {
		return command.NotOK
	}
	if c.Pattern.Match(lines.At(0).Trimmed().S).ExactMatch() {
		return command.OK
	}
	return command.NotOK
}

// Execute implements command.Command.
func (c *UBrexSingle[D]) Execute(d D, lines command.BlocLines, pass command.ParserPass) command.Result {
	if lines.Size() != 1 {
		Throw(IllegalArgumentException, "")
	}
	first := lines.At(0)
	line := first.Trimmed().S
	matcher := c.Pattern.Match(line)
	if !matcher.ExactMatch() {
		return command.Error("Cannot parse line " + line)
	}
	return c.Exec(d, first.Location, regex.NewResultFromCaptures(matcher), pass)
}

// EligibleFor implements command.Command.
func (c *UBrexSingle[D]) EligibleFor(pass command.ParserPass) bool {
	if len(c.Passes) == 0 {
		return pass == command.One
	}
	return slices.Contains(c.Passes, pass)
}
