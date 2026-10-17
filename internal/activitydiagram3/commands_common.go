// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandMultilines3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandDecoratorMultine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandFootboxIgnored.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/BlocLines.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"errors"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*ActivityDiagram3]

// colorResult turns an error of a color lookup into upstream's result for
// NoSuchColorException; other errors are bugs and panic.
func colorResult(err error) command.Result {
	var noSuch *color.NoSuchColorError
	if errors.As(err, &noSuch) {
		return command.BadColor()
	}
	panic(err)
}

// throwNoSuchColor panics with upstream's NoSuchColorException for an
// unknown color, for the commands that let it reach ProtectedCommand; other
// errors are bugs and panic as they are.
func throwNoSuchColor(err error) {
	var noSuch *color.NoSuchColorError
	if errors.As(err, &noSuch) {
		uml.Throw("net.sourceforge.plantuml.klimt.color.NoSuchColorException", noSuch.Error())
	}
	panic(err)
}

// getColor is HColorSet.getColor of the diagram's color set.
func getColor(d *ActivityDiagram3, s string) (color.HColor, error) {
	return d.Skin().IHtmlColorSet().GetColor(s)
}

// urlBuilder is the UrlBuilder of the diagram, with its "topurl" and the
// strict mode.
func urlBuilder(d *ActivityDiagram3) *klimt.URLBuilder {
	topurl, ok := d.Skin().Value("topurl")
	return klimt.NewURLBuilder(topurl, ok, klimt.URLModeStrict)
}

// optionalURL returns the link captured by klimt.URLBuilderOptional, nil
// when there is none.
func optionalURL(d *ActivityDiagram3, arg *regex.Result) *klimt.URL {
	s, ok := arg.Get(klimt.URLBuilderKey, 0)
	if !ok {
		return nil
	}
	return urlBuilder(d).URL(s)
}

// The arrow styles of CommandLinkElement: "-[#red,dashed]->", several
// styles separated by ";".
const (
	linkElementKey1      = `dotted|dashed|plain|bold|hidden|norank|single|node|thickness=\d+`
	linkElementKey2      = `,dotted|,dashed|,plain|,bold|,hidden|,norank|,single|,node|,thickness=\d+`
	linkElementLineStyle = `(?:#\w+|` + linkElementKey1 + `)(?:,#\w+|` + linkElementKey2 + `)*`
	lineStyleMultiples   = linkElementLineStyle + `(?:(?:;` + linkElementLineStyle + `)*)`
	styleColorsMultiples = `-\[(` + lineStyleMultiples + `*)\]->`
)

// rainbowOf is Rainbow.build with the separation space of the diagram: the
// colors of an arrow style captured under key; none when the group did not
// match.
func rainbowOf(d *ActivityDiagram3, arg *regex.Result, key string) (*Rainbow, error) {
	s, ok := arg.Get(key, 0)
	return RainbowBuild(d.Skin(), s, ok, d.Skin().ColorArrowSeparationSpace())
}

// backRendering is CommandBackward3.getBackRendering: the arrow whose
// colors are captured under name+"_COLOR" and whose label is captured
// under name.
func backRendering(d *ActivityDiagram3, arg *regex.Result, name string) (*LinkRendering, error) {
	in := LinkRenderingNone()
	if colorString, ok := arg.Get(name+"_COLOR", 0); ok {
		incomingColor, err := RainbowBuild(d.Skin(), colorString, true, d.Skin().ColorArrowSeparationSpace())
		if err != nil {
			return nil, err
		}
		in = LinkRenderingCreate(incomingColor)
	}
	label, ok := arg.Get(name, 0)
	return in.WithDisplay(klimt.DisplayWithNewlinesString(d.Pragma(), label, ok)), nil
}

// withNewlines is Display.getWithNewlines(pragma, arg.get(key, 0)): the
// null display when the group did not match.
func withNewlines(d *ActivityDiagram3, arg *regex.Result, key string) *klimt.Display {
	s, ok := arg.Get(key, 0)
	return klimt.DisplayWithNewlinesString(d.Pragma(), s, ok)
}

// testDisplay is the display of a "(test)" captured under key: an empty
// test is null, as the commands do with "if (test.length() == 0) test =
// null".
func testDisplay(d *ActivityDiagram3, arg *regex.Result, key string) *klimt.Display {
	s, ok := arg.Get(key, 0)
	if !ok {
		uml.ThrowNPE(`Cannot invoke "String.length()" because "test" is null`)
	}
	return klimt.DisplayWithNewlinesString(d.Pragma(), s, s != "")
}

