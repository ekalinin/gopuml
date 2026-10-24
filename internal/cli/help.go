// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/HelpPrint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/HelpGroup.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/HelpTable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/ExitStatus.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// Exit statuses, as upstream.
const (
	ExitOK          = 0
	ExitNoFileFound = 50
	ExitNoDiagram   = 100
	ExitSomeErrors  = 200
	ExitParseError  = 42 // CliParsingException, such as -D without a value
	ExitUnsupported = 1  // this implementation: an option it does not provide
)

// usage is upstream's getUsage: the left column of the help.
func (f Flag) usage() string {
	info := f.Info()
	if info.Usage != "" {
		if strings.HasPrefix(info.Usage, "--") {
			return "     " + info.Usage
		}
		return " " + info.Usage
	}
	for _, a := range info.Aliases {
		if len(a) == 2 {
			return " " + a + ", " + info.Flag
		}
	}
	return "     " + info.Flag
}

var nonWord = regexp.MustCompile(`\W`)

// sortKey is HelpGroup's comparator key.
func (f Flag) sortKey() string {
	u := f.usage()
	letters := strings.ToLower(nonWord.ReplaceAllString(u, ""))
	if strings.HasPrefix(strings.TrimSpace(u), "--") {
		return letters[:1] + letters
	}
	return letters
}

type helpGroup struct {
	title string
	flags []Flag
}

// add keeps the flags sorted and, like upstream's TreeSet, drops a flag
// whose key equals one already present.
func (g *helpGroup) add(f Flag) {
	i, found := slices.BinarySearchFunc(g.flags, f, func(a, b Flag) int { return strings.Compare(a.sortKey(), b.sortKey()) })
	if !found {
		g.flags = slices.Insert(g.flags, i, f)
	}
}

func helpRows(limit int) [][2]string {
	groups := []*helpGroup{{title: "General"}}
	for f := range Flag(flagCount) {
		info := f.Info()
		if info.Doc == "" || info.Level > limit {
			continue
		}
		if info.NewGroup != "" {
			groups = append(groups, &helpGroup{title: info.NewGroup})
		}
		groups[len(groups)-1].add(f)
	}
	var rows [][2]string
	for _, g := range groups {
		rows = append(rows, [2]string{"", g.title})
		for _, f := range g.flags {
			rows = append(rows, [2]string{f.usage(), f.Info().Doc})
		}
	}
	return rows
}

func pad(s string, size int, fill byte) string {
	if len(s) == size-1 {
		return s + " "
	}
	if len(s) >= size {
		return s
	}
	return s + strings.Repeat(string(fill), size-len(s))
}

func printTable(w io.Writer, rows [][2]string) {
	size0 := 0
	for _, r := range rows {
		first, _, _ := strings.Cut(r[0], "\n")
		size0 = max(size0, len(first))
	}
	for _, r := range rows {
		if r[0] == "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, r[1]+":")
			continue
		}
		first := strings.Split(r[0], "\n")
		second := strings.Split(r[1], "\n")
		if len(first) == 2 {
			fmt.Fprintln(w, first[0])
			fmt.Fprintln(w, "     "+pad(first[1]+" ", size0+2-5, '.')+" "+second[0])
			continue
		}
		fmt.Fprintln(w, pad(r[0]+" ", size0+2, '.')+" "+second[0])
		for _, s := range second[1:] {
			fmt.Fprintln(w, pad(" ", size0+2, ' ')+" "+s)
		}
	}
}

// PrintExitCodes prints the exit statuses, as upstream.
func PrintExitCodes(w io.Writer) {
	for _, e := range []struct {
		code int
		text string
	}{{ExitOK, "Success"}, {ExitNoFileFound, "No file found"}, {ExitNoDiagram, "No diagram found in file(s)"}, {ExitSomeErrors, "Some diagrams have syntax errors"}} {
		fmt.Fprintf(w, "  %-4d %s\n", e.code, e.text)
	}
}

// PrintHelp writes upstream's help; more includes the advanced options.
// prog replaces "java -jar plantuml.jar" in the examples.
func PrintHelp(w io.Writer, prog string, more bool) {
	limit := 0
	if more {
		limit = 1
	}
	lines := []string{
		"plantuml - generate diagrams from plain text",
		"",
		"Usage:",
		"  " + prog + " [options] [file|dir]...",
		"  " + prog + " [options] --gui",
		"",
		"Description:",
		"  Process PlantUML sources from files, directories (optionally recursive), or stdin (-pipe).",
		"",
		"Wildcards (for files/dirs):",
		"  *   any characters except '/' and '\\'",
		"  ?   exactly one character except '/' and '\\'",
		"  **  any characters across directories (recursive)",
		"  Tip: quote patterns to avoid shell expansion (e.g., \"**/*.puml\").",
		"",
		"General:",
	}
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	printTable(w, helpRows(limit))
	footer := []string{
		"",
		"",
		"Examples:",
		"  # Process all .puml recursively",
		"  " + prog + " \"**/*.puml\"",
		"",
		"  # Check syntax only (CI)",
		"  " + prog + " --check-syntax src/diagrams",
		"",
		"  # Read from stdin and write to stdout (SVG)",
		"  cat diagram.puml | " + prog + " --svg -pipe > out.svg",
		"",
		"  # Encode a sprite from an image",
		"  " + prog + " --sprite 16z myicon.png",
		"",
		"  # Use a define",
		"  " + prog + " -DAUTHOR=John diagram.puml",
		"",
		"  # Change output directory",
		"  " + prog + " --format svg --output-dir out diagrams/",
		"",
		"Exit codes:",
	}
	fmt.Fprintln(w, strings.Join(footer, "\n"))
	PrintExitCodes(w)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "See also:")
	fmt.Fprintln(w, "  "+prog+" --help-more")
	fmt.Fprintln(w, "  Documentation: https://plantuml.com")
}
