// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/Highlighted.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// highlightedPrefix starts the line of a highlight path.
const highlightedPrefix = "#highlight "

var highlightedPattern = regexp.MustCompile(`^([^<>]+)(<<.*>>)?$`)

// Highlighted is a highlight path of a JSON or YAML diagram: the keys that
// lead to a row from the box it applies to, and the stereotype whose style
// draws the row.
type Highlighted struct {
	paths      []string
	stereotype *klimt.Stereotype
}

// MatchesHighlightedDefinition reports whether line is a highlight path.
func MatchesHighlightedDefinition(line string) bool {
	return strings.HasPrefix(line, highlightedPrefix)
}

// BuildHighlighted reads the highlight path of line, such as
// `#highlight "a" / "b" <<s>>`. It panics with an IllegalStateException, as
// upstream throws, when line is not a highlight path.
func BuildHighlighted(line string) *Highlighted {
	if !MatchesHighlightedDefinition(line) {
		uml.Throw(uml.IllegalStateException, "")
	}
	line = text.JavaTrim(line[len(highlightedPrefix):])
	m := highlightedPattern.FindStringSubmatch(line)
	if m == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	var stereotype *klimt.Stereotype
	if m[2] != "" {
		stereotype = klimt.StereotypeBuild(m[2])
	}
	return &Highlighted{paths: highlightedPaths(text.JavaTrim(m[1])), stereotype: stereotype}
}

// highlightedPaths splits paths at "/" and removes the quotes of each key.
func highlightedPaths(paths string) []string {
	var result []string
	for _, s := range klimt.JavaSplit(paths, "/") {
		result = append(result, text.RemoveQuotesFormat(text.JavaTrim(s), `"`))
	}
	return result
}

// upOneLevel returns the path for the box of the value of key, or nil when
// the path does not lead there.
func (h *Highlighted) upOneLevel(key string) *Highlighted {
	if len(h.paths) <= 1 {
		return nil
	}
	first := h.paths[0]
	if first == "**" {
		return &Highlighted{paths: h.paths, stereotype: h.stereotype}
	}
	if first == "*" || first == key {
		return &Highlighted{paths: h.paths[1:], stereotype: h.stereotype}
	}
	return nil
}

// isKeyHighlight reports whether the path names the row of key.
func (h *Highlighted) isKeyHighlight(key string) bool {
	if len(h.paths) == 2 && h.paths[0] == "**" && h.paths[1] == key {
		return true
	}
	return len(h.paths) == 1 && h.paths[0] == key
}

// Stereotype returns the stereotype of the path, nil when it has none.
func (h *Highlighted) Stereotype() *klimt.Stereotype { return h.stereotype }
