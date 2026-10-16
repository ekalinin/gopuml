// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandMultilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandMultilines2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandMultilinesBracket.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/MultilinesStrategy.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/Trim.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/SingleLineCommand2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/regex/RegexConcat.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/BlocLines.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Strategy tells whether the comment lines of a multi-line command are
// kept, upstream's MultilinesStrategy.
type Strategy int

// Strategies.
const (
	RemoveStartingQuote Strategy = iota
	KeepStartingQuote
)

// CleanList applies the strategy: RemoveStartingQuote drops the lines whose
// trimmed text starts with a quote.
func (s Strategy) CleanList(lines command.BlocLines) command.BlocLines {
	if s != RemoveStartingQuote {
		return lines
	}
	var kept []text.StringLocated
	for _, l := range lines.Lines() {
		if !strings.HasPrefix(l.Trimmed().S, "'") {
			kept = append(kept, l)
		}
	}
	return command.NewBlocLines(kept)
}

// Trim tells how the last line of a multi-line command is trimmed before it
// is matched against the end pattern, upstream's Trim.
type Trim int

// Trims.
const (
	TrimBoth Trim = iota
	TrimLeftOnly
	TrimNone
)

var ltrim = regexp.MustCompile(`^[\t\n\x0B\f\r ]+`)

// Apply trims the line.
func (t Trim) Apply(s text.StringLocated) string {
	switch t {
	case TrimNone:
		return s.S
	case TrimBoth:
		return s.Trimmed().S
	}
	return ltrim.ReplaceAllString(s.S, "")
}

// SubExtract is upstream's BlocLines.subExtract: the lines without the
// first margeStart and the last margeEnd ones. It panics as Java's subList
// when there are fewer lines.
func SubExtract(lines command.BlocLines, margeStart, margeEnd int) command.BlocLines {
	all := lines.Lines()
	end := len(all) - margeEnd
	if margeStart > end || end > len(all) || margeStart < 0 {
		Throw(IllegalArgumentException, "fromIndex("+itoa(margeStart)+") > toIndex("+itoa(end)+")")
	}
	return command.NewBlocLines(all[margeStart:end])
}

// Location is upstream's BlocLines.getLocation: the location of the first
// line, nil without lines.
func Location(lines command.BlocLines) *text.LineLocation {
	first, ok := lines.First()
	if !ok {
		return nil
	}
	return first.Location
}

// TrimSmart is upstream's BlocLines.trimSmart: from line referenceLine on,
// the leading spaces and tabulations of that line are removed from every
// line (at most as many as each line has).
func TrimSmart(lines command.BlocLines, referenceLine int) command.BlocLines {
	all := lines.Lines()
	if len(all) <= referenceLine {
		return lines
	}
	out := append([]text.StringLocated(nil), all...)
	nb := nbStartingSpace(out[referenceLine].S)
	for i := referenceLine; i < len(out); i++ {
		s := out[i]
		if s.S == "" {
			continue
		}
		j := 0
		for j < nb && j < len(s.S) && (s.S[j] == ' ' || s.S[j] == '\t') {
			j++
		}
		if j > 0 {
			out[i] = s.Substring(j, len(s.S))
		}
	}
	return command.NewBlocLines(out)
}

func nbStartingSpace(s string) int {
	n := 0
	for n < len(s) && (s[n] == ' ' || s[n] == '\t') {
		n++
	}
	return n
}

// EventuallyMoveBracket is upstream's BlocLines.eventuallyMoveBracket: a
// "{" alone on the second line joins the first line.
func EventuallyMoveBracket(lines command.BlocLines) command.BlocLines {
	all := lines.Lines()
	if len(all) < 2 {
		return lines
	}
	first, second := all[0].Trimmed().S, all[1].Trimmed().S
	if strings.HasSuffix(first, "{") || second != "{" {
		return lines
	}
	out := []text.StringLocated{{S: all[0].S + " {", Location: all[0].Location,
		PreprocessorError: all[0].PreprocessorError}}
	return command.NewBlocLines(append(out, all[2:]...))
}

// Multilines is a command spanning a start line and an end line, upstream's
// CommandMultilines: the trimmed first line must match Start and the trimmed
// last line End. Both patterns are upstream patterns compiled with
// regex.Compile, anchored with ^ and $.
type Multilines[D command.Diagram] struct {
	Start, End *regexp.Regexp
	Exec       func(d D, lines command.BlocLines, pass command.ParserPass) command.Result
}

// IsValid implements command.Command.
func (c *Multilines[D]) IsValid(lines command.BlocLines) command.Control {
	first, ok := lines.First()
	if !ok {
		ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.text.StringLocated.getTrimmed()" because the return value of "net.sourceforge.plantuml.utils.BlocLines.getFirst()" is null`)
	}
	if !c.Start.MatchString(first.Trimmed().S) {
		return command.NotOK
	}
	if lines.Size() == 1 {
		return command.OKPartial
	}
	if !c.End.MatchString(lines.Last().Trimmed().S) {
		return command.OKPartial
	}
	return command.OK
}

// Execute implements command.Command.
func (c *Multilines[D]) Execute(d D, lines command.BlocLines, pass command.ParserPass) command.Result {
	return c.Exec(d, lines, pass)
}

// EligibleFor implements command.Command.
func (c *Multilines[D]) EligibleFor(pass command.ParserPass) bool { return pass == command.One }

