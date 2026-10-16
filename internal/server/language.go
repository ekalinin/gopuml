// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/syntax/LanguageDescriptor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package server

import (
	"bufio"
	"io"
	"slices"
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// The word lists of LanguageDescriptor, in upstream's source order.
var (
	languageDirectives = []string{
		"board", "bpm", "chart", "chen", "chronology", "creole", "def", "ditaa",
		"dot", "ebnf", "files", "flow", "gantt", "git", "hcl", "jcckit",
		"json", "latex", "math", "mindmap", "nwdiag", "project", "regex", "salt",
		"sprites", "uml", "wbs", "wire", "yaml", "packetdiag",
	}
	languageTypes = []string{
		"abstract", "action", "actor", "agent", "analog", "annotation", "archimate", "artifact",
		"binary", "boundary", "card", "class", "clock", "cloud", "collections", "component",
		"concise", "control", "database", "dataclass", "diamond", "entity", "enum", "exception",
		"file", "folder", "frame", "hexagon", "interface", "json", "label", "map",
		"metaclass", "network", "node", "nwdiag", "object", "package", "participant", "person",
		"process", "protocol", "queue", "record", "rectangle", "relationship", "robust", "stack",
		"state", "storage", "struct", "usecase", "packetdiag",
	}
	languageKeywords = []string{
		"across", "activate", "address", "area", "again", "allow_mixing", "allowmixing", "also",
		"alt", "annotation", "as", "at", "attribute", "attributes", "autonumber", "backward",
		"bar", "bold", "bottom", "box", "break", "case", "caption", "center",
		"circle", "circled", "circles", "color", "create", "critical", "dashed", "deactivate",
		"description", "destroy", "detach", "dotted", "down", "else", "else", "elseif",
		"empty", "end", "endcaption", "endfooter", "endfork", "endheader", "endif", "endlegend",
		"endmerge", "endswitch", "endtitle", "endwhile", "false", "field", "fields", "floating note",
		"footbox", "footer", "fork", "goto", "grid", "group", "h-axis", "header",
		"hide", "highlight", "hnote", "if", "is", "italic", "kill", "label",
		"left", "left to right direction", "legend", "line", "link", "loop", "mainframe", "member",
		"members", "merge", "method", "methods", "namespace", "newpage", "normal", "not",
		"note", "null", "of", "on", "opt", "order", "orientation", "over",
		"package", "page", "par", "partition", "plain", "private", "protected", "public",
		"ref", "repeat", "repeatwhile", "return", "right", "rnote", "rotate", "scatter",
		"shape", "show", "skin", "skinparam", "spacing", "split", "sprite", "stackMode",
		"start", "stereotype", "stereotypes", "stop", "style", "switch", "then", "ticks",
		"title", "together", "top", "top to bottom direction", "true", "up", "v-axis", "v2-axis",
		"while",
	}
	languagePreprocessor = []string{
		"!assert", "!define", "!definelong", "!dump_memory", "!else", "!end", "!enddefinelong", "!endfor",
		"!endfunction", "!endif", "!endprocedure", "!endsub", "!endwhile", "!exit", "!final", "!foreach",
		"!function", "!if", "!ifdef", "!ifndef", "!import", "!include", "!includedef", "!includesub",
		"!local", "!log", "!option", "!pragma", "!procedure", "!return", "!startsub", "!theme",
		"!undef", "!unquoted", "!while",
	}
)

// Lists of LanguageDescriptor.print that come from other parts of upstream.
var (
	languageSkinParameters = klimt.SkinParamPossibleValues
	languageColors         = func() []string { return color.Instance().Names() }
)

// sortedSet returns the words sorted and without duplicates, like the
// TreeSet of LanguageDescriptor. Byte order equals Java's String order for
// these ASCII words.
func sortedSet(lists ...[]string) []string {
	out := slices.Concat(lists...)
	slices.Sort(out)
	return slices.Compact(out)
}

// WriteLanguage writes upstream's language description
// (LanguageDescriptor.print), the body of GET /language: for each of the
// sections type, keyword, preprocessor, skinparameter and color, a line
// ";name", a line ";count", the words, and an empty line; then ";EOF".
func WriteLanguage(w io.Writer) error {
	keywords := slices.Clone(languageKeywords)
	for _, d := range languageDirectives {
		keywords = append(keywords, "@start"+d, "@end"+d)
	}
	bw := bufio.NewWriter(w)
	for _, section := range []struct {
		name  string
		words []string
	}{
		{"type", sortedSet(languageTypes)},
		{"keyword", sortedSet(keywords)},
		{"preprocessor", sortedSet(languagePreprocessor)},
		{"skinparameter", sortedSet(languageSkinParameters())},
		{"color", languageColors()},
	} {
		bw.WriteString(";" + section.name + "\n;" + strconv.Itoa(len(section.words)) + "\n")
		for _, word := range section.words {
			bw.WriteString(word + "\n")
		}
		bw.WriteString("\n")
	}
	bw.WriteString(";EOF\n")
	return bw.Flush()
}
