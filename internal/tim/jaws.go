// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jaws/Jaws.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// Private-use characters that the preprocessor puts in lines.
const (
	jawsNewline   = '\uE100'
	jawsBreakline = '\uE103'
)

// mutateExpands1 is Jaws.mutateExpands1: lines are split at every
// breakline character that is not inside a {{...}} block. Upstream's
// triple-mark merge is disabled (JawsFlags.PARSE_NEW_MULTILINE_TRIPLE_MARKS
// is false) and not ported.
func mutateExpands1(lines []text.StringLocated) []text.StringLocated {
	var out []text.StringLocated
	for _, sl := range lines {
		if !strings.ContainsRune(sl.S, jawsBreakline) {
			out = append(out, sl)
			continue
		}
		u := toUTF16(sl.S)
		var pending []uint16
		level := 0
		for j, ch := range u {
			if ch == '{' && j+1 < len(u) && u[j+1] == '{' {
				level++
			} else if ch == '}' && j+1 < len(u) && u[j+1] == '}' {
				level--
			}
			switch {
			case level > 0:
				pending = append(pending, ch)
			case ch == jawsBreakline:
				out = append(out, text.StringLocated{S: fromUTF16(pending), Location: sl.Location, PreprocessorError: sl.PreprocessorError})
				pending = pending[:0]
			default:
				pending = append(pending, ch)
			}
		}
		// Tail piece (always emitted, even when empty, to match the legacy
		// behaviour).
		out = append(out, text.StringLocated{S: fromUTF16(pending), Location: sl.Location, PreprocessorError: sl.PreprocessorError})
	}
	return out
}
