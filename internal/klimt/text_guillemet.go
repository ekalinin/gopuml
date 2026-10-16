// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/text/Guillemet.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"
)

// Guillemet is the way stereotypes are written: "<<" and ">>", the « and »
// guillemets, nothing, or a custom pair. Upstream defines it in the text
// package; it lives here because it has no other user outside the drawing
// model. The three predefined values are compared by identity.
type Guillemet struct {
	start string
	end   string
}

var (
	guillemetNone             = &Guillemet{start: "", end: ""}
	guillemetDoubleComparator = &Guillemet{start: "<<", end: ">>"}
	guillemetGuillemet        = &Guillemet{start: "«", end: "»"}
)

// GuillemetNone returns the empty pair.
func GuillemetNone() *Guillemet { return guillemetNone }

// GuillemetDoubleComparator returns "<<" and ">>", which leaves texts
// unchanged.
func GuillemetDoubleComparator() *Guillemet { return guillemetDoubleComparator }

// GuillemetGuillemet returns « and ».
func GuillemetGuillemet() *Guillemet { return guillemetGuillemet }

// FromDescription returns the pair of "skinparam guillemet": "false" or
// "<< >>" for "<<" and ">>", "none", a custom "start end", and « » otherwise
// (also without value, ok false).
func (g *Guillemet) FromDescription(value string, ok bool) *Guillemet {
	if ok {
		if javaEqualsIgnoreCase("false", value) {
			return guillemetDoubleComparator
		}
		if javaEqualsIgnoreCase("<< >>", value) {
			return guillemetDoubleComparator
		}
		if javaEqualsIgnoreCase("none", value) {
			return guillemetNone
		}
		if strings.Contains(value, " ") {
			tokens := javaTokenize(value, " ")
			if len(tokens) < 2 {
				throw(noSuchElement, "")
			}
			return &Guillemet{start: tokens[0], end: tokens[1]}
		}
	}
	return guillemetGuillemet
}

var guillemetPattern = regexp.MustCompile(`<<[\t\n\v\f\r ]?((?:<&\w+>|[^<>])+?)[\t\n\v\f\r ]?>>`)

// ManageGuillemet replaces every "<<x>>" by the pair around x.
func (g *Guillemet) ManageGuillemet(st string) string {
	if g == guillemetDoubleComparator {
		return st
	}
	if !strings.Contains(st, "<") {
		return st
	}
	quote := func(s string) string { return strings.ReplaceAll(s, "$", "$$") }
	return guillemetPattern.ReplaceAllString(st, quote(g.start)+"${1}"+quote(g.end))
}

// ManageGuillemetStrict replaces the "<<" and ">>" (with an inner space) at
// the ends of st by the pair.
func (g *Guillemet) ManageGuillemetStrict(st string) string {
	if g == guillemetDoubleComparator {
		return st
	}
	if strings.HasPrefix(st, "<< ") {
		st = g.start + st[3:]
	} else if strings.HasPrefix(st, "<<") {
		st = g.start + st[2:]
	}
	if strings.HasSuffix(st, " >>") {
		st = st[:len(st)-3] + g.end
	} else if strings.HasSuffix(st, ">>") {
		st = st[:len(st)-2] + g.end
	}
	return st
}

// String returns the pair, for debugging.
func (g *Guillemet) String() string { return g.start + " " + g.end }
