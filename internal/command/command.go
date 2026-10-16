// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/Command.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandControl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/ParserPass.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandExecutionResult.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/SingleLineCommand2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/AbstractDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Control is the answer of a command to some lines.
type Control int

// Controls.
const (
	OK        Control = iota // the lines form a complete command
	NotOK                    // the lines are not this command
	OKPartial                // the lines start a multi-line command
)

// ParserPass is one pass over the source; most diagrams need only One.
type ParserPass int

// Parser passes.
const (
	One ParserPass = iota
	Two
	Three
)

// Diagram is a diagram that commands build.
type Diagram interface {
	core.Diagram
	// RequiredPasses lists the passes, in order.
	RequiredPasses() []ParserPass
	StartingPass(ParserPass)
	// CheckFinalError returns an error message once all lines are read, or "".
	CheckFinalError() string
	MakeDiagramReady()
	IsIncomplete() bool
}

// Base provides upstream's default diagram hooks; diagrams embed it.
type Base struct{}

func (Base) RequiredPasses() []ParserPass { return []ParserPass{One} }
func (Base) StartingPass(ParserPass)      {}
func (Base) CheckFinalError() string      { return "" }
func (Base) MakeDiagramReady()            {}
func (Base) IsIncomplete() bool           { return false }

// Result is the outcome of executing a command.
type Result struct {
	err   string
	score int
	// NewDiagram, when set, replaces the diagram being built.
	NewDiagram Diagram
}

// Ok is a successful result.
func Ok() Result { return Result{} }

// NewDiagramResult replaces the diagram being built.
func NewDiagramResult(d Diagram) Result { return Result{NewDiagram: d} }

// Error is a failed result with score 0.
func Error(msg string) Result { return Result{err: msg} }

// ErrorScore is a failed result with a score.
func ErrorScore(msg string, score int) Result { return Result{err: msg, score: score} }

// BadColor is upstream's result for an unknown color.
func BadColor() Result { return Result{err: "No such color", score: 10} }

// IsOK reports whether the command succeeded.
func (r Result) IsOK() bool { return r.err == "" }

// Err is the error message of a failed result.
func (r Result) Err() string { return r.err }

// Score ranks the error against errors of other diagram types.
func (r Result) Score() int { return r.score }

// Command recognizes and executes lines for diagrams of type D.
type Command[D Diagram] interface {
	IsValid(lines BlocLines) Control
	Execute(d D, lines BlocLines, pass ParserPass) Result
	EligibleFor(pass ParserPass) bool
}

// SingleLine is a command made of one line matching a pattern.
type SingleLine[D Diagram] struct {
	Pattern *regex.Concat
	// Exec runs the command with the named groups of the match.
	Exec func(d D, location *text.LineLocation, arg *regex.Result, pass ParserPass) Result
	// NoTrim keeps the spaces around the line; lines are trimmed by default.
	NoTrim bool
	// FinalBracket accepts "line {" also written as "line" then "{".
	FinalBracket bool
	// Forbidden, when set, rejects lines the pattern would accept.
	Forbidden func(line string) bool
	// Passes are the passes the command runs in; One by default.
	Passes []ParserPass
}

func (c *SingleLine[D]) trim(l text.StringLocated) text.StringLocated {
	if c.NoTrim {
		return l
	}
	return l.Trimmed()
}

// IsValid implements Command.
func (c *SingleLine[D]) IsValid(lines BlocLines) Control {
	if lines.Size() == 2 && c.FinalBracket {
		return c.isValidBracket(lines)
	}
	if lines.Size() != 1 {
		return NotOK
	}
	line := c.trim(lines.At(0))
	if c.FinalBracket && !strings.HasSuffix(line.S, "{") {
		if c.IsValid(SingleString(lines.At(0).S+" {")) == OK {
			return OKPartial
		}
		return NotOK
	}
	if c.Pattern.Match(line.S) {
		return OK
	}
	return NotOK
}

func (c *SingleLine[D]) isValidBracket(lines BlocLines) Control {
	if c.trim(lines.At(1)).S != "{" {
		return NotOK
	}
	return c.IsValid(SingleString(lines.At(0).S + " {"))
}

// Execute implements Command.
func (c *SingleLine[D]) Execute(d D, lines BlocLines, pass ParserPass) Result {
	if c.FinalBracket && lines.Size() == 2 {
		lines = SingleString(lines.At(0).S + " {")
	}
	if lines.Size() != 1 {
		panic("command: single line command executed on several lines")
	}
	first := lines.At(0)
	line := c.trim(first).S
	if c.Forbidden != nil && c.Forbidden(line) {
		return Error("Syntax error: " + line)
	}
	arg := c.Pattern.Matcher(line)
	if arg == nil {
		return Error("Cannot parse line " + line)
	}
	return c.Exec(d, first.Location, arg, pass)
}

// EligibleFor implements Command.
func (c *SingleLine[D]) EligibleFor(pass ParserPass) bool {
	if len(c.Passes) == 0 {
		return pass == One
	}
	return slices.Contains(c.Passes, pass)
}