// innerColors is Stereogroup.getInnerColors with the diagram's color set.
func innerColors(d *ActivityDiagram3, stereogroup *klimt.Stereogroup) (*color.Colors, error) {
	return stereogroup.InnerColors(d.Skin().IHtmlColorSet())
}

// single returns a single-line command, upstream's SingleLineCommand2.
func single(exec func(d *ActivityDiagram3, location *text.LineLocation, arg *regex.Result) command.Result,
	parts ...regex.Part) *command.SingleLine[*ActivityDiagram3] {
	return uml.Single(regex.NewConcat(parts...),
		func(d *ActivityDiagram3, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return exec(d, location, arg)
		})
}

// commandFootboxIgnored is CommandFootboxIgnored: "hide footbox" is
// accepted and ignored.
func commandFootboxIgnored() cmd {
	return single(func(*ActivityDiagram3, *text.LineLocation, *regex.Result) command.Result { return command.Ok() },
		regex.Start(),
		uml.Leaf(1, "", "(hide|show)?"),
		regex.SpaceZeroOrMore(),
		uml.Lit("footbox"), regex.End())
}

// multilines3 is a command whose first line matches start and whose last
// line matches end, both composed patterns, upstream's CommandMultilines3.
// The lines are cleaned with strategy first; the last line is trimmed with
// trimEnd (TrimNone or TrimBoth).
type multilines3 struct {
	start    *regex.Concat
	strategy uml.Strategy
	trimEnd  uml.Trim
	end      *regex.Concat
	exec     func(d *ActivityDiagram3, lines command.BlocLines) command.Result
}

// IsValid implements command.Command.
func (c *multilines3) IsValid(lines command.BlocLines) command.Control {
	lines = c.strategy.CleanList(lines)
	first, ok := lines.First()
	if !ok {
		return command.NotOK
	}
	if !c.start.Match(first.Trimmed().S) {
		return command.NotOK
	}
	if lines.Size() == 1 {
		return command.OKPartial
	}
	var potentialLast text.StringLocated
	switch c.trimEnd {
	case uml.TrimNone:
		potentialLast = lines.Last()
	case uml.TrimBoth:
		potentialLast = lines.Last().Trimmed()
	default:
		uml.Throw(uml.IllegalStateException, "")
	}
	if !c.end.Match(potentialLast.S) {
		return command.OKPartial
	}
	return command.OK
}

// Execute implements command.Command.
func (c *multilines3) Execute(d *ActivityDiagram3, lines command.BlocLines, _ command.ParserPass) command.Result {
	return c.exec(d, c.strategy.CleanList(lines))
}

// EligibleFor implements command.Command.
func (c *multilines3) EligibleFor(pass command.ParserPass) bool { return pass == command.One }

// decoratorMultine is CommandDecoratorMultine: a single-line command whose
// line may be split over at most nbMaxLines lines. The lines are joined
// with the hidden newline of the preprocessor; any line that the joined
// text does not match may start the command.
type decoratorMultine struct {
	cmd        *command.SingleLine[*ActivityDiagram3]
	nbMaxLines int
}

func newDecoratorMultine(c *command.SingleLine[*ActivityDiagram3], nbMaxLines int) *decoratorMultine {
	return &decoratorMultine{cmd: c, nbMaxLines: nbMaxLines}
}

// toSingleLineWithHiddenNewLine is BlocLines.toSingleLineWithHiddenNewLine:
// the lines joined with the hidden newline, as a line without location.
func toSingleLineWithHiddenNewLine(lines command.BlocLines) command.BlocLines {
	return command.SingleString(strings.Join(lines.Strings(), string(klimt.JawsBlockE1Newline)))
}

// IsValid implements command.Command.
func (c *decoratorMultine) IsValid(lines command.BlocLines) command.Control {
	switch c.cmd.IsValid(toSingleLineWithHiddenNewLine(lines)) {
	case command.OK:
		return command.OK
	case command.OKPartial:
		uml.Throw(uml.IllegalStateException, "")
	}
	return command.OKPartial
}

// Execute implements command.Command.
func (c *decoratorMultine) Execute(d *ActivityDiagram3, lines command.BlocLines, pass command.ParserPass) command.Result {
	return c.cmd.Execute(d, toSingleLineWithHiddenNewLine(lines), pass)
}

// EligibleFor implements command.Command.
func (c *decoratorMultine) EligibleFor(pass command.ParserPass) bool { return c.cmd.EligibleFor(pass) }

// MaxLines implements command.MaxLinesCommand.
func (c *decoratorMultine) MaxLines() int { return c.nbMaxLines }
