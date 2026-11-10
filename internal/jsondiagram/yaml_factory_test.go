// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
)

// createYAML builds the diagram of a @startyaml block with lines.
func createYAML(t *testing.T, lines ...string) *JsonDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@startyaml\n" + strings.Join(lines, "\n") + "\n@endyaml\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := YamlFactory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	d, ok := result.(*JsonDiagram)
	if !ok {
		t.Fatalf("%q: %T, want a YAML diagram", lines, result)
	}
	return d
}

func TestYAMLFactory(t *testing.T) {
	d := createYAML(t, `#highlight "info" / "origin"`, "info:", "  origin: Spain")
	if d.DiagramType() != core.YAML || d.Description() != "(Yaml)" || d.root == nil || len(d.highlighted) != 1 {
		t.Fatalf("type %v, description %q, root %v, %d highlight paths", d.DiagramType(), d.Description(), d.root,
			len(d.highlighted))
	}
	if svg := svgOf(t, d); !strings.Contains(svg, ">origin</text>") || !strings.Contains(svg, `fill="#CCFF02"`) {
		t.Errorf("SVG without the highlighted row origin: %.500s", svg)
	}
}

func TestYAMLUnreadable(t *testing.T) {
	for _, lines := range [][]string{{"just some text"}, {"tags: [a, b, c]", "- x"}, {"#highlight <<h1>>", "a: 1"}, nil} {
		d := createYAML(t, lines...)
		want := ">" + strings.ReplaceAll("Your data does not sound like YAML data", " ", "\u00a0") + "</text>"
		if svg := svgOf(t, d); d.root != nil || !strings.Contains(svg, want) {
			t.Errorf("%q: root %v, SVG %.300s", lines, d.root, svg)
		}
	}
}

func TestYAMLTitleAndStyle(t *testing.T) {
	d := createYAML(t, "<style>", "yamlDiagram {", "node { BackGroundColor #ffeecc }", "}", "</style>", "title Fruits",
		"a:", "  b: 1")
	svg := svgOf(t, d)
	if !strings.Contains(svg, ">Fruits</text>") || !strings.Contains(svg, `fill="#FEC"`) {
		t.Errorf("SVG without the title and the style: %.600s", svg)
	}
}
