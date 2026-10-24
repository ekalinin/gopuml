// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/error/PSystemError.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/error/PSystemErrorV2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/error/PSystemErrorPreprocessor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/crash/ReportLog.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package errdiagram

import (
	"strconv"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/version"
)

// errorTextBlock is PSystemError.getTextBlock: the error, below the welcome
// help for a source of fewer than five lines. The time-based sponsor
// messages and the "arecibo" image are never added (see the package
// documentation).
func errorTextBlock(d *core.ErrorDiagram, src *core.UmlSource) klimt.TextBlock {
	result := graphicalFormatted(d)
	if src.TotalLineCount() < 5 {
		result = klimt.TextBlockUtilsMergeTB(welcomeTextBlock(geom.GraphicPositionBackgroundCornerTopRight), result,
			geom.HorizontalAlignmentLeft)
	}
	return result
}

// graphicalFormatted is PSystemError.getGraphicalFormatted: on black, the
// version line, the location of the error on green, the lines read before
// it in green bold, the last one with a red wave, and the message in red.
func graphicalFormatted(d *core.ErrorDiagram) klimt.TextBlock {
	fc0 := klimt.GraphicStringsSansSerif14(color.Black).Bold()
	fc1 := klimt.GraphicStringsSansSerif14(color.MyGreen).Bold()
	fc2 := klimt.GraphicStringsSansSerif14(color.Red).Bold()
	fc4 := klimt.GraphicStringsSansSerif12(color.MyGreen).Bold().Italic()

	fullBody := textFullBody(d.Trace)
	result0 := klimt.TextBlockUtilsAddBackcolor(
		klimt.TextBlockUtilsWithMargin4(klimt.NewTextBlockRaw(textFromStack(d.Trace), fc0), 1, 1, 1, 4), color.MyGreen)
	result1 := klimt.NewTextBlockRaw(fullBody[:len(fullBody)-1], fc1)
	result2 := klimt.NewTextBlockRaw(fullBody[len(fullBody)-1:], fc1.Wave(color.Red))
	result3 := klimt.NewTextBlockRaw([]string{" " + d.Error.Message()}, fc2)
	result4 := klimt.TextBlockUtilsWithMargin4(klimt.NewTextBlockRaw(header(), fc4), 0, 2, 0, 8)
	var result klimt.TextBlock = result0
	result = klimt.TextBlockUtilsMergeTB(result, result1, geom.HorizontalAlignmentLeft)
	result = klimt.TextBlockUtilsMergeTB(result, result2, geom.HorizontalAlignmentLeft)
	result = klimt.TextBlockUtilsMergeTB(result, result3, geom.HorizontalAlignmentLeft)
	result = klimt.TextBlockUtilsMergeTB(result4, result, geom.HorizontalAlignmentLeft)
	result = klimt.TextBlockUtilsWithMargin(result, 5, 5)
	return klimt.TextBlockUtilsAddBackcolor(result, color.Black)
}

// header is PSystemError.header: the version line. ReportLog's warning for
// a build older than 140 days depends on the clock and is never added.
func header() []string { return []string{version.FullDescription()} }

// textFromStack is PSystemError.getTextFromStack: the location of the last
// line of the trace, then the locations of the includes that led to it.
func textFromStack(trace []text.StringLocated) []string {
	var result []string
	if len(trace) == 0 {
		// Upstream's getLastLine fails on an empty trace.
		return result
	}
	for loc := trace[len(trace)-1].Location; loc != nil; loc = loc.Parent {
		result = append(result, "[From "+loc.Desc+" (line "+strconv.Itoa(loc.Position+1)+") ]")
	}
	return result
}

// textFullBody is PSystemError.getTextFullBody: a blank line, then the
// trace, shortened to its first 5 and last 20 lines when it has more than
// 40, each line cut after 120 characters.
func textFullBody(trace []text.StringLocated) []string {
	result := []string{" "}
	traceSize := len(trace)
	if traceSize > 40 {
		for _, s := range trace[:5] {
			result = addToResult(result, s)
		}
		result = append(result, "...")
		skipped := traceSize - 5 - 20
		result = append(result, "... ( skipping "+strconv.Itoa(skipped)+" lines )")
		result = append(result, "...")
		for _, s := range trace[traceSize-20:] {
			result = addToResult(result, s)
		}
	} else {
		for _, s := range trace {
			result = addToResult(result, s)
		}
	}
	return result
}

// addToResult adds a line, cut after 120 UTF-16 units. A surrogate pair cut
// in half leaves U+FFFD where Java keeps the lone high surrogate.
func addToResult(result []string, s text.StringLocated) []string {
	tmp := s.S
	if units := utf16.Encode([]rune(tmp)); len(units) > 120 {
		tmp = string(utf16.Decode(units[:120])) + " ..."
	}
	return append(result, tmp)
}

// titlePattern is UmlSource's TITLE.
var titlePattern = regex.MustCompile(`^[%s]*title[%s]+(.+)$`)

// errorWarningOrError is PSystemError.getWarningOrError: the description,
// the title of the source if any, and an empty line.
func errorWarningOrError(src *core.UmlSource) string {
	result := "(Error)\n"
	for _, s := range src.Lines() {
		if m := titlePattern.FindStringSubmatch(s.S); m != nil {
			result += m[1] + "\n"
			break
		}
	}
	return result + "\n"
}

// isPreprocessorError reports whether d is upstream's
// PSystemErrorPreprocessor, whose message is the preprocessor error of the
// last line of the trace: only the preprocessor sets that error.
func isPreprocessorError(d *core.ErrorDiagram) bool {
	n := len(d.Trace)
	return n > 0 && d.Trace[n-1].PreprocessorError != ""
}
