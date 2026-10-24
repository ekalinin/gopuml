// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/Command.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/PSystemCommandFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// FirstTokensCommand is a command that declares the first tokens of the
// lines it may accept (text.StringLocated.FirstToken), upstream's
// mandatoryFirstTokensFast. Nil means any line. A command that does not
// implement it may accept any line.
type FirstTokensCommand interface {
	MandatoryFirstTokens() []string
}

// StartPatternCommand is a command with a starting pattern in upstream's
// syntax, from which its first tokens are read.
type StartPatternCommand interface {
	StartPattern() string
}

// FirstTokensOf returns the first tokens cmd declares, nil when it may
// accept any line.
func FirstTokensOf[D Diagram](cmd Command[D]) []string {
	if c, ok := cmd.(FirstTokensCommand); ok {
		return c.MandatoryFirstTokens()
	}
	return nil
}

// MandatoryFirstTokens implements FirstTokensCommand from the pattern, or
// the declared tokens.
func (c *SingleLine[D]) MandatoryFirstTokens() []string {
	if c.FirstTokens != nil {
		return c.FirstTokens
	}
	return regex.FirstTokens(c.Pattern.Pattern())
}

// StartPattern implements StartPatternCommand.
func (c *SingleLine[D]) StartPattern() string { return c.Pattern.Pattern() }

// MandatoryFirstTokens implements FirstTokensCommand with the tokens of the
// wrapped command.
func (c *inPasses[D]) MandatoryFirstTokens() []string { return FirstTokensOf(c.Command) }

// StartPattern implements StartPatternCommand with the pattern of the
// wrapped command, "" when it has none.
func (c *inPasses[D]) StartPattern() string {
	if p, ok := c.Command.(StartPatternCommand); ok {
		return p.StartPattern()
	}
	return ""
}

// commandTokens are the first tokens of a command, computed once per
// factory: nil for a command that may accept any line.
type commandTokens struct {
	tokens []string
}

// accepts reports whether the command is tried on a line of that token.
func (t commandTokens) accepts(token string) bool {
	return t.tokens == nil || slices.Contains(t.tokens, token)
}

// CommandInfo describes a command of a factory: its starting pattern ("" when
// it has none) and its first tokens (nil for any line).
type CommandInfo struct {
	Pattern string
	Tokens  []string
}

// CommandInfos describes commands, in order.
func CommandInfos[D Diagram](cmds []Command[D]) []CommandInfo {
	var result []CommandInfo
	for _, cmd := range cmds {
		info := CommandInfo{Tokens: FirstTokensOf(cmd)}
		if p, ok := cmd.(StartPatternCommand); ok {
			info.Pattern = p.StartPattern()
		}
		result = append(result, info)
	}
	return result
}
