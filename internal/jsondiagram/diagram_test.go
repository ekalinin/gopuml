// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// createJSON builds the diagram of a @startjson block with lines.
func createJSON(t *testing.T, lines ...string) *JsonDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	src := "@startjson\n" + strings.Join(lines, "\n") + "\n@endjson\n"
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	d, ok := result.(*JsonDiagram)
	if !ok {
		t.Fatalf("%q: %T, want a JSON diagram", lines, result)
	}
	return d
}

// svgOf renders d as SVG.
func svgOf(t *testing.T, d *JsonDiagram) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.ExportDiagramContext(t.Context(), &buf, 0, export.NewFileFormatOption(font.FormatSVG)); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestJSONFactory(t *testing.T) {
	d := createJSON(t, `{"a": 1}`)
	if d.DiagramType() != core.JSON || d.Description() != "(Json)" || d.root == nil || !d.root.IsObject() {
		t.Errorf("type %v, description %q, root %v", d.DiagramType(), d.Description(), d.root)
	}
	if svg := svgOf(t, d); !strings.Contains(svg, ">a</text>") || !strings.Contains(svg, ">1</text>") {
		t.Errorf("SVG without the row a 1: %.500s", svg)
	}
}

func TestJSONRoots(t *testing.T) {
	d := createJSON(t, `"hello"`)
	if !d.root.IsArray() || d.root.AsArray().Size() != 1 {
		t.Errorf(`"hello": root %v, want an array of one item`, d.root)
	}
	d = createJSON(t, `{}`)
	if !d.root.IsArray() || d.root.AsArray().Size() != 1 {
		t.Errorf(`{}: root %v, want an array of one empty string`, d.root)
	} else if v, _ := d.root.AsArray().Get(0); !v.IsString() || v.AsString() != "" {
		t.Errorf(`{}: item %v, want the empty string`, v)
	}
}

func TestJSONUnreadable(t *testing.T) {
	for _, lines := range [][]string{{`{"a": }`}, nil} {
		d := createJSON(t, lines...)
		if d.root != nil {
			t.Errorf("%q: root %v, want none", lines, d.root)
		}
		// The monospaced text has no-break spaces, as upstream writes them.
		want := ">" + strings.ReplaceAll("Your data does not sound like JSON data", " ", "\u00a0") + "</text>"
		if svg := svgOf(t, d); !strings.Contains(svg, want) {
			t.Errorf("%q: SVG without the text: %.500s", lines, svg)
		}
	}
}

func TestJSONHashLine(t *testing.T) {
	d := createJSON(t, "#foo bar", `{"a": 1}`)
	if d.root == nil || len(d.highlighted) != 0 {
		t.Errorf("root %v, %d highlight paths, want the object and none", d.root, len(d.highlighted))
	}
	d = createJSON(t, `#highlight "a"`, `{"a": 1}`)
	if len(d.highlighted) != 1 {
		t.Errorf("%d highlight paths, want 1", len(d.highlighted))
	}
}

func TestJSONTitleAndScale(t *testing.T) {
	d := createJSON(t, "title My data", "scale 2", `{"a": 1}`)
	if d.Title().Display() == nil || d.Scale() == nil {
		t.Errorf("title %v, scale %v", d.Title(), d.Scale())
	}
	if svg := svgOf(t, d); !strings.Contains(svg, ">My data</text>") {
		t.Errorf("SVG without the title: %.500s", svg)
	}
}

func TestJSONCancelledContext(t *testing.T) {
	d := createJSON(t, `{"a": {"b": 1}}`)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var buf bytes.Buffer
	if _, err := d.ExportDiagramContext(ctx, &buf, 0, export.NewFileFormatOption(font.FormatSVG)); !errors.Is(err, context.Canceled) {
		t.Errorf("error %v, want context.Canceled", err)
	}
	if buf.Len() != 0 {
		t.Errorf("%d bytes written, want none", buf.Len())
	}
}
