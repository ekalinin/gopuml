// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/GraphvizUtils.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/OptionPrint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"context"
	"strings"
)

// dotVersionLimit is the oldest Graphviz version (2.26) considered recent
// enough.
const dotVersionLimit = 226

// testDotSource is the graph laid out by the Graphviz check.
const testDotSource = "digraph foo { test; }"

// CheckGraphviz runs the Graphviz check of --check-graphviz (upstream's
// -testdot). It returns the report lines and upstream's exit status: 0 when
// the check passed, and then the lines go to stdout; otherwise the lines go
// to stderr and the process exits with that status (12 to 16). insecure tells
// whether the security profile is INSECURE, which adds the GRAPHVIZ_DOT
// setting and the dot path to the report.
func CheckGraphviz(ctx context.Context, c Config, insecure bool) (lines []string, exitCode int) {
	return AddDotStatus(ctx, c, insecure, false)
}

// AddDotStatus returns the Graphviz status lines and an error code (0 when
// Graphviz works). withRichText adds the creole markup used by the "testdot"
// and "version" diagrams.
func AddDotStatus(ctx context.Context, c Config, insecure, withRichText bool) (lines []string, errorCode int) {
	red, bold := "", ""
	if withRichText {
		red = "<b><color:red>"
		bold = "<b>"
	}

	// The VizJs branch of upstream is not ported: the MIT build has no VizJs
	// engine, so upstream never takes it.

	dotExe := SearchDotExe(c.DotPath, c.Host)
	if insecure {
		if ent := GetenvGraphvizDot(c.DotPath, c.Host); ent == "" {
			lines = append(lines, "The environment variable GRAPHVIZ_DOT has not been set")
		} else {
			lines = append(lines, "The environment variable GRAPHVIZ_DOT has been set to "+ent)
		}
		lines = append(lines, "Dot executable is "+javaFileString(dotExe))
	}

	exeState := CheckFile(dotExe, c.Host)
	if exeState != OK {
		lines = append(lines, red+"Error: "+exeState.TextMessage())
		lines = append(lines, "Error: only sequence diagrams will be generated")
		return lines, 16
	}

	e := &Exec{Path: dotExe, Timeout: c.Timeout}
	version := e.DotVersion(ctx)
	if withRichText {
		lines = append(lines, `<b>GraphViz</b> ""`+version+`""`)
	} else {
		lines = append(lines, "GraphViz: "+version)
	}
	switch v := RetrieveVersion(version); {
	case v == -1:
		lines = append(lines, "Warning : cannot determine dot version")
		errorCode = 12
	case v < dotVersionLimit:
		lines = append(lines, bold+"Warning : Your dot installation seems old")
		lines = append(lines, bold+"Some diagrams may have issues")
		errorCode = 13
	default:
		if err := e.testCreateSimpleFile(ctx); err != "" {
			lines = append(lines, red+err)
			errorCode = 14
		} else {
			lines = append(lines, bold+"Installation seems OK. File generation OK")
		}
	}
	return lines, errorCode
}

// testCreateSimpleFile lays out a one-node graph and returns "" when dot
// produced SVG, or the error message. As upstream, dot's stderr is mixed into
// the output and the exit status is not checked.
func (e *Exec) testCreateSimpleFile(ctx context.Context) string {
	p, state := e.createFile3(ctx, testDotSource, true, "svg")
	if state.differs(terminatedOK) {
		return "Error: timeout " + state.String()
	}
	if len(p.stdout) == 0 {
		return "Error: dot generates empty file. Check you dot installation."
	}
	if !strings.Contains(string(p.stdout), "<svg") {
		return "Error: dot generates unreadable SVG file. Check you dot installation."
	}
	return ""
}
