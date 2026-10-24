// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandStyleImport.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandStyleMultilinesCSS.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandStyleSingleLineCSS.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandMultilines2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/MultilinesStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/Trim.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"errors"
	"io/fs"
	"path"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
)

// StyleDiagram is a diagram that has skin parameters, upstream's
// TitledDiagram as the style commands see it.
type StyleDiagram interface {
	command.Diagram
	SkinParam() ISkinParam
}

// applyPendingStyleMigration is upstream's cast of the skin parameters to
// SkinParam, which throws ClassCastException for other implementations.
func applyPendingStyleMigration(skinParam ISkinParam) {
	sp, ok := skinParam.(interface{ ApplyPendingStyleMigration() })
	if !ok {
		throw("java.lang.ClassCastException", "")
	}
	sp.ApplyPendingStyleMigration()
}

// muteStyles parses lines with the style builder of the diagram and merges
// the styles into it.
func muteStyles(skinParam ISkinParam, lines command.BlocLines) error {
	styleBuilder := skinParam.CurrentStyleBuilder()
	styles, err := NewStyleParserCounter(styleBuilder).Parse(lines)
	if err != nil {
		return err
	}
	skinParam.MuteStyle(styles)
	return nil
}

// styleErrorResult converts a style error to upstream's command result.
func styleErrorResult(err error) command.Result {
	var parsing *StyleParsingError
	if errors.As(err, &parsing) {
		return command.Error("Error in style definition: " + parsing.Message)
	}
	var noStyle *NoStyleAvailableError
	if errors.As(err, &noStyle) {
		return command.Error("General failure: no style available.")
	}
	return command.Error(err.Error())
}

// NewCommandStyleSingleLineCSS returns the "<style>...</style>" command on
// one line.
func NewCommandStyleSingleLineCSS[D StyleDiagram]() command.Command[D] {
	return &command.SingleLine[D]{
		Pattern: regex.NewConcat(regex.Start(),
			regex.NewLeaf(`\<style\>`),
			regex.NewNamedLeaf(1, "STYLE", `([-#.\w%s;:{}]+)`),
			regex.NewLeaf(`\</style\>`),
			regex.End()),
		Exec: func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result {
			line, _ := arg.Get("STYLE", 0)
			if err := muteStyles(d.SkinParam(), command.SingleString(line)); err != nil {
				return styleErrorResult(err)
			}
			applyPendingStyleMigration(d.SkinParam())
			return command.Ok()
		},
	}
}

var styleMultilinesEnd = regex.MustCompile(`^[%s]*\</?style\>[%s]*$`)

// commandStyleMultilinesCSS is the "<style>" ... "</style>" block, on
// upstream's CommandMultilines2 with the REMOVE_STARTING_QUOTE strategy and
// Trim.BOTH.
type commandStyleMultilinesCSS[D StyleDiagram] struct {
	starting *regex.Concat
}

// NewCommandStyleMultilinesCSS returns the "<style>" block command.
func NewCommandStyleMultilinesCSS[D StyleDiagram]() command.Command[D] {
	return &commandStyleMultilinesCSS[D]{
		starting: regex.NewConcat(regex.Start(), regex.NewLeaf(`\<style\>`), regex.End()),
	}
}

// cleanListRemoveStartingQuote drops the lines that start with a quote
// (comments), upstream's MultilinesStrategy.REMOVE_STARTING_QUOTE.
func cleanListRemoveStartingQuote(lines command.BlocLines) command.BlocLines {
	var kept []text.StringLocated
	for _, l := range lines.Lines() {
		if !strings.HasPrefix(l.Trimmed().S, "'") {
			kept = append(kept, l)
		}
	}
	return command.NewBlocLines(kept)
}

func (c *commandStyleMultilinesCSS[D]) IsValid(lines command.BlocLines) command.Control {
	lines = cleanListRemoveStartingQuote(lines)
	first, ok := lines.First()
	if !ok {
		return command.NotOK
	}
	if !c.starting.Match(first.Trimmed().S) {
		return command.NotOK
	}
	if lines.Size() == 1 {
		return command.OKPartial
	}
	if !styleMultilinesEnd.MatchString(lines.Last().Trimmed().S) {
		return command.OKPartial
	}
	return command.OK
}

func (c *commandStyleMultilinesCSS[D]) Execute(d D, lines command.BlocLines, pass command.ParserPass) command.Result {
	lines = cleanListRemoveStartingQuote(lines)
	all := lines.Lines()
	if len(all) < 2 {
		throw(illegalArgument, "fromIndex > toIndex")
	}
	if err := muteStyles(d.SkinParam(), command.NewBlocLines(all[1:len(all)-1])); err != nil {
		return styleErrorResult(err)
	}
	applyPendingStyleMigration(d.SkinParam())
	return command.Ok()
}

func (c *commandStyleMultilinesCSS[D]) EligibleFor(pass command.ParserPass) bool {
	return pass == command.One
}

// NewCommandStyleImport returns the "<style file=path>" command, which
// loads a style sheet from the render context's file system or from the
// embedded skins. Remote (http, https) style sheets are not supported yet
// and give an explicit error.
func NewCommandStyleImport[D StyleDiagram]() command.Command[D] {
	return &command.SingleLine[D]{
		Pattern: regex.NewConcat(regex.Start(),
			regex.NewLeaf(`\<style`),
			regex.SpaceZeroOrMore(),
			regex.NewLeaf(`\w+`),
			regex.SpaceZeroOrMore(),
			regex.NewLeaf("="),
			regex.SpaceZeroOrMore(),
			regex.NewLeaf("[%q%g]?"),
			regex.NewNamedLeaf(1, "PATH", "([^%q%g]*)"),
			regex.NewLeaf("[%q%g]?"),
			regex.NewLeaf(`\>`),
			regex.End()),
		Exec: func(d D, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result {
			p, _ := arg.Get("PATH", 0)
			if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
				return command.Error("Cannot read: " + p + " (remote style sheets are not supported yet)")
			}
			lines, found, err := readStyleFile(d.SkinParam(), p, location)
			if err != nil {
				return command.Error("Cannot read: " + p)
			}
			if !found || lines.Size() == 0 {
				return command.Error("Cannot read: " + p)
			}
			if err := muteStyles(d.SkinParam(), lines); err != nil {
				return styleErrorResult(err)
			}
			return command.Ok()
		},
	}
}

// readStyleFile reads a style sheet: the file relative to the current
// directory of the render context, else the embedded /skin resource.
func readStyleFile(skinParam ISkinParam, p string, location *text.LineLocation) (command.BlocLines, bool, error) {
	if ctx := skinParam.PathSystem(); ctx != nil && ctx.FS != nil {
		name := path.Join(ctx.CurrentDir, p)
		if fs.ValidPath(name) {
			data, err := fs.ReadFile(ctx.FS, name)
			if err == nil {
				return loadBlocLines(string(data), location), true, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return command.BlocLines{}, false, err
			}
		}
	}
	data, err := skinFiles.ReadFile("skin/" + p)
	if err != nil {
		return command.BlocLines{}, false, nil
	}
	return loadBlocLines(string(data), location), true, nil
}