// Multilines2 is a command whose first line matches a composed pattern and
// whose last line matches End, upstream's CommandMultilines2. The lines are
// cleaned with Strategy first; the last line is trimmed with TrimEnd.
type Multilines2[D command.Diagram] struct {
	Start    *regex.Concat
	Strategy Strategy
	TrimEnd  Trim
	End      *regexp.Regexp
	// FinalBracket accepts "line {" also written as "line" then "{".
	FinalBracket bool
	// Exec runs the command on the cleaned lines.
	Exec func(d D, lines command.BlocLines, pass command.ParserPass) command.Result
}

// IsValid implements command.Command.
func (c *Multilines2[D]) IsValid(lines command.BlocLines) command.Control {
	lines = c.Strategy.CleanList(lines)
	if c.FinalBracket {
		if lines.Size() == 1 && !strings.HasSuffix(lines.At(0).Trimmed().S, "{") {
			if c.IsValid(command.SingleString(lines.At(0).S+" {")) == command.OKPartial {
				return command.OKPartial
			}
			return command.NotOK
		}
		lines = EventuallyMoveBracket(lines)
	}
	first, ok := lines.First()
	if !ok {
		return command.NotOK
	}
	if !c.Start.Match(first.Trimmed().S) {
		return command.NotOK
	}
	if lines.Size() == 1 {
		return command.OKPartial
	}
	if !c.End.MatchString(c.TrimEnd.Apply(lines.Last())) {
		return command.OKPartial
	}
	return command.OK
}

// Execute implements command.Command.
func (c *Multilines2[D]) Execute(d D, lines command.BlocLines, pass command.ParserPass) command.Result {
	lines = c.Strategy.CleanList(lines)
	if c.FinalBracket {
		lines = EventuallyMoveBracket(lines)
	}
	return c.Exec(d, lines, pass)
}

// EligibleFor implements command.Command.
func (c *Multilines2[D]) EligibleFor(pass command.ParserPass) bool { return pass == command.One }

// StartArg matches the trimmed first line against the start pattern, as
// upstream's getStartingPattern().matcher(...); nil when it does not match.
func (c *Multilines2[D]) StartArg(lines command.BlocLines) *regex.Result {
	return c.Start.Matcher(lines.At(0).Trimmed().S)
}

// MultilinesBracket is a command made of a first line matching Start and of
// the following lines up to the balancing "}", upstream's
// CommandMultilinesBracket. Each inner line must satisfy Consistent.
type MultilinesBracket[D command.Diagram] struct {
	Start      *regexp.Regexp
	Consistent func(line string, level int) bool
	Exec       func(d D, lines command.BlocLines, pass command.ParserPass) command.Result
}

// IsValid implements command.Command.
func (c *MultilinesBracket[D]) IsValid(lines command.BlocLines) command.Control {
	if !c.Start.MatchString(lines.At(0).Trimmed().S) {
		return command.NotOK
	}
	if lines.Size() == 1 {
		return command.OKPartial
	}
	level := 1
	for _, cs := range lines.Lines()[1:] {
		s := cs.Trimmed().S
		if !c.Consistent(s, level) {
			return command.NotOK
		}
		if strings.HasSuffix(s, "{") {
			level++
		}
		if strings.HasSuffix(s, "}") {
			level--
		}
		if level < 0 {
			return command.NotOK
		}
	}
	if level != 0 {
		return command.OKPartial
	}
	return command.OK
}

// Execute implements command.Command.
func (c *MultilinesBracket[D]) Execute(d D, lines command.BlocLines, pass command.ParserPass) command.Result {
	return c.Exec(d, lines, pass)
}

// EligibleFor implements command.Command.
func (c *MultilinesBracket[D]) EligibleFor(pass command.ParserPass) bool { return pass == command.One }

// Guarded is a single-line command with the checks of upstream's
// SingleLineCommand2 and RegexConcat that command.SingleLine lacks: a
// maximum line length (RegexConcat.protectSize, in UTF-16 units) and a
// rejection of lines that a Java-only pattern feature (a lookbehind)
// refuses. Both checks run before the pattern is matched, so a rejected
// line is not this command.
type Guarded[D command.Diagram] struct {
	*command.SingleLine[D]
	// MaxSize, when not zero, rejects longer lines.
	MaxSize int
	// Reject, when set, rejects the trimmed lines it returns true for,
	// given the groups of the match.
	Reject func(arg *regex.Result) bool
}

// IsValid implements command.Command.
func (c *Guarded[D]) IsValid(lines command.BlocLines) command.Control {
	if lines.Size() == 1 {
		line := lines.At(0)
		if !c.NoTrim {
			line = line.Trimmed()
		}
		if c.MaxSize != 0 && JavaLength(line.S) > c.MaxSize {
			return command.NotOK
		}
		if c.Reject != nil {
			if arg := c.Pattern.Matcher(line.S); arg != nil && c.Reject(arg) {
				return command.NotOK
			}
		}
	}
	return c.SingleLine.IsValid(lines)
}

// Single returns a single-line command, upstream's SingleLineCommand2 with
// trimmed lines.
func Single[D command.Diagram](pattern *regex.Concat,
	exec func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result) *command.SingleLine[D] {
	return &command.SingleLine[D]{Pattern: pattern, Exec: exec}
}

func itoa(i int) string {
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}
