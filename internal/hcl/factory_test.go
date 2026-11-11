// SPDX-License-Identifier: MIT
// Provenance: original

package hcl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/jsondiagram"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// svgOfHCL renders a @starthcl block with lines as SVG.
func svgOfHCL(t *testing.T, lines ...string) (*jsondiagram.JsonDiagram, string) {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@starthcl\n" + strings.Join(lines, "\n") + "\n@endhcl\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	d, ok := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true)).(*jsondiagram.JsonDiagram)
	if !ok {
		t.Fatal("not a JSON diagram")
	}
	var buf bytes.Buffer
	if _, err := d.ExportDiagramContext(t.Context(), &buf, 0, export.NewFileFormatOption(font.FormatSVG)); err != nil {
		t.Fatal(err)
	}
	return d, buf.String()
}

func TestHCLFactory(t *testing.T) {
	d, svg := svgOfHCL(t, "a {", `  b = "c"`, "}")
	if d.DiagramType() != core.HCL || d.Description() != "(HCL)" {
		t.Errorf("type %v, description %q", d.DiagramType(), d.Description())
	}
	if !strings.Contains(svg, ">b</text>") || !strings.Contains(svg, ">c</text>") {
		t.Errorf("SVG without the row b c: %.500s", svg)
	}
}

func TestHCLStyleAndTitleIgnored(t *testing.T) {
	_, svg := svgOfHCL(t, "<style>", "yamlDiagram { node { BackGroundColor #ffeecc } }", "</style>", "title Ignored",
		"a {", `  b = "c"`, "}")
	if strings.Contains(svg, "Ignored") || strings.Contains(svg, `fill="#FEC"`) || !strings.Contains(svg, `fill="#F1F1F1"`) {
		t.Errorf("SVG with the title or the style: %.600s", svg)
	}
}

func TestHCLUnreadable(t *testing.T) {
	want := ">" + strings.ReplaceAll("Your data does not sound like HCL data", " ", "\u00a0") + "</text>"
	if _, svg := svgOfHCL(t, `a = "b"`); !strings.Contains(svg, want) {
		t.Errorf("SVG without the text: %.400s", svg)
	}
}
