// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/text"
)

// located returns the lines as located lines.
func located(lines ...string) []text.StringLocated {
	result := make([]text.StringLocated, len(lines))
	for i, l := range lines {
		result[i] = text.NewStringLocated(l, nil)
	}
	return result
}

func TestStyleExtractor(t *testing.T) {
	e := NewStyleExtractor(located("@startjson", "title My data", "scale 2", "<style>",
		"jsonDiagram { node { BackGroundColor #eeeeee } }", "</style>", "", `{"a": 1}`, "title not a title", "@endjson"))
	if want := []string{"@startjson", `{"a": 1}`, "title not a title", "@endjson"}; !slices.Equal(e.Lines(), want) {
		t.Errorf("lines %q, want %q", e.Lines(), want)
	}
	if e.Title() != "My data" || e.Scale() != "scale 2" || len(e.style) != 1 || e.IsHandwritten() {
		t.Errorf("title %q, scale %q, %d style lines, handwritten %v", e.Title(), e.Scale(), len(e.style), e.IsHandwritten())
	}
}

func TestStyleExtractorSkinparam(t *testing.T) {
	e := NewStyleExtractor(located("@startjson", "skinparam handwritten true", "skinparam x {", "a b", "}",
		"!pragma layout smetana", "hide empty", "skin rose", `{"a": 1}`, "skinparam y z", "@endjson"))
	if want := []string{"@startjson", `{"a": 1}`, "skinparam y z", "@endjson"}; !slices.Equal(e.Lines(), want) {
		t.Errorf("lines %q, want %q", e.Lines(), want)
	}
	if !e.IsHandwritten() || e.newSkin != "rose" {
		t.Errorf("handwritten %v, skin %q", e.IsHandwritten(), e.newSkin)
	}
}
