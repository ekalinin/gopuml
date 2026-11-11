// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/hcl/HclSource.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package hcl

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// hclSource holds the characters of the HCL lines, joined without line
// breaks, as upstream joins them. A line that starts with "#" once trimmed
// is dropped.
type hclSource struct {
	all []rune
}

func (s *hclSource) add(line string) {
	if strings.HasPrefix(text.JavaTrim(line), "#") {
		return
	}
	s.all = append(s.all, []rune(line)...)
}
